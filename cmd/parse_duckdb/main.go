package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	logic_duckdb "ba-torment-data-process/internal/logic/duckdb"
	"ba-torment-data-process/internal/logic/filter"
	"ba-torment-data-process/internal/logic/parse"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"ba-torment-data-process/internal/types"

	"github.com/joho/godotenv"
)

var (
	raids = []string{
		"3S27-1",
		"3S27-3",
		"3S27-4",
	}
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	flag.Parse()

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

		// Parse DuckDB directly to final format
		partyData, filterResult, err := logic_duckdb.ParseDuckDB(raidInfo.ContentID, raidInfo.StartDate.Time)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to parse duckdb: %w", err))
		}

		fileName := fmt.Sprintf("%s.json", raidInfo.ContentID)

		// Upload party data
		partyDataBytes, err := json.Marshal(partyData)
		if err != nil {
			log.Printf("Failed to marshal party data: %v", err)
			continue
		}
		logic_upload.UploadFile("batorment/v3/party", fileName, partyDataBytes, *dryRun)

		// Upload filter
		filterResultBytes, err := json.Marshal(filterResult)
		if err != nil {
			log.Printf("Failed to marshal filter result: %v", err)
			continue
		}
		logic_upload.UploadFile("batorment/v3/filter", fileName, filterResultBytes, *dryRun)

		// Create and upload lunatic filter
		lunaticFilter := filter.CreateLunaticFilter(partyData)
		lunaticFilterBytes, err := json.Marshal(lunaticFilter)
		if err != nil {
			log.Printf("Failed to create lunatic filter: %v", err)
		} else {
			logic_upload.UploadFile("batorment/v3/lunatic-filter", fileName, lunaticFilterBytes, *dryRun)
			log.Printf("루나틱 필터 업로드 완료: %s", raidInfo.ContentID)
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := filter.CreateNonLunaticFilter(partyData)
		nonLunaticFilterBytes, err := json.Marshal(nonLunaticFilter)
		if err != nil {
			log.Printf("Failed to create non-lunatic filter: %v", err)
		} else {
			logic_upload.UploadFile("batorment/v3/nonlunatic-filter", fileName, nonLunaticFilterBytes, *dryRun)
			log.Printf("논루나틱 필터 업로드 완료: %s", raidInfo.ContentID)
		}

		// Create and upload summary data
		summaryData, err := parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			log.Printf("Failed to process summary data: %v", err)
			continue
		}
		summaryDataBytes, err := json.Marshal(summaryData)
		if err != nil {
			log.Printf("Failed to marshal summary data: %v", err)
			continue
		}
		logic_upload.UploadFile("batorment/v3/summary", fileName, summaryDataBytes, *dryRun)

		log.Printf("Successfully processed raid: %s", raidInfo.ContentID)
	}

	fmt.Println("Successfully parsed and processed all raids")
}
