package main

import (
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/videoref"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

var (
	pendingRaids []string = []string{
		"3S22-1",
		"3S22-2",
		"3S22-3",
		"3S23-1",
		"3S23-2",
		"3S23-3",
		"3S24-1",
		"3S24-2",
		"3S24-3",
		"3S25-1",
		"3S25-3",
		"3S25-4",
		"S78-0",
		"S79-0",
		"S80-0",
		"S81-0",
		"S82-0",
		"S83-0",
	}
)

func main() {
	if logic.IsLocalEnv() {
		if err := godotenv.Load(); err != nil {
			log.Fatalf("Failed to load .env file: %v", err)
		}
	}

	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no actual uploads)")
	flag.Parse()

	videoref.UpdateVideoRef(*dryRun, pendingRaids)
}
