package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/GetStoreConnect/storeconnect-cli/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name              string
		setupFunc         func(t *testing.T, tmpDir string)
		wantServers       map[string]ServerInfo
		wantDefaultServer string
		wantErr           bool
	}{
		{
			name: "creates new project when config doesn't exist",
			setupFunc: func(t *testing.T, tmpDir string) {
				// Change to temp directory where no config exists
				require.NoError(t, os.Chdir(tmpDir))
			},
			wantServers:       map[string]ServerInfo{},
			wantDefaultServer: "",
			wantErr:           false,
		},
		{
			name: "loads existing project config",
			setupFunc: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.Chdir(tmpDir))

				configDir := filepath.Join(tmpDir, ".storeconnect")
				require.NoError(t, os.MkdirAll(configDir, 0755))

				existingProj := Project{
					DefaultServer: "production",
					Servers: map[string]ServerInfo{
						"production": {
							URL:                 "https://store.example.com",
							StoreconnectVersion: "1.2.3",
							BaseThemeVersion:    "2.0.0",
							LastSync:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
						},
					},
				}

				data, err := yaml.Marshal(&existingProj)
				require.NoError(t, err)
				testutil.WriteTestFile(t, filepath.Join(configDir, "config.yml"), string(data))
			},
			wantServers: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.2.3",
					BaseThemeVersion:    "2.0.0",
					LastSync:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			wantDefaultServer: "production",
			wantErr:           false,
		},
		{
			name: "returns error for invalid YAML",
			setupFunc: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.Chdir(tmpDir))

				configDir := filepath.Join(tmpDir, ".storeconnect")
				require.NoError(t, os.MkdirAll(configDir, 0755))

				invalidYAML := "invalid: yaml: content: [[[:"
				testutil.WriteTestFile(t, filepath.Join(configDir, "config.yml"), invalidYAML)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and save original working directory
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)

			// Setup test scenario
			tt.setupFunc(t, tmpDir)

			// Execute
			proj, err := NewProject()

			// Verify
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, proj)
			assert.Equal(t, filepath.Join(".storeconnect", "config.yml"), proj.path)
			assert.Equal(t, tt.wantDefaultServer, proj.DefaultServer)
			assert.Equal(t, tt.wantServers, proj.Servers)
		})
	}
}

func TestProject_Exists(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tmpDir string)
		want      bool
	}{
		{
			name: "returns true when config exists",
			setupFunc: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.Chdir(tmpDir))
				configDir := filepath.Join(tmpDir, ".storeconnect")
				require.NoError(t, os.MkdirAll(configDir, 0755))
				testutil.WriteTestFile(t, filepath.Join(configDir, "config.yml"), "servers: {}")
			},
			want: true,
		},
		{
			name: "returns false when config doesn't exist",
			setupFunc: func(t *testing.T, tmpDir string) {
				require.NoError(t, os.Chdir(tmpDir))
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)

			tt.setupFunc(t, tmpDir)

			proj := &Project{
				Servers: make(map[string]ServerInfo),
				path:    filepath.Join(".storeconnect", "config.yml"),
			}

			assert.Equal(t, tt.want, proj.Exists())
		})
	}
}

func TestProject_Init(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantDirs    []string
	}{
		{
			name:        "creates project directory structure",
			projectName: "my-store",
			wantDirs: []string{
				"my-store",
				"my-store/.storeconnect",
				"my-store/themes",
			},
		},
		{
			name:        "creates nested project",
			projectName: "clients/acme-corp",
			wantDirs: []string{
				"clients/acme-corp",
				"clients/acme-corp/.storeconnect",
				"clients/acme-corp/themes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			origDir, err := os.Getwd()
			require.NoError(t, err)
			defer os.Chdir(origDir)

			require.NoError(t, os.Chdir(tmpDir))

			proj := &Project{
				Servers: make(map[string]ServerInfo),
				path:    filepath.Join(".storeconnect", "config.yml"),
			}

			// Execute
			err = proj.Init(tt.projectName)

			// Verify
			require.NoError(t, err)

			// Check all directories were created
			for _, dir := range tt.wantDirs {
				fullPath := filepath.Join(tmpDir, dir)
				assert.True(t, testutil.DirExists(fullPath), "directory should exist: %s", dir)
			}

			// Check config file was created
			configPath := filepath.Join(tmpDir, tt.projectName, ".storeconnect", "config.yml")
			assert.True(t, testutil.FileExists(configPath), "config file should exist")

			// Verify path was updated
			assert.Equal(t, filepath.Join(tt.projectName, ".storeconnect", "config.yml"), proj.path)
		})
	}
}

