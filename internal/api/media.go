package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// MediaAsset represents a media asset
type MediaAsset struct {
	SCID        string `json:"sc_id"`
	SFID        string `json:"sfid,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Size        int64  `json:"size,omitempty"`
}

// UploadURLResponse represents the response from POST /api/v1/media/upload_url.
//
// The server hands back a target URL plus a set of form parameters that must
// accompany the binary in the follow-up multipart POST.
type UploadURLResponse struct {
	UploadURL    string            `json:"upload_url"`
	UploadParams map[string]string `json:"upload_params"`
	AssetID      string            `json:"asset_id,omitempty"`
}

// UploadResult is the parsed response from the upload host. The hosted URL is
// "secure_url" when present, falling back to "url".
type UploadResult struct {
	SecureURL string `json:"secure_url"`
	URL       string `json:"url"`
}

// HostedURL returns the canonical hosted location, preferring the secure URL.
func (r UploadResult) HostedURL() string {
	if r.SecureURL != "" {
		return r.SecureURL
	}
	return r.URL
}

// Media handles media-related endpoints
type Media struct {
	client *Client

	// uploader performs the multipart POST to the (external) upload host.
	// It is overridable in tests; production uses the default HTTP client.
	uploader *http.Client
}

// NewMedia creates a new Media service
func NewMedia(client *Client) *Media {
	return &Media{
		client:   client,
		uploader: &http.Client{Timeout: 60 * time.Second},
	}
}

// List returns all media assets
func (m *Media) List() ([]MediaAsset, error) {
	var result struct {
		Media []MediaAsset `json:"media"`
	}

	err := m.client.Get("/api/v1/media", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Media, nil
}

// UploadURL requests a signed upload target for a single file.
//
// fileType is the StoreConnect media kind ("image" or "document"), filename is
// the asset key/name, and contentType is the MIME type inferred from the file.
func (m *Media) UploadURL(fileType, filename, contentType string) (*UploadURLResponse, error) {
	body := map[string]interface{}{
		"file_type":    fileType,
		"filename":     filename,
		"content_type": contentType,
	}

	var result UploadURLResponse
	err := m.client.Post("/api/v1/media/upload_url", body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UploadFile streams the binary to the upload host as a multipart POST,
// sending every upload param as a form field plus the bytes as the "file"
// field. It returns the hosted URL the server assigned to the upload.
func (m *Media) UploadFile(upload *UploadURLResponse, filename string, content []byte) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range upload.UploadParams {
		if err := writer.WriteField(key, value); err != nil {
			return "", fmt.Errorf("failed to write upload param %q: %w", key, err)
		}
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create file field: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return "", fmt.Errorf("failed to write file contents: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize upload body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, upload.UploadURL, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to build upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := m.uploader.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result UploadResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	hosted := result.HostedURL()
	if hosted == "" {
		return "", fmt.Errorf("upload response missing hosted URL")
	}

	return hosted, nil
}
