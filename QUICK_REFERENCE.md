# TIT Architecture - Quick Reference Guide

## AppMode Values (14 Total)

```go
0  = ModeMenu               // Main menu
1  = ModeInput              // Generic input (deprecated)
2  = ModeConsole            // Git output streaming
3  = ModeConfirmation       // Y/N dialog
4  = ModeHistory            // Commit history browser
5  = ModeConflictResolve    // 3-way conflict UI
6  = ModeInitializeLocation // Init location picker
7  = ModeInitializeBranches // Init branch input
8  = ModeCloneURL           // Clone URL input
9  = ModeCloneLocation      // Clone location picker
10 = ModeClone              // Clone async operation
11 = ModeSelectBranch       // Clone: select branch
12 = ModeFileHistory        // File history browser
13 = ModeSetupWizard        // Git env setup
```

## Git State Detection (5-Axis Tuple)

```
AXIS 0: GitEnvironment (PRIORITY 0)
  Ready | NeedsSetup | MissingGit | MissingSSH

AXIS 1: Operation (PRIORITY 1)
  NotRepo | Normal | Conflicted | Merging | Rebasing | 
  DirtyOperation | TimeTraveling | Rewinding

AXIS 2: WorkingTree (PRIORITY 2)
  Clean | Dirty

AXIS 3: Remote (PRIORITY 3)
  NoRemote | HasRemote

AXIS 4: Timeline (PRIORITY 4, conditional)
  InSync | Ahead | Behind | Diverged | (empty if N/A)
  Only checked if: Operation == Normal && Remote == HasRemote
```

## Git State Detection Commands

```bash
# WorkingTree
git status --porcelain=v2

# Operation (priority order)
git status --porcelain=v2          # Check for "u " lines
stat .git/TIT_TIME_TRAVEL          # Time traveling?
stat .git/MERGE_HEAD               # Merging?
stat .git/rebase-merge             # Rebasing?
stat .git/rebase-apply             # Rebasing?

# Remote
git remote

# Timeline (only if Normal && HasRemote)
git rev-list --left-right --count HEAD...@{u}

# Branch info
git symbolic-ref --short HEAD      # Branch name
git rev-parse HEAD                 # Commit hash
git rev-parse @{u}                 # Upstream commit
```

## File Structure (56 Files, ~15,200 LOC)

```
internal/app/           (26 files, ~6000 LOC)
  ├─ app.go                      (1407 lines) - Core Application struct
  ├─ modes.go                    (162 lines)  - AppMode enumeration
  ├─ menu.go                     (300+ lines) - Menu generation
  ├─ menu_items.go               (200+ lines) - Menu definitions
  ├─ handlers.go                 (500+ lines) - Event handlers
  ├─ keyboard.go                 (400+ lines) - Key bindings
  ├─ git_handlers.go             (600+ lines) - Git operations
  ├─ messages.go                 (200+ lines) - Bubble Tea messages
  ├─ state_info.go               (150+ lines) - State formatting
  ├─ history_cache.go            (300+ lines) - Metadata caching
  ├─ dispatchers.go              (300+ lines) - Action dispatch
  ├─ operations.go               (400+ lines) - Long-running ops
  ├─ conflict_handlers.go         - Merge conflict handling
  ├─ conflict_state.go            - Conflict state tracking
  ├─ dirty_state.go               - Dirty operation tracking
  ├─ confirmation_handlers.go     - Dialog handlers
  ├─ setup_wizard.go              - Setup wizard logic
  ├─ async.go                     - Async operation management
  ├─ cursor_movement.go           - Cursor navigation
  ├─ key_builder.go               - Key handler builder
  ├─ location.go                  - Location state tracking
  ├─ errors.go                    - Error handling
  └─ config.go                    - App configuration

internal/ui/            (18 files, ~5000 LOC)
  ├─ layout.go                   (200+ lines) - Reactive layout
  ├─ sizing.go                   (64 lines)   - Dynamic sizing
  ├─ menu.go                     (250+ lines) - Menu rendering
  ├─ header.go                   (200+ lines) - State header
  ├─ console.go                  (300+ lines) - Git output
  ├─ history.go                  (400+ lines) - Commit history
  ├─ filehistory.go              (400+ lines) - File history
  ├─ conflictresolver.go         (400+ lines) - Conflict UI
  ├─ theme.go                    (300+ lines) - Color themes
  ├─ buffer.go                   (150+ lines) - Output buffering
  ├─ input.go                    (200+ lines) - Text input
  ├─ box.go                      - Box drawing
  ├─ confirmation.go             - Confirmation dialog
  ├─ spinner.go                  - Loading spinner
  ├─ statusbar.go                - Footer status bar
  ├─ textinput.go                - Text input UI
  ├─ textpane.go                 - Text display pane
  ├─ listpane.go                 - List display pane
  ├─ branchinput.go              - Branch input component
  ├─ validation.go               - Input validation
  ├─ formatters.go               - Text formatting
  └─ assets/                     - Asset files

internal/git/           (8 files, ~2500 LOC)
  ├─ state.go                    (507 lines)  - State detection
  ├─ types.go                    (110 lines)  - Type definitions
  ├─ execute.go                  (200+ lines) - Git execution
  ├─ environment.go              (150+ lines) - Env detection
  ├─ init.go                     (200+ lines) - Repo init
  ├─ ssh.go                      (300+ lines) - SSH keys
  ├─ messages.go                 (100+ lines) - Message types
  └─ dirtyop.go                  - Dirty op tracking

internal/config/        (1 file)
  └─ stash.go                    (100+ lines) - Stash tracking

internal/banner/        (2 files)
  ├─ svg.go                      - SVG rendering
  └─ braille.go                  - Braille characters

cmd/tit/                (1 file)
  └─ main.go                     (22 lines)   - Entry point
```

