package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		storeID  string
		apiKey   string
		opts     []ClientOption
		wantAuth string
	}{
		{
			name:     "basic client without org ID",
			baseURL:  "https://example.com",
			storeID:  "a0A7Z000000AbCdEUAV",
			apiKey:   "test-key",
			opts:     nil,
			wantAuth: "Bearer a0A7Z000000AbCdEUAV:test-key",
		},
		{
			name:     "client with org ID",
			baseURL:  "https://example.com",
			storeID:  "a0A7Z000000AbCdEUAV",
			apiKey:   "test-key",
			opts:     []ClientOption{WithOrgID("00D7Z000000AbCdEFG")},
			wantAuth: "Bearer 00D7Z000000AbCdEFG:a0A7Z000000AbCdEUAV:test-key",
		},
		{
			name:     "client with change set ID",
			baseURL:  "https://example.com",
			storeID:  "a0A7Z000000AbCdEUAV",
			apiKey:   "test-key",
			opts:     []ClientOption{WithChangeSetID("changeset-123")},
			wantAuth: "Bearer a0A7Z000000AbCdEUAV:test-key",
		},
		{
			name:    "client with org ID and change set ID",
			baseURL: "https://example.com",
			storeID: "a0A7Z000000AbCdEUAV",
			apiKey:  "test-key",
			opts: []ClientOption{
				WithOrgID("00D7Z000000AbCdEFG"),
				WithChangeSetID("changeset-123"),
			},
			wantAuth: "Bearer 00D7Z000000AbCdEFG:a0A7Z000000AbCdEUAV:test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.baseURL, tt.storeID, tt.apiKey, tt.opts...)

			assert.Equal(t, tt.baseURL, client.BaseURL)
			assert.Equal(t, tt.storeID, client.StoreSFID)
			assert.Equal(t, tt.apiKey, client.APIKey)
			assert.NotNil(t, client.httpClient)

			// Check authorization header
			headers := client.httpClient.Header
			assert.Equal(t, tt.wantAuth, headers.Get("Authorization"))
		})
	}
}

func TestBuildBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		orgID    string
		storeID  string
		apiKey   string
		expected string
	}{
		{
			name:     "legacy format without org ID",
			orgID:    "",
			storeID:  "a0A7Z000000AbCdEUAV",
			apiKey:   "test-key",
			expected: "Bearer a0A7Z000000AbCdEUAV:test-key",
		},
		{
			name:     "enhanced format with org ID",
			orgID:    "00D7Z000000AbCdEFG",
			storeID:  "a0A7Z000000AbCdEUAV",
			apiKey:   "test-key",
			expected: "Bearer 00D7Z000000AbCdEFG:a0A7Z000000AbCdEUAV:test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				OrgID:     tt.orgID,
				StoreSFID: tt.storeID,
				APIKey:    tt.apiKey,
			}
			got := client.buildBearerToken()
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestClientGet(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		params         map[string]string
		changeSetID    string
		serverResponse int
		serverBody     interface{}
		wantErr        bool
		wantStatusCode int
		wantResult     map[string]string
	}{
		{
			name:           "successful GET request",
			path:           "/api/v1/test",
			params:         nil,
			serverResponse: http.StatusOK,
			serverBody:     map[string]string{"status": "ok"},
			wantErr:        false,
			wantResult:     map[string]string{"status": "ok"},
		},
		{
			name:           "GET with query params",
			path:           "/api/v1/test",
			params:         map[string]string{"foo": "bar"},
			serverResponse: http.StatusOK,
			serverBody:     map[string]string{"foo": "bar"},
			wantErr:        false,
			wantResult:     map[string]string{"foo": "bar"},
		},
		{
			name:           "GET with change set header",
			path:           "/api/v1/test",
			changeSetID:    "changeset-123",
			serverResponse: http.StatusOK,
			serverBody:     map[string]string{"status": "ok"},
			wantErr:        false,
			wantResult:     map[string]string{"status": "ok"},
		},
		{
			name:           "404 not found",
			path:           "/api/v1/missing",
			serverResponse: http.StatusNotFound,
			serverBody:     map[string]string{"error": "Not found"},
			wantErr:        true,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "401 unauthorized",
			path:           "/api/v1/test",
			serverResponse: http.StatusUnauthorized,
			serverBody:     map[string]string{"error": "Unauthorized"},
			wantErr:        true,
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify method
				assert.Equal(t, http.MethodGet, r.Method)

				// Verify path
				assert.Equal(t, tt.path, r.URL.Path)

				// Verify query params
				if tt.params != nil {
					for k, v := range tt.params {
						assert.Equal(t, v, r.URL.Query().Get(k))
					}
				}

				// Verify change set header
				if tt.changeSetID != "" {
					assert.Equal(t, tt.changeSetID, r.Header.Get("X-SC-Change-Set-ID"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverResponse)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			opts := []ClientOption{}
			if tt.changeSetID != "" {
				opts = append(opts, WithChangeSetID(tt.changeSetID))
			}

			client := NewClient(server.URL, "test-store", "test-key", opts...)

			result := make(map[string]string)
			err := client.Get(tt.path, &result, tt.params)

			if tt.wantErr {
				require.Error(t, err)
				apiErr, ok := err.(*APIError)
				require.True(t, ok)
				if tt.wantStatusCode != 0 {
					assert.Equal(t, tt.wantStatusCode, apiErr.StatusCode)
				}
			} else {
				require.NoError(t, err)
				if tt.wantResult != nil {
					assert.Equal(t, tt.wantResult, result)
				}
			}
		})
	}
}

func TestClientPost(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		body           interface{}
		changeSetID    string
		serverResponse int
		serverBody     interface{}
		wantErr        bool
		wantStatusCode int
	}{
		{
			name:           "successful POST request",
			path:           "/api/v1/test",
			body:           map[string]string{"name": "test"},
			serverResponse: http.StatusCreated,
			serverBody:     map[string]string{"id": "123", "name": "test"},
			wantErr:        false,
		},
		{
			name:           "POST with change set header",
			path:           "/api/v1/test",
			body:           map[string]string{"name": "test"},
			changeSetID:    "changeset-123",
			serverResponse: http.StatusCreated,
			serverBody:     map[string]string{"id": "123"},
			wantErr:        false,
		},
		{
			name:           "400 bad request",
			path:           "/api/v1/test",
			body:           map[string]string{"invalid": "data"},
			serverResponse: http.StatusBadRequest,
			serverBody:     map[string]string{"error": "Invalid data"},
			wantErr:        true,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "409 conflict",
			path:           "/api/v1/test",
			body:           map[string]string{"name": "duplicate"},
			serverResponse: http.StatusConflict,
			serverBody:     map[string]string{"error": "Already exists"},
			wantErr:        true,
			wantStatusCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, tt.path, r.URL.Path)

				if tt.changeSetID != "" {
					assert.Equal(t, tt.changeSetID, r.Header.Get("X-SC-Change-Set-ID"))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverResponse)
				json.NewEncoder(w).Encode(tt.serverBody)
			}))
			defer server.Close()

			opts := []ClientOption{}
			if tt.changeSetID != "" {
				opts = append(opts, WithChangeSetID(tt.changeSetID))
			}

			client := NewClient(server.URL, "test-store", "test-key", opts...)

			result := make(map[string]string)
			err := client.Post(tt.path, tt.body, &result)

			if tt.wantErr {
				require.Error(t, err)
				apiErr, ok := err.(*APIError)
				require.True(t, ok)
				if tt.wantStatusCode != 0 {
					assert.Equal(t, tt.wantStatusCode, apiErr.StatusCode)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClientPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")

	result := make(map[string]string)
	err := client.Put("/api/v1/test/123", map[string]string{"name": "updated"}, &result)

	require.NoError(t, err)
	assert.Equal(t, "updated", result["status"])
}

func TestClientPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "patched"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")

	result := make(map[string]string)
	err := client.Patch("/api/v1/test/123", map[string]string{"field": "value"}, &result)

	require.NoError(t, err)
	assert.Equal(t, "patched", result["status"])
}

