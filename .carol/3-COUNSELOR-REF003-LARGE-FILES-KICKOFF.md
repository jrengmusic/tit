# Sprint 3: REF-003 Large Handler Files - Comprehensive Kickoff Plan

**Date:** 2026-01-30
**Objective:** Address AUDITOR's REF-003 finding - reduce `internal/app/` from 11,176 lines to manageable size
**Approach:** 6-phase incremental file splitting with clean build verification
**Estimated Duration:** 6-8 hours total (1-1.5 hours per phase)

---

## Pre-Flight: Current State Analysis

### Current Metrics (Post-Sprint 2):
- **Total Files:** 57 `.go` files in `internal/app/`
- **Total Lines:** ~11,176 lines
- **Largest File:** `app.go` - ~1,771 lines
- **Second Largest:** `confirmation_handlers.go` - ~972 lines (48 methods)

### Problem Areas Identified:
1. **`app.go` (1,771 lines)** - Contains Update(), View(), Init(), 60+ delegation methods
2. **`confirmation_handlers.go` (972 lines)** - 48 confirmation-related methods
3. **Handler Sprawl** - 100+ Application methods scattered across files
4. **No Sub-Packages** - Everything in flat `internal/app/` structure

---

## Phase 1: Extract Core Bubble Tea Methods from app.go

**Goal:** Split `app.go` into focused files by Bubble Tea interface
**Files to Create:** 3 new files
**Lines to Move:** ~600 lines from `app.go`
**Estimated Time:** 90 minutes

### Implementation:

**Step 1: Create `internal/app/app_update.go`**
Move the Update() method and all message handlers:
```go
package app

import tea "github.com/charmbracelet/bubbletea"

// Update handles all messages and routes to appropriate handlers.
// This is the core Bubble Tea Update method.
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Move entire Update() method here (lines 571-727 from app.go)
}

// Message handler helpers that are only used by Update()
func (a *Application) handleCacheProgress(msg CacheProgressMsg) (tea.Model, tea.Cmd) { ... }
func (a *Application) handleCacheRefreshTick() (tea.Model, tea.Cmd) { ... }
func (a *Application) handleRewind(msg RewindMsg) (tea.Model, tea.Cmd) { ... }
func (a *Application) handleRestoreTimeTravel(msg RestoreTimeTravelMsg) (tea.Model, tea.Cmd) { ... }
```

**Step 2: Create `internal/app/app_view.go`**
Move the View() method and rendering helpers:
```go
package app

// View renders the current view based on application mode.
// This is the core Bubble Tea View method.
func (a *Application) View() string {
    // Move entire View() method here (lines 729-905 from app.go)
}

// Rendering helpers that are only used by View()
func (a *Application) RenderStateHeader() string { ... }
func (a *Application) isInputMode() bool { ... }
```

**Step 3: Create `internal/app/app_init.go`**
Move the Init() method and initialization helpers:
```go
package app

import tea "github.com/charmbracelet/bubbletea"

// Init initializes the application state and returns initial commands.
// This is the core Bubble Tea Init method.
func (a *Application) Init() tea.Cmd {
    // Move entire Init() method here (lines 907-938 from app.go)
}

// Init helpers
func (a *Application) RestoreFromTimeTravel() tea.Cmd { ... }
```

**Step 4: Update `app.go`**
Remove the moved methods, keep:
- Application struct definition
- Constructor `NewApplication()`
- SSOT helpers (`reloadGitState`, `checkForConflicts`, `executeGitOp`)
- Mode transition helpers
- Delegation methods (can be removed in Phase 6)

**Expected Result:**
- `app.go`: 1,771 → ~900 lines
- New files: ~600 lines total
- Build must pass

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 2: Extract State Delegation Methods

**Goal:** Remove 60+ delegation methods from app.go to reduce bloat
**Files to Update:** 7 existing state files
**Lines to Remove:** ~300 lines from `app.go`
**Estimated Time:** 60 minutes

### Implementation:

**Step 1: Inline delegation methods into call sites**

Instead of:
```go
// In app.go
func (a *Application) resetCloneWorkflow() {
    a.workflowState.ResetClone()
}

// In caller
a.resetCloneWorkflow()
```

Change to direct access:
```go
// In caller
a.workflowState.ResetClone()
```

**Step 2: Update all call sites for these delegation method groups:**

| State | Methods to Inline |
|-------|-------------------|
| WorkflowState | 6 methods (resetCloneWorkflow, saveCurrentMode, etc.) |
| EnvironmentState | 11 methods (isEnvironmentReady, setSetupWizardStep, etc.) |
| PickerState | 10 methods (getHistoryState, setFileHistoryState, etc.) |
| ConsoleState | 10 methods (getConsoleBuffer, scrollConsoleUp, etc.) |
| ActivityState | 8 methods (markActivity, isMenuInactive, etc.) |
| DialogState | 8 methods (getDialog, showConfirmationDialog, etc.) |
| TimeTravelState | 7 methods (isTimeTravelActive, setTimeTravelInfo, etc.) |

