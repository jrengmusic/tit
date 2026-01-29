# Sprint 3B: Reduce app.go to 400 Lines - Kickoff Plan

**Date:** 2026-01-30
**Objective:** Reduce `internal/app/app.go` from 975 to ~400 lines through aggressive code extraction
**Approach:** 3-phase deletion and extraction with clean build verification
**Estimated Duration:** 2-3 hours total (45-60 min per phase)

---

## Current State Analysis

**app.go: 975 lines breakdown:**

| Section | Lines | Content |
|---------|-------|---------|
| Struct + imports | 1-96 | Application struct definition |
| Core helpers | 98-195 | transitionTo, reloadGitState, checkForConflicts, executeGitOp |
| Async helpers | 197-240 | startAsyncOp, endAsyncOp, abortAsyncOp, etc. |
| Constructor | 242-419 | newSetupWizardApp, NewApplication (177 lines) |
| Footer getter | 424-426 | GetFooterHint |
| Key handlers | 428-568 | buildKeyHandlers (140 lines) |
| Menu shortcuts | 570-650 | rebuildMenuShortcuts (80 lines) |
| Menu handlers | 652-711 | handleMenuUp, handleMenuDown, handleMenuEnter |
| Input helpers | 713-765 | insertTextAtCursor, deleteAtCursor, updateInputValidation |
| Input handlers | 767-802 | handleInputSubmit, handleHistoryRewind |
| **DEAD CODE** | **804-975** | **41 delegation methods (171 lines)** |

**Key Finding:** The 41 delegation methods are **NEVER CALLED**. The codebase uses direct state access.

---

## Phase 1: Delete Dead Delegation Methods

**Goal:** Remove 171 lines of unused code
**Risk:** ZERO - verified no call sites exist
**Estimated Time:** 15 minutes

### Implementation:

**Step 1: Delete lines 804-975 from app.go**

Remove these 41 methods entirely:
```go
// Workflow state delegation (6 methods)
func (a *Application) resetCloneWorkflow() { ... }
func (a *Application) saveCurrentMode() { ... }
func (a *Application) restorePreviousMode() (AppMode, int) { ... }
func (a *Application) setPendingRewind(commit string) { ... }
func (a *Application) getPendingRewind() string { ... }
func (a *Application) clearPendingRewind() { ... }

// Environment state delegation (11 methods)
func (a *Application) isEnvironmentReady() bool { ... }
func (a *Application) needsEnvironmentSetup() bool { ... }
func (a *Application) setEnvironment(env git.GitEnvironment) { ... }
func (a *Application) getSetupWizardStep() SetupWizardStep { ... }
func (a *Application) setSetupWizardStep(step SetupWizardStep) { ... }
func (a *Application) getSetupWizardError() string { ... }
func (a *Application) setSetupWizardError(err string) { ... }
func (a *Application) getSetupEmail() string { ... }
func (a *Application) setSetupEmail(email string) { ... }
func (a *Application) markSetupKeyCopied() { ... }
func (a *Application) isSetupKeyCopied() bool { ... }

// Picker state delegation (10 methods)
func (a *Application) getHistoryState() *ui.HistoryState { ... }
func (a *Application) setHistoryState(state *ui.HistoryState) { ... }
func (a *Application) resetHistoryState() { ... }
func (a *Application) getFileHistoryState() *ui.FileHistoryState { ... }
func (a *Application) setFileHistoryState(state *ui.FileHistoryState) { ... }
func (a *Application) resetFileHistoryState() { ... }
func (a *Application) getBranchPickerState() *ui.BranchPickerState { ... }
func (a *Application) setBranchPickerState(state *ui.BranchPickerState) { ... }
func (a *Application) resetBranchPickerState() { ... }
func (a *Application) resetAllPickerStates() { ... }

// Console state delegation (10 methods)
func (a *Application) getConsoleBuffer() *ui.OutputBuffer { ... }
func (a *Application) clearConsoleBuffer() { ... }
func (a *Application) scrollConsoleUp() { ... }
func (a *Application) scrollConsoleDown() { ... }
func (a *Application) pageConsoleUp() { ... }
func (a *Application) pageConsoleDown() { ... }
func (a *Application) toggleConsoleAutoScroll() { ... }
func (a *Application) isConsoleAutoScroll() bool { ... }
func (a *Application) getConsoleState() ui.ConsoleOutState { ... }
func (a *Application) setConsoleScrollOffset(offset int) { ... }

// Activity state delegation (4 methods)
func (a *Application) markMenuActivity() { ... }
func (a *Application) isMenuInactive() bool { ... }
func (a *Application) setMenuActivityTimeout(timeout time.Duration) { ... }
func (a *Application) getActivityTimeout() time.Duration { ... }
```

