package seed

import (
	"context"
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ba-torment-data-process/internal/constants"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/analysis"
	"ba-torment-data-process/internal/logic/party"
	"ba-torment-data-process/internal/logic/schaledb"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The tables that cannot be rebuilt from SchaleDB and the raw raid files: the
// raid list and the reviewed video data.
var tables = []string{"batorment_v3.contents", "batorment_v3.youtube_metadata", "batorment_v3.youtube_analysis"}

// Export writes each table to dir as CSV with a header row. It only reads, so
// it is the step that runs against production.
func Export(ctx context.Context, dir string) error {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return errorhandle.ErrDBOperation("acquire connection", err)
	}
	defer conn.Release()

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errorhandle.ErrInternal(err)
	}
	for _, table := range tables {
		file, err := os.Create(filepath.Join(dir, table+".csv"))
		if err != nil {
			return errorhandle.ErrInternal(err)
		}
		_, err = conn.Conn().PgConn().CopyTo(ctx, file, "COPY "+table+" TO STDOUT WITH (FORMAT csv, HEADER)")
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return errorhandle.ErrDBOperation("export "+table, err)
		}
	}
	return nil
}

// Import replaces the tables in the local test database with the files Export
// wrote, then rebuilds students, their translations, raid data and the total
// analysis from their public inputs, as the scheduled batch does. Presents are
// skipped: no service reads them, and the shared test schema lacks their table.
func Import(ctx context.Context, log logger.Logger, dir string) error {
	if host := env.GetEnv("POSTGRES_HOST", ""); host != "127.0.0.1" && host != "localhost" {
		return constants.ErrSeedTargetNotLocal(host)
	}
	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	if err := load(ctx, pool, dir); err != nil {
		return err
	}
	log.Info("Seeded tables from export", logger.Field{Key: "dir", Value: dir})

	queries := postgres.New(postgres.Bounded(pool))
	if err := schaledb.ParseSchaleDBStudents(log, queries); err != nil {
		return err
	}
	if err := schaledb.SaveI18nData(log, queries); err != nil {
		return err
	}
	if err := party.ProcessAllRaids(log, false, 5); err != nil {
		return err
	}
	return analysis.RunTotalAnalysisPipeline(log, false)
}

func load(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errorhandle.ErrDBOperation("begin seed", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op once Commit succeeds

	queries := postgres.New(tx)
	if err := queries.RegisterServiceToken(ctx, env.GetEnv("BA_ANALYZER_SERVICE_TOKEN", "")); err != nil {
		return errorhandle.ErrDBOperation("register service token", err)
	}
	if err := queries.TruncateSeedTables(ctx); err != nil {
		return errorhandle.ErrDBOperation("truncate seed tables", err)
	}
	for _, table := range tables {
		if err := copyIn(ctx, tx, table, filepath.Join(dir, table+".csv")); err != nil {
			return err
		}
	}
	if err := queries.ResetVideoIDs(ctx); err != nil {
		return errorhandle.ErrDBOperation("reset video ids", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return errorhandle.ErrDBOperation("commit seed", err)
	}
	return nil
}

// copyIn loads one file by its header's column names, so a column order that
// differs between the two databases does not matter.
func copyIn(ctx context.Context, tx pgx.Tx, table, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	defer file.Close()

	header, err := csv.NewReader(file).Read()
	if err != nil {
		return constants.ErrDataDecode(path, err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errorhandle.ErrInternal(err)
	}
	columns := make([]string, len(header))
	for i, name := range header {
		columns[i] = pgx.Identifier{name}.Sanitize()
	}
	sql := "COPY " + table + " (" + strings.Join(columns, ", ") + ") FROM STDIN WITH (FORMAT csv, HEADER)"
	if _, err := tx.Conn().PgConn().CopyFrom(ctx, file, sql); err != nil {
		return errorhandle.ErrDBOperation("import "+table, err)
	}
	return nil
}
