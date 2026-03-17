package utils

import (
	"fmt"
	"strings"
)

// NormalizeSalesforceID converts a 15-character Salesforce ID to 18 characters
// If already 18 characters, returns as-is
func NormalizeSalesforceID(id string) (string, error) {
	id = strings.TrimSpace(id)

	if len(id) == 18 {
		return id, nil
	}

	if len(id) != 15 {
		return "", fmt.Errorf("invalid Salesforce ID length: must be 15 or 18 characters")
	}

	// Convert to 18-character format by adding checksum suffix
	suffix := ""
	for i := 0; i < 3; i++ {
		flags := 0
		for j := 0; j < 5; j++ {
			c := id[i*5+j]
			if c >= 'A' && c <= 'Z' {
				flags |= 1 << j
			}
		}
		// Map flags to base32 character
		if flags <= 25 {
			suffix += string(rune('A' + flags))
		} else {
			suffix += string(rune('0' + (flags - 26)))
		}
	}

	return id + suffix, nil
}

// ValidateOrgID validates a Salesforce Organization ID format
func ValidateOrgID(orgID string) bool {
	if len(orgID) != 15 && len(orgID) != 18 {
		return false
	}

	// Must start with "00D"
	if len(orgID) >= 3 && !strings.EqualFold(orgID[:3], "00D") {
		return false
	}

	// Check alphanumeric
	for _, c := range orgID {
		if !isAlphaNumeric(c) {
			return false
		}
	}

	return true
}

// ValidateSalesforceID validates a Salesforce ID format
func ValidateSalesforceID(id string) bool {
	if len(id) != 15 && len(id) != 18 {
		return false
	}

	// Check alphanumeric
	for _, c := range id {
		if !isAlphaNumeric(c) {
			return false
		}
	}

	return true
}

func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
