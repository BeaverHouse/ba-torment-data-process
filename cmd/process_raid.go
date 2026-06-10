package cmd

import (
	"ba-torment-data-process/internal/logic/party"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var processRaidCmd = &cobra.Command{
	Use:   "process-raid",
	Short: "Process all raid content data",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		recent, _ := cmd.Flags().GetInt("recent")

		log, err := logger.NewLogger()
		if err != nil {
			return err
		}

		if err := party.ProcessAllRaids(log, dryRun, recent); err != nil {
			return err
		}

		log.Info("Successfully processed all raids")
		return nil
	},
}

func init() {
	processRaidCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	processRaidCmd.Flags().Int("recent", 5, "Number of recent raids to fully process from DuckDB (older ones use cached S3 data)")
	rootCmd.AddCommand(processRaidCmd)
}
