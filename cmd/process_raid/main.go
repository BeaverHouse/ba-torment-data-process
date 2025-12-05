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
	"ba-torment-data-process/internal/logic/videoref"

	"github.com/joho/godotenv"
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

	contentIDs, err := queries.ListContentIDs(context.Background())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to list content IDs: %w", err))
	}

	for _, contentID := range contentIDs {
		log.Printf("\n=== Processing content: %s ===", contentID)

		contentInfo, err := queries.GetContentByID(context.Background(), contentID)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to get content info: %w", err))
		}

		// Step 1: Parse DuckDB to create party data
		log.Printf("[1/6] Parsing DuckDB for %s...", contentID)
		partyData, filterResult, err := logic_duckdb.ParseDuckDB(contentID, contentInfo.StartDate.Time)
		if err != nil {
			log.Printf("Skipping content %s: %v", contentID, err)
			continue
		}

		fileName := fmt.Sprintf("%s.json", contentID)

		// Step 2: Update video references (without S3 download)
		log.Printf("[2/6] Updating video references for %s...", contentID)
		updated, err := videoref.UpdateVideoRefWithData(partyData, contentID)
		if err != nil {
			log.Printf("Warning: Failed to update video refs for %s: %v", contentID, err)
			// Continue even if video ref update fails
		} else {
			log.Printf("Updated %d video references for %s", updated, contentID)
		}

		// Step 3: Upload party data (with video refs if updated)
		log.Printf("[3/6] Uploading party data for %s...", contentID)
		if err := logic_upload.MarshalAndUpload(partyData, "batorment/v3/party", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload party data: %v", err)
			continue
		}

		// Step 4: Create and upload video filter
		log.Printf("[4/6] Creating and uploading video filter for %s...", contentID)
		videoFilter := filter.CreateVideoFilter(contentID)
		if videoFilter != nil {
			if err := logic_upload.MarshalAndUpload(videoFilter, "batorment/v3/video-filter", fileName, *dryRun, ""); err != nil {
				log.Printf("Warning: Failed to upload video filter: %v", err)
				// Continue even if video filter upload fails
			}
		} else {
			log.Printf("Warning: No video filter created for %s", contentID)
		}

		// Step 5: Upload additional filters
		log.Printf("[5/6] Uploading additional filters for %s...", contentID)

		// Upload basic filter
		if err := logic_upload.MarshalAndUpload(filterResult, "batorment/v3/filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload filter: %v", err)
			continue
		}

		// Create and upload lunatic filter
		lunaticFilter := filter.CreateLunaticFilter(partyData)
		if err := logic_upload.MarshalAndUpload(lunaticFilter, "batorment/v3/lunatic-filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload lunatic filter: %v", err)
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := filter.CreateNonLunaticFilter(partyData)
		if err := logic_upload.MarshalAndUpload(nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload non-lunatic filter: %v", err)
		}

		// Step 6: Create and upload summary data
		log.Printf("[6/6] Processing and uploading summary data for %s...", contentID)
		summaryData, err := parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			log.Printf("Failed to process summary data: %v", err)
			continue
		}

		// Add platinum cuts to summary data
		platinumCuts, err := logic_duckdb.GetPlatinumCuts(contentID, contentInfo.StartDate.Time)
		if err != nil {
			log.Printf("Warning: Failed to get platinum cuts for %s: %v", contentID, err)
			// Continue even if platinum cuts fail - it's not critical
		} else {
			summaryData.PlatinumCuts = platinumCuts
			log.Printf("Added %d platinum cuts for %s", len(platinumCuts), contentID)
		}

		if err := logic_upload.MarshalAndUpload(summaryData, "batorment/v3/summary", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload summary data: %v", err)
			continue
		}

		log.Printf("✓ Successfully processed content: %s\n", contentID)
	}

	fmt.Println("\n=== Successfully processed all raids ===")
}