**Step 3: Use find-and-replace or sed to update call sites**

Example pattern:
```bash
# Replace method calls with direct field access
sed -i 's/a\.resetCloneWorkflow()/a.workflowState.ResetClone()/g' internal/app/*.go
sed -i 's/a\.isEnvironmentReady()/a.environmentState.IsReady()/g' internal/app/*.go
# ... etc for all 60 methods
```

**Expected Result:**
- `app.go`: 900 → ~600 lines (300 delegation methods removed)
- All call sites use direct `a.xxxState.Method()` access
- Build must pass

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 3: Split Confirmation Handlers

**Goal:** Break up `confirmation_handlers.go` (972 lines, 48 methods)
**Files to Create:** 4 new files
**Estimated Time:** 90 minutes

### Implementation:

**Step 1: Create `internal/app/confirm_dialog.go`**
Core dialog infrastructure:
```go
package app

// Dialog display methods
func (a *Application) showConfirmation(config ui.ConfirmationConfig) { ... }
func (a *Application) showConfirmationFromMessage(confirmType ConfirmationType, customExplanation string) { ... }
func (a *Application) showAlert(title, explanation string) { ... }
func (a *Application) handleConfirmationResponse(confirmed bool) (tea.Model, tea.Cmd) { ... }

// Warning dialogs
func (a *Application) showNestedRepoWarning(path string) { ... }
func (a *Application) showForcePushWarning(branchName string) { ... }
func (a *Application) showHardResetWarning() { ... }
func (a *Application) showRewindConfirmation(commitHash string) tea.Cmd { ... }
```

**Step 2: Create `internal/app/confirm_time_travel.go`**
Time travel specific confirmations:
```go
package app

func (a *Application) executeConfirmTimeTravel() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectTimeTravel() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmTimeTravelReturn() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectTimeTravelReturn() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmTimeTravelMerge() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectTimeTravelMerge() (tea.Model, tea.Cmd) { ... }
// ... 12 more time travel confirmation handlers
```

**Step 3: Create `internal/app/confirm_operations.go`**
General operation confirmations:
```go
package app

func (a *Application) executeConfirmNestedRepoInit() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectNestedRepoInit() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmForcePush() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectForcePush() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmHardReset() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectHardReset() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmDirtyPull() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectDirtyPull() (tea.Model, tea.Cmd) { ... }
// ... etc
```

**Step 4: Create `internal/app/confirm_branch.go`**
Branch switch confirmations:
```go
package app

func (a *Application) executeConfirmBranchSwitchClean() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectBranchSwitch() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeConfirmBranchSwitchDirty() (tea.Model, tea.Cmd) { ... }
func (a *Application) executeRejectBranchSwitchDirty() (tea.Model, tea.Cmd) { ... }
```

**Step 5: Update `confirmation_handlers.go`**
Remove moved methods, keep only:
- Shared helper methods (if any)
- OR delete file entirely if all methods moved

**Expected Result:**
- `confirmation_handlers.go`: 972 → ~200 lines (or deleted)
- 4 new files: ~800 lines total
- Better organization by domain
- Build must pass

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 4: Create Sub-Packages

**Goal:** Move state structs to `internal/app/state/` sub-package
**Files to Move:** 10 state files
**Estimated Time:** 120 minutes (complex due to import changes)

### Implementation:

**Step 1: Create `internal/app/state/` directory**

**Step 2: Move state files to sub-package**
```bash
mkdir -p internal/app/state
mv internal/app/*_state.go internal/app/state/
```

Files to move:
- `activity_state.go`
- `async_state.go`
- `console_state.go`
- `conflict_state.go`
- `dirty_state.go`
- `dialog_state.go`
- `environment_state.go`
- `input_state.go`
- `picker_state.go`
- `time_travel_state.go`
- `workflow_state.go`

**Step 3: Update package declaration in moved files**
```go
// Change from:
package app

// To:
package state
```

**Step 4: Update imports in `internal/app/` files**
Add import and update type references:
```go
import (
    "tit/internal/app/state"
    // ... other imports
)

// Change field types:
activityState ActivityState  →  activityState state.ActivityState
inputState InputState        →  inputState state.InputState
// etc for all 10 state structs
```

**Step 5: Update constructor calls**
```go
// Change from:
activityState: NewActivityState(),

// To:
activityState: state.NewActivityState(),
```

**Step 6: Update all call sites to use state package**
```go
// Change from:
a.activityState.MarkActivity()

// To:
a.activityState.MarkActivity()  // Still works - field is now state.ActivityState
```

**Expected Result:**
- New `internal/app/state/` package with 10 files
- `internal/app/` file count: 57 → 46 files
- Better logical organization
- Build must pass

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 5: Extract Menu System

