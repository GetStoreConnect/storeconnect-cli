# CLAUDE.md - Go CLI Implementation

This file provides guidance to Claude Code when working with the Go implementation of the StoreConnect CLI.

## Overview

This is a **Go port** of the Ruby StoreConnect CLI (`/cli/`). It provides the same functionality but leverages Go's:
- Static typing and compile-time safety
- Fast execution and small binary size
- Cross-platform compilation
- Growing ecosystem of CLI tools

## Recent Updates

### Milestone 1: JSON Output & Exit Codes ✅ (2026-03-17)

**Status:** Complete and tested

Added agent-friendly JSON output mode and semantic exit codes to make the CLI fully usable by AI agents and automation scripts.

**New Features:**
- `--json` flag for machine-readable output (all commands)
- 9 semantic exit codes (0-8) for different error types
- Structured error responses with helpful suggestions
- Backward compatible (human-friendly output by default)

**Files Added:**
- `internal/commands/exit_codes.go` - Exit code constants
- `internal/commands/responses.go` - JSON response types
- `internal/commands/output.go` - Output handlers

**Commands Updated:**
- `sc theme list --json` - Returns ThemeListResponse
- `sc theme push --json` - Returns ContentChangeResponse
- `sc theme pull --json` - Returns ThemePullResponse
- `sc theme preview --json` - Returns PreviewURLResponse
- `sc theme publish --json` - Returns ThemePublishResponse
- `sc status --json` - Returns StatusResponse

**Usage:**
```bash
# Machine-readable output
sc theme list --json | jq -r '.data.themes[].name'

# Check exit codes
sc theme list --json
echo $?  # 0 = success, 8 = config error, etc.

# Error handling
if ! sc status --json > /tmp/status.json; then
    jq -r '.suggestion' /tmp/status.json
fi
```

**Documentation:** See `MILESTONE_1_COMPLETE.md` for full details.

## Technology Stack

- **Go 1.21+**: Core language
- **Cobra**: CLI framework (used by kubectl, hugo, docker)
- **Viper**: Configuration management
- **Resty**: HTTP client
- **fatih/color**: Terminal colors
- **briandowns/spinner**: Loading spinners
- **go-homedir**: Cross-platform home directory
- **testify**: Testing framework (planned)

## Project Structure (Standard Go Layout)

```
cli-go/
├── cmd/                 # Application entry points
│   └── sc/             # Main CLI binary
│       └── main.go     # Entry point
├── internal/            # Private application code
│   ├── api/            # API client and services
│   │   ├── client.go   # HTTP client with enhanced errors
│   │   ├── auth.go     # Authentication service
│   │   └── themes.go   # Theme service
│   ├── commands/       # Command implementations
│   │   ├── root.go     # Root command (--json flag)
│   │   ├── version.go  # Version constant
│   │   ├── exit_codes.go    # Exit code constants (NEW)
│   │   ├── responses.go     # JSON response types (NEW)
│   │   ├── output.go        # Output handlers (NEW)
│   │   ├── init.go     # Project initialization
│   │   ├── connect.go  # Server connection
│   │   ├── status.go   # Status display (JSON support)
│   │   └── theme*.go   # Theme commands (JSON support)
│   ├── config/         # Configuration management
│   │   ├── credentials.go  # Global credentials
│   │   └── project.go      # Project config
│   ├── content/        # Content serializers (planned)
│   ├── theme/          # Theme serializers
│   ├── ui/             # UI helpers
│   │   ├── formatter.go    # Colored output
│   │   └── spinner.go      # Loading spinners
│   ├── utils/          # Utilities
│   │   └── salesforce_id.go # Salesforce ID validation
│   └── validators/     # Validators
├── pkg/                # Public libraries (if needed)
├── docs/               # Documentation
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
├── Makefile            # Build tasks
├── README.md           # User documentation
├── CLAUDE.md           # This file
├── MILESTONE_1_COMPLETE.md  # Milestone 1 documentation
└── AGENT_FRIENDLY_CLI_PLAN.md  # Full 8-milestone roadmap
```

## Go Best Practices Applied

### 1. Standard Project Layout

Follows https://github.com/golang-standards/project-layout:
- `cmd/` - Application entry points
- `internal/` - Private application code (cannot be imported by other projects)
- `pkg/` - Public libraries (can be imported)

### 2. Error Handling

