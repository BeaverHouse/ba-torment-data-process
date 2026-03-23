package main

import (
	"flag"
	"fmt"
	"log"

	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/gridimage"

	"github.com/joho/godotenv"
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			panic(fmt.Sprintf("Failed to load .env file: %v", err))
		}
	}

	dryRun := flag.Bool("dry-run", false, "Save to local files/ directory instead of uploading")
	flag.Parse()

	if err := gridimage.GenerateGridImages(*dryRun); err != nil {
		panic(fmt.Sprintf("Failed to generate grid images: %v", err))
	}

	log.Println("Successfully generated all grid images")
}
