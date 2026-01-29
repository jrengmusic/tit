# Sprint 2: God Object Refactoring - Incremental Kickoff Plan

**Date:** 2026-01-29
**Objective:** Refactor Application struct from 47 fields to ~21 fields through incremental extraction
**Approach:** 9 phases, each with clean build verification
**Estimated Duration:** 4-6 hours total (30-45 min per phase)

---

## Pre-Flight: Read Required Documentation

Before starting, ENGINEER must read:
1. `SPEC.md` - State model and operation specifications
2. `ARCHITECTURE.md` - Current Application struct architecture
3. `internal/app/app.go` - Current Application struct (lines 32-124)
4. `internal/app/input_state.go` - Existing extraction pattern
5. `internal/app/cache_manager.go` - Existing extraction pattern
6. `internal/app/async_state.go` - Existing extraction pattern

**Pattern to Follow:**
- Named composition (NOT embedding)
- Constructor function: `NewXxxState()`
- Methods on struct, not on Application
- Delegation methods in app.go for backward compatibility

---

## Phase 0: Quick Wins (AUD-001, AUD-002, AUD-003)

**Goal:** Fix immediate issues before structural refactoring
**Files:** 3 files
**Estimated Time:** 15 minutes

### Tasks:

1. **Fix AUD-001: Hardcoded ".git" string**
   - File: `internal/git/state.go:443`
   - Change: `gitDir := ".git"` → `gitDir := internal.GitDirectoryName`
   - Verify: Check other usages use constant (lines 334, 401, 457)

2. **Delete AUD-002: Leftover backup file**
   - File: `internal/git/execute.go.backup`
   - Action: Delete file

3. **Delete AUD-003: Leftover temporary file**
   - File: `internal/app/part1.txt`
   - Action: Delete file

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 1: Extract WorkflowState

**Goal:** Extract clone/init workflow state and mode restoration
**Fields:** 7 (`cloneURL`, `clonePath`, `cloneMode`, `cloneBranches`, `previousMode`, `previousMenuIndex`, `pendingRewindCommit`)
**New File:** `internal/app/workflow_state.go`
**Estimated Time:** 45 minutes

### Implementation:

**Step 1: Create `internal/app/workflow_state.go`**
```go
package app

// WorkflowState manages transient state for multi-step workflows (clone, init).
// All fields reset when workflow completes or is cancelled.
type WorkflowState struct {
    // Clone workflow
    CloneURL      string
    ClonePath     string
    CloneMode     string   // "here" or "subdir"
    CloneBranches []string // Available branches after clone
    
    // Mode restoration (for ESC handling)
    PreviousMode      AppMode
    PreviousMenuIndex int
    
    // Pending operations
    PendingRewindCommit string
}

// NewWorkflowState creates a new WorkflowState with defaults.
func NewWorkflowState() WorkflowState {
    return WorkflowState{
        CloneMode:         "here",
        PreviousMode:      ModeMenu,
        PreviousMenuIndex: 0,
    }
}

// ResetClone clears all clone-related state.
func (w *WorkflowState) ResetClone() {
    w.CloneURL = ""
    w.ClonePath = ""
    w.CloneMode = "here"
    w.CloneBranches = nil
}

// SaveMode stores current mode and index for ESC restoration.
func (w *WorkflowState) SaveMode(mode AppMode, index int) {
    w.PreviousMode = mode
    w.PreviousMenuIndex = index
}

// RestoreMode returns the saved mode and index.
func (w *WorkflowState) RestoreMode() (AppMode, int) {
    return w.PreviousMode, w.PreviousMenuIndex
}

// SetPendingRewind stores a commit hash for rewind operation.
func (w *WorkflowState) SetPendingRewind(commit string) {
    w.PendingRewindCommit = commit
}

// GetPendingRewind returns the pending rewind commit (empty if none).
func (w *WorkflowState) GetPendingRewind() string {
    return w.PendingRewindCommit
}

// ClearPendingRewind removes the pending rewind commit.
func (w *WorkflowState) ClearPendingRewind() {
    w.PendingRewindCommit = ""
}
```

**Step 2: Update Application struct in `app.go`**
- Add field: `workflowState WorkflowState`
- Remove fields: `cloneURL`, `clonePath`, `cloneMode`, `cloneBranches`, `previousMode`, `previousMenuIndex`, `pendingRewindCommit`

