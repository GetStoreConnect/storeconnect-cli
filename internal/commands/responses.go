package commands

import "time"

// SuccessResponse is the standard JSON success response format
type SuccessResponse struct {
	Success bool                   `json:"success"`
	Data    interface{}            `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// ErrorResponse is the standard JSON error response format
type ErrorResponse struct {
	Success    bool        `json:"success"`
	Error      ErrorDetail `json:"error"`
	Suggestion string      `json:"suggestion,omitempty"`
}

// ErrorDetail provides structured error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ThemeListResponse is the response for theme list command
type ThemeListResponse struct {
	Themes []ThemeInfo `json:"themes"`
	Total  int         `json:"total"`
}

// ThemeInfo contains information about a single theme
type ThemeInfo struct {
	Name      string    `json:"name"`
	SCID      string    `json:"sc_id"`
	SFID      string    `json:"sfid,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// StatusResponse is the response for status command
type StatusResponse struct {
	Connected     bool                    `json:"connected"`
	Servers       map[string]ServerStatus `json:"servers"`
	DefaultServer string                  `json:"default_server,omitempty"`
	ProjectDir    string                  `json:"project_dir,omitempty"`
}

// ServerStatus contains status information for a single server
type ServerStatus struct {
	URL           string    `json:"url"`
	Version       string    `json:"version,omitempty"`
	Authenticated bool      `json:"authenticated"`
	LastChecked   time.Time `json:"last_checked,omitempty"`
}

// PreviewURLResponse is the response for theme preview command
type PreviewURLResponse struct {
	PreviewURL      string `json:"preview_url"`
	ContentChangeID string `json:"content_change_id"`
	ThemeID         string `json:"theme_id,omitempty"`
}

// ContentChangeResponse is the response for content change operations (push, publish)
type ContentChangeResponse struct {
	ContentChangeID string `json:"content_change_id"`
	ThemeName       string `json:"theme_name"`
	Status          string `json:"status"`
}

// ThemePullResponse is the response for theme pull command
type ThemePullResponse struct {
	ThemeName  string `json:"theme_name"`
	ThemeID    string `json:"theme_id"`
	FilesCount int    `json:"files_count"`
}

// ThemePublishResponse is the response for theme publish command
type ThemePublishResponse struct {
	ThemeName       string `json:"theme_name"`
	ContentChangeID string `json:"content_change_id"`
	Status          string `json:"status"`
	Message         string `json:"message,omitempty"`
}
