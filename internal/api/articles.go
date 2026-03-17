package api

// Article represents a StoreConnect article
type Article struct {
	SCID    string `json:"sc_id"`
	SFID    string `json:"sfid,omitempty"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Slug    string `json:"slug,omitempty"`
}

// Articles handles article-related endpoints
type Articles struct {
	client *Client
}

// NewArticles creates a new Articles service
func NewArticles(client *Client) *Articles {
	return &Articles{client: client}
}

// List returns all articles
func (a *Articles) List() ([]Article, error) {
	var result struct {
		Articles []Article `json:"articles"`
	}

	err := a.client.Get("/api/v1/articles", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Articles, nil
}

// Get retrieves a single article by ID
func (a *Articles) Get(id string) (*Article, error) {
	var result Article
	err := a.client.Get("/api/v1/articles/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
