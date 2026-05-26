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
	// zh is non-fatal: SchaleDB occasionally lags on Chinese translations.
	// Missing zh values leave name_zh as DEFAULT '' and i18n.Get() falls back
	// to ko/en when looking up.
	zh, err := loadLocalizationFull("zh")
	if err != nil {
		ui.Log.Warn("Failed to load Chinese localization (non-fatal)", logger.F("error", err))
		zh = &localizationRaw{School: map[string]string{}, Club: map[string]string{}}
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
			NameZh:   zh.School[key],
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
			NameZh:   zh.Club[key],
		})
		if err != nil {
			return err
		}
		ui.Log.Info("Saved i18n club", logger.F("key", key))
	}

	return nil
}
