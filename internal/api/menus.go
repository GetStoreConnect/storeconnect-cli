package api

// Menu represents a StoreConnect menu
type Menu struct {
	SCID  string     `json:"sc_id"`
	SFID  string     `json:"sfid,omitempty"`
	Name  string     `json:"name"`
	Items []MenuItem `json:"items,omitempty"`
}

// MenuItem represents a menu item
type MenuItem struct {
	Label    string     `json:"label"`
	URL      string     `json:"url"`
	Target   string     `json:"target,omitempty"`
	Children []MenuItem `json:"children,omitempty"`
}

// Menus handles menu-related endpoints
type Menus struct {
	client *Client
}

// NewMenus creates a new Menus service
func NewMenus(client *Client) *Menus {
	return &Menus{client: client}
}

// List returns all menus
func (m *Menus) List() ([]Menu, error) {
	var result struct {
		Menus []Menu `json:"menus"`
	}

	err := m.client.Get("/api/v1/menus", &result, nil)
	if err != nil {
		return nil, err
	}

	return result.Menus, nil
}

// Get retrieves a single menu by ID
func (m *Menus) Get(id string) (*Menu, error) {
	var result Menu
	err := m.client.Get("/api/v1/menus/"+id, &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
