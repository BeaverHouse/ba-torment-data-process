package batch

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/app/types"
	"slices"

	"go.uber.org/zap"
)

// Updates character id and name from SchaleDB
func UpdateStudentInfo() error {
	defer func() {
		common.LogInfo("학생 정보 업데이트 프로세스 완료")
	}()

	studentData, err := data.GetStudentDataFromSchaleDB()
	if err != nil {
		return common.WrapErrorWithContext("UpdateStudentInfo", err)
	}

	// Get current student IDs
	currentStudentIDs, _ := database.GetStudentIDs()

	// Prepare update data
	var updateInfo []types.SchaleDBStudentData
	for _, student := range studentData {
		if !slices.Contains(currentStudentIDs, student.ID) {
			updateInfo = append(updateInfo, student)
		}
	}

	if len(updateInfo) == 0 {
		common.LogInfo("업데이트할 학생이 없습니다.")
		return nil
	}

	// Insert new students
	for _, student := range updateInfo {
		common.LogInfo("학생 데이터 입력", zap.Int("ID", student.ID), zap.String("Name", student.Name))
		err = database.InsertStudent(student)
		if err != nil {
			return common.WrapErrorWithContext("UpdateStudentInfo", err)
		}

		if err := data.UploadCharacterImage(student.ID, false); err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateStudentInfo", err))
			continue
		}
	}

	return nil
}
