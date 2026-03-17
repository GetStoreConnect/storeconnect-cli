# Agent-Friendly CLI Implementation Plan

**Goal:** Make the StoreConnect CLI fully usable by AI coding agents (like Claude Code, GitHub Copilot, etc.)

**Target Completion:** 4 weeks (8 milestones, 2 per week)

---

## Overview of Changes

### Core Principles
1. **Machine-readable output** - JSON everywhere
2. **Non-interactive** - All inputs via flags
3. **Discoverable** - Self-documenting commands
4. **Predictable** - Standardized errors and exit codes
5. **Safe** - Dry-run and validation before actions

### Major Features
1. JSON output mode for all commands
2. Non-interactive flags for all prompts
3. Structured error handling
4. Self-documenting help system
5. Watch mode for live development
6. Dry-run mode for preview
7. Status inspection commands
8. Example workflows library

---

## Milestone 1: Foundation - JSON Output & Exit Codes (Week 1, Days 1-3)

### Goals
- Add `--json` flag to all commands
- Implement standardized exit codes
- Create JSON response structs

### Tasks

#### 1.1 Exit Code Constants
**File:** `internal/commands/exit_codes.go` (NEW)
```go
package commands

const (
    ExitSuccess          = 0  // Operation completed successfully
    ExitGenericError     = 1  // Generic error
    ExitAuthError        = 2  // Authentication failed
    ExitNotFound         = 3  // Resource not found (theme, server, etc.)
    ExitValidationError  = 4  // Validation failed (invalid Liquid, bad config)
    ExitNetworkError     = 5  // Network/connection error
    ExitConflict         = 6  // Resource conflict (draft already exists)
    ExitUserCancelled    = 7  // User cancelled operation
    ExitConfigError      = 8  // Configuration error
)
```

#### 1.2 JSON Response Types
**File:** `internal/commands/responses.go` (NEW)
```go
package commands

// Standard success response
type SuccessResponse struct {
    Success bool                   `json:"success"`
    Data    interface{}            `json:"data,omitempty"`
    Message string                 `json:"message,omitempty"`
    Meta    map[string]interface{} `json:"meta,omitempty"`
}

// Standard error response
type ErrorResponse struct {
    Success    bool                   `json:"success"`
    Error      ErrorDetail            `json:"error"`
    Suggestion string                 `json:"suggestion,omitempty"`
}

type ErrorDetail struct {
    Code    string                 `json:"code"`
    Message string                 `json:"message"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Theme list response
type ThemeListResponse struct {
    Themes []ThemeInfo `json:"themes"`
    Total  int         `json:"total"`
}

type ThemeInfo struct {
    Name      string    `json:"name"`
    SCID      string    `json:"sc_id"`
    SFID      string    `json:"sfid,omitempty"`
    Version   string    `json:"version,omitempty"`
    CreatedAt time.Time `json:"created_at,omitempty"`
    UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Theme status response
type ThemeStatusResponse struct {
    Theme           string    `json:"theme"`
    Exists          bool      `json:"exists"`
    HasDraft        bool      `json:"has_draft"`
    DraftID         string    `json:"draft_id,omitempty"`
    DraftStatus     string    `json:"draft_status,omitempty"`
    TemplateChanges int       `json:"template_changes,omitempty"`
    CanPublish      bool      `json:"can_publish"`
    PreviewURL      string    `json:"preview_url,omitempty"`
}

// Connection status response
type StatusResponse struct {
    Connected     bool                     `json:"connected"`
    Servers       map[string]ServerStatus  `json:"servers"`
    DefaultServer string                   `json:"default_server,omitempty"`
    ProjectDir    string                   `json:"project_dir,omitempty"`
}

type ServerStatus struct {
    URL           string    `json:"url"`
    Version       string    `json:"version,omitempty"`
    Authenticated bool      `json:"authenticated"`
    LastChecked   time.Time `json:"last_checked,omitempty"`
}
```

#### 1.3 Global JSON Flag
**File:** `internal/commands/root.go`
```go
var (
    jsonOutput bool
    verbose    bool
)

func init() {
    rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}

// Helper to output response
func outputResponse(data interface{}, err error) error {
    if err != nil {
        return outputError(err)
    }

    if jsonOutput {
        return outputJSON(SuccessResponse{Success: true, Data: data})
    }

    // Human-friendly output (existing code)
    return nil
}

func outputError(err error) error {
    exitCode := ExitGenericError
    errorCode := "GENERIC_ERROR"
    suggestion := ""

    // Map error types to exit codes
    switch {
    case errors.Is(err, ErrNotFound):
        exitCode = ExitNotFound
        errorCode = "NOT_FOUND"
    case errors.Is(err, ErrAuth):
        exitCode = ExitAuthError
        errorCode = "AUTH_ERROR"
    // ... more mappings
    }

    if jsonOutput {
        outputJSON(ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    errorCode,
                Message: err.Error(),
            },
            Suggestion: suggestion,
        })
    } else {
        // Human-friendly error (existing)
    }

    os.Exit(exitCode)
    return err
}
```

#### 1.4 Update All Commands
- Add JSON output to: theme list, theme pull, theme push, theme preview, theme publish, theme delete, status, connect
- Use `outputResponse()` helper for consistent formatting
- Set appropriate exit codes on errors

### Success Criteria
- [ ] All commands support `--json` flag
- [ ] Exit codes are consistent and documented
- [ ] JSON output follows standard format
- [ ] Test: `sc theme list --json | jq .` works
- [ ] Test: Exit codes match error types

### Estimated Effort: 2 days

---

## Milestone 2: Non-Interactive Mode (Week 1, Days 4-5)

### Goals
- Add flags for all interactive prompts
- Implement `--non-interactive` global flag
- Error when required input missing in non-interactive mode

### Tasks

#### 2.1 Non-Interactive Flag
**File:** `internal/commands/root.go`
```go
var nonInteractive bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&nonInteractive, "non-interactive", false, "Non-interactive mode (error if input required)")
}
```

#### 2.2 Update Connect Command
**File:** `internal/commands/connect.go`
```go
var connectCmd = &cobra.Command{
    Use: "connect URL",
    // ...
}

