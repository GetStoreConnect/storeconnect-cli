package theme

import (
	"path/filepath"
	"testing"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewDeserializer(t *testing.T) {
	basePath := "/tmp/test"
	deserializer := NewDeserializer(basePath)

	assert.NotNil(t, deserializer)
	assert.Equal(t, basePath, deserializer.basePath)
}

func TestDeserializer_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		themeName string
		setup     func(t *testing.T, basePath string) string
		wantErr   bool
		validate  func(t *testing.T, theme *api.Theme)
	}{
		{
			name:      "complete theme",
			themeName: "Test Theme",
			setup: func(t *testing.T, basePath string) string {
				// Serialize a complete theme first
				theme := testutil.TestTheme()
				serializer := NewSerializer(basePath)
				err := serializer.Serialize(theme)
				require.NoError(t, err)
				return theme.Name
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "Test Theme", theme.Name)
				assert.Equal(t, "test-theme", theme.SCID)
				assert.Equal(t, "a0A7Z000000AbCdEUAV", theme.SFID)
				assert.Len(t, theme.Templates, 2)
				assert.NotNil(t, theme.Variables)
				assert.Len(t, theme.Assets, 1)
			},
		},
		{
			name:      "minimal theme",
			themeName: "Minimal Theme",
			setup: func(t *testing.T, basePath string) string {
				theme := testutil.TestThemeMinimal()
				serializer := NewSerializer(basePath)
				err := serializer.Serialize(theme)
				require.NoError(t, err)
				return theme.Name
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "Minimal Theme", theme.Name)
				assert.Equal(t, "minimal-theme", theme.SCID)
				assert.Equal(t, "a0A7Z000000AbCdEUAV", theme.SFID)
				assert.Empty(t, theme.Templates)
				// Note: Even though the original theme had nil Variables/Assets,
				// the serializer creates empty files ({} and []) due to how Go
				// handles nil maps/slices through interface{}. When deserializing,
				// we get empty maps/slices back.
				assert.NotNil(t, theme.Variables)
				assert.Empty(t, theme.Variables)
				assert.NotNil(t, theme.Assets)
				assert.Empty(t, theme.Assets)
			},
		},
		{
			name:      "theme with nested templates",
			themeName: "Nested",
			setup: func(t *testing.T, basePath string) string {
				themePath := filepath.Join(basePath, "themes", "Nested")
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "layouts"))
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "pages"))
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "sections"))

				// Write theme.yml
				metadata := map[string]interface{}{
					"name":  "Nested",
					"sc_id": "nested-theme",
					"sfid":  "a0A7Z000000AbCdEUAV",
				}
				data, _ := yaml.Marshal(metadata)
				testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

				// Write templates
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "layouts", "main.liquid"), "main layout")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "pages", "home.liquid"), "home page")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "sections", "header.liquid"), "header section")

				return "Nested"
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "Nested", theme.Name)
				assert.Len(t, theme.Templates, 3)

				// Check template keys are correct
				keys := make(map[string]bool)
				for _, tmpl := range theme.Templates {
					keys[tmpl.Key] = true
				}
				assert.True(t, keys["layouts/main"])
				assert.True(t, keys["pages/home"])
				assert.True(t, keys["sections/header"])
			},
		},
		{
			name:      "missing theme",
			themeName: "NonExistent",
			setup: func(t *testing.T, basePath string) string {
				// Don't create anything
				return "NonExistent"
			},
			wantErr: true,
		},
		{
			name:      "missing templates directory",
			themeName: "NoTemplates",
			setup: func(t *testing.T, basePath string) string {
				themePath := filepath.Join(basePath, "themes", "NoTemplates")
				testutil.CreateTestDir(t, themePath)

				// Write theme.yml
				metadata := map[string]interface{}{
					"name":  "NoTemplates",
					"sc_id": "no-templates",
				}
				data, _ := yaml.Marshal(metadata)
				testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

				return "NoTemplates"
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "NoTemplates", theme.Name)
				assert.Empty(t, theme.Templates)
			},
		},
		{
			name:      "missing optional files",
			themeName: "Minimal",
			setup: func(t *testing.T, basePath string) string {
				themePath := filepath.Join(basePath, "themes", "Minimal")
				testutil.CreateTestDir(t, themePath)

				// Only write theme.yml
				metadata := map[string]interface{}{
					"name":  "Minimal",
					"sc_id": "minimal",
				}
				data, _ := yaml.Marshal(metadata)
				testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

				return "Minimal"
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "Minimal", theme.Name)
				assert.Equal(t, "minimal", theme.SCID)
				assert.Empty(t, theme.SFID)
				assert.Empty(t, theme.Templates)
				assert.Nil(t, theme.Variables)
				assert.Nil(t, theme.Assets)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themeName := tt.setup(t, basePath)

			deserializer := NewDeserializer(basePath)
			theme, err := deserializer.Deserialize(themeName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, theme)
			} else {
				require.NoError(t, err)
				require.NotNil(t, theme)
				if tt.validate != nil {
					tt.validate(t, theme)
				}
			}
		})
	}
}