**Verification:**
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

**Expected Result:** app.go: 975 → 804 lines

---

## Phase 2: Extract Constructor

**Goal:** Move constructor logic to separate file
**Lines to Move:** ~177 lines
**New File:** `internal/app/app_constructor.go`
**Estimated Time:** 45 minutes

### Implementation:

**Step 1: Create `internal/app/app_constructor.go`**

```go
package app

import (
	"fmt"
	"os"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"
)

// newSetupWizardApp creates a minimal Application for the setup wizard
func newSetupWizardApp(sizing ui.DynamicSizing, theme ui.Theme, gitEnv git.GitEnvironment) *Application {
	envState := NewEnvironmentState()
	envState.SetEnvironment(gitEnv)
	app := &Application{
		sizing:           sizing,
		theme:            theme,
		mode:             ModeSetupWizard,
		environmentState: envState,
		asyncState:       AsyncState{exitAllowed: true},
		consoleState:     NewConsoleState(),
		dialogState:      NewDialogState(),
	}
	app.keyHandlers = app.buildKeyHandlers()
	return app
}

// NewApplication creates a new application instance
func NewApplication(sizing ui.DynamicSizing, theme ui.Theme, cfg *config.Config) *Application {
	// PRIORITY 0: Check git environment BEFORE anything else
	gitEnv := git.DetectGitEnvironment()
	InitGitLogger()
	if gitEnv != git.Ready {
		return newSetupWizardApp(sizing, theme, gitEnv)
	}

	// Try to find and cd into git repository
	isRepo, repoPath := git.IsInitializedRepo()
	if !isRepo {
		isRepo, repoPath = git.HasParentRepo()
	}

	var gitState *git.State
	if isRepo && repoPath != "" {
		if err := os.Chdir(repoPath); err != nil {
			panic(fmt.Sprintf("cannot cd into repository at %s: %v", repoPath, err))
		}
		state, err := git.DetectState()
		if err != nil {
			panic(fmt.Sprintf("Failed to detect git state in repo %s: %v", repoPath, err))
		}
		gitState = state
	} else {
		gitState = &git.State{Operation: git.NotRepo}
	}

	// Build state info maps
	workingTreeInfo, timelineInfo, operationInfo := BuildStateInfo(theme)

	app := &Application{
		sizing:          sizing,
		theme:           theme,
		mode:            ModeMenu,
		gitState:        gitState,
		selectedIndex:   0,
		asyncState:      AsyncState{exitAllowed: true},
		workflowState:   NewWorkflowState(),
		consoleState:    NewConsoleState(),
		workingTreeInfo: workingTreeInfo,
		timelineInfo:    timelineInfo,
		operationInfo:   operationInfo,
		pickerState: PickerState{
			History: &ui.HistoryState{
				Commits:     make([]ui.CommitInfo, 0),
				SelectedIdx: 0,
				PaneFocused: true,
			},
			FileHistory: &ui.FileHistoryState{
				Commits:           make([]ui.CommitInfo, 0),
				Files:             make([]ui.FileInfo, 0),
				SelectedCommitIdx: 0,
				SelectedFileIdx:   0,
				FocusedPane:       ui.PaneCommits,
			},
			BranchPicker: &ui.BranchPickerState{
				Branches:    make([]ui.BranchInfo, 0),
				SelectedIdx: 0,
				PaneFocused: true,
			},
		},
		cacheManager:  NewCacheManager(),
		appConfig:     cfg,
		activityState: NewActivityState(),
		dialogState:   NewDialogState(),
	}

	// Build and cache key handler registry
	app.keyHandlers = app.buildKeyHandlers()

	// Check for incomplete time travel restoration
	hasTimeTravelMarker := git.FileExists(".git/TIT_TIME_TRAVEL") && isRepo
	shouldRestore := hasTimeTravelMarker && app.gitState.Operation != git.TimeTraveling

	if app.gitState.Operation == git.TimeTraveling && hasTimeTravelMarker {
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			shouldRestore = true
		} else {
			app.timeTravelState.SetInfo(ttInfo)
		}
	}

	if shouldRestore {
		app.mode = ModeConsole
		app.startAsyncOp()
		app.workflowState.PreviousMode = ModeMenu
		app.footerHint = "Restoring from incomplete time travel session..."
	}

	// Pre-generate menu
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 && !shouldRestore {
		app.footerHint = menu[0].Hint
	}

	app.rebuildMenuShortcuts(ModeMenu)

	if !shouldRestore {
		app.cacheManager.SetLoadingStarted(true)
	}

	if shouldRestore {
		app.startAsyncOp()
	}

	return app
}

// GetFooterHint returns the footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}
```

