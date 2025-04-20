package database

import (
	"database/sql"
	"fmt"

	"ba-torment-data-process/app/common"

	_ "github.com/lib/pq"
)

var db *sql.DB

// Initializes the database. It'll panic if it fails to connect to the database.
func InitPostgres() {
	user := common.GetEssentialEnv("POSTGRES_USER")
	password := common.GetEssentialEnv("POSTGRES_PASSWORD")
	dbname := common.GetEssentialEnv("POSTGRES_DBNAME")
	host := common.GetEssentialEnv("POSTGRES_HOST")
	port := common.GetEssentialEnv("POSTGRES_PORT")

	connStr := "user=" + user + " password=" + password + " dbname=" + dbname + " host=" + host + " port=" + port + " sslmode=disable TimeZone=Asia/Seoul"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Postgres: %v", err))
	}

	common.LogInfo("Connected to Postgres")
}

// Get the database object.
func GetDB() *sql.DB {
	return db
}

// Executes a query and returns the result rows or error. Usually used for SELECT.
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, common.WrapErrorWithContext("Query", err)
	}
	return rows, nil
}

// Executes a query and returns a single row.
//
// Error is not occured until the row's Scan method is called.
func QueryRow(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// Executes a query and returns the result or error. Usually used for INSERT, UPDATE, DELETE.
func Exec(query string, args ...interface{}) (sql.Result, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return nil, common.WrapErrorWithContext("Exec", err)
	}
	return result, nil
}
