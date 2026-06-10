package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ba-torment-data-process/internal/constants"

	"github.com/BeaverHouse/go-common/env"
	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/BeaverHouse/go-common/logger"
)

var (
	fileUploadURL = "https://api.tinyclover.com/file-manager/v1"
)

// MarshalAndUpload marshals data to JSON and uploads it.
func MarshalAndUpload(log logger.Logger, data any, path, fileName string, dryRun bool, successMsg string) error {
	var dataBytes []byte
	var err error

	if dryRun {
		dataBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		dataBytes, err = json.Marshal(data)
	}
	if err != nil {
		return constants.ErrDataEncode(fileName, err)
	}

	err = UploadFile(log, path, fileName, dataBytes, dryRun)
	if err != nil {
		return err
	}

	if successMsg != "" {
		log.Info(successMsg)
	}

	return nil
}

// UploadFile uploads a file to S3 via File Manager API.
func UploadFile(log logger.Logger, path string, fileName string, data []byte, dryRun bool) error {
	if dryRun {
		os.MkdirAll(filepath.Join("files", path), 0755)
		return os.WriteFile(filepath.Join("files", path, fileName), data, 0644)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return constants.ErrDataEncode(fileName, err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return constants.ErrDataEncode(fileName, err)
	}
	writer.WriteField("upload_path", path)
	writer.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/files/upload/raw", fileUploadURL),
		body,
	)
	if err != nil {
		return errorhandle.ErrInternal(err)
	}

	req.Header.Set("X-Access-Token", env.GetEnv("BA_ANALYZER_SERVICE_TOKEN", ""))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return constants.ErrUpload(fileName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return constants.ErrUpstreamBadStatus(fmt.Sprintf("upload status %d, body: %s", resp.StatusCode, string(respBody)))
	}

	log.Info("File uploaded successfully", logger.Field{Key: "file", Value: fileName})
	time.Sleep(2 * time.Second)

	return nil
}
