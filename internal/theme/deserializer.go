package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"gopkg.in/yaml.v3"
)

// Deserializer converts local file structure to API theme JSON
type Deserializer struct {
	basePath string
}

// NewDeserializer creates a new theme deserializer
func NewDeserializer(basePath string) *Deserializer {
	return &Deserializer{basePath: basePath}
}

// Deserialize reads a theme from the local filesystem
func (d *Deserializer) Deserialize(themeName string) (*api.Theme, error) {
	themePath := filepath.Join(d.basePath, "themes", themeName)

	// Check if theme exists
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("theme '%s' not found", themeName)
	}

	theme := &api.Theme{
		Name: themeName,
	}

	// Read theme metadata
	if err := d.readThemeMetadata(themePath, theme); err != nil {
		return nil, fmt.Errorf("failed to read theme metadata: %w", err)
	}

	// Read templates
	templates, err := d.readTemplates(themePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates: %w", err)
	}
	theme.Templates = templates

	// Read variables
	variables, err := d.readYAML(filepath.Join(themePath, "variables.json"))
	if err == nil {
		if varMap, ok := variables.(map[string]interface{}); ok {
			theme.Variables = varMap
		}
	}

	// Read assets
	assets, err := d.readYAML(filepath.Join(themePath, "assets.json"))
	if err == nil {
		if assetsList, ok := assets.([]interface{}); ok {
			theme.Assets = make([]api.ThemeAsset, 0, len(assetsList))
			for _, a := range assetsList {
				if assetMap, ok := a.(map[string]interface{}); ok {
					asset := api.ThemeAsset{}
					if filename, ok := assetMap["filename"].(string); ok {
						asset.Filename = filename
					}
					if contentType, ok := assetMap["content_type"].(string); ok {
						asset.ContentType = contentType
					}
					if url, ok := assetMap["url"].(string); ok {
						asset.URL = url
					}
					theme.Assets = append(theme.Assets, asset)
				}
			}
		}
	}

	return theme, nil
}

func (d *Deserializer) readThemeMetadata(themePath string, theme *api.Theme) error {
	metadataPath := filepath.Join(themePath, "theme.yml")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return err
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return err
	}

	if scID, ok := metadata["sc_id"].(string); ok {
		theme.SCID = scID
	}
	if sfid, ok := metadata["sfid"].(string); ok {
		theme.SFID = sfid
	}

	return nil
}

func (d *Deserializer) readTemplates(themePath string) ([]api.ThemeTemplate, error) {
	templatesDir := filepath.Join(themePath, "templates")

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return []api.ThemeTemplate{}, nil
	}

	var templates []api.ThemeTemplate

	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-liquid files
		if info.IsDir() || !strings.HasSuffix(path, ".liquid") {
			return nil
		}

		// Read template content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Convert file path to template key
		// e.g., "templates/pages/home.liquid" → "pages/home"
		relPath, err := filepath.Rel(templatesDir, path)
		if err != nil {
			return err
		}
		key := strings.TrimSuffix(relPath, ".liquid")
		key = filepath.ToSlash(key) // Convert to forward slashes

		templates = append(templates, api.ThemeTemplate{
			Key:     key,
			Content: string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
}

func (d *Deserializer) readYAML(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}
