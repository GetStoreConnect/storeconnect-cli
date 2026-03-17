package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestResponse defines a mock HTTP response
type TestResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// NewTestServer creates a test HTTP server with predefined responses
func NewTestServer(t *testing.T, responses map[string]TestResponse) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		response, ok := responses[key]
		if !ok {
			t.Logf("No mock response for: %s", key)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Set headers
		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}

		// Set status code
		w.WriteHeader(response.StatusCode)

		// Write body
		if response.Body != nil {
			if str, ok := response.Body.(string); ok {
				_, _ = w.Write([]byte(str))
			} else {
				_ = json.NewEncoder(w).Encode(response.Body)
			}
		}
	}))
}

// NewTestServerWithHandler creates a test server with custom handler
func NewTestServerWithHandler(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}
