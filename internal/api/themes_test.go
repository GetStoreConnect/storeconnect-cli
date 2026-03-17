package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThemesList(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantCount    int
	}{
		{
			name:         "successful list",
			serverStatus: http.StatusOK,
			serverBody: map[string]interface{}{
				"themes": []Theme{
					{SCID: "theme-1", Name: "Theme 1"},
					{SCID: "theme-2", Name: "Theme 2"},
				},
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:         "empty list",
			serverStatus: http.StatusOK,
			serverBody: map[string]interface{}{
				"themes": []Theme{},
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:         "unauthorized",
			serverStatus: http.StatusUnauthorized,
			serverBody:   map[string]string{"error": "Unauthorized"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/themes", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			themesService := NewThemes(client)

			themes, err := themesService.List()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, themes, tt.wantCount)
			}
		})
	}
}

func TestThemesGet(t *testing.T) {
	tests := []struct {
		name         string
		themeID      string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantName     string
	}{
		{
			name:         "successful get",
			themeID:      "theme-123",
			serverStatus: http.StatusOK,
			serverBody:   Theme{SCID: "theme-123", Name: "Test Theme"},
			wantErr:      false,
			wantName:     "Test Theme",
		},
		{
			name:         "not found",
			themeID:      "missing",
			serverStatus: http.StatusNotFound,
			serverBody:   map[string]string{"error": "Not found"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/themes/"+tt.themeID, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			themesService := NewThemes(client)

			theme, err := themesService.Get(tt.themeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantName, theme.Name)
			}
		})
	}
}

func TestThemesGetBase(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/themes/base", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Theme{SCID: "base", Name: "Base Theme"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	themesService := NewThemes(client)

	theme, err := themesService.GetBase()

	require.NoError(t, err)
	assert.Equal(t, "Base Theme", theme.Name)
	assert.Equal(t, "base", theme.SCID)
}

func TestThemesCreate(t *testing.T) {
	tests := []struct {
		name         string
		themeName    string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantSCID     string
	}{
		{
			name:         "successful create",
			themeName:    "New Theme",
			serverStatus: http.StatusCreated,
			serverBody:   Theme{SCID: "new-theme", Name: "New Theme"},
			wantErr:      false,
			wantSCID:     "new-theme",
		},
		{
			name:         "conflict - already exists",
			themeName:    "Duplicate",
			serverStatus: http.StatusConflict,
			serverBody:   map[string]string{"error": "Already exists"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/themes", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				// Verify request body
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, tt.themeName, body["name"])

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			themesService := NewThemes(client)

			theme, err := themesService.Create(tt.themeName)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSCID, theme.SCID)
				assert.Equal(t, tt.themeName, theme.Name)
			}
		})
	}
}

func TestThemesDelete(t *testing.T) {
	tests := []struct {
		name         string
		themeID      string
		serverStatus int
		wantErr      bool
	}{
		{
			name:         "successful delete",
			themeID:      "theme-123",
			serverStatus: http.StatusNoContent,
			wantErr:      false,
		},
		{
			name:         "not found",
			themeID:      "missing",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/themes/"+tt.themeID, r.URL.Path)
				assert.Equal(t, http.MethodDelete, r.Method)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			themesService := NewThemes(client)

			err := themesService.Delete(tt.themeID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
