package parse

import (
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/types"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func ProcessPartyDataToSummaryData(partyData *types.BATormentPartyData) (*types.BATormentSummaryData, error) {
	// 토먼트/루나틱 데이터 분리
	var lunaticData, tormentData []types.BATormentPartyDetail
	for _, data := range partyData.PartyDetail {
		if data.Level == "L" {
			lunaticData = append(lunaticData, data)
		} else if data.Level == "T" {
			tormentData = append(tormentData, data)
		}
	}

	lunaticCount := len(lunaticData)
	tormentCount := len(tormentData)

	// 최종 데이터 구성
	result := &types.BATormentSummaryData{
		Torment: processLevelData(tormentData, "torment", lunaticCount, tormentCount),
		Lunatic: processLevelData(lunaticData, "lunatic", lunaticCount, tormentCount),
	}

	return result, nil
}

// processLevelData는 각 레벨의 데이터를 처리합니다.
func processLevelData(data []types.BATormentPartyDetail, level string, lunaticCount, tormentCount int) types.BATormentLevelData {
	result := types.BATormentLevelData{
		ClearCount:    len(data),
		PartyCounts:   make(map[string][]int),
		Filters:       make(map[string][]int),
		AssistFilters: make(map[string][]int),
		Top5Partys:    make([][]interface{}, 0),
	}

	parties := make([]string, 0)

	// 필터 데이터 처리
	filters := make(map[string][]int)
	assistFilters := make(map[string][]int)
	totalCount := len(data)

	for _, entry := range data {
		var partyKeys []string
		for i := range len(entry.PartyData) {
			members := entry.PartyData["party_"+strconv.Itoa(i+1)]
			var charIDs []string
			for _, member := range members {

				charID := member / 1000

				if charID == 0 {
					continue
				}

				logic.UpdateSummaryFilters(filters, assistFilters, member)

				charIDs = append(charIDs, strconv.Itoa(charID))
			}
			sort.Strings(charIDs)
			partyKeys = append(partyKeys, strings.Join(charIDs, "_"))
		}
		key := strings.Join(partyKeys, "_")
		parties = append(parties, key)
	}

	// 1% 미만 사용률 필터 제거
	logic.DropLowUsageFilters(filters, totalCount)
	logic.DropLowUsageFilters(assistFilters, totalCount)

	result.Filters = filters
	result.AssistFilters = assistFilters

	// 파티 목록을 정렬
	sort.Strings(parties)

	type partyUsage struct {
		key   string
		count int
	}

	usages := make([]partyUsage, 0)
	currentKey := parties[0]
	currentCount := 1

	// 정렬된 목록에서 연속된 같은 값의 개수를 세어 저장
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

	// 사용 횟수로 정렬하고, 같은 횟수면 키로 정렬
	sort.Slice(usages, func(i, j int) bool {
		if usages[i].count != usages[j].count {
			return usages[i].count > usages[j].count
		}
		return usages[i].key < usages[j].key
	})
	// Top 5 파티 추출
	for i := 0; i < 5 && i < len(usages); i++ {
		party := []interface{}{
			usages[i].key,
			usages[i].count,
		}
		result.Top5Partys = append(result.Top5Partys, party)
	}

	// 임계값에 따른 파티 카운트 계산
	thresholds := getThresholds(lunaticCount, tormentCount, level == "torment")
	for _, threshold := range thresholds {
		// 파티 수에 따른 카운트 계산 [1파티, 2파티, 3파티, 4파티 이상]
		partyCounts := make([]int, 4)
		for i := range data {
			entry := data[i]
			rank := entry.TormentRank
			if rank > threshold {
				continue
			}
			partyData := entry.PartyData
			// 파티 데이터의 키 개수로 파티 수 계산
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

// getThresholds는 클리어 수에 따라 적절한 임계값 목록을 반환합니다.
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
