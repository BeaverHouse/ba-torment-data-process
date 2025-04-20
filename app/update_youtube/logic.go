package update_youtube

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

var (
	apiKey  string
	service *youtube.Service
)

func init() {
	common.LoadEnv()
	apiKey = common.GetEssentialEnv("YOUTUBE_API_KEY")

	var err error
	service, err = getYouTubeService()
	common.ExitIfError(err)
}

// getYouTubeService creates a YouTube API client.
func getYouTubeService() (*youtube.Service, error) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, common.WrapErrorWithContext("getYouTubeService", err)
	}
	return service, nil
}

// getVideoInfo is a function to get video information from a YouTube video ID.
func getVideoInfo(service *youtube.Service, videoID string) (*youtube.Video, error) {
	call := service.Videos.List([]string{"snippet"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return nil, common.WrapErrorWithContext("getVideoInfo", err)
	}

	if len(response.Items) == 0 {
		return nil, common.WrapErrorWithContext("getVideoInfo", fmt.Errorf("비디오를 찾을 수 없습니다"))
	}

	return response.Items[0], nil
}

// GetYouTubeChannelInfoFromVideoURL returns the channel URL and title from a YouTube video URL.
func GetYouTubeChannelInfoFromVideoURL(videoURL string) (*types.YouTubeChannelInfo, error) {
	videoID, err := ExtractVideoID(videoURL)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetYouTubeChannelInfoFromVideoURL", err)
	}

	video, err := getVideoInfo(service, videoID)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetYouTubeChannelInfoFromVideoURL", err)
	}

	channelID := video.Snippet.ChannelId
	channelURL := fmt.Sprintf("https://www.youtube.com/channel/%s", channelID)
	channelTitle := video.Snippet.ChannelTitle

	return &types.YouTubeChannelInfo{
		ChannelURL:   channelURL,
		ChannelTitle: channelTitle,
	}, nil
}

// extractVideoIDFromParts is a common logic to extract the video ID from the URL parts.
func extractVideoIDFromParts(parts []string, separator string) string {
	if len(parts) > 1 {
		videoID := parts[1]
		// Remove additional parameters
		if idx := strings.Index(videoID, separator); idx != -1 {
			videoID = videoID[:idx]
		}
		return videoID
	}
	return ""
}

// ExtractVideoID extracts the video ID from a YouTube URL.
//
// If the URL is invalid or the video ID cannot be extracted, it returns an error.
func ExtractVideoID(url string) (string, error) {
	if url == "" {
		return "", common.WrapErrorWithContext("ExtractVideoID", fmt.Errorf("URL이 비어있습니다"))
	}

	// The common YouTube URL format: https://www.youtube.com/watch?v=VIDEO_ID
	if strings.Contains(url, "youtube.com/watch?v=") {
		if videoID := extractVideoIDFromParts(strings.Split(url, "v="), "&"); videoID != "" {
			return videoID, nil
		}
	}

	// The embedded URL format: https://www.youtube.com/embed/VIDEO_ID
	if strings.Contains(url, "youtube.com/embed/") {
		if videoID := extractVideoIDFromParts(strings.Split(url, "embed/"), "?"); videoID != "" {
			return videoID, nil
		}
	}

	// The short URL format: https://youtu.be/VIDEO_ID
	if strings.Contains(url, "youtu.be/") {
		if videoID := extractVideoIDFromParts(strings.Split(url, "youtu.be/"), "?"); videoID != "" {
			return videoID, nil
		}
	}

	return "", common.WrapErrorWithContext("ExtractVideoID", fmt.Errorf("유효하지 않은 YouTube URL입니다"))
}
