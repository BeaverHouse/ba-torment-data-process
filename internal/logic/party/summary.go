package party

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
)

func ProcessPartyDataToSummaryData(partyData *types.BATormentPartyData) (*types.BATormentSummaryData, error) {
	var lunaticData, tormentData []types.BATormentPartyDetail

	isInsane := partyData.PartyDetail[0].Score < constants.TormentMinScore

	for _, data := range partyData.PartyDetail {
		if data.Rank > constants.PlatinumRankLimit {
			continue
		}
		if data.Score >= constants.LunaticMinScore {
			lunaticData = append(lunaticData, data)
		} else {
			if isInsane || data.Score >= constants.TormentMinScore {
				tormentData = append(tormentData, data)
			}
		}
	}

	lunaticCount := len(lunaticData)
	tormentCount := len(tormentData)

	result := &types.BATormentSummaryData{
		Torment: processLevelData(tormentData, "torment", lunaticCount, tormentCount),
		Lunatic: processLevelData(lunaticData, "lunatic", lunaticCount, tormentCount),
	}

	return result, nil
}

func processLevelData(data []types.BATormentPartyDetail, level string, lunaticCount, tormentCount int) types.BATormentLevelData {
	result := types.BATormentLevelData{
		ClearCount:  len(data),
		PartyCounts: make(map[string][]int),
		Top5Partys:  make([][]any, 0),
	}

	parties := make([]string, 0)

	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	for _, entry := range data {
		var partyKeys []string
		for i := range len(entry.PartyData) {
			members := entry.PartyData[i]
			var charIDs []string
			for _, member := range members {
				charID := id.GetStudentID(member)
				if charID == 0 {
					continue
				}

				updateSummaryFilters(filters, assistFilters, member)
				charIDs = append(charIDs, strconv.Itoa(charID))
			}
			sort.Strings(charIDs)
			partyKeys = append(partyKeys, strings.Join(charIDs, "_"))
		}
		key := strings.Join(partyKeys, "_")
		parties = append(parties, key)
	}

	sort.Strings(parties)

	if len(parties) == 0 {
		return result
	}

	type partyUsage struct {
		key   string
		count int
	}

	usages := make([]partyUsage, 0)
	currentKey := parties[0]
	currentCount := 1

	for i := 1; i < len(parties); i++ {
		if parties[i] == currentKey {
			currentCount++
		} else {
			usages = append(usages, partyUsage{currentKey, currentCount})
			currentKey = parties[i]
			currentCount = 1
		}
	}
	usages = append(usages, partyUsage{currentKey, currentCount})

	sort.Slice(usages, func(i, j int) bool {
		if usages[i].count != usages[j].count {
			return usages[i].count > usages[j].count
		}
		return usages[i].key < usages[j].key
	})

	for i := 0; i < 5 && i < len(usages); i++ {
		party := []any{
			usages[i].key,
			usages[i].count,
		}
		result.Top5Partys = append(result.Top5Partys, party)
	}

	thresholds := getThresholds(lunaticCount, tormentCount, level == "torment")
	for _, threshold := range thresholds {
		partyCounts := make([]int, 4)
		for i := range data {
			entry := data[i]
			rank := entry.Rank
			if rank > threshold {
				continue
			}
			partyData := entry.PartyData
			numParties := len(partyData)
			if numParties >= 4 {
				partyCounts[3]++
			} else if numParties > 0 {
				partyCounts[numParties-1]++
			}
		}
		result.PartyCounts[fmt.Sprintf("in%d", threshold)] = partyCounts
	}

	return result
}

func getThresholds(lunaticCount, tormentCount int, isTorment bool) []int {
	allThresholds := []int{100, 200, 500, 1000, 2000, 5000, 10000, 20000}
	var thresholds []int

	if isTorment {
		for _, t := range allThresholds {
			if lunaticCount < t && t <= (lunaticCount+tormentCount) {
				thresholds = append(thresholds, t)
			}
		}
		if len(thresholds) == 0 || thresholds[len(thresholds)-1] != (lunaticCount+tormentCount) {
			thresholds = append(thresholds, lunaticCount+tormentCount)
		}
	} else {
		for _, t := range allThresholds {
			if t <= lunaticCount {
				thresholds = append(thresholds, t)
			}
		}
		if len(thresholds) == 0 || thresholds[len(thresholds)-1] != lunaticCount {
			thresholds = append(thresholds, lunaticCount)
		}
	}

	return thresholds
}
