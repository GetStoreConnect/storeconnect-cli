# Theme Inheritance & Content Change Sets - Implementation Plan

**Date:** 2026-03-17
**Status:** Proposal
**Priority:** Critical (Security + Architecture)

## Executive Summary

This document proposes a comprehensive redesign of the StoreConnect theme and content change system to address three critical issues:

1. **Security**: Prevent CLI from bypassing the Salesforce approval flow
2. **Developer Experience**: Group multiple edits into single content change sessions
3. **Theme Management**: Implement theme inheritance for easier versioning and upgrades

## Problem Statement

### 1. Security Issue: Bypassing Approval Flow

**Current State:**
```
CLI → Heroku (live site) → Changes appear immediately ❌
      ↓
      Content Change → Salesforce → Approval (too late)
```

**Problem:** Changes are live on the site before Salesforce approval.

**Risk:**
- Unapproved content goes live
- No audit trail
- Bypasses governance
- Violates compliance requirements

### 2. Content Change Granularity

**Current State:**
- Each `sc theme push` creates a new ContentChange record
- 100 edits = 100 ContentChange records in Salesforce
- No concept of a "session" or "batch"
- Difficult to review changes holistically

**Problem:**
- Cluttered approval queue
- Hard to understand what changed
- No rollback to session start
- Poor developer experience

### 3. Theme Management

**Current State:**
- Base theme is hardcoded data in gem
- No versioning
- No inheritance
- Upgrades require manual migration
- Can't test new base theme version safely

**Problem:**
- Difficult to upgrade base theme
- Can't revert to previous version
- No incremental theme development
- Must duplicate entire theme to customize

## Proposed Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    DEVELOPER WORKFLOW                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. sc theme new "My Theme" --inherit "Base Theme v21"      │
│     ↓                                                        │
│  2. Edit files locally (templates/assets/variables)         │
│     ↓                                                        │
│  3. sc theme push "My Theme"                                │
│     → Creates/Updates ContentChangeSession (draft)          │
│     → All changes grouped in session                        │
│     ↓                                                        │
│  4. sc theme preview "My Theme"                             │
│     → Preview URL shows cumulative session changes          │
│     → NOT live on site                                      │
│     ↓                                                        │
│  5. sc theme publish "My Theme"                             │
│     → Sends ENTIRE session to Salesforce as one batch       │
│     → Creates ContentChange record in Salesforce            │
│     ↓                                                        │
│  6. SALESFORCE APPROVAL FLOW                                │
│     → Review all changes in session                         │
│     → Approve/Reject                                        │
│     ↓                                                        │
│  7. SC-SYNC PUBLISHES (only after approval)                 │
│     → sc-sync detects approved ContentChange                │
│     → Publishes to live site                                │
│     → Marks session as published                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **CLI NEVER touches live site directly**
2. **All changes go through ContentChangeSession → ContentChange → Approval → sc-sync**
3. **Preview is isolated (query param or subdomain)**
4. **Sessions group related edits**
5. **Themes can inherit from other themes**
6. **Base theme is versioned and upgradeable**

## Implementation Details

### Part 1: Content Change Sessions

#### 1.1 New Data Model

**Salesforce Objects:**

```apex
// New: ContentChangeSession (replaces multiple ContentChanges)
public class ContentChangeSession__c {
    Id Id;
    String Name;                          // Auto: "Session #1234"
    String Theme__c;                      // FK to Theme__c (sc_id)
    String Status__c;                     // draft, submitted, approved, rejected, published
    String CreatedBy__c;                  // Developer email/username
    DateTime CreatedDate;
    DateTime SubmittedDate__c;
    DateTime ApprovedDate__c;
    DateTime PublishedDate__c;
    String ApprovedBy__c;
    String RejectionReason__c;
    Integer ChangeCount__c;               // Number of changes in session

    // Session metadata
    String SessionId__c;                  // UUID from CLI
    String CLIVersion__c;                 // For diagnostics
    Text SessionNotes__c;                 // Developer notes
}

// Modified: ContentChangeRecord (now references session)
public class ContentChangeRecord__c {
    Id Id;
    String ContentChangeSession__c;       // FK to session (was ContentChange__c)
    String RecordType__c;                 // Theme, Template, Asset, etc.
    String RecordId__c;                   // SC ID of modified record
    String Action__c;                     // create, update, delete
    String FieldChanges__c;              // JSON of field changes
    Integer SequenceNumber__c;            // Order of changes in session
}

// Modified: ContentChangeField (now references record which references session)
public class ContentChangeField__c {
    Id Id;
    String ContentChangeRecord__c;        // FK to ContentChangeRecord__c
    String FieldAPIName__c;
    String OldValue__c;
    String NewValue__c;
}
```

