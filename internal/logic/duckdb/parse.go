package logic_duckdb

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	"github.com/andybalholm/brotli"
	_ "github.com/marcboeker/go-duckdb"
)

// downloadDuckDB downloads the DuckDB file from CloudFront
func downloadDuckDB(dateString string) error {
	baseURL := os.Getenv("BATORMENT_DUCKDB_REMOTE_URL")
	if baseURL == "" {
		return fmt.Errorf("BATORMENT_DUCKDB_REMOTE_URL environment variable is not set")
	}

	url := fmt.Sprintf("%s/v1/JP/%s.db", baseURL, dateString)
	fileName := fmt.Sprintf("%s.db", dateString)

	log.Printf("Downloading DuckDB from: %s", url)

	// Create HTTP client that accepts Brotli encoding
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Accept Brotli encoding (CloudFront sends Brotli-compressed files)
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
	log.Printf("Response Content-Type: %s, Content-Length: %d, Content-Encoding: %s",
		resp.Header.Get("Content-Type"), resp.ContentLength, contentEncoding)

	// Create output file
	out, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer out.Close()

	// Decompress if Brotli encoded
	var reader io.Reader = resp.Body
	if contentEncoding == "br" {
		log.Printf("Decompressing Brotli-encoded file...")
		reader = brotli.NewReader(resp.Body)
	}

	// Copy (and decompress) to file
	written, err := io.Copy(out, reader)
	if err != nil {
		os.Remove(fileName) // Clean up incomplete file
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	log.Printf("Downloaded and wrote %d bytes (%.2f MB) to %s", written, float64(written)/(1024*1024), fileName)

	// Verify minimum file size (DuckDB files should be substantial)
	if written < 10240 { // At least 10KB
		os.Remove(fileName)
		return fmt.Errorf("downloaded file is too small (%d bytes), likely not a valid DuckDB file", written)
	}

	return nil
}

func ParseDuckDB(contentID string, startDate time.Time) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	// Check if DuckDB file exists, if not, try to download it
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		log.Printf("DuckDB file %s not found, attempting to download...", dbFileName)
		if err := downloadDuckDB(dateString); err != nil {
			log.Printf("Info: Failed to download DuckDB file %s: %v. Skipping this raid.", dbFileName, err)
			return nil, nil, fmt.Errorf("duckdb file not available: %w", err)
		}
		log.Printf("Successfully downloaded DuckDB file: %s", dbFileName)
	}

	db, err := sql.Open("duckdb", dbFileName)
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
		// Skip data with score 0
		if run.score == 0 {
			break
		}

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

// isEliminationRaid checks if the contentID represents an elimination raid (대결전)
// Elimination raids have contentID starting with "3S" (e.g., "3S3-1")
func isEliminationRaid(contentID string) bool {
	return strings.HasPrefix(contentID, "3S")
}

// GetPlatinumCuts retrieves score cutoffs at specific ranks (2000, 4000, ..., 20000)
// For elimination raids, it uses combined scores from all armor types
func GetPlatinumCuts(contentID string, startDate time.Time) ([]types.PlatinumCut, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	// Check if DuckDB file exists
	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		log.Printf("DuckDB file %s not found for platinum cuts", dbFileName)
		return nil, fmt.Errorf("duckdb file not available: %w", err)
	}

	db, err := sql.Open("duckdb", dbFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open duckdb: %w", err)
	}
	defer db.Close()

	// Target ranks: 2000, 4000, 6000, ..., 20000
	ranks := []int{2000, 4000, 6000, 8000, 10000, 12000, 14000, 16000, 18000, 20000}

	var querySQL string
	if isEliminationRaid(contentID) {
		querySQL = GetEliminationPlatinumCutSQL(ranks)
	} else {
		querySQL = GetPlatinumCutSQL(ranks)
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
