# Missing Components Implementation Plan

## Overview

New-TIT aims to have identical flow with old-TIT but with better organization and structure. This document maps missing components and the organized structure new-TIT should adopt.

---

## Structure Comparison: Old-TIT vs New-TIT

### Old-TIT (`internal/app/` - 18 files)
```
app.go                         # Main app struct
appinitializers.go            # Initialization logic
cache.go                       # Cache manager
cachepreload.go               # Cache preload logic
confirmationhandlers.go       # Confirmation dialogs
conflicthandlers.go           # Conflict resolution handlers
conflictstate.go              # Conflict state tracking
diffpane_handlers.go          # Diff pane mode handlers
dispatchers.go                # Action dispatchers
githandlers.go                # Git operation result handlers
inputhandlers.go              # Input handling
keyboard.go                   # Key handler registry
menu.go                        # Menu generation
messages.go                   # Message types & constants
modehandlers_history.go       # History browsing mode
modehandlers_filehistory.go   # File history mode
operations.go                 # Git async operations
rendering.go                  # View rendering
```

### New-TIT (`internal/app/` - 17 files, but more granular)
```
app.go                         # Main app struct ✅
async.go                      # Async handling (extracted)
config.go                      # Config management
cursormovement.go             # Cursor utilities (extracted)
dispatchers.go                # Action dispatchers ✅
githandlers.go                # Git operation handlers ✅
handlers.go                   # Input handlers (combined)
keyboard.go                   # Key handler registry ✅
keybuilder.go                 # Key handler building (extracted)
location.go                   # Location selection (extracted)
menu.go                        # Menu generation ✅
menubuilder.go                # Menu building (extracted)
messages.go                   # Message types ✅
modes.go                      # Mode definitions
operations.go                 # Git async operations ✅
stateinfo.go                  # State display info (NEW - good!)
```

### New-TIT (`internal/ui/` - 14 files vs old-tit 14 files)
```
assets/                        # SVG logos ✅
buffer.go                      # Output buffer ✅
box.go                         # Box drawing (extracted)
branchinput.go                # Branch selection (NEW - good!)
console.go                     # Console output ✅
formatters.go                 # Text formatters (extracted)
input.go                       # Input fields (extracted)
layout.go                      # Main layout ✅
menu.go                        # Menu rendering ✅
sizing.go                      # Constants (extracted)
textinput.go                   # Text input ✅
theme.go                       # Theme system ✅
validation.go                 # Input validation (extracted)
```

**Assessment:** New-TIT is actually MORE organized! Better extraction of concerns:
- ✅ Utilities extracted (`cursormovement.go`, `formatters.go`, `sizing.go`, `validation.go`)
- ✅ New abstractions (`stateinfo.go`, `branchinput.go`)
- ❌ Missing: `confirmation.go`, `conflictstate.go`, `cache.go`

---

## Missing Components by Priority

### 🔴 PHASE 2 (Immediate - Core Functionality)

#### 1. **ConfirmationDialog System** (`internal/ui/confirmation.go`)
```go
// Location: internal/ui/confirmation.go
type ConfirmationConfig struct {
    Title       string
    Explanation string
    YesLabel    string
    NoLabel     string
    ActionID    string
}

type ConfirmationDialog struct {
    config  ConfirmationConfig
    width   int
    theme   *Theme
}

// Renders a centered confirmation dialog with explanation
func (cd *ConfirmationDialog) View() string { ... }
```

**Use cases:**
- Nested repo warning during init
- Force push confirmation
- Abort operation confirmation
- Branch deletion confirmation

**Flow:**
```
User selects action → Set mode=ModeConfirmation
                    → Show ConfirmationDialog
                    → User presses Y/N
                    → Handle confirmationhandlers.go
```

**Implementation steps:**
1. Create `internal/ui/confirmation.go` with `ConfirmationDialog`
2. Create `internal/app/confirmationhandlers.go` for confirmation logic
3. Add `ModeConfirmation` to `modes.go`
4. Add confirmation-related message types to `messages.go`
5. Wire handlers in `keyboard.go`

**Files to port from old-tit:**
- `internal/ui/confirmation.go`
- `internal/app/confirmationhandlers.go`

---

#### 2. **Conflict State Tracking** (`internal/app/conflictstate.go`)
```go
// Location: internal/app/conflictstate.go
type ConflictedFile struct {
    Path           string
    Status         string
    ConflictType   string  // "ours", "theirs", "both"
    OursSide       string
    TheirsSide     string
    BaseSide       string
}

type ConflictState struct {
    OperationType  string         // "merge", "rebase", "cherry-pick"
    ConflictedFiles []ConflictedFile
    TotalFiles     int
    ResolvedCount  int
}
```

**Use cases:**
- Tracking which files have conflicts
- 3-way conflict resolution
- Conflict status display

**Implementation steps:**
1. Create `internal/app/conflictstate.go`
2. Add conflict detection to `git/state.go`
3. Create `internal/app/conflicthandlers.go`
4. Wire to keyboard handlers

