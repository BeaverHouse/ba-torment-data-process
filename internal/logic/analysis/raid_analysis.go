package analysis

import (
	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"
)

// RunRaidAnalyses runs analysis for all raids
// sortedContentIDs provides the order (sorted by start_date)
func RunRaidAnalyses(partyDataMap map[string]*types.BATormentPartyData, sortedContentIDs []string) []types.RaidAnalysisResult {
	var results []types.RaidAnalysisResult

	for _, raidID := range sortedContentIDs {
		partyData, ok := partyDataMap[raidID]
		if !ok {
			continue
		}
		result := AnalyzeRaid(raidID, partyData)
		results = append(results, result)
	}

	return results
}

// AnalyzeRaid analyzes a single raid
func AnalyzeRaid(raidID string, partyData *types.BATormentPartyData) types.RaidAnalysisResult {
	strikerCount := make(map[int]int)
	specialCount := make(map[int]int)
	assistCount := make(map[int]int)
	lunaticClearCount := 0

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}

		if party.Score >= constants.LunaticMinScore {
			lunaticClearCount++
		}

		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}

				studentID, _, _, isAssist := logic.ParseStudentDetailID(member)

				if IsStriker(studentID) {
					strikerCount[studentID]++
				} else if IsSpecial(studentID) {
					specialCount[studentID]++
				}

				if isAssist {
					assistCount[studentID]++
				}
			}
		}
	}

	return types.RaidAnalysisResult{
		RaidID:            raidID,
		TopStrikers:       getTopN(strikerCount, 5),
		TopSpecials:       getTopN(specialCount, 5),
		TopAssists:        getTopN(assistCount, 3),
		LunaticClearCount: lunaticClearCount,
	}
}
