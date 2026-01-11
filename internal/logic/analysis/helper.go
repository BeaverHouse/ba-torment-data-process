package analysis

import (
	"sort"
	"strings"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/types"
)

// ExtractGroupID extracts group ID from content_id (e.g., "3S26-1" -> "3S26")
func ExtractGroupID(contentID string) string {
	if idx := strings.Index(contentID, "-"); idx != -1 {
		return contentID[:idx]
	}
	return contentID
}

// ParseStudentDetailID extracts studentID, star, weaponStar, isAssist from detailID
// Format: {studentID(5)}{star(1)}{weaponStar(1)}{isAssist(1)}
func ParseStudentDetailID(detailID int) (studentID int, star int, weaponStar int, isAssist bool) {
	studentID = detailID / 1000
	star = (detailID % 1000) / 100
	weaponStar = (detailID % 100) / 10
	isAssist = detailID%10 == 1
	return
}

// IsStriker checks if studentID is a striker (1xxxx)
func IsStriker(studentID int) bool {
	return studentID/10000 == 1
}

// IsSpecial checks if studentID is a special (2xxxx)
func IsSpecial(studentID int) bool {
	return studentID/10000 == 2
}

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
					studentID := member / 1000
					result[studentID] = true
				}
			}
		}
	}

	return result
}

// getPlatinumUserCount returns the count of platinum users (rank <= 20000)
func getPlatinumUserCount(partyData *types.BATormentPartyData) int {
	count := 0
	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}
		count++
	}
	return count
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