**Step 3: Initialize in `NewApplication()`**
```go
workflowState: NewWorkflowState(),
```

**Step 4: Add delegation methods in `app.go`**
```go
// Workflow state delegation
func (a *Application) resetCloneWorkflow() {
    a.workflowState.ResetClone()
}

func (a *Application) saveCurrentMode() {
    a.workflowState.SaveMode(a.mode, a.selectedIndex)
}

func (a *Application) restorePreviousMode() (AppMode, int) {
    return a.workflowState.RestoreMode()
}

func (a *Application) setPendingRewind(commit string) {
    a.workflowState.SetPendingRewind(commit)
}

func (a *Application) getPendingRewind() string {
    return a.workflowState.GetPendingRewind()
}

func (a *Application) clearPendingRewind() {
    a.workflowState.ClearPendingRewind()
}
```

**Step 5: Update all call sites**
Search for usages of the 7 extracted fields and update to use delegation methods or `a.workflowState.Xxx`.

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 2: Extract EnvironmentState

**Goal:** Extract git environment detection and setup wizard state
**Fields:** 5 (`gitEnvironment`, `setupWizardStep`, `setupWizardError`, `setupEmail`, `setupKeyCopied`)
**New File:** `internal/app/environment_state.go`
**Estimated Time:** 45 minutes

### Implementation:

**Step 1: Create `internal/app/environment_state.go`**
```go
package app

import "github.com/jrengmusic/tit/internal/git"

// SetupWizardStep represents the current step in the setup wizard.
type SetupWizardStep int

const (
    SetupStepWelcome SetupWizardStep = iota
    SetupStepPrerequisites
    SetupStepEmail
    SetupStepGenerate
    SetupStepDisplayKey
    SetupStepComplete
    SetupStepFatalMissingGit
    SetupStepFatalMissingSSH
)

// EnvironmentState manages git environment detection and setup wizard state.
// This is only relevant before the main application loop starts.
type EnvironmentState struct {
    GitEnvironment   git.GitEnvironment  // Ready, NeedsSetup, MissingGit, MissingSSH
    SetupWizardStep  SetupWizardStep     // Current step in wizard
    SetupWizardError string              // Error message for SetupStepError
    SetupEmail       string              // Email for SSH key generation
    SetupKeyCopied   bool                // Public key copied to clipboard
}

// NewEnvironmentState creates a new EnvironmentState with defaults.
func NewEnvironmentState() EnvironmentState {
    return EnvironmentState{
        GitEnvironment:  git.GitEnvironmentReady,
        SetupWizardStep: SetupStepWelcome,
    }
}

// IsReady returns true if git environment is ready for operation.
func (e *EnvironmentState) IsReady() bool {
    return e.GitEnvironment == git.GitEnvironmentReady
}

// NeedsSetup returns true if setup wizard is required.
func (e *EnvironmentState) NeedsSetup() bool {
    return e.GitEnvironment == git.GitEnvironmentNeedsSetup
}

// SetEnvironment updates the git environment state.
func (e *EnvironmentState) SetEnvironment(env git.GitEnvironment) {
    e.GitEnvironment = env
}

// SetWizardStep updates the current setup wizard step.
func (e *EnvironmentState) SetWizardStep(step SetupWizardStep) {
    e.SetupWizardStep = step
}

// SetWizardError sets an error message for the wizard.
func (e *EnvironmentState) SetWizardError(err string) {
    e.SetupWizardError = err
}

// GetEmail returns the setup email.
func (e *EnvironmentState) GetEmail() string {
    return e.SetupEmail
}

// SetEmail sets the setup email.
func (e *EnvironmentState) SetEmail(email string) {
    e.SetupEmail = email
}

// MarkKeyCopied marks the SSH key as copied.
func (e *EnvironmentState) MarkKeyCopied() {
    e.SetupKeyCopied = true
}

// IsKeyCopied returns true if the SSH key has been copied.
func (e *EnvironmentState) IsKeyCopied() bool {
    return e.SetupKeyCopied
}
```

**Step 2: Update Application struct in `app.go`**
- Add field: `environmentState EnvironmentState`
- Remove fields: `gitEnvironment`, `setupWizardStep`, `setupWizardError`, `setupEmail`, `setupKeyCopied`
- Note: May need to keep `gitEnvironment` as exported getter if used externally

**Step 3: Initialize in `NewApplication()`**
```go
environmentState: NewEnvironmentState(),
```

