package api

// Page represents a StoreConnect page
type Page struct {
	ID       string `json:"id,omitempty"`
	SCID     string `json:"sc_id"`
	SFID     string `json:"sfid,omitempty"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Slug     string `json:"slug,omitempty"`
	Path     string `json:"path,omitempty"`
	Visible  bool   `json:"visible,omitempty"`
}

// Pages handles page-related endpoints
type Pages struct {
	client *Client
}

// NewPages creates a new Pages service
func NewPages(client *Client) *Pages {
	return &Pages{client: client}
}

// List returns all pages
func (p *Pages) List() ([]Page, error) {
	var result struct {
		Data []Page `json:"data"`
	}

	err := p.client.Get("/api/v1/pages/content", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Get retrieves a single page by sc_id, id or path
func (p *Pages) Get(id string) (*Page, error) {
	var result Page
	err := p.client.Get("/api/v1/pages/"+id+"/content", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create stages a new page as a draft content change
func (p *Pages) Create(page *Page) (*Page, error) {
	var result Page
	err := p.client.Post("/api/v1/pages/content", page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update stages changes to an existing page as a draft content change
func (p *Pages) Update(id string, page *Page) (*Page, error) {
	var result Page
	err := p.client.Patch("/api/v1/pages/"+id+"/content", page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
