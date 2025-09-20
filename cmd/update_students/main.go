package main

import (
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/schaledb"
	"ba-torment-data-process/internal/types"
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
	_, err = schaledb.ParseSchaleDBStudents(queries)
	if err != nil {
		log.Fatalf("Failed to parse SchaleDB students: %v", err)
	}

	log.Println("Successfully parsed SchaleDB students")
}
