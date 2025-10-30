package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	logic_duckdb "ba-torment-data-process/internal/logic/duckdb"
	"ba-torment-data-process/internal/logic/filter"
	"ba-torment-data-process/internal/logic/parse"
	logic_upload "ba-torment-data-process/internal/logic/upload"

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
	pool := postgres.InitFromEnv()
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
		if err := logic_upload.MarshalAndUpload(partyData, "batorment/v3/party", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload party data: %v", err)
			continue
		}

		// Upload filter
		if err := logic_upload.MarshalAndUpload(filterResult, "batorment/v3/filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload filter: %v", err)
			continue
		}

		// Create and upload lunatic filter
		lunaticFilter := filter.CreateLunaticFilter(partyData)
		if err := logic_upload.MarshalAndUpload(lunaticFilter, "batorment/v3/lunatic-filter", fileName, *dryRun, fmt.Sprintf("루나틱 필터 업로드 완료: %s", raidInfo.ContentID)); err != nil {
			log.Printf("Failed to upload lunatic filter: %v", err)
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := filter.CreateNonLunaticFilter(partyData)
		if err := logic_upload.MarshalAndUpload(nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, *dryRun, fmt.Sprintf("논루나틱 필터 업로드 완료: %s", raidInfo.ContentID)); err != nil {
			log.Printf("Failed to upload non-lunatic filter: %v", err)
		}

		// Create and upload summary data
		summaryData, err := parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			log.Printf("Failed to process summary data: %v", err)
			continue
		}
		if err := logic_upload.MarshalAndUpload(summaryData, "batorment/v3/summary", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload summary data: %v", err)
			continue
		}

		log.Printf("Successfully processed raid: %s", raidInfo.ContentID)
	}

	fmt.Println("Successfully parsed and processed all raids")
}