func TestProject_AddServer(t *testing.T) {
	tests := []struct {
		name             string
		initial          map[string]ServerInfo
		initialDefault   string
		alias            string
		url              string
		setAsDefault     bool
		wantServer       ServerInfo
		wantDefaultAlias string
	}{
		{
			name:           "adds new server and sets as default when none exist",
			initial:        map[string]ServerInfo{},
			initialDefault: "",
			alias:          "production",
			url:            "https://store.example.com",
			setAsDefault:   false,
			wantServer: ServerInfo{
				URL: "https://store.example.com",
			},
			wantDefaultAlias: "production",
		},
		{
			name: "adds new server without changing default",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			initialDefault: "production",
			alias:          "staging",
			url:            "https://staging.example.com",
			setAsDefault:   false,
			wantServer: ServerInfo{
				URL: "https://staging.example.com",
			},
			wantDefaultAlias: "production",
		},
		{
			name: "adds new server and sets as default explicitly",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			initialDefault: "production",
			alias:          "staging",
			url:            "https://staging.example.com",
			setAsDefault:   true,
			wantServer: ServerInfo{
				URL: "https://staging.example.com",
			},
			wantDefaultAlias: "staging",
		},
		{
			name: "updates existing server without changing default",
			initial: map[string]ServerInfo{
				"production": {
					URL:                 "https://old.example.com",
					StoreconnectVersion: "1.0.0",
				},
			},
			initialDefault: "production",
			alias:          "production",
			url:            "https://new.example.com",
			setAsDefault:   false,
			wantServer: ServerInfo{
				URL: "https://new.example.com",
			},
			wantDefaultAlias: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			configPath := filepath.Join(tmpDir, "config.yml")
			proj := &Project{
				DefaultServer: tt.initialDefault,
				Servers:       tt.initial,
				path:          configPath,
			}

			// Execute
			err := proj.AddServer(tt.alias, tt.url, tt.setAsDefault)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.wantServer, proj.Servers[tt.alias])
			assert.Equal(t, tt.wantDefaultAlias, proj.DefaultServer)

			// Verify file was saved
			loadedProj, err := loadProjectFromPath(configPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantServer, loadedProj.Servers[tt.alias])
			assert.Equal(t, tt.wantDefaultAlias, loadedProj.DefaultServer)
		})
	}
}

func TestProject_UpdateServerInfo(t *testing.T) {
	tests := []struct {
		name                    string
		initial                 map[string]ServerInfo
		alias                   string
		storeconnectVersion     string
		baseThemeVersion        string
		wantStoreconnectVersion string
		wantBaseThemeVersion    string
		wantErr                 bool
		checkLastSync           bool
	}{
		{
			name: "updates both versions",
			initial: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.0.0",
					BaseThemeVersion:    "1.0.0",
				},
			},
			alias:                   "production",
			storeconnectVersion:     "2.0.0",
			baseThemeVersion:        "3.0.0",
			wantStoreconnectVersion: "2.0.0",
			wantBaseThemeVersion:    "3.0.0",
			checkLastSync:           true,
		},
		{
			name: "updates only storeconnect version",
			initial: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.0.0",
					BaseThemeVersion:    "2.0.0",
				},
			},
			alias:                   "production",
			storeconnectVersion:     "1.5.0",
			baseThemeVersion:        "",
			wantStoreconnectVersion: "1.5.0",
			wantBaseThemeVersion:    "2.0.0",
			checkLastSync:           true,
		},
		{
			name: "updates only base theme version",
			initial: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.0.0",
					BaseThemeVersion:    "2.0.0",
				},
			},
			alias:                   "production",
			storeconnectVersion:     "",
			baseThemeVersion:        "2.5.0",
			wantStoreconnectVersion: "1.0.0",
			wantBaseThemeVersion:    "2.5.0",
			checkLastSync:           true,
		},
		{
			name: "returns error for non-existent server",
			initial: map[string]ServerInfo{
				"production": {URL: "https://store.example.com"},
			},
			alias:         "nonexistent",
			wantErr:       true,
			checkLastSync: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			configPath := filepath.Join(tmpDir, "config.yml")
			proj := &Project{
				Servers: tt.initial,
				path:    configPath,
			}

			// Record time before update
			timeBefore := time.Now()

			// Execute
			err := proj.UpdateServerInfo(tt.alias, tt.storeconnectVersion, tt.baseThemeVersion)

			// Verify error case
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			// Verify success case
			require.NoError(t, err)

			server, ok := proj.Servers[tt.alias]
			require.True(t, ok, "server should exist")
			assert.Equal(t, tt.wantStoreconnectVersion, server.StoreconnectVersion)
			assert.Equal(t, tt.wantBaseThemeVersion, server.BaseThemeVersion)

			// Verify LastSync was updated
			if tt.checkLastSync {
				assert.True(t, server.LastSync.After(timeBefore) || server.LastSync.Equal(timeBefore),
					"LastSync should be updated to current time")
			}

			// Verify file was saved
			loadedProj, err := loadProjectFromPath(configPath)
			require.NoError(t, err)
			loadedServer, ok := loadedProj.Servers[tt.alias]
			require.True(t, ok)
			assert.Equal(t, tt.wantStoreconnectVersion, loadedServer.StoreconnectVersion)
			assert.Equal(t, tt.wantBaseThemeVersion, loadedServer.BaseThemeVersion)
		})
	}
}

