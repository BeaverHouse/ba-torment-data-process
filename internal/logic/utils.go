package logic

import (
	"log"
	"strconv"
	"strings"
)

// Generates a detailed student ID as an integer.
//
// Its format is {studentID}{star}{weaponStar}{1 if isAssist, else 0}
func GetStudentDetailIDInt(studentID int, star int, weaponStar int, isAssist bool) int {
	isAssistInt := 0
	if isAssist {
		isAssistInt = 1
	}
	return studentID*1000 + star*100 + weaponStar*10 + isAssistInt
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
