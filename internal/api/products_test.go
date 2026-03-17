package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/products", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"products": []Product{
				{SCID: "prod-1", Name: "Product 1"},
				{SCID: "prod-2", Name: "Product 2"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	productsService := NewProducts(client)

	products, err := productsService.List()

	require.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, "Product 1", products[0].Name)
}

func TestProductsGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/products/prod-123", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Product{SCID: "prod-123", Name: "Test Product"})
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-store", "test-key")
	productsService := NewProducts(client)

	product, err := productsService.Get("prod-123")

	require.NoError(t, err)
	assert.Equal(t, "Test Product", product.Name)
}