var (
    connectOrgID    string
    connectStoreID  string
    connectAPIKey   string
    connectAlias    string
)

func init() {
    connectCmd.Flags().StringVar(&connectOrgID, "org-id", "", "Organization ID (15-18 chars)")
    connectCmd.Flags().StringVar(&connectStoreID, "store-id", "", "Store Salesforce ID")
    connectCmd.Flags().StringVar(&connectAPIKey, "api-key", "", "API key")
    connectCmd.Flags().StringVar(&connectAlias, "alias", "", "Server alias")
}

func runConnect(cmd *cobra.Command, args []string) error {
    url := args[0]

    // Get org ID
    orgID := connectOrgID
    if orgID == "" {
        if nonInteractive {
            return fmt.Errorf("--org-id required in non-interactive mode")
        }
        orgID = promptForOrgID() // existing interactive prompt
    }

    // Similar for store-id, api-key, alias
    // ...
}
```

#### 2.3 Update Other Interactive Commands
- `init` - add `--name` flag for project name
- All prompts should check `nonInteractive` flag

#### 2.4 Add Environment Variable Support
```go
// Allow credentials from environment
func getCredential(flagValue, envVar, promptText string) (string, error) {
    if flagValue != "" {
        return flagValue, nil
    }

    if envValue := os.Getenv(envVar); envValue != "" {
        return envValue, nil
    }

    if nonInteractive {
        return "", fmt.Errorf("%s required (set --%s flag or %s env var)",
            promptText, flagName, envVar)
    }

    return prompt(promptText), nil
}
```

### Success Criteria
- [ ] No interactive prompts when `--non-interactive` set
- [ ] Clear errors when required inputs missing
- [ ] Environment variables work for credentials
- [ ] Test: `sc connect URL --org-id X --store-id Y --api-key Z --alias dev --non-interactive`

### Estimated Effort: 1.5 days

---

## Milestone 3: Structured Errors & Help System (Week 2, Days 1-3)

### Goals
- Self-documenting help in JSON format
- Rich error messages with suggestions
- Command discovery system

### Tasks

#### 3.1 JSON Help Format
**File:** `internal/commands/help.go` (NEW)
```go
package commands

type CommandHelp struct {
    Name        string                `json:"name"`
    Usage       string                `json:"usage"`
    Short       string                `json:"short"`
    Long        string                `json:"long"`
    Flags       []FlagHelp            `json:"flags"`
    Subcommands []string              `json:"subcommands,omitempty"`
    Examples    []Example             `json:"examples"`
    ExitCodes   map[int]string        `json:"exit_codes"`
}

type FlagHelp struct {
    Name      string `json:"name"`
    Shorthand string `json:"shorthand,omitempty"`
    Type      string `json:"type"`
    Default   string `json:"default,omitempty"`
    Required  bool   `json:"required"`
    Usage     string `json:"usage"`
}

type Example struct {
    Description string `json:"description"`
    Command     string `json:"command"`
}

func getCommandHelp(cmd *cobra.Command) CommandHelp {
    help := CommandHelp{
        Name:  cmd.Name(),
        Usage: cmd.UseLine(),
        Short: cmd.Short,
        Long:  cmd.Long,
    }

    // Extract flags
    cmd.Flags().VisitAll(func(flag *pflag.Flag) {
        help.Flags = append(help.Flags, FlagHelp{
            Name:      flag.Name,
            Shorthand: flag.Shorthand,
            Type:      flag.Value.Type(),
            Default:   flag.DefValue,
            Usage:     flag.Usage,
        })
    })

    // Extract subcommands
    for _, subCmd := range cmd.Commands() {
        if !subCmd.Hidden {
            help.Subcommands = append(help.Subcommands, subCmd.Name())
        }
    }

    return help
}
```

Add `--help-format json` flag:
```bash
sc theme push --help-format json
{
  "name": "push",
  "usage": "push THEME_NAME",
  "short": "Upload theme to server",
  "flags": [
    {
      "name": "server",
      "shorthand": "s",
      "type": "string",
      "usage": "Server alias to use"
    },
    {
      "name": "json",
      "type": "bool",
      "usage": "Output in JSON format"
    }
  ],
  "examples": [
    {
      "description": "Push theme to default server",
      "command": "sc theme push my-theme"
    },
    {
      "description": "Push theme to specific server with JSON output",
      "command": "sc theme push my-theme --server prod --json"
    }
  ],
  "exit_codes": {
    "0": "Success",
    "3": "Theme not found",
    "4": "Validation error"
  }
}
```

#### 3.2 Command Discovery
**File:** `internal/commands/commands.go` (NEW)
```go
var commandsCmd = &cobra.Command{
    Use:   "commands",
    Short: "List all available commands",
    RunE:  runCommands,
}

func runCommands(cmd *cobra.Command, args []string) error {
    if jsonOutput {
        allCommands := extractAllCommands(rootCmd)
        return outputJSON(map[string]interface{}{
            "commands": allCommands,
            "version":  Version,
        })
    }

    // Human-friendly list
    return nil
}