**Step 2: Remove from app.go**
- Delete `newSetupWizardApp()` (lines 242-258)
- Delete `NewApplication()` (lines 260-419)
- Delete `GetFooterHint()` (lines 424-426)

**Verification:**
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

**Expected Result:** app.go: 804 → 627 lines

---

## Phase 3: Extract Key Handlers

**Goal:** Move key handler registration to separate file
**Lines to Move:** ~220 lines
**New File:** `internal/app/app_keys.go`
**Estimated Time:** 60 minutes

### Implementation:

**Step 1: Create `internal/app/app_keys.go`**

```go
package app

// buildKeyHandlers creates the complete key handler registry
func (a *Application) buildKeyHandlers() map[AppMode]map[string]KeyHandler {
	// Global handlers - highest priority
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash,
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	cursorNavMixin := CursorNavigationMixin{}

	// Generic input cursor handlers
	genericInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string { return a.inputState.Value },
		func(a *Application) int { return a.inputState.CursorPosition },
		func(a *Application, pos int) { a.inputState.CursorPosition = pos },
	)

	// Mode-specific handlers
	modeHandlers := map[AppMode]map[string]KeyHandler{
		ModeMenu: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			Build(),
		ModeConsole: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeInput: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleInputSubmit).
			Build(),
		ModeInitializeLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleInitLocationSelection).
			On("1", a.handleInitLocationChoice1).
			On("2", a.handleInitLocationChoice2).
			Build(),
		ModeCloneURL: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleCloneURLSubmit).
			Build(),
		ModeCloneLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleCloneLocationSelection).
			On("1", a.handleCloneLocationChoice1).
			On("2", a.handleCloneLocationChoice2).
			Build(),
		ModeConfirmation: NewModeHandlers().
			On("left", a.handleConfirmationLeft).
			On("right", a.handleConfirmationRight).
			On("h", a.handleConfirmationLeft).
			On("l", a.handleConfirmationRight).
			On("y", a.handleConfirmationYes).
			On("n", a.handleConfirmationNo).
			On("enter", a.handleConfirmationEnter).
			Build(),
		ModeHistory: NewModeHandlers().
			On("up", a.handleHistoryUp).
			On("k", a.handleHistoryUp).
			On("down", a.handleHistoryDown).
			On("j", a.handleHistoryDown).
			On("tab", a.handleHistoryTab).
			On("enter", a.handleHistoryEnter).
			On("ctrl+r", a.handleHistoryRewind).
			On("esc", a.handleHistoryEsc).
			Build(),
		ModeFileHistory: NewModeHandlers().
			On("up", a.handleFileHistoryUp).
			On("down", a.handleFileHistoryDown).
			On("k", a.handleFileHistoryUp).
			On("j", a.handleFileHistoryDown).
			On("tab", a.handleFileHistoryTab).
			On("y", a.handleFileHistoryCopy).
			On("v", a.handleFileHistoryVisualMode).
			On("esc", a.handleFileHistoryEsc).
			Build(),
		ModeConflictResolve: NewModeHandlers().
			On("up", a.handleConflictUp).
			On("k", a.handleConflictUp).
			On("down", a.handleConflictDown).
			On("j", a.handleConflictDown).
			On("tab", a.handleConflictTab).
			On(" ", a.handleConflictSpace).
			On("enter", a.handleConflictEnter).
			Build(),
		ModeClone: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeSelectBranch: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleSelectBranchEnter).
			Build(),
		ModeSetupWizard: NewModeHandlers().
			On("enter", a.handleSetupWizardEnter).
			Build(),
		ModeConfig: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			Build(),
		ModeBranchPicker: NewModeHandlers().
			On("up", a.handleBranchPickerUp).
			On("k", a.handleBranchPickerUp).
			On("down", a.handleBranchPickerDown).
			On("j", a.handleBranchPickerDown).
			On("enter", a.handleBranchPickerEnter).
			Build(),
		ModePreferences: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter).
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
			Build(),
	}

	// Merge global handlers into each mode
	for mode := range modeHandlers {
		for key, handler := range globalHandlers {
			modeHandlers[mode][key] = handler
		}
	}

	return modeHandlers
}

// rebuildMenuShortcuts dynamically registers keyboard handlers for menu shortcuts
func (a *Application) rebuildMenuShortcuts(mode AppMode) {
	if a.keyHandlers[mode] == nil {
		a.keyHandlers[mode] = make(map[string]KeyHandler)
	}

	var baseHandlers map[string]KeyHandler
	if mode == ModeMenu {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			On(" ", a.handleMenuEnter).
			Build()
	} else if mode == ModeConfig {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			On(" ", a.handleConfigMenuEnter).
			Build()
	} else if mode == ModePreferences {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter).
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
			Build()
	}

	newHandlers := make(map[string]KeyHandler)

	for key, handler := range baseHandlers {
		newHandlers[key] = handler
	}

	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash,
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	for key, handler := range globalHandlers {
		if _, exists := baseHandlers[key]; !exists {
			newHandlers[key] = handler
		}
	}

	// Register shortcuts for current menu items
	for i, item := range a.menuItems {
		if item.Shortcut != "" && item.Enabled && !item.Separator {
			itemIndex := i
			itemID := item.ID
			itemHint := item.Hint

			newHandlers[item.Shortcut] = func(app *Application) (tea.Model, tea.Cmd) {
				app.selectedIndex = itemIndex
				app.footerHint = itemHint
				return app, app.dispatchAction(itemID)
			}
		}
	}

	a.keyHandlers[mode] = newHandlers
}
```