Go uses explicit error returns instead of exceptions:

```go
// Good: Explicit error handling with context
result, err := client.Get("/api/v1/themes")
if err != nil {
    return fmt.Errorf("failed to fetch themes: %w", err)
}

// Milestone 1: Semantic exit codes
if err != nil {
    return outputError(err)  // Maps to appropriate exit code
}
```

### 3. Interfaces for Testability

Define interfaces for services to enable mocking:

```go
type ThemeService interface {
    List() ([]Theme, error)
    Get(id string) (*Theme, error)
}
```

### 4. Functional Options Pattern

Used in API client for flexible configuration:

```go
client := api.NewClient(url, storeID, apiKey,
    api.WithOrgID(orgID),
    api.WithChangeSetID(changeSetID),
)
```

### 5. Context for Cancellation

Use `context.Context` for request cancellation and timeouts (to be added):

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.GetWithContext(ctx, "/api/v1/themes")
```

## Key Differences from Ruby Version

### Ruby → Go Translations

| Ruby | Go | Notes |
|------|-----|-------|
| `Thor` | `Cobra` | CLI framework |
| `HTTP` gem | `Resty` | HTTP client |
| `TTY::Spinner` | `briandowns/spinner` | Loading spinners |
| `Pastel` | `fatih/color` | Terminal colors |
| `RSpec` | `testify` + `go test` | Testing |
| `module StoreConnect` | `package` system | Namespacing |
| Classes | Structs + Methods | OOP → composition |
| `attr_reader` | Exported fields | `Field` vs `field` |
| Blocks/Procs | Functions as values | `func() error` |
| Exceptions | Explicit errors | `error` return value |

### Code Style Differences

**Ruby (implicit returns):**
```ruby
def get_theme(id)
  client.get("/api/v1/themes/#{id}")
end
```

**Go (explicit returns):**
```go
func (t *Themes) Get(id string) (*Theme, error) {
    var result Theme
    err := t.client.Get("/api/v1/themes/"+id, &result, nil)
    if err != nil {
        return nil, err
    }
    return &result, nil
}
```

## Development Workflow

### Setup

```bash
cd cli-go
make deps      # Download dependencies
make build     # Build binary to bin/sc
```

### Development Cycle

```bash
make fmt       # Format code (go fmt)
make lint      # Run linters (golangci-lint)
make test      # Run tests
make build     # Build binary
./bin/sc --help
```

### Testing Against Local Rails

```bash
# Terminal 1: Start Rails server
cd ../gem && bin/dev

