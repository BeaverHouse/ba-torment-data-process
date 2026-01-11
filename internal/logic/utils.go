package logic

import (
	"log"
	"strconv"
	"strings"
)

// StudentDetailID format: {studentID(5)}{star(1)}{weaponStar(1)}{isAssist(1)}
// Example: 10000310 = studentID=10000, star=3, weaponStar=1, isAssist=false

// ComposeStudentDetailID creates a detail ID from components.
func ComposeStudentDetailID(studentID, star, weaponStar int, isAssist bool) int {
	isAssistInt := 0
	if isAssist {
		isAssistInt = 1
	}
	return studentID*1000 + star*100 + weaponStar*10 + isAssistInt
}

// ParseStudentDetailID extracts all components from a detail ID.
func ParseStudentDetailID(detailID int) (studentID, star, weaponStar int, isAssist bool) {
	studentID = detailID / 1000
	star = (detailID % 1000) / 100
	weaponStar = (detailID % 100) / 10
	isAssist = detailID%10 == 1
	return
}

// GetStudentID extracts only the student ID from a detail ID.
func GetStudentID(detailID int) int {
	return detailID / 1000
}

// GetStarWeapon extracts star and weaponStar combined (for filter keys).
func GetStarWeapon(detailID int) int {
	return (detailID % 1000) / 10
}

// IsAssist checks if the detail ID represents an assist character.
func IsAssist(detailID int) bool {
	return detailID%10 == 1
}

// Splits the season string into season & category. (Ex. S16-1 >> S16, 1)
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