**PostgreSQL (gem):**

```ruby
# New table
create_table :content_change_sessions do |t|
  t.string :sc_id, null: false, index: {unique: true}
  t.string :sfid, index: true
  t.references :theme, foreign_key: true, null: false
  t.string :status, default: 'draft'  # draft, submitted, approved, rejected, published
  t.string :session_id, null: false   # UUID from CLI
  t.text :session_notes
  t.integer :change_count, default: 0
  t.datetime :submitted_at
  t.datetime :approved_at
  t.datetime :published_at
  t.string :approved_by
  t.text :rejection_reason
  t.timestamps
end

# Modified: content_change_records
add_column :content_change_records, :content_change_session_id, :bigint
add_column :content_change_records, :sequence_number, :integer
add_index :content_change_records, :content_change_session_id
# Remove old content_change_id column after migration
```

#### 1.2 Session Lifecycle

**States:**

1. **draft** - CLI is making changes (can preview)
2. **submitted** - Developer published (sent to Salesforce)
3. **approved** - Salesforce admin approved (ready for sc-sync)
4. **rejected** - Salesforce admin rejected (developer must fix)
5. **published** - sc-sync published to live site (immutable)

**Transitions:**

```
draft → submitted    (sc theme publish)
submitted → approved (Salesforce admin action)
submitted → rejected (Salesforce admin action)
approved → published (sc-sync background job)
rejected → draft     (sc theme reopen - for fixes)
```

#### 1.3 CLI Workflow

**Session Management:**

```bash
# Start new session (automatic on first push)
sc theme push "My Theme"
# → Creates ContentChangeSession in draft
# → Tracks session ID in .storeconnect/sessions/my-theme.yml

# Continue working (adds to same session)
sc theme push "My Theme"
# → Updates existing draft session
# → Cumulative changes

# Preview cumulative changes
sc theme preview "My Theme"
# → Returns: https://mystore.com/?session=SESSION_UUID
# → Shows ALL changes in current draft session

# Submit session for approval
sc theme publish "My Theme"
# → Changes session status: draft → submitted
# → Sends to Salesforce as ONE ContentChange
# → CLI cannot modify this session anymore

# Check session status
sc theme status "My Theme"
# → Shows: draft (N changes), submitted (pending), approved, rejected, published

# Handle rejection
sc theme reopen "My Theme"
# → Changes status: rejected → draft
# → Developer can fix and re-submit
```

**Session Storage (.storeconnect/sessions/):**

```yaml
# .storeconnect/sessions/my-theme.yml
session_id: "550e8400-e29b-41d4-a716-446655440000"
theme_id: "theme-sc-id"
theme_name: "My Theme"
status: draft
change_count: 15
created_at: 2026-03-17T10:30:00Z
last_push_at: 2026-03-17T14:22:00Z
submitted_at: null
notes: "Adding new product page template"
```

### Part 2: Theme Inheritance

#### 2.1 Data Model Changes

**Salesforce:**

```apex
// Modified: Theme__c
public class Theme__c {
    Id Id;
    String Name;                          // "My Custom Theme"
    String SCID__c;                       // sc_abc123

    // NEW: Inheritance
    String ParentTheme__c;                // FK to Theme__c (sc_id)
    String BaseThemeVersion__c;           // "21.0.0" (denormalized for quick lookup)
    Boolean IsBaseTheme__c;               // true for official base themes
    Integer InheritanceDepth__c;          // 0=base, 1=child of base, 2=grandchild

    // Existing fields
    Text Variables__c;                    // JSON (only overrides if has parent)
    // ... templates are in separate table
}

// New: ThemeTemplate__c (extract from Theme__c)
public class ThemeTemplate__c {
    Id Id;
    String Theme__c;                      // FK to Theme__c
    String TemplateKey__c;                // "pages/home", "layouts/default"
    Text Content__c;                      // Liquid template content
    String ContentHash__c;                // SHA256 for change detection
    Boolean OverridesParent__c;           // true if parent has same template
}

// New: ThemeAsset__c (extract from Theme__c)
public class ThemeAsset__c {
    Id Id;
    String Theme__c;                      // FK to Theme__c
    String Filename__c;                   // "css/main.css"
    String ContentType__c;                // "text/css"
    String URL__c;                        // Cloudinary/S3 URL
    String ContentHash__c;                // SHA256
    Boolean OverridesParent__c;           // true if parent has same asset
}
```

