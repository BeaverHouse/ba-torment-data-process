package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/analysis"
	"ba-torment-data-process/internal/logic/storage"

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

	// Get all content IDs sorted by start_date
	contents, err := queries.ListContentIDsWithStartDate(context.Background())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to list content IDs: %w", err))
	}

	// Extract content IDs in order
	contentIDs := make([]string, len(contents))
	for i, c := range contents {
		contentIDs[i] = c.ContentID
	}

	log.Printf("Found %d content IDs", len(contentIDs))

	// Download all party data
	log.Println("Downloading party data from S3...")
	partyDataMap := analysis.DownloadAllPartyData(contentIDs)
	log.Printf("Successfully downloaded %d/%d party data", len(partyDataMap), len(contentIDs))

	if len(partyDataMap) == 0 {
		log.Fatal("No party data available for analysis")
	}

	// Run analysis
	log.Println("Running total analysis...")
	result := analysis.RunTotalAnalysis(partyDataMap, contentIDs)

	// Upload result
	err = storage.MarshalAndUpload(
		result,
		"batorment/v3",
		"total-analysis.json",
		*dryRun,
		"Total analysis completed",
	)
	if err != nil {
		log.Fatalf("Failed to upload analysis result: %v", err)
	}

	log.Println("Total analysis completed successfully!")
}
