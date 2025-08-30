package batch

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/app/parse"
	"ba-torment-data-process/app/types"

	"go.uber.org/zap"
)

func UpdateData() {
	defer func() {
		common.LogInfo("총력전 데이터 업데이트 프로세스 완료")
	}()

	pendingRaids, err := database.GetPendingRaids()
	if err != nil {
		common.ExitIfError(common.WrapErrorWithContext("UpdateData", err))
	}

	if len(pendingRaids) == 0 {
		common.LogInfo("업데이트할 총력전 ID가 없습니다.")
		return
	}

	for _, raid := range pendingRaids {
		var partyData *types.BATormentPartyData
		var summaryData *types.BATormentSummaryData

		// Get from Arona AI
		partyData, filterResult, err := parse.ParsePartyDataFromAronaAI(raid.RaidID)

		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		data.UploadPartyDataJSON(partyData, raid.RaidID, false)
		data.UploadFilterResultJSON(filterResult, raid.RaidID, false)

		summaryData, err = parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		data.UploadSummaryDataJSON(summaryData, raid.RaidID, false)

		err = database.UpdateRaidStatusToComplete(raid.RaidID)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		common.LogInfo("총력전 ID 업데이트 완료", zap.String("raidID", raid.RaidID))
	}
}
