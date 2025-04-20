package database

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"fmt"
)

// Get the list of raid IDs that have been created more than a certain number of days ago.
func GetOldRaidIDs(days int) ([]string, error) {
	query := `
		SELECT raid_id
		FROM ba_torment.raids
		WHERE created_at < NOW() - INTERVAL '` + fmt.Sprint(days) + ` days'
		AND deleted_at IS NULL
	`
	rows, err := Query(query)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetOldRaidIDs", err)
	}
	defer rows.Close()

	var raidIDs []string
	for rows.Next() {
		var raidID string
		if err := rows.Scan(&raidID); err != nil {
			return nil, common.WrapErrorWithContext("GetOldRaidIDs", err)
		}
		raidIDs = append(raidIDs, raidID)
	}
	return raidIDs, nil
}

// Get the list of raid IDs that are in pending state.
func GetPendingRaids() ([]types.Raid, error) {
	query := `
		SELECT raid_id, name, status, created_at, updated_at, deleted_at, top_level
		FROM ba_torment.raids
		WHERE status = 'PENDING'
		AND deleted_at IS NULL
	`
	rows, err := Query(query)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetPendingRaids", err)
	}
	defer rows.Close()

	var raids []types.Raid
	for rows.Next() {
		var raid types.Raid
		if err := rows.Scan(&raid.RaidID, &raid.Name, &raid.Status, &raid.CreatedAt, &raid.UpdatedAt, &raid.DeletedAt, &raid.TopLevel); err != nil {
			return nil, common.WrapErrorWithContext("GetPendingRaids > rows.Scan", err)
		}
		raids = append(raids, raid)
	}
	return raids, nil
}

// Deletes data with specific raid ID from raids table.
func DeleteRaidByID(raidID string) error {
	_, err := Exec(`
		UPDATE ba_torment.raids
		SET deleted_at = NOW()
		WHERE raid_id = $1
		AND deleted_at IS NULL
	`, raidID)
	if err != nil {
		return common.WrapErrorWithContext("DeleteRaidByID", err)
	}
	return nil
}

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
