package update

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"ba-torment-data-process/internal/db/postgres"
	"ba-torment-data-process/internal/logic"
	"ba-torment-data-process/internal/logic/filter"
	"ba-torment-data-process/internal/logic/parse"
	logic_upload "ba-torment-data-process/internal/logic/upload"
	"ba-torment-data-process/internal/types"
)

func UpdateData(dryRun bool) {
	defer func() {
		log.Println("총력전 데이터 업데이트 프로세스 완료")
	}()

	// Create postgres config from environment
	postgresPort := logic.GetEnv("POSTGRES_PORT", "5432")
	postgresPortInt, err := strconv.Atoi(postgresPort)
	if err != nil {
		panic("Failed to convert POSTGRES_PORT to int: " + postgresPort)
	}

	// Create postgres config from environment
	cfg := types.PostgresConfig{
		Host:     logic.GetEnv("POSTGRES_HOST", "localhost"),
		Port:     postgresPortInt,
		User:     logic.GetEnv("POSTGRES_USER", "postgres"),
		Password: logic.GetEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:   logic.GetEnv("POSTGRES_DB", "postgres"),
		SSLMode:  logic.GetEnv("POSTGRES_SSLMODE", "disable"),
	}

	// Create database pool
	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Printf("UpdateData - creating pool: %v", err)
		return
	}
	defer pool.Close()

	pendingRaids := []types.Raid{
		{
			RaidID: "S80-0",
			Name:   "테스트용",
			Status: "PENDING",
		},
	}

	if len(pendingRaids) == 0 {
		log.Println("업데이트할 총력전 ID가 없습니다.")
		return
	}

	for _, raid := range pendingRaids {
		var partyData *types.BATormentPartyData
		var summaryData *types.BATormentSummaryData

		fileName := fmt.Sprintf("%s.json", raid.RaidID)

		// Get from Arona AI
		partyData, filterResult := parse.ParsePartyDataFromAronaAI(raid.RaidID)

		partyDataBytes, err := json.Marshal(partyData)
		if err != nil {
			log.Printf("UpdateData: %v", err)
			continue
		}
		filterResultBytes, err := json.Marshal(filterResult)
		if err != nil {
			log.Printf("UpdateData: %v", err)
			continue
		}
		logic_upload.UploadFile("v3/party", fileName, partyDataBytes, dryRun)
		logic_upload.UploadFile("v3/filter", fileName, filterResultBytes, dryRun)

		// Create and upload lunatic filter
		lunaticFilter := filter.CreateLunaticFilter(partyData, filterResult)
		lunaticFilterBytes, err := json.Marshal(lunaticFilter)
		if err != nil {
			log.Printf("UpdateData - lunatic filter: %v", err)
		} else {
			logic_upload.UploadFile("v3/lunatic-filter", fileName, lunaticFilterBytes, dryRun)
			log.Printf("루나틱 필터 업로드 완료: %s", raid.RaidID)
		}

		// Create and upload non-lunatic filter
		nonLunaticFilter := filter.CreateNonLunaticFilter(partyData, filterResult)
		nonLunaticFilterBytes, err := json.Marshal(nonLunaticFilter)
		if err != nil {
			log.Printf("UpdateData - non-lunatic filter: %v", err)
		} else {
			logic_upload.UploadFile("v3/nonlunatic-filter", fileName, nonLunaticFilterBytes, dryRun)
			log.Printf("논루나틱 필터 업로드 완료: %s", raid.RaidID)
		}

		summaryData, err = parse.ProcessPartyDataToSummaryData(partyData)
		if err != nil {
			log.Printf("UpdateData: %v", err)
			continue
		}
		summaryDataBytes, err := json.Marshal(summaryData)
		if err != nil {
			log.Printf("UpdateData: %v", err)
			continue
		}
		logic_upload.UploadFile("v3/summary", fileName, summaryDataBytes, dryRun)
	}
}

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
