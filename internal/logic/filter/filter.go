package filter

import (
	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

func CreateVideoFilter(raidID string) *types.BATormentFilter {
	url := "https://api.tinyclover.com/ba-analyzer/v1/video/analysis?raid_id=" + raidID + "&page=1&limit=1000"

	serviceToken := logic.GetEnv("BA_ANALYZER_SERVICE_API_KEY", "")

	req, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return nil
	}

	req.Header.Set("X-Access-Token", serviceToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}

	fmt.Println(string(body))

	var data types.APIResponse[types.VideoAnalysisListResponse]
	if err := json.Unmarshal(body, &data); err != nil {
		return nil
	}

	var partyTeams [][6]int

	for _, analysis := range data.Data.Data {
		if !analysis.IsVerified {
			continue
		}
		// Convert [][]int to [][6]int
		for _, party := range analysis.PartyData {
			var team [6]int
			copy(team[:], party)
			partyTeams = append(partyTeams, team)
		}
	}

	return createVideoFilterFromPartyTeams(partyTeams)
}