## Key Data Structures

### Application Struct (Core State Container)
```go
type Application struct {
    // Dimensions & UI
    width, height int
    sizing ui.DynamicSizing
    theme ui.Theme
    
    // Current state
    mode AppMode
    gitState *git.State
    gitEnvironment git.GitEnvironment
    
    // Menu
    selectedIndex int
    menuItems []MenuItem
    keyHandlers map[AppMode]map[string]KeyHandler
    
    // Input state
    inputPrompt, inputValue string
    inputCursorPosition int
    inputValidationMsg string
    
    // Workflow states
    cloneURL, clonePath string
    cloneBranches []string
    
    // Async state
    asyncOperationActive, asyncOperationAborted bool
    isExitAllowed bool
    previousMode AppMode
    
    // Output
    consoleState ui.ConsoleOutState
    outputBuffer *ui.OutputBuffer
    
    // Special states
    conflictResolveState *ConflictResolveState
    historyState *ui.HistoryState
    fileHistoryState *ui.FileHistoryState
    timeTravelInfo *git.TimeTravelInfo
    
    // Caches
    historyMetadataCache map[string]*git.CommitDetails
    fileHistoryDiffCache map[string]string
    cacheLoadingStarted, cacheMetadata, cacheDiffs bool
}
```

### Git State Struct
```go
type State struct {
    WorkingTree WorkingTree          // Clean | Dirty
    Timeline Timeline                // InSync | Ahead | Behind | Diverged | ""
    Operation Operation              // NotRepo | Normal | ...
    Remote Remote                    // NoRemote | HasRemote
    CurrentBranch string
    CurrentHash string
    RemoteHash string
    CommitsAhead, CommitsBehind int
    LocalBranchOnRemote bool
    Detached bool
}
```

### MenuItem Struct
```go
type MenuItem struct {
    ID string        // Unique identifier
    Shortcut string  // Single character
    Emoji string     // Leading emoji
    Label string     // Max 21 chars
    Hint string      // Footer hint
    Enabled bool     // Selectable
    Separator bool   // Visual separator
}
```

## Menu Item Categories (30+ Items)

```
NotRepo:
  - init, clone

Working Tree (Dirty):
  - commit, commit_push

Timeline (HasRemote):
  - push, force_push, pull_merge, dirty_pull_merge,
    replace_local, pull_merge_diverged, reset_discard_changes

Remote (NoRemote):
  - add_remote

History (Always):
  - history, file_history

Time Travel:
  - time_travel_history, time_travel_files_history,
    rewind, return_from_timetravel

Conflict:
  - resolve_conflicts, abort_merge
```

## Render Functions (UI Components)

```go
// Layout
RenderReactiveLayout()    // Full page with header/content/footer
RenderMenuWithBanner()    // Menu + ASCII banner
RenderMenuWithHeight()    // Menu only

// Content modes
RenderConsoleOutput()     // Git command output
RenderHistorySplitPane()  // Commit history browser
RenderFileHistorySplitPane() // File history browser
RenderConflictResolveGeneric() // Merge conflict UI
RenderTextInput()         // Text input field

// Components
RenderHeader()            // State information (5 rows)
RenderStatusBar()         // Footer status line
RenderBox()               // Bordered box
RenderHeaderInfo()        // Header content
RenderMenuHighlight()     // Highlighted menu item

// Internal
RenderListPane()          // Generic list rendering
RenderDiffPane()          // Diff content rendering
```

## Sizing Constants

```go
MinWidth = 69         // Minimum terminal width
MinHeight = 19        // Minimum terminal height
HeaderHeight = 9      // Header rows
FooterHeight = 1      // Footer rows
MinContentHeight = 4  // Minimum content rows
HorizontalMargin = 2  // Left/right padding
BannerWidth = 30      // Menu banner width
```

## Message Types (Bubble Tea)