**Step 2: Remove from app.go**
- Delete `buildKeyHandlers()` (lines 428-568)
- Delete `rebuildMenuShortcuts()` (lines 570-650)

**Verification:**
```bash
cd /Users/jreng/Documents/Poems/dev/tit
go build ./...
# Must compile without errors
```

**Expected Result:** app.go: 627 → 407 lines

---

## Summary of Changes

| Phase | Action | Lines Removed | app.go After |
|-------|--------|---------------|--------------|
| 1 | Delete delegation methods | 171 | 804 |
| 2 | Extract constructor | 177 | 627 |
| 3 | Extract key handlers | 220 | **407** |

**Final Result: app.go ~407 lines** ✅ (within 300-400 target)

---

## Final app.go Structure (~407 lines)

```go
package app

// Imports (14 lines)

// Application struct (96 lines)
type Application struct { ... }

// ModeTransition struct (8 lines)
type ModeTransition struct { ... }

// Core helpers (~100 lines)
func (a *Application) transitionTo(config ModeTransition) { ... }
func (a *Application) reloadGitState() error { ... }
func (a *Application) checkForConflicts(...) *GitOperationMsg { ... }
func (a *Application) executeGitOp(...) tea.Cmd { ... }

// Async helpers (~45 lines)
func (a *Application) startAsyncOp() { ... }
func (a *Application) endAsyncOp() { ... }
// ... etc

// Menu handlers (~60 lines)
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) { ... }
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) { ... }
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) { ... }

// Input helpers (~50 lines)
func (a *Application) insertTextAtCursor(text string) { ... }
func (a *Application) deleteAtCursor() { ... }
func (a *Application) updateInputValidation() { ... }

// Input handlers (~35 lines)
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) { ... }
func (a *Application) handleHistoryRewind(app *Application) (tea.Model, tea.Cmd) { ... }
```

---

## Critical Rules

1. **Clean Build After EVERY Phase**
   - Run `go build ./...` before proceeding
   - No warnings, no errors

2. **No Logic Changes**
   - Only move or delete code
   - Keep all existing functionality

3. **Maintain Code Style**
   - Use positive checks (not negative early returns)
   - Match existing indentation and formatting
   - Keep comments intact

4. **Verify Call Sites**
   - Phase 1: Confirm delegation methods have zero callers
   - Phase 2: Confirm constructor only called from main
   - Phase 3: Confirm key handlers only reference existing methods

---

## Rollback Plan

If any phase fails:
1. Stop immediately
2. Do not proceed to next phase
3. Report issue to user
4. User decides: fix forward or git revert

---

**End of Kickoff Plan**

Ready for ENGINEER to begin Phase 1 (delete dead delegation methods).
