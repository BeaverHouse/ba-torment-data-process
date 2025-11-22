package filter

import (
	"context"
	"log"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"
)

// createFilterFromPartyTeams creates a filter from a list of party teams for summary data
func createFilterFromPartyTeams(partyTeams [][6]int) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	for _, partyTeam := range partyTeams {
		for _, studentDetailID := range partyTeam {
			if studentDetailID != 0 {
				logic.UpdateSummaryFilters(filters, assistFilters, studentDetailID)
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}

// createVideoFilterFromPartyTeams creates a filter from a list of party teams for video data
func createVideoFilterFromPartyTeams(partyTeams [][6]int) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	for _, partyTeam := range partyTeams {
		for _, studentDetailID := range partyTeam {
			if studentDetailID != 0 {
				logic.UpdatePartyFilters(filters, assistFilters, studentDetailID)
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

	// Filter parties with score >= LunaticMinScore
	for _, party := range partyData.PartyDetail {
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

// CreateVideoFilter creates a video filter from verified YouTube analysis data in the database
func CreateVideoFilter(raidID string) *types.BATormentFilter {
	// Connect to database
	pool := postgres.InitFromEnv()
	defer pool.Close()

	ctx := context.Background()
	queries := postgres.New(pool)

	// Get verified YouTube analysis from database
	analysisRows, err := queries.GetVerifiedYoutubeAnalysisByRaidID(ctx, raidID)
	if err != nil {
		log.Printf("Failed to query YouTube analysis for %s: %v", raidID, err)
		return nil
	}

	if len(analysisRows) == 0 {
		log.Printf("No verified YouTube analysis found for %s", raidID)
		return nil
	}

	var partyTeams [][6]int

	// Parse analysis results and extract party data
	for _, row := range analysisRows {
		partyTeams = append(partyTeams, row.AnalysisResult.PartyData...)
	}

	log.Printf("Created video filter for %s with %d party teams from %d verified videos",
		raidID, len(partyTeams), len(analysisRows))

	return createVideoFilterFromPartyTeams(partyTeams)
}
