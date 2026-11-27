package analysis

import (
	"fmt"
	"slices"
	"sort"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
)

// characterTotals holds one character's figures summed over the raids added so far.
type characterTotals struct {
	usage        map[string]types.RaidUsage // by raid ID
	star         map[string]map[string]int  // by group ID (e.g., "3S26-1", "3S26-3" -> "3S26")
	asOwnByGroup map[string]int
	asAssist     int
	asOwn        int
	coUsage      map[int]int // co-appearances with other characters
	appearances  int
}

// characterAnalysis sums every character's figures raid by raid, so a raid's
// party data can be dropped as soon as it has been added.
type characterAnalysis struct {
	raidIDs  []string // in the order added
	groupIDs []string // in order of first appearance
	totals   map[int]*characterTotals
}

// addRaid adds one raid's parties to every character in them.
func (a *characterAnalysis) addRaid(raidID string, partyData *types.BATormentPartyData) {
	a.raidIDs = append(a.raidIDs, raidID)
	groupID := id.ExtractGroupID(raidID)
	if !slices.Contains(a.groupIDs, groupID) {
		a.groupIDs = append(a.groupIDs, groupID)
	}

	for _, party := range partyData.PartyDetail {
		if party.Rank > constants.PlatinumRankLimit {
			break
		}
		a.addParty(raidID, groupID, party)
	}
}

func (a *characterAnalysis) addParty(raidID, groupID string, party types.BATormentPartyDetail) {
	isLunatic := party.Score >= constants.LunaticMinScore
	inParty := make(map[int]bool)

	for _, squad := range party.PartyData {
		var squadMembers []int

		for _, member := range squad {
			if member == 0 {
				continue
			}

			studentID, star, weaponStar, isAssist := id.ParseStudentDetailID(member)
			squadMembers = append(squadMembers, studentID)
			inParty[studentID] = true

			if a.totals[studentID] == nil {
				a.totals[studentID] = &characterTotals{
					usage:        make(map[string]types.RaidUsage),
					star:         make(map[string]map[string]int),
					asOwnByGroup: make(map[string]int),
					coUsage:      make(map[int]int),
				}
			}
			a.totals[studentID].addUse(raidID, groupID, isLunatic, isAssist, fmt.Sprintf("%d%d", star, weaponStar))
		}

		// A character counts its squad mates once per squad, however often it is in it.
		for i, studentID := range squadMembers {
			if slices.Contains(squadMembers[:i], studentID) {
				continue
			}
			for _, memberID := range squadMembers {
				if memberID != studentID {
					a.totals[studentID].coUsage[memberID]++
				}
			}
		}
	}

	for studentID := range inParty {
		a.totals[studentID].appearances++
	}
}

func (t *characterTotals) addUse(raidID, groupID string, isLunatic, isAssist bool, starKey string) {
	usage := t.usage[raidID]
	usage.UserCount++
	if isLunatic {
		usage.LunaticUserCount++
	}
	t.usage[raidID] = usage

	if isAssist {
		t.asAssist++
		return
	}
	t.asOwn++
	t.asOwnByGroup[groupID]++
	// Star distribution (exclude assist)
	if t.star[groupID] == nil {
		t.star[groupID] = make(map[string]int)
	}
	t.star[groupID][starKey]++
}

// results returns every character's analysis, sorted by studentID.
func (a *characterAnalysis) results() []types.CharacterAnalysisResult {
	var results []types.CharacterAnalysisResult

	for studentID, totals := range a.totals {
		var usageHistory []types.RaidUsage
		for _, raidID := range a.raidIDs {
			usage := totals.usage[raidID]
			usage.RaidID = raidID
			usageHistory = append(usageHistory, usage)
		}

		// Get latest star distribution (200+ own usage)
		var starDistribution *types.RaidStarDistribution
		for i := len(a.groupIDs) - 1; i >= 0; i-- {
			groupID := a.groupIDs[i]
			if totals.asOwnByGroup[groupID] >= 200 {
				starDistribution = &types.RaidStarDistribution{
					RaidID:       groupID,
					Distribution: totals.star[groupID],
				}
				break
			}
		}

		totalCount := totals.asAssist + totals.asOwn
		assistRatio := 0.0
		if totalCount > 0 {
			assistRatio = float64(totals.asAssist) / float64(totalCount)
		}

		results = append(results, types.CharacterAnalysisResult{
			StudentID:        studentID,
			UsageHistory:     usageHistory,
			StarDistribution: starDistribution,
			AssistStats: types.AssistUsageStats{
				AsAssistCount: totals.asAssist,
				AsOwnCount:    totals.asOwn,
				TotalCount:    totalCount,
				AssistRatio:   assistRatio,
			},
			// Synergy: top 3, >= 5%
			TopSynergyChars: calculateTopSynergy(totals.coUsage, totals.appearances, 3),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StudentID < results[j].StudentID
	})

	return results
}

// calculateTopSynergy calculates top N synergy characters (>= 5% only)
func calculateTopSynergy(coUsageCount map[int]int, totalAppearances int, n int) []types.CharacterSynergy {
	if totalAppearances == 0 {
		return nil
	}

	// Filter to >= 5% threshold
	filtered := make(map[int]int)
	for k, v := range coUsageCount {
		if float64(v)/float64(totalAppearances) >= 0.05 {
			filtered[k] = v
		}
	}

	sorted := sortMapDescending(filtered)

	var result []types.CharacterSynergy
	for i := 0; i < n && i < len(sorted); i++ {
		result = append(result, types.CharacterSynergy{
			StudentID:    sorted[i].Key,
			CoUsageRate:  float64(sorted[i].Value) / float64(totalAppearances),
			CoUsageCount: sorted[i].Value,
		})
	}
	return result
}
