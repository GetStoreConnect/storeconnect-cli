package ui

import (
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

// Spinner represents a loading spinner
type Spinner struct {
	spinner *spinner.Spinner
	message string
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	_ = s.Color("cyan")

	return &Spinner{
		spinner: s,
		message: message,
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	s.spinner.Start()
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	s.spinner.Stop()
}

// Success stops the spinner and prints a success message
func (s *Spinner) Success(message string) {
	s.spinner.Stop()
	color.Green("✓ " + message)
}

// Error stops the spinner and prints an error message
func (s *Spinner) Error(message string) {
	s.spinner.Stop()
	color.Red("✗ " + message)
}

// UpdateMessage updates the spinner message
func (s *Spinner) UpdateMessage(message string) {
	s.message = message
	s.spinner.Suffix = " " + message
}

// Spin runs a function with a spinner
func Spin(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()
	defer s.Stop()
	return fn()
}
