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
)

var (
	fileUploadURL string
	adminAPIKey   string
)

func init() {
	common.LoadEnv()
	fileUploadURL = "https://api.tinyclover.com/file-manager/v1"
	adminAPIKey = logic.GetEnv("ADMIN_API_KEY", "")
}

// Uploads a file to the S3 via File Manager API.
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
			return fmt.Errorf("failed to create form file: %v", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
			return fmt.Errorf("failed to copy file data: %v", err)
		}
		writer.WriteField("upload_path", path)
		writer.Close()

		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/files/upload", fileUploadURL),
			body)
		if err != nil {
			return fmt.Errorf("API request failed: %v", err)
		}

		req.Header.Set("X-API-Key", adminAPIKey)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("API request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to upload image: status %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
}
