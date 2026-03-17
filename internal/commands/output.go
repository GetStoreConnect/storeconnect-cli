package commands

import (
	"encoding/json"
	"os"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
)

// jsonOutput is set by the --json flag and determines output format
var jsonOutput bool

// outputJSON marshals data to JSON and prints it to stdout
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// outputResponse handles unified output for both success and error cases
// In JSON mode, it outputs structured JSON. In human mode, it returns nil
// to let the command handle its own human-friendly output.
func outputResponse(data interface{}, err error) error {
	if err != nil {
		return outputError(err)
	}

	if jsonOutput {
		return outputJSON(SuccessResponse{
			Success: true,
			Data:    data,
		})
	}

	// Human-friendly output is handled by the command itself
	return nil
}

// outputError handles error output and determines the appropriate exit code
func outputError(err error) error {
	exitCode := ExitGenericError
	errorCode := "GENERIC_ERROR"
	suggestion := ""

	// Map error types to exit codes and suggestions
	if apiErr, ok := err.(*api.APIError); ok {
		switch apiErr.StatusCode {
		case 401:
			exitCode = ExitAuthError
			errorCode = "AUTH_ERROR"
			suggestion = "Check your API key and credentials with 'sc status'"
		case 404:
			exitCode = ExitNotFound
			errorCode = "NOT_FOUND"
			suggestion = "Verify the resource exists with 'sc theme list' or similar command"
		case 409:
			exitCode = ExitConflict
			errorCode = "CONFLICT"
			suggestion = "Check for existing drafts or conflicting resources"
		case 422:
			exitCode = ExitValidationError
			errorCode = "VALIDATION_ERROR"
			suggestion = "Check your input for errors. Use 'sc theme validate' for templates"
		case 0:
			// Connection error (no status code)
			exitCode = ExitNetworkError
			errorCode = "NETWORK_ERROR"
			suggestion = "Check your network connection and server URL"
		case 500, 502, 503:
			exitCode = ExitNetworkError
			errorCode = "SERVER_ERROR"
			suggestion = "The server is experiencing issues. Try again later"
		}
	}

	// Check for config-related errors
	errMsg := err.Error()
	if contains(errMsg, "no server configured") || contains(errMsg, "server not found") ||
		contains(errMsg, "failed to load project") || contains(errMsg, "failed to load credentials") {
		exitCode = ExitConfigError
		errorCode = "CONFIG_ERROR"
		suggestion = "Run 'sc status' to check configuration, or 'sc connect' to set up a server"
	}

	// Output error in appropriate format
	if jsonOutput {
		_ = outputJSON(ErrorResponse{
			Success: false,
			Error: ErrorDetail{
				Code:    errorCode,
				Message: errMsg,
			},
			Suggestion: suggestion,
		})
	}

	os.Exit(exitCode)
	return err
}

// contains checks if a string contains a substring (case-sensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
