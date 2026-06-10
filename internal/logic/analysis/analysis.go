package analysis

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"time"

	"ba-torment-data-process/internal/types"

	"github.com/BeaverHouse/go-common/logger"
)

const PartyDataBaseURL = "https://twauaebyyujvvvusbrwe.supabase.co/storage/v1/object/public/pb7h4uvn2b6m0lyu7i6r3j8ac/batorment/v3/party"

// DownloadPartyData downloads party data for a single content ID, returns nil on failure.
func DownloadPartyData(log logger.Logger, contentID string) *types.BATormentPartyData {
	url := PartyDataBaseURL + "/" + contentID + ".json"
	data, err := fetchPartyData(log, url)
	if err != nil {
		log.Warn("Failed to fetch party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
		return nil
	}

	var partyData types.BATormentPartyData
	if err := json.Unmarshal(data, &partyData); err != nil {
		log.Warn("Failed to parse party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
		return nil
	}

	log.Info("Downloaded party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "parties", Value: len(partyData.PartyDetail)})
	return &partyData
}

// DownloadAllPartyData downloads party data for all content IDs
func DownloadAllPartyData(log logger.Logger, contentIDs []string) map[string]*types.BATormentPartyData {
	result := make(map[string]*types.BATormentPartyData)

	for _, contentID := range contentIDs {
		url := PartyDataBaseURL + "/" + contentID + ".json"
		data, err := fetchPartyData(log, url)
		if err != nil {
			log.Warn("Failed to fetch party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
			continue
		}

		var partyData types.BATormentPartyData
		if err := json.Unmarshal(data, &partyData); err != nil {
			log.Warn("Failed to parse party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
			continue
		}

		result[contentID] = &partyData
		log.Info("Downloaded party data", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "parties", Value: len(partyData.PartyDetail)})
	}

	return result
}

// fetchPartyData fetches party data from URL (non-fatal on error)
func fetchPartyData(log logger.Logger, url string) ([]byte, error) {
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

	log.Info("Fetched", logger.Field{Key: "url", Value: url}, logger.Field{Key: "duration", Value: time.Since(start)})
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
func RunTotalAnalysis(log logger.Logger, partyDataMap map[string]*types.BATormentPartyData, sortedContentIDs []string) *types.TotalAnalysisOutput {
	log.Info("Starting total analysis", logger.Field{Key: "raids", Value: len(partyDataMap)})

	// Raid analysis
	log.Info("Running raid analyses...")
	raidAnalyses := RunRaidAnalyses(partyDataMap, sortedContentIDs)
	log.Info("Completed raid analyses", logger.Field{Key: "raids", Value: len(raidAnalyses)})

	// Character analysis
	log.Info("Running character analyses...")
	characterAnalyses := RunCharacterAnalyses(partyDataMap, sortedContentIDs)
	log.Info("Completed character analyses", logger.Field{Key: "characters", Value: len(characterAnalyses)})

	// Calculate and assign overall rankings
	log.Info("Calculating overall rankings...")
	AssignOverallRankings(characterAnalyses)
	log.Info("Completed overall ranking calculation")

	return &types.TotalAnalysisOutput{
		GeneratedAt:       time.Now().Format(time.RFC3339),
		RaidAnalyses:      raidAnalyses,
		CharacterAnalyses: characterAnalyses,
	}
}

// AssignOverallRankings calculates and assigns overall usage rankings to characterAnalyses
func AssignOverallRankings(characterAnalyses []types.CharacterAnalysisResult) {
	// Calculate total usage for each character
	for i := range characterAnalyses {
		totalUsage := 0
		for _, ru := range characterAnalyses[i].UsageHistory {
			totalUsage += ru.UserCount
		}
		characterAnalyses[i].TotalUsage = totalUsage
	}

	// Create index slice for sorting
	indices := make([]int, len(characterAnalyses))
	for i := range indices {
		indices[i] = i
	}

	// Sort indices by total usage (descending)
	sort.Slice(indices, func(i, j int) bool {
		return characterAnalyses[indices[i]].TotalUsage > characterAnalyses[indices[j]].TotalUsage
	})

	// Assign overall ranks based on sorted order
	for rank, idx := range indices {
		characterAnalyses[idx].OverallRank = rank + 1
	}

	// Calculate category ranks (striker: 1xxxx, special: 2xxxx)
	strikerRank := 1
	specialRank := 1
	for _, idx := range indices {
		studentID := characterAnalyses[idx].StudentID
		if studentID >= 10000 && studentID < 20000 {
			characterAnalyses[idx].CategoryRank = strikerRank
			strikerRank++
		} else if studentID >= 20000 && studentID < 30000 {
			characterAnalyses[idx].CategoryRank = specialRank
			specialRank++
		}
	}
}
