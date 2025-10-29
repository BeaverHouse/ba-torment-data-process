package update

import (
	"encoding/json"
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
		videoFilterBytes, err := json.Marshal(videoFilter)
		if err != nil {
			log.Printf("UpdateData - video filter: %v", err)
		} else {
			logic_upload.UploadFile("batorment/v3/video-filter", fileName, videoFilterBytes, dryRun)
			log.Printf("비디오 필터 업로드 완료: %s", raid)
		}
	}
}
