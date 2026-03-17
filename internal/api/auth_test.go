package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthInfo(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		serverBody   interface{}
		wantErr      bool
		wantVersion  string
	}{
		{
			name:         "successful authentication",
			serverStatus: http.StatusOK,
			serverBody: InfoResponse{
				StoreconnectVersion: "20.13.0",
				BaseThemeVersion:    "1.0.0",
				OrgID:               "00D7Z000000AbCdEFG",
				StoreSFID:           "a0A7Z000000AbCdEUAV",
			},
			wantErr:     false,
			wantVersion: "20.13.0",
		},
		{
			name:         "unauthorized - invalid credentials",
			serverStatus: http.StatusUnauthorized,
			serverBody:   map[string]string{"error": "Invalid credentials"},
			wantErr:      true,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			serverBody:   map[string]string{"error": "Server error"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v1/info", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				// Verify authorization header is present
				authHeader := r.Header.Get("Authorization")
				assert.NotEmpty(t, authHeader)
				assert.Contains(t, authHeader, "Bearer")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			authService := NewAuth(client)

			info, err := authService.Info()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantVersion, info.StoreconnectVersion)
			}
		})
	}
}