**PostgreSQL:**

```ruby
# Modified: themes table
add_column :themes, :parent_theme_id, :bigint
add_column :themes, :base_theme_version, :string
add_column :themes, :is_base_theme, :boolean, default: false
add_column :themes, :inheritance_depth, :integer, default: 0
add_index :themes, :parent_theme_id
add_foreign_key :themes, :themes, column: :parent_theme_id

# New: theme_templates table
create_table :theme_templates do |t|
  t.references :theme, foreign_key: true, null: false
  t.string :template_key, null: false      # pages/home
  t.text :content, null: false             # Liquid
  t.string :content_hash                   # SHA256
  t.boolean :overrides_parent, default: false
  t.timestamps

  t.index [:theme_id, :template_key], unique: true
end

# New: theme_assets table
create_table :theme_assets do |t|
  t.references :theme, foreign_key: true, null: false
  t.string :filename, null: false          # css/main.css
  t.string :content_type
  t.string :url                            # External storage
  t.string :content_hash
  t.boolean :overrides_parent, default: false
  t.timestamps

  t.index [:theme_id, :filename], unique: true
end
```

#### 2.2 Theme Resolution Algorithm

When rendering a page with inherited theme:

```ruby
class ThemeRenderer
  def resolve_template(theme, template_key)
    # 1. Check if theme has this template
    template = theme.templates.find_by(template_key: template_key)
    return template.content if template

    # 2. Check parent theme (recursive)
    if theme.parent_theme
      return resolve_template(theme.parent_theme, template_key)
    end

    # 3. Not found in chain
    raise TemplateNotFound, "#{template_key} not found in theme hierarchy"
  end

  def resolve_variables(theme)
    # Merge variables up the chain (child overrides parent)
    variables = {}

    # Walk up the chain
    current = theme
    chain = []
    while current
      chain << current
      current = current.parent_theme
    end

    # Apply from base to child (so child overrides)
    chain.reverse.each do |t|
      variables.merge!(JSON.parse(t.variables || '{}'))
    end

    variables
  end
end
```

#### 2.3 CLI Commands

```bash
# Create theme inheriting from base
sc theme new "My Theme" --inherit "Base Theme v21"
# → Creates theme with parent_theme_id = base_theme_sc_id
# → Starts with empty templates (inherits all from parent)

# Create standalone theme (no inheritance)
sc theme new "My Theme" --no-inherit
# → Creates theme with parent_theme_id = null
# → Must provide all templates

# Override a parent template
# Edit: themes/my-theme/templates/pages/home.liquid
sc theme push "My Theme"
# → Creates template override
# → Sets overrides_parent = true

# Change parent theme (upgrade)
sc theme set-parent "My Theme" "Base Theme v22"
# → Updates parent_theme_id
# → Creates ContentChangeSession for review
# → Shows diff of what will change

# View inheritance chain
sc theme info "My Theme"
# Output:
#   My Theme (sc_xyz789)
#   └─ Base Theme v21 (sc_base21)
#      └─ (root)
#
#   Overrides: 3 templates, 2 assets
#   Inherited: 47 templates, 15 assets

# Show what's overridden
sc theme diff "My Theme"
# Output:
#   Overridden Templates:
#     ✓ pages/home.liquid
#     ✓ layouts/default.liquid
#     ✓ snippets/header.liquid
#
#   Inherited Templates: 47 (use --show-inherited to list)
```

### Part 3: Base Theme Installation

#### 3.1 Base Theme Bootstrap Process

**First Installation (Package Install/Upgrade):**

