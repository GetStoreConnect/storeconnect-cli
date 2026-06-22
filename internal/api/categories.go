package api

// Category represents a StoreConnect product category
type Category struct {
	ID          string     `json:"id,omitempty"`
	SCID        string     `json:"sc_id"`
	SFID        string     `json:"sfid,omitempty"`
	Name        string     `json:"name"`
	DisplayName string     `json:"display_name,omitempty"`
	Path        string     `json:"path,omitempty"`
	Children    []Category `json:"children,omitempty"`
}

// Categories handles category-related endpoints
type Categories struct {
	client *Client
}

// NewCategories creates a new Categories service
func NewCategories(client *Client) *Categories {
	return &Categories{client: client}
}

// List returns all categories
func (c *Categories) List() ([]Category, error) {
	var result struct {
		Data []Category `json:"data"`
	}

	err := c.client.Get("/api/v1/categories/content", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Tree returns the category hierarchy
func (c *Categories) Tree() ([]Category, error) {
	var result struct {
		Data []Category `json:"data"`
	}

	err := c.client.Get("/api/v1/categories/tree", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Get retrieves a single category by sc_id, id or path
func (c *Categories) Get(id string) (*Category, error) {
	var result Category
	err := c.client.Get("/api/v1/categories/"+id+"/content", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
