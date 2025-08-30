package types

import (
	"database/sql"
	"time"
)

type BATormentFilter struct {
	Filters       map[string](map[string]int) `json:"filters"`
	AssistFilters map[string](map[string]int) `json:"assist_filters"`
}

type BATormentPartyData struct {
	MinPartys   int                    `json:"min_partys"`
	MaxPartys   int                    `json:"max_partys"`
	PartyDetail []BATormentPartyDetail `json:"parties"`
}

type BATormentPartyDetail struct {
	FinalRank   int     `json:"FINAL_RANK"`
	TormentRank int     `json:"TORMENT_RANK"`
	Score       int     `json:"SCORE"`
	PartyData   [][]int `json:"PARTY_DATA"`
}

type BATormentSummaryData struct {
	Torment BATormentLevelData `json:"torment"`
	Lunatic BATormentLevelData `json:"lunatic"`
}

type BATormentLevelData struct {
	ClearCount  int              `json:"clear_count"`
	PartyCounts map[string][]int `json:"party_counts"`
	Top5Partys  [][]any          `json:"top5_partys"`
}

// *******************************
// ********** DB Schema **********
// *******************************

type Raid struct {
	RaidID    string         `json:"raid_id"`
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt sql.NullTime   `json:"updated_at"`
	DeletedAt sql.NullTime   `json:"deleted_at"`
	TopLevel  sql.NullString `json:"top_level"`
}

type NamedUser struct {
	UserID      int            `json:"user_id"`
	RaidID      sql.NullString `json:"raid_id"`
	Description string         `json:"description"`
	YouTubeURL  string         `json:"youtube_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
	DeletedAt   sql.NullTime   `json:"deleted_at"`
	Score       int            `json:"score"`
}

// *************************************
// ********** Parsed from CSV **********
// *************************************

type RankData struct {
	UserID    int
	FinalRank int
	Score     int
	PartScore int
}
