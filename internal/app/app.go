package app

import (
	"context"
	"time"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Use ui.FileHistoryState and ui.FileHistoryPane (no duplication)

// Application is the central state container for the TIT (Terminal Interface for Time Travel) application.
// It manages all application state, UI rendering, git operations, and user interactions.
//
// The Application follows a strict UI THREAD / WORKER THREAD separation pattern:
// - UI THREAD: Handles rendering, input, and immediate user feedback
// - WORKER THREAD: Executes async git operations and state updates
//
// Key Responsibilities:
// - Maintains current application mode and state
// - Manages git repository state and operations
// - Handles user input and menu navigation
// - Coordinates async operations and UI updates
// - Enforces application invariants and contracts

type Application struct {
	width             int
	height            int
	sizing            ui.DynamicSizing
	theme             ui.Theme
	mode              AppMode // Current application mode
	quitConfirmActive bool
	quitConfirmTime   time.Time
	footerHint        string // Footer hint/message text
	gitState          *git.State
	selectedIndex     int                               // Current menu item selection
	menuItems         []MenuItem                        // Cached menu items
	keyHandlers       map[AppMode]map[string]KeyHandler // Cached key handlers

	// Input mode state
	inputState InputState // Text input field state

	// Workflow state (clone, init, mode restoration)
	workflowState WorkflowState // Transient multi-step workflow state

	// Remote operation state

	// Async operation state
	asyncState AsyncState

	// Console output state (for clone, init, etc)
	consoleState ConsoleState

	// Process cancellation
	cancelContext context.CancelFunc

	// Confirmation dialog state
	dialogState DialogState

	// Conflict resolution state
	conflictResolveState *ConflictResolveState

	// Dirty operation tracking
	dirtyOperationState *DirtyOperationState // nil when no dirty op in progress

	// State display info maps
	workingTreeInfo map[git.WorkingTree]StateInfo
	timelineInfo    map[git.Timeline]StateInfo
	operationInfo   map[git.Operation]StateInfo

	// Picker state (history, file history, branch picker)
	pickerState PickerState

	// Time Travel state
	timeTravelState TimeTravelState

	// Environment state (git detection + setup wizard)
	environmentState EnvironmentState // Git environment and setup wizard state

	// History cache
	cacheManager *CacheManager

	// Config state (Session 86)
	appConfig *config.Config // Loaded from ~/.config/tit/config.toml

	// Preferences state (Session 86)

	// Activity tracking (Session 2 - Lazy auto-update)
	activityState ActivityState
}

// ModeTransition configuration for streamlined mode changes
type ModeTransition struct {
	Mode        AppMode
	InputPrompt string
	InputAction string
	FooterHint  string
	InputHeight int
	ResetFields []string
}

// transitionTo handles standardized mode transitions and state resets.
func (a *Application) transitionTo(config ModeTransition) {
	a.mode = config.Mode

	// Always reset common input state
	a.selectedIndex = 0
	a.inputState.Reset()
	a.inputState.ClearConfirming = false

	// Set new input config from the transition configuration
	if config.InputPrompt != "" {
		a.inputState.Prompt = config.InputPrompt
	}
	if config.InputAction != "" {
		a.inputState.Action = config.InputAction
	}
	if config.FooterHint != "" {
		a.footerHint = config.FooterHint
	}
	if config.InputHeight > 0 {
		a.inputState.Height = config.InputHeight
	} else if config.Mode == ModeInput || config.Mode == ModeCloneURL {
		// Default to single-line input (4 = label + 3-line box)
		a.inputState.Height = 4
	}

	// Reset workflow-specific fields based on the configuration
	for _, field := range config.ResetFields {
		switch field {
		case "clone":
			a.workflowState.ResetClone()
		case "all":
			// Reset all workflow states
			a.workflowState.ResetClone()
		}
	}
}

// reloadGitState refreshes git state from repository.
// This is SSOT for all git state reloads in the application.
func (a *Application) reloadGitState() error {
	state, err := git.DetectState()
	if err != nil {
		return err
	}
	a.gitState = state
	return nil
}

// checkForConflicts detects if git is in conflicted state after an operation.
// Returns GitOperationMsg if conflicts detected, nil otherwise.
// successFlag: set to true when caller wants to trigger conflict resolver (e.g., dirty pull merge)
// successFlag: set to false for normal conflict detection during operations
func (a *Application) checkForConflicts(step string, successFlag bool) *GitOperationMsg {
	if err := a.reloadGitState(); err != nil {
		return nil
	}
	if a.gitState.Operation == git.Conflicted {
		return &GitOperationMsg{
			Step:             step,
			Success:          successFlag,
			ConflictDetected: true,
			Error:            "Merge conflicts detected",
		}
	}
	return nil
}

// executeGitOp executes a git command and returns appropriate message.
// This is SSOT for git command execution with standard error handling.
func (a *Application) executeGitOp(step string, args ...string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, args...)
		if !result.Success {
			return GitOperationMsg{
				Step:    step,
				Success: false,
				Error:   result.Stderr,
			}
		}
		return GitOperationMsg{
			Step:    step,
			Success: true,
		}
	}
}

