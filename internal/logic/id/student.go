package id

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

// IsStriker checks if studentID is a striker (1xxxx).
func IsStriker(studentID int) bool {
	return studentID/10000 == 1
}

// IsSpecial checks if studentID is a special (2xxxx).
func IsSpecial(studentID int) bool {
	return studentID/10000 == 2
}
