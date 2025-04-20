package batch

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/database"

	"go.uber.org/zap"
)

// Deletes raid data that is older than certain days.
func DeleteOldRaidData(days int) {
	defer func() {
		common.LogInfo("오래된 총력전 데이터 삭제 프로세스 완료")
	}()

	oldRaidIDs, err := database.GetOldRaidIDs(days)
	if err != nil {
		common.LogError(common.WrapErrorWithContext("DeleteOldRaidData", err))
		return
	}

	if len(oldRaidIDs) == 0 {
		common.LogInfo("삭제할 총력전 ID가 없습니다.")
		return
	}

	common.LogInfo("삭제할 총력전 ID 목록", zap.Any("oldRaidIDs", oldRaidIDs))

	for _, raidID := range oldRaidIDs {
		if err := database.DeleteRaidByID(raidID); err != nil {
			common.LogError(common.WrapErrorWithContext("DeleteOldRaidData", err))
			continue
		}

		if err := database.DeleteNamedUsersByRaidID(raidID); err != nil {
			common.LogError(common.WrapErrorWithContext("DeleteOldRaidData", err))
			continue
		}

		common.LogInfo("총력전 ID 삭제 완료", zap.String("raidID", raidID))
	}
}
