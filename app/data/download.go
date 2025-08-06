package data

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"encoding/json"
)

var (
	schaleDBURL string = "https://schaledb.com/"
)

// Get student data from SchaleDB.
func GetStudentDataFromSchaleDB() ([]types.SchaleDBStudentData, error) {
	url := schaleDBURL + "data/kr/students.min.json"

	data, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetStudentDataFromSchaleDB", err)
	}

	var jsonData map[string]types.SchaleDBStudentData
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetStudentDataFromSchaleDB > json.Unmarshal", err)
	}

	var studentData []types.SchaleDBStudentData
	for _, student := range jsonData {
		studentData = append(studentData, student)
	}

	return studentData, nil
}