```apex
// Apex class: BaseThemeInstaller
public class BaseThemeInstaller {

    public static void installBaseTheme(String version) {
        // Check if this version already exists
        Theme__c existing = [
            SELECT Id FROM Theme__c
            WHERE IsBaseTheme__c = true
            AND BaseThemeVersion__c = :version
            LIMIT 1
        ];

        if (existing != null) {
            System.debug('Base Theme v' + version + ' already installed');
            return;
        }

        // Create base theme record
        Theme__c baseTheme = new Theme__c(
            Name = 'Base Theme v' + version,
            SCID__c = 's_c__base_theme_' + version.replace('.', '_'),
            IsBaseTheme__c = true,
            BaseThemeVersion__c = version,
            ParentTheme__c = null,
            InheritanceDepth__c = 0,
            Variables__c = getDefaultVariables()
        );
        insert baseTheme;

        // Request templates/assets from gem via API
        requestBaseThemeAssets(baseTheme.SCID__c, version);
    }

    @future(callout=true)
    private static void requestBaseThemeAssets(String themeScId, String version) {
        // Call gem API to send base theme content
        HttpRequest req = new HttpRequest();
        req.setEndpoint(getGemApiUrl() + '/api/v1/base_theme/install');
        req.setMethod('POST');
        req.setHeader('Authorization', 'Bearer ' + getInternalApiKey());
        req.setBody(JSON.serialize(new Map<String, String>{
            'theme_sc_id' => themeScId,
            'version' => version
        }));

        Http http = new Http();
        HttpResponse res = http.send(req);

        if (res.getStatusCode() != 200) {
            throw new InstallException('Failed to install base theme: ' + res.getBody());
        }
    }
}

// Post-install script
global class PostInstallScript implements InstallHandler {
    global void onInstall(InstallContext context) {
        // Get package version
        String version = getPackageVersion(); // "21.0.0"

        // Install base theme
        BaseThemeInstaller.installBaseTheme(version);
    }
}
```

**Gem API Endpoint:**

```ruby
# app/controllers/api/v1/base_theme_controller.rb
class Api::V1::BaseThemeController < Api::V1::VersionController

  # POST /api/v1/base_theme/install
  def install
    theme_sc_id = params[:theme_sc_id]
    version = params[:version]

    # Find or create theme in gem
    theme = Theme.find_or_create_by(sc_id: theme_sc_id) do |t|
      t.name = "Base Theme v#{version}"
      t.is_base_theme = true
      t.base_theme_version = version
      t.parent_theme_id = nil
      t.inheritance_depth = 0
    end

    # Load base theme from gem assets
    base_theme_data = BaseTheme.load_from_assets(version)

    # Create templates
    base_theme_data[:templates].each do |template_data|
      ThemeTemplate.find_or_create_by(
        theme: theme,
        template_key: template_data[:key]
      ) do |t|
        t.content = template_data[:content]
        t.content_hash = Digest::SHA256.hexdigest(template_data[:content])
        t.overrides_parent = false
      end
    end

    # Create assets
    base_theme_data[:assets].each do |asset_data|
      ThemeAsset.find_or_create_by(
        theme: theme,
        filename: asset_data[:filename]
      ) do |a|
        a.content_type = asset_data[:content_type]
        a.url = upload_to_cloudinary(asset_data[:content])
        a.content_hash = Digest::SHA256.hexdigest(asset_data[:content])
        a.overrides_parent = false
      end
    end

    render json: {success: true, theme_id: theme.sc_id}
  end
end

# lib/base_theme.rb
class BaseTheme
  def self.load_from_assets(version)
    # Load from gem's packaged base theme
    base_path = Rails.root.join('app', 'themes', 'base', version)

    {
      templates: load_templates(base_path),
      assets: load_assets(base_path),
      variables: load_variables(base_path)
    }
  end

  private

  def self.load_templates(base_path)
    Dir.glob(base_path.join('templates', '**', '*')).map do |file|
      next if File.directory?(file)

      key = file.sub(base_path.join('templates').to_s + '/', '')
      {
        key: key,
        content: File.read(file)
      }
    end.compact
  end

  # ... similar for assets and variables
end
```

#### 3.2 Base Theme Versioning

**Version Naming:**

```
Base Theme v21.0.0  (major.minor.patch)
Base Theme v21.1.0
Base Theme v22.0.0

Format: "Base Theme v{package_major_version}.{theme_minor}.{theme_patch}"
```

**Upgrade Strategy:**

1. **Package upgrade installs new base theme**
   - v21.0.0 → v22.0.0 package upgrade
   - Post-install creates "Base Theme v22" record
   - Old "Base Theme v21" remains (for backward compatibility)

