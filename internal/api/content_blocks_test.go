package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentBlocksList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/content_blocks", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content_blocks": []ContentBlock{
				{SCID: "block-1", Name: "Block 1"},
				{SCID: "block-2", Name: "Block 2"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	blocksService := NewContentBlocks(client)

	blocks, err := blocksService.List()

	require.NoError(t, err)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "Block 1", blocks[0].Name)
}

func TestContentBlocksGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/content_blocks/block-123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ContentBlock{SCID: "block-123", Name: "Test Block"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	blocksService := NewContentBlocks(client)

	block, err := blocksService.Get("block-123")

	require.NoError(t, err)
	assert.Equal(t, "Test Block", block.Name)
}
