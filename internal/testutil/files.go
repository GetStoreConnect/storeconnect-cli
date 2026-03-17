package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// CreateTempProject creates a temporary project structure for testing
func CreateTempProject(t *testing.T) (dir string, cleanup func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "sc-test-*")
	require.NoError(t, err, "failed to create temp directory")

	cleanup = func() {
		_ = os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// WriteTestFile writes content to a test file
func WriteTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err, "failed to create directory: %s", dir)
	}

	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err, "failed to write file: %s", path)
}

// ReadTestFile reads content from a test file
func ReadTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %s", path)

	return string(content)
}

// CreateTestDir creates a test directory
func CreateTestDir(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, 0755)
	require.NoError(t, err, "failed to create directory: %s", path)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
