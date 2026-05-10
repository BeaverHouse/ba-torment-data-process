package cmd

import (
	"context"
	"fmt"
	"time"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/analysis"
	"ba-torment-data-process/internal/logic/party"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/types"
	"ba-torment-data-process/internal/ui"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

type raidListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TopLevel     string `json:"top_level"`
	PartyUpdated bool   `json:"party_updated"`
}

var processRaidCmd = &cobra.Command{
	Use:   "process-raid",
	Short: "Process all raid content data",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		recent, _ := cmd.Flags().GetInt("recent")

		pool := gopostgres.InitFromEnv()
		defer pool.Close()

		queries := postgres.New(pool)

		contents, err := queries.ListContentsForRaidList(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list contents: %w", err)
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
				processErr = processCachedRaid(contentID, dryRun)
			} else {
				processErr = processFullRaid(queries, contentID, dryRun)
			}
			if processErr != nil {
				ui.Log.Warn("Raid processing failed", logger.F("contentID", contentID), logger.F("error", processErr))
				partyUpdated[contentID] = false
				continue
			}
			partyUpdated[contentID] = true
			ui.Log.Info("Successfully processed content", logger.F("contentID", contentID))
		}

		ui.Log.Info("Generating raids.json")
		raidList := make([]raidListItem, 0, len(contents))
		for _, content := range contents {
			raidList = append(raidList, raidListItem{
				ID:           content.ContentID,
				Name:         content.Title,
				TopLevel:     string(content.TopLevel),
				PartyUpdated: partyUpdated[content.ContentID],
			})
		}

		if err := storage.MarshalAndUpload(raidList, "batorment/v3", "raids.json", dryRun, "Raids list uploaded"); err != nil {
			return fmt.Errorf("failed to upload raids.json: %w", err)
		}

		ui.Log.Info("Successfully processed all raids")
		return nil
	},
}

// processCachedRaid refreshes video refs against an existing party JSON in S3
// and re-uploads it together with the video filter. Used for older raids whose
// party data is static — DuckDB parsing is skipped entirely.
//
// Any failure short-circuits and returns an error so the caller can mark the
// raid as not-updated. Partial uploads may remain in S3, but raids.json will
// reflect the failure honestly.
func processCachedRaid(contentID string, dryRun bool) error {
	ui.Log.Info("Updating video refs only (cached)", logger.F("contentID", contentID))

	partyData := analysis.DownloadPartyData(contentID)
	if partyData == nil {
		return fmt.Errorf("no cached party data found")
	}

	fileName := fmt.Sprintf("%s.json", contentID)

	if err := refreshVideoRefsAndUpload(partyData, contentID, fileName, dryRun); err != nil {
		return err
	}

	return nil
}

// processFullRaid runs the full DuckDB-backed pipeline: parse, refresh video
// refs, upload party + filters, and produce the summary aggregate.
func processFullRaid(queries *postgres.Queries, contentID string, dryRun bool) error {
	ui.Log.Info("Full processing (DuckDB)", logger.F("contentID", contentID))

	contentInfo, err := queries.GetContentByID(context.Background(), contentID)
	if err != nil {
		return fmt.Errorf("failed to get content info: %w", err)
	}

	ui.Log.Info("[1/6] Parsing DuckDB", logger.F("contentID", contentID))
	partyData, filterResult, err := party.ParseDuckDB(contentID, contentInfo.StartDate.Time)
	if err != nil {
		return fmt.Errorf("failed to parse DuckDB: %w", err)
	}

	fileName := fmt.Sprintf("%s.json", contentID)

	ui.Log.Info("[2/6] Updating video references and uploading party + video filter", logger.F("contentID", contentID))
	if err := refreshVideoRefsAndUpload(partyData, contentID, fileName, dryRun); err != nil {
		return err
	}

	ui.Log.Info("[3/6] Uploading additional filters", logger.F("contentID", contentID))
	if err := storage.MarshalAndUpload(filterResult, "batorment/v3/filter", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload filter: %w", err)
	}

	lunaticFilter := party.CreateLunaticFilter(partyData)
	if err := storage.MarshalAndUpload(lunaticFilter, "batorment/v3/lunatic-filter", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload lunatic filter: %w", err)
	}

	nonLunaticFilter := party.CreateNonLunaticFilter(partyData)
	if err := storage.MarshalAndUpload(nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload non-lunatic filter: %w", err)
	}

	ui.Log.Info("[4/6] Building summary data", logger.F("contentID", contentID))
	summaryData, err := party.ProcessPartyDataToSummaryData(partyData)
	if err != nil {
		return fmt.Errorf("failed to process summary data: %w", err)
	}

	ui.Log.Info("[5/6] Annotating summary with platinum cuts and character stats", logger.F("contentID", contentID))
	annotateSummary(summaryData, partyData, contentID, contentInfo.StartDate.Time)

	ui.Log.Info("[6/6] Uploading summary data", logger.F("contentID", contentID))
	if err := storage.MarshalAndUpload(summaryData, "batorment/v3/summary", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload summary data: %w", err)
	}

	return nil
}

