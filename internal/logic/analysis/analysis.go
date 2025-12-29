package analysis

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"ba-torment-data-process/internal/types"
)

const PartyDataBaseURL = "https://twauaebyyujvvvusbrwe.supabase.co/storage/v1/object/public/pb7h4uvn2b6m0lyu7i6r3j8ac/batorment/v3/party"

// DownloadAllPartyData downloads party data for all content IDs
func DownloadAllPartyData(contentIDs []string) map[string]*types.BATormentPartyData {
	result := make(map[string]*types.BATormentPartyData)

	for _, contentID := range contentIDs {
		url := PartyDataBaseURL + "/" + contentID + ".json"
		data, err := fetchPartyData(url)
		if err != nil {
			log.Printf("Failed to fetch party data for %s: %v", contentID, err)
			continue
		}

		var partyData types.BATormentPartyData
		if err := json.Unmarshal(data, &partyData); err != nil {
			log.Printf("Failed to parse party data for %s: %v", contentID, err)
			continue
		}

		result[contentID] = &partyData
		log.Printf("Downloaded party data for %s: %d parties", contentID, len(partyData.PartyDetail))
	}

	return result
}

// fetchPartyData fetches party data from URL (non-fatal on error)
func fetchPartyData(url string) ([]byte, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, &httpError{StatusCode: resp.StatusCode, URL: url}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("Fetched: url=%s, duration=%s", url, time.Since(start))
	return body, nil
}

type httpError struct {
	StatusCode int
	URL        string
}

func (e *httpError) Error() string {
	return "HTTP " + string(rune(e.StatusCode)) + " for " + e.URL
}

// RunTotalAnalysis runs the complete analysis
// sortedContentIDs provides the order for raidAnalyses (sorted by start_date)
func RunTotalAnalysis(partyDataMap map[string]*types.BATormentPartyData, sortedContentIDs []string) *types.TotalAnalysisOutput {
	log.Printf("Starting total analysis with %d raids", len(partyDataMap))

	// Raid analysis
	log.Println("Running raid analyses...")
	raidAnalyses := RunRaidAnalyses(partyDataMap, sortedContentIDs)
	log.Printf("Completed raid analyses: %d raids", len(raidAnalyses))

	// Character analysis
	log.Println("Running character analyses...")
	characterAnalyses := RunCharacterAnalyses(partyDataMap, sortedContentIDs)
	log.Printf("Completed character analyses: %d characters", len(characterAnalyses))

	return &types.TotalAnalysisOutput{
		GeneratedAt:       time.Now().Format(time.RFC3339),
		RaidAnalyses:      raidAnalyses,
		CharacterAnalyses: characterAnalyses,
	}
}
