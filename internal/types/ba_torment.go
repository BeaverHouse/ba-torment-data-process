package types

type BATormentFilter struct {
	Filters       map[string](map[string]int) `json:"filters"`
	AssistFilters map[string](map[string]int) `json:"assistFilters"`
}

type BATormentPartyData struct {
	MinPartys   int                    `json:"minPartys"`
	MaxPartys   int                    `json:"maxPartys"`
	PartyDetail []BATormentPartyDetail `json:"parties"`
}

type BATormentPartyDetail struct {
	Rank      int      `json:"rank"`
	Score     int      `json:"score"`
	PartyData [][6]int `json:"partyData"`
}

type BATormentSummaryData struct {
	Torment BATormentLevelData `json:"torment"`
	Lunatic BATormentLevelData `json:"lunatic"`
}

type BATormentLevelData struct {
	ClearCount  int              `json:"clearCount"`
	PartyCounts map[string][]int `json:"partyCounts"`
	Top5Partys  [][]any          `json:"top5Partys"`
}

// ********************************************
// ********* BA Analyzer - Video Data *********
// ********************************************

// VideoAnalysisSummary represents a video analysis summary item
type VideoAnalysisSummary struct {
	VideoID     string  `json:"video_id"`
	Score       int64   `json:"score"`
	Title       string  `json:"title"`
	RaidID      *string `json:"raid_id"`
	CreatedAt   string  `json:"created_at"`
	IsVerified  bool    `json:"is_verified"`
	PartyData   [][]int `json:"party_data"`
	VerifyLevel int     `json:"verify_level"`
}

// VideoAnalysisListResponse represents the response for video analysis list
type VideoAnalysisListResponse struct {
	Data []VideoAnalysisSummary `json:"data"`
}
