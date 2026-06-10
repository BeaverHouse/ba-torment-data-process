package storage

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"ba-torment-data-process/internal/constants"

	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

const maxRetries = 3

// GetDataFromURL gets data from URL with retry on transient (5xx) errors.
func GetDataFromURL(log logger.Logger, url string) ([]byte, error) {
	start := time.Now()

	var lastErr error
	for attempt := range maxRetries {
		resp, err := http.Get(url)
		if err != nil {
			lastErr = errorhandle.ErrInternal(err)
			if attempt < maxRetries-1 {
				log.Warn("Retrying request after transport error", logger.Field{Key: "url", Value: url}, logger.Field{Key: "attempt", Value: attempt + 1}, logger.Field{Key: "err", Value: err.Error()})
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode >= 500 && attempt < maxRetries-1 {
			resp.Body.Close()
			log.Warn("Retrying request", logger.Field{Key: "url", Value: url}, logger.Field{Key: "status", Value: resp.StatusCode}, logger.Field{Key: "attempt", Value: attempt + 1})
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, constants.ErrUpstreamBadStatus(fmt.Sprintf("status %d for URL: %s", resp.StatusCode, url))
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, errorhandle.ErrInternal(err)
		}
		if len(body) == 0 {
			return nil, constants.ErrEmptyResponse(url)
		}

		log.Info("Response successfully fetched", logger.Field{Key: "url", Value: url}, logger.Field{Key: "duration", Value: time.Since(start)})
		return body, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, constants.ErrMaxRetriesExceeded(url)
}
