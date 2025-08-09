package types

// Complete student data structure based on processed SchaleDB data
type StudentData struct {
	// Basic Info
	Name         string `json:"Name"`
	PersonalName string `json:"PersonalName"`
	FamilyName   string `json:"FamilyName"`
	DevName      string `json:"DevName"`

	// Character Details
	School              string `json:"School"`
	Club                string `json:"Club"`
	SchoolYear          string `json:"SchoolYear"`
	CharacterAge        string `json:"CharacterAge"`
	BirthDay            string `json:"BirthDay"`
	CharHeightMetric    string `json:"CharHeightMetric"`
	Hobby               string `json:"Hobby"`
	ProfileIntroduction string `json:"ProfileIntroduction"`

	// Battle Stats
	Position   string `json:"Position"`
	TacticRole string `json:"TacticRole"`
	SquadType  string `json:"SquadType"`
	BulletType string `json:"BulletType"`
	ArmorType  string `json:"ArmorType"`
	WeaponType string `json:"WeaponType"`
	Range      int    `json:"Range"`

	// Base Stats
	MaxHP1          int `json:"MaxHP1"`
	MaxHP100        int `json:"MaxHP100"`
	AttackPower1    int `json:"AttackPower1"`
	AttackPower100  int `json:"AttackPower100"`
	DefensePower1   int `json:"DefensePower1"`
	DefensePower100 int `json:"DefensePower100"`
	HealPower1      int `json:"HealPower1"`
	HealPower100    int `json:"HealPower100"`
	AccuracyPoint   int `json:"AccuracyPoint"`
	CriticalPoint   int `json:"CriticalPoint"`
	StabilityPoint  int `json:"StabilityPoint"`
	DodgePoint      int `json:"DodgePoint"`

	// Adaptation
	StreetBattleAdaptation  int `json:"StreetBattleAdaptation"`
	OutdoorBattleAdaptation int `json:"OutdoorBattleAdaptation"`
	IndoorBattleAdaptation  int `json:"IndoorBattleAdaptation"`

	// Equipment & Favor
	Equipment []string `json:"Equipment"`
	StarGrade int      `json:"StarGrade"`
	IsLimited []int    `json:"IsLimited"`

	// Favor System
	FavorStatType       []string `json:"FavorStatType"`
	FavorStatValue      [][]int  `json:"FavorStatValue"`
	FavorAlts           []int    `json:"FavorAlts"`
	FavorItemTags       []string `json:"FavorItemTags"`
	FavorItemUniqueTags []string `json:"FavorItemUniqueTags"`

	// Complex nested structures
	Skills               StudentSkills   `json:"Skills"`
	Weapon               StudentWeapon   `json:"Weapon"`
	Gear                 StudentGear     `json:"Gear"`
	Summons              []interface{}   `json:"Summons"`
	SearchTags           []string        `json:"SearchTags"`
	FurnitureInteraction [][]interface{} `json:"FurnitureInteraction"`
}

// Skills structure
type StudentSkills struct {
	Ex            *StudentSkill `json:"Ex,omitempty"`
	Passive       *StudentSkill `json:"Passive,omitempty"`
	ExtraPassive  *StudentSkill `json:"ExtraPassive,omitempty"`
	Public        *StudentSkill `json:"Public,omitempty"`
	GearPublic    *StudentSkill `json:"GearPublic,omitempty"`
	WeaponPassive *StudentSkill `json:"WeaponPassive,omitempty"`
}

// Individual skill
type StudentSkill struct {
	Cost    []int         `json:"Cost,omitempty"`
	Desc    string        `json:"Desc"`
	Effects []SkillEffect `json:"Effects"`
}

// Skill effect
type SkillEffect struct {
	Type                 string                `json:"Type"`
	Target               []string              `json:"Target,omitempty"`
	Scale                []int                 `json:"Scale,omitempty"`
	Value                [][]int               `json:"Value,omitempty"`
	AdditionalValue      interface{}           `json:"AdditionalValue,omitempty"`
	Channel              int                   `json:"Channel,omitempty"`
	Stat                 string                `json:"Stat,omitempty"`
	Duration             int                   `json:"Duration,omitempty"`
	Period               int                   `json:"Period,omitempty"`
	ApplyFrame           int                   `json:"ApplyFrame,omitempty"`
	Block                int                   `json:"Block,omitempty"`
	CriticalCheck        string                `json:"CriticalCheck,omitempty"`
	DescParamId          int                   `json:"DescParamId,omitempty"`
	Hits                 []int                 `json:"Hits,omitempty"`
	ExtraStatRate        []int                 `json:"ExtraStatRate,omitempty"`
	ExtraStatSource      string                `json:"ExtraStatSource,omitempty"`
	HpRateDamageModifier *HpRateDamageModifier `json:"HpRateDamageModifier,omitempty"`
}

// HP-based damage modifier
type HpRateDamageModifier struct {
	MaxHpRate     int     `json:"MaxHpRate"`
	MinHpRate     int     `json:"MinHpRate"`
	MultiplierMax float64 `json:"MultiplierMax"`
	MultiplierMin float64 `json:"MultiplierMin"`
}

// Student weapon
type StudentWeapon struct {
	Name            string `json:"Name"`
	Desc            string `json:"Desc"`
	AdaptationType  string `json:"AdaptationType"`
	AdaptationValue int    `json:"AdaptationValue"`
	AttackPower1    int    `json:"AttackPower1"`
	AttackPower100  int    `json:"AttackPower100"`
	MaxHP1          int    `json:"MaxHP1"`
	MaxHP100        int    `json:"MaxHP100"`
	HealPower1      int    `json:"HealPower1"`
	HealPower100    int    `json:"HealPower100"`
	StatLevelUpType string `json:"StatLevelUpType"`
}

// Student gear
type StudentGear struct {
	Name      string   `json:"Name"`
	StatType  []string `json:"StatType"`
	StatValue [][]int  `json:"StatValue"`
}