**Step 4: Add delegation methods and update call sites**

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 3: Extract PickerState

**Goal:** Extract history, file history, and branch picker UI states
**Fields:** 3 (`historyState`, `fileHistoryState`, `branchPickerState`)
**New File:** `internal/app/picker_state.go`
**Estimated Time:** 30 minutes

### Implementation:

**Step 1: Create `internal/app/picker_state.go`**
```go
package app

import "github.com/jrengmusic/tit/internal/ui"

// PickerState manages all picker mode states (history, file history, branch picker).
// These share a common pattern: list pane + details pane with coordinated scrolling.
type PickerState struct {
    History      *ui.HistoryState
    FileHistory  *ui.FileHistoryState
    BranchPicker *ui.BranchPickerState
}

// NewPickerState creates a new PickerState with nil states.
func NewPickerState() PickerState {
    return PickerState{}
}

// GetHistory returns the history state (may be nil).
func (p *PickerState) GetHistory() *ui.HistoryState {
    return p.History
}

// SetHistory sets the history state.
func (p *PickerState) SetHistory(state *ui.HistoryState) {
    p.History = state
}

// ResetHistory clears the history state.
func (p *PickerState) ResetHistory() {
    p.History = nil
}

// GetFileHistory returns the file history state (may be nil).
func (p *PickerState) GetFileHistory() *ui.FileHistoryState {
    return p.FileHistory
}

// SetFileHistory sets the file history state.
func (p *PickerState) SetFileHistory(state *ui.FileHistoryState) {
    p.FileHistory = state
}

// ResetFileHistory clears the file history state.
func (p *PickerState) ResetFileHistory() {
    p.FileHistory = nil
}

// GetBranchPicker returns the branch picker state (may be nil).
func (p *PickerState) GetBranchPicker() *ui.BranchPickerState {
    return p.BranchPicker
}

// SetBranchPicker sets the branch picker state.
func (p *PickerState) SetBranchPicker(state *ui.BranchPickerState) {
    p.BranchPicker = state
}

// ResetBranchPicker clears the branch picker state.
func (p *PickerState) ResetBranchPicker() {
    p.BranchPicker = nil
}

// ResetAll clears all picker states.
func (p *PickerState) ResetAll() {
    p.History = nil
    p.FileHistory = nil
    p.BranchPicker = nil
}
```

**Step 2-4:** Follow same pattern as Phase 1

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 4: Extract ConsoleState

**Goal:** Extract console output management state
**Fields:** 3 (`consoleState`, `outputBuffer`, `consoleAutoScroll`)
**New File:** `internal/app/console_state.go`
**Estimated Time:** 30 minutes

### Implementation:

**Step 1: Create `internal/app/console_state.go`**
```go
package app

import "github.com/jrengmusic/tit/internal/ui"

// ConsoleState manages console output display and scrolling.
// Thread-safe: Uses thread-safe OutputBuffer.
type ConsoleState struct {
    state      ui.ConsoleOutState  // Scroll position, etc.
    buffer     *ui.OutputBuffer    // Thread-safe output buffer
    autoScroll bool                // Auto-scroll to bottom
}

// NewConsoleState creates a new ConsoleState.
func NewConsoleState() ConsoleState {
    return ConsoleState{
        buffer:     ui.NewOutputBuffer(),
        autoScroll: true,
    }
}

// GetBuffer returns the output buffer.
func (c *ConsoleState) GetBuffer() *ui.OutputBuffer {
    return c.buffer
}

// Clear clears the console buffer.
func (c *ConsoleState) Clear() {
    c.buffer.Clear()
}

// ScrollUp scrolls the console view up.
func (c *ConsoleState) ScrollUp() {
    if c.state.ScrollOffset > 0 {
        c.state.ScrollOffset--
    }
}

// ScrollDown scrolls the console view down.
func (c *ConsoleState) ScrollDown() {
    c.state.ScrollOffset++
}

// PageUp scrolls up by page.
func (c *ConsoleState) PageUp() {
    if c.state.ScrollOffset > 10 {
        c.state.ScrollOffset -= 10
    } else {
        c.state.ScrollOffset = 0
    }
}

// PageDown scrolls down by page.
func (c *ConsoleState) PageDown() {
    c.state.ScrollOffset += 10
}

// ToggleAutoScroll toggles auto-scroll behavior.
func (c *ConsoleState) ToggleAutoScroll() {
    c.autoScroll = !c.autoScroll
}

// IsAutoScroll returns true if auto-scroll is enabled.
func (c *ConsoleState) IsAutoScroll() bool {
    return c.autoScroll
}

// GetState returns the console state.
func (c *ConsoleState) GetState() ui.ConsoleOutState {
    return c.state
}

// SetScrollOffset sets the scroll offset directly.
func (c *ConsoleState) SetScrollOffset(offset int) {
    c.state.ScrollOffset = offset
}
```

