package filter

import (
	"ba-torment-data-process/app/types"
	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
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

	// Filter parties with score < LunaticMinScore
	for _, party := range partyData.PartyDetail {
		if party.Score < constants.LunaticMinScore {
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
