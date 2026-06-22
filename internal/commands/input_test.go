package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCredentialInput(t *testing.T) {
	// Save original values
	origNonInteractive := nonInteractive
	defer func() { nonInteractive = origNonInteractive }()

	tests := []struct {
		name             string
		flagValue        string
		envVar           string
		envValue         string
		nonInteractiveFn func() bool
		wantValue        string
		wantErr          bool
		errContains      string
	}{
		{
			name:      "returns flag value when provided",
			flagValue: "flag-value",
			envVar:    "TEST_ENV",
			envValue:  "env-value",
			wantValue: "flag-value",
			wantErr:   false,
		},
		{
			name:      "returns env value when flag not provided",
			flagValue: "",
			envVar:    "TEST_ENV",
			envValue:  "env-value",
			wantValue: "env-value",
			wantErr:   false,
		},
		{
			name:        "errors in non-interactive mode without flag or env",
			flagValue:   "",
			envVar:      "TEST_ENV",
			envValue:    "",
			wantValue:   "",
			wantErr:     true,
			errContains: "required in non-interactive mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			// Set non-interactive mode for error test
			if tt.wantErr {
				nonInteractive = true
			} else {
				nonInteractive = false
			}

			got, err := getCredentialInput(tt.flagValue, tt.envVar, "Test Input", "test-flag")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestGetSecretInput(t *testing.T) {
	// Save original values
	origNonInteractive := nonInteractive
	defer func() { nonInteractive = origNonInteractive }()

	tests := []struct {
		name        string
		flagValue   string
		envVar      string
		envValue    string
		wantValue   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "returns flag value when provided",
			flagValue: "secret-from-flag",
			envVar:    "TEST_SECRET",
			envValue:  "secret-from-env",
			wantValue: "secret-from-flag",
			wantErr:   false,
		},
		{
			name:      "returns env value when flag not provided",
			flagValue: "",
			envVar:    "TEST_SECRET",
			envValue:  "secret-from-env",
			wantValue: "secret-from-env",
			wantErr:   false,
		},
		{
			name:        "errors in non-interactive mode without flag or env",
			flagValue:   "",
			envVar:      "TEST_SECRET",
			envValue:    "",
			wantValue:   "",
			wantErr:     true,
			errContains: "required in non-interactive mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}

			// Set non-interactive mode for error test
			if tt.wantErr {
				nonInteractive = true
			} else {
				nonInteractive = false
			}

			got, err := getSecretInput(tt.flagValue, tt.envVar, "API Key", "api-key")

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestConfirmAction(t *testing.T) {
	// Save original values
	origNonInteractive := nonInteractive
	origYesFlag := yesFlag
	defer func() {
		nonInteractive = origNonInteractive
		yesFlag = origYesFlag
	}()

	tests := []struct {
		name      string
		yesFlag   bool
		nonInt    bool
		wantValue bool
		wantErr   bool
	}{
		{
			name:      "returns true when yes flag set",
			yesFlag:   true,
			nonInt:    false,
			wantValue: true,
			wantErr:   false,
		},
		{
			name:      "returns true when yes flag set (non-interactive)",
			yesFlag:   true,
			nonInt:    true,
			wantValue: true,
			wantErr:   false,
		},
		{
			name:      "errors in non-interactive mode without yes flag",
			yesFlag:   false,
			nonInt:    true,
			wantValue: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yesFlag = tt.yesFlag
			nonInteractive = tt.nonInt

			got, err := confirmAction("Proceed?", tt.yesFlag)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "confirmation required")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}
