package schaledb

import (
	"context"
	"encoding/json"
	"log"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"
)

type localizationRaw struct {
	School map[string]string `json:"School"`
	Club   map[string]string `json:"Club"`
}

func loadLocalizationFull(lang string) *localizationRaw {
	url := constants.SchaleDBURL + "data/" + lang + "/localization.min.json"
	byteValue := storage.GetDataFromURL(url)

	var data localizationRaw
	if err := json.Unmarshal(byteValue, &data); err != nil {
		log.Fatalf("Failed to unmarshal localization (%s): %v", lang, err)
	}

	return &data
}

func SaveI18nData(db *postgres.Queries) error {
	kr := loadLocalizationFull("kr")
	ja := loadLocalizationFull("jp")
	en := loadLocalizationFull("en")

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
		log.Printf("Saved i18n school: %s", key)
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
		log.Printf("Saved i18n club: %s", key)
	}

	return nil
}
