package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

// ServerCredentials represents credentials for a single server
type ServerCredentials struct {
	URL       string `yaml:"url"`
	OrgID     string `yaml:"org_id,omitempty"`
	StoreSFID string `yaml:"store_sfid"`
	APIKey    string `yaml:"api_key"`
}

// Credentials manages global credentials stored in ~/.storeconnect/credentials.yml
type Credentials struct {
	Servers map[string]ServerCredentials `yaml:"servers"`
	path    string
}

// NewCredentials loads or creates credentials
func NewCredentials() (*Credentials, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	credPath := filepath.Join(home, ".storeconnect", "credentials.yml")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return nil, fmt.Errorf("failed to create credentials directory: %w", err)
	}

	creds := &Credentials{
		Servers: make(map[string]ServerCredentials),
		path:    credPath,
	}

	// Load existing credentials if file exists
	if _, err := os.Stat(credPath); err == nil {
		data, err := os.ReadFile(credPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials: %w", err)
		}

		if err := yaml.Unmarshal(data, creds); err != nil {
			return nil, fmt.Errorf("failed to parse credentials: %w", err)
		}
	}

	return creds, nil
}

// AddServer adds or updates server credentials
func (c *Credentials) AddServer(alias, url, storeSFID, apiKey string, orgID string) error {
	c.Servers[alias] = ServerCredentials{
		URL:       url,
		OrgID:     orgID,
		StoreSFID: storeSFID,
		APIKey:    apiKey,
	}
	return c.Save()
}

// GetServer retrieves server credentials by alias
func (c *Credentials) GetServer(alias string) (*ServerCredentials, bool) {
	cred, ok := c.Servers[alias]
	return &cred, ok
}

// RemoveServer removes server credentials
func (c *Credentials) RemoveServer(alias string) error {
	delete(c.Servers, alias)
	return c.Save()
}

// Save writes credentials to disk with secure permissions
func (c *Credentials) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with 0600 permissions (read/write owner only)
	if err := os.WriteFile(c.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// Path returns the path to the credentials file
func (c *Credentials) Path() string {
	return c.path
}
