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

// GetDataFromURL gets data from URL with retry on transient (5xx) errors.
func GetDataFromURL(url string) ([]byte, error) {
	start := time.Now()

	var lastErr error
	for attempt := range maxRetries {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("http get failed: %w", err)
			if attempt < maxRetries-1 {
				ui.Log.Warn("Retrying request after transport error", logger.F("url", url), logger.F("attempt", attempt+1), logger.F("err", err.Error()))
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode >= 500 && attempt < maxRetries-1 {
			resp.Body.Close()
			ui.Log.Warn("Retrying request", logger.F("url", url), logger.F("status", resp.StatusCode), logger.F("attempt", attempt+1))
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d for URL: %s", resp.StatusCode, url)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		if len(body) == 0 {
			return nil, fmt.Errorf("empty response body for URL: %s", url)
		}

		ui.Log.Info("Response successfully fetched", logger.F("url", url), logger.F("duration", time.Since(start)))
		return body, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("max retries exceeded for URL: %s", url)
}
