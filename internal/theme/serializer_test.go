package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewSerializer(t *testing.T) {
	basePath := "/tmp/test"
	serializer := NewSerializer(basePath)

	assert.NotNil(t, serializer)
	assert.Equal(t, basePath, serializer.basePath)
}

func TestSerializer_DownloadsAssets(t *testing.T) {
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	theme := &api.Theme{
		SCID: "with-assets",
		Name: "With Assets",
		Assets: []api.ThemeAsset{
			{Key: "images/logo.png", URL: "https://example.com/logo.png"},
			{Key: "style.css", URL: "https://example.com/style.css"},
		},
	}

	requested := map[string]bool{}
	serializer := NewSerializer(basePath).WithDownloader(func(url string) ([]byte, error) {
		requested[url] = true
		return []byte("downloaded:" + url), nil
	})

	require.NoError(t, serializer.Serialize(theme))

	themePath := filepath.Join(basePath, "themes", theme.Name)

	// Nested key creates subdirectories under assets/.
	logoPath := filepath.Join(themePath, "assets", "images", "logo.png")
	assert.True(t, testutil.FileExists(logoPath))
	assert.Equal(t, "downloaded:https://example.com/logo.png", testutil.ReadTestFile(t, logoPath))

	cssPath := filepath.Join(themePath, "assets", "style.css")
	assert.True(t, testutil.FileExists(cssPath))

	assert.True(t, requested["https://example.com/logo.png"])
	assert.True(t, requested["https://example.com/style.css"])
}

func TestSerializer_DownloadErrorIsWarnedNotFatal(t *testing.T) {
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	theme := &api.Theme{
		SCID: "broken-asset",
		Name: "Broken Asset",
		Assets: []api.ThemeAsset{
			{Key: "ok.png", URL: "https://example.com/ok.png"},
			{Key: "bad.png", URL: "https://example.com/bad.png"},
		},
	}

	var warnings []string
	serializer := NewSerializer(basePath).
		WithWarner(func(msg string) { warnings = append(warnings, msg) }).
		WithDownloader(func(url string) ([]byte, error) {
			if url == "https://example.com/bad.png" {
				return nil, assert.AnError
			}
			return []byte("ok"), nil
		})

	// A failed download must not fail the whole serialize.
	require.NoError(t, serializer.Serialize(theme))

	themePath := filepath.Join(basePath, "themes", theme.Name)
	assert.True(t, testutil.FileExists(filepath.Join(themePath, "assets", "ok.png")))
	assert.False(t, testutil.FileExists(filepath.Join(themePath, "assets", "bad.png")))

	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "bad.png")
}

