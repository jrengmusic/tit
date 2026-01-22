# TIT Codebase Architecture Analysis

**Project**: TIT (Terminal Interactive Tool) - A git-focused terminal UI in Go  
**Language**: Go  
**Size**: ~15,200 lines of code  
**Framework**: Bubble Tea (Charmbracelet)  

---

## 1. PACKAGE STRUCTURE & RESPONSIBILITIES

### Package Hierarchy

```
tit/
├── cmd/
│   └── tit/
│       └── main.go              # Entry point: theme init → app creation → tea.Run()
├── internal/
│   ├── app/                     # Application state management, mode handling
│   ├── git/                     # Git state detection, command execution
│   ├── ui/                      # UI rendering components and layouts
│   ├── config/                  # Configuration management (stash, etc)
│   └── banner/                  # ASCII art rendering (SVG, Braille)
└── infra/                       # Infrastructure/utilities
```

### Package Responsibilities

#### `internal/app` (Core Application Logic)
**Purpose**: Manages application state, mode transitions, menu generation, and event handling

**Key Components**:
- `app.go` - Main Application struct and lifecycle
- `modes.go` - AppMode enumeration (14 modes)
- `menu.go` - Menu generation based on git state
- `menu_items.go` - Menu item definitions (single source of truth)
- `handlers.go` - Event handlers for user actions
- `keyboard.go` - Key binding management
- `git_handlers.go` - Git operation handlers
- `conflict_handlers.go` - Merge conflict resolution
- `state_info.go` - State display formatting
- `messages.go` - Bubble Tea message types
- `async.go` - Asynchronous operation management
- `history_cache.go` - History metadata caching

**Responsibilities**:
- Maintain application state lifecycle
- Route events to appropriate handlers based on current mode
- Generate context-aware menus
- Manage async git operations (clone, push, pull, etc)
- Handle state transitions between modes
- Render views for each mode

#### `internal/git` (Git State Management)
**Purpose**: Detects and manages git repository state; executes git commands

**Key Components**:
- `state.go` - Core state detection (WorkingTree, Timeline, Operation, Remote)
- `types.go` - State type definitions
- `execute.go` - Git command execution
- `environment.go` - Git/SSH environment readiness checks
- `init.go` - Repository initialization
- `messages.go` - Git operation result types
- `ssh.go` - SSH key generation and management
- `dirtyop.go` - Dirty operation state tracking

**State Detection Hierarchy** (Priority order):
1. **GitEnvironment**: Ready, NeedsSetup, MissingGit, MissingSSH
2. **Operation**: NotRepo, Normal, Conflicted, Merging, Rebasing, DirtyOperation, TimeTraveling, Rewinding
3. **WorkingTree**: Clean, Dirty
4. **Timeline**: InSync, Ahead, Behind, Diverged (only when Remote=HasRemote)
5. **Remote**: NoRemote, HasRemote

**Responsibilities**:
- Detect git repository state with multi-level priority system
- Execute git commands safely
- Track incomplete operations (time travel, dirty state)
- Manage SSH key generation and environment setup
- Monitor timeline synchronization (local vs remote)

#### `internal/ui` (UI Rendering)
**Purpose**: Renders all terminal UI components

**Key Components**:
- `layout.go` - Reactive layout engine (header/content/footer)
- `sizing.go` - Dynamic sizing calculations
- `menu.go` - Menu rendering
- `header.go` - State header rendering (5 rows)
- `console.go` - Git command output console
- `history.go` - Commit history split-pane view
- `filehistory.go` - File(s) history split-pane view
- `conflictresolver.go` - 3-way conflict resolution UI
- `input.go` - Text input components
- `theme.go` - Color theme system
- `buffer.go` - Output line buffering
- `spinner.go` - Loading indicator animation

**Rendering Components**:
- `RenderMenuWithBanner()` - Menu with SVG banner
- `RenderConsoleOutput()` - Streaming git output
- `RenderHistorySplitPane()` - Commit history browser
- `RenderFileHistorySplitPane()` - File history browser
- `RenderConflictResolveGeneric()` - Merge conflict resolver
- `RenderReactiveLayout()` - Full page with header/content/footer

**Responsibilities**:
- Render terminal UI components with consistent styling
- Handle dynamic terminal resizing
- Display streaming git operation output
- Implement split-pane navigation
- Support color themes