**Step 2-4:** Follow same pattern as Phase 1

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 5: Extract ActivityState

**Goal:** Extract menu activity tracking and auto-update state
**Fields:** 4 (`lastMenuActivity`, `menuActivityTimeout`, `autoUpdateInProgress`, `autoUpdateFrame`)
**New File:** `internal/app/activity_state.go`
**Estimated Time:** 30 minutes

### Implementation:

**Step 1: Create `internal/app/activity_state.go`**
```go
package app

import "time"

// ActivityState tracks menu navigation activity and auto-update status.
type ActivityState struct {
    lastActivity         time.Time
    activityTimeout      time.Duration
    autoUpdateInProgress bool
    autoUpdateFrame      int
}

// NewActivityState creates a new ActivityState with defaults.
func NewActivityState() ActivityState {
    return ActivityState{
        lastActivity:    time.Now(),
        activityTimeout: 30 * time.Second,
    }
}

// MarkActivity updates the last activity timestamp to now.
func (a *ActivityState) MarkActivity() {
    a.lastActivity = time.Now()
}

// IsInactive returns true if no activity for longer than timeout.
func (a *ActivityState) IsInactive() bool {
    return time.Since(a.lastActivity) > a.activityTimeout
}

// SetActivityTimeout sets the inactivity timeout duration.
func (a *ActivityState) SetActivityTimeout(timeout time.Duration) {
    a.activityTimeout = timeout
}

// StartAutoUpdate marks auto-update as in progress.
func (a *ActivityState) StartAutoUpdate() {
    a.autoUpdateInProgress = true
    a.autoUpdateFrame = 0
}

// StopAutoUpdate marks auto-update as complete.
func (a *ActivityState) StopAutoUpdate() {
    a.autoUpdateInProgress = false
}

// IsAutoUpdateInProgress returns true if auto-update is running.
func (a *ActivityState) IsAutoUpdateInProgress() bool {
    return a.autoUpdateInProgress
}

// IncrementFrame advances the auto-update animation frame.
func (a *ActivityState) IncrementFrame() {
    a.autoUpdateFrame++
}

// GetFrame returns the current auto-update frame.
func (a *ActivityState) GetFrame() int {
    return a.autoUpdateFrame
}
```

**Step 2-4:** Follow same pattern as Phase 1

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 6: Extract DialogState

**Goal:** Extract confirmation dialog state
**Fields:** 2 (`confirmationDialog`, `confirmContext`)
**New File:** `internal/app/dialog_state.go`
**Estimated Time:** 20 minutes

### Implementation:

**Step 1: Create `internal/app/dialog_state.go`**
```go
package app

import "github.com/jrengmusic/tit/internal/ui"

// DialogState manages confirmation dialog display and context.
type DialogState struct {
    dialog  *ui.ConfirmationDialog
    context map[string]string  // Context data for dialog actions
}

// NewDialogState creates a new DialogState.
func NewDialogState() DialogState {
    return DialogState{
        context: make(map[string]string),
    }
}

// Show sets the dialog to display with context.
func (d *DialogState) Show(dialog *ui.ConfirmationDialog, ctx map[string]string) {
    d.dialog = dialog
    if ctx != nil {
        d.context = ctx
    } else {
        d.context = make(map[string]string)
    }
}

// Hide clears the current dialog.
func (d *DialogState) Hide() {
    d.dialog = nil
    d.context = make(map[string]string)
}

// GetDialog returns the current dialog (may be nil).
func (d *DialogState) GetDialog() *ui.ConfirmationDialog {
    return d.dialog
}

// IsVisible returns true if a dialog is currently shown.
func (d *DialogState) IsVisible() bool {
    return d.dialog != nil
}

// GetContext returns the dialog context map.
func (d *DialogState) GetContext() map[string]string {
    return d.context
}

// SetContextValue sets a single context value.
func (d *DialogState) SetContextValue(key, value string) {
    d.context[key] = value
}

// GetContextValue returns a single context value (empty if not found).
func (d *DialogState) GetContextValue(key string) string {
    return d.context[key]
}
```

