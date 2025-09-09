package batch

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/data"
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/parse"
	"ba-torment-data-process/app/types"
	"ba-torment-data-process/internal/constants"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func createLunaticFilter(partyData *types.BATormentPartyData, _ *types.BATormentFilter) *types.BATormentFilter {
	filters := make(map[string](map[string]int))
	assistFilters := make(map[string](map[string]int))

	// Filter parties with score >= LunaticMinScore
	for _, party := range partyData.PartyDetail {
		if party.Score >= constants.LunaticMinScore {
			// Process each party's characters
			for _, partyTeam := range party.PartyData {
				for _, studentDetailID := range partyTeam {
					if studentDetailID != 0 {
						logic.UpdateSummaryFilters(filters, assistFilters, studentDetailID)
					}
				}
			}
		}
	}

	return &types.BATormentFilter{
		Filters:       filters,
		AssistFilters: assistFilters,
	}
}

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

	// pendingRaids, err := database.GetPendingRaids()
	// if err != nil {
	// 	common.ExitIfError(common.WrapErrorWithContext("UpdateData", err))
	// }

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
		data.UploadFile("v3/party", fmt.Sprintf("%s.json", raid.RaidID), partyDataBytes, dryRun)
		data.UploadFile("v3/filter", fmt.Sprintf("%s.json", raid.RaidID), filterResultBytes, dryRun)

		// Create and upload lunatic filter
		lunaticFilter := createLunaticFilter(partyData, filterResult)
		lunaticFilterBytes, err := json.Marshal(lunaticFilter)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData - lunatic filter", err))
		} else {
			data.UploadFile("v3/lunatic-filter", fmt.Sprintf("%s.json", raid.RaidID), lunaticFilterBytes, dryRun)
			common.LogInfo("루나틱 필터 업로드 완료", zap.String("raidID", raid.RaidID))
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
		data.UploadFile("v3/summary", fmt.Sprintf("%s.json", raid.RaidID), summaryDataBytes, dryRun)

		err = database.UpdateRaidStatusToComplete(raid.RaidID)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateData", err))
			continue
		}
		common.LogInfo("총력전 ID 업데이트 완료", zap.String("raidID", raid.RaidID))
	}
}