#### `internal/config`
**Purpose**: Persistent configuration management

**Key Components**:
- `stash.go` - Stash state tracking for dirty operations

#### `internal/banner`
**Purpose**: ASCII art rendering

**Key Components**:
- `svg.go` - SVG to terminal rendering
- `braille.go` - Braille character patterns

---

## 2. APPLICATION STATE MANAGEMENT

### Main Application Struct

```go
type Application struct {
    // Core dimensions
    width, height int
    sizing ui.DynamicSizing
    theme ui.Theme
    
    // Current mode
    mode AppMode
    
    // Git state (master)
    gitState *git.State
    gitEnvironment git.GitEnvironment
    
    // Menu state
    selectedIndex int
    menuItems []MenuItem
    keyHandlers map[AppMode]map[string]KeyHandler
    
    // Input mode state
    inputPrompt, inputValue string
    inputCursorPosition int
    inputValidationMsg string
    
    // Clone workflow state
    cloneURL, clonePath string
    cloneBranches []string
    
    // Async operations
    asyncOperationActive bool
    asyncOperationAborted bool
    isExitAllowed bool
    previousMode AppMode
    
    // Console/output state
    consoleState ui.ConsoleOutState
    outputBuffer *ui.OutputBuffer
    
    // Conflict resolution
    conflictResolveState *ConflictResolveState
    
    // History modes
    historyState *ui.HistoryState
    fileHistoryState *ui.FileHistoryState
    
    // Time travel
    timeTravelInfo *git.TimeTravelInfo
    
    // Caches
    historyMetadataCache map[string]*git.CommitDetails
    fileHistoryDiffCache map[string]string
    cacheLoadingStarted bool
}
```

### State Initialization Flow

```
main.go
  ↓
NewApplication()
  ├── DetectGitEnvironment()
  │   └── If not Ready → Show setup wizard
  ├── IsInitializedRepo() / HasParentRepo()
  │   └── Find and cd into repo
  ├── DetectState()
  │   ├── Check for DirtyOperation
  │   ├── Detect WorkingTree (Clean/Dirty)
  │   ├── Detect Operation (Normal/Conflicted/etc)
  │   ├── Detect Remote (HasRemote/NoRemote)
  │   ├── Detect Timeline (if Remote && !Detached)
  │   └── Get branch name and commit hashes
  ├── GenerateMenu()
  │   └── Create menu items based on git state
  ├── Build cache metadata
  └── Return initialized Application
```

---

## 3. APPLICATION MODES (14 Total)

### Mode Enumeration with Metadata

```go
const (
    ModeMenu              = iota  // Main state-driven menu
    ModeInput                     // Generic text input (deprecated)
    ModeConsole                   // Streaming git output
    ModeConfirmation              // Yes/No dialog
    ModeHistory                   // Commit history browser
    ModeConflictResolve           // 3-way conflict resolver
    ModeInitializeLocation        // init: choose location
    ModeInitializeBranches        // init: choose branch names
    ModeCloneURL                  // clone: input URL
    ModeCloneLocation             // clone: choose location
    ModeClone                     // clone: async operation
    ModeSelectBranch              // clone: select canon branch
    ModeFileHistory               // File history browser
    ModeSetupWizard               // Git environment setup
)
```

### Mode Metadata (ModeDescriptions Map)

Each mode has:
- **Name**: String representation (e.g., "menu", "history")
- **Description**: Human-readable purpose
- **AcceptsInput**: Whether mode handles keyboard input
- **IsAsync**: Whether mode is blocking/async

**Input-Accepting Modes**:
- ModeMenu
- ModeInput
- ModeConsole
- ModeConfirmation
- ModeHistory
- ModeFileHistory
- ModeConflictResolve
- ModeInitializeLocation / ModeInitializeBranches
- ModeCloneURL / ModeCloneLocation
- ModeSelectBranch
- ModeSetupWizard

**Async Modes**:
- ModeConsole (git output streaming)
- ModeClone (clone operation)

### Setup Wizard Steps (5 Steps)

```go
const (
    SetupStepWelcome SetupWizardStep = iota  // Welcome message
    SetupStepPrerequisites                    // Check git + ssh
    SetupStepEmail                            // Input email
    SetupStepGenerate                         // Generate SSH key + agent
    SetupStepDisplayKey                       // Show public key
    SetupStepComplete                         // Setup complete
)
```