func extractAllCommands(cmd *cobra.Command) []CommandInfo {
    var commands []CommandInfo

    for _, subCmd := range cmd.Commands() {
        if subCmd.Hidden {
            continue
        }

        info := CommandInfo{
            Name:        subCmd.Name(),
            Description: subCmd.Short,
            Usage:       subCmd.UseLine(),
        }

        // Recursively get subcommands
        if subCmd.HasSubCommands() {
            info.Subcommands = extractAllCommands(subCmd)
        }

        commands = append(commands, info)
    }

    return commands
}
```

#### 3.3 Enhanced Error Messages
**File:** `internal/commands/errors.go` (NEW)
```go
type CLIError struct {
    Code       string
    Message    string
    Details    map[string]interface{}
    Suggestion string
    ExitCode   int
}

func (e *CLIError) Error() string {
    return e.Message
}

// Predefined errors with suggestions
var (
    ErrThemeNotFound = func(themeName string, available []string) *CLIError {
        return &CLIError{
            Code:    "THEME_NOT_FOUND",
            Message: fmt.Sprintf("Theme '%s' not found", themeName),
            Details: map[string]interface{}{
                "theme_name":       themeName,
                "available_themes": available,
            },
            Suggestion: "Run 'sc theme list' to see available themes",
            ExitCode:   ExitNotFound,
        }
    }

    ErrNoDraft = func(themeName string) *CLIError {
        return &CLIError{
            Code:    "NO_DRAFT",
            Message: fmt.Sprintf("No draft found for theme '%s'", themeName),
            Details: map[string]interface{}{
                "theme_name": themeName,
            },
            Suggestion: "Run 'sc theme push " + themeName + "' to create a draft first",
            ExitCode:   ExitNotFound,
        }
    }

    // More predefined errors...
)
```

### Success Criteria
- [ ] `sc --help-format json` returns structured help
- [ ] `sc commands --json` lists all commands
- [ ] Errors include suggestions for fixing
- [ ] Test: Parse help JSON and generate command from it

### Estimated Effort: 2 days

---

## Milestone 4: Dry-Run & Validation (Week 2, Days 4-5)

### Goals
- Implement `--dry-run` flag for preview
- Add validation commands
- Show what would happen without doing it

### Tasks

#### 4.1 Dry-Run Flag
**File:** `internal/commands/root.go`
```go
var dryRun bool

func init() {
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview action without executing")
}
```

#### 4.2 Theme Push Dry-Run
**File:** `internal/commands/theme_push.go`
```go
func runThemePush(cmd *cobra.Command, args []string) error {
    // ... existing code to deserialize theme ...

    if dryRun {
        // Calculate what would happen
        changes := calculateChanges(themeData, serverTheme)

        dryRunResult := DryRunResult{
            Action: "theme_push",
            WouldCreate: map[string]interface{}{
                "content_change":  true,
                "templates_count": len(themeData.Templates),
            },
            Changes: changes,
            Warnings: []string{},
        }

        if jsonOutput {
            return outputJSON(dryRunResult)
        }

        // Human-friendly dry-run output
        formatter.Info("Dry-run mode - no changes will be made")
        formatter.Info(fmt.Sprintf("Would create draft with %d templates", len(themeData.Templates)))
        return nil
    }

    // ... existing push logic ...
}

type DryRunResult struct {
    Action      string                 `json:"action"`
    WouldCreate map[string]interface{} `json:"would_create"`
    Changes     []FileChange           `json:"changes"`
    Warnings    []string               `json:"warnings"`
}

type FileChange struct {
    File   string `json:"file"`
    Action string `json:"action"` // "create", "update", "delete"
}
```

#### 4.3 Validation Commands
**File:** `internal/commands/theme_validate.go`
```go
func runThemeValidate(cmd *cobra.Command, args []string) error {
    themeName := args[0]

    // Read theme from filesystem
    deserializer := theme.NewDeserializer(".")
    themeData, err := deserializer.Deserialize(themeName)
    if err != nil {
        return err
    }

    // Validate structure
    errors := []ValidationError{}

    // Check theme.yml exists
    if themeData.Name == "" {
        errors = append(errors, ValidationError{
            File:  "theme.yml",
            Error: "Missing or invalid theme.yml",
        })
    }

    // Validate each template
    for _, template := range themeData.Templates {
        if errs := validators.ValidateLiquid(template.Content); len(errs) > 0 {
            for _, err := range errs {
                errors = append(errors, ValidationError{
                    File:  fmt.Sprintf("templates/%s.liquid", template.Key),
                    Line:  err.Line,
                    Error: err.Message,
                })
            }
        }
    }

    result := ValidationResult{
        Valid:  len(errors) == 0,
        Errors: errors,
    }

    if jsonOutput {
        return outputJSON(result)
    }

    // Human-friendly output
    if result.Valid {
        formatter.Success(fmt.Sprintf("Theme '%s' is valid", themeName))
    } else {
        formatter.Error(fmt.Sprintf("Theme '%s' has %d validation errors", themeName, len(errors)))
        for _, err := range errors {
            formatter.Error(fmt.Sprintf("  %s:%d - %s", err.File, err.Line, err.Error))
        }
        os.Exit(ExitValidationError)
    }

    return nil
}

type ValidationResult struct {
    Valid  bool              `json:"valid"`
    Errors []ValidationError `json:"errors"`
}

