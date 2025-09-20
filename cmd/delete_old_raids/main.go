package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

func main() {
	defer func() {
		log.Println("오래된 총력전 데이터 삭제 프로세스 완료")
	}()

	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	// Get days from args or default to 200
	days := 200
	if len(os.Args) > 1 {
		if parsed, err := strconv.Atoi(os.Args[1]); err == nil {
			days = parsed
		}
	}

	// Create postgres config from environment
	postgresPort := logic.GetEnv("POSTGRES_PORT", "5432")
	postgresPortInt, err := strconv.Atoi(postgresPort)
	if err != nil {
		panic("Failed to convert POSTGRES_PORT to int: " + postgresPort)
	}

	// Initialize database connection
	postgresConfig := types.PostgresConfig{
		Host:     logic.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     postgresPortInt,
		User:     logic.GetEnv("POSTGRES_USER", "postgres"),
		Password: logic.GetEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:   logic.GetEnv("POSTGRES_DB", "postgres"),
		SSLMode:  logic.GetEnv("POSTGRES_SSLMODE", "disable"),
	}
	pool, err := postgres.NewPool(postgresConfig)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Create queries
	queries := postgres.New(pool)

	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Convert to pgtype.Timestamp
	pgTimestamp := pgtype.Timestamp{
		Time:  cutoffDate,
		Valid: true,
	}

	// Execute soft delete
	ctx := context.Background()
	result, err := queries.SoftDeleteOldRaids(ctx, pgTimestamp)
	if err != nil {
		log.Printf("main - soft delete: %v", err)
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("삭제할 총력전 데이터가 없습니다. days: %d", days)
	} else {
		log.Printf("총력전 데이터 soft delete 완료 - deletedCount: %d, days: %d, cutoffDate: %v", rowsAffected, days, cutoffDate)
	}
}
