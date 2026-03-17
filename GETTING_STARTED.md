# Getting Started with StoreConnect CLI (Go)

This guide will help you get started with the Go version of the StoreConnect CLI.

## Prerequisites

- **Go 1.21 or later**: [Download Go](https://golang.org/dl/)
- **Make** (optional): For build automation
- **Git**: For version control

## Installation

### Option 1: Build from Source

```bash
# Navigate to the Go CLI directory
cd cli-go

# Download dependencies
make deps

# Build the binary
make build

# The binary is now available at: bin/sc
./bin/sc --help

# Optional: Install to $GOPATH/bin (adds to PATH)
make install
sc --help
```

### Option 2: Direct Go Install

```bash
cd cli-go
go install ./cmd/sc
sc --help
```

## Quick Start Tutorial

### 1. Initialize a New Project

```bash
# Create a new project
sc init my-store-project

# Navigate into the project
cd my-store-project
```

This creates:
```
my-store-project/
├── .storeconnect/
│   └── config.yml        # Project config (safe to commit)
└── themes/               # Theme files go here
```

### 2. Connect to Your StoreConnect Server

```bash
sc connect https://dev.mystore.com --alias dev
```

You'll be prompted for:

**Organization ID:**
- Find: Salesforce Setup → Company Information
- Format: 15-18 characters starting with "00D"
- Example: `00D000000000062ABC`

**Store Salesforce ID:**
- Find: Your Store record URL in Salesforce
- Format: 15-18 alphanumeric characters
- Example: `a0A7Z00000AbCdE`

**API Key:**
- Generate: Store record → API Keys → New API Key
- Security: Input is hidden for security
- Store securely: Don't share or commit to git

### 3. Verify Connection

```bash
sc status
```

Output shows:
```
ℹ Project Configuration

* dev
  URL: https://dev.mystore.com
  Version: 20.13.0
  Base Theme: 1.2.3

* = default server

Credentials stored in: /Users/you/.storeconnect/credentials.yml
```

### 4. List Available Themes

```bash
sc theme list
```

### 5. Download a Theme

```bash
sc theme pull my-theme
```

This downloads the theme to `themes/my-theme/`:
```
themes/my-theme/
├── theme.yml
├── templates/
│   ├── pages/
│   ├── snippets/
│   └── layouts/
├── assets.json
├── variables.json
└── locales/
```

### 6. Create a New Theme

```bash
sc theme new my-new-theme
```

### 7. Work with Multiple Servers

```bash
# Connect to staging
sc connect https://staging.mystore.com --alias staging

# Connect to production
sc connect https://mystore.com --alias prod

# Use specific server
sc theme list --server staging
sc theme pull my-theme --server prod

# Set default server
sc config set-default-server prod
```

## Common Workflows

### Local Theme Development

```bash
# 1. Pull theme from server
sc theme pull my-theme

# 2. Edit files in themes/my-theme/
# 3. Push changes (creates draft)
sc theme push my-theme

# 4. Preview changes
sc theme preview my-theme

# 5. Publish to live site
sc theme publish my-theme
```

### Multi-Environment Development

```bash
# Develop on dev server
sc theme pull my-theme --server dev
# ... make changes ...
sc theme push my-theme --server dev
sc theme preview my-theme --server dev

# Promote to staging
sc theme push my-theme --server staging
sc theme publish my-theme --server staging

# Promote to production
sc theme push my-theme --server prod
sc theme publish my-theme --server prod
```

## Configuration Files

### Global Credentials (~/.storeconnect/credentials.yml)

**Location:** Your home directory
**Permissions:** 0600 (read/write owner only)
**Contains:** API keys and Store IDs
**Security:** NEVER commit to git

```yaml
servers:
  dev:
    url: https://dev.mystore.com
    org_id: 00D000000000062ABC
    store_sfid: a0A7Z00000AbCdE
    api_key: your-dev-api-key-here
  prod:
    url: https://mystore.com
    org_id: 00D000000000062ABC
    store_sfid: a0A7Z00000XyZ123
    api_key: your-prod-api-key-here
```

### Project Config (.storeconnect/config.yml)

**Location:** Project directory
**Permissions:** 0644 (readable by all)
**Contains:** Server URLs and metadata
**Security:** SAFE to commit to git

```yaml
default_server: dev
servers:
  dev:
    url: https://dev.mystore.com
    storeconnect_version: "20.13.0"
    base_theme_version: "1.2.3"
    last_sync: 2024-03-15T10:30:00Z
```

## Available Commands

### Global Commands
- `sc version` - Show CLI version
- `sc init PROJECT` - Create new project
- `sc connect URL --alias NAME` - Connect to server
- `sc disconnect ALIAS` - Remove server
- `sc status` - Show connection status

### Theme Commands
- `sc theme list` - List themes
- `sc theme pull NAME` - Download theme
- `sc theme new NAME` - Create new theme
- `sc theme push NAME` - Upload theme (draft)
- `sc theme preview NAME` - Get preview URL
- `sc theme publish NAME` - Publish to live
- `sc theme delete NAME` - Delete theme
- `sc theme validate NAME` - Validate structure

### Global Flags
- `--server ALIAS` or `-s ALIAS` - Use specific server
- `--config PATH` - Custom config file

## Troubleshooting

### Authentication Failed

```bash
# Verify your credentials
sc status

# Check that credentials file exists
ls -la ~/.storeconnect/credentials.yml

# Reconnect to server
sc disconnect dev
sc connect https://dev.mystore.com --alias dev
```

### "Not in a StoreConnect project"

```bash
# Initialize project
sc init my-project
cd my-project

# Or if already in project directory
ls -la .storeconnect/config.yml
```

### Permission Denied (credentials.yml)

```bash
# Fix permissions
chmod 600 ~/.storeconnect/credentials.yml
```

## Getting Help

### CLI Help

```bash
# General help
sc --help

# Command-specific help
sc theme --help
sc theme push --help
```

### Documentation

- **README.md** - Installation and overview
- **CLAUDE.md** - Architecture and development guide
- **API Documentation** - See `/gem/CLAUDE.md` for API endpoints

### Support

- GitHub Issues: https://github.com/GetStoreConnect/storeconnect-cli/issues
- Documentation: https://docs.storeconnect.com

## Next Steps

1. Read the [README.md](README.md) for full feature list
2. Check [CLAUDE.md](CLAUDE.md) for architecture details
3. Explore theme structure in `themes/` directory
4. Review API documentation in `/gem/CLAUDE.md`
5. Set up your development workflow

## Development vs Production

**Development Mode:**
- Auto-publishes changes
- Faster iteration
- Safe to experiment

**Production Mode:**
- Changes require review
- Manual approval in Salesforce
- Safer deployment

Set environment appropriately when connecting!
