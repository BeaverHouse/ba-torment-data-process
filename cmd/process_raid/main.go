package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/party"
	"ba-torment-data-process/internal/logic/storage"

	"github.com/joho/godotenv"
)

// RaidListItem represents an item in raids.json
type RaidListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TopLevel     string `json:"top_level"`
	PartyUpdated bool   `json:"party_updated"`
}

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

	// Get all contents for raid list
	contents, err := queries.ListContentsForRaidList(context.Background())
	if err != nil {
		log.Fatal(fmt.Errorf("failed to list contents: %w", err))
	}

	// Track party_updated status for each content
	partyUpdated := make(map[string]bool)

	for _, content := range contents {
		contentID := content.ContentID
		log.Printf("\n=== Processing content: %s ===", contentID)

		contentInfo, err := queries.GetContentByID(context.Background(), contentID)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to get content info: %w", err))
		}

		// Step 1: Parse DuckDB to create party data
		log.Printf("[1/6] Parsing DuckDB for %s...", contentID)
		partyData, filterResult, err := party.ParseDuckDB(contentID, contentInfo.StartDate.Time)
		if err != nil {
			log.Printf("Skipping content %s: %v", contentID, err)
			partyUpdated[contentID] = false
			continue
		}
		partyUpdated[contentID] = true

		fileName := fmt.Sprintf("%s.json", contentID)

		// Step 2: Update video references (without S3 download)
		log.Printf("[2/6] Updating video references for %s...", contentID)
		updated, err := party.UpdateVideoRefWithData(partyData, contentID)
		if err != nil {
			log.Printf("Warning: Failed to update video refs for %s: %v", contentID, err)
		} else {
			log.Printf("Updated %d video references for %s", updated, contentID)
		}

		// Step 3: Upload party data (with video refs if updated)
		log.Printf("[3/6] Uploading party data for %s...", contentID)
		if err := storage.MarshalAndUpload(partyData, "batorment/v3/party", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload party data: %v", err)
			continue
		}

		// Step 4: Create and upload video filter
		log.Printf("[4/6] Creating and uploading video filter for %s...", contentID)
		videoFilter := party.CreateVideoFilter(contentID)
		if videoFilter != nil {
			if err := storage.MarshalAndUpload(videoFilter, "batorment/v3/video-filter", fileName, *dryRun, ""); err != nil {
				log.Printf("Warning: Failed to upload video filter: %v", err)
			}
		} else {
			log.Printf("Warning: No video filter created for %s", contentID)
		}

		// Step 5: Upload additional filters
		log.Printf("[5/6] Uploading additional filters for %s...", contentID)

		// Upload basic filter
		if err := storage.MarshalAndUpload(filterResult, "batorment/v3/filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload filter: %v", err)
			continue
		}

		// Create and upload lunatic filter
		lunaticFilter := party.CreateLunaticFilter(partyData)
		if err := storage.MarshalAndUpload(lunaticFilter, "batorment/v3/lunatic-filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload lunatic filter: %v", err)
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := party.CreateNonLunaticFilter(partyData)
		if err := storage.MarshalAndUpload(nonLunaticFilter, "batorment/v3/nonlunatic-filter", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload non-lunatic filter: %v", err)
		}

		// Step 6: Create and upload summary data
		log.Printf("[6/6] Processing and uploading summary data for %s...", contentID)
		summaryData, err := party.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			log.Printf("Failed to process summary data: %v", err)
			continue
		}

		// Add platinum cuts to summary data
		platinumCuts, err := party.GetPlatinumCuts(contentID, contentInfo.StartDate.Time)
		if err != nil {
			log.Printf("Warning: Failed to get platinum cuts for %s: %v", contentID, err)
		} else {
			summaryData.PlatinumCuts = platinumCuts
			log.Printf("Added %d platinum cuts for %s", len(platinumCuts), contentID)
		}

		// Add part platinum cuts for grand assault (individual part cuts from partyData)
		if party.IsGrandAssault(contentID) {
			partPlatinumCuts := party.GetPartPlatinumCutsFromPartyData(partyData)
			if len(partPlatinumCuts) > 0 {
				summaryData.PartPlatinumCuts = partPlatinumCuts
				log.Printf("Added %d part platinum cuts for %s", len(partPlatinumCuts), contentID)
			}
		}

		// Add essential characters (70%+ usage in platinum ranks)
		essentialTorment, essentialLunatic := party.GetEssentialCharacters(partyData)
		summaryData.Torment.EssentialCharacters = essentialTorment
		summaryData.Lunatic.EssentialCharacters = essentialLunatic
		log.Printf("Essential characters for %s: Torment=%d, Lunatic=%d", contentID, len(essentialTorment), len(essentialLunatic))

		// Add high impact characters (biggest score gap when missing)
		highImpactTorment, highImpactLunatic := party.GetHighImpactCharacters(partyData)
		summaryData.Torment.HighImpactCharacters = highImpactTorment
		summaryData.Lunatic.HighImpactCharacters = highImpactLunatic
		log.Printf("High impact characters for %s: Torment=%d, Lunatic=%d", contentID, len(highImpactTorment), len(highImpactLunatic))

		// Add min UE users (users who cleared with minimum unique equipment)
		minUETorment, minUELunatic := party.GetMinUEUsers(partyData)
		summaryData.Torment.MinUEUser = minUETorment
		summaryData.Lunatic.MinUEUser = minUELunatic
		if minUETorment != nil {
			log.Printf("Min UE user (Torment) for %s: rank %d, %d UE, %d parties", contentID, minUETorment.Rank, minUETorment.UECount, len(minUETorment.PartyData))
		}
		if minUELunatic != nil {
			log.Printf("Min UE user (Lunatic) for %s: rank %d, %d UE, %d parties", contentID, minUELunatic.Rank, minUELunatic.UECount, len(minUELunatic.PartyData))
		}

		// Add max party users (users who cleared with maximum party count)
		maxPartyTorment, maxPartyLunatic := party.GetMaxPartyUsers(partyData)
		summaryData.Torment.MaxPartyUser = maxPartyTorment
		summaryData.Lunatic.MaxPartyUser = maxPartyLunatic
		if maxPartyTorment != nil {
			log.Printf("Max party user (Torment) for %s: rank %d, %d parties", contentID, maxPartyTorment.Rank, len(maxPartyTorment.PartyData))
		}
		if maxPartyLunatic != nil {
			log.Printf("Max party user (Lunatic) for %s: rank %d, %d parties", contentID, maxPartyLunatic.Rank, len(maxPartyLunatic.PartyData))
		}

		if err := storage.MarshalAndUpload(summaryData, "batorment/v3/summary", fileName, *dryRun, ""); err != nil {
			log.Printf("Failed to upload summary data: %v", err)
			continue
		}

		log.Printf("Successfully processed content: %s\n", contentID)
	}

	// Generate raids.json
	log.Println("\n=== Generating raids.json ===")
	var raidList []RaidListItem
	for _, content := range contents {
		raidList = append(raidList, RaidListItem{
			ID:           content.ContentID,
			Name:         content.Title,
			TopLevel:     string(content.TopLevel),
			PartyUpdated: partyUpdated[content.ContentID],
		})
	}

	if err := storage.MarshalAndUpload(raidList, "batorment/v3", "raids.json", *dryRun, "Raids list uploaded"); err != nil {
		log.Printf("Failed to upload raids.json: %v", err)
	}

	fmt.Println("\n=== Successfully processed all raids ===")
}
