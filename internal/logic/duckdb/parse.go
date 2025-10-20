package logic_duckdb

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	_ "github.com/marcboeker/go-duckdb"
)

func ParseDuckDB(contentID string, startDate time.Time) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	dateString := startDate.Format("20060102")

	db, err := sql.Open("duckdb", fmt.Sprintf("%s.db", dateString))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	_, category := logic.SplitSeasonString(contentID)

	armorType := constants.ArmorTypeMapping[category]

	partyData, filterResult, err := processArmorType(db, armorType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process armor type: %w", err)
	}

	if len(partyData.PartyDetail) == 0 {
		return nil, nil, fmt.Errorf("no details found for armor type: %s", armorType)
	}

	return partyData, filterResult, nil
}

func processArmorType(db *sql.DB, armorType string) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	// Step 1: Get complete run IDs and scores
	completeRunsSQL := GetCompleteRunIDAndScoreSQL(armorType)
	rows, err := db.Query(completeRunsSQL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query complete runs: %w", err)
	}
	defer rows.Close()

	type runData struct {
		completeRunID int
		score         int
	}
	var runs []runData

	for rows.Next() {
		var completeRunID, score int
		if err := rows.Scan(&completeRunID, &score); err != nil {
			return nil, nil, fmt.Errorf("failed to scan complete run: %w", err)
		}
		runs = append(runs, runData{completeRunID, score})
	}

	// Sort by score descending
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].score > runs[j].score
	})

	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))
	var parties []types.BATormentPartyDetail

	minPartys := 99
	maxPartys := 0

	for rank, run := range runs {
		// Step 2: Get run IDs by complete run ID
		partyData, err := getPartiesByCompleteRunID(db, armorType, run.completeRunID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get parties: %w", err)
		}

		// Update filters
		for _, party := range partyData {
			for _, member := range party {
				if member == 0 {
					continue
				}
				logic.UpdatePartyFilters(filters, assistFilters, member)
			}
		}

		partyInfo := types.BATormentPartyDetail{
			Rank:      rank + 1,
			Score:     run.score,
			PartyData: partyData,
		}
		parties = append(parties, partyInfo)

		if len(partyData) < minPartys {
			minPartys = len(partyData)
		}
		if len(partyData) > maxPartys {
			maxPartys = len(partyData)
		}
	}

	filterResult := &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}

	result := &types.BATormentPartyData{
		MinPartys:   minPartys,
		MaxPartys:   maxPartys,
		PartyDetail: parties,
	}

	return result, filterResult, nil
}

func getPartiesByCompleteRunID(db *sql.DB, armorType string, completeRunID int) ([][6]int, error) {
	runIDsSQL := GetRunIDsByCompleteRunIDSQL(armorType, completeRunID)
	rows, err := db.Query(runIDsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query run IDs: %w", err)
	}
	defer rows.Close()

	var parties [][6]int

	for rows.Next() {
		var runID int
		if err := rows.Scan(&runID); err != nil {
			return nil, fmt.Errorf("failed to scan run ID: %w", err)
		}

		// Step 3: Get party info by run ID
		party, err := getPartyByRunID(db, armorType, runID)
		if err != nil {
			return nil, fmt.Errorf("failed to get party: %w", err)
		}

		parties = append(parties, party)
	}

	return parties, nil
}

func getPartyByRunID(db *sql.DB, armorType string, runID int) ([6]int, error) {
	partySQL := GetPartyInfoByRunIDSQL(armorType, runID)
	rows, err := db.Query(partySQL)
	if err != nil {
		return [6]int{}, fmt.Errorf("failed to query party info: %w", err)
	}
	defer rows.Close()

	// Initialize party members array [4 strikers + 2 specials]
	var partyMembers [6]int
	specialIndex := 0

	for rows.Next() {
		var sid, level, slot int
		var build string
		var assist bool

		if err := rows.Scan(&sid, &build, &level, &slot, &assist); err != nil {
			return [6]int{}, fmt.Errorf("failed to scan party info: %w", err)
		}

		// Map build to weapon star
		weaponValue, exists := constants.WeaponStarMapping[build]
		if !exists {
			return [6]int{}, fmt.Errorf("unknown build value: %s", build)
		}

		star := weaponValue / 10
		weaponStar := weaponValue % 10

		// Create student detail ID (8 digits)
		studentDetailID := logic.GetStudentDetailIDInt(sid, star, weaponStar, assist)

		// slot 0-3: strikers (positions 0-3)
		// slot 4: specials (position 4-5, append sequentially)
		if slot <= 3 {
			partyMembers[slot] = studentDetailID
		} else {
			// Find first empty special slot
			if specialIndex < 2 {
				partyMembers[4+specialIndex] = studentDetailID
				specialIndex++
			}
		}
	}

	return partyMembers, nil
}