func TestClientDelete(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		serverBody     interface{}
		wantErr        bool
		wantStatusCode int
	}{
		{
			name:           "successful DELETE",
			serverResponse: http.StatusNoContent,
			serverBody:     nil,
			wantErr:        false,
		},
		{
			name:           "DELETE with 200 response",
			serverResponse: http.StatusOK,
			serverBody:     map[string]string{"status": "deleted"},
			wantErr:        false,
		},
		{
			name:           "404 not found",
			serverResponse: http.StatusNotFound,
			serverBody:     map[string]string{"error": "Not found"},
			wantErr:        true,
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodDelete, r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.serverResponse)
				if tt.serverBody != nil {
					json.NewEncoder(w).Encode(tt.serverBody)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			err := client.Delete("/api/v1/test/123", nil)

			if tt.wantErr {
				require.Error(t, err)
				apiErr, ok := err.(*APIError)
				require.True(t, ok)
				if tt.wantStatusCode != 0 {
					assert.Equal(t, tt.wantStatusCode, apiErr.StatusCode)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   interface{}
		wantErr        bool
		wantStatusCode int
		wantMessage    string
	}{
		{
			name:         "200 success",
			statusCode:   http.StatusOK,
			responseBody: map[string]string{"status": "ok"},
			wantErr:      false,
		},
		{
			name:         "201 created",
			statusCode:   http.StatusCreated,
			responseBody: map[string]string{"id": "123"},
			wantErr:      false,
		},
		{
			name:       "204 no content",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:           "400 bad request",
			statusCode:     http.StatusBadRequest,
			responseBody:   map[string]string{"error": "Invalid input"},
			wantErr:        true,
			wantStatusCode: http.StatusBadRequest,
			wantMessage:    "Bad request: Invalid input",
		},
		{
			name:           "401 unauthorized",
			statusCode:     http.StatusUnauthorized,
			responseBody:   map[string]string{"error": "Unauthorized"},
			wantErr:        true,
			wantStatusCode: http.StatusUnauthorized,
			wantMessage:    "Authentication failed. Please check your API key.",
		},
		{
			name:           "404 not found",
			statusCode:     http.StatusNotFound,
			responseBody:   map[string]string{"error": "Not found"},
			wantErr:        true,
			wantStatusCode: http.StatusNotFound,
			wantMessage:    "Resource not found (404)",
		},
		{
			name:           "409 conflict",
			statusCode:     http.StatusConflict,
			responseBody:   map[string]string{"error": "Already exists"},
			wantErr:        true,
			wantStatusCode: http.StatusConflict,
			wantMessage:    "Conflict: Already exists",
		},
		{
			name:           "422 unprocessable entity",
			statusCode:     http.StatusUnprocessableEntity,
			responseBody:   map[string]string{"message": "Validation failed"},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantMessage:    "Validation error: Validation failed",
		},
		{
			name:           "422 with errors array",
			statusCode:     http.StatusUnprocessableEntity,
			responseBody:   map[string]interface{}{"errors": []interface{}{"Error 1", "Error 2"}},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantMessage:    "Validation error: [Error 1 Error 2]",
		},
		{
			name:           "500 internal server error",
			statusCode:     http.StatusInternalServerError,
			responseBody:   map[string]string{"error": "Server error"},
			wantErr:        true,
			wantStatusCode: http.StatusInternalServerError,
			wantMessage:    "Server error (500). Please try again later.",
		},
		{
			name:           "502 bad gateway",
			statusCode:     http.StatusBadGateway,
			responseBody:   map[string]string{"error": "Bad gateway"},
			wantErr:        true,
			wantStatusCode: http.StatusBadGateway,
			wantMessage:    "Server error (502). Please try again later.",
		},
		{
			name:           "503 service unavailable",
			statusCode:     http.StatusServiceUnavailable,
			responseBody:   map[string]string{"error": "Service unavailable"},
			wantErr:        true,
			wantStatusCode: http.StatusServiceUnavailable,
			wantMessage:    "Server error (503). Please try again later.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")

			result := make(map[string]string)
			err := client.Get("/test", &result, nil)

			if tt.wantErr {
				require.Error(t, err)
				apiErr, ok := err.(*APIError)
				require.True(t, ok, "error should be of type *APIError")
				assert.Equal(t, tt.wantStatusCode, apiErr.StatusCode)
				if tt.wantMessage != "" {
					assert.Equal(t, tt.wantMessage, apiErr.Message)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