# Terminal 2: Build and test CLI
cd cli-go
make build
./bin/sc connect http://localhost:3000 --alias local
./bin/sc theme list --server local
./bin/sc theme list --server local --json  # Test JSON mode
```

## Building and Distribution

### Single Platform

```bash
make build     # Build for current platform
make install   # Install to $GOPATH/bin
```

### Multi-Platform

```bash
make build-all
```

Creates binaries for:
- macOS (Intel & ARM)
- Linux (AMD64 & ARM64)
- Windows (AMD64)

### Release Process (Planned)

1. Update version in `internal/commands/version.go`
2. Tag release: `git tag v0.1.0`
3. Use GoReleaser for automated builds
4. Publish to GitHub Releases

## Dependencies

Current dependencies (see `go.mod`):

```
github.com/spf13/cobra        # CLI framework
github.com/spf13/viper        # Configuration
github.com/go-resty/resty/v2  # HTTP client
github.com/fatih/color        # Terminal colors
github.com/briandowns/spinner # Loading spinners
github.com/mitchellh/go-homedir # Home directory
golang.org/x/term             # Terminal input (passwords)
gopkg.in/yaml.v3              # YAML parsing
```

To add a new dependency:

```bash
go get github.com/some/package
make tidy
```

## Testing Strategy (To Be Implemented)

### Unit Tests

```go
func TestNormalizeSalesforceID(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"15 chars", "a0A7Z00000AbCdE", "a0A7Z00000AbCdEFGH", false},
        {"18 chars", "a0A7Z00000AbCdEFGH", "a0A7Z00000AbCdEFGH", false},
        {"invalid", "invalid", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NormalizeSalesforceID(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.expected {
                t.Errorf("got %v, expected %v", got, tt.expected)
            }
        })
    }
}
```

### Integration Tests

Use table-driven tests with httptest:

```go
func TestAPIClient(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer server.Close()

    client := api.NewClient(server.URL, "storeID", "apiKey")
    // Test client methods...
}
```

## TODO: Features to Port from Ruby CLI

- [x] Theme serialization/deserialization (COMPLETE)
- [x] Theme push/preview/publish workflow (COMPLETE)
- [x] ContentChange API integration (COMPLETE)
- [x] Product, Article, Block content management (COMPLETE)
- [x] Category and Trait management (COMPLETE)
- [x] Page management (COMPLETE)
- [x] Media upload (COMPLETE)
- [x] Menu management (COMPLETE)
- [x] Liquid template validation (COMPLETE)
- [x] **JSON output mode (COMPLETE - Milestone 1)**
- [x] **Semantic exit codes (COMPLETE - Milestone 1)**
- [ ] Non-interactive mode (Milestone 2)
- [ ] Watch mode for live development (Milestone 3)
- [ ] Self-documenting help system (Milestone 4)
- [ ] Dry-run and validation (Milestone 5)
- [ ] Progress tracking (Milestone 6)
- [ ] Examples library (Milestone 7)
- [ ] Comprehensive test suite (Milestone 8)
- [ ] CI/CD pipeline (Milestone 8)
- [ ] Release automation (GoReleaser) (Milestone 8)

## Code Style Guidelines

### Naming Conventions

- **Exported** (public): `CapitalCase` - `Client`, `NewClient`, `GetTheme`
- **Unexported** (private): `lowerCamelCase` - `buildURL`, `handleResponse`
- **Constants**: `CapitalCase` or `ALL_CAPS` - `Version`, `DefaultTimeout`, `ExitSuccess`
- **Acronyms**: Keep case - `HTTPClient`, `APIURL`, `SFID`

### File Organization

- One package per directory
- Group related types in same file
- Separate `_test.go` files for tests
- Filename matches primary type: `client.go` contains `Client` struct

### Comments

```go
// Package api provides HTTP client for StoreConnect API
package api

// Client is the HTTP client for StoreConnect API.
// It handles authentication, request/response processing, and error handling.
type Client struct {
    // Exported fields are documented
    BaseURL string

    // Unexported fields use inline comments
    httpClient *resty.Client // underlying HTTP client
}

// NewClient creates a new API client.
// It accepts baseURL, storeSFID, and apiKey as required parameters.
// Optional configuration can be provided using ClientOption functions.
func NewClient(baseURL, storeSFID, apiKey string, opts ...ClientOption) *Client {
    // Implementation
}
```

## Performance Considerations

### Go Advantages

- **Fast compilation**: ~1-2 seconds for full rebuild
- **Small binaries**: ~10-15MB statically linked
- **Low memory**: ~10-20MB runtime memory
- **Fast execution**: No interpreter overhead

### Concurrency (Future)

Use goroutines for parallel operations:

```go
// Download multiple themes concurrently
var wg sync.WaitGroup
for _, themeName := range themes {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        downloadTheme(name)
    }(themeName)
}
wg.Wait()
```

## Cross-Platform Considerations

### File Paths

Always use `filepath.Join()` for cross-platform paths:

```go
// Good
configPath := filepath.Join(home, ".storeconnect", "config.yml")

// Bad (Unix-only)
configPath := home + "/.storeconnect/config.yml"
```

### Home Directory

Use `go-homedir` package:

```go
home, err := homedir.Dir()
```

### Line Endings

Go normalizes line endings on Windows. No special handling needed.

## Common Patterns

### Service Pattern

```go
type ThemesService struct {
    client *Client
}

func NewThemes(client *Client) *Themes {
    return &Themes{client: client}
}

