package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/parse"
	"ba-torment-data-process/app/types"
)

func TestAronaAIParsing(t *testing.T) {
	seasonString := "S74-0"
	filesDir := "files"
	season := "S74"

	common.InitLogger()

	baTormentPartyPath := filepath.Join(filesDir, season+"-ba-torment-party.json")
	baTormentSummaryPath := filepath.Join(filesDir, season+"-ba-torment-summary.json")

	parsedPartyData, err := parse.ParsePartyDataFromAronaAI(seasonString)
	if err != nil {
		t.Fatalf("파티 데이터 파싱 실패: %v", err)
	}

	baTormentPartyBytes, err := os.ReadFile(baTormentPartyPath)
	if err != nil {
		t.Fatalf("BA Torment Party 파일 읽기 실패: %v", err)
	}

	var baTormentPartyData types.BATormentPartyData
	if err := json.Unmarshal(baTormentPartyBytes, &baTormentPartyData); err != nil {
		t.Fatalf("BA Torment Party JSON 파싱 실패: %v", err)
	}

	// Save parsed data
	parsedPartyDataBytes, err := json.Marshal(parsedPartyData)
	if err != nil {
		t.Fatalf("파티 데이터 직렬화 실패: %v", err)
	}
	os.WriteFile(filepath.Join(filesDir, season+"-parsed-party-data.json"), parsedPartyDataBytes, 0644)

	ComparePartyData(t, parsedPartyData, &baTormentPartyData, true)

	parsedSummaryData, err := parse.ProcessPartyDataToSummaryData(parsedPartyData)
	if err != nil {
		t.Fatalf("요약 데이터 파싱 실패: %v", err)
	}

	baTormentSummaryBytes, err := os.ReadFile(baTormentSummaryPath)
	if err != nil {
		t.Fatalf("BA Torment Summary 파일 읽기 실패: %v", err)
	}

	var baTormentSummaryData types.BATormentSummaryData
	if err := json.Unmarshal(baTormentSummaryBytes, &baTormentSummaryData); err != nil {
		t.Fatalf("BA Torment Summary JSON 파싱 실패: %v", err)
	}

	CompareSummaryData(t, parsedSummaryData, &baTormentSummaryData)
}
