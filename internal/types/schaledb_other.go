package types

type FavorItem struct {
	Id       int      `json:"Id"`
	Category string   `json:"Category"`
	Rarity   string   `json:"Rarity"`
	Tags     []string `json:"Tags"`
	ExpValue int      `json:"ExpValue"`
	Name     string   `json:"Name"`
}