func TestProject_GetServer(t *testing.T) {
	tests := []struct {
		name      string
		servers   map[string]ServerInfo
		alias     string
		wantInfo  ServerInfo
		wantFound bool
	}{
		{
			name: "returns existing server",
			servers: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.2.3",
					BaseThemeVersion:    "2.0.0",
					LastSync:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
			alias: "production",
			wantInfo: ServerInfo{
				URL:                 "https://store.example.com",
				StoreconnectVersion: "1.2.3",
				BaseThemeVersion:    "2.0.0",
				LastSync:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			wantFound: true,
		},
		{
			name: "returns false for non-existent server",
			servers: map[string]ServerInfo{
				"production": {URL: "https://store.example.com"},
			},
			alias:     "nonexistent",
			wantInfo:  ServerInfo{},
			wantFound: false,
		},
		{
			name:      "returns false for empty servers map",
			servers:   map[string]ServerInfo{},
			alias:     "any",
			wantInfo:  ServerInfo{},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj := &Project{
				Servers: tt.servers,
				path:    "/tmp/test.yml",
			}

			info, found := proj.GetServer(tt.alias)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantInfo, *info)
			}
		})
	}
}

func TestProject_RemoveServer(t *testing.T) {
	tests := []struct {
		name              string
		initial           map[string]ServerInfo
		initialDefault    string
		removeAlias       string
		wantRemaining     map[string]ServerInfo
		wantDefaultServer string
	}{
		{
			name: "removes server and keeps default",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
				"staging":    {URL: "https://staging.example.com"},
			},
			initialDefault: "production",
			removeAlias:    "staging",
			wantRemaining: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			wantDefaultServer: "production",
		},
		{
			name: "removes default server and sets new default",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
				"staging":    {URL: "https://staging.example.com"},
			},
			initialDefault: "production",
			removeAlias:    "production",
			wantRemaining: map[string]ServerInfo{
				"staging": {URL: "https://staging.example.com"},
			},
			wantDefaultServer: "staging",
		},
		{
			name: "removes last server and clears default",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			initialDefault:    "production",
			removeAlias:       "production",
			wantRemaining:     map[string]ServerInfo{},
			wantDefaultServer: "",
		},
		{
			name: "handles removing non-existent server",
			initial: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			initialDefault: "production",
			removeAlias:    "nonexistent",
			wantRemaining: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			wantDefaultServer: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			configPath := filepath.Join(tmpDir, "config.yml")
			proj := &Project{
				DefaultServer: tt.initialDefault,
				Servers:       tt.initial,
				path:          configPath,
			}

			// Execute
			err := proj.RemoveServer(tt.removeAlias)

			// Verify
			require.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, proj.Servers)
			assert.Equal(t, tt.wantDefaultServer, proj.DefaultServer)

			// Verify removal
			_, found := proj.Servers[tt.removeAlias]
			assert.False(t, found)

			// Verify file was saved
			loadedProj, err := loadProjectFromPath(configPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantRemaining, loadedProj.Servers)
			assert.Equal(t, tt.wantDefaultServer, loadedProj.DefaultServer)
		})
	}
}

func TestProject_SetDefaultServer(t *testing.T) {
	tests := []struct {
		name        string
		servers     map[string]ServerInfo
		oldDefault  string
		newDefault  string
		wantErr     bool
		wantDefault string
	}{
		{
			name: "sets default to existing server",
			servers: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
				"staging":    {URL: "https://staging.example.com"},
			},
			oldDefault:  "production",
			newDefault:  "staging",
			wantErr:     false,
			wantDefault: "staging",
		},
		{
			name: "returns error for non-existent server",
			servers: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			oldDefault:  "production",
			newDefault:  "nonexistent",
			wantErr:     true,
			wantDefault: "production",
		},
		{
			name: "sets default when none was set",
			servers: map[string]ServerInfo{
				"production": {URL: "https://prod.example.com"},
			},
			oldDefault:  "",
			newDefault:  "production",
			wantErr:     false,
			wantDefault: "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			configPath := filepath.Join(tmpDir, "config.yml")
			proj := &Project{
				DefaultServer: tt.oldDefault,
				Servers:       tt.servers,
				path:          configPath,
			}

			// Execute
			err := proj.SetDefaultServer(tt.newDefault)

			// Verify
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantDefault, proj.DefaultServer)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantDefault, proj.DefaultServer)

			// Verify file was saved
			loadedProj, err := loadProjectFromPath(configPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDefault, loadedProj.DefaultServer)
		})
	}
}

