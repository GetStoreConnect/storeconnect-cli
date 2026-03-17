# StoreConnect CLI - Quick Reference

## Installation & Setup

```bash
# Build
cd cli-go
make build

# Install to $GOPATH/bin
make install

# Or run directly
./bin/sc --help
```

## Core Commands

```bash
sc version                          # Show version
sc init my-project                  # Create project
sc connect URL --alias dev          # Connect to server
sc disconnect dev                   # Remove connection
sc status                           # Show connections
```

## Theme Commands

```bash
# List & Browse
sc theme list                       # List all themes

# Download
sc theme pull THEME                 # Download theme
sc theme new THEME                  # Create new theme

# Edit & Validate
vim themes/THEME/templates/...      # Edit locally
sc theme validate THEME             # Validate syntax

# Upload & Deploy
sc theme push THEME                 # Upload as draft
sc theme preview THEME              # Get preview URL
sc theme publish THEME              # Publish to live
sc theme delete THEME               # Delete from server
```

## Content Commands

```bash
# Products
sc product list                     # List products

# Articles
sc article list                     # List articles

# Content Blocks
sc block list                       # List content blocks

# Categories
sc category list                    # List categories
sc category tree                    # Show hierarchy

# Pages
sc page list                        # List pages

# Menus
sc menu list                        # List menus
sc menu show MENU_ID                # Show menu structure

# Media
sc media list                       # List media assets
```

## Global Flags

```bash
--server ALIAS or -s ALIAS          # Use specific server
--config PATH                       # Custom config file
--help or -h                        # Show help
```

## File Structure

```
my-project/
├── .storeconnect/
│   └── config.yml                  # Project config (git-safe)
└── themes/
    └── my-theme/
        ├── theme.yml               # Theme metadata
        ├── templates/              # Liquid templates
        │   ├── pages/
        │   ├── snippets/
        │   └── layouts/
        ├── variables.json          # Theme variables
        ├── assets.json             # Asset URLs
        └── .sync-state.yml         # Sync tracking
```

## Configuration

**Global credentials** (`~/.storeconnect/credentials.yml`):
- API keys and Store IDs
- Never commit to git
- Permissions: 0600

**Project config** (`.storeconnect/config.yml`):
- Server URLs and metadata
- Safe to commit to git
- Permissions: 0644

## Typical Workflow

```bash
# 1. Setup
sc init my-store && cd my-store
sc connect https://dev.mystore.com --alias dev

# 2. Download theme
sc theme pull my-theme

# 3. Edit locally
vim themes/my-theme/templates/pages/home.liquid

# 4. Validate
sc theme validate my-theme

# 5. Deploy
sc theme push my-theme
sc theme preview my-theme           # Check preview URL
sc theme publish my-theme           # Go live
```

## Multi-Environment Deployment

```bash
# Connect to all environments
sc connect https://dev.mystore.com --alias dev
sc connect https://staging.mystore.com --alias staging
sc connect https://mystore.com --alias prod

# Deploy to dev
sc theme push my-theme --server dev
sc theme publish my-theme --server dev

# Promote to staging
sc theme push my-theme --server staging
sc theme publish my-theme --server staging

# Promote to production
sc theme push my-theme --server prod
sc theme publish my-theme --server prod
```

## Troubleshooting

```bash
# Check connection
sc status

# Verify credentials
ls -la ~/.storeconnect/credentials.yml

# Reconnect
sc disconnect dev
sc connect https://dev.mystore.com --alias dev

# Fix permissions
chmod 600 ~/.storeconnect/credentials.yml

# Validate theme
sc theme validate my-theme

# Check theme exists
ls -la themes/my-theme
```

## Development

```bash
# Build
make build

# Test
make test

# Format code
make fmt

# Lint
make lint

# Clean
make clean

# Build for all platforms
make build-all
```

## Help

```bash
# General help
sc --help

# Command help
sc theme --help
sc theme push --help

# All commands
sc --help | grep "^  "
```

## Example Output

```bash
$ sc theme push my-theme
✓ Reading theme 'my-theme'
✓ Creating draft
✓ Uploading templates
✓ Pushed theme 'my-theme' as draft

Content Change ID: cc_abc123
Run 'sc theme preview my-theme' to get preview URL
Run 'sc theme publish my-theme' to publish to live site
```

```bash
$ sc theme preview my-theme
✓ Preview URL generated

Preview URL:
https://dev.mystore.com?theme-preview=theme_xyz&content-change=cc_abc123

Open this URL in your browser to preview the theme
Run 'sc theme publish my-theme' when ready to publish
```

```bash
$ sc category tree
ℹ Category tree on dev:

├─ Electronics
│  ├─ Computers
│  │  ├─ Laptops
│  │  └─ Desktops
│  └─ Phones
└─ Clothing
   ├─ Men
   └─ Women
```

## Performance

- **Startup:** ~10ms (20x faster than Ruby)
- **Binary size:** 11 MB (self-contained)
- **Memory:** ~15 MB working set
- **Build time:** ~2 seconds

## Links

- **Documentation:** See README.md
- **Architecture:** See CLAUDE.md
- **Tutorial:** See GETTING_STARTED.md
- **Full feature list:** See IMPLEMENTATION_COMPLETE.md
