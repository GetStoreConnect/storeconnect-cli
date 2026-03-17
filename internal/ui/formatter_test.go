package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	formatter := NewFormatter()
	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.success)
	assert.NotNil(t, formatter.error)
	assert.NotNil(t, formatter.warning)
	assert.NotNil(t, formatter.info)
	assert.NotNil(t, formatter.dim)
}

func TestFormatter_Success(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		formatter.Success("Operation completed")
	})
}

func TestFormatter_Error(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		formatter.Error("Operation failed")
	})
}

func TestFormatter_Warning(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		formatter.Warning("Please check this")
	})
}

func TestFormatter_Info(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		formatter.Info("Information message")
	})
}

func TestFormatter_Dim(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		formatter.Dim("Dimmed text")
	})
}

func TestFormatter_Print(t *testing.T) {
	formatter := NewFormatter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.Print("Plain text")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, "Plain text\n", output)
}

func TestFormatter_Newline(t *testing.T) {
	formatter := NewFormatter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.Newline()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, "\n", output)
}

func TestFormatter_MultipleMessages(t *testing.T) {
	formatter := NewFormatter()
	// Just verify it doesn't panic when called multiple times
	assert.NotPanics(t, func() {
		formatter.Success("First")
		formatter.Error("Second")
		formatter.Info("Third")
		formatter.Warning("Fourth")
		formatter.Dim("Fifth")
	})
}

func TestFormatter_EmptyMessage(t *testing.T) {
	formatter := NewFormatter()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.Print("")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, "\n", output)
}

func TestFormatter_LongMessage(t *testing.T) {
	formatter := NewFormatter()
	longMsg := strings.Repeat("x", 200)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	formatter.Print(longMsg)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	output := buf.String()
	assert.Equal(t, fmt.Sprintf("%s\n", longMsg), output)
}

func TestFormatter_AllMethods(t *testing.T) {
	formatter := NewFormatter()

	// Test that all methods can be called without panicking
	tests := []struct {
		name string
		fn   func()
	}{
		{"Success", func() { formatter.Success("test") }},
		{"Error", func() { formatter.Error("test") }},
		{"Warning", func() { formatter.Warning("test") }},
		{"Info", func() { formatter.Info("test") }},
		{"Dim", func() { formatter.Dim("test") }},
		{"Print", func() { formatter.Print("test") }},
		{"Newline", func() { formatter.Newline() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.fn)
		})
	}
}
