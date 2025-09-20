package logic_download

import (
	"io"
	"log"
	"net/http"
	"time"
)

// Gets data from URL.
//
// If the URL is invalid or the data is empty, it returns an error.
func GetDataFromURL(url string) []byte {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("failed to get data from URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("invalid status code: %d, URL: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read data from URL: %v", err)
	} else if len(body) == 0 {
		log.Fatalf("the data from URL is empty: %s", url)
	}

	log.Printf("Response successfully fetched: url=%s, duration=%s", url, time.Since(start))
	return body
}
