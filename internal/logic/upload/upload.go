package logic_upload

import (
	"ba-torment-data-process/app/common"
	"ba-torment-data-process/internal/logic"
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"
)

const schaleDBURL = "https://schale-db.com/"

var (
	fileUploadURL string
	adminAPIKey   string
)

func init() {
	common.LoadEnv()
	fileUploadURL = "https://api.tinyclover.com/file-manager/v1"
	adminAPIKey = logic.GetEnv("ADMIN_API_KEY", "")
}

// Uploads a file to the Oracle Object Storage.
func UploadFile(path string, fileName string, data []byte, dryRun bool) error {
	if dryRun {
		// Create directory if it doesn't exist
		os.MkdirAll(filepath.Join("files", path), 0755)
		// Save to JSON
		return os.WriteFile(filepath.Join("files", path, fileName), data, 0644)
	} else {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			return common.WrapErrorWithContext("UploadFile", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
			return common.WrapErrorWithContext("UploadFile", err)
		}
		writer.WriteField("upload_path", path)
		writer.Close()

		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/files/upload", fileUploadURL),
			body)
		if err != nil {
			return common.WrapErrorWithContext("UploadFile", err)
		}

		req.Header.Set("X-API-Key", adminAPIKey)
		req.Header.Set("Content-Type", writer.FormDataContentType())

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
}

// Uploads the character image from SchaleDB to the Oracle Object Storage.
func UploadCharacterImage(id int, isTest bool, dryRun bool) error {

	imgBytes, err := common.GetDataFromURL(schaleDBURL + "images/student/icon/" + strconv.Itoa(id) + ".webp")
	if err != nil {
		return common.WrapErrorWithContext("UploadCharacterImage", err)
	}

	path := "batorment/character"

	err = UploadFile(path, strconv.Itoa(id)+".webp", imgBytes, dryRun)
	if err != nil {
		return common.WrapErrorWithContext("UploadCharacterImage", err)
	}

	return nil
}
