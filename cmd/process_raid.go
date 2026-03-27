package cmd

import (
	"context"
	"fmt"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/party"
	"ba-torment-data-process/internal/logic/storage"
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

		pool := gopostgres.InitFromEnv()
		defer pool.Close()

		queries := postgres.New(pool)

		contents, err := queries.ListContentsForRaidList(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list contents: %w", err)
		}

		partyUpdated := make(map[string]bool)

		for _, content := range contents {
			contentID := content.ContentID
			ui.Log.Info("Processing content", logger.F("contentID", contentID))

			contentInfo, err := queries.GetContentByID(context.Background(), contentID)
			if err != nil {
				return fmt.Errorf("failed to get content info: %w", err)
			}

			ui.Log.Info("[1/6] Parsing DuckDB", logger.F("contentID", contentID))
			partyData, filterResult, err := party.ParseDuckDB(contentID, contentInfo.StartDate.Time)
			if err != nil {
				ui.Log.Warn("Skipping content", logger.F("contentID", contentID), logger.F("error", err))
				partyUpdated[contentID] = false
				continue
			}
			partyUpdated[contentID] = true

			fileName := fmt.Sprintf("%s.json", contentID)

			ui.Log.Info("[2/6] Updating video references", logger.F("contentID", contentID))
			updated, err := party.UpdateVideoRefWithData(partyData, contentID)
			if err != nil {
				ui.Log.Warn("Failed to update video refs", logger.F("contentID", contentID), logger.F("error", err))
			} else {
				ui.Log.Info("Updated video references", logger.F("count", updated), logger.F("contentID", contentID))
			}

			ui.Log.Info("[3/6] Uploading party data", logger.F("contentID", contentID))
			if err := storage.MarshalAndUpload(partyData, "batorment/v3/party", fileName, dryRun, ""); err != nil {
				ui.Log.Warn("Failed to upload party data", logger.F("error", err))
				continue
			}

			ui.Log.Info("[4/6] Creating and uploading video filter", logger.F("contentID", contentID))
			videoFilter := party.CreateVideoFilter(contentID)
			if videoFilter != nil {
				if err := storage.MarshalAndUpload(videoFilter, "batorment/v3/video-filter", fileName, dryRun, ""); err != nil {
					ui.Log.Warn("Failed to upload video filter", logger.F("error", err))
				}
			} else {
				ui.Log.Warn("No video filter created", logger.F("contentID", contentID))
			}

			ui.Log.Info("[5/6] Uploading additional filters", logger.F("contentID", contentID))
			if err := storage.MarshalAndUpload(filterResult, "batorment/v3/filter", fileName, dryRun, ""); err != nil {
				ui.Log.Warn("Failed to upload filter", logger.F("error", err))
				continue
			}

			lunaticFilter := party.CreateLunaticFilter(partyData)
			if err := storage.MarshalAndUpload(lunaticFilter, "batorment/v3/lunatic-filter", fileName, dryRun, ""); err != nil {
				ui.Log.Warn("Failed to upload lunatic filter", logger.F("error", err))
			}

			nonLunaticFilter := party.CreateNonLunaticFilter(partyData)
			if err := storage.MarshalAndUpload(nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, dryRun, ""); err != nil {
				ui.Log.Warn("Failed to upload non-lunatic filter", logger.F("error", err))
			}

			ui.Log.Info("[6/6] Processing and uploading summary data", logger.F("contentID", contentID))
			summaryData, err := party.ProcessPartyDataToSummaryData(partyData)
			if err != nil {
				ui.Log.Warn("Failed to process summary data", logger.F("error", err))
				continue
			}

			platinumCuts, err := party.GetPlatinumCuts(contentID, contentInfo.StartDate.Time)
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

			if err := storage.MarshalAndUpload(summaryData, "batorment/v3/summary", fileName, dryRun, ""); err != nil {
				ui.Log.Warn("Failed to upload summary data", logger.F("error", err))
				continue
			}

			ui.Log.Info("Successfully processed content", logger.F("contentID", contentID))
		}

		ui.Log.Info("Generating raids.json")
		var raidList []raidListItem
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

func init() {
	processRaidCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	rootCmd.AddCommand(processRaidCmd)
}
