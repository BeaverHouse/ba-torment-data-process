package batch

import (
	"time"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/database"
	"ba-torment-data-process/app/update_youtube"

	"go.uber.org/zap"
)

var excludeUserIDs = map[int]bool{
	13547823: true, // Not specified
}

// Update YouTube channels from recently added videos.
func UpdateYouTubeChannels() error {
	defer func() {
		common.LogInfo("유튜브 채널 업데이트 프로세스 완료")
	}()

	videos, err := database.GetVideosAfterDate(time.Now().AddDate(0, 0, -14))
	if err != nil {
		return common.WrapErrorWithContext("UpdateYouTubeChannels", err)
	}

	for _, video := range videos {
		hasChannel, err := database.HasChannel(video.UserID)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateYouTubeChannels", err))
			continue
		}

		if hasChannel {
			common.LogInfo("User already has a channel", zap.Int("userID", video.UserID))
			continue
		}

		common.LogInfo("User does not have a channel", zap.Int("userID", video.UserID))
		if excludeUserIDs[video.UserID] {
			common.LogInfo("User is in the exclude list", zap.Int("userID", video.UserID))
			continue
		}
		channelInfo, err := update_youtube.GetYouTubeChannelInfoFromVideoURL(video.YouTubeURL)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateYouTubeChannels", err))
			continue
		}

		err = database.UpdateUserChannel(video.UserID, channelInfo.ChannelTitle, channelInfo.ChannelURL)
		if err != nil {
			common.LogError(common.WrapErrorWithContext("UpdateYouTubeChannels", err))
			continue
		}
	}

	return nil
}