// ========================================
// Async Operation State Helpers
// ========================================
// SSOT for async operation lifecycle.

// startAsyncOp marks an async operation as active.
func (a *Application) startAsyncOp() {
	a.asyncState.Start()
}

// endAsyncOp marks an async operation as complete.
func (a *Application) endAsyncOp() {
	a.asyncState.End()
}

// abortAsyncOp marks an async operation as aborted by user.
func (a *Application) abortAsyncOp() {
	a.asyncState.Abort()
}

// isAsyncActive returns true if an async operation is running.
func (a *Application) isAsyncActive() bool {
	return a.asyncState.IsActive()
}

// isAsyncAborted returns true if current async operation was aborted.
func (a *Application) isAsyncAborted() bool {
	return a.asyncState.IsAborted()
}

// clearAsyncAborted resets the aborted flag.
func (a *Application) clearAsyncAborted() {
	a.asyncState.ClearAborted()
}

// setExitAllowed sets whether exit is allowed during operation.
func (a *Application) setExitAllowed(allowed bool) {
	a.asyncState.SetExitAllowed(allowed)
}

// canExit returns true if exit is allowed.
func (a *Application) canExit() bool {
	return a.asyncState.CanExit()
}

// handleCacheProgress handles cache building progress updates

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if len(app.menuItems) > 0 {
		startIdx := app.selectedIndex
		app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for app.menuItems[app.selectedIndex].Separator || !app.menuItems[app.selectedIndex].Enabled {
			app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
			// Prevent infinite loop if all items disabled
			if app.selectedIndex == startIdx {
				break
			}
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuDown moves selection down
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if len(app.menuItems) > 0 {
		startIdx := app.selectedIndex
		app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for app.menuItems[app.selectedIndex].Separator || !app.menuItems[app.selectedIndex].Enabled {
			app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
			// Prevent infinite loop if all items disabled
			if app.selectedIndex == startIdx {
				break
			}
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuEnter selects current menu item and dispatches action
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.menuItems) {
		return app, nil
	}
	item := app.menuItems[app.selectedIndex]

	// CONTRACT: Cannot execute separators or disabled items (cache still building)
	if item.Separator || !item.Enabled {
		return app, nil
	}

	// Dispatch action
	return app, app.dispatchAction(item.ID)
}

// Input mode helpers

// insertTextAtCursor inserts text at current cursor position (UTF-8 safe)
func (a *Application) insertTextAtCursor(text string) {
	// Defensive bounds checking
	valueLen := len(a.inputState.Value)
	if a.inputState.CursorPosition < 0 {
		a.inputState.CursorPosition = 0
	}
	if a.inputState.CursorPosition > valueLen {
		a.inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputState.Value[:a.inputState.CursorPosition]
	after := a.inputState.Value[a.inputState.CursorPosition:]
	a.inputState.Value = before + text + after
	a.inputState.CursorPosition += len(text)
}

// deleteAtCursor deletes character before cursor (UTF-8 safe)
func (a *Application) deleteAtCursor() {
	valueLen := len(a.inputState.Value)
	if a.inputState.CursorPosition <= 0 || valueLen == 0 {
		return
	}
	if a.inputState.CursorPosition > valueLen {
		a.inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputState.Value[:a.inputState.CursorPosition-1]
	after := a.inputState.Value[a.inputState.CursorPosition:]
	a.inputState.Value = before + after
	a.inputState.CursorPosition--
}

// updateInputValidation updates validation message for current input
func (a *Application) updateInputValidation() {
	if a.inputState.Action == "clone_url" {
		currentValue := a.inputState.Value
		if a.mode == ModeInitializeBranches {
			return // No validation in branch mode
		}
		if currentValue == "" {
			a.inputState.ValidationMsg = ""
		} else if ui.ValidateRemoteURL(currentValue) {
			a.inputState.ValidationMsg = ""
		} else {
			a.inputState.ValidationMsg = "Invalid URL format"
		}
	}
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	switch app.inputState.Action {
	case "init_branch_name":
		return app.handleInitBranchNameSubmit()
	case "init_subdir_name":
		return app.handleInputSubmitSubdirName(app)
	case "add_remote_url":
		return app.handleAddRemoteSubmit(app)
	case "commit_message":
		return app.handleCommitSubmit(app)
	case "commit_push_message":
		return app.handleCommitPushSubmit(app)
	default:
		return app, nil
	}
}

// handleHistoryRewind handles Ctrl+ENTER in history browser to initiate rewind
func (a *Application) handleHistoryRewind(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil || len(app.pickerState.History.Commits) == 0 {
		return app, nil
	}

	if app.pickerState.History.SelectedIdx < 0 || app.pickerState.History.SelectedIdx >= len(app.pickerState.History.Commits) {
		return app, nil
	}

	selectedCommit := app.pickerState.History.Commits[app.pickerState.History.SelectedIdx]
	app.workflowState.PendingRewindCommit = selectedCommit.Hash

	return app, app.showRewindConfirmation(selectedCommit.Hash)
}
