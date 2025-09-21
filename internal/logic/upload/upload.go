package logic_upload

import (
	"ba-torment-data-process/internal/logic"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

var (
	fileUploadURL = "https://api.tinyclover.com/file-manager/v1"
)

func getAdminAPIKey() string {
	return logic.GetEnv("FILEMANAGER_ADMIN_API_KEY", "")
}

// Uploads a file to the S3 via File Manager API.
func UploadFile(path string, fileName string, data []byte, dryRun bool) error {
	if dryRun {
		os.MkdirAll(filepath.Join("files", path), 0755)
		return os.WriteFile(filepath.Join("files", path, fileName), data, 0644)
	} else {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			log.Fatalf("Failed to create form file: %v", err)
		}
		if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
			log.Fatalf("Failed to copy file data: %v", err)
		}
		writer.WriteField("upload_path", path)
		writer.Close()

		req, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s/files/upload", fileUploadURL),
			body,
		)
		if err != nil {
			log.Fatalf("API request failed: %v", err)
		}

		req.Header.Set("X-API-Key", getAdminAPIKey())
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("API request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("failed to upload image: status %d, body: %s", resp.StatusCode, string(body))
		}
		return nil
	}
}
