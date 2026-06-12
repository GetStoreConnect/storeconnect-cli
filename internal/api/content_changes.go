package api

import "errors"

// ContentChange represents a draft change set
type ContentChange struct {
	SCID       string                 `json:"sc_id"`
	SFID       string                 `json:"sfid,omitempty"`
	Status     string                 `json:"status"`
	CustomData map[string]interface{} `json:"custom_data,omitempty"`
}

// ContentChangeRequest represents a request to create/update content changes
type ContentChangeRequest struct {
	ThemeID   string                  `json:"theme_id,omitempty"`
	Templates []ContentChangeTemplate `json:"templates,omitempty"`
}

// ContentChangeTemplate represents a template change
type ContentChangeTemplate struct {
	Key     string `json:"key"`
	Content string `json:"content"`
	Action  string `json:"action"` // "create", "update", or "delete"
}

// PreviewURLResponse represents the preview URL response
type PreviewURLResponse struct {
	PreviewURL string `json:"preview_url"`
}

// PublishResponse represents the publish response
type PublishResponse struct {
	Message     string `json:"message"`
	Status      string `json:"status"`
	SCID        string `json:"sc_id"`
	Note        string `json:"note,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

// ContentChanges handles content change endpoints
type ContentChanges struct {
	client *Client
}

// NewContentChanges creates a new ContentChanges service
func NewContentChanges(client *Client) *ContentChanges {
	return &ContentChanges{client: client}
}

// Create creates a new content change (draft)
func (cc *ContentChanges) Create(themeID string) (*ContentChange, error) {
	body := map[string]interface{}{
		"theme_id": themeID,
	}

	var result ContentChange
	err := cc.client.Post("/api/v1/content_changes", body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Update adds template changes to an existing content change
func (cc *ContentChanges) Update(id string, themeID string, templates []ContentChangeTemplate) error {
	body := map[string]interface{}{
		"theme_id":  themeID,
		"templates": templates,
	}

	var result ContentChange
	return cc.client.Patch("/api/v1/content_changes/"+id, body, &result)
}

// GetPreviewURL gets the preview URL for a content change
func (cc *ContentChanges) GetPreviewURL(id string) (string, error) {
	var result PreviewURLResponse
	err := cc.client.Get("/api/v1/content_changes/"+id+"/preview_url", &result, nil)
	if err != nil {
		return "", err
	}
	return result.PreviewURL, nil
}

// Publish publishes a content change to live.
//
// Production stores answer 403 when the change still needs approval - the
// server moves the draft to review, so that is a successful submission for
// the CLI, not an error.
func (cc *ContentChanges) Publish(id string) (*PublishResponse, error) {
	var result PublishResponse
	err := cc.client.Post("/api/v1/content_changes/"+id+"/publish", nil, &result)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 403 {
			return &PublishResponse{
				Message: apiErr.Message,
				Status:  "review",
				SCID:    id,
			}, nil
		}
		return nil, err
	}
	return &result, nil
}

// Submit pushes a draft content change to Salesforce for review without
// publishing it - the Change Request flow
func (cc *ContentChanges) Submit(id string) (*ContentChange, error) {
	var result ContentChange
	err := cc.client.Post("/api/v1/content_changes/"+id+"/submit", nil, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Get retrieves a content change by ID
func (cc *ContentChanges) Get(id string) (*ContentChange, error) {
	var result ContentChange
	err := cc.client.Get("/api/v1/content_changes/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