**Step 2-4:** Follow same pattern as Phase 1

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 7: Extract TimeTravelState

**Goal:** Extract time travel operation state
**Fields:** 2 (`timeTravelInfo`, `restoreTimeTravelInitiated`)
**New File:** `internal/app/time_travel_state.go`
**Estimated Time:** 20 minutes

### Implementation:

**Step 1: Create `internal/app/time_travel_state.go`**
```go
package app

import "github.com/jrengmusic/tit/internal/git"

// TimeTravelState manages time travel operation state.
type TimeTravelState struct {
    info             *git.TimeTravelInfo
    restoreInitiated bool
}

// NewTimeTravelState creates a new TimeTravelState.
func NewTimeTravelState() TimeTravelState {
    return TimeTravelState{}
}

// IsActive returns true if currently in time travel mode.
func (t *TimeTravelState) IsActive() bool {
    return t.info != nil
}

// GetInfo returns the time travel info (may be nil).
func (t *TimeTravelState) GetInfo() *git.TimeTravelInfo {
    return t.info
}

// SetInfo sets the time travel info.
func (t *TimeTravelState) SetInfo(info *git.TimeTravelInfo) {
    t.info = info
}

// Clear removes time travel state.
func (t *TimeTravelState) Clear() {
    t.info = nil
    t.restoreInitiated = false
}

// IsRestoreInitiated returns true if restore operation has started.
func (t *TimeTravelState) IsRestoreInitiated() bool {
    return t.restoreInitiated
}

// MarkRestoreInitiated marks the restore as initiated.
func (t *TimeTravelState) MarkRestoreInitiated() {
    t.restoreInitiated = true
}

// ClearRestore resets the restore initiated flag.
func (t *TimeTravelState) ClearRestore() {
    t.restoreInitiated = false
}
```

**Step 2-4:** Follow same pattern as Phase 1

### Verification:
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

---

## Phase 8: Final Cleanup

**Goal:** Remove delegation methods, update documentation
**Estimated Time:** 30 minutes

### Tasks:

1. **Remove temporary delegation methods** from `app.go`
   - Only if all call sites have been updated to use struct access
   - Keep delegation for commonly used methods if they improve readability

2. **Update ARCHITECTURE.md**
   - Document the new state struct architecture
   - Update Application struct field count (47 → ~21)
   - Add section explaining the composition pattern

3. **Update CODEBASE-MAP.md** (if exists)
   - Add new state files to file listing

4. **Final verification**
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
go test ./...
# Must compile and pass all tests
```

---

## Summary of Changes

| Phase | File Created | Fields Extracted | Application Fields After |
|-------|--------------|------------------|--------------------------|
| 0 | - | 0 | 47 |
| 1 | `workflow_state.go` | 7 | 40 |
| 2 | `environment_state.go` | 5 | 35 |
| 3 | `picker_state.go` | 3 | 32 |
| 4 | `console_state.go` | 3 | 29 |
| 5 | `activity_state.go` | 4 | 25 |
| 6 | `dialog_state.go` | 2 | 23 |
| 7 | `time_travel_state.go` | 2 | 21 |
| 8 | - | Cleanup | 21 |

**Total:** 7 new files, 26 fields extracted, 55% reduction in Application struct size

---

## Critical Rules

1. **Clean Build After EVERY Phase**
   - Run `go build ./...` before proceeding to next phase
   - No warnings, no errors

2. **Named Composition Only**
   - Use `workflowState WorkflowState` NOT `WorkflowState` (embedding)
   - Be explicit about field access

3. **Delegation Methods for Backward Compatibility**
   - Add delegation methods during transition
   - Remove only after all call sites updated

4. **Follow Existing Patterns**
   - Copy structure from `input_state.go`, `cache_manager.go`, `async_state.go`
   - Use same naming conventions
   - Same method receiver patterns

5. **No Logic Changes**
   - Only move fields and add accessor methods
   - Do NOT change behavior
   - Do NOT optimize

---

## Rollback Plan

If any phase fails:
1. Stop immediately
2. Do not proceed to next phase
3. Report issue to user
4. User decides: fix forward or git revert

---

**End of Kickoff Plan**

Ready for ENGINEER to begin Phase 0.
