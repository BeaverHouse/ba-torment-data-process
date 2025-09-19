package update

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/app/types"
	"ba-torment-data-process/internal/logic/filter"
	"ba-torment-data-process/internal/logic/parse"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func UpdateData() {
	defer func() {
		common.LogInfo("총력전 데이터 업데이트 프로세스 완료")
	}()

	dryRun := true
	pendingRaids := []types.Raid{
		{
			RaidID: "S80-0",
			Name:   "테스트용",
			Status: "PENDING",
		},
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
		partyDataBytes, err := json.Marshal(partyData)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		filterResultBytes, err := json.Marshal(filterResult)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		logic_upload.UploadFile("v3/party", fmt.Sprintf("%s.json", raid.RaidID), partyDataBytes, dryRun)
		logic_upload.UploadFile("v3/filter", fmt.Sprintf("%s.json", raid.RaidID), filterResultBytes, dryRun)

		// Create and upload lunatic filter
		lunaticFilter := filter.CreateLunaticFilter(partyData, filterResult)
		lunaticFilterBytes, err := json.Marshal(lunaticFilter)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData - lunatic filter", err))
		} else {
			logic_upload.UploadFile("v3/lunatic-filter", fmt.Sprintf("%s.json", raid.RaidID), lunaticFilterBytes, dryRun)
			common.LogInfo("루나틱 필터 업로드 완료", zap.String("raidID", raid.RaidID))
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := filter.CreateNonLunaticFilter(partyData, filterResult)
		nonLunaticFilterBytes, err := json.Marshal(nonLunaticFilter)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData - non-lunatic filter", err))
		} else {
			logic_upload.UploadFile("v3/nonlunatic-filter", fmt.Sprintf("%s.json", raid.RaidID), nonLunaticFilterBytes, dryRun)
			common.LogInfo("논루나틱 필터 업로드 완료", zap.String("raidID", raid.RaidID))
		}

		summaryData, err = parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		summaryDataBytes, err := json.Marshal(summaryData)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		logic_upload.UploadFile("v3/summary", fmt.Sprintf("%s.json", raid.RaidID), summaryDataBytes, dryRun)

		err = database.UpdateRaidStatusToComplete(raid.RaidID)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		common.LogInfo("총력전 ID 업데이트 완료", zap.String("raidID", raid.RaidID))
	}
}
