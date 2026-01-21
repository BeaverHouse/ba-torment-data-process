package id

import (
	"log"
	"strconv"
	"strings"
)

// ExtractGroupID extracts group ID from content_id (e.g., "3S26-1" -> "3S26")
func ExtractGroupID(contentID string) string {
	if idx := strings.Index(contentID, "-"); idx != -1 {
		return contentID[:idx]
	}
	return contentID
}

// SplitSeasonString splits the season string into season & category. (Ex. S16-1 >> S16, 1)
func SplitSeasonString(season string) (string, int) {
	parts := strings.Split(season, "-")
	if len(parts) != 2 {
		log.Fatalf("Invalid season string: %s", season)
	}
	category, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Invalid season string: %s", season)
	}
	return strings.Replace(parts[0], "3S", "S", 1), category
}