**Goal:** Move menu-related files to `internal/app/menu/` sub-package
**Files to Move:** 4 files
**Estimated Time:** 90 minutes

### Implementation:

**Step 1: Create `internal/app/menu/` directory**

**Step 2: Move menu files**
```bash
mkdir -p internal/app/menu
mv internal/app/menu.go internal/app/menu/
mv internal/app/menu_items.go internal/app/menu/
mv internal/app/menu_builders.go internal/app/menu/
mv internal/app/state_info.go internal/app/menu/  # Contains menu state info
```

**Step 3: Update package declarations**
```go
package menu
```

**Step 4: Update imports and references**
Files that use menu types will need:
```go
import "tit/internal/app/menu"

// Change types:
menuItems []MenuItem  →  menuItems []menu.MenuItem
```

**Step 5: Handle circular dependencies**
If menu package needs Application, use interface:
```go
// In menu package
type MenuGenerator interface {
    GetGitState() *git.State
    // ... minimal interface
}

func GenerateMenu(app MenuGenerator) []MenuItem { ... }
```

**Expected Result:**
- New `internal/app/menu/` package with 4 files
- `internal/app/` file count: 46 → 42 files
- Menu system is now isolated and testable
- Build must pass

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 6: Final Cleanup and Documentation

**Goal:** Remove dead code, update documentation, verify metrics
**Estimated Time:** 60 minutes

### Implementation:

**Step 1: Remove any remaining unused delegation methods**
Search for methods that are no longer called:
```bash
grep -n "func (a \*Application)" internal/app/app.go | head -40
```

**Step 2: Delete empty or near-empty files**
Check if any files have <50 lines:
```bash
wc -l internal/app/*.go | sort -n | head -10
```

**Step 3: Update ARCHITECTURE.md**
Document the new package structure:
```markdown
## Package Structure

### internal/app/
Core application logic and Bubble Tea interface
- `app.go` - Application struct and core methods
- `app_update.go` - Update() method and message handlers
- `app_view.go` - View() method and rendering
- `app_init.go` - Init() method and initialization

### internal/app/state/
State management structs
- `activity_state.go` - Menu activity tracking
- `async_state.go` - Async operation state
- `console_state.go` - Console output state
- ... etc

### internal/app/menu/
Menu generation and display
- `menu.go` - Core menu logic
- `menu_items.go` - Menu item definitions
- `menu_builders.go` - Menu builders by git state
- `state_info.go` - State display info
```

**Step 4: Update CODEBASE-MAP.md**
Reflect new file organization

**Step 5: Verify final metrics**
```bash
# Count lines in each package
echo "=== internal/app ==="
find internal/app -maxdepth 1 -name "*.go" -exec wc -l {} + | tail -1
echo "=== internal/app/state ==="
find internal/app/state -name "*.go" -exec wc -l {} + | tail -1
echo "=== internal/app/menu ==="
find internal/app/menu -name "*.go" -exec wc -l {} + | tail -1
```

**Expected Result:**
- Clean, documented package structure
- All files have clear purpose
- Total line count documented
- Build passes

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
go test ./...
# Must compile and pass all tests
```

---

## Summary of Changes

| Phase | Action | Files Changed | Est. Lines Impact |
|-------|--------|---------------|-------------------|
| 1 | Extract Update/View/Init | +3 files | app.go: -600 |
| 2 | Inline delegation methods | ~20 files | app.go: -300 |
| 3 | Split confirmation handlers | +4 files | confirmation_handlers.go: -800 |
| 4 | Create state/ sub-package | 10 files moved | organizational |
| 5 | Create menu/ sub-package | 4 files moved | organizational |
| 6 | Cleanup and docs | all | documentation |

**Expected Final Metrics:**
- `internal/app/app.go`: 1,771 → ~600 lines (66% reduction)
- `internal/app/`: 57 files → ~42 files (26% reduction)
- New `internal/app/state/`: 10 files, ~800 lines
- New `internal/app/menu/`: 4 files, ~400 lines
- **Total lines remain ~11,000** but better organized

---

## Critical Rules

1. **Clean Build After EVERY Phase**
   - Run `go build ./...` before proceeding
   - No warnings, no errors

2. **No Logic Changes**
   - Only move code, don't modify behavior
   - Keep all existing functionality

3. **Import Path Updates**
   - Use `tit/internal/app/state` not relative paths
   - Update all imports consistently

4. **Handle Circular Dependencies**
   - Use interfaces to break cycles
   - Keep dependencies unidirectional

5. **Test After Each Phase**
   - Run `./build.sh` or `go build ./...`
   - Verify application still runs

---

## Rollback Plan

If any phase fails:
1. Stop immediately
2. Do not proceed to next phase
3. Report issue to user
4. User decides: fix forward or git revert

---

**End of Sprint 3 Kickoff Plan**

Ready for ENGINEER to begin Phase 1.
