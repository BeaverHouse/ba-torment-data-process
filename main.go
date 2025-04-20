package main

import (
	"ba-torment-data-process/app/batch"
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/database"
)

func main() {
	common.LoadEnv()
	common.InitLogger()
	database.InitPostgres()

	batch.UpdateData()
	batch.UpdateStudentInfo()
	// batch.UpdateYouTubeChannels()
	batch.DeleteOldRaidData(200)
}
