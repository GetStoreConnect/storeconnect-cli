package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCredentials(t *testing.T) {
	t.Run("creates directory with correct permissions", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sc-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		t.Setenv("HOME", tmpDir)
		homedir.Reset() // Reset homedir cache

		creds, err := NewCredentials()
		require.NoError(t, err)
		assert.NotNil(t, creds)

		// Verify directory was created with correct permissions
		credDir := filepath.Join(tmpDir, ".storeconnect")
		info, err := os.Stat(credDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())

		// Verify empty servers map
		assert.Equal(t, map[string]ServerCredentials{}, creds.Servers)
	})

	t.Run("loads existing credentials from file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sc-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create credentials file before initializing
		credDir := filepath.Join(tmpDir, ".storeconnect")
		require.NoError(t, os.MkdirAll(credDir, 0700))
		credPath := filepath.Join(credDir, "credentials.yml")

		existingCreds := Credentials{
			Servers: map[string]ServerCredentials{
				"production": {
					URL:       "https://store.example.com",
					OrgID:     "00D7Z000000AbCd",
					StoreSFID: "a0A7Z000000AbCdEUAV",
					APIKey:    "secret-key-123",
				},
			},
		}

		data, err := yaml.Marshal(&existingCreds)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(credPath, data, 0600))

		t.Setenv("HOME", tmpDir)
		homedir.Reset() // Reset homedir cache

		// Load credentials
		creds, err := NewCredentials()
		require.NoError(t, err)
		assert.Equal(t, existingCreds.Servers, creds.Servers)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "sc-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		credDir := filepath.Join(tmpDir, ".storeconnect")
		require.NoError(t, os.MkdirAll(credDir, 0700))
		credPath := filepath.Join(credDir, "credentials.yml")

		invalidYAML := "servers:\n  production:\n    url: test\n    invalid: [[[\n      bad: yaml"
		require.NoError(t, os.WriteFile(credPath, []byte(invalidYAML), 0600))

		t.Setenv("HOME", tmpDir)
		homedir.Reset() // Reset homedir cache

		_, err = NewCredentials()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse credentials")
	})
}

func TestCredentials_AddServer(t *testing.T) {
	tests := []struct {
		name      string
		initial   map[string]ServerCredentials
		alias     string
		url       string
		storeSFID string
		apiKey    string
		orgID     string
		want      ServerCredentials
	}{
		{
			name:      "adds new server credentials",
			initial:   map[string]ServerCredentials{},
			alias:     "local",
			url:       "http://localhost:3000",
			storeSFID: "a0A7Z000000AbCdEUAV",
			apiKey:    "test-key",
			orgID:     "00D7Z000000AbCd",
			want: ServerCredentials{
				URL:       "http://localhost:3000",
				OrgID:     "00D7Z000000AbCd",
				StoreSFID: "a0A7Z000000AbCdEUAV",
				APIKey:    "test-key",
			},
		},
		{
			name: "updates existing server credentials",
			initial: map[string]ServerCredentials{
				"production": {
					URL:       "https://old.example.com",
					OrgID:     "00D7Z000000OldId",
					StoreSFID: "a0A7Z000000OldSFID",
					APIKey:    "old-key",
				},
			},
			alias:     "production",
			url:       "https://new.example.com",
			storeSFID: "a0A7Z000000NewSFID",
			apiKey:    "new-key",
			orgID:     "00D7Z000000NewId",
			want: ServerCredentials{
				URL:       "https://new.example.com",
				OrgID:     "00D7Z000000NewId",
				StoreSFID: "a0A7Z000000NewSFID",
				APIKey:    "new-key",
			},
		},
		{
			name:      "adds server without orgID",
			initial:   map[string]ServerCredentials{},
			alias:     "dev",
			url:       "https://dev.example.com",
			storeSFID: "a0A7Z000000DevSFID",
			apiKey:    "dev-key",
			orgID:     "",
			want: ServerCredentials{
				URL:       "https://dev.example.com",
				OrgID:     "",
				StoreSFID: "a0A7Z000000DevSFID",
				APIKey:    "dev-key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir, err := os.MkdirTemp("", "sc-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			credPath := filepath.Join(tmpDir, "credentials.yml")
			creds := &Credentials{
				Servers: tt.initial,
				path:    credPath,
			}

			// Execute
			err = creds.AddServer(tt.alias, tt.url, tt.storeSFID, tt.apiKey, tt.orgID)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.want, creds.Servers[tt.alias])

			// Verify file was saved
			assert.FileExists(t, credPath)

			// Verify file can be loaded
			loadedCreds, err := loadCredentialsFromPath(credPath)
			require.NoError(t, err)
			assert.Equal(t, tt.want, loadedCreds.Servers[tt.alias])
		})
	}
}

