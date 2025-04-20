package database

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"strconv"
)

// Gets all student IDs from the database.
func GetStudentIDs() ([]int, error) {
	rows, err := Query("SELECT student_id FROM ba_torment.students")
	if err != nil {
		return nil, common.WrapErrorWithContext("GetStudentIDs", err)
	}
	defer rows.Close()

	var studentIDs []int
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, common.WrapErrorWithContext("GetStudentIDs", err)
		}
		studentIDInt, err := strconv.Atoi(studentID)
		if err != nil {
			return nil, common.WrapErrorWithContext("GetStudentIDs", err)
		}
		studentIDs = append(studentIDs, studentIDInt)
	}
	return studentIDs, nil
}

// Insert the student data.
func InsertStudent(studentData types.SchaleDBStudentData) error {
	query := `
		INSERT INTO ba_torment.students (student_id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (student_id) DO NOTHING
	`
	_, err := Exec(query, studentData.ID, studentData.Name)
	if err != nil {
		return common.WrapErrorWithContext("InsertStudent", err)
	}
	return nil
}
