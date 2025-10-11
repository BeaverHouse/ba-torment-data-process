package update

import (
	"encoding/json"
	"fmt"
	"log"

	"ba-torment-data-process/internal/logic/filter"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"ba-torment-data-process/internal/types"
)

func UpdateVideoFilter(dryRun bool) {
	defer func() {
		log.Println("비디오 필터 업데이트 프로세스 완료")
	}()

	pendingRaids := []types.Raid{
		{
			RaidID: "3S22-1",
			Name:   "테스트용",
			Status: "PENDING",
		},
		{
			RaidID: "3S22-2",
			Name:   "테스트용",
			Status: "PENDING",
		},
		{
			RaidID: "3S22-3",
			Name:   "테스트용",
			Status: "PENDING",
		},
	}

	if len(pendingRaids) == 0 {
		log.Println("업데이트할 총력전 ID가 없습니다.")
		return
	}

	for _, raid := range pendingRaids {
		fileName := fmt.Sprintf("%s.json", raid.RaidID)

		// Create and upload video filter
		videoFilter := filter.CreateVideoFilter(raid.RaidID)
		videoFilterBytes, err := json.Marshal(videoFilter)
		if err != nil {
			log.Printf("UpdateData - video filter: %v", err)
		} else {
			logic_upload.UploadFile("v3/video-filter", fileName, videoFilterBytes, dryRun)
			log.Printf("비디오 필터 업로드 완료: %s", raid.RaidID)
		}
	}
}