func TestCredentials_GetServer(t *testing.T) {
	tests := []struct {
		name      string
		servers   map[string]ServerCredentials
		alias     string
		wantCred  ServerCredentials
		wantFound bool
	}{
		{
			name: "returns existing server",
			servers: map[string]ServerCredentials{
				"production": {
					URL:       "https://store.example.com",
					OrgID:     "00D7Z000000AbCd",
					StoreSFID: "a0A7Z000000AbCdEUAV",
					APIKey:    "secret-key",
				},
			},
			alias: "production",
			wantCred: ServerCredentials{
				URL:       "https://store.example.com",
				OrgID:     "00D7Z000000AbCd",
				StoreSFID: "a0A7Z000000AbCdEUAV",
				APIKey:    "secret-key",
			},
			wantFound: true,
		},
		{
			name: "returns false for non-existent server",
			servers: map[string]ServerCredentials{
				"production": {
					URL:       "https://store.example.com",
					StoreSFID: "a0A7Z000000AbCdEUAV",
					APIKey:    "secret-key",
				},
			},
			alias:     "nonexistent",
			wantCred:  ServerCredentials{},
			wantFound: false,
		},
		{
			name:      "returns false for empty servers map",
			servers:   map[string]ServerCredentials{},
			alias:     "any",
			wantCred:  ServerCredentials{},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &Credentials{
				Servers: tt.servers,
				path:    "/tmp/test.yml",
			}

			cred, found := creds.GetServer(tt.alias)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantCred, *cred)
			}
		})
	}
}

func TestCredentials_RemoveServer(t *testing.T) {
	tests := []struct {
		name            string
		initial         map[string]ServerCredentials
		removeAlias     string
		wantRemaining   map[string]ServerCredentials
		shouldBeRemoved bool
	}{
		{
			name: "removes existing server",
			initial: map[string]ServerCredentials{
				"production": {
					URL:       "https://prod.example.com",
					StoreSFID: "a0A7Z000000ProdSFID",
					APIKey:    "prod-key",
				},
				"staging": {
					URL:       "https://staging.example.com",
					StoreSFID: "a0A7Z000000StagSFID",
					APIKey:    "staging-key",
				},
			},
			removeAlias: "staging",
			wantRemaining: map[string]ServerCredentials{
				"production": {
					URL:       "https://prod.example.com",
					StoreSFID: "a0A7Z000000ProdSFID",
					APIKey:    "prod-key",
				},
			},
			shouldBeRemoved: true,
		},
		{
			name: "handles removing non-existent server",
			initial: map[string]ServerCredentials{
				"production": {
					URL:       "https://prod.example.com",
					StoreSFID: "a0A7Z000000ProdSFID",
					APIKey:    "prod-key",
				},
			},
			removeAlias: "nonexistent",
			wantRemaining: map[string]ServerCredentials{
				"production": {
					URL:       "https://prod.example.com",
					StoreSFID: "a0A7Z000000ProdSFID",
					APIKey:    "prod-key",
				},
			},
			shouldBeRemoved: false,
		},
		{
			name: "removes last server leaving empty map",
			initial: map[string]ServerCredentials{
				"only": {
					URL:       "https://only.example.com",
					StoreSFID: "a0A7Z000000OnlySFID",
					APIKey:    "only-key",
				},
			},
			removeAlias:     "only",
			wantRemaining:   map[string]ServerCredentials{},
			shouldBeRemoved: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir, err := os.MkdirTemp("", "sc-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			credPath := filepath.Join(tmpDir, "credentials.yml")
			creds := &Credentials{
				Servers: tt.initial,
				path:    credPath,
			}

			// Execute
			err = creds.RemoveServer(tt.removeAlias)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, creds.Servers)

			// Verify removal
			_, found := creds.Servers[tt.removeAlias]
			assert.False(t, found)

			// Verify file was saved
			loadedCreds, err := loadCredentialsFromPath(credPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, loadedCreds.Servers)
		})
	}
}

