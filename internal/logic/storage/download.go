package storage

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"ba-torment-data-process/internal/ui"

	"github.com/BeaverHouse/go-common/logger"
)

const maxRetries = 3

// GetDataFromURL gets data from URL with retry on transient errors.
func GetDataFromURL(url string) []byte {
	start := time.Now()

	for attempt := range maxRetries {
		resp, err := http.Get(url)
		if err != nil {
			panic(fmt.Sprintf("failed to get data from URL: %v", err))
		}

		if resp.StatusCode >= 500 && attempt < maxRetries-1 {
			resp.Body.Close()
			ui.Log.Warn("Retrying request", logger.F("url", url), logger.F("status", resp.StatusCode), logger.F("attempt", attempt+1))
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			panic(fmt.Sprintf("invalid status code: %d, URL: %s", resp.StatusCode, url))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			panic(fmt.Sprintf("failed to read data from URL: %v", err))
		} else if len(body) == 0 {
			panic(fmt.Sprintf("the data from URL is empty: %s", url))
		}

		ui.Log.Info("Response successfully fetched", logger.F("url", url), logger.F("duration", time.Since(start)))
		return body
	}

	panic(fmt.Sprintf("unreachable: max retries exceeded for URL: %s", url))
}
