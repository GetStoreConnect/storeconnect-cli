package theme

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"gopkg.in/yaml.v3"
)

// Downloader fetches the bytes at a URL. It is a seam so asset downloading can
// be driven by httptest (or stubbed) without reaching the network.
type Downloader func(url string) ([]byte, error)

// defaultDownloader fetches a URL over HTTP with a sane timeout.
func defaultDownloader(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Serializer converts API theme JSON to local file structure
type Serializer struct {
	basePath string

	// download fetches asset binaries; overridable for tests.
	download Downloader
	// warn reports a non-fatal problem (e.g. a failed asset download).
	warn func(string)
}

// NewSerializer creates a new theme serializer
func NewSerializer(basePath string) *Serializer {
	return &Serializer{
		basePath: basePath,
		download: defaultDownloader,
		warn:     func(msg string) { fmt.Fprintln(os.Stderr, msg) },
	}
}

// WithDownloader overrides how asset binaries are fetched (used in tests).
func (s *Serializer) WithDownloader(d Downloader) *Serializer {
	s.download = d
	return s
}

// WithWarner overrides where non-fatal warnings are reported (used in tests).
func (s *Serializer) WithWarner(w func(string)) *Serializer {
	s.warn = w
	return s
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

	// Write assets.json metadata
	if err := s.writeJSON(filepath.Join(themePath, "assets.json"), theme.Assets); err != nil {
		return fmt.Errorf("failed to write assets: %w", err)
	}

	// Download asset binaries into assets/. A failed download is a warning,
	// not a fatal error - the rest of the theme is still usable.
	s.downloadAssets(themePath, theme.Assets)

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

// downloadAssets fetches each asset's URL into assets/<key>, creating any
// subdirectories the key implies. Download failures are reported via warn and
// skipped so a single broken asset never fails the whole pull.
func (s *Serializer) downloadAssets(themePath string, assets []api.ThemeAsset) {
	for _, asset := range assets {
		if asset.Key == "" || asset.URL == "" {
			continue
		}

		content, err := s.download(asset.URL)
		if err != nil {
			s.warn(fmt.Sprintf("warning: failed to download asset %q: %v", asset.Key, err))
			continue
		}

		// Key may contain slashes - build the nested path and ensure dirs exist.
		destPath := filepath.Join(themePath, AssetsDirName, filepath.FromSlash(asset.Key))
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			s.warn(fmt.Sprintf("warning: failed to create directory for asset %q: %v", asset.Key, err))
			continue
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			s.warn(fmt.Sprintf("warning: failed to write asset %q: %v", asset.Key, err))
			continue
		}
	}
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