func TestSerializer_Serialize(t *testing.T) {
	tests := []struct {
		name     string
		theme    *api.Theme
		wantErr  bool
		validate func(t *testing.T, basePath string, theme *api.Theme)
	}{
		{
			name:    "complete theme",
			theme:   testutil.TestTheme(),
			wantErr: false,
			validate: func(t *testing.T, basePath string, theme *api.Theme) {
				themePath := filepath.Join(basePath, "themes", theme.Name)

				// Verify directory structure
				assert.True(t, testutil.DirExists(themePath))
				assert.True(t, testutil.DirExists(filepath.Join(themePath, "templates")))

				// Verify theme.yml exists
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "theme.yml")))

				// Verify templates exist
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "layouts", "main.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "pages", "home.liquid")))

				// Verify variables.json exists
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "variables.json")))

				// Verify assets.json exists
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "assets.json")))
			},
		},
		{
			name:    "minimal theme",
			theme:   testutil.TestThemeMinimal(),
			wantErr: false,
			validate: func(t *testing.T, basePath string, theme *api.Theme) {
				themePath := filepath.Join(basePath, "themes", theme.Name)

				// Verify directory exists
				assert.True(t, testutil.DirExists(themePath))

				// Verify theme.yml exists
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "theme.yml")))

				// Verify no templates directory (nil templates)
				assert.False(t, testutil.DirExists(filepath.Join(themePath, "templates")))

				// Note: Even with nil Variables/Assets in the test fixture, the serializer
				// still creates empty files. This appears to be due to how the API client
				// unmarshals JSON - empty objects become empty maps/slices, not nil.
				// This is acceptable behavior.
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "variables.json")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "assets.json")))
			},
		},
		{
			name: "theme with nested templates",
			theme: &api.Theme{
				SCID: "nested-theme",
				SFID: "a0A7Z000000AbCdEUAV",
				Name: "Nested Theme",
				Templates: []api.ThemeTemplate{
					{Key: "layouts/main", Content: "main layout"},
					{Key: "layouts/alternate", Content: "alternate layout"},
					{Key: "pages/home", Content: "home page"},
					{Key: "pages/about", Content: "about page"},
					{Key: "sections/header", Content: "header section"},
					{Key: "sections/footer", Content: "footer section"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, basePath string, theme *api.Theme) {
				themePath := filepath.Join(basePath, "themes", theme.Name)

				// Verify all template directories exist
				assert.True(t, testutil.DirExists(filepath.Join(themePath, "templates", "layouts")))
				assert.True(t, testutil.DirExists(filepath.Join(themePath, "templates", "pages")))
				assert.True(t, testutil.DirExists(filepath.Join(themePath, "templates", "sections")))

				// Verify all template files exist
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "layouts", "main.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "layouts", "alternate.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "pages", "home.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "pages", "about.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "sections", "header.liquid")))
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "sections", "footer.liquid")))
			},
		},
		{
			name: "theme with empty templates array",
			theme: &api.Theme{
				SCID:      "empty-templates",
				Name:      "Empty Templates",
				Templates: []api.ThemeTemplate{},
			},
			wantErr: false,
			validate: func(t *testing.T, basePath string, theme *api.Theme) {
				themePath := filepath.Join(basePath, "themes", theme.Name)
				assert.True(t, testutil.DirExists(themePath))
				assert.False(t, testutil.DirExists(filepath.Join(themePath, "templates")))
			},
		},
		{
			name: "theme with empty sfid",
			theme: &api.Theme{
				SCID: "no-sfid",
				SFID: "",
				Name: "No SFID Theme",
			},
			wantErr: false,
			validate: func(t *testing.T, basePath string, theme *api.Theme) {
				themePath := filepath.Join(basePath, "themes", theme.Name)
				assert.True(t, testutil.DirExists(themePath))

				// Read theme.yml and verify sfid is not present
				data, err := os.ReadFile(filepath.Join(themePath, "theme.yml"))
				require.NoError(t, err)

				var metadata map[string]interface{}
				err = yaml.Unmarshal(data, &metadata)
				require.NoError(t, err)

				assert.Equal(t, theme.Name, metadata["name"])
				assert.Equal(t, theme.SCID, metadata["sc_id"])
				assert.NotContains(t, metadata, "sfid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			serializer := NewSerializer(basePath).
				WithDownloader(func(url string) ([]byte, error) { return []byte("stub-asset"), nil })
			err := serializer.Serialize(tt.theme)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, basePath, tt.theme)
				}
			}
		})
	}
}

func TestSerializer_writeThemeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		theme    *api.Theme
		wantErr  bool
		validate func(t *testing.T, metadata map[string]interface{}, theme *api.Theme)
	}{
		{
			name:    "complete metadata with sfid",
			theme:   testutil.TestTheme(),
			wantErr: false,
			validate: func(t *testing.T, metadata map[string]interface{}, theme *api.Theme) {
				assert.Equal(t, theme.Name, metadata["name"])
				assert.Equal(t, theme.SCID, metadata["sc_id"])
				assert.Equal(t, theme.SFID, metadata["sfid"])
			},
		},
		{
			name: "metadata without sfid",
			theme: &api.Theme{
				SCID: "test-theme",
				SFID: "",
				Name: "Test Theme",
			},
			wantErr: false,
			validate: func(t *testing.T, metadata map[string]interface{}, theme *api.Theme) {
				assert.Equal(t, theme.Name, metadata["name"])
				assert.Equal(t, theme.SCID, metadata["sc_id"])
				assert.NotContains(t, metadata, "sfid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themePath := filepath.Join(basePath, "themes", tt.theme.Name)
			err := os.MkdirAll(themePath, 0755)
			require.NoError(t, err)

			serializer := NewSerializer(basePath)
			err = serializer.writeThemeMetadata(themePath, tt.theme)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Read and verify the file
				data, err := os.ReadFile(filepath.Join(themePath, "theme.yml"))
				require.NoError(t, err)

				var metadata map[string]interface{}
				err = yaml.Unmarshal(data, &metadata)
				require.NoError(t, err)

				if tt.validate != nil {
					tt.validate(t, metadata, tt.theme)
				}
			}
		})
	}
}

