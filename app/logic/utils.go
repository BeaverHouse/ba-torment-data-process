package logic

import (
	"ba-torment-data-process/app/common"
	"fmt"
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

// Returns the level of the score. (Lunatic / Torment / Insane)
func GetLevelFromScore(score int) string {
	if score >= 44000000 {
		return "L"
	} else if score >= 31076000 {
		return "T"
	}
	return "I"
}

// Splits the season string into season & category. (Ex. S16-1 >> S16, 1)
func SplitSeasonString(season string) (string, int, error) {
	parts := strings.Split(season, "-")
	if len(parts) != 2 {
		return "", -1, common.WrapErrorWithContext("SplitSeasonString", fmt.Errorf("invalid season string: %s", season))
	}
	category, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", -1, common.WrapErrorWithContext("SplitSeasonString", err)
	}
	return parts[0], category, nil
}