func TestDeserializer_readThemeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantErr  bool
		validate func(t *testing.T, theme *api.Theme)
	}{
		{
			name: "complete metadata",
			metadata: map[string]interface{}{
				"name":  "Test Theme",
				"sc_id": "test-theme",
				"sfid":  "a0A7Z000000AbCdEUAV",
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "test-theme", theme.SCID)
				assert.Equal(t, "a0A7Z000000AbCdEUAV", theme.SFID)
			},
		},
		{
			name: "metadata without sfid",
			metadata: map[string]interface{}{
				"name":  "Test Theme",
				"sc_id": "test-theme",
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "test-theme", theme.SCID)
				assert.Empty(t, theme.SFID)
			},
		},
		{
			name: "metadata with extra fields",
			metadata: map[string]interface{}{
				"name":        "Test Theme",
				"sc_id":       "test-theme",
				"sfid":        "a0A7Z000000AbCdEUAV",
				"extra_field": "ignored",
			},
			wantErr: false,
			validate: func(t *testing.T, theme *api.Theme) {
				assert.Equal(t, "test-theme", theme.SCID)
				assert.Equal(t, "a0A7Z000000AbCdEUAV", theme.SFID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themePath := filepath.Join(basePath, "theme")
			testutil.CreateTestDir(t, themePath)

			// Write theme.yml
			data, err := yaml.Marshal(tt.metadata)
			require.NoError(t, err)
			testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

			deserializer := NewDeserializer(basePath)
			theme := &api.Theme{Name: "Test Theme"}
			err = deserializer.readThemeMetadata(themePath, theme)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, theme)
				}
			}
		})
	}
}

func TestDeserializer_readThemeMetadata_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "missing file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid yaml",
			content: "invalid: yaml: content: [unclosed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themePath := filepath.Join(basePath, "theme")
			testutil.CreateTestDir(t, themePath)

			if tt.content != "" {
				testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), tt.content)
			}

			deserializer := NewDeserializer(basePath)
			theme := &api.Theme{Name: "Test Theme"}
			err := deserializer.readThemeMetadata(themePath, theme)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeserializer_readTemplates(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, themePath string)
		wantErr  bool
		validate func(t *testing.T, templates []api.ThemeTemplate)
	}{
		{
			name: "single template",
			setup: func(t *testing.T, themePath string) {
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "layouts"))
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "layouts", "main.liquid"), "main layout")
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Len(t, templates, 1)
				assert.Equal(t, "layouts/main", templates[0].Key)
				assert.Equal(t, "main layout", templates[0].Content)
			},
		},
		{
			name: "multiple templates",
			setup: func(t *testing.T, themePath string) {
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "layouts"))
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "pages"))

				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "layouts", "main.liquid"), "main")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "pages", "home.liquid"), "home")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "pages", "about.liquid"), "about")
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Len(t, templates, 3)

				keys := make(map[string]string)
				for _, tmpl := range templates {
					keys[tmpl.Key] = tmpl.Content
				}

				assert.Equal(t, "main", keys["layouts/main"])
				assert.Equal(t, "home", keys["pages/home"])
				assert.Equal(t, "about", keys["pages/about"])
			},
		},
		{
			name: "nested template paths",
			setup: func(t *testing.T, themePath string) {
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "pages", "category"))
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "pages", "category", "list.liquid"), "category list")
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Len(t, templates, 1)
				assert.Equal(t, "pages/category/list", templates[0].Key)
				assert.Equal(t, "category list", templates[0].Content)
			},
		},
		{
			name: "ignores non-liquid files",
			setup: func(t *testing.T, themePath string) {
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates"))
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "main.liquid"), "liquid file")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "README.md"), "readme")
				testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "config.json"), "{}")
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Len(t, templates, 1)
				assert.Equal(t, "main", templates[0].Key)
			},
		},
		{
			name: "missing templates directory",
			setup: func(t *testing.T, themePath string) {
				// Don't create templates directory
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Empty(t, templates)
			},
		},
		{
			name: "empty templates directory",
			setup: func(t *testing.T, themePath string) {
				testutil.CreateTestDir(t, filepath.Join(themePath, "templates"))
			},
			wantErr: false,
			validate: func(t *testing.T, templates []api.ThemeTemplate) {
				assert.Empty(t, templates)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			themePath := filepath.Join(basePath, "theme")
			testutil.CreateTestDir(t, themePath)

			tt.setup(t, themePath)

			deserializer := NewDeserializer(basePath)
			templates, err := deserializer.readTemplates(themePath)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, templates)
				}
			}
		})
	}
}

