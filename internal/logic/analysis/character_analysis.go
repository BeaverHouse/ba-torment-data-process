package analysis

import (
	"fmt"
	"sort"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/types"
)

// RunCharacterAnalyses runs analysis for all characters
// sortedContentIDs provides the order for usageHistory (sorted by start_date)
func RunCharacterAnalyses(partyDataMap map[string]*types.BATormentPartyData, sortedContentIDs []string) []types.CharacterAnalysisResult {
	allStudentIDs := collectAllStudentIDs(partyDataMap)

	var results []types.CharacterAnalysisResult
	for studentID := range allStudentIDs {
		result := AnalyzeCharacter(studentID, partyDataMap, sortedContentIDs)
		results = append(results, result)
	}

	// Sort by studentID
	sort.Slice(results, func(i, j int) bool {
		return results[i].StudentID < results[j].StudentID
	})

	return results
}

// AnalyzeCharacter analyzes a single character across all raids
// sortedContentIDs provides the order for usageHistory (sorted by start_date)
func AnalyzeCharacter(studentID int, partyDataMap map[string]*types.BATormentPartyData, sortedContentIDs []string) types.CharacterAnalysisResult {
	var usageHistory []types.RaidUsage
	var starDistribution []types.RaidStarDistribution

	totalAsAssist := 0
	totalAsOwn := 0

	// Synergy calculation: count of co-appearances with other characters
	coUsageCount := make(map[int]int)
	totalAppearances := 0

	for _, raidID := range sortedContentIDs {
		partyData, ok := partyDataMap[raidID]
		if !ok {
			continue
		}
		raidUsage, raidStar, asAssist, asOwn, coChars, appearances := analyzeCharacterInRaid(studentID, partyData)

		usageHistory = append(usageHistory, types.RaidUsage{
			RaidID:           raidID,
			UserCount:        raidUsage.UserCount,
			LunaticUserCount: raidUsage.LunaticUserCount,
		})

		// Only include star distribution if own usage (excluding assist) >= 200 (1%)
		if asOwn >= 200 {
			starDistribution = append(starDistribution, types.RaidStarDistribution{
				RaidID:       raidID,
				Distribution: raidStar,
			})
		}

		totalAsAssist += asAssist
		totalAsOwn += asOwn

		for coCharID, count := range coChars {
			coUsageCount[coCharID] += count
		}
		totalAppearances += appearances
	}

	// Calculate synergy (top 3, >= 5%)
	topSynergyChars := calculateTopSynergy(coUsageCount, totalAppearances, 3)

	totalCount := totalAsAssist + totalAsOwn
	assistRatio := 0.0
	if totalCount > 0 {
		assistRatio = float64(totalAsAssist) / float64(totalCount)
	}

	return types.CharacterAnalysisResult{
		StudentID:        studentID,
		UsageHistory:     usageHistory,
		StarDistribution: starDistribution,
		AssistStats: types.AssistUsageStats{
			AsAssistCount: totalAsAssist,
			AsOwnCount:    totalAsOwn,
			TotalCount:    totalCount,
			AssistRatio:   assistRatio,
		},
		TopSynergyChars: topSynergyChars,
	}
}

// analyzeCharacterInRaid analyzes a character in a single raid
func analyzeCharacterInRaid(studentID int, partyData *types.BATormentPartyData) (
	usage types.RaidUsage,
	starDist map[string]int,
	asAssist int,
	asOwn int,
	coChars map[int]int,
	appearances int,
) {
	starDist = make(map[string]int)
	coChars = make(map[int]int)
	lunaticUserCount := 0

	for _, party := range partyData.PartyDetail {
		if party.Rank > PlatinumRankLimit {
			break
		}

		isLunatic := party.Score >= constants.LunaticMinScore
		foundInParty := false
		var partyMembers []int

		for _, members := range party.PartyData {
			for _, member := range members {
				if member == 0 {
					continue
				}

				memberStudentID, star, weaponStar, isAssist := ParseStudentDetailID(member)

				// Collect party members for synergy
				partyMembers = append(partyMembers, memberStudentID)

				if memberStudentID == studentID {
					foundInParty = true

					if isAssist {
						asAssist++
						if isLunatic {
							lunaticUserCount++
						}
					} else {
						asOwn++
						if isLunatic {
							lunaticUserCount++
						}
						// Star distribution (exclude assist)
						key := fmt.Sprintf("%d%d", star, weaponStar)
						starDist[key]++
					}
				}
			}
		}

		if foundInParty {
			appearances++
			// Count co-usage characters
			for _, memberID := range partyMembers {
				if memberID != studentID {
					coChars[memberID]++
				}
			}
		}
	}

	usage = types.RaidUsage{
		UserCount:        asAssist + asOwn,
		LunaticUserCount: lunaticUserCount,
	}

	return
}

// calculateTopSynergy calculates top N synergy characters (>= 5% only)
func calculateTopSynergy(coUsageCount map[int]int, totalAppearances int, n int) []types.CharacterSynergy {
	if totalAppearances == 0 {
		return nil
	}

	type kv struct {
		Key   int
		Value int
	}

	var candidates []kv
	for k, v := range coUsageCount {
		rate := float64(v) / float64(totalAppearances)
		if rate >= 0.05 { // 5% threshold
			candidates = append(candidates, kv{k, v})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Value != candidates[j].Value {
			return candidates[i].Value > candidates[j].Value
		}
		return candidates[i].Key < candidates[j].Key
	})

	var result []types.CharacterSynergy
	for i := 0; i < n && i < len(candidates); i++ {
		result = append(result, types.CharacterSynergy{
			StudentID:    candidates[i].Key,
			CoUsageRate:  float64(candidates[i].Value) / float64(totalAppearances),
			CoUsageCount: candidates[i].Value,
		})
	}

	return result
}
