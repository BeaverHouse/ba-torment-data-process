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

	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/andybalholm/brotli"
	_ "github.com/marcboeker/go-duckdb"
)

func downloadDuckDB(log logger.Logger, dateString string) error {
	baseURL := env.GetEnv("BATORMENT_DUCKDB_REMOTE_URL", "")
	if baseURL == "" {
		return errorhandle.ErrConfigMissing("BATORMENT_DUCKDB_REMOTE_URL")
	}

	url := fmt.Sprintf("%s/v2/JP/%s.db", baseURL, dateString)
	fileName := fmt.Sprintf("%s.db", dateString)

	log.Info("Downloading DuckDB", logger.Field{Key: "url", Value: url})

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}

	req.Header.Set("Accept-Encoding", "br, gzip, deflate")
	req.Header.Set("User-Agent", "ba-torment-data-process/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return constants.ErrDataFetch(fmt.Sprintf("duckdb from %s", url), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return constants.ErrUpstreamBadStatus(fmt.Sprintf("HTTP %d from %s", resp.StatusCode, url))
	}

	contentEncoding := resp.Header.Get("Content-Encoding")
	log.Info("Response metadata",
		logger.Field{Key: "contentType", Value: resp.Header.Get("Content-Type")},
		logger.Field{Key: "contentLength", Value: resp.ContentLength},
		logger.Field{Key: "contentEncoding", Value: contentEncoding})

	out, err := os.Create(fileName)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	defer out.Close()

	var reader io.Reader = resp.Body
	if contentEncoding == "br" {
		log.Info("Decompressing Brotli-encoded file...")
		reader = brotli.NewReader(resp.Body)
	}

	written, err := io.Copy(out, reader)
	if err != nil {
		os.Remove(fileName)
		return errorhandle.ErrInternal(err)
	}

	log.Info("Downloaded file", logger.Field{Key: "bytes", Value: written}, logger.Field{Key: "mb", Value: fmt.Sprintf("%.2f", float64(written)/(1024*1024))}, logger.Field{Key: "file", Value: fileName})

	if written < 10240 {
		os.Remove(fileName)
		return constants.ErrDuckDBTooSmall(written)
	}

	return nil
}

func ParseDuckDB(log logger.Logger, contentID string, startDate time.Time) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		log.Info("DuckDB file not found, attempting to download", logger.Field{Key: "file", Value: dbFileName})
		if err := downloadDuckDB(log, dateString); err != nil {
			log.Info("Failed to download DuckDB file, skipping this raid", logger.Field{Key: "file", Value: dbFileName}, logger.Field{Key: "error", Value: err})
			return nil, nil, constants.ErrDuckDBUnavailable(err)
		}
		log.Info("Successfully downloaded DuckDB file", logger.Field{Key: "file", Value: dbFileName})
	}

	db, err := sql.Open("duckdb", dbFileName)
	if err != nil {
		return nil, nil, errorhandle.ErrDBOperation("open duckdb", err)
	}
	defer db.Close()

	_, category := id.SplitSeasonString(contentID)

	armorType := constants.ArmorTypeMapping[category]

	partyData, filterResult, err := processArmorType(db, armorType)
	if err != nil {
		return nil, nil, err
	}

	if len(partyData.PartyDetail) == 0 {
		return nil, nil, constants.ErrArmorTypeNoDetails(armorType)
	}

	removeFraudUsers(log, contentID, partyData)

	return partyData, filterResult, nil
}

func removeFraudUsers(log logger.Logger, contentID string, partyData *types.BATormentPartyData) {
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
				log.Warn("Found fraud user at rank 13622 with Mika star1 UE3 in S80-0")
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

	log.Info("Removed fraud user and adjusted ranks", logger.Field{Key: "adjustedRanks", Value: len(partyData.PartyDetail) - fraudIndex})
}