2. **Customers can upgrade incrementally**
   ```bash
   # Safe upgrade testing
   sc theme set-parent "My Theme" "Base Theme v22" --preview
   # → Shows what will change
   # → Preview URL to test

   # If good, publish
   sc theme set-parent "My Theme" "Base Theme v22" --publish
   # → Goes through approval flow

   # If bad, revert
   sc theme set-parent "My Theme" "Base Theme v21"
   ```

3. **Deprecation Timeline**
   - v21 supported for 12 months after v22 release
   - Customers must upgrade before EOL
   - Clear migration guides

### Part 4: Security & Approval Flow

#### 4.1 Heroku Connect Configuration

**CRITICAL: Read-Only Access for Themes**

```ruby
# config/initializers/heroku_connect.rb

# Themes should ONLY be modified via ContentChangeSessions
# Direct writes to theme tables are BLOCKED

ActiveRecord::Base.connection.execute(<<-SQL)
  -- Revoke direct write access
  REVOKE UPDATE, DELETE ON themes FROM heroku_connect_user;
  REVOKE UPDATE, DELETE ON theme_templates FROM heroku_connect_user;
  REVOKE UPDATE, DELETE ON theme_assets FROM heroku_connect_user;

  -- Heroku Connect can still INSERT (for new themes from Salesforce)
  -- But updates MUST go through ContentChangeSession flow
SQL

# App-level enforcement
class Theme < ApplicationRecord
  before_update :prevent_direct_updates

  private

  def prevent_direct_updates
    unless ContentChangeSession.publishing_context?
      raise SecurityError, "Themes can only be updated via ContentChangeSession"
    end
  end
end
```

#### 4.2 Preview Mode Implementation

**Query Parameter Approach:**

```ruby
# app/controllers/application_controller.rb
class ApplicationController < ActionController::Base
  before_action :load_theme_with_session

  private

  def load_theme_with_session
    @theme = current_store.active_theme

    # Check for preview session
    if params[:session]
      session = ContentChangeSession.find_by(session_id: params[:session])

      if session&.draft? || session&.submitted?
        # Overlay session changes on theme
        @theme = ThemeWithSession.new(@theme, session)
      end
    end
  end
end

# app/models/theme_with_session.rb
class ThemeWithSession
  delegate_missing_to :@base_theme

  def initialize(base_theme, session)
    @base_theme = base_theme
    @session = session
  end

  def template(key)
    # Check session for override
    change = @session.template_changes.find_by(template_key: key)
    return change.new_content if change

    # Fall back to base theme (with inheritance)
    @base_theme.template(key)
  end

  def variables
    # Merge session variable changes
    base_vars = @base_theme.variables
    session_vars = @session.variable_changes || {}
    base_vars.merge(session_vars)
  end
end
```

#### 4.3 sc-sync Publishing Process

**Only sc-sync can publish approved changes:**

```ruby
# app/jobs/content_change_publisher_job.rb
class ContentChangePublisherJob < ApplicationJob
  queue_as :critical

  def perform
    # Find approved sessions ready to publish
    ContentChangeSession.where(status: 'approved').find_each do |session|
      publish_session(session)
    end
  end

  private

  def publish_session(session)
    ContentChangeSession.with_publishing_context do
      ActiveRecord::Base.transaction do
        theme = session.theme

        # Apply all changes in session
        session.change_records.order(:sequence_number).each do |change|
          case change.record_type
          when 'template'
            apply_template_change(theme, change)
          when 'asset'
            apply_asset_change(theme, change)
          when 'variable'
            apply_variable_change(theme, change)
          when 'parent_theme'
            apply_parent_change(theme, change)
          end
        end

        # Mark session as published
        session.update!(
          status: 'published',
          published_at: Time.current
        )

        # Clear cache
        Rails.cache.delete([:theme, theme.sc_id])

        # Sync back to Salesforce
        sync_to_salesforce(session)
      end
    end
  rescue => e
    session.update(
      status: 'failed',
      rejection_reason: e.message
    )
    raise
  end

  def apply_template_change(theme, change)
    case change.action
    when 'create', 'update'
      template = theme.templates.find_or_initialize_by(
        template_key: change.field_changes['template_key']
      )
      template.update!(
        content: change.field_changes['content'],
        content_hash: Digest::SHA256.hexdigest(change.field_changes['content'])
      )
    when 'delete'
      theme.templates.find_by(template_key: change.record_id)&.destroy
    end
  end
end
```

### Part 5: CLI Implementation

#### 5.1 New Commands