**Files to port from old-tit:**
- `internal/app/conflictstate.go`
- `internal/app/conflicthandlers.go`

---

### 🟡 PHASE 3-4 (History & Advanced Features)

#### 3. **History Mode Handlers** (`internal/app/modehandlers_history.go`)
```go
// Location: internal/app/modehandlers_history.go
// Handles:
// - Commit list browsing
// - Cherry-pick selection
// - Time travel (detached HEAD exploration)

type HistoryState struct {
    CommitList  []CommitInfo
    SelectedIdx int
    ScrollPos   int
}
```

**Use cases:**
- Browse commit history
- Cherry-pick commits
- Time travel (read-only exploration)
- Show commit details

**Implementation steps:**
1. Create `internal/app/modehandlers_history.go`
2. Create `internal/ui/listpane.go` (reusable list component)
3. Create `internal/ui/history.go` (history view)
4. Add `ModeHistory` mode
5. Implement commit caching with `CacheManager`

**Files to port from old-tit:**
- `internal/app/modehandlers_history.go`
- `internal/ui/history.go`
- `internal/ui/listpane.go`

---

#### 4. **File History & Diff Pane** (`internal/app/modehandlers_filehistory.go`)
```go
// Location: internal/app/modehandlers_filehistory.go
// Handles:
// - Per-file change history
// - Side-by-side diff view
// - Syntax highlighting
```

**Use cases:**
- View changes to specific file
- Review patches
- Side-by-side comparison

**Implementation steps:**
1. Create `internal/app/modehandlers_filehistory.go`
2. Create `internal/ui/diffpane.go` (side-by-side diff)
3. Create `internal/ui/filehistory.go` (file history view)
4. Add `ModeFileHistory` and `ModeDiffPane` modes
5. Implement diff caching

**Files to port from old-tit:**
- `internal/app/modehandlers_filehistory.go`
- `internal/app/diffpane_handlers.go`
- `internal/ui/diffpane.go`
- `internal/ui/filehistory.go`

---

### 🟢 PHASE 5+ (Polish & Optimization)

#### 5. **Parallel Cache System** (`internal/app/cache.go`)
```go
// Location: internal/app/cache.go
type CacheManager struct {
    config      CacheConfig
    cache       map[string]interface{}
    itemKeys    []string
    worker      CacheWorker
}

// Parallel loading of history, file lists, diffs
// Progress reporting with CacheProgressMsg
```

**Use cases:**
- Preload commit history in background
- Cache file lists
- Cache diff results

**Implementation steps:**
1. Create `internal/app/cache.go` with `CacheManager`
2. Create `internal/app/cachepreload.go` for preload logic
3. Add cache manager to `Application` struct
4. Implement progress messages

**Files to port from old-tit:**
- `internal/app/cache.go`
- `internal/app/cachepreload.go`

---

#### 6. **Rendering Helpers** (`internal/app/rendering.go`)
```go
// Location: internal/app/rendering.go
// Utility functions for complex View() composition
// This keeps app.go View() method clean
```

**Implementation steps:**
1. Create `internal/app/rendering.go`
2. Extract complex rendering logic from `app.go` View()
3. Keep `stateinfo.go` pattern (state → display mapping)

**Files to port from old-tit:**
- `internal/app/rendering.go` (selectively)

---

## File Organization Checklist

### `internal/app/` Structure
```
Core:
- ✅ app.go              # Main Application struct, Update(), View()
- ✅ modes.go            # AppMode enum and constants
- ✅ messages.go         # Message types, FooterMessageType map

Dispatching & Handling:
- ✅ dispatchers.go      # Action dispatchers (menu → mode)
- ✅ keyboard.go         # Key handler registry + builders
- ✅ handlers.go         # Input submission handlers
- ✅ githandlers.go      # Git operation result handlers
- ❌ confirmationhandlers.go   # Confirmation result handlers (MISSING)
- ❌ conflicthandlers.go       # Conflict resolution handlers (MISSING)

Operations:
- ✅ operations.go       # Async git commands (cmd*)
- ✅ githandlers.go      # Operation chaining

Advanced Modes:
- ❌ modehandlers_history.go        # History browsing (MISSING)
- ❌ modehandlers_filehistory.go    # File history (MISSING)
- ❌ diffpane_handlers.go           # Diff pane (MISSING)

State & Config:
- ✅ config.go           # Repo config
- ✅ stateinfo.go        # State → display mapping (NEW!)
- ❌ conflictstate.go    # Conflict tracking (MISSING)

Utilities:
- ✅ async.go            # Async helpers (already extracted!)
- ✅ cursormovement.go   # Cursor movement
- ✅ location.go         # Location selection
- ✅ menubuilder.go      # Menu building
- ✅ keybuilder.go       # Key handler building

Optimization:
- ❌ cache.go            # Parallel cache manager (MISSING)
- ❌ cachepreload.go     # Cache preload logic (MISSING)
- ❌ rendering.go        # Rendering helpers (MISSING)
```

