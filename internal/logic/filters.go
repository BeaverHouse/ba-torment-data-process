package logic

import "strconv"

// Update the filter for a student.
func updateFilter(filters map[string](map[string]int), studentDetailID int) {
	studentID := studentDetailID / 1000
	starWeapon := (studentDetailID % 1000) / 10

	studentIDString := strconv.Itoa(studentID)
	starWeaponString := strconv.Itoa(starWeapon)

	if _, exists := filters[studentIDString]; !exists {
		filters[studentIDString] = make(map[string]int)
	}

	filters[studentIDString][starWeaponString]++
}

// Update the filters for a party.
//
// It updates assist filters if the student is an assist, ad updates normal filters if the student is not an assist.
func UpdatePartyFilters(filters map[string](map[string]int), assistFilters map[string](map[string]int), studentDetailID int) {
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
func UpdateSummaryFilters(filters map[string](map[string]int), assistFilters map[string](map[string]int), studentDetailID int) {
	isAssist := studentDetailID%10 == 1

	updateFilter(filters, studentDetailID)
	if isAssist {
		updateFilter(assistFilters, studentDetailID)
	}
}
