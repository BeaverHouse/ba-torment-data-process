package update

import (
	"fmt"
	"log"

	"ba-torment-data-process/internal/logic/filter"
	logic_upload "ba-torment-data-process/internal/logic/upload"
)

func UpdateVideoFilter(dryRun bool, pendingRaids []string) {
	defer func() {
		log.Println("비디오 필터 업데이트 프로세스 완료")
	}()

	for _, raid := range pendingRaids {
		fileName := fmt.Sprintf("%s.json", raid)

		// Create and upload video filter
		videoFilter := filter.CreateVideoFilter(raid)
		if err := logic_upload.MarshalAndUpload(videoFilter, "batorment/v3/video-filter", fileName, dryRun, fmt.Sprintf("비디오 필터 업로드 완료: %s", raid)); err != nil {
			log.Printf("Failed to upload video filter: %v", err)
		}
	}
}
