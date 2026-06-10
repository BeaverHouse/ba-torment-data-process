package cmd

import (
	"ba-torment-data-process/internal/logic/schaledb"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/spf13/cobra"
)

var updateFromSchaleDBCmd = &cobra.Command{
	Use:   "update-from-schaledb",
	Short: "Update student and present data from SchaleDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		log, err := logger.NewLogger()
		if err != nil {
			return err
		}

		if err := schaledb.UpdateFromSchaleDB(log); err != nil {
			return err
		}

		log.Info("Successfully updated from SchaleDB")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateFromSchaleDBCmd)
}
