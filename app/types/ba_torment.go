package types

import (
	"database/sql"
	"time"
)

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
