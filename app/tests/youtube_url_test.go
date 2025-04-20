package tests

import (
	"testing"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/update_youtube"

	"github.com/stretchr/testify/require"
)

var (
	videoURLs []string = []string{
		"https://www.youtube.com/embed/JpnNFe1k9zA?feature=shared",
		"https://www.youtube.com/embed/Vns6cAtpQ8w",
	}
	channelURLs []string = []string{
		"https://www.youtube.com/channel/UCbqe7u7_489e5zs8Bj6NdFQ",
		"https://www.youtube.com/channel/UCbWZkdwVQDY4jcLvRB9UtLQ",
	}
	usernames []string = []string{
		"みじん子ちゃんねる。",
		"kkbn",
	}
	videoURLsWithVariousFormats = []string{
		"https://www.youtube.com/watch?v=bn8gP5N8hqM",
		"https://www.youtube.com/watch?v=bn8gP5N8hqM&feature=youtu.be",
		"https://www.youtube.com/embed/bn8gP5N8hqM",
		"https://www.youtube.com/embed/bn8gP5N8hqM?feature=shared",
		"https://youtu.be/bn8gP5N8hqM",
		"https://youtu.be/bn8gP5N8hqM?t=43",
	}
	expectedVideoID string = "bn8gP5N8hqM"
)

func TestYouTubeChannelInfo(t *testing.T) {
	common.InitLogger()

	for i, videoURL := range videoURLs {
		channelInfo, err := update_youtube.GetYouTubeChannelInfoFromVideoURL(videoURL)
		require.NoError(t, err)
		require.Equal(t, channelInfo.ChannelURL, channelURLs[i], "URL: %s, Channel URL expected: %s, Got: %s", videoURL, channelURLs[i], channelInfo.ChannelURL)
		require.Equal(t, channelInfo.ChannelTitle, usernames[i], "URL: %s, Channel Title expected: %s, Got: %s", videoURL, usernames[i], channelInfo.ChannelTitle)
	}
}

func TestExtractVideoID(t *testing.T) {
	common.InitLogger()

	for _, url := range videoURLsWithVariousFormats {
		result, err := update_youtube.ExtractVideoID(url)
		require.NoError(t, err)
		require.Equal(t, result, expectedVideoID, "URL: %s, ID expected: %s, Got: %s", url, expectedVideoID, result)
	}
}
