package party

import (
	"context"
	"fmt"
	"time"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/analysis"
	"ba-torment-data-process/internal/logic/raidname"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

type raidListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"` // legacy = name_ko; FE migrating to per-locale fields below
	NameKO       string `json:"name_ko"`
	NameEN       string `json:"name_en"`
	NameZH       string `json:"name_zh"`
	TopLevel     string `json:"top_level"`
	PartyUpdated bool   `json:"party_updated"`
}

// ProcessAllRaids orchestrates the full raid-processing run: it lists raid
// content from the database, processes each raid (full DuckDB pipeline for the
// most recent `recent` raids, cached S3 refresh for older ones), and uploads
// the aggregate raids.json. The CLI adapter only triggers it and reports
// completion.
func ProcessAllRaids(log logger.Logger, dryRun bool, recent int) error {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()

	queries := postgres.New(pool)

	contents, err := queries.ListContentsForRaidList(context.Background())
	if err != nil {
		return errorhandle.ErrDBOperation("list contents for raid list", err)
	}

	recentStart := len(contents) - recent
	if recentStart < 0 {
		recentStart = 0
	}

	partyUpdated := make(map[string]bool)
	for i, content := range contents {
		contentID := content.ContentID
		var processErr error
		if i < recentStart {
			processErr = processCachedRaid(log, contentID, dryRun)
		} else {
			processErr = processFullRaid(log, queries, contentID, dryRun)
		}
		if processErr != nil {
			log.Warn("Raid processing failed", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: processErr})
			partyUpdated[contentID] = false
			continue
		}
		partyUpdated[contentID] = true
		log.Info("Successfully processed content", logger.Field{Key: "contentID", Value: contentID})
	}

	log.Info("Generating raids.json")
	raidList := make([]raidListItem, 0, len(contents))
	for _, content := range contents {
		raidList = append(raidList, raidListItem{
			ID:           content.ContentID,
			Name:         content.Title,
			NameKO:       content.Title,
			NameEN:       raidname.Translate(content.Title, raidname.LangEN),
			NameZH:       raidname.Translate(content.Title, raidname.LangZH),
			TopLevel:     string(content.TopLevel),
			PartyUpdated: partyUpdated[content.ContentID],
		})
	}

	if err := storage.MarshalAndUpload(log, raidList, "batorment/v3", "raids.json", dryRun, "Raids list uploaded"); err != nil {
		return err
	}

	return nil
}

// processCachedRaid refreshes video refs against an existing party JSON in S3
// and re-uploads it together with the video filter. Used for older raids whose
// party data is static — DuckDB parsing is skipped entirely.
//
// Any failure short-circuits and returns an error so the caller can mark the
// raid as not-updated. Partial uploads may remain in S3, but raids.json will
// reflect the failure honestly.
func processCachedRaid(log logger.Logger, contentID string, dryRun bool) error {
	log.Info("Updating video refs only (cached)", logger.Field{Key: "contentID", Value: contentID})

	partyData := analysis.DownloadPartyData(log, contentID)
	if partyData == nil {
		return errorhandle.ErrNotFound(fmt.Sprintf("cached party data for %s", contentID))
	}

	fileName := fmt.Sprintf("%s.json", contentID)

	if err := refreshVideoRefsAndUpload(log, partyData, contentID, fileName, dryRun); err != nil {
		return err
	}

	return nil
}

// processFullRaid runs the full DuckDB-backed pipeline: parse, refresh video
// refs, upload party + filters, and produce the summary aggregate.
func processFullRaid(log logger.Logger, queries *postgres.Queries, contentID string, dryRun bool) error {
	log.Info("Full processing (DuckDB)", logger.Field{Key: "contentID", Value: contentID})

	contentInfo, err := queries.GetContentByID(context.Background(), contentID)
	if err != nil {
		return errorhandle.ErrDBOperation("get content by id", err)
	}

	log.Info("[1/6] Parsing DuckDB", logger.Field{Key: "contentID", Value: contentID})
	partyData, filterResult, err := ParseDuckDB(log, contentID, contentInfo.StartDate.Time)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s.json", contentID)

	log.Info("[2/6] Updating video references and uploading party + video filter", logger.Field{Key: "contentID", Value: contentID})
	if err := refreshVideoRefsAndUpload(log, partyData, contentID, fileName, dryRun); err != nil {
		return err
	}

	log.Info("[3/6] Uploading additional filters", logger.Field{Key: "contentID", Value: contentID})
	if err := storage.MarshalAndUpload(log, filterResult, "batorment/v3/filter", fileName, dryRun, ""); err != nil {
		return err
	}

	lunaticFilter := CreateLunaticFilter(partyData)
	if err := storage.MarshalAndUpload(log, lunaticFilter, "batorment/v3/lunatic-filter", fileName, dryRun, ""); err != nil {
		return err
	}

	nonLunaticFilter := CreateNonLunaticFilter(partyData)
	if err := storage.MarshalAndUpload(log, nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, dryRun, ""); err != nil {
		return err
	}

	log.Info("[4/6] Building summary data", logger.Field{Key: "contentID", Value: contentID})
	summaryData, err := ProcessPartyDataToSummaryData(partyData)
	if err != nil {
		return err
	}

	log.Info("[5/6] Annotating summary with platinum cuts and character stats", logger.Field{Key: "contentID", Value: contentID})
	annotateSummary(log, summaryData, partyData, contentID, contentInfo.StartDate.Time)

	log.Info("[6/6] Uploading summary data", logger.Field{Key: "contentID", Value: contentID})
	if err := storage.MarshalAndUpload(log, summaryData, "batorment/v3/summary", fileName, dryRun, ""); err != nil {
		return err
	}

	return nil
}

