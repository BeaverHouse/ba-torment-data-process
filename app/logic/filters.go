package logic

import "strconv"

// Update the filter for a student.
func updateFilter(filters map[string][]int, studentDetailID int) {
	studentID := studentDetailID / 1000
	star := (studentDetailID % 1000) / 100
	weaponStar := (studentDetailID % 100) / 10

	studentIDString := strconv.Itoa(studentID)

	if _, exists := filters[studentIDString]; !exists {
		filters[studentIDString] = make([]int, 9)
	}

	if star < 5 {
		filters[studentIDString][star]++
	} else {
		filters[studentIDString][5+weaponStar]++
	}
}

// Update the filters for a party.
//
// It updates assist filters if the student is an assist, ad updates normal filters if the student is not an assist.
func UpdatePartyFilters(filters map[string][]int, assistFilters map[string][]int, studentDetailID int) {
	isAssist := studentDetailID%10 == 1

	targetFilters := filters
	if isAssist {
		targetFilters = assistFilters
	}

	updateFilter(targetFilters, studentDetailID)
}

// Update the filters for a summary.
//
// It always updates normal filters, and updates assist filters if the student is an assist.
func UpdateSummaryFilters(filters map[string][]int, assistFilters map[string][]int, studentDetailID int) {
	isAssist := studentDetailID%10 == 1

	updateFilter(filters, studentDetailID)
	if isAssist {
		updateFilter(assistFilters, studentDetailID)
	}
}

// Drops filters with usage less than 1% of the clear count.
func DropLowUsageFilters(filters map[string][]int, clearCount int) {
	for key := range filters {
		sum := 0
		for _, count := range filters[key] {
			sum += count
		}
		if float64(sum) < 0.01*float64(clearCount) {
			delete(filters, key)
		}
	}
}
