package party

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
	"ba-torment-data-process/internal/ui"

	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/andybalholm/brotli"
	_ "github.com/marcboeker/go-duckdb"
)

func downloadDuckDB(dateString string) error {
	baseURL := env.GetEnv("BATORMENT_DUCKDB_REMOTE_URL", "")
	if baseURL == "" {
		return fmt.Errorf("BATORMENT_DUCKDB_REMOTE_URL environment variable is not set")
	}

	url := fmt.Sprintf("%s/v2/JP/%s.db", baseURL, dateString)
	fileName := fmt.Sprintf("%s.db", dateString)

	ui.Log.Info("Downloading DuckDB", logger.F("url", url))

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept-Encoding", "br, gzip, deflate")
	req.Header.Set("User-Agent", "ba-torment-data-process/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d from %s", resp.StatusCode, url)
	}

	contentEncoding := resp.Header.Get("Content-Encoding")
	ui.Log.Info("Response metadata",
		logger.F("contentType", resp.Header.Get("Content-Type")),
		logger.F("contentLength", resp.ContentLength),
		logger.F("contentEncoding", contentEncoding))

	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if contentEncoding == "br" {
		ui.Log.Info("Decompressing Brotli-encoded file...")
		reader = brotli.NewReader(resp.Body)
	}

	written, err := io.Copy(out, reader)
	if err != nil {
		os.Remove(fileName)
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	ui.Log.Info("Downloaded file", logger.F("bytes", written), logger.F("mb", fmt.Sprintf("%.2f", float64(written)/(1024*1024))), logger.F("file", fileName))

	if written < 10240 {
		os.Remove(fileName)
		return fmt.Errorf("downloaded file is too small (%d bytes), likely not a valid DuckDB file", written)
	}

	return nil
}

func ParseDuckDB(contentID string, startDate time.Time) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		ui.Log.Info("DuckDB file not found, attempting to download", logger.F("file", dbFileName))
		if err := downloadDuckDB(dateString); err != nil {
			ui.Log.Info("Failed to download DuckDB file, skipping this raid", logger.F("file", dbFileName), logger.F("error", err))
			return nil, nil, fmt.Errorf("duckdb file not available: %w", err)
		}
		ui.Log.Info("Successfully downloaded DuckDB file", logger.F("file", dbFileName))
	}

	db, err := sql.Open("duckdb", dbFileName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	_, category := id.SplitSeasonString(contentID)

	armorType := constants.ArmorTypeMapping[category]

	partyData, filterResult, err := processArmorType(db, armorType)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to process armor type: %w", err)
	}

	if len(partyData.PartyDetail) == 0 {
		return nil, nil, fmt.Errorf("no details found for armor type: %s", armorType)
	}

	removeFraudUsers(contentID, partyData)

	return partyData, filterResult, nil
}

func removeFraudUsers(contentID string, partyData *types.BATormentPartyData) {
	if contentID != "S80-0" {
		return
	}

	fraudIndex := -1
	for i, party := range partyData.PartyDetail {
		if party.Rank == 13622 {
			hasFraudChar := false
			for _, members := range party.PartyData {
				for _, member := range members {
					if member == 10059510 {
						hasFraudChar = true
						break
					}
				}
				if hasFraudChar {
					break
				}
			}
			if hasFraudChar {
				fraudIndex = i
				ui.Log.Warn("Found fraud user at rank 13622 with Mika star1 UE3 in S80-0")
			}
			break
		}
	}

	if fraudIndex == -1 {
		return
	}

	partyData.PartyDetail = append(partyData.PartyDetail[:fraudIndex], partyData.PartyDetail[fraudIndex+1:]...)

	for i := fraudIndex; i < len(partyData.PartyDetail); i++ {
		partyData.PartyDetail[i].Rank--
	}

	ui.Log.Info("Removed fraud user and adjusted ranks", logger.F("adjustedRanks", len(partyData.PartyDetail)-fraudIndex))
}