type ValidationError struct {
    File  string `json:"file"`
    Line  int    `json:"line,omitempty"`
    Error string `json:"error"`
}
```

### Success Criteria
- [ ] `sc theme push my-theme --dry-run --json` shows what would happen
- [ ] `sc theme validate my-theme --json` validates Liquid templates
- [ ] Validation catches unclosed tags, syntax errors
- [ ] Test: Dry-run doesn't create server resources

### Estimated Effort: 1.5 days

---

## Milestone 5: Status & Inspection Commands (Week 3, Days 1-2)

### Goals
- Add `theme status` command
- Add `theme diff` command
- Inspect state before acting

### Tasks

#### 5.1 Theme Status Command
**File:** `internal/commands/theme_status.go` (NEW)
```go
var themeStatusCmd = &cobra.Command{
    Use:   "status THEME_NAME",
    Short: "Show theme status and draft information",
    Args:  cobra.ExactArgs(1),
    RunE:  runThemeStatus,
}

func runThemeStatus(cmd *cobra.Command, args []string) error {
    themeName := args[0]

    // Load sync state
    syncState, _ := config.NewSyncState(fmt.Sprintf("themes/%s", themeName))

    // Get server credentials
    // ... load credentials ...

    // Check if theme exists on server
    client := api.NewClient(cred.URL, cred.StoreSFID, cred.APIKey, api.WithOrgID(cred.OrgID))
    themesService := api.NewThemes(client)

    themes, err := themesService.List()
    if err != nil {
        return err
    }

    var serverTheme *api.Theme
    for _, t := range themes {
        if t.Name == themeName {
            serverTheme = &t
            break
        }
    }

    // Check for draft
    var draft *api.ContentChange
    var previewURL string
    if syncState != nil && syncState.ContentChangeSCID != "" {
        contentChangesService := api.NewContentChanges(client)
        draft, err = contentChangesService.Get(syncState.ContentChangeSCID)
        if err == nil {
            previewURL, _ = contentChangesService.GetPreviewURL(draft.SCID)
        }
    }

    // Build status response
    status := ThemeStatusResponse{
        Theme:      themeName,
        Exists:     serverTheme != nil,
        HasDraft:   draft != nil,
        CanPublish: draft != nil && draft.Status == "draft",
    }

    if draft != nil {
        status.DraftID = draft.SCID
        status.DraftStatus = draft.Status
        status.PreviewURL = previewURL
    }

    if jsonOutput {
        return outputJSON(status)
    }

    // Human-friendly output
    formatter.Info(fmt.Sprintf("Theme: %s", themeName))
    if serverTheme != nil {
        formatter.Success("  Exists on server: yes")
        formatter.Info(fmt.Sprintf("  SC ID: %s", serverTheme.SCID))
    } else {
        formatter.Warning("  Exists on server: no")
    }

    if draft != nil {
        formatter.Info(fmt.Sprintf("  Draft: %s (status: %s)", draft.SCID, draft.Status))
        if previewURL != "" {
            formatter.Info(fmt.Sprintf("  Preview URL: %s", previewURL))
        }
    } else {
        formatter.Info("  Draft: none")
    }

    return nil
}
```

#### 5.2 Theme Diff Command
**File:** `internal/commands/theme_diff.go` (NEW)
```go
var themeDiffCmd = &cobra.Command{
    Use:   "diff THEME_NAME",
    Short: "Show differences between local and server theme",
    Args:  cobra.ExactArgs(1),
    RunE:  runThemeDiff,
}

func runThemeDiff(cmd *cobra.Command, args []string) error {
    themeName := args[0]

    // Load local theme
    deserializer := theme.NewDeserializer(".")
    localTheme, err := deserializer.Deserialize(themeName)
    if err != nil {
        return err
    }

    // Load server theme
    // ... get server theme ...

    // Calculate differences
    diffs := calculateDiffs(localTheme, serverTheme)

    if jsonOutput {
        return outputJSON(map[string]interface{}{
            "theme":   themeName,
            "changes": diffs,
        })
    }

    // Human-friendly diff output
    for _, diff := range diffs {
        formatter.Info(fmt.Sprintf("%s %s", diff.Action, diff.File))
    }

    return nil
}

func calculateDiffs(local, server *api.Theme) []FileChange {
    var diffs []FileChange

    // Build map of server templates
    serverTemplates := make(map[string]string)
    for _, t := range server.Templates {
        serverTemplates[t.Key] = t.Content
    }

    // Check local templates
    for _, localTemplate := range local.Templates {
        serverContent, exists := serverTemplates[localTemplate.Key]

        if !exists {
            diffs = append(diffs, FileChange{
                File:   fmt.Sprintf("templates/%s.liquid", localTemplate.Key),
                Action: "create",
            })
        } else if localTemplate.Content != serverContent {
            diffs = append(diffs, FileChange{
                File:   fmt.Sprintf("templates/%s.liquid", localTemplate.Key),
                Action: "update",
            })
        }

        delete(serverTemplates, localTemplate.Key)
    }

    // Remaining server templates would be deleted
    for key := range serverTemplates {
        diffs = append(diffs, FileChange{
            File:   fmt.Sprintf("templates/%s.liquid", key),
            Action: "delete",
        })
    }

    return diffs
}
```

### Success Criteria
- [ ] `sc theme status my-theme --json` shows complete status
- [ ] `sc theme diff my-theme --json` shows file differences
- [ ] Test: Status shows draft info when draft exists
- [ ] Test: Diff shows create/update/delete actions

### Estimated Effort: 1.5 days

---

## Milestone 6: Watch Mode (Week 3, Days 3-5)

### Goals
- Implement file watching
- Auto-push on changes
- Live preview URL

### Tasks

#### 6.1 File Watcher
**File:** `internal/watcher/watcher.go` (NEW)
```go
package watcher

import (
    "path/filepath"
    "time"

    "github.com/fsnotify/fsnotify"
)