func TestDeserializer_readTemplates_KeyGeneration(t *testing.T) {
	// Test that template keys are generated correctly across platforms
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themePath := filepath.Join(basePath, "theme")
	templatesPath := filepath.Join(themePath, "templates")

	testutil.CreateTestDir(t, filepath.Join(templatesPath, "layouts"))
	testutil.CreateTestDir(t, filepath.Join(templatesPath, "pages", "deep", "nested"))

	testutil.WriteTestFile(t, filepath.Join(templatesPath, "layouts", "main.liquid"), "main")
	testutil.WriteTestFile(t, filepath.Join(templatesPath, "pages", "deep", "nested", "page.liquid"), "nested")

	deserializer := NewDeserializer(basePath)
	templates, err := deserializer.readTemplates(themePath)
	require.NoError(t, err)

	assert.Len(t, templates, 2)

	// Keys should always use forward slashes regardless of OS
	keys := make(map[string]bool)
	for _, tmpl := range templates {
		keys[tmpl.Key] = true
		// Verify no backslashes in keys (Windows compatibility)
		assert.NotContains(t, tmpl.Key, "\\")
	}

	assert.True(t, keys["layouts/main"])
	assert.True(t, keys["pages/deep/nested/page"])
}

func TestDeserializer_readYAML(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantErr  bool
		validate func(t *testing.T, result interface{})
	}{
		{
			name: "map data",
			content: `primary_color: "#FF0000"
font_family: Arial`,
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "#FF0000", m["primary_color"])
				assert.Equal(t, "Arial", m["font_family"])
			},
		},
		{
			name: "array data",
			content: `- filename: logo.png
  content_type: image/png
  url: https://example.com/logo.png
- filename: style.css
  content_type: text/css
  url: https://example.com/style.css`,
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				arr, ok := result.([]interface{})
				require.True(t, ok)
				assert.Len(t, arr, 2)

				first, ok := arr[0].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "logo.png", first["filename"])
			},
		},
		{
			name:    "invalid yaml",
			content: "invalid: yaml: [unclosed",
			wantErr: true,
		},
		{
			name:    "empty file",
			content: "",
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			path := filepath.Join(basePath, "test.yml")
			testutil.WriteTestFile(t, path, tt.content)

			deserializer := NewDeserializer(basePath)
			result, err := deserializer.readYAML(path)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestDeserializer_readYAML_MissingFile(t *testing.T) {
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	path := filepath.Join(basePath, "missing.yml")

	deserializer := NewDeserializer(basePath)
	_, err := deserializer.readYAML(path)

	require.Error(t, err)
}

func TestDeserializer_VariablesTypeAssertion(t *testing.T) {
	// Test that variables are correctly parsed as map[string]interface{}
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themePath := filepath.Join(basePath, "themes", "Test")
	testutil.CreateTestDir(t, themePath)

	// Write theme.yml
	metadata := map[string]interface{}{
		"name":  "Test",
		"sc_id": "test",
	}
	data, _ := yaml.Marshal(metadata)
	testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

	// Write variables.json with various types
	variables := `primary_color: "#FF0000"
font_size: 16
show_header: true
nested:
  key: value`
	testutil.WriteTestFile(t, filepath.Join(themePath, "variables.json"), variables)

	deserializer := NewDeserializer(basePath)
	theme, err := deserializer.Deserialize("Test")
	require.NoError(t, err)

	assert.NotNil(t, theme.Variables)
	assert.Equal(t, "#FF0000", theme.Variables["primary_color"])
	assert.Equal(t, 16, theme.Variables["font_size"])
	assert.Equal(t, true, theme.Variables["show_header"])

	nested, ok := theme.Variables["nested"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", nested["key"])
}

func TestDeserializer_AssetsTypeAssertion(t *testing.T) {
	// Test that assets are correctly parsed as []api.ThemeAsset
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themePath := filepath.Join(basePath, "themes", "Test")
	testutil.CreateTestDir(t, themePath)

	// Write theme.yml
	metadata := map[string]interface{}{
		"name":  "Test",
		"sc_id": "test",
	}
	data, _ := yaml.Marshal(metadata)
	testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

	// Write assets.json
	assets := `- filename: logo.png
  content_type: image/png
  url: https://example.com/logo.png
- filename: style.css
  content_type: text/css
  url: https://example.com/style.css`
	testutil.WriteTestFile(t, filepath.Join(themePath, "assets.json"), assets)

	deserializer := NewDeserializer(basePath)
	theme, err := deserializer.Deserialize("Test")
	require.NoError(t, err)

	require.Len(t, theme.Assets, 2)
	assert.Equal(t, "logo.png", theme.Assets[0].Filename)
	assert.Equal(t, "image/png", theme.Assets[0].ContentType)
	assert.Equal(t, "https://example.com/logo.png", theme.Assets[0].URL)
	assert.Equal(t, "style.css", theme.Assets[1].Filename)
}

func TestDeserializer_InvalidAssetsFormat(t *testing.T) {
	// Test that invalid assets format doesn't crash, just returns nil assets
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themePath := filepath.Join(basePath, "themes", "Test")
	testutil.CreateTestDir(t, themePath)

	// Write theme.yml
	metadata := map[string]interface{}{
		"name":  "Test",
		"sc_id": "test",
	}
	data, _ := yaml.Marshal(metadata)
	testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

	// Write assets.json as a map instead of array (invalid)
	assets := `invalid: format`
	testutil.WriteTestFile(t, filepath.Join(themePath, "assets.json"), assets)

	deserializer := NewDeserializer(basePath)
	theme, err := deserializer.Deserialize("Test")
	require.NoError(t, err)

	// Should handle gracefully - assets will be nil/empty
	assert.Nil(t, theme.Assets)
}

func TestDeserializer_RoundTrip(t *testing.T) {
	// Test that deserializing and then serializing produces the same files
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	// Create a theme manually
	themePath := filepath.Join(basePath, "themes", "Original")
	testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "layouts"))
	testutil.CreateTestDir(t, filepath.Join(themePath, "templates", "pages"))

	// Write theme.yml
	metadata := map[string]interface{}{
		"name":  "Original",
		"sc_id": "original",
		"sfid":  "a0A7Z000000AbCdEUAV",
	}
	data, _ := yaml.Marshal(metadata)
	testutil.WriteTestFile(t, filepath.Join(themePath, "theme.yml"), string(data))

	// Write templates
	testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "layouts", "main.liquid"), "main layout")
	testutil.WriteTestFile(t, filepath.Join(themePath, "templates", "pages", "home.liquid"), "home page")

	// Write variables.json
	variables := map[string]interface{}{
		"primary_color": "#FF0000",
	}
	varData, _ := yaml.Marshal(variables)
	testutil.WriteTestFile(t, filepath.Join(themePath, "variables.json"), string(varData))

	// Deserialize
	deserializer := NewDeserializer(basePath)
	theme, err := deserializer.Deserialize("Original")
	require.NoError(t, err)

	// Change the name to avoid overwriting
	theme.Name = "Copy"

	// Serialize
	serializer := NewSerializer(basePath)
	err = serializer.Serialize(theme)
	require.NoError(t, err)

	// Deserialize the copy
	themeCopy, err := deserializer.Deserialize("Copy")
	require.NoError(t, err)

	// Compare
	assert.Equal(t, "Copy", themeCopy.Name)
	assert.Equal(t, theme.SCID, themeCopy.SCID)
	assert.Equal(t, theme.SFID, themeCopy.SFID)
	assert.Len(t, themeCopy.Templates, len(theme.Templates))
	assert.Equal(t, theme.Variables, themeCopy.Variables)
}
