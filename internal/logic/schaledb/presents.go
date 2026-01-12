package schaledb

import (
	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Reads /data/<lang>/items.json file
func loadItems(lang string) map[string]types.FavorItem {
	url := fmt.Sprintf("%s/data/%s/items.json", constants.SchaleDBURL, lang)
	byteValue := storage.GetDataFromURL(url)

	var items map[string]types.FavorItem
	err := json.Unmarshal(byteValue, &items)
	if err != nil {
		log.Fatalf("Failed to unmarshal localization: %v", err)
	}

	return items
}

func ParseSchaleDBPresents(db *postgres.Queries) (map[string]*types.FavorItem, error) {
	items := loadItems("kr")

	for _, item := range items {
		db.InsertPresent(context.Background(), postgres.InsertPresentParams{
			PresentID: int32(item.Id),
			NameKo:    item.Name,
			Rarity:    item.Rarity,
			Tags:      item.Tags,
			ExpValue:  int32(item.ExpValue),
		})
	}

	return nil, nil
}
