package logic_duckdb

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"math"
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

	// Apply fraud user cleanup for specific content
	removeFraudUsers(contentID, partyData)

	return partyData, filterResult, nil
}

// removeFraudUsers removes known fraudulent user data and adjusts ranks
// S80-0 rank 13622: User with Mika (10059) star 5, UE1 (studentDetailID contains 10059510)
func removeFraudUsers(contentID string, partyData *types.BATormentPartyData) {
	if contentID != "S80-0" {
		return
	}

	// Find the fraud user at rank 13622 with Mika star 5 UE1 (studentDetailID 10059510)
	fraudIndex := -1
	for i, party := range partyData.PartyDetail {
		if party.Rank == 13622 {
			// Check if this user has Mika star 5 UE1 (studentDetailID 10059510)
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
				log.Printf("Found fraud user at rank 13622 with Mika star1 UE3 in S80-0")
			}
			break
		}
	}

	if fraudIndex == -1 {
		return
	}

	// Remove the fraud entry
	partyData.PartyDetail = append(partyData.PartyDetail[:fraudIndex], partyData.PartyDetail[fraudIndex+1:]...)

	// Shift ranks for all entries after the removed one (13623+ becomes 13622+)
	for i := fraudIndex; i < len(partyData.PartyDetail); i++ {
		partyData.PartyDetail[i].Rank--
	}

	log.Printf("Removed fraud user and adjusted %d ranks", len(partyData.PartyDetail)-fraudIndex)
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
		studentDetailID := logic.ComposeStudentDetailID(sid, star, weaponStar, assist)

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

// IsGrandAssault checks if the contentID represents an grand assault (대결전)
// Grand assaults have contentID starting with "3S" (e.g., "3S3-1")
func IsGrandAssault(contentID string) bool {
	return strings.HasPrefix(contentID, "3S")
}

// getExistingPointColumns returns the list of *_point columns that exist in complete_runs table
func getExistingPointColumns(db *sql.DB) ([]string, error) {
	rows, err := db.Query(GetCompleteRunsColumnsSQL())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	armorTypes := []string{"Light_point", "Heavy_point", "Special_point", "Elastic_point"}
	armorTypeSet := make(map[string]bool)
	for _, at := range armorTypes {
		armorTypeSet[at] = true
	}

	var existingColumns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		if armorTypeSet[columnName] {
			existingColumns = append(existingColumns, columnName)
		}
	}

	return existingColumns, nil
}

// GetEssentialCharacters returns characters used by 70%+ of users (excluding assists)
// Results are sorted by usage ratio in descending order
// Returns separate results for Torment and Lunatic difficulties
func GetEssentialCharacters(partyData *types.BATormentPartyData) (torment []types.EssentialCharacter, lunatic []types.EssentialCharacter) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	// Check if this is INSANE difficulty (top score < TormentMinScore)
	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	tormentCharCount := make(map[int]int)
	lunaticCharCount := make(map[int]int)
	var tormentUsers, lunaticUsers int

	for _, party := range partyData.PartyDetail {
		// Only count platinum ranks (top 20000)
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		// Determine difficulty
		isLunatic := party.Score >= constants.LunaticMinScore
		isTorment := isInsane || party.Score >= constants.TormentMinScore

		if isLunatic {
			lunaticUsers++
		} else if isTorment {
			tormentUsers++
		} else {
			continue
		}

		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				// Skip assists (last digit is 1)
				if member%10 == 1 {
					continue
				}
				// Extract studentID (first 5 digits)
				studentID := logic.GetStudentID(member)
				if isLunatic {
					lunaticCharCount[studentID]++
				} else {
					tormentCharCount[studentID]++
				}
			}
		}
	}

	// Helper function to calculate essential characters
	calcEssential := func(charCount map[int]int, totalUsers int) []types.EssentialCharacter {
		if totalUsers == 0 {
			return nil
		}

		threshold := float64(totalUsers) * 0.7
		type charUsage struct {
			studentID int
			count     int
		}
		var usages []charUsage
		for sid, count := range charCount {
			if float64(count) >= threshold {
				usages = append(usages, charUsage{sid, count})
			}
		}

		sort.Slice(usages, func(i, j int) bool {
			return usages[i].count > usages[j].count
		})

		var result []types.EssentialCharacter
		for _, u := range usages {
			ratio := float64(u.count) / float64(totalUsers)
			result = append(result, types.EssentialCharacter{
				StudentID: u.studentID,
				Ratio:     math.Round(ratio*1000) / 1000,
			})
		}
		return result
	}

	torment = calcEssential(tormentCharCount, tormentUsers)
	lunatic = calcEssential(lunaticCharCount, lunaticUsers)

	return torment, lunatic
}