type Watcher struct {
    path       string
    onChange   func(path string)
    debounce   time.Duration
    watcher    *fsnotify.Watcher
    changes    map[string]time.Time
}

func NewWatcher(path string, onChange func(string)) (*Watcher, error) {
    fsWatcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    w := &Watcher{
        path:     path,
        onChange: onChange,
        debounce: 500 * time.Millisecond,
        watcher:  fsWatcher,
        changes:  make(map[string]time.Time),
    }

    // Watch all subdirectories
    if err := w.watchRecursive(path); err != nil {
        return nil, err
    }

    return w, nil
}

func (w *Watcher) watchRecursive(root string) error {
    return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return w.watcher.Add(path)
        }
        return nil
    })
}

func (w *Watcher) Start() {
    ticker := time.NewTicker(w.debounce)
    defer ticker.Stop()

    for {
        select {
        case event := <-w.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                w.changes[event.Name] = time.Now()
            }

        case <-ticker.C:
            // Process debounced changes
            now := time.Now()
            for path, changeTime := range w.changes {
                if now.Sub(changeTime) >= w.debounce {
                    w.onChange(path)
                    delete(w.changes, path)
                }
            }

        case err := <-w.watcher.Errors:
            // Log error
        }
    }
}

func (w *Watcher) Stop() {
    w.watcher.Close()
}
```

#### 6.2 Theme Watch Command
**File:** `internal/commands/theme_watch.go` (NEW)
```go
var themeWatchCmd = &cobra.Command{
    Use:   "watch THEME_NAME",
    Short: "Watch theme directory and auto-push changes",
    Long: `Watch theme directory for changes and automatically push updates to the server.
Creates a preview URL that you can refresh to see changes.`,
    Args: cobra.ExactArgs(1),
    RunE: runThemeWatch,
}

var (
    watchOpenBrowser bool
)

func init() {
    themeWatchCmd.Flags().BoolVar(&watchOpenBrowser, "open", false, "Open preview URL in browser")
}

func runThemeWatch(cmd *cobra.Command, args []string) error {
    themeName := args[0]
    themePath := fmt.Sprintf("themes/%s", themeName)

    formatter := ui.NewFormatter()

    // Initial push to create draft
    formatter.Info(fmt.Sprintf("Creating initial draft for '%s'...", themeName))
    contentChangeID, err := pushTheme(themeName)
    if err != nil {
        return err
    }

    // Get preview URL
    previewURL, err := getPreviewURL(contentChangeID)
    if err != nil {
        return err
    }

    if jsonOutput {
        // JSON mode: output events as stream
        outputJSON(map[string]interface{}{
            "event":       "started",
            "theme":       themeName,
            "preview_url": previewURL,
            "draft_id":    contentChangeID,
        })
    } else {
        formatter.Success(fmt.Sprintf("Watching %s for changes...", themePath))
        formatter.Info(fmt.Sprintf("Preview URL: %s", previewURL))
        formatter.Dim("Press Ctrl+C to stop")
        formatter.Newline()

        if watchOpenBrowser {
            // Open browser
            openBrowser(previewURL)
        }
    }

    // Create watcher
    w, err := watcher.NewWatcher(themePath, func(path string) {
        handleFileChange(path, themeName, contentChangeID)
    })
    if err != nil {
        return err
    }
    defer w.Stop()

    // Block until interrupted
    w.Start()

    return nil
}

func handleFileChange(path, themeName, contentChangeID string) {
    timestamp := time.Now().Format("15:04:05")

    if jsonOutput {
        outputJSON(map[string]interface{}{
            "event":     "file_changed",
            "timestamp": timestamp,
            "file":      path,
        })
    } else {
        formatter.Info(fmt.Sprintf("[%s] %s changed", timestamp, path))
        formatter.Dim("         → Pushing update...")
    }

    // Push update
    if err := pushThemeUpdate(themeName, contentChangeID); err != nil {
        if jsonOutput {
            outputJSON(map[string]interface{}{
                "event":     "error",
                "timestamp": timestamp,
                "error":     err.Error(),
            })
        } else {
            formatter.Error(fmt.Sprintf("         Failed: %v", err))
        }
        return
    }

    if jsonOutput {
        outputJSON(map[string]interface{}{
            "event":     "pushed",
            "timestamp": timestamp,
            "status":    "success",
        })
    } else {
        formatter.Success("         → Done! Refresh preview to see changes")
    }
}
```

Add dependency:
```bash
go get github.com/fsnotify/fsnotify
```

### Success Criteria
- [ ] `sc theme watch my-theme` monitors file changes
- [ ] Changes trigger automatic push
- [ ] Preview URL stays valid throughout session
- [ ] JSON mode outputs event stream
- [ ] Test: Edit file, save, see event in JSON output

### Estimated Effort: 2 days

---

## Milestone 7: Examples & Workflows Library (Week 4, Days 1-2)

### Goals
- Built-in example workflows
- Copy-pasteable commands
- Common patterns documented

### Tasks

#### 7.1 Examples Command
**File:** `internal/commands/examples.go` (NEW)
```go
var examplesCmd = &cobra.Command{
    Use:   "examples",
    Short: "Show example workflows",
    RunE:  runExamples,
}

var examplesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List all available examples",
    RunE:  runExamplesList,
}

var examplesShowCmd = &cobra.Command{
    Use:   "show EXAMPLE_NAME",
    Short: "Show specific example",
    Args:  cobra.ExactArgs(1),
    RunE:  runExamplesShow,
}

func init() {
    examplesCmd.AddCommand(examplesListCmd)
    examplesCmd.AddCommand(examplesShowCmd)
}

type Example struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Steps       []Step   `json:"steps"`
    Tags        []string `json:"tags"`
}

