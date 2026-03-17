package api

// InfoResponse represents the server info response
type InfoResponse struct {
	StoreconnectVersion string `json:"storeconnect_version"`
	BaseThemeVersion    string `json:"base_theme_version"`
	OrgID               string `json:"org_id"`
	StoreSFID           string `json:"store_sfid"`
}

// Auth handles authentication endpoints
type Auth struct {
	client *Client
}

// NewAuth creates a new Auth service
func NewAuth(client *Client) *Auth {
	return &Auth{client: client}
}

// Info validates credentials and returns server information
func (a *Auth) Info() (*InfoResponse, error) {
	var result InfoResponse
	err := a.client.Get("/api/v1/info", &result, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