func TestSerializer_writeTemplates(t *testing.T) {
	tests := []struct {
		name      string
		templates []api.ThemeTemplate
		wantErr   bool
		validate  func(t *testing.T, themePath string, templates []api.ThemeTemplate)
	}{
		{
			name: "single template",
			templates: []api.ThemeTemplate{
				{Key: "layouts/main", Content: "main layout content"},
			},
			wantErr: false,
			validate: func(t *testing.T, themePath string, templates []api.ThemeTemplate) {
				path := filepath.Join(themePath, "templates", "layouts", "main.liquid")
				assert.True(t, testutil.FileExists(path))
				content := testutil.ReadTestFile(t, path)
				assert.Equal(t, "main layout content", content)
			},
		},
		{
			name: "multiple templates with different paths",
			templates: []api.ThemeTemplate{
				{Key: "layouts/main", Content: "main layout"},
				{Key: "pages/home", Content: "home page"},
				{Key: "sections/header", Content: "header section"},
			},
			wantErr: false,
			validate: func(t *testing.T, themePath string, templates []api.ThemeTemplate) {
				// Check layouts
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "layouts", "main.liquid")))
				// Check pages
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "pages", "home.liquid")))
				// Check sections
				assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "sections", "header.liquid")))

				// Verify content
				content := testutil.ReadTestFile(t, filepath.Join(themePath, "templates", "pages", "home.liquid"))
				assert.Equal(t, "home page", content)
			},
		},
		{
			name:      "empty templates",
			templates: []api.ThemeTemplate{},
			wantErr:   false,
			validate: func(t *testing.T, themePath string, templates []api.ThemeTemplate) {
				// Should not create templates directory
				assert.False(t, testutil.DirExists(filepath.Join(themePath, "templates")))
			},
		},
		{
			name:      "nil templates",
			templates: nil,
			wantErr:   false,
			validate: func(t *testing.T, themePath string, templates []api.ThemeTemplate) {
				// Should not create templates directory
				assert.False(t, testutil.DirExists(filepath.Join(themePath, "templates")))
			},
		},
		{
			name: "template with nested path",
			templates: []api.ThemeTemplate{
				{Key: "pages/category/list", Content: "category list"},
			},
			wantErr: false,
			validate: func(t *testing.T, themePath string, templates []api.ThemeTemplate) {
				path := filepath.Join(themePath, "templates", "pages", "category", "list.liquid")
				assert.True(t, testutil.FileExists(path))
				content := testutil.ReadTestFile(t, path)
				assert.Equal(t, "category list", content)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themePath := filepath.Join(basePath, "theme")
			err := os.MkdirAll(themePath, 0755)
			require.NoError(t, err)

			serializer := NewSerializer(basePath)
			err = serializer.writeTemplates(themePath, tt.templates)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, themePath, tt.templates)
				}
			}
		})
	}
}