```bash
# Session management
sc theme session status "My Theme"
sc theme session list
sc theme session reopen "My Theme"
sc theme session discard "My Theme"

# Theme inheritance
sc theme new "Theme Name" [--inherit "Parent Theme"] [--no-inherit]
sc theme set-parent "Theme Name" "New Parent" [--preview] [--publish]
sc theme info "Theme Name"
sc theme diff "Theme Name" [--show-inherited]

# Enhanced preview
sc theme preview "Theme Name" [--open]
# → Returns preview URL with session parameter
# → Optionally opens in browser
```

#### 5.2 Modified Workflow

```bash
# 1. Create new theme (inheriting from base)
sc theme new "My Custom Theme" --inherit "Base Theme v21"

# 2. Pull to local (gets parent templates for reference)
sc theme pull "My Custom Theme"
# Creates:
#   themes/my-custom-theme/
#     theme.yml              (metadata, parent reference)
#     templates/             (only overrides, empty at first)
#     assets/                (only overrides, empty at first)
#     variables.json         (only overrides)
#   themes/_inherited/
#     base-theme-v21/        (read-only reference)
#       templates/           (all parent templates)
#       assets/
#       variables.json

# 3. Edit - create override
cp themes/_inherited/base-theme-v21/templates/pages/home.liquid \
   themes/my-custom-theme/templates/pages/home.liquid
# Edit the file...

# 4. Push changes (starts session)
sc theme push "My Custom Theme"
# Output:
#   ✓ Created session: abc-123-def-456
#   ✓ Added 1 template override
#   ✓ Session has 1 change
#
#   Preview: https://mystore.com/?session=abc-123-def-456
#   Publish when ready: sc theme publish "My Custom Theme"

# 5. Make more changes
# Edit more files...
sc theme push "My Custom Theme"
# Output:
#   ✓ Updated session: abc-123-def-456
#   ✓ Added 2 template overrides
#   ✓ Session has 3 changes total
#
#   Preview: https://mystore.com/?session=abc-123-def-456

# 6. Preview all changes
sc theme preview "My Custom Theme" --open
# Opens browser to preview URL

# 7. Publish (submit for approval)
sc theme publish "My Custom Theme"
# Output:
#   ✓ Session submitted for approval
#   ✓ Session ID: abc-123-def-456
#   ✓ Changes: 3 templates
#
#   Status: Pending approval in Salesforce
#
#   Check status: sc theme session status "My Custom Theme"
#
#   Next: Ask Salesforce admin to approve the ContentChange

# 8. Check status
sc theme session status "My Custom Theme"
# Output:
#   Session: abc-123-def-456
#   Status: submitted
#   Submitted: 2026-03-17 14:30:00
#   Changes: 3 templates
#
#   Waiting for Salesforce approval...

# Later, after approval...
sc theme session status "My Custom Theme"
# Output:
#   Session: abc-123-def-456
#   Status: approved
#   Approved: 2026-03-17 15:45:00
#   Approved by: admin@example.com
#
#   sc-sync will publish to live site within 5 minutes

# Even later, after sc-sync runs...
sc theme session status "My Custom Theme"
# Output:
#   Session: abc-123-def-456
#   Status: published
#   Published: 2026-03-17 15:47:23
#
#   Changes are now live!
```

## Implementation Phases

### Phase 1: Security Fix (CRITICAL - Week 1)

**Priority:** Immediate

**Tasks:**
1. Add database constraints preventing direct theme updates
2. Implement ContentChangeSession model
3. Update CLI to use sessions
4. Implement preview mode with query params
5. Update sc-sync to publish only approved sessions

**Success Criteria:**
- CLI cannot modify live site directly
- All changes go through approval
- Preview works without affecting live

### Phase 2: Theme Inheritance (Week 2-3)

**Tasks:**
1. Add parent_theme_id to themes table
2. Implement theme resolution algorithm
3. Create ThemeTemplate and ThemeAsset tables
4. Update CLI to support inheritance
5. Implement template override detection

**Success Criteria:**
- Themes can inherit from other themes
- Overrides work correctly
- Resolution algorithm tested

### Phase 3: Base Theme Installation (Week 3-4)

**Tasks:**
1. Package base theme templates/assets
2. Create post-install script
3. Implement gem API endpoint
4. Create base theme installer
5. Test installation process

**Success Criteria:**
- Base theme installed on package install
- Versioned correctly
- Available for inheritance

