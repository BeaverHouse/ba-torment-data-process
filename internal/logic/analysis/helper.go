package analysis

import (
	"sort"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
)

// collectAllStudentIDs collects all unique studentIDs from party data
func collectAllStudentIDs(partyDataMap map[string]*types.BATormentPartyData) map[int]bool {
	result := make(map[int]bool)

	for _, partyData := range partyDataMap {
		for _, party := range partyData.PartyDetail {
			if party.Rank > constants.PlatinumRankLimit {
				break
			}
			for _, members := range party.PartyData {
				for _, member := range members {
					if member == 0 {
						continue
					}
					result[id.GetStudentID(member)] = true
				}
			}
		}
	}

	return result
}

// intKV is a key-value pair for sorting map[int]int
type intKV struct {
	Key   int
	Value int
}

// sortMapDescending sorts a map[int]int by value descending, then key ascending
func sortMapDescending(m map[int]int) []intKV {
	sorted := make([]intKV, 0, len(m))
	for k, v := range m {
		sorted = append(sorted, intKV{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Value != sorted[j].Value {
			return sorted[i].Value > sorted[j].Value
		}
		return sorted[i].Key < sorted[j].Key
	})
	return sorted
}

// getTopN extracts top N characters by usage count
func getTopN(countMap map[int]int, n int) []types.CharacterUsage {
	sorted := sortMapDescending(countMap)

	var result []types.CharacterUsage
	for i := 0; i < n && i < len(sorted); i++ {
		result = append(result, types.CharacterUsage{
			StudentID:  sorted[i].Key,
			UsageCount: sorted[i].Value,
		})
	}
	return result
}
