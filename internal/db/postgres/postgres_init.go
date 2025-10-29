package postgres

import (
	"context"
	"fmt"
	"log"

	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/types"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Initializes a PostgreSQL connection from environment variables.
// Panics if connection fails.
func InitFromEnv() *pgxpool.Pool {
	postgresConfig := types.PostgresConfig{
		Host:     logic.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     logic.GetIntEnv("POSTGRES_PORT", 5432),
		User:     logic.GetEnv("POSTGRES_USER", "postgres"),
		Password: logic.GetEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:   logic.GetEnv("POSTGRES_DB", "postgres"),
		SSLMode:  logic.GetEnv("POSTGRES_SSLMODE", "disable"),
	}

	pool, err := NewPool(postgresConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return pool
}

func NewPool(cfg types.PostgresConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}