type Step struct {
    Description string `json:"description"`
    Command     string `json:"command"`
}

var exampleLibrary = []Example{
    {
        Name:        "create-and-publish-theme",
        Description: "Create a new theme, edit it, and publish to live site",
        Tags:        []string{"beginner", "theme", "publish"},
        Steps: []Step{
            {
                Description: "Initialize a new project",
                Command:     "sc init my-project",
            },
            {
                Description: "Connect to your StoreConnect server",
                Command:     "sc connect https://mystore.com --alias prod",
            },
            {
                Description: "Create a new theme",
                Command:     "sc theme new my-custom-theme",
            },
            {
                Description: "Edit theme files (use your editor)",
                Command:     "# Edit files in themes/my-custom-theme/",
            },
            {
                Description: "Validate theme before pushing",
                Command:     "sc theme validate my-custom-theme --json",
            },
            {
                Description: "Push theme as draft",
                Command:     "sc theme push my-custom-theme --json",
            },
            {
                Description: "Get preview URL",
                Command:     "sc theme preview my-custom-theme --json",
            },
            {
                Description: "Publish to live site",
                Command:     "sc theme publish my-custom-theme --json",
            },
        },
    },
    {
        Name:        "multi-server-workflow",
        Description: "Develop on staging, promote to production",
        Tags:        []string{"intermediate", "multi-server"},
        Steps: []Step{
            {
                Description: "Connect to staging server",
                Command:     "sc connect https://staging.mystore.com --alias staging",
            },
            {
                Description: "Connect to production server",
                Command:     "sc connect https://mystore.com --alias prod",
            },
            {
                Description: "Pull theme from production",
                Command:     "sc theme pull my-theme --server prod",
            },
            {
                Description: "Push changes to staging",
                Command:     "sc theme push my-theme --server staging --json",
            },
            {
                Description: "Preview on staging",
                Command:     "sc theme preview my-theme --server staging --json",
            },
            {
                Description: "When ready, push to production",
                Command:     "sc theme push my-theme --server prod --json",
            },
            {
                Description: "Publish on production",
                Command:     "sc theme publish my-theme --server prod --json",
            },
        },
    },
    {
        Name:        "watch-mode-development",
        Description: "Use watch mode for rapid iteration",
        Tags:        []string{"intermediate", "development", "watch"},
        Steps: []Step{
            {
                Description: "Start watch mode (opens preview URL)",
                Command:     "sc theme watch my-theme --open",
            },
            {
                Description: "Edit files in your editor",
                Command:     "# Changes auto-push on save",
            },
            {
                Description: "Refresh browser to see changes",
                Command:     "# No manual push needed!",
            },
            {
                Description: "When done, publish",
                Command:     "sc theme publish my-theme --json",
            },
        },
    },
    {
        Name:        "agent-workflow",
        Description: "Automated workflow for AI agents",
        Tags:        []string{"advanced", "automation", "agent"},
        Steps: []Step{
            {
                Description: "Check connection status",
                Command:     "sc status --json",
            },
            {
                Description: "Get available commands",
                Command:     "sc commands --json",
            },
            {
                Description: "List themes to verify target exists",
                Command:     "sc theme list --json",
            },
            {
                Description: "Check theme status before modifying",
                Command:     "sc theme status my-theme --json",
            },
            {
                Description: "Validate before pushing",
                Command:     "sc theme validate my-theme --json",
            },
            {
                Description: "Dry-run to preview changes",
                Command:     "sc theme push my-theme --dry-run --json",
            },
            {
                Description: "Push if dry-run looks good",
                Command:     "sc theme push my-theme --json",
            },
            {
                Description: "Get preview URL for verification",
                Command:     "sc theme preview my-theme --json",
            },
            {
                Description: "Verify preview (screenshot, accessibility, etc.)",
                Command:     "# Use external tools to verify preview_url",
            },
            {
                Description: "Publish if verification passes",
                Command:     "sc theme publish my-theme --json",
            },
        },
    },
}

func runExamplesList(cmd *cobra.Command, args []string) error {
    if jsonOutput {
        list := []map[string]interface{}{}
        for _, ex := range exampleLibrary {
            list = append(list, map[string]interface{}{
                "name":        ex.Name,
                "description": ex.Description,
                "tags":        ex.Tags,
            })
        }
        return outputJSON(map[string]interface{}{
            "examples": list,
        })
    }

    formatter := ui.NewFormatter()
    formatter.Info("Available examples:")
    formatter.Newline()

    for _, ex := range exampleLibrary {
        formatter.Success(fmt.Sprintf("  %s", ex.Name))
        formatter.Dim(fmt.Sprintf("    %s", ex.Description))
        formatter.Dim(fmt.Sprintf("    Tags: %s", strings.Join(ex.Tags, ", ")))
        formatter.Newline()
    }

    formatter.Info("Run 'sc examples show <name>' to see details")

    return nil
}

func runExamplesShow(cmd *cobra.Command, args []string) error {
    name := args[0]

    var example *Example
    for _, ex := range exampleLibrary {
        if ex.Name == name {
            example = &ex
            break
        }
    }

    if example == nil {
        return fmt.Errorf("example '%s' not found", name)
    }

    if jsonOutput {
        return outputJSON(example)
    }

    formatter := ui.NewFormatter()
    formatter.Success(example.Name)
    formatter.Info(example.Description)
    formatter.Newline()

    for i, step := range example.Steps {
        formatter.Info(fmt.Sprintf("%d. %s", i+1, step.Description))
        formatter.Dim(fmt.Sprintf("   $ %s", step.Command))
        formatter.Newline()
    }

    return nil
}
```

### Success Criteria
- [ ] `sc examples list --json` returns all examples
- [ ] `sc examples show agent-workflow --json` returns steps
- [ ] Human-friendly output is readable
- [ ] Test: Can copy-paste commands and they work

### Estimated Effort: 1 day

---

## Milestone 8: Documentation & Polish (Week 4, Days 3-5)

### Goals
- Update all documentation
- Add agent-specific guides
- Final testing and polish

### Tasks

#### 8.1 Agent Usage Guide
**File:** `docs/AGENT_USAGE.md` (NEW)
```markdown
# StoreConnect CLI for AI Agents

