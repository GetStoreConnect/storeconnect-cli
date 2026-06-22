package api

// ContentBlock represents a StoreConnect content block
type ContentBlock struct {
	ID         string `json:"id,omitempty"`
	SCID       string `json:"sc_id"`
	SFID       string `json:"sfid,omitempty"`
	Name       string `json:"name"`
	Identifier string `json:"identifier,omitempty"`
	Template   string `json:"template,omitempty"`
	Title      string `json:"title,omitempty"`
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
		Data []ContentBlock `json:"data"`
	}

	err := cb.client.Get("/api/v1/content_blocks/content", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Get retrieves a single content block by sc_id, id or identifier
func (cb *ContentBlocks) Get(id string) (*ContentBlock, error) {
	var result ContentBlock
	err := cb.client.Get("/api/v1/content_blocks/"+id+"/content", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
