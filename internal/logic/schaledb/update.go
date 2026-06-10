package schaledb

import (
	"ba-torment-data-process/internal/db/postgres"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/logger"
)

// UpdateFromSchaleDB runs the full SchaleDB sync: students, presents, and i18n
// data. It owns the DB pool lifecycle so the CLI adapter only triggers and
// reports the run.
func UpdateFromSchaleDB(log logger.Logger) error {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()

	queries := postgres.New(pool)

	if _, err := ParseSchaleDBStudents(log, queries); err != nil {
		return err
	}

	if _, err := ParseSchaleDBPresents(log, queries); err != nil {
		return err
	}

	if err := SaveI18nData(log, queries); err != nil {
		return err
	}

	return nil
}
