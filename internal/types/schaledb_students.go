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
	Equipment    []string `json:"Equipment"`
	StarGrade    int      `json:"StarGrade"`
	IsLimited    []int    `json:"IsLimited"`    // 0:통상, 1:한정, 2:배포, 3:페스
	IsReleased   []bool   `json:"IsReleased"`   // [JP, Global, CN] 서버별 출시 여부
	DefaultOrder int      `json:"DefaultOrder"` // 출시 순서 (높을수록 신캐)

	// Skill level-up materials: item IDs and amounts per level step.
	// EX uses SkillExMaterial (lv2-5), the other three skills share
	// SkillMaterial (lv2-9). PotentialMaterial is the potential-unlock item.
	SkillMaterial         [][]int `json:"SkillMaterial,omitempty"`
	SkillMaterialAmount   [][]int `json:"SkillMaterialAmount,omitempty"`
	SkillExMaterial       [][]int `json:"SkillExMaterial,omitempty"`
	SkillExMaterialAmount [][]int `json:"SkillExMaterialAmount,omitempty"`
	PotentialMaterial     int     `json:"PotentialMaterial,omitempty"`

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
	Normal        *StudentSkill `json:"Normal,omitempty"`
	Ex            *StudentSkill `json:"Ex,omitempty"`
	Passive       *StudentSkill `json:"Passive,omitempty"`
	ExtraPassive  *StudentSkill `json:"ExtraPassive,omitempty"`
	Public        *StudentSkill `json:"Public,omitempty"`
	GearPublic    *StudentSkill `json:"GearPublic,omitempty"`
	WeaponPassive *StudentSkill `json:"WeaponPassive,omitempty"`
}

// Individual skill. Name locales are injected from the jp/en/zh SchaleDB
// datasets after parsing; Parameters keeps the raw per-level values so
// downstream consumers can reason about level breakpoints, while Desc bakes
// in the max-level values only.
type StudentSkill struct {
	Name        string        `json:"Name,omitempty"`
	NameJa      string        `json:"NameJa,omitempty"`
	NameEn      string        `json:"NameEn,omitempty"`
	NameZh      string        `json:"NameZh,omitempty"`
	Cost        []int         `json:"Cost,omitempty"`
	Desc        string        `json:"Desc"`
	Parameters  [][]string    `json:"Parameters,omitempty"`
	Duration    int           `json:"Duration,omitempty"`
	Range       int           `json:"Range,omitempty"`
	Effects     []SkillEffect `json:"Effects"`
	ExtraSkills []ExtraSkill  `json:"ExtraSkills,omitempty"`
}

// Extra skill for selectable EX skills
type ExtraSkill struct {
	Id         string        `json:"Id"`
	Name       string        `json:"Name"`
	NameJa     string        `json:"NameJa,omitempty"`
	NameEn     string        `json:"NameEn,omitempty"`
	NameZh     string        `json:"NameZh,omitempty"`
	Desc       string        `json:"Desc"`
	Parameters [][]string    `json:"Parameters,omitempty"`
	Cost       []int         `json:"Cost,omitempty"`
	Duration   int           `json:"Duration,omitempty"`
	Range      int           `json:"Range,omitempty"`
	Icon       string        `json:"Icon,omitempty"`
	Effects    []SkillEffect `json:"Effects"`
	Radius     []any         `json:"Radius,omitempty"`
}

// Skill effect
type SkillEffect struct {
	Type                 string                `json:"Type"`
	Restrictions         []SkillRestriction    `json:"Restrictions,omitempty"`
	Target               []string              `json:"Target,omitempty"`
	Scale                []int                 `json:"Scale,omitempty"`
	Value                [][]float64           `json:"Value,omitempty"`
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
	NameJa          string `json:"NameJa,omitempty"`
	NameEn          string `json:"NameEn,omitempty"`
	NameZh          string `json:"NameZh,omitempty"`
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
	NameJa    string   `json:"NameJa,omitempty"`
	NameEn    string   `json:"NameEn,omitempty"`
	NameZh    string   `json:"NameZh,omitempty"`
	Desc      string   `json:"Desc,omitempty"`
	StatType  []string `json:"StatType"`
	StatValue [][]int  `json:"StatValue"`
}
