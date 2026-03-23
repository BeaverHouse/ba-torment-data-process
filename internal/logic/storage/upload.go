package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	fileUploadURL = "https://api.tinyclover.com/file-manager/v1"
)

// MarshalAndUpload marshals data to JSON and uploads it.
func MarshalAndUpload(data any, path, fileName string, dryRun bool, successMsg string) error {
	var dataBytes []byte
	var err error

	if dryRun {
		dataBytes, err = json.MarshalIndent(data, "", "  ")
	} else {
		dataBytes, err = json.Marshal(data)
	}
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = UploadFile(path, fileName, dataBytes, dryRun)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	if successMsg != "" {
		log.Println(successMsg)
	}

	return nil
}

// UploadFile uploads a file to S3 via File Manager API.
func UploadFile(path string, fileName string, data []byte, dryRun bool) error {
	if dryRun {
		os.MkdirAll(filepath.Join("files", path), 0755)
		return os.WriteFile(filepath.Join("files", path, fileName), data, 0644)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create form file: %v", err))
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		panic(fmt.Sprintf("Failed to copy file data: %v", err))
	}
	writer.WriteField("upload_path", path)
	writer.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/files/upload", fileUploadURL),
		body,
	)
	if err != nil {
		panic(fmt.Sprintf("API request failed: %v", err))
	}

	req.Header.Set("X-Access-Token", os.Getenv("BA_ANALYZER_SERVICE_TOKEN"))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("API request failed: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("failed to upload image: status %d, body: %s", resp.StatusCode, string(body)))
	}

	log.Println("File uploaded successfully: ", fileName)
	time.Sleep(2 * time.Second)

	return nil
}
