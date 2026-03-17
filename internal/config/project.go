package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ServerInfo represents project-level server information
type ServerInfo struct {
	URL                 string    `yaml:"url"`
	StoreconnectVersion string    `yaml:"storeconnect_version,omitempty"`
	BaseThemeVersion    string    `yaml:"base_theme_version,omitempty"`
	LastSync            time.Time `yaml:"last_sync,omitempty"`
}

// Project manages project-level configuration in .storeconnect/config.yml
type Project struct {
	DefaultServer string                `yaml:"default_server,omitempty"`
	Servers       map[string]ServerInfo `yaml:"servers"`
	path          string
}

// NewProject loads or creates project configuration
func NewProject() (*Project, error) {
	configPath := filepath.Join(".storeconnect", "config.yml")

	proj := &Project{
		Servers: make(map[string]ServerInfo),
		path:    configPath,
	}

	// Load existing config if file exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read project config: %w", err)
		}

		if err := yaml.Unmarshal(data, proj); err != nil {
			return nil, fmt.Errorf("failed to parse project config: %w", err)
		}
	}

	return proj, nil
}

// Exists checks if project configuration exists
func (p *Project) Exists() bool {
	_, err := os.Stat(p.path)
	return err == nil
}

// Init initializes a new project directory
func (p *Project) Init(projectName string) error {
	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create .storeconnect directory
	configDir := filepath.Join(projectName, ".storeconnect")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create themes directory
	themesDir := filepath.Join(projectName, "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Create initial config file
	p.path = filepath.Join(configDir, "config.yml")
	return p.Save()
}

// AddServer adds or updates server information
func (p *Project) AddServer(alias, url string, setAsDefault bool) error {
	p.Servers[alias] = ServerInfo{
		URL: url,
	}

	if setAsDefault || p.DefaultServer == "" {
		p.DefaultServer = alias
	}

	return p.Save()
}

// UpdateServerInfo updates server version information
func (p *Project) UpdateServerInfo(alias, storeconnectVersion, baseThemeVersion string) error {
	info, ok := p.Servers[alias]
	if !ok {
		return fmt.Errorf("server %s not found", alias)
	}

	if storeconnectVersion != "" {
		info.StoreconnectVersion = storeconnectVersion
	}
	if baseThemeVersion != "" {
		info.BaseThemeVersion = baseThemeVersion
	}
	info.LastSync = time.Now()

	p.Servers[alias] = info
	return p.Save()
}

// GetServer retrieves server information by alias
func (p *Project) GetServer(alias string) (*ServerInfo, bool) {
	info, ok := p.Servers[alias]
	return &info, ok
}

// RemoveServer removes a server from project config
func (p *Project) RemoveServer(alias string) error {
	delete(p.Servers, alias)

	// Clear default if it was the removed server
	if p.DefaultServer == alias {
		p.DefaultServer = ""
		// Set first available server as default
		for serverAlias := range p.Servers {
			p.DefaultServer = serverAlias
			break
		}
	}

	return p.Save()
}

// Save writes project configuration to disk
func (p *Project) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(p.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(p.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

// GetDefaultServer returns the default server alias
func (p *Project) GetDefaultServer() string {
	return p.DefaultServer
}

// SetDefaultServer sets the default server
func (p *Project) SetDefaultServer(alias string) error {
	if _, ok := p.Servers[alias]; !ok {
		return fmt.Errorf("server %s not found", alias)
	}
	p.DefaultServer = alias
	return p.Save()
}
