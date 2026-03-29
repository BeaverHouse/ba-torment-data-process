package party

import (
	"context"
	"fmt"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/id"
	"ba-torment-data-process/internal/types"
	"ba-torment-data-process/internal/ui"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/logger"
)

const (
	hoshinoMusoCodeA = 10098
	hoshinoMusoCodeB = 10099
)

// UpdateVideoRefWithData updates video references for party data
func UpdateVideoRefWithData(partyData *types.BATormentPartyData, raidID string) (int, error) {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()

	ctx := context.Background()
	queries := postgres.New(pool)

	analysisRows, err := queries.GetVerifiedYoutubeAnalysisByRaidID(ctx, raidID)
	if err != nil {
		return 0, fmt.Errorf("failed to query YouTube analysis: %w", err)
	}

	if len(analysisRows) == 0 {
		ui.Log.Info("No verified YouTube analysis found", logger.F("raidID", raidID))
		return 0, nil
	}

	analysisResults := make([]types.YoutubeAnalysisResult, 0, len(analysisRows))
	videoIDMap := make(map[int]string)

	for idx, row := range analysisRows {
		analysisResults = append(analysisResults, row.AnalysisResult)
		videoIDMap[idx] = row.VideoID
	}

	updated := matchAndUpdateVideoRefs(partyData, analysisResults, videoIDMap)
	ui.Log.Info("Updated video references", logger.F("count", updated), logger.F("raidID", raidID))

	return updated, nil
}

func matchAndUpdateVideoRefs(partyData *types.BATormentPartyData, analysisResults []types.YoutubeAnalysisResult, videoIDMap map[int]string) int {
	updated := 0
	usedVideos := make(map[string]bool)

	for i := range partyData.PartyDetail {
		partyData.PartyDetail[i].VideoID = nil
	}

	for i := range partyData.PartyDetail {
		party := &partyData.PartyDetail[i]

		for idx, analysis := range analysisResults {
			videoID := videoIDMap[idx]

			if usedVideos[videoID] {
				continue
			}

			if isMatch(party, &analysis) {
				party.VideoID = &videoID
				usedVideos[videoID] = true
				updated++
				ui.Log.Info("Matched party with video", logger.F("rank", party.Rank), logger.F("score", party.Score), logger.F("videoID", videoID))
				break
			}
		}
	}

	return updated
}

func isMatch(party *types.BATormentPartyDetail, analysis *types.YoutubeAnalysisResult) bool {
	if party.Score != analysis.Score {
		return false
	}

	if len(party.PartyData) != len(analysis.PartyData) {
		return false
	}

	for i := range party.PartyData {
		if !isPartyMatch(party.PartyData[i], analysis.PartyData[i]) {
			return false
		}
	}

	return true
}

// isPartyMatch checks if a single party matches, with exceptions:
// 1. Hoshino Muso: 10098 <-> 10099 (allowed only once per party)
// 2. Missing assistant in analysis (last digit difference only, allowed only once per party)
func isPartyMatch(partyA [6]int, partyB [6]int) bool {
	hoshinoExceptionCount := 0
	assistantExceptionCount := 0

	for i := 0; i < 6; i++ {
		if partyA[i] == partyB[i] {
			continue
		}

		if isHoshinoMusoException(partyA[i], partyB[i]) {
			hoshinoExceptionCount++
			if hoshinoExceptionCount > 2 {
				return false
			}
			continue
		}

		if isAssistantException(partyA[i], partyB[i]) {
			assistantExceptionCount++
			if assistantExceptionCount > 1 {
				return false
			}
			continue
		}

		return false
	}

	return true
}

func isHoshinoMusoException(a, b int) bool {
	studentA := id.GetStudentID(a)
	studentB := id.GetStudentID(b)

	return (studentA == hoshinoMusoCodeA && studentB == hoshinoMusoCodeB) ||
		(studentA == hoshinoMusoCodeB && studentB == hoshinoMusoCodeA)
}

func isAssistantException(a, b int) bool {
	return a/10 == b/10 && a%10 != b%10
}
