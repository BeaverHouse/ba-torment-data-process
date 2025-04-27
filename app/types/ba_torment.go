package types

import (
	"database/sql"
	"time"
)

type BATormentPartyData struct {
	Filters       map[string][]int       `json:"filters"`
	AssistFilters map[string][]int       `json:"assist_filters"`
	MinPartys     int                    `json:"min_partys"`
	MaxPartys     int                    `json:"max_partys"`
	PartyDetail   []BATormentPartyDetail `json:"parties"`
}

type BATormentPartyDetail struct {
	FinalRank   int              `json:"FINAL_RANK"`
	TormentRank int              `json:"TORMENT_RANK"`
	Score       int64            `json:"SCORE"`
	UserID      int64            `json:"USER_ID"`
	Level       string           `json:"LEVEL"`
	PartyData   map[string][]int `json:"PARTY_DATA"`
}

type BATormentSummaryData struct {
	Torment BATormentLevelData `json:"torment"`
	Lunatic BATormentLevelData `json:"lunatic"`
}

type BATormentLevelData struct {
	ClearCount    int              `json:"clear_count"`
	PartyCounts   map[string][]int `json:"party_counts"`
	Filters       map[string][]int `json:"filters"`
	AssistFilters map[string][]int `json:"assist_filters"`
	Top5Partys    [][]interface{}  `json:"top5_partys"`
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
	UserID    int64
	FinalRank int
	Score     int64
	PartScore int64
}
