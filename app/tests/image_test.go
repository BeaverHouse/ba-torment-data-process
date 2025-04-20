package tests

import (
	"fmt"
	"os"
	"testing"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
)

func TestUploadCharacterImage(t *testing.T) {
	testID := 26003

	common.InitLogger()

	// Upload test image
	if err := data.UploadCharacterImage(testID, true); err != nil {
		t.Errorf("failed to upload test image: %v", err)
		return
	}

	common.LoadEnv()
	oracleDownloadURL := common.GetEssentialEnv("BATORMENT_DOWNLOAD_URL")

	// Download test image
	imgBytes, err := common.GetDataFromURL(fmt.Sprintf("%s/o/batorment/test/character/%d.webp", oracleDownloadURL, testID))
	if err != nil {
		t.Errorf("failed to get image: %v", err)
		return
	}

	// Save test image to tests/files directory
	if err := os.WriteFile(fmt.Sprintf("files/%d.webp", testID), imgBytes, 0644); err != nil {
		t.Errorf("failed to save test image: %v", err)
		return
	}

	t.Logf("Successfully downloaded test image for ID %d", testID)
}
