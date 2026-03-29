package cmd

import (
	"fmt"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/schaledb"
	"ba-torment-data-process/internal/ui"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/spf13/cobra"
)

var updateFromSchaleDBCmd = &cobra.Command{
	Use:   "update-from-schaledb",
	Short: "Update student and present data from SchaleDB",
	RunE: func(cmd *cobra.Command, args []string) error {
		pool := gopostgres.InitFromEnv()
		defer pool.Close()

		queries := postgres.New(pool)

		if _, err := schaledb.ParseSchaleDBStudents(queries); err != nil {
			return fmt.Errorf("failed to parse SchaleDB students: %w", err)
		}

		if _, err := schaledb.ParseSchaleDBPresents(queries); err != nil {
			return fmt.Errorf("failed to parse SchaleDB presents: %w", err)
		}

		if err := schaledb.SaveI18nData(queries); err != nil {
			return fmt.Errorf("failed to save i18n data: %w", err)
		}

		ui.Log.Info("Successfully updated from SchaleDB")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateFromSchaleDBCmd)
}
