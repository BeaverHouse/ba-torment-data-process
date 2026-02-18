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
	Skills               StudentSkills `json:"Skills"`
	Weapon               StudentWeapon `json:"Weapon"`
	Gear                 StudentGear   `json:"Gear"`
	Summons              []any         `json:"Summons"`
	SearchTags           []string      `json:"SearchTags"`
	FurnitureInteraction [][]any       `json:"FurnitureInteraction"`
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
	Cost        []int          `json:"Cost,omitempty"`
	Desc        string         `json:"Desc"`
	Effects     []SkillEffect  `json:"Effects"`
	ExtraSkills []ExtraSkill   `json:"ExtraSkills,omitempty"`
}

// Extra skill for selectable EX skills
type ExtraSkill struct {
	Id       string        `json:"Id"`
	Name     string        `json:"Name"`
	Desc     string        `json:"Desc"`
	Cost     []int         `json:"Cost,omitempty"`
	Duration int           `json:"Duration,omitempty"`
	Range    int           `json:"Range,omitempty"`
	Icon     string        `json:"Icon,omitempty"`
	Effects  []SkillEffect `json:"Effects"`
	Radius   []any         `json:"Radius,omitempty"`
}

// Skill effect
type SkillEffect struct {
	Type                 string                `json:"Type"`
	Restrictions         []SkillRestriction    `json:"Restrictions,omitempty"`
	Target               []string              `json:"Target,omitempty"`
	Scale                []int                 `json:"Scale,omitempty"`
	Value                [][]int               `json:"Value,omitempty"`
	AdditionalValue      any                   `json:"AdditionalValue,omitempty"`
	Stat                 string                `json:"Stat,omitempty"`
	Duration             int                   `json:"Duration,omitempty"`
	Period               int                   `json:"Period,omitempty"`
	Block                int                   `json:"Block,omitempty"`
	CriticalCheck        string                `json:"CriticalCheck,omitempty"`
	Hits                 []int                 `json:"Hits,omitempty"`
	ExtraStatRate        []int                 `json:"ExtraStatRate,omitempty"`
	ExtraStatSource      string                `json:"ExtraStatSource,omitempty"`
	TargetHpRateModifier *TargetHpRateModifier `json:"TargetHpRateModifier,omitempty"`
}

// Skill restriction (e.g., BulletType == Pierce)
type SkillRestriction struct {
	Property string `json:"Property"`
	Operand  string `json:"Operand"`
	Value    any    `json:"Value"` // Can be string, number, or array
}

// HP-based damage modifier
type TargetHpRateModifier struct {
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
