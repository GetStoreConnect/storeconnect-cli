package api

// Category represents a StoreConnect category
type Category struct {
	SCID     string     `json:"sc_id"`
	SFID     string     `json:"sfid,omitempty"`
	Name     string     `json:"name"`
	Slug     string     `json:"slug,omitempty"`
	ParentID string     `json:"parent_id,omitempty"`
	Children []Category `json:"children,omitempty"`
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
		Categories []Category `json:"categories"`
	}

	err := c.client.Get("/api/v1/categories", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Categories, nil
}

// Tree returns category hierarchy
func (c *Categories) Tree() ([]Category, error) {
	var result struct {
		Categories []Category `json:"categories"`
	}

	err := c.client.Get("/api/v1/categories/tree", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Categories, nil
}

// Get retrieves a single category by ID
func (c *Categories) Get(id string) (*Category, error) {
	var result Category
	err := c.client.Get("/api/v1/categories/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
