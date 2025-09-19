package database

import (
	"ba-torment-data-process/app/common"
)

func UpdateRaidStatusToComplete(raidID string) error {
	_, err := Exec(`
		UPDATE ba_torment.raids
		SET status = 'COMPLETE', updated_at = NOW()
		WHERE raid_id = $1
	`, raidID)
	if err != nil {
		return common.WrapErrorWithContext("UpdateRaidStatusToComplete", err)
	}
	return nil
}
