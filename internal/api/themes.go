package api

// Theme represents a StoreConnect theme
type Theme struct {
	SCID      string                 `json:"sc_id"`
	SFID      string                 `json:"sfid,omitempty"`
	Name      string                 `json:"name"`
	Templates []ThemeTemplate        `json:"templates,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Assets    []ThemeAsset           `json:"assets,omitempty"`
}

// ThemeTemplate represents a template in a theme
type ThemeTemplate struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

// ThemeAsset represents an asset in a theme.
//
// The server contract for GET /api/v1/themes/:id returns assets as
// [{key, url, content_type, content_hash}]. Key is the asset's path within
// the theme (it may contain slashes), content_hash is the SHA-256 hex digest
// of the binary, used to skip uploads of unchanged files.
type ThemeAsset struct {
	Key         string `json:"key"`
	URL         string `json:"url,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	ContentHash string `json:"content_hash,omitempty"`
}

// Themes handles theme-related endpoints
type Themes struct {
	client *Client
}

// NewThemes creates a new Themes service
func NewThemes(client *Client) *Themes {
	return &Themes{client: client}
}

// List returns all themes for the store
func (t *Themes) List() ([]Theme, error) {
	var result struct {
		Themes []Theme `json:"themes"`
	}

	err := t.client.Get("/api/v1/themes", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Themes, nil
}

// Get retrieves a single theme by ID (sc_id or sfid)
func (t *Themes) Get(id string) (*Theme, error) {
	var result Theme
	err := t.client.Get("/api/v1/themes/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetBase retrieves the base theme
func (t *Themes) GetBase() (*Theme, error) {
	var result Theme
	err := t.client.Get("/api/v1/themes/base", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Create creates a new theme
func (t *Themes) Create(name string) (*Theme, error) {
	body := map[string]interface{}{
		"name": name,
	}

	var result Theme
	err := t.client.Post("/api/v1/themes", body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Delete deletes a theme by ID
func (t *Themes) Delete(id string) error {
	return t.client.Delete("/api/v1/themes/"+id, nil)
}
