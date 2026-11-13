package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Every query here takes milliseconds; one still waiting after this has lost
// its connection, and without a deadline the job would wait on it forever.
const queryTimeout = time.Minute

// Bounded gives every query on the pool a deadline.
func Bounded(pool *pgxpool.Pool) DBTX {
	return bounded{pool}
}

type bounded struct {
	pool *pgxpool.Pool
}

func (b bounded) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	return b.pool.Exec(ctx, sql, args...)
}

func (b bounded) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	rows, err := b.pool.Query(ctx, sql, args...)
	if err != nil {
		cancel()
		return nil, err
	}
	return boundedRows{rows, cancel}, nil
}

func (b bounded) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	return boundedRow{b.pool.QueryRow(ctx, sql, args...), cancel}
}

// The deadline lasts until the rows are read.
type boundedRows struct {
	pgx.Rows
	cancel context.CancelFunc
}

func (r boundedRows) Close() {
	r.Rows.Close()
	r.cancel()
}

type boundedRow struct {
	row    pgx.Row
	cancel context.CancelFunc
}

func (r boundedRow) Scan(dest ...any) error {
	defer r.cancel()
	return r.row.Scan(dest...)
}
