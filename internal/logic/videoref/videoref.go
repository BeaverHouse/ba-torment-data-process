package videoref

import (
	"context"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/types"
)

const (
	hoshinoMusoCodeA = 10098
	hoshinoMusoCodeB = 10099
)

// Updates video references for party data (directly from DuckDB)
// This function is designed to be used in integrated pipelines where partyData is already in memory
func UpdateVideoRefWithData(partyData *types.BATormentPartyData, raidID string) (int, error) {
	// Connect to database
	pool := postgres.InitFromEnv()
	defer pool.Close()

	ctx := context.Background()
	queries := postgres.New(pool)

	// Get verified YouTube analysis from database
	analysisRows, err := queries.GetVerifiedYoutubeAnalysisByRaidID(ctx, raidID)
	if err != nil {
		return 0, fmt.Errorf("failed to query YouTube analysis: %w", err)
	}

	if len(analysisRows) == 0 {
		log.Printf("No verified YouTube analysis found for %s", raidID)
		return 0, nil
	}

	// Parse analysis results
	analysisResults := make([]types.YoutubeAnalysisResult, 0, len(analysisRows))
	videoIDMap := make(map[int]string) // index -> video_id

	for idx, row := range analysisRows {
		analysisResults = append(analysisResults, row.AnalysisResult)
		videoIDMap[idx] = row.VideoID
	}

	// Match and update video references
	updated := matchAndUpdateVideoRefs(partyData, analysisResults, videoIDMap)
	log.Printf("Updated %d video references for raid %s", updated, raidID)

	return updated, nil
}

// matchAndUpdateVideoRefs matches YouTube analysis with party data and updates video references
func matchAndUpdateVideoRefs(partyData *types.BATormentPartyData, analysisResults []types.YoutubeAnalysisResult, videoIDMap map[int]string) int {
	updated := 0
	usedVideos := make(map[string]bool) // Track used video IDs

	// Initialize all video_ids to nil before matching
	for i := range partyData.PartyDetail {
		partyData.PartyDetail[i].VideoID = nil
	}

	for i := range partyData.PartyDetail {
		party := &partyData.PartyDetail[i]

		// Try to match with each analysis result
		for idx, analysis := range analysisResults {
			videoID := videoIDMap[idx]

			// Skip if this video is already used by another party
			if usedVideos[videoID] {
				continue
			}

			if isMatch(party, &analysis) {
				party.VideoID = &videoID
				usedVideos[videoID] = true // Mark this video as used
				updated++
				log.Printf("Matched party (rank=%d, score=%d) with video: %s", party.Rank, party.Score, videoID)
				break
			}
		}
	}

	return updated
}

// isMatch checks if a party matches a YouTube analysis result
func isMatch(party *types.BATormentPartyDetail, analysis *types.YoutubeAnalysisResult) bool {
	// Check score
	if party.Score != analysis.Score {
		return false
	}

	// Check party count
	if len(party.PartyData) != len(analysis.PartyData) {
		return false
	}

	// Check each party
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

		// Exception 1: Hoshino Muso (10098 <-> 10099)
		if isHoshinoMusoException(partyA[i], partyB[i]) {
			hoshinoExceptionCount++
			if hoshinoExceptionCount > 2 {
				return false // More than two Hoshino exception = different data or analysis error (Regarding own Hoshino + assist Hoshino)
			}
			continue
		}

		// Exception 2: Missing assistant (only last digit differs, e.g., 10064520 <-> 10064521)
		if isAssistantException(partyA[i], partyB[i]) {
			assistantExceptionCount++
			if assistantExceptionCount > 1 {
				return false // More than one assistant exception = different data or analysis error
			}
			continue
		}

		return false
	}

	return true
}

// isHoshinoMusoException checks if the difference is due to Hoshino Muso form change
func isHoshinoMusoException(a, b int) bool {
	studentA := a / 1000
	studentB := b / 1000

	return (studentA == hoshinoMusoCodeA && studentB == hoshinoMusoCodeB) ||
		(studentA == hoshinoMusoCodeB && studentB == hoshinoMusoCodeA)
}

// isAssistantException checks if the difference is only in the assistant flag (last digit)
func isAssistantException(a, b int) bool {
	// Check if only the last digit differs
	return a/10 == b/10 && a%10 != b%10
}
