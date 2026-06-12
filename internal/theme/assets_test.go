package theme

import (
	"path/filepath"
	"testing"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
	"github.com/GetStoreConnect/storeconnect-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectChangedAssets(t *testing.T) {
	tests := []struct {
		name     string
		local    []LocalAsset
		server   []api.ThemeAsset
		wantKeys []string
	}{
		{
			name: "new asset absent on server is selected",
			local: []LocalAsset{
				{Key: "images/logo.png", ContentHash: "hashA"},
			},
			server:   nil,
			wantKeys: []string{"images/logo.png"},
		},
		{
			name: "changed hash is selected",
			local: []LocalAsset{
				{Key: "images/logo.png", ContentHash: "hashNEW"},
			},
			server: []api.ThemeAsset{
				{Key: "images/logo.png", ContentHash: "hashOLD"},
			},
			wantKeys: []string{"images/logo.png"},
		},
		{
			name: "unchanged hash is skipped",
			local: []LocalAsset{
				{Key: "images/logo.png", ContentHash: "same"},
			},
			server: []api.ThemeAsset{
				{Key: "images/logo.png", ContentHash: "same"},
			},
			wantKeys: nil,
		},
		{
			name: "mixed selects only new and changed",
			local: []LocalAsset{
				{Key: "a.png", ContentHash: "1"},     // unchanged
				{Key: "b.png", ContentHash: "2-new"}, // changed
				{Key: "c.png", ContentHash: "3"},     // new
			},
			server: []api.ThemeAsset{
				{Key: "a.png", ContentHash: "1"},
				{Key: "b.png", ContentHash: "2-old"},
			},
			wantKeys: []string{"b.png", "c.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changed := SelectChangedAssets(tt.local, tt.server)

			var gotKeys []string
			for _, a := range changed {
				gotKeys = append(gotKeys, a.Key)
			}
			assert.Equal(t, tt.wantKeys, gotKeys)
		})
	}
}

func TestReadLocalAssets(t *testing.T) {
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themeDir := filepath.Join(basePath, "themes", "Test")
	assetsDir := filepath.Join(themeDir, AssetsDirName)
	testutil.CreateTestDir(t, filepath.Join(assetsDir, "images"))

	testutil.WriteTestFile(t, filepath.Join(assetsDir, "style.css"), "body{}")
	testutil.WriteTestFile(t, filepath.Join(assetsDir, "images", "logo.png"), "PNGDATA")

	assets, err := ReadLocalAssets(themeDir)
	require.NoError(t, err)
	require.Len(t, assets, 2)

	byKey := map[string]LocalAsset{}
	for _, a := range assets {
		byKey[a.Key] = a
	}

	// Nested key uses forward slashes regardless of platform.
	logo, ok := byKey["images/logo.png"]
	require.True(t, ok)
	assert.Equal(t, "image/png", logo.ContentType)
	assert.Equal(t, HashContent([]byte("PNGDATA")), logo.ContentHash)
	assert.Equal(t, []byte("PNGDATA"), logo.Content)

	css, ok := byKey["style.css"]
	require.True(t, ok)
	assert.Equal(t, "text/css", css.ContentType)
}

func TestReadLocalAssetsMissingDir(t *testing.T) {
	basePath, cleanup := testutil.CreateTempProject(t)
	defer cleanup()

	themeDir := filepath.Join(basePath, "themes", "NoAssets")
	testutil.CreateTestDir(t, themeDir)

	assets, err := ReadLocalAssets(themeDir)
	require.NoError(t, err)
	assert.Empty(t, assets)
}

func TestInferContentType(t *testing.T) {
	cases := map[string]string{
		"logo.png":   "image/png",
		"photo.JPG":  "image/jpeg",
		"icon.svg":   "image/svg+xml",
		"style.css":  "text/css",
		"app.js":     "application/javascript",
		"data.bin":   "application/octet-stream",
		"noext":      "application/octet-stream",
		"manual.pdf": "application/pdf",
	}
	for name, want := range cases {
		assert.Equal(t, want, InferContentType(name), name)
	}
}

func TestInferFileType(t *testing.T) {
	cases := map[string]string{
		"logo.png":   "image",
		"photo.jpeg": "image",
		"icon.SVG":   "image",
		"style.css":  "document",
		"manual.pdf": "document",
		"noext":      "document",
	}
	for name, want := range cases {
		assert.Equal(t, want, InferFileType(name), name)
	}
}
