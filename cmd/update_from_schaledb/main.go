package main

import (
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/schaledb"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	// Initialize database connection
	pool := postgres.InitFromEnv()
	defer pool.Close()

	queries := postgres.New(pool)
	_, err := schaledb.ParseSchaleDBStudents(queries)
	if err != nil {
		log.Fatalf("Failed to parse SchaleDB students: %v", err)
	}

	_, err = schaledb.ParseSchaleDBPresents(queries)
	if err != nil {
		log.Fatalf("Failed to parse SchaleDB presents: %v", err)
	}

	log.Println("Successfully parsed SchaleDB students")
}