// refreshVideoRefsAndUpload is the shared step between cached and full
// pipelines: update video refs in partyData, upload it, then build and upload
// the video filter. Video-ref update failures are logged but not fatal — the
// party JSON itself is still authoritative and must succeed.
func refreshVideoRefsAndUpload(log logger.Logger, partyData *types.BATormentPartyData, contentID, fileName string, dryRun bool) error {
	updated, err := UpdateVideoRefWithData(log, partyData, contentID)
	if err != nil {
		log.Warn("Failed to update video refs", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
	} else {
		log.Info("Updated video references", logger.Field{Key: "count", Value: updated}, logger.Field{Key: "contentID", Value: contentID})
	}

	if err := storage.MarshalAndUpload(log, partyData, "batorment/v3/party", fileName, dryRun, ""); err != nil {
		return err
	}

	videoFilter := CreateVideoFilter(log, contentID)
	if videoFilter == nil {
		log.Warn("No video filter created", logger.Field{Key: "contentID", Value: contentID})
		return nil
	}
	if err := storage.MarshalAndUpload(log, videoFilter, "batorment/v3/video-filter", fileName, dryRun, ""); err != nil {
		return err
	}
	return nil
}

// annotateSummary fills in the aggregate fields (platinum cuts, essential and
// high-impact characters, min-UE / max-party users). Each subfield is logged
// but a per-field failure is non-fatal — the summary is still uploaded with
// whatever was successfully computed.
func annotateSummary(log logger.Logger, summaryData *types.BATormentSummaryData, partyData *types.BATormentPartyData, contentID string, startDate time.Time) {
	platinumCuts, err := GetPlatinumCuts(log, contentID, startDate)
	if err != nil {
		log.Warn("Failed to get platinum cuts", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "error", Value: err})
	} else {
		summaryData.PlatinumCuts = platinumCuts
		log.Info("Added platinum cuts", logger.Field{Key: "count", Value: len(platinumCuts)}, logger.Field{Key: "contentID", Value: contentID})
	}

	if IsGrandAssault(contentID) {
		partPlatinumCuts := GetPartPlatinumCutsFromPartyData(partyData)
		if len(partPlatinumCuts) > 0 {
			summaryData.PartPlatinumCuts = partPlatinumCuts
			log.Info("Added part platinum cuts", logger.Field{Key: "count", Value: len(partPlatinumCuts)}, logger.Field{Key: "contentID", Value: contentID})
		}
	}

	essentialTorment, essentialLunatic := GetEssentialCharacters(partyData)
	summaryData.Torment.EssentialCharacters = essentialTorment
	summaryData.Lunatic.EssentialCharacters = essentialLunatic
	log.Info("Essential characters", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "torment", Value: len(essentialTorment)}, logger.Field{Key: "lunatic", Value: len(essentialLunatic)})

	highImpactTorment, highImpactLunatic := GetHighImpactCharacters(partyData)
	summaryData.Torment.HighImpactCharacters = highImpactTorment
	summaryData.Lunatic.HighImpactCharacters = highImpactLunatic
	log.Info("High impact characters", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "torment", Value: len(highImpactTorment)}, logger.Field{Key: "lunatic", Value: len(highImpactLunatic)})

	minUETorment, minUELunatic := GetMinUEUsers(partyData)
	summaryData.Torment.MinUEUser = minUETorment
	summaryData.Lunatic.MinUEUser = minUELunatic
	if minUETorment != nil {
		log.Info("Min UE user (Torment)", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "rank", Value: minUETorment.Rank}, logger.Field{Key: "ueCount", Value: minUETorment.UECount}, logger.Field{Key: "parties", Value: len(minUETorment.PartyData)})
	}
	if minUELunatic != nil {
		log.Info("Min UE user (Lunatic)", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "rank", Value: minUELunatic.Rank}, logger.Field{Key: "ueCount", Value: minUELunatic.UECount}, logger.Field{Key: "parties", Value: len(minUELunatic.PartyData)})
	}

	maxPartyTorment, maxPartyLunatic := GetMaxPartyUsers(partyData)
	summaryData.Torment.MaxPartyUser = maxPartyTorment
	summaryData.Lunatic.MaxPartyUser = maxPartyLunatic
	if maxPartyTorment != nil {
		log.Info("Max party user (Torment)", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "rank", Value: maxPartyTorment.Rank}, logger.Field{Key: "parties", Value: len(maxPartyTorment.PartyData)})
	}
	if maxPartyLunatic != nil {
		log.Info("Max party user (Lunatic)", logger.Field{Key: "contentID", Value: contentID}, logger.Field{Key: "rank", Value: maxPartyLunatic.Rank}, logger.Field{Key: "parties", Value: len(maxPartyLunatic.PartyData)})
	}
}
