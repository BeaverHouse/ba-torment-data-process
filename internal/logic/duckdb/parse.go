package logic_duckdb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	_ "github.com/marcboeker/go-duckdb"
)

func ParseDuckDB(contentID string, startDate time.Time) error {
	dateString := startDate.Format("20060102")

	db, err := sql.Open("duckdb", fmt.Sprintf("../../%s.db", dateString))
	if err != nil {
		return fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	if err := os.MkdirAll("../../data", 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	_, category := logic.SplitSeasonString(contentID)

	armorType := constants.ArmorTypeMapping[category]

	details, err := processArmorType(db, armorType)
	if err != nil {
		return fmt.Errorf("failed to process armor type: %w", err)
	}

	if len(details) == 0 {
		return fmt.Errorf("no details found for armor type: %s", armorType)
	}

	result := types.AronaAIData{D: details}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal json for %s: %w", armorType, err)
	}

	filename := fmt.Sprintf("../../data/%s.json", contentID)
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write json file for %s: %w", armorType, err)
	}

	return nil
}

func processArmorType(db *sql.DB, armorType string) ([]types.AronaAIDetail, error) {
	// Step 1: Get complete run IDs and scores
	completeRunsSQL := GetCompleteRunIDAndScoreSQL(armorType)
	rows, err := db.Query(completeRunsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query complete runs: %w", err)
	}
	defer rows.Close()

	var details []types.AronaAIDetail
	rank := 1

	for rows.Next() {
		var completeRunID, score int
		if err := rows.Scan(&completeRunID, &score); err != nil {
			return nil, fmt.Errorf("failed to scan complete run: %w", err)
		}

		// Step 2: Get run IDs by complete run ID
		parties, err := getPartiesByCompleteRunID(db, armorType, completeRunID)
		if err != nil {
			return nil, fmt.Errorf("failed to get parties: %w", err)
		}

		details = append(details, types.AronaAIDetail{
			R: rank,
			S: score,
			T: parties,
		})
		rank++
	}

	return details, nil
}

func getPartiesByCompleteRunID(db *sql.DB, armorType string, completeRunID int) ([]types.AronaAIParty, error) {
	runIDsSQL := GetRunIDsByCompleteRunIDSQL(armorType, completeRunID)
	rows, err := db.Query(runIDsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query run IDs: %w", err)
	}
	defer rows.Close()

	var parties []types.AronaAIParty

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

func getPartyByRunID(db *sql.DB, armorType string, runID int) (types.AronaAIParty, error) {
	partySQL := GetPartyInfoByRunIDSQL(armorType, runID)
	rows, err := db.Query(partySQL)
	if err != nil {
		return types.AronaAIParty{}, fmt.Errorf("failed to query party info: %w", err)
	}
	defer rows.Close()

	// Initialize fixed-size arrays
	strikers := make([]types.AronaAICharacter, 4)
	specials := make([]types.AronaAICharacter, 2)

	// Fill with empty characters
	emptyChar := types.AronaAICharacter{
		StudentID:  0,
		Star:       0,
		HasWeapon:  false,
		WeaponStar: 0,
		IsAssist:   false,
	}
	for i := range 4 {
		strikers[i] = emptyChar
	}
	for i := range 2 {
		specials[i] = emptyChar
	}

	for rows.Next() {
		var sid, level, slot int
		var build string
		var assist bool

		if err := rows.Scan(&sid, &build, &level, &slot, &assist); err != nil {
			return types.AronaAIParty{}, fmt.Errorf("failed to scan party info: %w", err)
		}

		// Map build to weapon star
		weaponValue, exists := constants.WeaponStarMapping[build]
		if !exists {
			return types.AronaAIParty{}, fmt.Errorf("unknown build value: %s", build)
		}

		star := weaponValue / 10
		weaponStar := weaponValue % 10
		hasWeapon := weaponStar >= 1

		character := types.AronaAICharacter{
			StudentID:  sid,
			Star:       star,
			HasWeapon:  hasWeapon,
			WeaponStar: weaponStar,
			IsAssist:   assist,
		}

		// slot 0-3: strikers (positions 0-3)
		// slot 4: specials (position 0-1, append sequentially)
		if slot <= 3 {
			strikers[slot] = character
		} else {
			// Find first empty special slot
			for i := range 2 {
				if specials[i].StudentID == 0 {
					specials[i] = character
					break
				}
			}
		}
	}

	return types.AronaAIParty{
		M: strikers,
		S: specials,
	}, nil
}