---

## 4. UI COMPONENTS & SIZING

### Major UI Components

**Layout Components**:
- `RenderReactiveLayout()` - Full page layout with header/content/footer
- `RenderMenuWithBanner()` - Menu with ASCII banner (2-column)
- `RenderConsoleOutput()` - Streaming git output with spinner

**Interactive Components**:
- `RenderTextInput()` - Single/multiline text input
- `RenderHistorySplitPane()` - Commit list + details pane
- `RenderFileHistorySplitPane()` - Files + diff pane
- `RenderConflictResolveGeneric()` - N-way conflict resolver
- `RenderConfirmationDialog()` - Yes/No confirmation

**Display Components**:
- `RenderHeader()` - State information header (5 rows)
- `RenderStatusBar()` - Footer status bar
- `RenderBox()` - Box drawing with borders

### Dynamic Sizing System

```go
type DynamicSizing struct {
    TerminalWidth     int
    TerminalHeight    int
    ContentHeight     int
    ContentInnerWidth int
    HeaderInnerWidth  int
    FooterInnerWidth  int
    MenuColumnWidth   int
    IsTooSmall        bool
}

// Sizing Constants
MinWidth = 69, MinHeight = 19
HeaderHeight = 9, FooterHeight = 1
HorizontalMargin = 2, BannerWidth = 30
```

**Reactive Behavior**:
- On `tea.WindowSizeMsg`: Update sizing and recalculate all dimensions
- Menu column width auto-adjusts: `contentInnerWidth - bannerWidth - 2`
- Content height auto-adjusts: `termHeight - headerHeight - footerHeight`
- Raises "terminal too small" warning if width < 69 or height < 19

### State Header Layout (5 Rows)

```
Row 1: CWD (left) | OPERATION emoji + label (right)
Row 2: Remote URL (left) | BRANCH emoji + label (right)
Row 3: WorkingTree emoji + status (left) | [spacer] (right)
Row 4: Timeline emoji + status (left) | [spacer] (right)
Row 5: [separator line]
```

---

## 5. GIT STATE DETECTION MECHANISM

### Multi-Axis State Detection

TIT models git state as a **5-axis tuple**:

```
(GitEnvironment, Operation, WorkingTree, Timeline, Remote)
```

### Priority Order (Checked Sequentially)

**Axis 0 - GitEnvironment** (Highest priority)
```
1. Check if git installed
2. Check if ssh installed
3. Check if SSH key exists in ~/.ssh
4. Return: Ready | NeedsSetup | MissingGit | MissingSSH
```

If GitEnvironment ≠ Ready:
- Show Setup Wizard, exit other state detection

**Axis 1 - Operation** (Highest in repo context)
```
Priority:
1. Check for conflicts (git status --porcelain=v2, lines starting with "u ")
2. Check for time traveling (.git/TIT_TIME_TRAVEL file)
3. Check for merge (MERGE_HEAD)
4. Check for rebase (rebase-merge, rebase-apply)
Result: Conflicted | TimeTraveling | Merging | Rebasing | Normal | NotRepo
```

**Axis 2 - WorkingTree**
```
Run: git status --porcelain=v2
- Empty output → Clean
- Lines starting with '1', '2', or '?' → Dirty
Result: Clean | Dirty
```

**Axis 3 - Remote**
```
Run: git remote
- No remotes → NoRemote
- Has remotes → HasRemote
Result: NoRemote | HasRemote
```

**Axis 4 - Timeline** (Conditional)
```
Only checked if: Operation == Normal && Remote == HasRemote && hasCommits

Compares local vs remote tracking:
1. Try @{u} (upstream tracking branch)
2. Fallback to refs/remotes/origin/[branch] if no upstream

Run: git rev-list --left-right --count HEAD...@{u}
- ahead > 0, behind == 0 → Ahead
- ahead == 0, behind > 0 → Behind
- ahead > 0, behind > 0 → Diverged
- ahead == 0, behind == 0 → InSync
Result: InSync | Ahead | Behind | Diverged | (empty if N/A)
```

### State Usage in Menu Generation

