package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

// Client is the HTTP client for StoreConnect API
type Client struct {
	BaseURL     string
	OrgID       string
	StoreSFID   string
	APIKey      string
	ChangeSetID string
	httpClient  *resty.Client
}

// APIError represents an API error response with enhanced error information
type APIError struct {
	StatusCode int
	Message    string
	Suggestion string // Helpful suggestion for resolving the error
}

func (e *APIError) Error() string {
	return e.Message
}

// NewClient creates a new API client
func NewClient(baseURL, storeSFID, apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		BaseURL:   baseURL,
		StoreSFID: storeSFID,
		APIKey:    apiKey,
		httpClient: resty.New().
			SetTimeout(30*time.Second).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json"),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Set authorization header
	c.httpClient.SetHeader("Authorization", c.buildBearerToken())

	return c
}

// ClientOption is a functional option for configuring the client
type ClientOption func(*Client)

// WithOrgID sets the organization ID for the client
func WithOrgID(orgID string) ClientOption {
	return func(c *Client) {
		c.OrgID = orgID
	}
}

// WithChangeSetID sets the change set ID for the client
func WithChangeSetID(changeSetID string) ClientOption {
	return func(c *Client) {
		c.ChangeSetID = changeSetID
	}
}

func (c *Client) buildBearerToken() string {
	if c.OrgID != "" {
		// Enhanced format: org_id:store_sfid:api_key
		return fmt.Sprintf("Bearer %s:%s:%s", c.OrgID, c.StoreSFID, c.APIKey)
	}
	// Legacy format: store_sfid:api_key
	return fmt.Sprintf("Bearer %s:%s", c.StoreSFID, c.APIKey)
}

// Get performs a GET request
func (c *Client) Get(path string, result interface{}, params map[string]string) error {
	req := c.httpClient.R().SetResult(result)

	if params != nil {
		req.SetQueryParams(params)
	}

	if c.ChangeSetID != "" {
		req.SetHeader("X-SC-Change-Set-ID", c.ChangeSetID)
	}

	resp, err := req.Get(c.BaseURL + path)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("Connection error: %v", err),
			Suggestion: "Check your network connection and server URL",
		}
	}

	return c.handleResponse(resp)
}

// Post performs a POST request
func (c *Client) Post(path string, body, result interface{}) error {
	req := c.httpClient.R().
		SetBody(body).
		SetResult(result)

	if c.ChangeSetID != "" {
		req.SetHeader("X-SC-Change-Set-ID", c.ChangeSetID)
	}

	resp, err := req.Post(c.BaseURL + path)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("Connection error: %v", err),
			Suggestion: "Check your network connection and server URL",
		}
	}

	return c.handleResponse(resp)
}

// Put performs a PUT request
func (c *Client) Put(path string, body, result interface{}) error {
	req := c.httpClient.R().
		SetBody(body).
		SetResult(result)

	if c.ChangeSetID != "" {
		req.SetHeader("X-SC-Change-Set-ID", c.ChangeSetID)
	}

	resp, err := req.Put(c.BaseURL + path)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("Connection error: %v", err),
			Suggestion: "Check your network connection and server URL",
		}
	}

	return c.handleResponse(resp)
}

// Patch performs a PATCH request
func (c *Client) Patch(path string, body, result interface{}) error {
	req := c.httpClient.R().
		SetBody(body).
		SetResult(result)

	if c.ChangeSetID != "" {
		req.SetHeader("X-SC-Change-Set-ID", c.ChangeSetID)
	}

	resp, err := req.Patch(c.BaseURL + path)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("Connection error: %v", err),
			Suggestion: "Check your network connection and server URL",
		}
	}

	return c.handleResponse(resp)
}

// Delete performs a DELETE request
func (c *Client) Delete(path string, result interface{}) error {
	req := c.httpClient.R().SetResult(result)

	if c.ChangeSetID != "" {
		req.SetHeader("X-SC-Change-Set-ID", c.ChangeSetID)
	}

	resp, err := req.Delete(c.BaseURL + path)
	if err != nil {
		return &APIError{
			StatusCode: 0,
			Message:    fmt.Sprintf("Connection error: %v", err),
			Suggestion: "Check your network connection and server URL",
		}
	}

	return c.handleResponse(resp)
}

func (c *Client) handleResponse(resp *resty.Response) error {
	statusCode := resp.StatusCode()

	// Success
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}

	// Parse error response
	var errorData map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &errorData); err != nil {
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("Request failed with status %d", statusCode),
		}
	}

	// Extract error message - prefer detailed "message" over generic "error"
	var message string
	if msg, ok := errorData["message"].(string); ok && msg != "" {
		message = msg
	} else if msg, ok := errorData["error"].(string); ok && msg != "" {
		message = msg
	} else if errors, ok := errorData["errors"].([]interface{}); ok && len(errors) > 0 {
		// Join array of errors
		var msgs []string
		for _, e := range errors {
			if s, ok := e.(string); ok {
				msgs = append(msgs, s)
			}
		}
		message = fmt.Sprintf("%v", msgs)
	}

	switch statusCode {
	case http.StatusBadRequest:
		if message == "" {
			message = "Bad request"
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("Bad request: %s", message),
			Suggestion: "Check your input parameters and try again",
		}
	case http.StatusUnauthorized:
		return &APIError{
			StatusCode: statusCode,
			Message:    "Authentication failed. Please check your API key.",
			Suggestion: "Run 'sc status' to check credentials, or 'sc connect' to reconfigure",
		}
	case http.StatusNotFound:
		return &APIError{
			StatusCode: statusCode,
			Message:    "Resource not found (404)",
			Suggestion: "Verify the resource exists with 'sc theme list' or similar command",
		}
	case http.StatusConflict:
		if message == "" {
			message = "Resource already exists"
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("Conflict: %s", message),
			Suggestion: "Check for existing drafts or conflicting resources",
		}
	case http.StatusUnprocessableEntity:
		if message == "" {
			message = "Unprocessable Content"
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("Validation error: %s", message),
			Suggestion: "Check your input for errors. Use 'sc theme validate' for templates",
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return &APIError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("Server error (%d). Please try again later.", statusCode),
			Suggestion: "The server is experiencing issues. Wait a moment and try again",
		}
	default:
		if message == "" {
			message = fmt.Sprintf("Request failed with status %d", statusCode)
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    message,
			Suggestion: "Check server logs or contact support if the issue persists",
		}
	}
}