### Phase 4: Upgrade & Migration (Week 4-5)

**Tasks:**
1. Implement hybrid base theme installation (Approach 4)
   - Static resource installation (primary)
   - API callout fallback
   - CLI manual installation (last resort)
2. Implement parent theme switching with preview
3. Create migration tools and conflict detection
4. Build diff/preview for upgrades
5. Write upgrade documentation
6. Test upgrade paths

**Success Criteria:**
- Base theme auto-installs on package upgrade
- Customers can preview base theme upgrades
- Switch parent theme with approval flow
- Conflict detection between overrides and new parent
- Rollback capability if issues found
- Clear upgrade path documentation

**Base Theme Installation (Hybrid Approach)**:

When StoreConnect package upgrades from v21 to v22:

**Method 1: Static Resource (Primary)**
```apex
// Post-install script tries static resource first
global class StoreConnectPostInstall implements InstallHandler {
    global void onInstall(InstallContext context) {
        String version = getTargetVersion(context);
        Boolean installed = false;

        // Try static resource (fast, reliable, offline)
        try {
            installed = installFromStaticResource(version);
            if (installed) {
                System.debug('✓ Installed from Static Resource');
                return;
            }
        } catch (Exception e) {
            System.debug('Static Resource install failed: ' + e.getMessage());
        }

        // Fallback to API callout
        if (!installed) {
            System.enqueueJob(new BaseThemeInstallJob(version));
            sendAdminEmail('Base theme installation queued', version);
        }
    }
}
```

**Method 2: API Callout (Fallback)**
- Async job calls gem API: `GET /api/v1/base_themes/:version`
- Gem returns JSON with all templates
- Salesforce creates Theme + Templates
- Used for hot-fixes or if static resource fails

**Method 3: CLI Manual (Last Resort)**
```bash
sc base-theme install --version 22.0.0 --org production
```

**Parent Theme Preview & Switching**:

```bash
# Preview upgrade (doesn't commit)
sc theme preview-upgrade "Acme Store" --to "Base Theme v22"
# Returns: https://acme.example.com/?preview-parent=base-v22-sc-id
# Shows: Conflict detection report

# If good, commit upgrade
sc theme upgrade "Acme Store" --to "Base Theme v22"
# Creates ContentChange for approval
# After approval: updates parent_theme_id

# If issues, rollback
sc theme rollback "Acme Store"
# Reverts to previous parent (from history)
```

### Phase 5: CLI Polish (Week 5-6)

**Tasks:**
1. Enhance session management commands
2. Improve diff output
3. Add inheritance visualization
4. Better error messages
5. Integration tests

**Success Criteria:**
- Intuitive developer experience
- Clear feedback
- Comprehensive documentation

## Migration Strategy

### Existing Themes

**Backward Compatibility:**

```ruby
# Migration: Convert existing themes to new structure
class MigrateToThemeInheritance < ActiveRecord::Migration[7.0]
  def up
    # 1. Install base theme v21
    BaseTheme.install!('21.0.0')
    base_theme = Theme.find_by(is_base_theme: true, base_theme_version: '21.0.0')

    # 2. Convert each existing custom theme
    Theme.where(is_base_theme: false).find_each do |theme|
      # Extract templates to separate table
      migrate_templates(theme)

      # Extract assets to separate table
      migrate_assets(theme)

      # Diff against base theme to find overrides
      mark_overrides(theme, base_theme)

      # Set parent to base theme
      theme.update(
        parent_theme_id: base_theme.id,
        inheritance_depth: 1,
        base_theme_version: '21.0.0'
      )
    end
  end

  private

  def migrate_templates(theme)
    # Existing themes have templates in JSON field
    templates = JSON.parse(theme.templates_json || '[]')

    templates.each do |template|
      ThemeTemplate.create!(
        theme: theme,
        template_key: template['key'],
        content: template['content'],
        content_hash: Digest::SHA256.hexdigest(template['content'])
      )
    end
  end

  def mark_overrides(theme, base_theme)
    theme.templates.each do |template|
      base_template = base_theme.templates.find_by(template_key: template.template_key)

      if base_template && template.content_hash != base_template.content_hash
        template.update(overrides_parent: true)
      end
    end
  end
end
```

### Data Migration Steps

