package cmd

import (
	"context"
	"fmt"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/analysis"
	"ba-torment-data-process/internal/logic/storage"
	"ba-torment-data-process/internal/ui"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var totalAnalysisCmd = &cobra.Command{
	Use:   "total-analysis",
	Short: "Run total analysis across all raid content",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		pool := gopostgres.InitFromEnv()
		defer pool.Close()

		queries := postgres.New(pool)

		contents, err := queries.ListContentIDsWithStartDate(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list content IDs: %w", err)
		}

		contentIDs := make([]string, len(contents))
		for i, c := range contents {
			contentIDs[i] = c.ContentID
		}

		ui.Log.Info("Found content IDs", logger.F("count", len(contentIDs)))

		ui.Log.Info("Downloading party data from S3...")
		partyDataMap := analysis.DownloadAllPartyData(contentIDs)
		ui.Log.Info("Successfully downloaded party data", logger.F("downloaded", len(partyDataMap)), logger.F("total", len(contentIDs)))

		if len(partyDataMap) == 0 {
			return fmt.Errorf("no party data available for analysis")
		}

		ui.Log.Info("Running total analysis...")
		result := analysis.RunTotalAnalysis(partyDataMap, contentIDs)

		if err := storage.MarshalAndUpload(result, "batorment/v3", "total-analysis.json", dryRun, "Total analysis completed"); err != nil {
			return fmt.Errorf("failed to upload analysis result: %w", err)
		}

		ui.Log.Info("Total analysis completed successfully!")
		return nil
	},
}

func init() {
	totalAnalysisCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	rootCmd.AddCommand(totalAnalysisCmd)
}
