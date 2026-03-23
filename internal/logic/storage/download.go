package storage

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
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

	log.Printf("Response successfully fetched: url=%s, duration=%s", url, time.Since(start))
	return body
}
