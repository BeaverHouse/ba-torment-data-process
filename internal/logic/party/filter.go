package party

import (
	"context"
	"strconv"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/logger"
)

// === Filter update helpers ===

func updateFilter(filters map[string](map[string]int), studentDetailID int) {
	studentIDString := strconv.Itoa(id.GetStudentID(studentDetailID))
	starWeaponString := strconv.Itoa(id.GetStarWeapon(studentDetailID))

	if _, exists := filters[studentIDString]; !exists {
		filters[studentIDString] = make(map[string]int)
	}

	filters[studentIDString][starWeaponString]++
}

// updatePartyFilters updates assist filters if the student is an assist,
// and updates normal filters if the student is not an assist.
func updatePartyFilters(filters map[string](map[string]int), assistFilters map[string](map[string]int), studentDetailID int) {
	isAssist := id.IsAssist(studentDetailID)

	targetFilters := filters
	if isAssist {
		targetFilters = assistFilters
	}

	updateFilter(targetFilters, studentDetailID)
}

// updateSummaryFilters always updates normal filters,
// and updates assist filters if the student is an assist.
func updateSummaryFilters(filters map[string](map[string]int), assistFilters map[string](map[string]int), studentDetailID int) {
	isAssist := id.IsAssist(studentDetailID)

	updateFilter(filters, studentDetailID)
	if isAssist {
		updateFilter(assistFilters, studentDetailID)
	}
}

// === Party Filters ===

func createFilterFromPartyTeams(partyTeams [][6]int) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	for _, partyTeam := range partyTeams {
		for _, studentDetailID := range partyTeam {
			if studentDetailID != 0 {
				updateSummaryFilters(filters, assistFilters, studentDetailID)
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}

func CreateLunaticFilter(partyData *types.BATormentPartyData) *types.BATormentFilter {
	var partyTeams [][6]int

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			continue
		}
		if party.Score >= constants.LunaticMinScore {
			partyTeams = append(partyTeams, party.PartyData...)
		}
	}

	return createFilterFromPartyTeams(partyTeams)
}

func CreateNonLunaticFilter(partyData *types.BATormentPartyData) *types.BATormentFilter {
	var partyTeams [][6]int

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			continue
		}
		maxScore := constants.LunaticMinScore
		minScore := constants.TormentMinScore
		if isInsane {
			maxScore = constants.TormentMinScore
			minScore = 0
		}
		if party.Score >= minScore && party.Score < maxScore {
			partyTeams = append(partyTeams, party.PartyData...)
		}
	}

	return createFilterFromPartyTeams(partyTeams)
}

// === Video Filter ===

func createVideoFilterFromPartyTeams(partyTeams [][6]int) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	for _, partyTeam := range partyTeams {
		for _, studentDetailID := range partyTeam {
			if studentDetailID != 0 {
				updatePartyFilters(filters, assistFilters, studentDetailID)
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}

// CreateVideoFilter creates a video filter from verified YouTube analysis data in the database
func CreateVideoFilter(log logger.Logger, raidID string) *types.BATormentFilter {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()

	ctx := context.Background()
	queries := postgres.New(pool)

	analysisRows, err := queries.GetVerifiedYoutubeAnalysisByRaidID(ctx, raidID)
	if err != nil {
		log.Warn("Failed to query YouTube analysis", logger.Field{Key: "raidID", Value: raidID}, logger.Field{Key: "error", Value: err})
		return nil
	}

	if len(analysisRows) == 0 {
		log.Info("No verified YouTube analysis found", logger.Field{Key: "raidID", Value: raidID})
		return nil
	}

	var partyTeams [][6]int

	for _, row := range analysisRows {
		partyTeams = append(partyTeams, row.AnalysisResult.PartyData...)
	}

	log.Info("Created video filter", logger.Field{Key: "raidID", Value: raidID}, logger.Field{Key: "partyTeams", Value: len(partyTeams)}, logger.Field{Key: "videos", Value: len(analysisRows)})

	return createVideoFilterFromPartyTeams(partyTeams)
}