The menu generator uses this priority:
```
if Operation == NotRepo
  → Show: Init, Clone
  
if Operation == TimeTraveling
  → Show: History (time travel variant), File History, Rewind, Return
  
if Operation == Conflicted
  → Lock down: Show only conflict resolution menu
  
if Operation == Normal
  → Generate based on WorkingTree + Timeline + Remote
```

### Dirty Operation State

Tracks incomplete operations (merge/rebase in progress) via `.git/TIT_DIRTY_OP` file:
```
Contains: operation type + state snapshot
Prevents: Conflicting operations
```

### Time Travel State Tracking

Time travel uses `.git/TIT_TIME_TRAVEL` marker file:
```
Format:
  Line 1: Original branch name (e.g., "main")
  Line 2: Original stash ID (if user had uncommitted work)

Detection:
  1. On app init: if marker exists but Operation ≠ TimeTraveling
     → Perform restoration (discard changes, return to branch, reapply work)
  2. During app: if marker exists and Operation == TimeTraveling
     → Load metadata for detached HEAD display
```

---

## 6. MENU SYSTEM

### Menu Generation Pipeline

```
Application.GenerateMenu()
  ↓
Select generator based on gitState.Operation
  ├── menuNotRepo()           // NotRepo state
  ├── menuNormal()            // Normal state
  └── menuTimeTraveling()     // TimeTraveling state
    ↓
Build MenuItem[] from git state
  ├── Working tree section (if Dirty)
  ├── Timeline section (if HasRemote)
  ├── History section
  └── Remote section (if NoRemote)
    ↓
Set Enabled flag based on cache status
  └── Disable history items while cache building
    ↓
Return MenuItem[]
```

### Menu Item Structure

```go
type MenuItem struct {
    ID        string  // Unique action identifier
    Shortcut  string  // Single keyboard shortcut
    Emoji     string  // Leading emoji
    Label     string  // Max 21 chars (for fixed-width menu)
    Hint      string  // Shown in footer on focus
    Enabled   bool    // Can be selected (disabled during cache build)
    Separator bool    // If true, visual separator only
}
```

### MenuItems Map (Single Source of Truth)

Defined in `menu_items.go`, contains 30+ items:
- **NotRepo**: init, clone
- **Working Tree**: commit, commit_push
- **Timeline**: push, force_push, pull_merge, replace_local, etc
- **History**: history, file_history
- **Remote**: add_remote
- **Time Travel**: time_travel_history, time_travel_files_history, rewind, return_from_timetravel

### Menu Rendering

```
RenderMenuWithBanner(sizing, items[], selectedIndex, theme)
  ├── Left column (width: menuColumnWidth)
  │   └── Render menu items with highlight
  └── Right column (width: 30)
      └── Render ASCII banner (TIT logo)
```

### Keyboard Shortcuts

Dynamically registered in `buildKeyHandlers()`:
```
Global handlers (all modes):
  - ctrl+c / q  → Quit with confirmation
  - esc         → Return to menu or parent mode
  - ctrl+v      → Paste mode trigger

Mode-specific handlers:
  - ModeMenu: up/down (j/k), enter, [1-9] for shortcuts
  - ModeInput: left/right, backspace, enter
  - ModeHistory: up/down (j/k), tab, enter, ctrl+r (rewind)
  - ModeFileHistory: up/down (j/k), tab, v (visual mode), y (copy)
  - ModeConflictResolve: up/down (j/k), tab, space, enter
```

---

## 7. EVENT FLOW & MESSAGE TYPES

### Update Loop Flow

```
tea.Update(msg)
  ↓
Switch on message type:
  ├── tea.WindowSizeMsg
  │   └── Update sizing, recalculate dimensions
  ├── tea.KeyMsg
  │   ├── Check if paste mode
  │   ├── Look up handler in keyHandlers[mode][keyStr]
  │   ├── Route to handler function
  │   └── Handler returns (model, cmd)
  ├── GitOperationMsg
  │   └── Handle result (clone, push, pull, etc)
  ├── CacheProgressMsg
  │   └── Update cache state, regenerate menu
  ├── OutputRefreshMsg
  │   └── Re-render console, schedule next tick
  ├── RestoreTimeTravelMsg
  │   └── Restore from incomplete time travel
  ├── SetupCompleteMsg
  │   └── Advance setup wizard step
  └── [Other operation-specific messages]
    ↓
Return updated Application + optional Command
```

### Message Types (in messages.go)

