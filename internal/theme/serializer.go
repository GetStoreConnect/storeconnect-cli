package theme

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"gopkg.in/yaml.v3"
)

// Serializer converts API theme JSON to local file structure
type Serializer struct {
	basePath string
}

// NewSerializer creates a new theme serializer
func NewSerializer(basePath string) *Serializer {
	return &Serializer{basePath: basePath}
}

// Serialize writes a theme to the local filesystem
func (s *Serializer) Serialize(theme *api.Theme) error {
	themePath := filepath.Join(s.basePath, "themes", theme.Name)

	// Create theme directory
	if err := os.MkdirAll(themePath, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	// Write theme.yml metadata
	if err := s.writeThemeMetadata(themePath, theme); err != nil {
		return fmt.Errorf("failed to write theme metadata: %w", err)
	}

	// Write templates
	if err := s.writeTemplates(themePath, theme.Templates); err != nil {
		return fmt.Errorf("failed to write templates: %w", err)
	}

	// Write variables.json
	if err := s.writeJSON(filepath.Join(themePath, "variables.json"), theme.Variables); err != nil {
		return fmt.Errorf("failed to write variables: %w", err)
	}

	// Write assets.json
	if err := s.writeJSON(filepath.Join(themePath, "assets.json"), theme.Assets); err != nil {
		return fmt.Errorf("failed to write assets: %w", err)
	}

	return nil
}

func (s *Serializer) writeThemeMetadata(themePath string, theme *api.Theme) error {
	metadata := map[string]interface{}{
		"name":  theme.Name,
		"sc_id": theme.SCID,
	}
	if theme.SFID != "" {
		metadata["sfid"] = theme.SFID
	}

	data, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(themePath, "theme.yml"), data, 0644)
}

func (s *Serializer) writeTemplates(themePath string, templates []api.ThemeTemplate) error {
	if len(templates) == 0 {
		return nil
	}

	templatesDir := filepath.Join(themePath, "templates")

	for _, template := range templates {
		// Parse template key (e.g., "pages/home" → "templates/pages/home.liquid")
		templatePath := filepath.Join(templatesDir, template.Key+".liquid")

		// Create directory structure
		if err := os.MkdirAll(filepath.Dir(templatePath), 0755); err != nil {
			return err
		}

		// Write template content
		if err := os.WriteFile(templatePath, []byte(template.Content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (s *Serializer) writeJSON(path string, data interface{}) error {
	if data == nil {
		return nil
	}

	// Use YAML for JSON files (cleaner, more readable)
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(path, yamlData, 0644)
}