func processArmorType(db *sql.DB, armorType string) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	completeRunsSQL := getCompleteRunIDAndScoreSQL(armorType)
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

	sort.Slice(runs, func(i, j int) bool {
		return runs[i].score > runs[j].score
	})

	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))
	var parties []types.BATormentPartyDetail

	minPartys := 99
	maxPartys := 0

	for rank, run := range runs {
		if run.score == 0 {
			break
		}

		partyData, err := getPartiesByCompleteRunID(db, armorType, run.completeRunID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get parties: %w", err)
		}

		for _, party := range partyData {
			for _, member := range party {
				if member == 0 {
					continue
				}
				updatePartyFilters(filters, assistFilters, member)
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
	runIDsSQL := getRunIDsByCompleteRunIDSQL(armorType, completeRunID)
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

		party, err := getPartyByRunID(db, armorType, runID)
		if err != nil {
			return nil, fmt.Errorf("failed to get party: %w", err)
		}

		parties = append(parties, party)
	}

	return parties, nil
}

func getPartyByRunID(db *sql.DB, armorType string, runID int) ([6]int, error) {
	partySQL := getPartyInfoByRunIDSQL(armorType, runID)
	rows, err := db.Query(partySQL)
	if err != nil {
		return [6]int{}, fmt.Errorf("failed to query party info: %w", err)
	}
	defer rows.Close()

	var partyMembers [6]int
	specialIndex := 0

	for rows.Next() {
		var sid, level, slot int
		var build string
		var assist bool

		if err := rows.Scan(&sid, &build, &level, &slot, &assist); err != nil {
			return [6]int{}, fmt.Errorf("failed to scan party info: %w", err)
		}

		weaponValue, exists := constants.WeaponStarMapping[build]
		if !exists {
			return [6]int{}, fmt.Errorf("unknown build value: %s", build)
		}

		star := weaponValue / 10
		weaponStar := weaponValue % 10

		studentDetailID := id.ComposeStudentDetailID(sid, star, weaponStar, assist)

		if slot <= 3 {
			partyMembers[slot] = studentDetailID
		} else {
			if specialIndex < 2 {
				partyMembers[4+specialIndex] = studentDetailID
				specialIndex++
			}
		}
	}

	return partyMembers, nil
}

// IsGrandAssault checks if the contentID represents a grand assault
func IsGrandAssault(contentID string) bool {
	return strings.HasPrefix(contentID, "3S")
}

// GetPlatinumCuts retrieves score cutoffs at specific ranks
func GetPlatinumCuts(contentID string, startDate time.Time) ([]types.PlatinumCut, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		ui.Log.Warn("DuckDB file not found for platinum cuts", logger.F("file", dbFileName))
		return nil, fmt.Errorf("duckdb file not available: %w", err)
	}

	db, err := sql.Open("duckdb", dbFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	ranks := []int{100, 200, 500, 1000, 2000, 4000, 6000, 8000, 10000, 12000, 14000, 16000, 18000, 20000}

	var querySQL string
	if IsGrandAssault(contentID) {
		existingColumns, err := getExistingPointColumns(db)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing columns: %w", err)
		}
		if len(existingColumns) == 0 {
			return nil, fmt.Errorf("no point columns found for elimination raid")
		}
		querySQL = getPartialPlatinumCutSQL(ranks, existingColumns)
	} else {
		querySQL = getPlatinumCutSQL(ranks)
	}

	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query platinum cuts: %w", err)
	}
	defer rows.Close()

	var cuts []types.PlatinumCut
	for rows.Next() {
		var rank, score int
		if err := rows.Scan(&rank, &score); err != nil {
			return nil, fmt.Errorf("failed to scan platinum cut: %w", err)
		}
		cuts = append(cuts, types.PlatinumCut{
			Rank:  rank,
			Score: score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating platinum cuts: %w", err)
	}

	return cuts, nil
}

// GetPartPlatinumCutsFromPartyData extracts platinum cuts from partyData for grand assault
func GetPartPlatinumCutsFromPartyData(partyData *types.BATormentPartyData) []types.PlatinumCut {
	if partyData == nil || len(partyData.PartyDetail) == 0 {
		return nil
	}

	targetRanks := []int{100, 200, 500, 1000, 2000, 4000, 6000, 8000, 10000, 12000, 14000, 16000, 18000, 20000}

	var cuts []types.PlatinumCut
	for _, targetRank := range targetRanks {
		for _, party := range partyData.PartyDetail {
			if party.Rank == targetRank {
				cuts = append(cuts, types.PlatinumCut{
					Rank:  party.Rank,
					Score: party.Score,
				})
				break
			}
		}
	}

	return cuts
}
