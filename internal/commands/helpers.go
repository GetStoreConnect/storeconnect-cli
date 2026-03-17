package commands

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	// Simple check: UUIDs are 36 chars with hyphens at positions 8, 13, 18, 23
	if len(s) != 36 {
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	return true
}
