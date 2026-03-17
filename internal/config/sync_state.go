package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// SyncState tracks synchronization state for a theme
type SyncState struct {
	ThemeSCID         string    `yaml:"theme_sc_id"`
	ThemeSFID         string    `yaml:"theme_sfid,omitempty"`
	ContentChangeSCID string    `yaml:"content_change_sc_id,omitempty"`
	LastPush          time.Time `yaml:"last_push,omitempty"`
	LastPull          time.Time `yaml:"last_pull,omitempty"`
	path              string
}

// NewSyncState loads or creates sync state for a theme
func NewSyncState(themePath string) (*SyncState, error) {
	statePath := filepath.Join(themePath, ".sync-state.yml")

	state := &SyncState{
		path: statePath,
	}

	// Load existing state if file exists
	if _, err := os.Stat(statePath); err == nil {
		data, err := os.ReadFile(statePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read sync state: %w", err)
		}

		if err := yaml.Unmarshal(data, state); err != nil {
			return nil, fmt.Errorf("failed to parse sync state: %w", err)
		}
	}

	return state, nil
}

// Save writes sync state to disk
func (s *SyncState) Save() error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal sync state: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write sync state: %w", err)
	}

	return nil
}

// UpdatePush updates the state after a push operation
func (s *SyncState) UpdatePush(contentChangeSCID string) error {
	s.ContentChangeSCID = contentChangeSCID
	s.LastPush = time.Now()
	return s.Save()
}

// UpdatePull updates the state after a pull operation
func (s *SyncState) UpdatePull(themeSCID, themeSFID string) error {
	s.ThemeSCID = themeSCID
	s.ThemeSFID = themeSFID
	s.LastPull = time.Now()
	return s.Save()
}

// ClearContentChange clears the content change ID (after publish)
func (s *SyncState) ClearContentChange() error {
	s.ContentChangeSCID = ""
	return s.Save()
}
