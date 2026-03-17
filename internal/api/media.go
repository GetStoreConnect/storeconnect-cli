package api

// MediaAsset represents a media asset
type MediaAsset struct {
	SCID        string `json:"sc_id"`
	SFID        string `json:"sfid,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Size        int64  `json:"size,omitempty"`
}

// UploadURLResponse represents the upload URL response
type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	AssetID   string `json:"asset_id"`
}

// Media handles media-related endpoints
type Media struct {
	client *Client
}

// NewMedia creates a new Media service
func NewMedia(client *Client) *Media {
	return &Media{client: client}
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

// GetUploadURL gets a presigned upload URL
func (m *Media) GetUploadURL(filename, contentType string) (*UploadURLResponse, error) {
	body := map[string]interface{}{
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

// ConfirmUpload confirms a completed upload
func (m *Media) ConfirmUpload(assetID string) (*MediaAsset, error) {
	body := map[string]interface{}{
		"asset_id": assetID,
	}

	var result MediaAsset
	err := m.client.Post("/api/v1/media/confirm_upload", body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
