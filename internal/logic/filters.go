package logic

import "strconv"

// Update the filter for a student.
func updateFilter(filters map[string](map[string]int), studentDetailID int) {
	studentIDString := strconv.Itoa(GetStudentID(studentDetailID))
	starWeaponString := strconv.Itoa(GetStarWeapon(studentDetailID))

	if _, exists := filters[studentIDString]; !exists {
		filters[studentIDString] = make(map[string]int)
	}

	filters[studentIDString][starWeaponString]++
}

// Update the filters for a party.
//
// It updates assist filters if the student is an assist, ad updates normal filters if the student is not an assist.
func UpdatePartyFilters(filters map[string](map[string]int), assistFilters map[string](map[string]int), studentDetailID int) {
	isAssist := IsAssist(studentDetailID)

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
	isAssist := IsAssist(studentDetailID)

	updateFilter(filters, studentDetailID)
	if isAssist {
		updateFilter(assistFilters, studentDetailID)
	}
}
