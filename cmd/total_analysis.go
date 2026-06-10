package cmd

import (
	"ba-torment-data-process/internal/logic/analysis"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var totalAnalysisCmd = &cobra.Command{
	Use:   "total-analysis",
	Short: "Run total analysis across all raid content",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		log, err := logger.NewLogger()
		if err != nil {
			return err
		}

		if err := analysis.RunTotalAnalysisPipeline(log, dryRun); err != nil {
			return err
		}

		log.Info("Total analysis completed successfully!")
		return nil
	},
}

func init() {
	totalAnalysisCmd.Flags().Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	rootCmd.AddCommand(totalAnalysisCmd)
}
