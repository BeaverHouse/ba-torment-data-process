package videoref

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	logic_download "ba-torment-data-process/internal/logic/download"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"ba-torment-data-process/internal/types"
)

const (
	supabaseStorageURL = "https://twauaebyyujvvvusbrwe.supabase.co/storage/v1/object/public/pb7h4uvn2b6m0lyu7i6r3j8ac/batorment/v3/party"
	hoshinoMusoCodeA   = 10098
	hoshinoMusoCodeB   = 10099
)

// UpdateVideoRef updates video references for party data based on verified YouTube analysis
func UpdateVideoRef(dryRun bool, raidIDs []string) {
	defer func() {
		log.Println("비디오 ref 업데이트 프로세스 완료")
	}()

	// Connect to database using existing init function
	pool := postgres.InitFromEnv()
	defer pool.Close()

	ctx := context.Background()
	queries := postgres.New(pool)

	for _, raidID := range raidIDs {
		log.Printf("Processing raid: %s", raidID)

		// 1. Download party data from Supabase
		partyData, err := downloadPartyData(raidID)
		if err != nil {
			log.Printf("Failed to download party data for %s: %v", raidID, err)
			continue
		}

		// 2. Get verified YouTube analysis from database
		analysisRows, err := queries.GetVerifiedYoutubeAnalysisByRaidID(ctx, raidID)
		if err != nil {
			log.Printf("Failed to query YouTube analysis for %s: %v", raidID, err)
			continue
		}

		if len(analysisRows) == 0 {
			log.Printf("No verified YouTube analysis found for %s", raidID)
			continue
		}

		// 3. Parse analysis results
		analysisResults := make([]types.YoutubeAnalysisResult, 0, len(analysisRows))
		videoIDMap := make(map[int]string) // index -> video_id

		for idx, row := range analysisRows {
			var result types.YoutubeAnalysisResult
			if err := json.Unmarshal(row.AnalysisResult, &result); err != nil {
				log.Printf("Failed to unmarshal analysis result: %v", err)
				continue
			}
			analysisResults = append(analysisResults, result)
			videoIDMap[idx] = row.VideoID
		}

		// 4. Match and update video references
		updated := matchAndUpdateVideoRefs(partyData, analysisResults, videoIDMap)

		// 5. Upload updated party data
		fileName := fmt.Sprintf("%s.json", raidID)
		if err := logic_upload.MarshalAndUpload(partyData, "batorment/v3/party", fileName, dryRun, fmt.Sprintf("파티 데이터 video ref 업데이트 완료: %s (updated: %d)", raidID, updated)); err != nil {
			log.Printf("Failed to upload party data for %s: %v", raidID, err)
		}
	}
}

// downloadPartyData downloads party data from Supabase storage
func downloadPartyData(raidID string) (*types.BATormentPartyData, error) {
	url := fmt.Sprintf("%s/%s.json", supabaseStorageURL, raidID)
	data := logic_download.GetDataFromURL(url)

	var partyData types.BATormentPartyData
	if err := json.Unmarshal(data, &partyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal party data: %w", err)
	}

	return &partyData, nil
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