// GetMinUEUsers returns users who cleared with minimum unique equipment usage
// Sorted by UE count ascending, then party count ascending
// Returns separate results for Torment and Lunatic difficulties
func GetMinUEUsers(partyData *types.BATormentPartyData) (torment *types.MinUEUser, lunatic *types.MinUEUser) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	// Check if this is INSANE difficulty (top score < TormentMinScore)
	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	type userUEData struct {
		rank       int
		score      int
		ueCount    int
		partyCount int
		partyData  [][6]int
	}

	var tormentUsers, lunaticUsers []userUEData

	for _, party := range partyData.PartyDetail {
		// Only consider platinum users
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		// Count UE usage (excluding assists)
		ueCount := 0
		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				// Skip assists (last digit is 1)
				if member%10 == 1 {
					continue
				}
				// Check weapon star (7th digit): (member % 100) / 10
				weaponStar := (member % 100) / 10
				if weaponStar > 0 {
					ueCount++
				}
			}
		}

		userData := userUEData{
			rank:       party.Rank,
			score:      party.Score,
			ueCount:    ueCount,
			partyCount: len(party.PartyData),
			partyData:  party.PartyData,
		}

		// Classify by difficulty
		if party.Score >= constants.LunaticMinScore {
			lunaticUsers = append(lunaticUsers, userData)
		} else if isInsane || party.Score >= constants.TormentMinScore {
			// INSANE: all users go to torment
			tormentUsers = append(tormentUsers, userData)
		}
	}

	// Sort function: UE count ascending, then party count ascending, then rank ascending
	sortFunc := func(users []userUEData) {
		sort.Slice(users, func(i, j int) bool {
			if users[i].ueCount != users[j].ueCount {
				return users[i].ueCount < users[j].ueCount
			}
			if users[i].partyCount != users[j].partyCount {
				return users[i].partyCount < users[j].partyCount
			}
			return users[i].rank < users[j].rank
		})
	}

	if len(tormentUsers) > 0 {
		sortFunc(tormentUsers)
		torment = &types.MinUEUser{
			Rank:      tormentUsers[0].rank,
			Score:     tormentUsers[0].score,
			UECount:   tormentUsers[0].ueCount,
			PartyData: tormentUsers[0].partyData,
		}
	}

	if len(lunaticUsers) > 0 {
		sortFunc(lunaticUsers)
		lunatic = &types.MinUEUser{
			Rank:      lunaticUsers[0].rank,
			Score:     lunaticUsers[0].score,
			UECount:   lunaticUsers[0].ueCount,
			PartyData: lunaticUsers[0].partyData,
		}
	}

	return torment, lunatic
}

// GetMaxPartyUsers returns users who cleared with maximum party count
// Returns separate results for Torment and Lunatic difficulties
func GetMaxPartyUsers(partyData *types.BATormentPartyData) (torment *types.MaxPartyUser, lunatic *types.MaxPartyUser) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	// Check if this is INSANE difficulty (top score < TormentMinScore)
	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	var tormentMaxCount, lunaticMaxCount int

	for _, party := range partyData.PartyDetail {
		// Only consider platinum users
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		partyCount := len(party.PartyData)

		// Classify by difficulty and track max
		if party.Score >= constants.LunaticMinScore {
			if partyCount > lunaticMaxCount {
				lunaticMaxCount = partyCount
				lunatic = &types.MaxPartyUser{
					Rank:      party.Rank,
					Score:     party.Score,
					PartyData: party.PartyData,
				}
			}
		} else if isInsane || party.Score >= constants.TormentMinScore {
			// INSANE: all users go to torment
			if partyCount > tormentMaxCount {
				tormentMaxCount = partyCount
				torment = &types.MaxPartyUser{
					Rank:      party.Rank,
					Score:     party.Score,
					PartyData: party.PartyData,
				}
			}
		}
	}

	return torment, lunatic
}

