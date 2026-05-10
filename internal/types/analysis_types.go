package types

// CharacterUsage represents character usage statistics
type CharacterUsage struct {
	StudentID  int `json:"studentId"`
	UsageCount int `json:"usageCount"`
}

// RaidAnalysisResult represents analysis result for a single raid
type RaidAnalysisResult struct {
	RaidID            string           `json:"raidId"`
	TopStrikers       []CharacterUsage `json:"topStrikers"`
	TopSpecials       []CharacterUsage `json:"topSpecials"`
	TopAssists        []CharacterUsage `json:"topAssists"`
	LunaticClearCount int              `json:"lunaticClearCount"`
}

// RaidUsage represents usage statistics for a character in a specific raid
type RaidUsage struct {
	RaidID           string `json:"raidId"`
	UserCount        int    `json:"userCount"`
	LunaticUserCount int    `json:"lunaticUserCount"`
}

// RaidStarDistribution represents star distribution for a character in a specific raid
type RaidStarDistribution struct {
	RaidID       string         `json:"raidId"`
	Distribution map[string]int `json:"distribution"` // "star_weaponStar" -> count
}

// AssistUsageStats represents assist vs own usage statistics
type AssistUsageStats struct {
	AsAssistCount int     `json:"asAssistCount"`
	AsOwnCount    int     `json:"asOwnCount"`
	TotalCount    int     `json:"totalCount"`
	AssistRatio   float64 `json:"assistRatio"`
}

// CharacterSynergy represents co-usage statistics with another character
type CharacterSynergy struct {
	StudentID    int     `json:"studentId"`
	CoUsageRate  float64 `json:"coUsageRate"`
	CoUsageCount int     `json:"coUsageCount"`
}

// CharacterAnalysisResult represents analysis result for a single character
type CharacterAnalysisResult struct {
	StudentID        int                   `json:"studentId"`
	UsageHistory     []RaidUsage           `json:"usageHistory"`
	StarDistribution *RaidStarDistribution `json:"starDistribution"` // Latest distribution (200+ usage), null if none
	AssistStats      AssistUsageStats      `json:"assistStats"`
	TopSynergyChars  []CharacterSynergy    `json:"topSynergyChars"`
	TotalUsage       int                   `json:"totalUsage"`
	OverallRank      int                   `json:"overallRank"`
	CategoryRank     int                   `json:"categoryRank"`
}

// TotalAnalysisOutput represents the final output of total analysis
type TotalAnalysisOutput struct {
	GeneratedAt       string                    `json:"generatedAt"`
	RaidAnalyses      []RaidAnalysisResult      `json:"raidAnalyses"`
	CharacterAnalyses []CharacterAnalysisResult `json:"characterAnalyses"`
}
