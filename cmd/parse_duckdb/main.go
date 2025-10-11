package main

import (
	"context"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	logic_duckdb "ba-torment-data-process/internal/logic/duckdb"
	"ba-torment-data-process/internal/types"

	"github.com/joho/godotenv"
)

var (
	raids = []string{
		"S82-0",
	}
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	// Initialize database connection
	postgresConfig := types.PostgresConfig{
		Host:     logic.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     logic.GetIntEnv("POSTGRES_PORT", 5432),
		User:     logic.GetEnv("POSTGRES_USER", "postgres"),
		Password: logic.GetEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:   logic.GetEnv("POSTGRES_DB", "postgres"),
		SSLMode:  logic.GetEnv("POSTGRES_SSLMODE", "disable"),
	}
	pool, err := postgres.NewPool(postgresConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	queries := postgres.New(pool)
	for _, raid := range raids {
		raidInfo, err := queries.GetContents(context.Background(), raid)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to get raid info: %w", err))
		}
		if err := logic_duckdb.ParseDuckDB(raidInfo.ContentID, raidInfo.StartDate.Time); err != nil {
			log.Fatal(fmt.Errorf("failed to parse duckdb: %w", err))
		}
	}

	fmt.Println("Successfully parsed DuckDB data")
}
