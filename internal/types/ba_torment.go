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
	VideoID   *string  `json:"video_id,omitempty"`
}

type BATormentSummaryData struct {
	Torment              BATormentLevelData    `json:"torment"`
	Lunatic              BATormentLevelData    `json:"lunatic"`
	PlatinumCuts         []PlatinumCut         `json:"platinumCuts,omitempty"`
	EssentialCharacters  []EssentialCharacter  `json:"essentialCharacters,omitempty"`
	HighImpactCharacters []HighImpactCharacter `json:"highImpactCharacters,omitempty"`
	MinUEUsers           *MinUEUsers           `json:"minUEUsers,omitempty"`
	MaxPartyUsers        *MaxPartyUsers        `json:"maxPartyUsers,omitempty"`
}

// EssentialCharacter represents a character used by 70%+ of platinum users
type EssentialCharacter struct {
	StudentID int     `json:"studentId"`
	Ratio     float64 `json:"ratio"`
}

// HighImpactCharacter represents a character with high score impact when missing
type HighImpactCharacter struct {
	StudentID  int `json:"studentId"`
	ScoreGap   int `json:"scoreGap"`   // 1st place score - best score without this character
	TopScore   int `json:"topScore"`   // 1st place score (with this character)
	WithoutMax int `json:"withoutMax"` // best score without this character
}

// MinUEUser represents a user who cleared with minimum unique equipment usage
type MinUEUser struct {
	UECount    int      `json:"ueCount"`
	PartyCount int      `json:"partyCount"`
	PartyData  [][6]int `json:"partyData"`
}

// MinUEUsers contains min UE users for each difficulty level
type MinUEUsers struct {
	Torment *MinUEUser `json:"torment,omitempty"`
	Lunatic *MinUEUser `json:"lunatic,omitempty"`
}

// MaxPartyUser represents a user who cleared with maximum party count
type MaxPartyUser struct {
	PartyCount int      `json:"partyCount"`
	PartyData  [][6]int `json:"partyData"`
}

// MaxPartyUsers contains max party users for each difficulty level
type MaxPartyUsers struct {
	Torment *MaxPartyUser `json:"torment,omitempty"`
	Lunatic *MaxPartyUser `json:"lunatic,omitempty"`
}

// PlatinumCut represents a score cutoff at a specific rank
// Used for showing score thresholds at 2000, 4000, ..., 20000 ranks
type PlatinumCut struct {
	Rank  int `json:"rank"`
	Score int `json:"score"`
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

// YoutubeAnalysisResult represents the analysis_result JSONB field in youtube_analysis table
type YoutubeAnalysisResult struct {
	URL         string   `json:"url"`
	Score       int      `json:"score"`
	PartyData   [][6]int `json:"partyData"`
	SkillOrders []any    `json:"skillOrders"`
}
