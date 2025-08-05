package batch

import (
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/scrape"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func ScrapeAronaAI() {
	log.Println("Starting to scrape data from arona.ai")

	pendingRaids, err := database.GetPendingRaids()
	if err != nil {
		log.Fatalf("Failed to get pending raids: %v", err)
	}

	for _, raid := range pendingRaids {
		seasonString := raid.RaidID
		data, err := scrape.GetDataFromAronaAI(seasonString)
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
		os.WriteFile("data/"+seasonString+".json", []byte(data), 0644)
	}

	log.Println("Successfully scraped data.")
}
