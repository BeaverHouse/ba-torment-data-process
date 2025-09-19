package database

import (
	"database/sql"

	"ba-torment-data-process/app/common"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Executes a query and returns the result or error. Usually used for INSERT, UPDATE, DELETE.
func Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, common.WrapErrorWithContext("Exec", err)
	}
	return result, nil
}
