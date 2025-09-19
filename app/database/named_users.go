package database

import (
	"ba-torment-data-process/app/common"
)

// Delete data with specific raid ID from named_users table.
func DeleteNamedUsersByRaidID(raidID string) error {
	_, err := Exec(`
		UPDATE ba_torment.named_users
		SET deleted_at = NOW()
		WHERE raid_id = $1
		AND deleted_at IS NULL
	`, raidID)
	if err != nil {
		return common.WrapErrorWithContext("DeleteNamedUsersByRaidID", err)
	}
	return nil
}