```go
type GitOperationMsg struct {
    OpType string  // "clone", "push", "pull", etc
    Success bool
    Output string
    Error string
}

type CacheProgressMsg struct {
    Type string      // "metadata", "diffs"
    Progress int
    Total int
    Complete bool
}

type OutputRefreshMsg struct {}  // Tick for console refresh

type RestoreTimeTravelMsg struct {
    Success bool
    Error string
}

type SetupCompleteMsg struct {
    Step string
    Data map[string]string
}

// ... and 20+ others
```

### Handler Chain

```
Update()
  ↓
Dispatch via mode handler map
  ↓
Specific handler (e.g., handleMenuEnter)
  ├── Validate state
  ├── Execute action
  ├── Update application state
  └── Return (app, cmd)
    ↓
Command execution (if any):
  └── Runs async git operation
    └── Returns result as message
      └── Next Update() processes result
```

---

## 8. KEY FILES & PURPOSES

### Application Logic (internal/app/)

| File | Lines | Purpose |
|------|-------|---------|
| app.go | 1407 | Main Application struct, lifecycle, rendering |
| modes.go | 162 | AppMode enumeration with metadata |
| menu.go | 300+ | Menu generation based on git state |
| menu_items.go | 200+ | Menu item definitions |
| handlers.go | 500+ | Event handlers for user actions |
| keyboard.go | 400+ | Key binding management |
| git_handlers.go | 600+ | Git operation handlers |
| messages.go | 200+ | Bubble Tea message types |
| state_info.go | 150+ | State display formatting |
| history_cache.go | 300+ | History metadata caching |
| dispatchers.go | 300+ | Action dispatchers |
| operations.go | 400+ | Long-running operation management |

### Git Operations (internal/git/)

| File | Lines | Purpose |
|------|-------|---------|
| state.go | 507 | Core state detection logic |
| types.go | 110 | State type definitions |
| execute.go | 200+ | Git command execution |
| environment.go | 150+ | Git/SSH environment checks |
| init.go | 200+ | Repository initialization |
| ssh.go | 300+ | SSH key generation |
| messages.go | 100+ | Git operation message types |

### UI Components (internal/ui/)

| File | Lines | Purpose |
|------|-------|---------|
| layout.go | 200+ | Reactive layout engine |
| sizing.go | 64 | Dynamic sizing calculations |
| menu.go | 250+ | Menu rendering |
| header.go | 200+ | State header rendering |
| console.go | 300+ | Git output console |
| history.go | 400+ | Commit history browser |
| filehistory.go | 400+ | File history browser |
| conflictresolver.go | 400+ | Conflict resolution UI |
| theme.go | 300+ | Color theme system |
| buffer.go | 150+ | Output line buffering |
| input.go | 200+ | Text input components |

### Configuration (internal/config/)

| File | Lines | Purpose |
|------|-------|---------|
| stash.go | 100+ | Stash state tracking |

---

## 9. CURRENT LIMITATIONS & ARCHITECTURAL NOTES

### Known Limitations

1. **ModeInput is deprecated**
   - Being phased out in favor of mode-specific input modes
   - Still used as fallback for some operations

2. **Cache building blocks history menus**
   - History/FileHistory disabled until caches built
   - First startup has ~5-10s delay on large repos

3. **Git environment setup not fully integrated**
   - Setup wizard exists but missing provider URLs step
   - SSH key generation works, but user education minimal

4. **Time travel restoration timing**
   - Restoration happens on startup if incomplete session detected
   - No user confirmation (automatic, could lose work)

5. **Conflict resolver not tested with cherry-pick/rebase**
   - Built for merge conflicts specifically
   - Other operation conflicts may not display correctly

6. **Terminal too small handling**
   - App locks up if terminal resized below minimum
   - Needs graceful degradation or min-window enforcement

### Architectural Decisions

1. **5-Axis State Model**
   - Cleanly separates concerns (git env vs repo state vs operation)
   - Priority order prevents invalid state combinations
   - Hard to extend with new axes (refactor needed)

2. **Single-Source-of-Truth for Menu Items**
   - MenuItems map in menu_items.go is canonical
   - Prevents inconsistencies between rendering and handlers
   - Makes updates centralized and easy to track

3. **State-Driven Menu Generation**
   - Menu changes automatically when git state changes
   - No manual menu rebuilding needed
   - Cache status integrated (disables items during build)

