package schaledb

import (
	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
)

// Reads /data/<lang>/items.json file
func loadItems(log logger.Logger, lang string) (map[string]types.FavorItem, error) {
	url := fmt.Sprintf("%s/data/%s/items.json", constants.SchaleDBURL, lang)
	byteValue, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch(fmt.Sprintf("items (%s)", lang), err)
	}

	var items map[string]types.FavorItem
	if err := json.Unmarshal(byteValue, &items); err != nil {
		return nil, constants.ErrDataDecode(fmt.Sprintf("items (%s)", lang), err)
	}

	return items, nil
}

func ParseSchaleDBPresents(log logger.Logger, db *postgres.Queries) (map[string]*types.FavorItem, error) {
	items, err := loadItems(log, "kr")
	if err != nil {
		return nil, err
	}

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
