package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentChangesCreate(t *testing.T) {
	tests := []struct {
		name         string
		themeID      string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantSCID     string
	}{
		{
			name:         "successful create",
			themeID:      "theme-123",
			serverStatus: http.StatusCreated,
			serverBody:   ContentChange{SCID: "cc-123", Status: "draft"},
			wantErr:      false,
			wantSCID:     "cc-123",
		},
		{
			name:         "bad request",
			themeID:      "invalid",
			serverStatus: http.StatusBadRequest,
			serverBody:   map[string]string{"error": "Invalid theme ID"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/content_changes", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				// Verify request body contains theme_id
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.themeID, body["theme_id"])

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			ccService := NewContentChanges(client)

			cc, err := ccService.Create(tt.themeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSCID, cc.SCID)
				assert.Equal(t, "draft", cc.Status)
			}
		})
	}
}

func TestContentChangesUpdate(t *testing.T) {
	templates := []ContentChangeTemplate{
		{Key: "pages/home", Content: "<h1>Home</h1>", Action: "update"},
		{Key: "pages/about", Content: "<h1>About</h1>", Action: "create"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/content_changes/cc-123", r.URL.Path)
		assert.Equal(t, http.MethodPatch, r.Method)

		// Verify request body contains templates
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Contains(t, body, "templates")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ContentChange{SCID: "cc-123", Status: "draft"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	ccService := NewContentChanges(client)

	err := ccService.Update("cc-123", "theme-123", templates, nil)
	require.NoError(t, err)
}

func TestContentChangesUpdateIncludesAssets(t *testing.T) {
	templates := []ContentChangeTemplate{
		{Key: "pages/home", Content: "<h1>Home</h1>", Action: "update"},
	}
	assets := []ContentChangeAsset{
		{Key: "images/logo.png", URL: "https://cdn.example.com/logo.png", ContentType: "image/png", ContentHash: "hash1"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/content_changes/cc-123", r.URL.Path)
		assert.Equal(t, http.MethodPatch, r.Method)

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))

		require.Contains(t, body, "assets")
		rawAssets, ok := body["assets"].([]interface{})
		require.True(t, ok)
		require.Len(t, rawAssets, 1)

		asset, ok := rawAssets[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "images/logo.png", asset["key"])
		assert.Equal(t, "https://cdn.example.com/logo.png", asset["url"])
		assert.Equal(t, "image/png", asset["content_type"])
		assert.Equal(t, "hash1", asset["content_hash"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ContentChange{SCID: "cc-123", Status: "draft"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	ccService := NewContentChanges(client)

	err := ccService.Update("cc-123", "theme-123", templates, assets)
	require.NoError(t, err)
}

func TestContentChangesList(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantStatus string
		wantLen    int
	}{
		{name: "no status filter", status: "", wantStatus: "", wantLen: 2},
		{name: "filtered by status", status: "draft", wantStatus: "draft", wantLen: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/content_changes", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, tt.wantStatus, r.URL.Query().Get("status"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": []ContentChange{
						{SCID: "cc-1", Status: "draft", Summary: "First", RecordsCount: 3, CreatedAt: "2026-06-12T00:00:00Z"},
						{SCID: "cc-2", Status: "draft", Summary: "Second", RecordsCount: 1},
					},
				})
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			ccService := NewContentChanges(client)

			changes, err := ccService.List(tt.status)
			require.NoError(t, err)
			require.Len(t, changes, tt.wantLen)
			assert.Equal(t, "cc-1", changes[0].SCID)
			assert.Equal(t, "First", changes[0].Summary)
			assert.Equal(t, 3, changes[0].RecordsCount)
		})
	}
}

func TestContentChangesGetPreviewURL(t *testing.T) {
	tests := []struct {
		name         string
		ccID         string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantURL      string
	}{
		{
			name:         "successful preview URL retrieval",
			ccID:         "cc-123",
			serverStatus: http.StatusOK,
			serverBody:   PreviewURLResponse{PreviewURL: "https://example.com/preview/cc-123"},
			wantErr:      false,
			wantURL:      "https://example.com/preview/cc-123",
		},
		{
			name:         "not found",
			ccID:         "missing",
			serverStatus: http.StatusNotFound,
			serverBody:   map[string]string{"error": "Not found"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/content_changes/"+tt.ccID+"/preview_url", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			ccService := NewContentChanges(client)

			url, err := ccService.GetPreviewURL(tt.ccID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantURL, url)
			}
		})
	}
}

func TestContentChangesPublish(t *testing.T) {
	tests := []struct {
		name         string
		ccID         string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
	}{
		{
			name:         "successful publish",
			ccID:         "cc-123",
			serverStatus: http.StatusOK,
			serverBody:   ContentChange{SCID: "cc-123", Status: "published"},
			wantErr:      false,
		},
		{
			name:         "validation error",
			ccID:         "cc-invalid",
			serverStatus: http.StatusUnprocessableEntity,
			serverBody:   map[string]string{"error": "Validation failed"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/content_changes/"+tt.ccID+"/publish", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			ccService := NewContentChanges(client)

			_, err := ccService.Publish(tt.ccID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestContentChangesGet(t *testing.T) {
	tests := []struct {
		name         string
		ccID         string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantStatus   string
	}{
		{
			name:         "successful get",
			ccID:         "cc-123",
			serverStatus: http.StatusOK,
			serverBody:   ContentChange{SCID: "cc-123", Status: "draft"},
			wantErr:      false,
			wantStatus:   "draft",
		},
		{
			name:         "not found",
			ccID:         "missing",
			serverStatus: http.StatusNotFound,
			serverBody:   map[string]string{"error": "Not found"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/content_changes/"+tt.ccID, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			ccService := NewContentChanges(client)

			cc, err := ccService.Get(tt.ccID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, cc.Status)
			}
		})
	}
}