1. **Backup all theme data**
2. **Install base theme v21**
3. **Migrate existing themes to inheritance model**
4. **Verify all sites still render correctly**
5. **Update CLI to new version**
6. **Communicate changes to developers**

## Testing Strategy

### Unit Tests

```ruby
# spec/models/theme_spec.rb
RSpec.describe Theme do
  describe 'inheritance' do
    it 'resolves templates from parent chain' do
      base = create(:theme, :base)
      child = create(:theme, parent: base)
      grandchild = create(:theme, parent: child)

      create(:template, theme: base, key: 'pages/home', content: 'Base')
      create(:template, theme: child, key: 'pages/about', content: 'Child')
      create(:template, theme: grandchild, key: 'pages/home', content: 'Grandchild')

      expect(grandchild.resolve_template('pages/home')).to eq('Grandchild')
      expect(grandchild.resolve_template('pages/about')).to eq('Child')
    end
  end
end

# spec/models/content_change_session_spec.rb
RSpec.describe ContentChangeSession do
  describe 'lifecycle' do
    it 'transitions from draft to published' do
      session = create(:content_change_session, status: 'draft')

      session.submit!
      expect(session.status).to eq('submitted')

      session.approve!
      expect(session.status).to eq('approved')

      ContentChangePublisherJob.perform_now

      session.reload
      expect(session.status).to eq('published')
    end
  end
end
```

### Integration Tests

```ruby
# spec/requests/theme_preview_spec.rb
RSpec.describe 'Theme preview with session' do
  it 'shows session changes without affecting live site' do
    theme = create(:theme)
    session = create(:content_change_session, theme: theme, status: 'draft')
    create(:template_change, session: session, key: 'pages/home', content: 'Preview')

    # Live site shows original
    get "/pages/home"
    expect(response.body).not_to include('Preview')

    # Preview shows session changes
    get "/pages/home?session=#{session.session_id}"
    expect(response.body).to include('Preview')

    # After publish and approval
    session.submit!
    session.approve!
    ContentChangePublisherJob.perform_now

    # Now live site shows changes
    get "/pages/home"
    expect(response.body).to include('Preview')
  end
end
```

### Security Tests

```ruby
# spec/security/theme_update_spec.rb
RSpec.describe 'Theme update security' do
  it 'prevents direct theme updates outside session context' do
    theme = create(:theme)

    expect {
      theme.update!(name: 'Hacked')
    }.to raise_error(SecurityError, /only be updated via ContentChangeSession/)
  end

  it 'allows updates within publishing context' do
    theme = create(:theme)

    ContentChangeSession.with_publishing_context do
      expect {
        theme.update!(name: 'Updated')
      }.not_to raise_error
    end
  end
end
```

## Rollback Plan

If issues are discovered post-deployment:

1. **Immediate:** Disable ContentChangeSession enforcement
2. **Short-term:** Revert to direct theme updates
3. **Investigation:** Fix issues in staging
4. **Re-deploy:** With fixes and additional tests

## Documentation Updates

### For Developers

1. **Updated CLI guide** with session workflow
2. **Theme inheritance tutorial**
3. **Base theme upgrade guide**
4. **Troubleshooting guide** for approval flow

### For Admins

1. **ContentChange approval guide**
2. **Session review checklist**
3. **Reject/approve best practices**
4. **Emergency rollback procedures**

## Success Metrics

### Security
- ✅ Zero unauthorized live site changes
- ✅ 100% of changes go through approval
- ✅ Audit trail for all changes

### Developer Experience
- ✅ < 5 commands for typical workflow
- ✅ Clear session status feedback
- ✅ Easy base theme upgrades

### Performance
- ✅ Template resolution < 50ms
- ✅ Preview mode no slower than live
- ✅ Session publish < 5 minutes

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Breaking existing themes | High | Medium | Comprehensive migration testing, rollback plan |
| Performance degradation | Medium | Low | Caching, indexing, load testing |
| Complex debugging | Medium | Medium | Enhanced logging, session history |
| Migration data loss | High | Low | Multiple backups, dry-run migrations |
| Developer resistance | Medium | Medium | Clear documentation, training, support |

## Conclusion

This proposal addresses critical security issues while significantly improving the developer experience and theme management capabilities. The phased approach allows for incremental delivery and risk mitigation.

**Recommended Decision:** Approve for implementation starting with Phase 1 (Security Fix) immediately.

---

**Questions or Feedback:** Review and approve to proceed with implementation.
