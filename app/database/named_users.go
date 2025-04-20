package database

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"time"

	"go.uber.org/zap"
)

// Get video information that created after some time.
func GetVideosAfterDate(date time.Time) ([]types.NamedUser, error) {
	rows, err := Query(`
		SELECT user_id, raid_id, description, youtube_url, score FROM ba_torment.named_users 
		WHERE created_at >= $1 
		AND raid_id IS NOT NULL
	`, date)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetVideosAfterDate", err)
	}
	defer rows.Close()

	var videos []types.NamedUser
	for rows.Next() {
		var video types.NamedUser
		err := rows.Scan(&video.UserID, &video.RaidID, &video.Description, &video.YouTubeURL, &video.Score)
		if err != nil {
			return nil, common.WrapErrorWithContext("GetVideosAfterDate", err)
		}
		videos = append(videos, video)
	}
	return videos, nil
}

// Check if there is channel information for a specific user ID.
func HasChannel(userID int) (bool, error) {
	var exists bool
	err := QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM ba_torment.named_users 
			WHERE user_id = $1 
			AND raid_id IS NULL
		)
	`, userID).Scan(&exists)
	if err != nil {
		return false, common.WrapErrorWithContext("HasChannel", err)
	}
	return exists, nil
}

// Update user channel information.
func UpdateUserChannel(userID int, channelName, channelURL string) error {
	_, err := Exec(`
		INSERT INTO ba_torment.named_users (user_id, raid_id, description, youtube_url, created_at, updated_at, score)
		VALUES ($1, NULL, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0)
	`, userID, channelName, channelURL)
	if err != nil {
		return common.WrapErrorWithContext("UpdateUserChannel", err)
	}
	common.LogInfo("유튜브 채널 정보 업데이트", zap.Int("userID", userID), zap.String("channelName", channelName), zap.String("channelURL", channelURL))
	return nil
}

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
