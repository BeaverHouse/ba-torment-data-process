package schaledb

import (
	"context"
	"encoding/json"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"

	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

type localizationRaw struct {
	School map[string]string `json:"School"`
	Club   map[string]string `json:"Club"`
}

func loadLocalizationFull(log logger.Logger, lang string) (*localizationRaw, error) {
	url := constants.SchaleDBURL + "data/" + lang + "/localization.min.json"
	byteValue, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch("localization ("+lang+")", err)
	}

	var data localizationRaw
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, constants.ErrDataDecode("localization ("+lang+")", err)
	}

	return &data, nil
}

func SaveI18nData(log logger.Logger, db *postgres.Queries) error {
	kr, err := loadLocalizationFull(log, "kr")
	if err != nil {
		return err
	}
	ja, err := loadLocalizationFull(log, "jp")
	if err != nil {
		return err
	}
	en, err := loadLocalizationFull(log, "en")
	if err != nil {
		return err
	}
	// zh is non-fatal: SchaleDB occasionally lags on Chinese translations.
	// Missing zh values leave name_zh as DEFAULT '' and i18n.Get() falls back
	// to ko/en when looking up.
	zh, err := loadLocalizationFull(log, "zh")
	if err != nil {
		log.Warn("Failed to load Chinese localization (non-fatal)", logger.Field{Key: "error", Value: err})
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
			return errorhandle.ErrDBOperation("upsert i18n school", err)
		}
		log.Info("Saved i18n school", logger.Field{Key: "key", Value: key})
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
			return errorhandle.ErrDBOperation("upsert i18n club", err)
		}
		log.Info("Saved i18n club", logger.Field{Key: "key", Value: key})
	}

	return nil
}
