package api

// ContentBlock represents a StoreConnect content block
type ContentBlock struct {
	SCID    string `json:"sc_id"`
	SFID    string `json:"sfid,omitempty"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Type    string `json:"type,omitempty"`
}

// ContentBlocks handles content block-related endpoints
type ContentBlocks struct {
	client *Client
}

// NewContentBlocks creates a new ContentBlocks service
func NewContentBlocks(client *Client) *ContentBlocks {
	return &ContentBlocks{client: client}
}

// List returns all content blocks
func (cb *ContentBlocks) List() ([]ContentBlock, error) {
	var result struct {
		ContentBlocks []ContentBlock `json:"content_blocks"`
	}

	err := cb.client.Get("/api/v1/content_blocks", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.ContentBlocks, nil
}

// Get retrieves a single content block by ID
func (cb *ContentBlocks) Get(id string) (*ContentBlock, error) {
	var result ContentBlock
	err := cb.client.Get("/api/v1/content_blocks/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