```go
type TickMsg struct{}              // Periodic timer
type ClearTickMsg struct{}         // Clear timeout
type OutputRefreshMsg struct{}     // Console refresh
type GitOperationMsg struct{}      // Git result
type CacheProgressMsg struct{}     // Cache update
type RestoreTimeTravelMsg struct{} // Time travel restore
type SetupCompleteMsg struct{}     // Setup wizard step
type RewindMsg struct{}            // Reset --hard done
type RemoteFetchMsg struct{}       // Fetch remote done
// ... 10+ more
```

## Key Handler Registry

```go
// Global handlers (all modes)
ctrl+c, q       → Quit
esc             → Return/Cancel
ctrl+v, cmd+v   → Paste

// Mode-specific
ModeMenu:       up/down/j/k, enter, [shortcuts]
ModeInput:      left/right/home/end, backspace, enter
ModeHistory:    up/down/j/k, tab, enter, ctrl+r
ModeFileHistory: up/down/j/k, tab, v, y, esc
ModeConflict:   up/down/j/k, tab, space, enter
```

## Setup Wizard Steps (5)

```
0. Welcome      → Display welcome message
1. Prerequisites → Check git + ssh installed
2. Email        → Input email for key
3. Generate     → Generate SSH key + agent config
4. Display Key  → Show public key
5. Complete     → Setup finished
```

## State Info Maps

### WorkingTree Display Info
```
Clean:
  Emoji: ✅
  Label: "Clean"
  Color: theme.StatusClean
  Desc: "No uncommitted changes"

Dirty:
  Emoji: 📝
  Label: "Dirty"
  Color: theme.StatusDirty
  Desc: "Uncommitted changes pending"
```

### Timeline Display Info
```
InSync:
  Emoji: 🔗, Label: "Sync", Desc: "Local and remote in sync"

Ahead:
  Emoji: 🌎, Label: "Local ahead", Desc: "N commits ahead of remote"

Behind:
  Emoji: 🪐, Label: "Local behind", Desc: "N commits behind remote"

Diverged:
  Emoji: 💥, Label: "Diverged", Desc: "N ahead, N behind (conflict)"
```

### Operation Display Info
```
Normal:
  Emoji: 🟢, Label: "READY"

Conflicted:
  Emoji: ⚡, Label: "CONFLICTED"

Merging:
  Emoji: 🔀, Label: "MERGING"

Rebasing:
  Emoji: 🔄, Label: "REBASING"

TimeTraveling:
  Emoji: 📌, Label: "DETACHED @ <hash>"
```

## Cache System

```go
type Application struct {
    // Metadata cache (commit details)
    historyMetadataCache map[string]*git.CommitDetails
    cacheMetadata bool
    cacheMetadataProgress, cacheMetadataTotal int
    
    // Diff cache (file change diffs)
    fileHistoryDiffCache map[string]string
    cacheDiffs bool
    cacheDiffsProgress, cacheDiffsTotal int
    
    // File list cache
    fileHistoryFilesCache map[string][]git.FileInfo
    
    // Guard flags
    cacheLoadingStarted bool
    cacheAnimationFrame int
    
    // Mutexes
    historyCacheMutex sync.Mutex
    diffCacheMutex sync.Mutex
}

// CONTRACT: Mandatory precomputation
// History menus DISABLED until caches ready
// Progress shown during loading
```

## Time Travel State File

```
Location: .git/TIT_TIME_TRAVEL
Format:
  Line 1: Original branch name (e.g., "main")
  Line 2: Original stash ID (if dirty work saved)

Example:
  main
  stash@{0}
```

## Error Handling Pattern

```go
type ErrorConfig struct {
    Level ErrorLevel          // ErrorWarn, ErrorError
    Message string            // Human-readable message
    InnerError error          // Original error
    BufferLine string         // Console output
    FooterLine string         // Footer status
}

a.LogError(ErrorConfig{...})
```

## Common Workflows

### Initialize Repository
```
NotRepo state
  → ModeInitializeLocation (here/subdir)
  → ModeInitializeBranches (canon + working)
  → ModeConsole (git init, setup)
  → ModeMenu (now Ready)
```

### Clone Repository
```
NotRepo state
  → ModeCloneURL (input URL)
  → ModeCloneLocation (here/subdir)
  → ModeClone (async clone)
  → ModeSelectBranch (select canon)
  → ModeConsole (setup)
  → ModeMenu (now Ready)
```

### Merge Conflict
```
Normal state + Changes pushed → Conflicted
  → Detect conflicts (git status)
  → ModeConflictResolve
  → Select resolution (base/ours/theirs)
  → Confirm resolution
  → ModeConsole (git add, continue)
  → ModeMenu (back to Normal)
```

### Time Travel
```
History browser (ModeHistory)
  → Select commit to view
  → Save current work (stash if dirty)
  → Checkout commit (detached HEAD)
  → Write .git/TIT_TIME_TRAVEL marker
  → Operation = TimeTraveling
  → Can browse files, view diffs
  → Press Ctrl+R to reset/return
  → Restore original branch + work
```

---

**Generated**: January 22, 2026  
**For**: Architecture Documentation
