package storage

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"ba-torment-data-process/internal/ui"

	"github.com/BeaverHouse/go-common/logger"
)

// GetDataFromURL gets data from URL.
func GetDataFromURL(url string) []byte {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Sprintf("failed to get data from URL: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Sprintf("invalid status code: %d, URL: %s", resp.StatusCode, url))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read data from URL: %v", err))
	} else if len(body) == 0 {
		panic(fmt.Sprintf("the data from URL is empty: %s", url))
	}

	ui.Log.Info("Response successfully fetched", logger.F("url", url), logger.F("duration", time.Since(start)))
	return body
}
