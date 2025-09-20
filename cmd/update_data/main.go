package main

import (
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/update"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	dryRun := true

	update.UpdateData(dryRun)
}
