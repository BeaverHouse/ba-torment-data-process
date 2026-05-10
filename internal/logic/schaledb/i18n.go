package schaledb

import (
	"context"
	"encoding/json"
	"fmt"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/ui"

	"github.com/BeaverHouse/go-common/logger"
)

type localizationRaw struct {
	School map[string]string `json:"School"`
	Club   map[string]string `json:"Club"`
}

func loadLocalizationFull(lang string) (*localizationRaw, error) {
	url := constants.SchaleDBURL + "data/" + lang + "/localization.min.json"
	byteValue, err := storage.GetDataFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch localization (%s): %w", lang, err)
	}

	var data localizationRaw
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal localization (%s): %w", lang, err)
	}

	return &data, nil
}

func SaveI18nData(db *postgres.Queries) error {
	kr, err := loadLocalizationFull("kr")
	if err != nil {
		return err
	}
	ja, err := loadLocalizationFull("jp")
	if err != nil {
		return err
	}
	en, err := loadLocalizationFull("en")
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Save School
	for key := range kr.School {
		err := db.UpsertI18n(ctx, postgres.UpsertI18nParams{
			Category: "school",
			Key:      key,
			NameKo:   kr.School[key],
			NameJa:   ja.School[key],
			NameEn:   en.School[key],
		})
		if err != nil {
			return err
		}
		ui.Log.Info("Saved i18n school", logger.F("key", key))
	}

	// Save Club
	for key := range kr.Club {
		err := db.UpsertI18n(ctx, postgres.UpsertI18nParams{
			Category: "club",
			Key:      key,
			NameKo:   kr.Club[key],
			NameJa:   ja.Club[key],
			NameEn:   en.Club[key],
		})
		if err != nil {
			return err
		}
		ui.Log.Info("Saved i18n club", logger.F("key", key))
	}

	return nil
}