func TestProject_GetDefaultServer(t *testing.T) {
	tests := []struct {
		name    string
		project Project
		want    string
	}{
		{
			name: "returns default server",
			project: Project{
				DefaultServer: "production",
				Servers: map[string]ServerInfo{
					"production": {URL: "https://store.example.com"},
				},
			},
			want: "production",
		},
		{
			name: "returns empty string when no default",
			project: Project{
				Servers: map[string]ServerInfo{
					"production": {URL: "https://store.example.com"},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.project.GetDefaultServer())
		})
	}
}

func TestProject_Save(t *testing.T) {
	tests := []struct {
		name          string
		defaultServer string
		servers       map[string]ServerInfo
	}{
		{
			name:          "saves project with single server",
			defaultServer: "production",
			servers: map[string]ServerInfo{
				"production": {
					URL:                 "https://store.example.com",
					StoreconnectVersion: "1.2.3",
					BaseThemeVersion:    "2.0.0",
					LastSync:            time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:          "saves project with multiple servers",
			defaultServer: "production",
			servers: map[string]ServerInfo{
				"production": {
					URL:                 "https://prod.example.com",
					StoreconnectVersion: "1.2.3",
				},
				"staging": {
					URL:              "https://staging.example.com",
					BaseThemeVersion: "2.0.0",
				},
			},
		},
		{
			name:          "saves empty project",
			defaultServer: "",
			servers:       map[string]ServerInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, cleanup := testutil.CreateTempProject(t)
			defer cleanup()

			configPath := filepath.Join(tmpDir, ".storeconnect", "config.yml")
			proj := &Project{
				DefaultServer: tt.defaultServer,
				Servers:       tt.servers,
				path:          configPath,
			}

			// Execute
			err := proj.Save()

			// Verify
			require.NoError(t, err)
			assert.True(t, testutil.FileExists(configPath))

			// Verify directory was created
			assert.True(t, testutil.DirExists(filepath.Dir(configPath)))

			// Verify content can be loaded
			loadedProj, err := loadProjectFromPath(configPath)
			require.NoError(t, err)
			assert.Equal(t, tt.defaultServer, loadedProj.DefaultServer)
			assert.Equal(t, tt.servers, loadedProj.Servers)
		})
	}
}

func TestProject_YAMLSerialization(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		wantYAML string
	}{
		{
			name: "serializes complete project",
			project: Project{
				DefaultServer: "production",
				Servers: map[string]ServerInfo{
					"production": {
						URL:                 "https://store.example.com",
						StoreconnectVersion: "1.2.3",
						BaseThemeVersion:    "2.0.0",
					},
				},
			},
			wantYAML: `default_server: production
servers:
    production:
        url: https://store.example.com
        storeconnect_version: 1.2.3
        base_theme_version: 2.0.0
`,
		},
		{
			name: "omits empty fields",
			project: Project{
				DefaultServer: "local",
				Servers: map[string]ServerInfo{
					"local": {
						URL: "http://localhost:3000",
					},
				},
			},
			wantYAML: `default_server: local
servers:
    local:
        url: http://localhost:3000
`,
		},
		{
			name: "serializes project without default server",
			project: Project{
				Servers: map[string]ServerInfo{
					"production": {
						URL: "https://store.example.com",
					},
				},
			},
			wantYAML: `servers:
    production:
        url: https://store.example.com
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.project)
			require.NoError(t, err)
			assert.YAMLEq(t, tt.wantYAML, string(data))

			// Verify deserialization
			var loaded Project
			err = yaml.Unmarshal(data, &loaded)
			require.NoError(t, err)
			assert.Equal(t, tt.project.DefaultServer, loaded.DefaultServer)
			assert.Equal(t, tt.project.Servers, loaded.Servers)
		})
	}
}

// Helper function to load project from a specific path
func loadProjectFromPath(path string) (*Project, error) {
	proj := &Project{
		Servers: make(map[string]ServerInfo),
		path:    path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, proj); err != nil {
		return nil, err
	}

	return proj, nil
}
