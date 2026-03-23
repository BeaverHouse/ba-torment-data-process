package main

import (
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/schaledb"
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			panic(fmt.Sprintf("Failed to load .env file: %v", err))
		}
	}

	// Initialize database connection
	pool := postgres.InitFromEnv()
	defer pool.Close()

	queries := postgres.New(pool)
	_, err := schaledb.ParseSchaleDBStudents(queries)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse SchaleDB students: %v", err))
	}

	_, err = schaledb.ParseSchaleDBPresents(queries)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse SchaleDB presents: %v", err))
	}

	err = schaledb.SaveI18nData(queries)
	if err != nil {
		panic(fmt.Sprintf("Failed to save i18n data: %v", err))
	}

	log.Println("Successfully updated from SchaleDB")
}