This guide explains how to use the StoreConnect CLI from AI coding agents.

## Key Features for Agents

- **JSON Output**: All commands support `--json` flag
- **Non-Interactive**: All inputs via flags (no prompts)
- **Discoverable**: Self-documenting help system
- **Predictable**: Standardized exit codes and errors
- **Safe**: Dry-run and validation before actions

## Quick Start

### 1. Discover Available Commands

```bash
sc commands --json
```

Returns all available commands with descriptions.

### 2. Get Command Help

```bash
sc theme push --help-format json
```

Returns structured help including flags, examples, exit codes.

### 3. Check Status

```bash
sc status --json
```

Returns connection status, available servers, project info.

### 4. List Resources

```bash
sc theme list --json
```

Returns all themes with SC IDs, names, versions.

## Complete Workflow Example

```bash
# 1. Check connection
sc status --json
# Output: {"connected": true, "servers": {...}}

# 2. List themes
sc theme list --json
# Output: {"themes": [{"name": "my-theme", "sc_id": "..."}]}

# 3. Check theme status
sc theme status my-theme --json
# Output: {"theme": "my-theme", "exists": true, "has_draft": false}

# 4. Validate before pushing
sc theme validate my-theme --json
# Output: {"valid": true, "errors": []}

# 5. Preview changes (dry-run)
sc theme push my-theme --dry-run --json
# Output: {"would_create": {...}, "changes": [...]}

# 6. Push if validation passes
sc theme push my-theme --json
# Output: {"success": true, "data": {"draft_id": "..."}}

# 7. Get preview URL
sc theme preview my-theme --json
# Output: {"preview_url": "https://..."}

# 8. Publish
sc theme publish my-theme --json
# Output: {"success": true, "message": "Published"}
```

## Error Handling

All errors return JSON with:
- `code`: Machine-readable error code
- `message`: Human-readable message
- `details`: Additional context
- `suggestion`: How to fix

```json
{
  "success": false,
  "error": {
    "code": "THEME_NOT_FOUND",
    "message": "Theme 'nonexistent' not found",
    "details": {
      "theme_name": "nonexistent",
      "available_themes": ["my-theme", "other-theme"]
    }
  },
  "suggestion": "Run 'sc theme list' to see available themes"
}
```

## Exit Codes

- `0` - Success
- `1` - Generic error
- `2` - Authentication error
- `3` - Not found
- `4` - Validation error
- `5` - Network error
- `6` - Conflict
- `7` - User cancelled
- `8` - Configuration error

Check exit code to determine error type:
```bash
sc theme push nonexistent --json
echo $?  # Prints: 3 (NOT_FOUND)
```

## Best Practices

### 1. Always Use JSON Mode

```bash
sc theme list --json | jq '.themes[0].sc_id'
```

### 2. Validate Before Pushing

```bash
sc theme validate my-theme --json
if [ $? -eq 0 ]; then
    sc theme push my-theme --json
fi
```

### 3. Check Status Before Acting

```bash
STATUS=$(sc theme status my-theme --json)
HAS_DRAFT=$(echo $STATUS | jq '.has_draft')

if [ "$HAS_DRAFT" = "true" ]; then
    # Publish existing draft
    sc theme publish my-theme --json
else
    # Create new draft
    sc theme push my-theme --json
fi
```

### 4. Use Dry-Run for Planning

```bash
# See what would happen
sc theme push my-theme --dry-run --json

# If acceptable, do it
sc theme push my-theme --json
```

### 5. Parse Errors for Recovery

```python
import json
import subprocess

result = subprocess.run(
    ['sc', 'theme', 'push', 'my-theme', '--json'],
    capture_output=True,
    text=True
)

if result.returncode != 0:
    error = json.loads(result.stdout)

    if error['error']['code'] == 'THEME_NOT_FOUND':
        # Create theme first
        subprocess.run(['sc', 'theme', 'new', 'my-theme', '--json'])
        # Retry push
        subprocess.run(['sc', 'theme', 'push', 'my-theme', '--json'])
```

## Common Patterns

### Check if Resource Exists

```bash
sc theme list --json | jq -e '.themes[] | select(.name == "my-theme")'
if [ $? -eq 0 ]; then
    echo "Theme exists"
else
    echo "Theme does not exist"
fi
```

### Get Specific Field

```bash
PREVIEW_URL=$(sc theme preview my-theme --json | jq -r '.preview_url')
echo "Preview at: $PREVIEW_URL"
```

### Handle Multiple Servers

```bash
# List themes on staging
sc theme list --server staging --json

# Push to production
sc theme push my-theme --server prod --json
```

### Batch Operations

```bash
# Get all theme names
THEMES=$(sc theme list --json | jq -r '.themes[].name')

# Validate each
for theme in $THEMES; do
    sc theme validate $theme --json
done
```

## Example Workflows

See built-in examples:
```bash
sc examples list --json
sc examples show agent-workflow --json
```

## Troubleshooting

### Command Not Found

```bash
sc commands --json
```

Lists all available commands.

### Flag Not Recognized

