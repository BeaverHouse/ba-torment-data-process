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

func CreateLunaticFilter(partyData *types.BATormentPartyData, _ *types.BATormentFilter) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	// Filter parties with score >= LunaticMinScore
	for _, party := range partyData.PartyDetail {
		if party.Score >= constants.LunaticMinScore {
			// Process each party's characters
			for _, partyTeam := range party.PartyData {
				for _, studentDetailID := range partyTeam {
					if studentDetailID != 0 {
						logic.UpdateSummaryFilters(filters, assistFilters, studentDetailID)
					}
				}
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}

func CreateNonLunaticFilter(partyData *types.BATormentPartyData, _ *types.BATormentFilter) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	for _, party := range partyData.PartyDetail {
		maxScore := constants.LunaticMinScore
		minScore := constants.TormentMinScore
		if isInsane {
			maxScore = constants.TormentMinScore
			minScore = 0
		}
		if party.Score >= minScore && party.Score < maxScore {
			// Process each party's characters
			for _, partyTeam := range party.PartyData {
				for _, studentDetailID := range partyTeam {
					if studentDetailID != 0 {
						logic.UpdateSummaryFilters(filters, assistFilters, studentDetailID)
					}
				}
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
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

	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	// Filter parties with score >= LunaticMinScore
	for _, analysis := range data.Data.Data {
		for _, partyTeam := range analysis.PartyData {
			for _, studentDetailID := range partyTeam {
				if studentDetailID != 0 {
					logic.UpdateSummaryFilters(filters, assistFilters, studentDetailID)
				}
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}
