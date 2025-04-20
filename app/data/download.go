package data

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/logic"
	"ba-torment-data-process/app/types"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
)

// Words
// Boss: Binah / ShiroKuro / Chesed / Kaitenger
// Armor: LightArmor (경장갑) / HeavyArmor (중장갑) / Unarmed (특수장갑) / ElasticArmor (탄력장갑)
// Location: Indoor (실내) / Outdoor (야외) / Street (시가지)
var (
	aronaAIURLs = map[string]string{
		"S74-0": "https://media.arona.ai/data/v3/raid/74/team-in20000",
		"S76-0": "https://media.arona.ai/data/v3/raid/76/team-in20000",
		"S21-1": "https://media.arona.ai/data/v3/eraid/21/team-in20000-Chesed_Indoor_LightArmor",
		"S21-2": "https://media.arona.ai/data/v3/eraid/21/team-in20000-Chesed_Indoor_HeavyArmor",
		"S21-3": "https://media.arona.ai/data/v3/eraid/21/team-in20000-Chesed_Indoor_ElasticArmor",
	}
	schaleDBURL  string = "https://schaledb.com/"
	googleAPIURL string = "https://storage.googleapis.com/info.herdatasam.me/BlueArchiveJP/"
)

// Get data from Arona.AI API.
func GetDataFromAronaAI(seasonString string) (*types.AronaAIData, error) {
	url, exists := aronaAIURLs[seasonString]
	if !exists {
		return nil, common.WrapErrorWithContext("GetDataFromAronaAI", fmt.Errorf("invalid season code: %s", seasonString))
	}

	data, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetDataFromAronaAI", err)
	}

	var jsonData types.AronaAIData
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetDataFromAronaAI > json.Unmarshal", err)
	}

	return &jsonData, nil
}

// Get rank data CSV from Google API.
func GetRankCSVFromGoogleAPI(seasonString string) (*csv.Reader, error) {
	season, category, err := logic.SplitSeasonString(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetRankCSVFromGoogleAPI", err)
	}

	var url string
	if category == 0 {
		url = googleAPIURL + "RaidRankData/" + season + "/FullData_Original.csv"
	} else {
		url = googleAPIURL + "RaidRankDataER/" + season + "/FullData_Original.csv"
	}

	reader, err := common.GetCSVReaderFromURL(url)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetRankCSVFromGoogleAPI", err)
	}

	return reader, nil
}

// Get party data CSV from Google API.
func GetPartyCSVFromGoogleAPI(seasonString string) (*csv.Reader, error) {
	season, category, err := logic.SplitSeasonString(seasonString)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetPartyCSVFromGoogleAPI", err)
	}

	var url string
	if category == 0 {
		url = googleAPIURL + "RaidRankData/" + season + "/TeamDataDetail_Original.csv"
	} else {
		url = googleAPIURL + "RaidRankDataER/" + season + "/TeamDataDetailBoss" + strconv.Itoa(category) + "_Original.csv"
	}

	reader, err := common.GetCSVReaderFromURL(url)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetPartyCSVFromGoogleAPI", err)
	}

	return reader, nil
}

// Get student data from SchaleDB.
func GetStudentDataFromSchaleDB() ([]types.SchaleDBStudentData, error) {
	url := schaleDBURL + "data/kr/students.min.json"

	data, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetStudentDataFromSchaleDB", err)
	}

	var jsonData map[string]types.SchaleDBStudentData
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetStudentDataFromSchaleDB > json.Unmarshal", err)
	}

	var studentData []types.SchaleDBStudentData
	for _, student := range jsonData {
		studentData = append(studentData, student)
	}

	return studentData, nil
}