4. **Async Operations via Messages**
   - Git operations run in goroutines
   - Results returned as Bubble Tea messages
   - Prevents UI freezing during long operations

5. **DynamicSizing for Terminal Responsiveness**
   - All dimensions recomputed on WindowSizeMsg
   - Enables responsive layout without complex CSS
   - Menu column width auto-adjusts based on banner

6. **Split-Pane Navigation**
   - History/FileHistory use full terminal (no header/footer)
   - Tab key switches focused pane
   - Cursor position tracked per pane

### Design Patterns Used

1. **Builder Pattern**: MenuItem, ModeHandlers, ConfirmationConfig
2. **Visitor Pattern**: MenuGenerators for different Operation states
3. **Handler Registry**: keyHandlers map for dynamic key binding
4. **State Machine**: Mode transitions with automatic cleanup
5. **Observer**: OutputRefreshMsg ticks for async operation progress
6. **Repository**: GitState as single source of git truth

### Thread Safety

- **Mutex Protected**: historyCacheMutex, fileHistoryCacheMutex, diffCacheMutex
- **Async Operations**: Use channels and goroutines
- **UI Thread**: Bubble Tea single-threaded event loop
- **Git Execution**: Each command spawned in separate process (thread-safe)

---

## 10. COMPLETE LIST OF MODES

| Mode | Purpose | Accepts Input | Async | Typical Transitions |
|------|---------|---------------|-------|-------------------|
| ModeMenu | Main state-driven menu | Yes | No | → Input, Console, History, Confirmation, SetupWizard |
| ModeInput | Generic text input (deprecated) | Yes | No | → Menu |
| ModeConsole | Streaming git output | Yes (ESC) | Yes | → Menu |
| ModeConfirmation | Yes/No confirmation | Yes | No | → Menu, Console, History |
| ModeHistory | Commit history browser | Yes | No | → Menu, Console |
| ModeConflictResolve | 3-way conflict resolver | Yes | No | → Menu |
| ModeInitializeLocation | Choose: init here or subdir | Yes | No | → InitializeBranches |
| ModeInitializeBranches | Input branch names | Yes | No | → Console |
| ModeCloneURL | Input clone URL | Yes | No | → CloneLocation |
| ModeCloneLocation | Choose: clone here or subdir | Yes | No | → Console |
| ModeClone | Clone operation (async) | Yes (ESC) | Yes | → SelectBranch |
| ModeSelectBranch | Select canon branch | Yes | No | → Console |
| ModeFileHistory | File(s) history browser | Yes | No | → Menu |
| ModeSetupWizard | Git environment setup | Yes | No | → Menu |

---

## SUMMARY: DATA FLOW OVERVIEW

```
START
  ↓
main.go: Initialize theme, create app
  ↓
NewApplication()
  ├── DetectGitEnvironment()
  │   └── If not Ready → Setup Wizard
  ├── FindRepository()
  ├── DetectState() → (Operation, WorkingTree, Timeline, Remote)
  └── GenerateMenu() → MenuItems[]
  ↓
tea.NewProgram(app).Run()
  ├── Init() → Load cache, fetch remote
  ├── View() → Render current mode
  └── Update() → Handle events
    ├── KeyMsg → keyHandlers[mode][key] → Handler()
    ├── GitOperationMsg → handleGitOperation()
    ├── CacheProgressMsg → updateMenu()
    └── WindowSizeMsg → recalculateSizing()
      ↓
Handler executes action
  ├── Update Application state
  ├── Return tea.Cmd if async
  └── Next Update processes result
    ↓
View() renders new mode
  ├── Header: State info (git state, branch, remote)
  ├── Content: Mode-specific rendering
  └── Footer: Action hints
    ↓
REPEAT until user quits
```

---

## FILE ORGANIZATION SUMMARY

### Internal Package File Count

| Package | File Count | Total Lines |
|---------|-----------|------------|
| internal/app | 26 files | ~6000 |
| internal/ui | 18 files | ~5000 |
| internal/git | 8 files | ~2500 |
| internal/config | 1 file | ~100 |
| internal/banner | 2 files | ~500 |
| cmd | 1 file | ~22 |
| **TOTAL** | **56 files** | **~15,200** |

---

**Generated**: January 22, 2026  
**Last Updated**: From TIT codebase analysis
