package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticlesList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/articles/content", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []Article{
				{SCID: "art-1", Title: "Article 1"},
				{SCID: "art-2", Title: "Article 2"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	articlesService := NewArticles(client)

	articles, err := articlesService.List()

	require.NoError(t, err)
	assert.Len(t, articles, 2)
	assert.Equal(t, "Article 1", articles[0].Title)
}

func TestArticlesGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/articles/art-123/content", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Article{SCID: "art-123", Title: "Test Article"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	articlesService := NewArticles(client)

	article, err := articlesService.Get("art-123")

	require.NoError(t, err)
	assert.Equal(t, "Test Article", article.Title)
}
