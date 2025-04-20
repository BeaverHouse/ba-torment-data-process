package tests

import (
	"ba-torment-data-process/app/types"
	"testing"

	"github.com/stretchr/testify/require"
)

// Compare the 2 filters.
func compareFilters(t *testing.T, expected map[string][]int, actual map[string][]int, title string) {
	require.Equal(t, len(expected), len(actual), title+" - 개수가 일치하지 않습니다")
	for key, value := range expected {
		require.Equal(t, value, actual[key], title+" - 값이 일치하지 않습니다: "+key)
	}
}

// Compare the party data. If it's Arona.AI data, ignores the user ID data.
func ComparePartyData(t *testing.T, parsedPartyData *types.BATormentPartyData, baTormentPartyData *types.BATormentPartyData, isAronaAI bool) {
	compareFilters(t, parsedPartyData.Filters, baTormentPartyData.Filters, "필터 데이터")
	compareFilters(t, parsedPartyData.AssistFilters, baTormentPartyData.AssistFilters, "어시스트 필터 데이터")

	// 최소, 최대 파티 수 비교
	require.Equal(t, parsedPartyData.MinPartys, baTormentPartyData.MinPartys, "최소 파티 수가 일치하지 않습니다")
	require.Equal(t, parsedPartyData.MaxPartys, baTormentPartyData.MaxPartys, "최대 파티 수가 일치하지 않습니다")

	// 파티 데이터 개수 비교
	if len(parsedPartyData.PartyDetail) != len(baTormentPartyData.PartyDetail) {
		t.Errorf("파티 데이터 개수가 일치하지 않습니다. 예상: %d, 실제: %d", len(baTormentPartyData.PartyDetail), len(parsedPartyData.PartyDetail))
	}

	// 파티 데이터 비교
	for i := range parsedPartyData.PartyDetail {
		if isAronaAI {
			baTormentPartyData.PartyDetail[i].UserID = int64(-i - 1)
		}
		require.Equal(t, parsedPartyData.PartyDetail[i], baTormentPartyData.PartyDetail[i], "%d번째 파티 데이터가 일치하지 않습니다", i)
	}
}

// Compare the summary data.
func CompareSummaryData(t *testing.T, parsedSummaryData *types.BATormentSummaryData, baTormentSummaryData *types.BATormentSummaryData) {
	// 토먼트 데이터 비교
	require.Equal(t, parsedSummaryData.Torment.ClearCount, baTormentSummaryData.Torment.ClearCount, "토먼트 클리어 수가 일치하지 않습니다")
	require.Equal(t, parsedSummaryData.Torment.PartyCounts, baTormentSummaryData.Torment.PartyCounts, "토먼트 파티 수가 일치하지 않습니다")

	compareFilters(t, parsedSummaryData.Torment.Filters, baTormentSummaryData.Torment.Filters, "토먼트 필터 데이터")
	compareFilters(t, parsedSummaryData.Torment.AssistFilters, baTormentSummaryData.Torment.AssistFilters, "토먼트 어시스트 필터 데이터")

	require.Equal(t, len(parsedSummaryData.Torment.Top5Partys), len(baTormentSummaryData.Torment.Top5Partys), "토먼트 상위 5개 파티 데이터 개수가 일치하지 않습니다")
	for i := range parsedSummaryData.Torment.Top5Partys {
		require.Equal(t, parsedSummaryData.Torment.Top5Partys[i][0], baTormentSummaryData.Torment.Top5Partys[i][0], "토먼트 상위 5개 파티 키가 일치하지 않습니다: %d", i)
		require.EqualValues(t, parsedSummaryData.Torment.Top5Partys[i][1], baTormentSummaryData.Torment.Top5Partys[i][1], "토먼트 상위 5개 파티 수가 일치하지 않습니다: %d", i)
	}

	// 루나틱 데이터 비교
	require.Equal(t, parsedSummaryData.Lunatic.ClearCount, baTormentSummaryData.Lunatic.ClearCount, "루나틱 클리어 수가 일치하지 않습니다")
	require.Equal(t, parsedSummaryData.Lunatic.PartyCounts, baTormentSummaryData.Lunatic.PartyCounts, "루나틱 파티 수가 일치하지 않습니다")

	compareFilters(t, parsedSummaryData.Lunatic.Filters, baTormentSummaryData.Lunatic.Filters, "루나틱 필터 데이터")
	compareFilters(t, parsedSummaryData.Lunatic.AssistFilters, baTormentSummaryData.Lunatic.AssistFilters, "루나틱 어시스트 필터 데이터")

	require.Equal(t, len(parsedSummaryData.Lunatic.Top5Partys), len(baTormentSummaryData.Lunatic.Top5Partys), "루나틱 상위 5개 파티 데이터 개수가 일치하지 않습니다")
	for i := range parsedSummaryData.Lunatic.Top5Partys {
		require.Equal(t, parsedSummaryData.Lunatic.Top5Partys[i][0], baTormentSummaryData.Lunatic.Top5Partys[i][0], "루나틱 상위 5개 파티 키가 일치하지 않습니다: %d", i)
		require.EqualValues(t, parsedSummaryData.Lunatic.Top5Partys[i][1], baTormentSummaryData.Lunatic.Top5Partys[i][1], "루나틱 상위 5개 파티 수가 일치하지 않습니다: %d", i)
	}
}
