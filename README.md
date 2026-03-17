# StoreConnect CLI (Go)

A command-line interface for developing and managing StoreConnect themes, written in Go.

## Features

- **Local Development**: Edit theme files locally with version control
- **Multi-Server Support**: Work with multiple servers (dev, staging, production)
- **Theme Management**: Push/pull themes between local machine and StoreConnect servers
- **Preview & Validation**: Preview and validate themes before publishing to Salesforce

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/GetStoreConnect/storeconnect-cli.git
cd storeconnect-cli/cli-go

# Install dependencies
make deps

# Build and install
make install
```

The `sc` command will be installed to `$GOPATH/bin`. Make sure it's in your PATH.

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/GetStoreConnect/storeconnect-cli/releases).

## Quick Start

### 1. Initialize a New Project

```bash
sc init my-project
cd my-project
```

This creates:
- `my-project/` - Root project directory
- `.storeconnect/` - Configuration directory (git-safe, no secrets)
- `themes/` - Directory for theme files

### 2. Connect to Your StoreConnect Server

```bash
sc connect https://dev.mystore.com --alias dev
```

You'll be prompted for:
- **Organization ID** (15-18 chars starting with "00D")
- **Store Salesforce ID** (15-18 alphanumeric characters)
- **API Key** (hidden input)

The CLI stores credentials securely in `~/.storeconnect/credentials.yml` with 0600 permissions.

### 3. List Available Themes

```bash
sc theme list
```

### 4. Download a Theme

```bash
sc theme pull my-theme
```

### 5. Create a New Theme

```bash
sc theme new my-new-theme
```

## Commands

### Global Commands

- `sc version` - Show CLI version
- `sc init PROJECT_NAME` - Create new project directory
- `sc connect URL --alias NAME` - Connect to StoreConnect server
- `sc disconnect ALIAS` - Remove server connection
- `sc status` - Show connection status and project info

### Theme Commands

- `sc theme list` - List all themes
- `sc theme pull THEME_NAME` - Download theme from server
- `sc theme new THEME_NAME` - Create a new empty theme
- `sc theme push THEME_NAME` - Upload theme to server (creates draft)
- `sc theme preview THEME_NAME` - Get preview URL for draft
- `sc theme publish THEME_NAME` - Publish draft to live site
- `sc theme delete THEME_NAME` - Delete theme from server
- `sc theme validate THEME_NAME` - Validate theme structure

### Global Flags

- `--server ALIAS` or `-s ALIAS` - Use specific server (defaults to project's default server)
- `--config PATH` - Custom config file path

## Configuration

### Two-Tier Configuration System

StoreConnect CLI separates sensitive credentials from project configuration:

#### Global Credentials (`~/.storeconnect/credentials.yml`)

- Stored in user's home directory
- Contains API keys and Store Salesforce IDs
- **Never committed to git**
- Permissions: 0600 (read/write owner only)

```yaml
servers:
  dev:
    url: https://dev.mystore.com
    org_id: 00D000000000062ABC
    store_sfid: a0A7Z00000AbCdEFGH
    api_key: your-api-key-here
  prod:
    url: https://mystore.com
    org_id: 00D000000000062ABC
    store_sfid: a0A7Z00000XyZ123DEF
    api_key: your-prod-api-key
```

#### Project Config (`.storeconnect/config.yml`)

- Stored in project directory
- Contains server URLs, versions, sync times
- **Safe to commit to git** (no secrets)
- Shared across team members

```yaml
default_server: dev
servers:
  dev:
    url: https://dev.mystore.com
    storeconnect_version: "20.13.0"
    base_theme_version: "1.2.3"
    last_sync: 2024-03-15T10:30:00Z
  prod:
    url: https://mystore.com
    storeconnect_version: "20.13.0"
    base_theme_version: "1.2.3"
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for build tasks)

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to $GOPATH/bin
make install
```

### Testing

```bash
# Run tests
make test

# Run tests with coverage
make coverage
```

### Code Quality

```bash
# Format code
make fmt

# Run linters
make lint

# Tidy dependencies
make tidy
```

## Project Structure

```
cli-go/
├── cmd/
│   └── sc/              # Main entry point
│       └── main.go
├── internal/
│   ├── api/             # API client
│   │   ├── client.go
│   │   ├── auth.go
│   │   └── themes.go
│   ├── commands/        # Command implementations
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── connect.go
│   │   ├── status.go
│   │   └── theme.go
│   ├── config/          # Configuration management
│   │   ├── credentials.go
│   │   └── project.go
│   ├── ui/              # UI helpers
│   │   ├── formatter.go
│   │   └── spinner.go
│   └── utils/           # Utilities
│       └── salesforce_id.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Architecture

### CLI Framework

Uses [Cobra](https://github.com/spf13/cobra) for CLI structure, the same framework used by:
- kubectl
- hugo
- docker
- GitHub CLI

### HTTP Client

Uses [Resty](https://github.com/go-resty/resty) for clean and easy HTTP requests.

### Configuration

Uses [Viper](https://github.com/spf13/viper) for configuration management.

### Terminal UI

Uses:
- [fatih/color](https://github.com/fatih/color) for colored output
- [briandowns/spinner](https://github.com/briandowns/spinner) for loading spinners

## Authentication

Authentication uses Bearer tokens with a composite format:

```
Authorization: Bearer {org_id}:{store_sfid}:{api_key}
```

The client also supports legacy format (without org_id):

```
Authorization: Bearer {store_sfid}:{api_key}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- GitHub Issues: https://github.com/GetStoreConnect/storeconnect-cli/issues
- Documentation: https://docs.storeconnect.com