func TestCredentials_Save(t *testing.T) {
	tests := []struct {
		name       string
		servers    map[string]ServerCredentials
		checkPerms bool
		wantPerms  os.FileMode
	}{
		{
			name: "saves credentials with correct permissions",
			servers: map[string]ServerCredentials{
				"production": {
					URL:       "https://store.example.com",
					OrgID:     "00D7Z000000AbCd",
					StoreSFID: "a0A7Z000000AbCdEUAV",
					APIKey:    "secret-key-123",
				},
			},
			checkPerms: true,
			wantPerms:  0600,
		},
		{
			name:       "saves empty credentials",
			servers:    map[string]ServerCredentials{},
			checkPerms: true,
			wantPerms:  0600,
		},
		{
			name: "saves multiple servers",
			servers: map[string]ServerCredentials{
				"prod": {
					URL:       "https://prod.example.com",
					StoreSFID: "a0A7Z000000ProdSFID",
					APIKey:    "prod-key",
				},
				"staging": {
					URL:       "https://staging.example.com",
					OrgID:     "00D7Z000000StagOrg",
					StoreSFID: "a0A7Z000000StagSFID",
					APIKey:    "staging-key",
				},
			},
			checkPerms: true,
			wantPerms:  0600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir, err := os.MkdirTemp("", "sc-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			credPath := filepath.Join(tmpDir, "credentials.yml")
			creds := &Credentials{
				Servers: tt.servers,
				path:    credPath,
			}

			// Execute
			err = creds.Save()

			// Verify
			require.NoError(t, err)
			assert.FileExists(t, credPath)

			// Check file permissions
			if tt.checkPerms {
				info, err := os.Stat(credPath)
				require.NoError(t, err)
				assert.Equal(t, tt.wantPerms, info.Mode().Perm(), "file should have 0600 permissions")
			}

			// Verify content can be loaded
			loadedCreds, err := loadCredentialsFromPath(credPath)
			require.NoError(t, err)
			assert.Equal(t, tt.servers, loadedCreds.Servers)
		})
	}
}

func TestCredentials_YAMLSerialization(t *testing.T) {
	tests := []struct {
		name     string
		creds    Credentials
		wantYAML string
	}{
		{
			name: "serializes single server correctly",
			creds: Credentials{
				Servers: map[string]ServerCredentials{
					"production": {
						URL:       "https://store.example.com",
						OrgID:     "00D7Z000000AbCd",
						StoreSFID: "a0A7Z000000AbCdEUAV",
						APIKey:    "secret-key",
					},
				},
			},
			wantYAML: `servers:
    production:
        url: https://store.example.com
        org_id: 00D7Z000000AbCd
        store_sfid: a0A7Z000000AbCdEUAV
        api_key: secret-key
`,
		},
		{
			name: "omits empty orgID field",
			creds: Credentials{
				Servers: map[string]ServerCredentials{
					"local": {
						URL:       "http://localhost:3000",
						StoreSFID: "a0A7Z000000LocalSFID",
						APIKey:    "local-key",
					},
				},
			},
			wantYAML: `servers:
    local:
        url: http://localhost:3000
        store_sfid: a0A7Z000000LocalSFID
        api_key: local-key
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.creds)
			require.NoError(t, err)
			assert.YAMLEq(t, tt.wantYAML, string(data))

			// Verify deserialization
			var loaded Credentials
			err = yaml.Unmarshal(data, &loaded)
			require.NoError(t, err)
			assert.Equal(t, tt.creds.Servers, loaded.Servers)
		})
	}
}

func TestCredentials_Path(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sc-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	expectedPath := filepath.Join(tmpDir, "credentials.yml")
	creds := &Credentials{
		Servers: make(map[string]ServerCredentials),
		path:    expectedPath,
	}

	assert.Equal(t, expectedPath, creds.Path())
}

// Helper function to load credentials from a specific path
func loadCredentialsFromPath(path string) (*Credentials, error) {
	creds := &Credentials{
		Servers: make(map[string]ServerCredentials),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, creds); err != nil {
		return nil, err
	}

	return creds, nil
}