### `internal/ui/` Structure
```
Core:
- ✅ layout.go           # Main layout container
- ✅ theme.go            # Theme system
- ✅ sizing.go           # Constants (EXTRACTED - good!)

Rendering:
- ✅ box.go              # Box drawing (EXTRACTED - good!)
- ✅ textinput.go        # Text input rendering
- ✅ branchinput.go      # Branch selection (NEW!)
- ✅ menu.go             # Menu rendering
- ✅ console.go          # Console output

State Display:
- ✅ buffer.go           # Output buffer
- ✅ formatters.go       # Text formatting (EXTRACTED - good!)
- ✅ input.go            # Input fields (EXTRACTED - good!)
- ✅ validation.go       # Validation (EXTRACTED - good!)

Advanced Components:
- ❌ confirmation.go     # Confirmation dialog (MISSING)
- ❌ conflictresolve.go  # Conflict UI (MISSING)
- ❌ listpane.go         # Reusable list (MISSING)
- ❌ history.go          # History view (MISSING)
- ❌ filehistory.go      # File history view (MISSING)
- ❌ diffpane.go         # Diff view (MISSING)

Assets:
- ✅ assets/             # SVG logos
```

---

## Implementation Roadmap

### Current Phase (Session 24): ✅ Complete
- ✅ Git operations (cmd* pattern)
- ✅ Operation chaining (add_remote → fetch → set_upstream)
- ✅ State header with emoji display
- ✅ Footer message map

### Phase 2 (Next): 🔴 HIGH PRIORITY
**Goal:** Complete core operations with safety confirmations
1. **ConfirmationDialog** (2-3 hours)
   - Port `internal/ui/confirmation.go`
   - Port `internal/app/confirmationhandlers.go`
   - Add nested repo warning to init flow
   - Test confirmation flow

2. **Conflict State** (2-3 hours)
   - Port `internal/app/conflictstate.go`
   - Port `internal/app/conflicthandlers.go`
   - Integrate with conflict detection
   - Test conflict resolution flow

### Phase 3: 🟡 MEDIUM PRIORITY
**Goal:** Add history browsing and time travel
1. **History Mode** (4-5 hours)
   - Create `internal/app/modehandlers_history.go`
   - Create `internal/ui/listpane.go` (reusable!)
   - Create `internal/ui/history.go`
   - Test history browsing

### Phase 4: 🟡 MEDIUM PRIORITY
**Goal:** File-level diff viewing
1. **File History Mode** (4-5 hours)
   - Create `internal/app/modehandlers_filehistory.go`
   - Create `internal/ui/diffpane.go`
   - Create `internal/ui/filehistory.go`
   - Test file history + diff

### Phase 5: 🟢 LOW PRIORITY
**Goal:** Performance optimization
1. **Cache System** (2-3 hours)
   - Port `internal/app/cache.go`
   - Port `internal/app/cachepreload.go`
   - Integrate with history preloading

---

## Design Principles (New-TIT Improvements)

### ✅ New-TIT Already Does Better
1. **Extracted Utilities** - `sizing.go`, `formatters.go`, `validation.go`
2. **State Display Maps** - `stateinfo.go` pattern (state → display)
3. **Cleaner Separation** - `confirmationhandlers.go` separate from `dispatchers.go`
4. **Better Naming** - `modehandlers_history.go` vs generic handler names

### 🎯 Continue This Pattern
- Keep handler types distinct: `ActionHandler`, `ConfirmationHandler`, `ConflictHandler`
- Map types: `FooterMessageType`, `ConfirmationType`, `ConflictType`
- UI components: `confirmation.go`, `listpane.go`, `diffpane.go` (focused, single-purpose)
- Mode handlers: `modehandlers_*.go` (one file per major mode)
- Utilities: Extract to separate files if >50 lines (`async.go`, `cursormovement.go`)

### 📦 Module Organization
```
internal/app/
├── Core (app.go, modes.go, messages.go)
├── Dispatching (dispatchers.go, keyboard.go, handlers.go)
├── Operations (operations.go, githandlers.go)
├── Confirmations (confirmationhandlers.go)
├── Conflict Resolution (conflicthandlers.go, conflictstate.go)
├── History Modes (modehandlers_history.go, modehandlers_filehistory.go)
├── Caching (cache.go, cachepreload.go)
└── Utilities (async.go, cursormovement.go, location.go, etc.)

internal/ui/
├── Core (layout.go, theme.go, sizing.go)
├── Components (textinput.go, branchinput.go, menu.go, console.go)
├── Advanced (confirmation.go, conflictresolve.go, listpane.go, diffpane.go, history.go, filehistory.go)
└── Utilities (box.go, buffer.go, formatters.go, validation.go, input.go)
```

---

## Verification Checklist

When porting each component:
- [ ] Read source from old-tit thoroughly
- [ ] Identify SSOT (git state types, theme colors, sizing)
- [ ] Use existing abstractions (FooterMessageType, StateInfo pattern)
- [ ] Extract utilities if >50 lines
- [ ] Test the specific flow manually
- [ ] No hardcoded strings (use maps)
- [ ] No duplicated logic (use helpers)
- [ ] Proper thread safety (closures for async)
