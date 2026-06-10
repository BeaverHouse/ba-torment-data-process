package analysis

import (
	"context"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic/storage"

	gopostgres "github.com/BeaverHouse/go-common/database/postgres"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

// RunTotalAnalysisPipeline orchestrates the full total-analysis run: it lists
// content IDs from the database, downloads party data from S3, runs the
// analysis, and uploads the result. The CLI adapter only triggers it and
// reports completion.
func RunTotalAnalysisPipeline(log logger.Logger, dryRun bool) error {
	pool := gopostgres.InitFromEnv()
	defer pool.Close()

	queries := postgres.New(pool)

	contents, err := queries.ListContentIDsWithStartDate(context.Background())
	if err != nil {
		return errorhandle.ErrDBOperation("list content IDs", err)
	}

	contentIDs := make([]string, len(contents))
	for i, c := range contents {
		contentIDs[i] = c.ContentID
	}

	log.Info("Found content IDs", logger.Field{Key: "count", Value: len(contentIDs)})

	log.Info("Downloading party data from S3...")
	partyDataMap := DownloadAllPartyData(log, contentIDs)
	log.Info("Successfully downloaded party data", logger.Field{Key: "downloaded", Value: len(partyDataMap)}, logger.Field{Key: "total", Value: len(contentIDs)})

	if len(partyDataMap) == 0 {
		return errorhandle.ErrInternalMsg("no party data available for analysis")
	}

	log.Info("Running total analysis...")
	result := RunTotalAnalysis(log, partyDataMap, contentIDs)

	if err := storage.MarshalAndUpload(log, result, "batorment/v3", "total-analysis.json", dryRun, "Total analysis completed"); err != nil {
		return err
	}

	return nil
}
