package update

import (
	"encoding/json"
	"fmt"
	"log"

	"ba-torment-data-process/internal/logic/filter"
	logic_upload "ba-torment-data-process/internal/logic/upload"
)

func UpdateVideoFilter(dryRun bool) {
	defer func() {
		log.Println("비디오 필터 업데이트 프로세스 완료")
	}()

	pendingRaids := []string{
		"3S22-1",
		"3S22-2",
		"3S22-3",
		"3S23-1",
		"3S23-2",
		"3S23-3",
		"S78-0",
		"S79-0",
		"S82-0",
	}

	if len(pendingRaids) == 0 {
		log.Println("업데이트할 총력전 ID가 없습니다.")
		return
	}

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