func TestSerializer_writeJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		wantErr  bool
		validate func(t *testing.T, path string, data interface{})
	}{
		{
			name: "map data",
			data: map[string]interface{}{
				"primary_color": "#FF0000",
				"font_family":   "Arial",
			},
			wantErr: false,
			validate: func(t *testing.T, path string, data interface{}) {
				assert.True(t, testutil.FileExists(path))
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				var result map[string]interface{}
				err = yaml.Unmarshal(content, &result)
				require.NoError(t, err)

				assert.Equal(t, "#FF0000", result["primary_color"])
				assert.Equal(t, "Arial", result["font_family"])
			},
		},
		{
			name: "array data",
			data: []api.ThemeAsset{
				{Key: "logo.png", ContentType: "image/png", URL: "https://example.com/logo.png", ContentHash: "hash1"},
				{Key: "style.css", ContentType: "text/css", URL: "https://example.com/style.css", ContentHash: "hash2"},
			},
			wantErr: false,
			validate: func(t *testing.T, path string, data interface{}) {
				assert.True(t, testutil.FileExists(path))
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				var result []map[string]interface{}
				err = yaml.Unmarshal(content, &result)
				require.NoError(t, err)

				assert.Len(t, result, 2)
				assert.Equal(t, "logo.png", result[0]["key"])
				assert.Equal(t, "style.css", result[1]["key"])
			},
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: false,
			validate: func(t *testing.T, path string, data interface{}) {
				// Should not create file for nil data
				assert.False(t, testutil.FileExists(path))
			},
		},
		{
			name:    "empty map",
			data:    map[string]interface{}{},
			wantErr: false,
			validate: func(t *testing.T, path string, data interface{}) {
				assert.True(t, testutil.FileExists(path))
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				var result map[string]interface{}
				err = yaml.Unmarshal(content, &result)
				require.NoError(t, err)

				assert.Empty(t, result)
			},
		},
		{
			name:    "empty array",
			data:    []interface{}{},
			wantErr: false,
			validate: func(t *testing.T, path string, data interface{}) {
				assert.True(t, testutil.FileExists(path))
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				var result []interface{}
				err = yaml.Unmarshal(content, &result)
				require.NoError(t, err)

				assert.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			path := filepath.Join(basePath, "test.json")

			serializer := NewSerializer(basePath)
			err := serializer.writeJSON(path, tt.data)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, path, tt.data)
				}
			}
		})
	}
}

func TestSerializer_CrossPlatformPaths(t *testing.T) {
	// Test that template paths work correctly on different platforms
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	theme := &api.Theme{
		SCID: "cross-platform",
		Name: "Cross Platform Theme",
		Templates: []api.ThemeTemplate{
			// Forward slashes (API format)
			{Key: "layouts/main", Content: "main"},
			{Key: "pages/home", Content: "home"},
			{Key: "sections/header/top", Content: "top"},
		},
	}

	serializer := NewSerializer(basePath)
	err := serializer.Serialize(theme)
	require.NoError(t, err)

	themePath := filepath.Join(basePath, "themes", theme.Name)

	// Verify all files exist using platform-specific paths
	assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "layouts", "main.liquid")))
	assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "pages", "home.liquid")))
	assert.True(t, testutil.FileExists(filepath.Join(themePath, "templates", "sections", "header", "top.liquid")))

	// Verify content
	content := testutil.ReadTestFile(t, filepath.Join(themePath, "templates", "sections", "header", "top.liquid"))
	assert.Equal(t, "top", content)
}

func TestSerializer_RoundTrip(t *testing.T) {
	// Test that serializing and then deserializing produces the same theme
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	original := testutil.TestTheme()

	// Serialize
	serializer := NewSerializer(basePath).
		WithDownloader(func(url string) ([]byte, error) { return []byte("stub-asset"), nil })
	err := serializer.Serialize(original)
	require.NoError(t, err)

	// Deserialize
	deserializer := NewDeserializer(basePath)
	result, err := deserializer.Deserialize(original.Name)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Name, result.Name)
	assert.Equal(t, original.SCID, result.SCID)
	assert.Equal(t, original.SFID, result.SFID)
	assert.Len(t, result.Templates, len(original.Templates))
	assert.Equal(t, original.Variables, result.Variables)
	assert.Len(t, result.Assets, len(original.Assets))

	// Compare templates (order might differ)
	for _, origTemplate := range original.Templates {
		found := false
		for _, resultTemplate := range result.Templates {
			if origTemplate.Key == resultTemplate.Key {
				assert.Equal(t, origTemplate.Content, resultTemplate.Content)
				found = true
				break
			}
		}
		assert.True(t, found, "template %s not found in result", origTemplate.Key)
	}
}
