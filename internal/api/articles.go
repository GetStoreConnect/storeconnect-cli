package api

// Article represents a StoreConnect article
type Article struct {
	ID           string `json:"id,omitempty"`
	SCID         string `json:"sc_id"`
	SFID         string `json:"sfid,omitempty"`
	Title        string `json:"title"`
	Path         string `json:"path,omitempty"`
	Author       string `json:"author,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	Published    bool   `json:"published,omitempty"`
	BodyMarkdown string `json:"body_markdown,omitempty"`
}

// Articles handles article-related endpoints
type Articles struct {
	client *Client
}

// NewArticles creates a new Articles service
func NewArticles(client *Client) *Articles {
	return &Articles{client: client}
}

// List returns the store's articles
func (a *Articles) List() ([]Article, error) {
	var result struct {
		Data []Article `json:"data"`
	}

	err := a.client.Get("/api/v1/articles/content", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// Get retrieves a single article by sc_id, id, path or slug
func (a *Articles) Get(id string) (*Article, error) {
	var result Article
	err := a.client.Get("/api/v1/articles/"+id+"/content", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
