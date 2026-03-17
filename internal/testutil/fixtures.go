package testutil

import "github.com/GetStoreConnect/storeconnect-cli/internal/api"

// TestTheme returns a sample theme for testing
func TestTheme() *api.Theme {
	return &api.Theme{
		SCID: "test-theme",
		SFID: "a0A7Z000000AbCdEUAV",
		Name: "Test Theme",
		Templates: []api.ThemeTemplate{
			{
				Key:     "layouts/main",
				Content: `<!DOCTYPE html><html><body>{{ content_for_layout }}</body></html>`,
			},
			{
				Key:     "pages/home",
				Content: `<h1>Welcome</h1>`,
			},
		},
		Assets: []api.ThemeAsset{
			{
				Filename:    "logo.png",
				ContentType: "image/png",
				URL:         "https://example.com/logo.png",
			},
		},
		Variables: map[string]interface{}{
			"primary_color": "#FF0000",
		},
	}
}

// TestThemeMinimal returns a minimal theme for testing
// Note: Variables and Assets are explicitly nil (not just omitted)
func TestThemeMinimal() *api.Theme {
	return &api.Theme{
		SCID:      "minimal-theme",
		SFID:      "a0A7Z000000AbCdEUAV",
		Name:      "Minimal Theme",
		Templates: nil,
		Variables: nil,
		Assets:    nil,
	}
}

// TestProduct returns a sample product for testing
func TestProduct() *api.Product {
	return &api.Product{
		SCID:        "test-product",
		SFID:        "a0B7Z000000XyZ1UAK",
		Name:        "Test Product",
		SKU:         "TEST-SKU-001",
		Description: "A test product",
		Price:       29.99,
	}
}

// TestArticle returns a sample article for testing
func TestArticle() *api.Article {
	return &api.Article{
		SCID:    "test-article",
		SFID:    "a0C7Z000000XyZ2UAK",
		Title:   "Test Article",
		Content: "This is test article content",
		Slug:    "test-article",
	}
}

// TestContentBlock returns a sample content block for testing
func TestContentBlock() *api.ContentBlock {
	return &api.ContentBlock{
		SCID:    "test-block",
		SFID:    "a0D7Z000000XyZ3UAK",
		Name:    "Test Block",
		Content: "This is test block content",
		Type:    "html",
	}
}
