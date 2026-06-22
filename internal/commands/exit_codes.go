package commands

// Exit codes for the CLI
// These provide semantic information about the type of error that occurred,
// allowing calling scripts and agents to handle different errors appropriately.
const (
	// ExitSuccess indicates the operation completed successfully
	ExitSuccess = 0

	// ExitGenericError indicates a generic error occurred
	ExitGenericError = 1

	// ExitAuthError indicates authentication failed (invalid credentials, expired token)
	ExitAuthError = 2

	// ExitNotFound indicates the requested resource was not found
	// (theme, server, content change, etc.)
	ExitNotFound = 3

	// ExitValidationError indicates validation failed
	// (invalid Liquid syntax, bad configuration, missing required fields)
	ExitValidationError = 4

	// ExitNetworkError indicates a network or connection error
	// (cannot reach server, timeout, DNS failure)
	ExitNetworkError = 5

	// ExitConflict indicates a resource conflict
	// (draft already exists, theme name taken, concurrent modification)
	ExitConflict = 6

	// ExitUserCancelled indicates the user cancelled the operation
	// (responded "no" to confirmation prompt, interrupted with Ctrl+C)
	ExitUserCancelled = 7

	// ExitConfigError indicates a configuration error
	// (missing config file, invalid YAML, no server configured)
	ExitConfigError = 8

	// ExitDiffChanges indicates `sc theme diff --exit-code` found differences
	// between the local theme and the server. It is NOT an error — the command
	// succeeded; this code lets CI distinguish "theme has undeployed drift"
	// from a genuine failure (which uses ExitGenericError and friends).
	ExitDiffChanges = 9
)
