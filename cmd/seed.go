package cmd

import (
	"ba-torment-data-process/internal/logic/seed"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed-test",
	Short: "Copy the tables an empty test database cannot rebuild, then rebuild the rest",
}

var seedExportCmd = &cobra.Command{
	Use:   "export <dir>",
	Short: "Read the raid list and video tables into CSV files (run against production)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		return seed.Export(cmd.Context(), args[0], dryRun)
	},
}

var seedImportCmd = &cobra.Command{
	Use:   "import <dir>",
	Short: "Load exported files into the local test database and run the data batch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		log, err := logger.NewLogger()
		if err != nil {
			return err
		}
		return seed.Import(cmd.Context(), log, args[0], dryRun)
	},
}

func init() {
	seedCmd.PersistentFlags().Bool("dry-run", false, "Print what would run without connecting to the database or writing anything")
	seedCmd.AddCommand(seedExportCmd, seedImportCmd)
	rootCmd.AddCommand(seedCmd)
}
