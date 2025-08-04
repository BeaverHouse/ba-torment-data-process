package batch

import (
	"ba-torment-data-process/scrape"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ScrapeAronaAI() {
	log.Println("Starting to scrape data from arona.ai")

	// seasons := []string{"S79-0", "S78-0"}
	seasons := []string{"S23-3"}
	// seasons := []string{"S33-1"}

	for _, season := range seasons {
		data, err := scrape.GetDataFromAronaAI(season)
		if err != nil {
			log.Fatalf("Failed to get data from arona.ai: %v", err)
		}

		// For now, we'll just print the first 100 characters of the data.
		// In a real scenario, you might want to save it to a file or process it further.
		if len(data) > 100 {
			fmt.Println(data[:100])
		} else {
			fmt.Println(data)
		}

		// save tValue to file
		json.Marshal(data)
		os.WriteFile("data/"+season+".json", []byte(data), 0644)
	}

	log.Println("Successfully scraped data.")
}
