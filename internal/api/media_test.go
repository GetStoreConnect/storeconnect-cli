package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaUploadURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/media/upload_url", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "image", body["file_type"])
		assert.Equal(t, "logo.png", body["filename"])
		assert.Equal(t, "image/png", body["content_type"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"upload_url": "https://uploads.example.com/signed",
			"upload_params": map[string]string{
				"signature": "abc",
				"timestamp": "123",
			},
			"asset_id": "asset-1",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	media := NewMedia(client)

	resp, err := media.UploadURL("image", "logo.png", "image/png")
	require.NoError(t, err)
	assert.Equal(t, "https://uploads.example.com/signed", resp.UploadURL)
	assert.Equal(t, "abc", resp.UploadParams["signature"])
	assert.Equal(t, "123", resp.UploadParams["timestamp"])
	assert.Equal(t, "asset-1", resp.AssetID)
}

func TestMediaUploadFile(t *testing.T) {
	tests := []struct {
		name        string
		response    map[string]interface{}
		wantURL     string
		wantErr     bool
		errContains string
	}{
		{
			name:     "prefers secure_url",
			response: map[string]interface{}{"secure_url": "https://cdn.example.com/secure.png", "url": "http://cdn.example.com/plain.png"},
			wantURL:  "https://cdn.example.com/secure.png",
		},
		{
			name:     "falls back to url",
			response: map[string]interface{}{"url": "http://cdn.example.com/plain.png"},
			wantURL:  "http://cdn.example.com/plain.png",
		},
		{
			name:        "missing hosted url",
			response:    map[string]interface{}{"public_id": "x"},
			wantErr:     true,
			errContains: "missing hosted URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)

				// Parse the multipart body and assert the params + file arrive.
				require.NoError(t, r.ParseMultipartForm(10<<20))

				assert.Equal(t, "sig-value", r.FormValue("signature"))
				assert.Equal(t, "1700000000", r.FormValue("timestamp"))

				file, header, err := r.FormFile("file")
				require.NoError(t, err)
				defer file.Close()
				// Go's multipart reader returns the base name (path components in
				// the Content-Disposition filename are stripped by the stdlib).
				assert.Equal(t, "logo.png", header.Filename)
				contents, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Equal(t, []byte("binary-bytes"), contents)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			client := NewClient(server.URL, "test-store", "test-key")
			media := NewMedia(client)

			upload := &UploadURLResponse{
				UploadURL: server.URL,
				UploadParams: map[string]string{
					"signature": "sig-value",
					"timestamp": "1700000000",
				},
			}

			hosted, err := media.UploadFile(upload, "images/logo.png", []byte("binary-bytes"))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantURL, hosted)
			}
		})
	}
}

func TestMediaUploadFileServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	media := NewMedia(client)

	upload := &UploadURLResponse{UploadURL: server.URL, UploadParams: map[string]string{}}
	_, err := media.UploadFile(upload, "logo.png", []byte("x"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}
