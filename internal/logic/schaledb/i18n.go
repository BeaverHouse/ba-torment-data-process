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

// itemName is the subset of items.json needed to resolve material IDs
// (students.detail stores skill materials as bare IDs).
type itemName struct {
	Id       int    `json:"Id"`
	Name     string `json:"Name"`
	Category string `json:"Category"`
}

func loadItemNames(log logger.Logger, lang string) (map[string]itemName, error) {
	url := constants.SchaleDBURL + "data/" + lang + "/items.min.json"
	byteValue, err := storage.GetDataFromURL(log, url)
	if err != nil {
		return nil, constants.ErrDataFetch("items ("+lang+")", err)
	}

	var data map[string]itemName
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, constants.ErrDataDecode("items ("+lang+")", err)
	}

	return data, nil
}

// saveMaterialNames stores Material item names under the "item" i18n category
// so consumers can turn skill-material IDs into names without fetching
// SchaleDB. Non-kr locales are best-effort.
func saveMaterialNames(ctx context.Context, log logger.Logger, db *postgres.Queries) error {
	kr, err := loadItemNames(log, "kr")
	if err != nil {
		return err
	}
	locales := map[string]map[string]itemName{}
	for _, lang := range []string{"jp", "en", "zh"} {
		items, err := loadItemNames(log, lang)
		if err != nil {
			log.Warn("Failed to load item names (non-fatal)", logger.Field{Key: "lang", Value: lang}, logger.Field{Key: "error", Value: err})
			items = map[string]itemName{}
		}
		locales[lang] = items
	}

	saved := 0
	for key, item := range kr {
		if item.Category != "Material" {
			continue
		}
		err := db.UpsertI18n(ctx, postgres.UpsertI18nParams{
			Category: "item",
			Key:      key,
			NameKo:   item.Name,
			NameJa:   locales["jp"][key].Name,
			NameEn:   locales["en"][key].Name,
			NameZh:   locales["zh"][key].Name,
		})
		if err != nil {
			return errorhandle.ErrDBOperation("upsert i18n item", err)
		}
		saved++
	}
	log.Info("Saved i18n materials", logger.Field{Key: "count", Value: saved})

	return nil
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

	return saveMaterialNames(ctx, log, db)
}
