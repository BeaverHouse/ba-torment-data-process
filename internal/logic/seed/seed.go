package seed

import (
	"context"
	"encoding/csv"
	"fmt"
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
	"github.com/jackc/pgx/v5/pgconn"
)

// The tables that cannot be rebuilt from SchaleDB and the raw raid files: the
// raid list and the reviewed video data.
var tables = []string{"batorment_v3.contents", "batorment_v3.youtube_metadata", "batorment_v3.youtube_analysis"}

// The rest is rebuilt from its public inputs, as the scheduled batch does.
// Presents are skipped: no service reads them.
var rebuildSteps = []struct {
	name string
	run  func(logger.Logger, *postgres.Queries) error
}{
	{"update students from SchaleDB", schaledb.ParseSchaleDBStudents},
	{"save translations from SchaleDB", schaledb.SaveI18nData},
	{"process the 5 most recent raids", func(log logger.Logger, _ *postgres.Queries) error {
		return party.ProcessAllRaids(log, false, 5)
	}},
	{"run the total analysis", func(log logger.Logger, _ *postgres.Queries) error {
		return analysis.RunTotalAnalysisPipeline(log, false)
	}},
}

// target runs the import statements, or prints them for a dry run.
type target interface {
	postgres.DBTX
	copyCSV(ctx context.Context, r io.Reader, sql string) error
}

type txTarget struct{ pgx.Tx }

func (t txTarget) copyCSV(ctx context.Context, r io.Reader, sql string) error {
	_, err := t.Conn().PgConn().CopyFrom(ctx, r, sql)
	return err
}

// printTarget leaves Query and QueryRow nil, since the seed queries only Exec.
type printTarget struct{ postgres.DBTX }

func (printTarget) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	fmt.Println(sql)
	return pgconn.CommandTag{}, nil
}

func (printTarget) copyCSV(_ context.Context, _ io.Reader, sql string) error {
	fmt.Println(sql)
	return nil
}

// Export writes each table to dir as CSV with a header row. It is the step
// that runs against production, so it reads inside a read-only transaction.
func Export(ctx context.Context, dir string, dryRun bool) error {
	if err := refuseRepoDir(dir); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", env.GetEnv("POSTGRES_HOST", ""))
		for _, table := range tables {
			fmt.Println("Export", table, "to", dumpFile(dir, table))
		}
		return nil
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return errorhandle.ErrDBOperation("begin export", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a read-only transaction has nothing to keep

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return errorhandle.ErrInternal(err)
	}
	for _, table := range tables {
		file, err := os.Create(dumpFile(dir, table))
		if err != nil {
			return errorhandle.ErrInternal(err)
		}
		_, err = tx.Conn().PgConn().CopyTo(ctx, file, "COPY "+table+" TO STDOUT WITH (FORMAT csv, HEADER)")
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return errorhandle.ErrDBOperation("export "+table, err)
		}
	}
	return nil
}

// Reset drops the batorment_v3 schema of the local test database and rebuilds
// it empty from this repository's schema file.
func Reset(ctx context.Context, log logger.Logger, dryRun bool) error {
	host, err := localHost()
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", host)
		return postgres.ResetSchema(ctx, printTarget{})
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	if err := postgres.ResetSchema(ctx, pool); err != nil {
		return errorhandle.ErrDBOperation("reset schema", err)
	}
	log.Info("Rebuilt the batorment_v3 schema")
	return nil
}

// localHost refuses any database but a local one, since these commands erase.
func localHost() (string, error) {
	host := env.GetEnv("POSTGRES_HOST", "")
	if host != "127.0.0.1" && host != "localhost" {
		return "", constants.ErrSeedTargetNotLocal(host)
	}
	return host, nil
}

// Import replaces the tables in the local test database with the files Export
// wrote, then runs the rebuild steps.
func Import(ctx context.Context, log logger.Logger, dir string, dryRun bool) error {
	host, err := localHost()
	if err != nil {
		return err
	}
	if err := refuseRepoDir(dir); err != nil {
		return err
	}
	if dryRun {
		fmt.Println("Target host:", host)
		if err := load(ctx, printTarget{}, dir); err != nil {
			return err
		}
		for _, step := range rebuildSteps {
			fmt.Println("Then:", step.name)
		}
		return nil
	}

	pool := gopostgres.InitFromEnv()
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return errorhandle.ErrDBOperation("begin seed", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op once Commit succeeds

	if err := load(ctx, txTarget{tx}, dir); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errorhandle.ErrDBOperation("commit seed", err)
	}
	log.Info("Seeded tables from export", logger.Field{Key: "dir", Value: dir})

	queries := postgres.New(postgres.Bounded(pool))
	for _, step := range rebuildSteps {
		if err := step.run(log, queries); err != nil {
			return err
		}
	}
	return nil
}

// refuseRepoDir keeps production data out of the repository. The CLI runs
// with `go run .`, so the working directory is the repository root.
func refuseRepoDir(dir string) error {
	root, err := os.Getwd()
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}
	if rel, err := filepath.Rel(root, abs); err == nil && filepath.IsLocal(rel) {
		return constants.ErrSeedDirInRepo(abs)
	}
	return nil
}

func dumpFile(dir, table string) string {
	return filepath.Join(dir, table+".csv")
}

func load(ctx context.Context, db target, dir string) error {
	queries := postgres.New(db)
	if err := queries.RegisterServiceToken(ctx, env.GetEnv("BA_ANALYZER_SERVICE_TOKEN", "")); err != nil {
		return errorhandle.ErrDBOperation("register service token", err)
	}
	if err := queries.TruncateSeedTables(ctx); err != nil {
		return errorhandle.ErrDBOperation("truncate seed tables", err)
	}
	for _, table := range tables {
		if err := copyIn(ctx, db, table, dumpFile(dir, table)); err != nil {
			return err
		}
	}
	if err := queries.ResetVideoIDs(ctx); err != nil {
		return errorhandle.ErrDBOperation("reset video ids", err)
	}
	return nil
}

// copyIn loads one file by its header's column names, so a column order that
// differs between the two databases does not matter.
func copyIn(ctx context.Context, db target, table, path string) error {
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
	if err := db.copyCSV(ctx, file, sql); err != nil {
		return errorhandle.ErrDBOperation("import "+table, err)
	}
	return nil
}