```bash
sc theme push --help-format json
```

Returns all valid flags for the command.

### Unexpected Error

```bash
sc theme push my-theme --verbose --json
```

Returns detailed execution steps.

## Integration Examples

### Python

```python
import json
import subprocess

def sc_command(args):
    result = subprocess.run(
        ['sc'] + args + ['--json'],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return json.loads(result.stdout)
    else:
        error = json.loads(result.stdout)
        raise Exception(error['error']['message'])

# Usage
themes = sc_command(['theme', 'list'])
print(f"Found {len(themes['themes'])} themes")
```

### Node.js

```javascript
const { execSync } = require('child_process');

function sc(args) {
  try {
    const output = execSync(`sc ${args.join(' ')} --json`, {
      encoding: 'utf-8'
    });
    return JSON.parse(output);
  } catch (err) {
    const error = JSON.parse(err.stdout);
    throw new Error(error.error.message);
  }
}

// Usage
const themes = sc(['theme', 'list']);
console.log(`Found ${themes.themes.length} themes`);
```

### Go

```go
import (
    "encoding/json"
    "os/exec"
)

func sc(args ...string) (map[string]interface{}, error) {
    args = append(args, "--json")
    cmd := exec.Command("sc", args...)

    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    var result map[string]interface{}
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, err
    }

    return result, nil
}

// Usage
result, _ := sc("theme", "list")
themes := result["themes"].([]interface{})
fmt.Printf("Found %d themes\n", len(themes))
```

## Support

For issues or questions:
- GitHub Issues: https://github.com/GetStoreConnect/storeconnect-cli/issues
- Documentation: https://docs.storeconnect.com
```

#### 8.2 Update Main README
- Add section: "For AI Agents" with link to AGENT_USAGE.md
- Update all examples to show JSON output option
- Add exit code reference

#### 8.3 API Reference Documentation
**File:** `docs/API_REFERENCE.md` (NEW)
- Document all JSON response formats
- List all exit codes
- Document all error codes
- Include request/response examples for every command

#### 8.4 Testing
- Create integration test suite that uses JSON output
- Test all commands with `--json` flag
- Test error conditions return proper exit codes
- Test `--dry-run` doesn't create server resources
- Test `--non-interactive` fails when input needed

### Success Criteria
- [ ] `docs/AGENT_USAGE.md` is comprehensive
- [ ] README updated with agent information
- [ ] All examples work as documented
- [ ] Integration tests pass
- [ ] Can use CLI from Python/Node/Go scripts

### Estimated Effort: 2 days

---

## Summary Timeline

| Week | Days | Milestone | Focus |
|------|------|-----------|-------|
| 1 | 1-3 | M1: Foundation | JSON output, exit codes, response types |
| 1 | 4-5 | M2: Non-Interactive | Flags for all prompts, env vars |
| 2 | 1-3 | M3: Help & Errors | JSON help, command discovery, error suggestions |
| 2 | 4-5 | M4: Validation | Dry-run mode, validation commands |
| 3 | 1-2 | M5: Inspection | Status commands, diff commands |
| 3 | 3-5 | M6: Watch Mode | File watching, auto-push, live preview |
| 4 | 1-2 | M7: Examples | Example workflows, common patterns |
| 4 | 3-5 | M8: Documentation | Agent guide, API reference, testing |

**Total: 4 weeks (20 work days)**

---

## Success Metrics

### Quantitative
- [ ] 100% of commands support `--json` flag
- [ ] 100% of interactive prompts have flag alternatives
- [ ] All commands have JSON help available
- [ ] Exit codes cover all error types
- [ ] Integration tests achieve 80%+ coverage

### Qualitative
- [ ] AI agent can discover all commands programmatically
- [ ] AI agent can handle errors and recover
- [ ] AI agent can complete full workflow without human intervention
- [ ] Documentation is clear and comprehensive
- [ ] JSON output is consistent across all commands

---

## Future Enhancements (Post-Launch)

### Phase 2 Features
1. **Batch Operations** - Multi-theme push, bulk validation
2. **Transaction Support** - Changeset management, rollback
3. **Webhook Integration** - Callbacks for long operations
4. **Config Profiles** - Predefined configurations for different scenarios
5. **Plugin System** - Custom commands and workflows

### Phase 3 Features
1. **Language SDKs** - Python, Node.js, Go packages
2. **GraphQL API** - Alternative to CLI for programmatic access
3. **Real-time Events** - WebSocket stream for watch mode
4. **Advanced Diff** - Line-by-line template diffs
5. **Conflict Resolution** - Handle concurrent edits

---

## Implementation Notes

### Dependencies to Add
```bash
go get github.com/fsnotify/fsnotify  # File watching
```

### Breaking Changes
- None - all features are additive and backward compatible
- Existing commands continue to work with human-friendly output
- `--json` is opt-in

### Testing Strategy
1. **Unit Tests** - Test JSON serialization, exit codes
2. **Integration Tests** - Test full workflows with JSON output
3. **Agent Tests** - Python scripts using CLI programmatically
4. **Manual Tests** - Verify human-friendly output still works

### Rollout Plan
1. Implement features in milestone order
2. Release beta version for testing with agents
3. Gather feedback from AI agent usage
4. Iterate based on feedback
5. Stable release with comprehensive documentation

---

## Getting Started

To begin implementation:

```bash
cd cli-go
git checkout -b feature/agent-friendly-cli

# Start with Milestone 1
# 1. Create exit_codes.go
# 2. Create responses.go
# 3. Add --json flag to theme list
# 4. Test: sc theme list --json | jq .
```

Track progress by checking off tasks in this document as they're completed.