func processArmorType(db *sql.DB, armorType string) (*types.BATormentPartyData, *types.BATormentFilter, error) {
	hasSkillOrder, err := hasMulliganColumn(db, armorType)
	if err != nil {
		return nil, nil, errorhandle.ErrDBOperation("check starting skill order column", err)
	}

	completeRunsSQL := getCompleteRunIDAndScoreSQL(armorType)
	rows, err := db.Query(completeRunsSQL)
	if err != nil {
		return nil, nil, errorhandle.ErrDBOperation("query complete runs", err)
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
			return nil, nil, errorhandle.ErrDBOperation("scan complete run", err)
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

		partyData, skillOrders, err := getPartiesByCompleteRunID(db, armorType, run.completeRunID, hasSkillOrder)
		if err != nil {
			return nil, nil, err
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
			Rank:        rank + 1,
			Score:       run.score,
			PartyData:   partyData,
			SkillOrders: skillOrders,
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

func getPartiesByCompleteRunID(db *sql.DB, armorType string, completeRunID int, hasSkillOrder bool) ([][6]int, [][6]int, error) {
	runIDsSQL := getRunIDsByCompleteRunIDSQL(armorType, completeRunID)
	rows, err := db.Query(runIDsSQL)
	if err != nil {
		return nil, nil, errorhandle.ErrDBOperation("query run IDs", err)
	}
	defer rows.Close()

	var parties [][6]int
	var skillOrders [][6]int
	hasAnySkillOrder := false

	for rows.Next() {
		var runID int
		if err := rows.Scan(&runID); err != nil {
			return nil, nil, errorhandle.ErrDBOperation("scan run ID", err)
		}

		party, skillOrder, hasRunSkillOrder, err := getPartyByRunID(db, armorType, runID, hasSkillOrder)
		if err != nil {
			return nil, nil, err
		}

		parties = append(parties, party)
		skillOrders = append(skillOrders, skillOrder)
		hasAnySkillOrder = hasAnySkillOrder || hasRunSkillOrder
	}

	if !hasAnySkillOrder {
		skillOrders = nil
	}

	return parties, skillOrders, nil
}

func getPartyByRunID(db *sql.DB, armorType string, runID int, hasSkillOrder bool) ([6]int, [6]int, bool, error) {
	partySQL := getPartyInfoByRunIDSQL(armorType, runID, hasSkillOrder)
	rows, err := db.Query(partySQL)
	if err != nil {
		return [6]int{}, [6]int{}, false, errorhandle.ErrDBOperation("query party info", err)
	}
	defer rows.Close()

	var partyMembers [6]int
	var skillOrder [6]int
	hasAnySkillOrder := false
	specialIndex := 0

	for rows.Next() {
		var sid, level, slot, mulligan int
		var build string
		var assist bool

		if err := rows.Scan(&sid, &build, &level, &slot, &assist, &mulligan); err != nil {
			return [6]int{}, [6]int{}, false, errorhandle.ErrDBOperation("scan party info", err)
		}
		if mulligan < 0 || mulligan > 5 {
			return [6]int{}, [6]int{}, false, constants.ErrInvalidSkillOrder(mulligan)
		}

		weaponValue, exists := constants.WeaponStarMapping[build]
		if !exists {
			return [6]int{}, [6]int{}, false, constants.ErrUnknownBuildValue(build)
		}

		star := weaponValue / 10
		weaponStar := weaponValue % 10

		studentDetailID := id.ComposeStudentDetailID(sid, star, weaponStar, assist)

		memberIndex := slot
		if slot > 3 {
			if specialIndex >= 2 {
				continue
			}
			memberIndex = 4 + specialIndex
			specialIndex++
		}
		partyMembers[memberIndex] = studentDetailID
		skillOrder[memberIndex] = mulligan
		hasAnySkillOrder = hasAnySkillOrder || mulligan > 0
	}

	return partyMembers, skillOrder, hasAnySkillOrder, nil
}

// IsGrandAssault checks if the contentID represents a grand assault
func IsGrandAssault(contentID string) bool {
	return strings.HasPrefix(contentID, "3S")
}

// GetPlatinumCuts retrieves score cutoffs at specific ranks
func GetPlatinumCuts(log logger.Logger, contentID string, startDate time.Time) ([]types.PlatinumCut, error) {
	dateString := startDate.Format("20060102")
	dbFileName := fmt.Sprintf("%s.db", dateString)

	if _, err := os.Stat(dbFileName); os.IsNotExist(err) {
		log.Warn("DuckDB file not found for platinum cuts", logger.Field{Key: "file", Value: dbFileName})
		return nil, constants.ErrDuckDBUnavailable(err)
	}

	db, err := sql.Open("duckdb", dbFileName)
	if err != nil {
		return nil, errorhandle.ErrDBOperation("open duckdb", err)
	}
	defer db.Close()

	ranks := []int{100, 200, 500, 1000, 2000, 4000, 6000, 8000, 10000, 12000, 14000, 16000, 18000, 20000}

	var querySQL string
	if IsGrandAssault(contentID) {
		existingColumns, err := getExistingPointColumns(db)
		if err != nil {
			return nil, errorhandle.ErrDBOperation("get existing point columns", err)
		}
		if len(existingColumns) == 0 {
			return nil, constants.ErrNoPointColumns()
		}
		querySQL = getPartialPlatinumCutSQL(ranks, existingColumns)
	} else {
		querySQL = getPlatinumCutSQL(ranks)
	}

	rows, err := db.Query(querySQL)
	if err != nil {
		return nil, errorhandle.ErrDBOperation("query platinum cuts", err)
	}
	defer rows.Close()

	var cuts []types.PlatinumCut
	for rows.Next() {
		var rank, score int
		if err := rows.Scan(&rank, &score); err != nil {
			return nil, errorhandle.ErrDBOperation("scan platinum cut", err)
		}
		cuts = append(cuts, types.PlatinumCut{
			Rank:  rank,
			Score: score,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, errorhandle.ErrDBOperation("iterate platinum cuts", err)
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
