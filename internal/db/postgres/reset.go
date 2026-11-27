package postgres

import (
	"context"
	_ "embed"
)

//go:embed sql/schema/schema.sql
var schema string

// Drops the batorment_v3 schema with its types and rebuilds it empty.
func ResetSchema(ctx context.Context, db DBTX) error {
	_, err := db.Exec(ctx, `DROP SCHEMA IF EXISTS batorment_v3 CASCADE;
DROP TYPE IF EXISTS top_level CASCADE;
DROP TYPE IF EXISTS analysis_type CASCADE;
`+schema)
	return err
}
