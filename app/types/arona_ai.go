package types

type AronaAIData struct {
	D []AronaAIDetail `json:"d"`
}

type AronaAIDetail struct {
	R int            `json:"r"` // 랭크
	S int            `json:"s"` // 점수
	T []AronaAIParty `json:"t"` // 파티 정보
}

type AronaAIParty struct {
	M []AronaAICharacter `json:"m"` // 스트라이커
	S []AronaAICharacter `json:"s"` // 서포터
}

type AronaAICharacter struct {
	StudentID  int  `json:"id"`         // 학생 ID
	Star       int  `json:"star"`       // 학생 성급
	HasWeapon  bool `json:"hasWeapon"`  // 무기 보유 여부
	WeaponStar int  `json:"weaponStar"` // 무기 성급
	IsAssist   bool `json:"isAssist"`   // 조력자 여부
}
