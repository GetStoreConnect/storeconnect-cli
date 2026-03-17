package ui

import (
	"fmt"

	"github.com/fatih/color"
)

// Formatter handles colored terminal output
type Formatter struct {
	success *color.Color
	error   *color.Color
	warning *color.Color
	info    *color.Color
	dim     *color.Color
}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{
		success: color.New(color.FgGreen),
		error:   color.New(color.FgRed),
		warning: color.New(color.FgYellow),
		info:    color.New(color.FgCyan),
		dim:     color.New(color.Faint),
	}
}

// Success prints a success message
func (f *Formatter) Success(message string) {
	f.success.Println("✓ " + message)
}

// Error prints an error message
func (f *Formatter) Error(message string) {
	f.error.Println("✗ " + message)
}

// Warning prints a warning message
func (f *Formatter) Warning(message string) {
	f.warning.Println("⚠ " + message)
}

// Info prints an info message
func (f *Formatter) Info(message string) {
	f.info.Println("ℹ " + message)
}

// Dim prints a dimmed message
func (f *Formatter) Dim(message string) {
	f.dim.Println(message)
}

// Newline prints a blank line
func (f *Formatter) Newline() {
	fmt.Println()
}

// Print prints a plain message
func (f *Formatter) Print(message string) {
	fmt.Println(message)
}