// refreshVideoRefsAndUpload is the shared step between cached and full
// pipelines: update video refs in partyData, upload it, then build and upload
// the video filter. Video-ref update failures are logged but not fatal — the
// party JSON itself is still authoritative and must succeed.
func refreshVideoRefsAndUpload(partyData *types.BATormentPartyData, contentID, fileName string, dryRun bool) error {
	updated, err := party.UpdateVideoRefWithData(partyData, contentID)
	if err != nil {
		ui.Log.Warn("Failed to update video refs", logger.F("contentID", contentID), logger.F("error", err))
	} else {
		ui.Log.Info("Updated video references", logger.F("count", updated), logger.F("contentID", contentID))
	}

	if err := storage.MarshalAndUpload(partyData, "batorment/v3/party", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload party data: %w", err)
	}

	videoFilter := party.CreateVideoFilter(contentID)
	if videoFilter == nil {
		ui.Log.Warn("No video filter created", logger.F("contentID", contentID))
		return nil
	}
	if err := storage.MarshalAndUpload(videoFilter, "batorment/v3/video-filter", fileName, dryRun, ""); err != nil {
		return fmt.Errorf("failed to upload video filter: %w", err)
	}
	return nil
}

// annotateSummary fills in the aggregate fields (platinum cuts, essential and
// high-impact characters, min-UE / max-party users). Each subfield is logged
// but a per-field failure is non-fatal — the summary is still uploaded with
// whatever was successfully computed.
func annotateSummary(summaryData *types.BATormentSummaryData, partyData *types.BATormentPartyData, contentID string, startDate time.Time) {
	platinumCuts, err := party.GetPlatinumCuts(contentID, startDate)
	if err != nil {
		ui.Log.Warn("Failed to get platinum cuts", logger.F("contentID", contentID), logger.F("error", err))
	} else {
		summaryData.PlatinumCuts = platinumCuts
		ui.Log.Info("Added platinum cuts", logger.F("count", len(platinumCuts)), logger.F("contentID", contentID))
	}

	if party.IsGrandAssault(contentID) {
		partPlatinumCuts := party.GetPartPlatinumCutsFromPartyData(partyData)
		if len(partPlatinumCuts) > 0 {
			summaryData.PartPlatinumCuts = partPlatinumCuts
			ui.Log.Info("Added part platinum cuts", logger.F("count", len(partPlatinumCuts)), logger.F("contentID", contentID))
		}
	}

	essentialTorment, essentialLunatic := party.GetEssentialCharacters(partyData)
	summaryData.Torment.EssentialCharacters = essentialTorment
	summaryData.Lunatic.EssentialCharacters = essentialLunatic
	ui.Log.Info("Essential characters", logger.F("contentID", contentID), logger.F("torment", len(essentialTorment)), logger.F("lunatic", len(essentialLunatic)))

	highImpactTorment, highImpactLunatic := party.GetHighImpactCharacters(partyData)
	summaryData.Torment.HighImpactCharacters = highImpactTorment
	summaryData.Lunatic.HighImpactCharacters = highImpactLunatic
	ui.Log.Info("High impact characters", logger.F("contentID", contentID), logger.F("torment", len(highImpactTorment)), logger.F("lunatic", len(highImpactLunatic)))

	minUETorment, minUELunatic := party.GetMinUEUsers(partyData)
	summaryData.Torment.MinUEUser = minUETorment
	summaryData.Lunatic.MinUEUser = minUELunatic
	if minUETorment != nil {
		ui.Log.Info("Min UE user (Torment)", logger.F("contentID", contentID), logger.F("rank", minUETorment.Rank), logger.F("ueCount", minUETorment.UECount), logger.F("parties", len(minUETorment.PartyData)))
	}
	if minUELunatic != nil {
		ui.Log.Info("Min UE user (Lunatic)", logger.F("contentID", contentID), logger.F("rank", minUELunatic.Rank), logger.F("ueCount", minUELunatic.UECount), logger.F("parties", len(minUELunatic.PartyData)))
	}

	maxPartyTorment, maxPartyLunatic := party.GetMaxPartyUsers(partyData)
	summaryData.Torment.MaxPartyUser = maxPartyTorment
	summaryData.Lunatic.MaxPartyUser = maxPartyLunatic
	if maxPartyTorment != nil {
		ui.Log.Info("Max party user (Torment)", logger.F("contentID", contentID), logger.F("rank", maxPartyTorment.Rank), logger.F("parties", len(maxPartyTorment.PartyData)))
	}
	if maxPartyLunatic != nil {
		ui.Log.Info("Max party user (Lunatic)", logger.F("contentID", contentID), logger.F("rank", maxPartyLunatic.Rank), logger.F("parties", len(maxPartyLunatic.PartyData)))
	}
}

func init() {
	processRaidCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	processRaidCmd.Flags().Int("recent", 5, "Number of recent raids to fully process from DuckDB (older ones use cached S3 data)")
	rootCmd.AddCommand(processRaidCmd)
}