func (t *Themes) List() ([]Theme, error) {
    // Implementation
}
```

### Error Wrapping

```go
if err != nil {
    return fmt.Errorf("failed to connect to server: %w", err)
}
```

### Defer for Cleanup

```go
func processFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close() // Automatically closes when function returns

    // Process file...
    return nil
}
```

## Comparison with Ruby CLI

Both implementations provide identical functionality:
- Same API endpoints
- Same configuration format (YAML)
- Same directory structure
- Same command interface
- **New in Go CLI:** JSON output mode, semantic exit codes

Choose based on:
- **Ruby CLI**: Easier for Ruby developers, dynamic language flexibility
- **Go CLI**: Faster execution, easier distribution, static typing safety, agent-friendly

## Contributing to Go CLI

1. Follow Go best practices and idioms
2. Use `make fmt` before committing
3. Run `make lint` and fix all issues
4. Write tests for new features
5. Update this CLAUDE.md for architectural changes
6. Keep parity with Ruby CLI functionality
7. Test JSON output mode for new commands

## API Integration Details

### Authentication

Bearer token format (enhanced, preferred):
```
Authorization: Bearer {org_id}:{store_sfid}:{api_key}
```

Legacy format (also supported):
```
Authorization: Bearer {store_sfid}:{api_key}
```

### API Endpoints Used by CLI

| Command | Endpoint | Method | Description |
|---------|----------|--------|-------------|
| `sc connect` | `/api/v1/info` | GET | Validate credentials and get server info |
| `sc theme list` | `/api/v1/themes` | GET | List store's custom themes |
| `sc theme pull` | `/api/v1/themes/:id` | GET | Download full theme (by sc_id or sfid) |
| `sc theme new` | `/api/v1/themes` | POST | Create new theme (returns sc_id) |
| `sc theme push` | `/api/v1/content_changes` | POST | Create draft content change |
| `sc theme push` | `/api/v1/content_changes/:id` | PATCH | Add template changes to draft |
| `sc theme preview` | `/api/v1/content_changes/:id/preview_url` | GET | Get preview URL for draft |
| `sc theme publish` | `/api/v1/content_changes/:id/publish` | POST | Publish changes to live site |

### Draft/Preview/Publish Workflow

The push/preview/publish workflow uses a **ContentChange** system to stage template modifications:

**Models:**
- **ContentChange**: Draft container (status: draft → review → published)
- **ContentChangeRecord**: Tracks modified records (action: create/update/delete)
- **ContentChangeField**: Tracks field changes (field_api_name, new_value)

**Workflow:**

1. **Push** - Creates ContentChange with theme_id, sends template changes
2. **Preview** - Returns URL with `?content-change={id}` to preview draft
3. **Publish** - Applies changes permanently (auto in dev/staging, requires approval in prod)

## Agent-Friendly Features (Milestone 1+)

### JSON Output Mode

All commands support `--json` flag for machine-readable output:

```bash
# List themes (human-friendly)
sc theme list

# List themes (JSON)
sc theme list --json | jq -r '.data.themes[].name'

# Status check
sc status --json | jq -r '.data.connected'

# Error handling
if ! sc theme push my-theme --json > /tmp/result.json; then
    jq -r '.suggestion' /tmp/result.json
fi
```

### Semantic Exit Codes

Exit codes indicate error types for automation:

- `0` - Success
- `1` - Generic error
- `2` - Authentication failed (check credentials)
- `3` - Resource not found (verify resource exists)
- `4` - Validation error (check input)
- `5` - Network error (check connection)
- `6` - Resource conflict (check for duplicates)
- `7` - User cancelled (handle cancellation)
- `8` - Configuration error (run `sc status`)

### Error Messages with Suggestions

All errors include helpful suggestions:

```json
{
  "success": false,
  "error": {
    "code": "CONFIG_ERROR",
    "message": "no server configured"
  },
  "suggestion": "Run 'sc status' to check configuration, or 'sc connect' to set up a server"
}
```

## Roadmap

See [AGENT_FRIENDLY_CLI_PLAN.md](AGENT_FRIENDLY_CLI_PLAN.md) for the full 8-milestone roadmap to make CLI fully usable by AI agents.

**Completed:**
- ✅ Milestone 1: JSON Output & Exit Codes (2026-03-17)

**Next:**
- ⏳ Milestone 2: Non-Interactive Mode
- ⏳ Milestone 3: Watch Mode
- ⏳ Milestone 4: Self-Documenting Help
- ⏳ Milestone 5: Dry-Run & Validation
- ⏳ Milestone 6: Progress Tracking
- ⏳ Milestone 7: Examples Library
- ⏳ Milestone 8: Production Hardening

## Testing Against Local Rails

To test CLI against local core-gem instance:

```bash
# Start Rails server
cd ../gem && bin/dev

# Connect CLI to localhost
sc connect http://localhost:3000 --alias local

# Run commands
sc theme list --server local
sc theme push my-theme --server local

# Test JSON mode
sc theme list --server local --json | jq .
sc status --json | jq -r '.data.servers.local.authenticated'
```