// GetHighImpactCharacters returns top 3 characters with the biggest rank gap when missing
// It finds characters used by top 100 players of each difficulty and calculates the gap between
// best rank and the best rank achieved without that character
// For 100% usage characters, it falls back to lower difficulty or uses -1 if unavailable
// Returns separate results for Torment and Lunatic difficulties
func GetHighImpactCharacters(partyData *types.BATormentPartyData) (torment []types.HighImpactCharacter, lunatic []types.HighImpactCharacter) {
	if len(partyData.PartyDetail) == 0 {
		return nil, nil
	}

	// Check if this is INSANE difficulty (top score < TormentMinScore)
	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	// Separate data by difficulty
	type partyInfo struct {
		rank      int
		usedChars map[int]bool
	}

	var tormentParties, lunaticParties []partyInfo

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		usedChars := make(map[int]bool)
		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}
				// Include both own characters and assists
				usedChars[logic.GetStudentID(member)] = true
			}
		}

		info := partyInfo{rank: party.Rank, usedChars: usedChars}

		if party.Score >= constants.LunaticMinScore {
			lunaticParties = append(lunaticParties, info)
		} else if isInsane || party.Score >= constants.TormentMinScore {
			tormentParties = append(tormentParties, info)
		}
	}

	// Helper function to find best rank (lowest) without a character in given parties
	// Returns 0 if no one cleared without the character
	findBestRankWithout := func(parties []partyInfo, charID int) int {
		bestRank := 0
		for _, p := range parties {
			if !p.usedChars[charID] {
				if bestRank == 0 || p.rank < bestRank {
					bestRank = p.rank
				}
			}
		}
		return bestRank
	}

	// Helper function to calculate high impact characters for a difficulty
	// fallbackParties is used when 100% usage (no one cleared without the character)
	calcHighImpact := func(parties []partyInfo, fallbackParties []partyInfo, top100Limit int) []types.HighImpactCharacter {
		if len(parties) == 0 {
			return nil
		}

		topRank := parties[0].rank

		// Collect characters used by top 100 of this difficulty
		top100Chars := make(map[int]bool)
		for i, p := range parties {
			if i >= top100Limit {
				break
			}
			for charID := range p.usedChars {
				top100Chars[charID] = true
			}
		}

		// Calculate gaps
		type charGap struct {
			studentID       int
			rankGap         int
			withoutBestRank int
		}
		var gaps []charGap

		for charID := range top100Chars {
			withoutBestRank := findBestRankWithout(parties, charID)

			// If 100% usage in this difficulty, try fallback (lower difficulty)
			if withoutBestRank == 0 && len(fallbackParties) > 0 {
				withoutBestRank = findBestRankWithout(fallbackParties, charID)
			}

			// Calculate gap (-1 if still no data)
			var rankGap int
			if withoutBestRank > 0 {
				rankGap = withoutBestRank - topRank
			} else {
				rankGap = -1 // 100% usage, no comparison available
			}

			gaps = append(gaps, charGap{charID, rankGap, withoutBestRank})
		}

		// Sort: -1 (100% usage) first, then by gap descending
		sort.Slice(gaps, func(i, j int) bool {
			if gaps[i].rankGap == -1 && gaps[j].rankGap != -1 {
				return true
			}
			if gaps[i].rankGap != -1 && gaps[j].rankGap == -1 {
				return false
			}
			return gaps[i].rankGap > gaps[j].rankGap
		})

		var result []types.HighImpactCharacter
		for i := 0; i < 3 && i < len(gaps); i++ {
			result = append(result, types.HighImpactCharacter{
				StudentID:       gaps[i].studentID,
				RankGap:         gaps[i].rankGap,
				TopRank:         topRank,
				WithoutBestRank: gaps[i].withoutBestRank,
			})
		}
		return result
	}

	// For torment, no fallback (it's the lowest difficulty)
	torment = calcHighImpact(tormentParties, nil, 100)
	// For lunatic, fallback to torment
	lunatic = calcHighImpact(lunaticParties, tormentParties, 100)

	return torment, lunatic
}

// GetPlatinumCuts retrieves score cutoffs at specific ranks (2000, 4000, ..., 20000)
// For grand assault, it uses combined scores from all armor types
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

	// Target ranks: 100, 200, 500, 1000, 2000, 4000, 6000, ..., 20000
	ranks := []int{100, 200, 500, 1000, 2000, 4000, 6000, 8000, 10000, 12000, 14000, 16000, 18000, 20000}

	var querySQL string
	if IsGrandAssault(contentID) {
		// For grand assault, first check which point columns exist
		existingColumns, err := getExistingPointColumns(db)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing columns: %w", err)
		}
		if len(existingColumns) == 0 {
			return nil, fmt.Errorf("no point columns found for elimination raid")
		}
		querySQL = GetPartialPlatinumCutSQL(ranks, existingColumns)
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

// GetPartPlatinumCutsFromPartyData extracts platinum cuts from partyData for grand assault (대결전)
// This uses the partyData's rank order directly instead of querying DuckDB
func GetPartPlatinumCutsFromPartyData(partyData *types.BATormentPartyData) []types.PlatinumCut {
	if partyData == nil || len(partyData.PartyDetail) == 0 {
		return nil
	}

	// Target ranks: 100, 200, 500, 1000, 2000, 4000, 6000, ..., 20000
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
