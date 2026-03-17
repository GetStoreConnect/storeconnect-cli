package utils

import (
	"testing"
)

func TestNormalizeSalesforceID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid 15 char ID",
			input:    "a0A7Z00000AbCdE",
			expected: "a0A7Z00000AbCdEUAV",
			wantErr:  false,
		},
		{
			name:     "valid 18 char ID",
			input:    "a0A7Z00000AbCdEUAV",
			expected: "a0A7Z00000AbCdEUAV",
			wantErr:  false,
		},
		{
			name:     "invalid length",
			input:    "invalid",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "too long",
			input:    "a0A7Z00000AbCdEFGHI",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeSalesforceID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeSalesforceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("NormalizeSalesforceID() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestValidateOrgID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid 15 char org ID",
			input: "00D7Z000000AbCd",
			valid: true,
		},
		{
			name:  "valid 18 char org ID",
			input: "00D7Z000000AbCdEFG",
			valid: true,
		},
		{
			name:  "valid lowercase prefix",
			input: "00d7Z000000AbCd",
			valid: true,
		},
		{
			name:  "invalid prefix",
			input: "a0A7Z000000AbCd",
			valid: false,
		},
		{
			name:  "invalid length",
			input: "00D7Z00",
			valid: false,
		},
		{
			name:  "empty string",
			input: "",
			valid: false,
		},
		{
			name:  "special characters",
			input: "00D7Z000000@bCd",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateOrgID(tt.input); got != tt.valid {
				t.Errorf("ValidateOrgID() = %v, expected %v", got, tt.valid)
			}
		})
	}
}

func TestValidateSalesforceID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{
			name:  "valid 15 char ID",
			input: "a0A7Z000000AbCd",
			valid: true,
		},
		{
			name:  "valid 18 char ID",
			input: "a0A7Z000000AbCdEFG",
			valid: true,
		},
		{
			name:  "invalid length",
			input: "a0A7Z00",
			valid: false,
		},
		{
			name:  "empty string",
			input: "",
			valid: false,
		},
		{
			name:  "special characters",
			input: "a0A7Z000000@bCd",
			valid: false,
		},
		{
			name:  "spaces",
			input: "a0A7Z00 000AbCd",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateSalesforceID(tt.input); got != tt.valid {
				t.Errorf("ValidateSalesforceID() = %v, expected %v", got, tt.valid)
			}
		})
	}
}
