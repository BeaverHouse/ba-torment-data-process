package data

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/app/types"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

var (
	oracleUploadURL string
	// SchaleDB URL is already declared
)

func init() {
	common.LoadEnv()
	oracleUploadURL = common.GetEssentialEnv("BATORMENT_UPLOAD_URL")
}

// Uploads a file to the Oracle Object Storage.
func uploadFile(path string, fileName string, data []byte) error {
	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/o/batorment/%s/%s", oracleUploadURL, path, fileName),
		bytes.NewReader(data))
	if err != nil {
		return common.WrapErrorWithContext("UploadFile", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return common.WrapErrorWithContext("UploadFile", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return common.WrapErrorWithContext("UploadFile", fmt.Errorf("failed to upload image: status %d, body: %s", resp.StatusCode, string(body)))
	}

	common.LogInfo("File uploaded", zap.String("path", path), zap.String("fileName", fileName))
	return nil
}

// Uploads the character image from SchaleDB to the Oracle Object Storage.
func UploadCharacterImage(id int, isTest bool) error {

	imgBytes, err := common.GetDataFromURL(schaleDBURL + "images/student/icon/" + strconv.Itoa(id) + ".webp")
	if err != nil {
		return common.WrapErrorWithContext("UploadCharacterImage", err)
	}

	path := "character"
	if isTest {
		path = "test/character"
	}

	err = uploadFile(path, strconv.Itoa(id)+".webp", imgBytes)
	if err != nil {
		return common.WrapErrorWithContext("UploadCharacterImage", err)
	}

	return nil
}

// Uploads the party data JSON to the Oracle Object Storage.
func UploadPartyDataJSON(data *types.BATormentPartyData, seasonString string, isTest bool) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return common.WrapErrorWithContext("UploadPartyDataJSON > json.Marshal", err)
	}

	path := "v2/party"
	if isTest {
		path = "test/party"
	}

	err = uploadFile(path, seasonString+".json", jsonData)
	if err != nil {
		return common.WrapErrorWithContext("UploadPartyDataJSON", err)
	}

	return nil
}

// Uploads the summary data JSON to the Oracle Object Storage.
func UploadSummaryDataJSON(data *types.BATormentSummaryData, seasonString string, isTest bool) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return common.WrapErrorWithContext("UploadSummaryDataJSON > json.Marshal", err)
	}

	path := "v2/summary"
	if isTest {
		path = "test/summary"
	}

	err = uploadFile(path, seasonString+".json", jsonData)
	if err != nil {
		return common.WrapErrorWithContext("UploadSummaryDataJSON", err)
	}

	return nil
}
