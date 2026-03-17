package api

// Page represents a StoreConnect page
type Page struct {
	SCID    string `json:"sc_id"`
	SFID    string `json:"sfid,omitempty"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Slug    string `json:"slug"`
	Path    string `json:"path,omitempty"`
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
		Pages []Page `json:"pages"`
	}

	err := p.client.Get("/api/v1/pages", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Pages, nil
}

// Get retrieves a single page by ID
func (p *Pages) Get(id string) (*Page, error) {
	var result Page
	err := p.client.Get("/api/v1/pages/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new page
func (p *Pages) Create(page *Page) (*Page, error) {
	var result Page
	err := p.client.Post("/api/v1/pages", page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates an existing page
func (p *Pages) Update(id string, page *Page) (*Page, error) {
	var result Page
	err := p.client.Put("/api/v1/pages/"+id, page, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
