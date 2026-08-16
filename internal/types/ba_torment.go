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
	Rank        int      `json:"rank"`
	Score       int      `json:"score"`
	PartyData   [][6]int `json:"partyData"`
	SkillOrders [][6]int `json:"skillOrders,omitempty"`
	VideoID     *string  `json:"video_id,omitempty"`
}

type BATormentSummaryData struct {
	Torment          BATormentLevelData `json:"torment"`
	Lunatic          BATormentLevelData `json:"lunatic"`
	PlatinumCuts     []PlatinumCut      `json:"platinumCuts,omitempty"`
	PartPlatinumCuts []PlatinumCut      `json:"partPlatinumCuts,omitempty"` // For grand assault: individual part cuts
}

// EssentialCharacter represents a character used by 70%+ of platinum users
type EssentialCharacter struct {
	StudentID int     `json:"studentId"`
	Ratio     float64 `json:"ratio"`
}

// HighImpactCharacter represents a character with high rank impact when missing
type HighImpactCharacter struct {
	StudentID       int `json:"studentId"`
	RankGap         int `json:"rankGap"`         // WithoutBestRank - TopRank, or -1 if 100% usage
	TopRank         int `json:"topRank"`         // best rank in this difficulty (uses this character)
	WithoutBestRank int `json:"withoutBestRank"` // best rank achieved without this character (-1 if 100% usage)
}

// MinUEUser represents a user who cleared with minimum unique equipment usage
type MinUEUser struct {
	Rank        int      `json:"rank"`
	Score       int      `json:"score"`
	UECount     int      `json:"ueCount"`
	PartyData   [][6]int `json:"partyData"`
	SkillOrders [][6]int `json:"skillOrders,omitempty"`
}

// MaxPartyUser represents a user who cleared with maximum party count
type MaxPartyUser struct {
	Rank        int      `json:"rank"`
	Score       int      `json:"score"`
	PartyData   [][6]int `json:"partyData"`
	SkillOrders [][6]int `json:"skillOrders,omitempty"`
}

// PlatinumCut represents a score cutoff at a specific rank
// Used for showing score thresholds at 2000, 4000, ..., 20000 ranks
type PlatinumCut struct {
	Rank  int `json:"rank"`
	Score int `json:"score"`
}

type BATormentLevelData struct {
	ClearCount           int                   `json:"clearCount"`
	PartyCounts          map[string][]int      `json:"partyCounts"`
	Top5Partys           [][]any               `json:"top5Partys"`
	EssentialCharacters  []EssentialCharacter  `json:"essentialCharacters,omitempty"`
	HighImpactCharacters []HighImpactCharacter `json:"highImpactCharacters,omitempty"`
	MinUEUser            *MinUEUser            `json:"minUEUser,omitempty"`
	MaxPartyUser         *MaxPartyUser         `json:"maxPartyUser,omitempty"`
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
	Score     int      `json:"score"`
	PartyData [][6]int `json:"partyData"`
}
