package parse

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/types"
	"fmt"
	"io"
	"strconv"
)

var (
	// Rank cutoff for platinum tier
	PlatinumCut = 20000
)

// GetRankData returns rank data for a given season
// season: season starts with "S" or "3S"
// category: target boss number if it's Grand Assault (대결전)
func GetRankData(seasonString string) ([]types.RankData, error) {
	_, category, err := logic.SplitSeasonString(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetRankData", err)
	}

	if category != 0 {
		return getGrandAssaultRankData(seasonString, category)
	}
	return getTotalAssaultRankData(seasonString)
}

// Get "Total Assault (총력전)" rank data
func getTotalAssaultRankData(seasonString string) ([]types.RankData, error) {
	reader, err := data.GetRankCSVFromGoogleAPI(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("getTotalAssaultRankData", err)
	}

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Find the column indices
	var userIDIdx, rankIdx, scoreIdx int
	for i, col := range header {
		switch col {
		case "AccountId":
			userIDIdx = i
		case "Rank":
			rankIdx = i
		case "BestRankingPoint":
			scoreIdx = i
		}
	}

	var rankData []types.RankData

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		userID, _ := strconv.ParseInt(record[userIDIdx], 10, 64)
		rank, _ := strconv.Atoi(record[rankIdx])
		score, _ := strconv.ParseInt(record[scoreIdx], 10, 64)

		if rank > PlatinumCut {
			break
		}

		rankData = append(rankData, types.RankData{
			UserID:    userID,
			FinalRank: rank,
			Score:     score,
		})
	}

	return rankData, nil
}

// Get "Grand Assault (대결전)" rank data
func getGrandAssaultRankData(seasonString string, category int) ([]types.RankData, error) {
	reader, err := data.GetRankCSVFromGoogleAPI(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("getGrandAssaultRankData", err)
	}

	header, err := reader.Read()
	if err != nil {
		return nil, common.WrapErrorWithContext("getGrandAssaultRankData", err)
	}

	// Find the column indices
	var userIDIdx, rankIdx, scoreIdx int
	for i, col := range header {
		switch col {
		case "AccountId":
			userIDIdx = i
		case "Rank":
			rankIdx = i
		case fmt.Sprintf("Boss%d", category):
			scoreIdx = i
		}
	}

	var rankData []types.RankData

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read record: %w", err)
		}

		rank, _ := strconv.Atoi(record[rankIdx])
		score, _ := strconv.ParseInt(record[scoreIdx], 10, 64)
		userID, _ := strconv.ParseInt(record[userIDIdx], 10, 64)

		if rank > PlatinumCut {
			break
		}

		rankData = append(rankData, types.RankData{
			UserID:    userID,
			FinalRank: rank,
			Score:     score,
		})
	}

	return rankData, nil
}
