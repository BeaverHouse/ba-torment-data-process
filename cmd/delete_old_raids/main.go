package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

func main() {
	defer func() {
		common.LogInfo("오래된 총력전 데이터 삭제 프로세스 완료")
	}()

	// Get days from args or default to 200
	days := 200
	if len(os.Args) > 1 {
		if parsed, err := strconv.Atoi(os.Args[1]); err == nil {
			days = parsed
		}
	}

	// Create postgres config from environment
	cfg := types.PostgresConfig{
		Host:     common.GetEssentialEnv("POSTGRES_HOST"),
		Port:     5432,
		User:     common.GetEssentialEnv("POSTGRES_USER"),
		Password: common.GetEssentialEnv("POSTGRES_PASSWORD"),
		DBName:   common.GetEssentialEnv("POSTGRES_DBNAME"),
		SSLMode:  "disable",
	}

	if portStr := logic.GetEnv("POSTGRES_PORT", "5432"); portStr != "" {
		if parsed, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = parsed
		}
	}

	// Create database pool
	pool, err := postgres.NewPool(cfg)
	if err != nil {
		common.LogError(common.WrapErrorWithContext("main - creating pool", err))
		return
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
		common.LogError(common.WrapErrorWithContext("main - soft delete", err))
		return
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		common.LogInfo("삭제할 총력전 데이터가 없습니다.", zap.Int("days", days))
	} else {
		common.LogInfo("총력전 데이터 soft delete 완료",
			zap.Int64("deletedCount", rowsAffected),
			zap.Int("days", days),
			zap.Time("cutoffDate", cutoffDate))
	}
}