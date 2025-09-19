package logic_download

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Gets data from URL.
//
// If the URL is invalid or the data is empty, it returns an error.
func GetDataFromURL(url string) ([]byte, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get data from URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %d, URL: %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from URL: %v", err)
	} else if len(body) == 0 {
		return nil, fmt.Errorf("the data from URL is empty: %s", url)
	}
	fmt.Println("Response successfully fetched", "url", url, "duration", time.Since(start))
	return body, nil
}
