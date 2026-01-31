package app

import (
	"context"

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
	// Embedded state clusters
	*UIState
	*NavigationState
	*OperationState
	*DialogManager

	// Core business logic (standalone)
	gitState        *git.State
	workingTreeInfo map[git.WorkingTree]StateInfo
	timelineInfo    map[git.Timeline]StateInfo
	operationInfo   map[git.Operation]StateInfo

	// Feature-specific state (standalone)
	timeTravelState  TimeTravelState
	environmentState EnvironmentState

	// Infrastructure (standalone)
	cacheManager  *CacheManager
	appConfig     *config.Config
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
	a.NavigationState.SetMode(config.Mode)

	// Always reset common input state
	inputState := a.OperationState.GetInputState()
	a.NavigationState.SetSelectedIndex(0)
	inputState.Reset()
	inputState.ClearConfirming = false

	// Set new input config from the transition configuration
	if config.InputPrompt != "" {
		inputState.Prompt = config.InputPrompt
	}
	if config.InputAction != "" {
		inputState.Action = config.InputAction
	}
	if config.FooterHint != "" {
		a.UIState.SetFooterHint(config.FooterHint)
	}
	if config.InputHeight > 0 {
		inputState.Height = config.InputHeight
	} else if config.Mode == ModeInput || config.Mode == ModeCloneURL {
		// Default to single-line input (label + 3-line box)
		inputState.Height = InputHeight
	}

	// Reset workflow-specific fields based on the configuration
	workflowState := a.OperationState.GetWorkflowState()
	for _, field := range config.ResetFields {
		switch field {
		case "clone":
			workflowState.ResetClone()
		case "all":
			// Reset all workflow states
			workflowState.ResetClone()
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
	if err := a.reloadGitState(); err == nil {
		if a.gitState.Operation == git.Conflicted {
			return &GitOperationMsg{
				Step:             step,
				Success:          successFlag,
				ConflictDetected: true,
				Error:            "Merge conflicts detected",
			}
		}
	}
	return nil
}

// executeGitOp executes a git command and returns appropriate message.
// This is SSOT for git command execution with standard error handling.
func (a *Application) executeGitOp(step string, args ...string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.OperationState.SetCancelContext(cancel)
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, args...)
		if result.Success {
			return GitOperationMsg{
				Step:    step,
				Success: true,
			}
		}
		return GitOperationMsg{
			Step:    step,
			Success: false,
			Error:   result.Stderr,
		}
	}
}

// ========================================
// Async Operation State Helpers
// ========================================
// SSOT for async operation lifecycle.

// startAsyncOp marks an async operation as active.
func (a *Application) startAsyncOp() {
	a.OperationState.StartAsyncOp()
}

// endAsyncOp marks an async operation as complete.
func (a *Application) endAsyncOp() {
	a.OperationState.EndAsyncOp()
}

// abortAsyncOp marks an async operation as aborted by user.
func (a *Application) abortAsyncOp() {
	a.OperationState.AbortAsyncOp()
}

// isAsyncActive returns true if an async operation is running.
func (a *Application) isAsyncActive() bool {
	return a.OperationState.IsAsyncActive()
}

// isAsyncAborted returns true if current async operation was aborted.
func (a *Application) isAsyncAborted() bool {
	return a.OperationState.IsAsyncAborted()
}

// clearAsyncAborted resets the aborted flag.
func (a *Application) clearAsyncAborted() {
	a.OperationState.ClearAsyncAborted()
}

// setExitAllowed sets whether exit is allowed during operation.
func (a *Application) setExitAllowed(allowed bool) {
	a.OperationState.SetExitAllowed(allowed)
}

// canExit returns true if exit is allowed.
func (a *Application) canExit() bool {
	return a.OperationState.CanExit()
}

// handleCacheProgress handles cache building progress updates

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	menuItems := app.NavigationState.GetMenuItems()
	if len(menuItems) > 0 {
		startIdx := app.NavigationState.GetSelectedIndex()
		newIdx := (startIdx - 1 + len(menuItems)) % len(menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for menuItems[newIdx].Separator || !menuItems[newIdx].Enabled {
			newIdx = (newIdx - 1 + len(menuItems)) % len(menuItems)
			// Prevent infinite loop if all items disabled
			if newIdx == startIdx {
				break
			}
		}
		app.NavigationState.SetSelectedIndex(newIdx)
		// Update footer hint
		if newIdx < len(menuItems) {
			app.UIState.SetFooterHint(menuItems[newIdx].Hint)
		}
	}
	return app, nil
}

// handleMenuDown moves selection down
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	menuItems := app.NavigationState.GetMenuItems()
	if len(menuItems) > 0 {
		startIdx := app.NavigationState.GetSelectedIndex()
		newIdx := (startIdx + 1) % len(menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for menuItems[newIdx].Separator || !menuItems[newIdx].Enabled {
			newIdx = (newIdx + 1) % len(menuItems)
			// Prevent infinite loop if all items disabled
			if newIdx == startIdx {
				break
			}
		}
		app.NavigationState.SetSelectedIndex(newIdx)
		// Update footer hint
		if newIdx < len(menuItems) {
			app.UIState.SetFooterHint(menuItems[newIdx].Hint)
		}
	}
	return app, nil
}

// handleMenuEnter selects current menu item and dispatches action
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	item, ok := app.NavigationState.GetSelectedItem()
	if ok {
		// CONTRACT: Cannot execute separators or disabled items (cache still building)
		if !item.Separator && item.Enabled {
			// Dispatch action
			return app, app.dispatchAction(item.ID)
		}
	}
	return app, nil
}

// Input mode helpers

// insertTextAtCursor inserts text at current cursor position (UTF-8 safe)
func (a *Application) insertTextAtCursor(text string) {
	inputState := a.OperationState.GetInputState()
	// Defensive bounds checking
	valueLen := len(inputState.Value)
	if inputState.CursorPosition < 0 {
		inputState.CursorPosition = 0
	}
	if inputState.CursorPosition > valueLen {
		inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := inputState.Value[:inputState.CursorPosition]
	after := inputState.Value[inputState.CursorPosition:]
	inputState.Value = before + text + after
	inputState.CursorPosition += len(text)
}

// deleteAtCursor deletes character before cursor (UTF-8 safe)
func (a *Application) deleteAtCursor() {
	inputState := a.OperationState.GetInputState()
	valueLen := len(inputState.Value)
	if inputState.CursorPosition <= 0 || valueLen == 0 {
		return
	}
	if inputState.CursorPosition > valueLen {
		inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := inputState.Value[:inputState.CursorPosition-1]
	after := inputState.Value[inputState.CursorPosition:]
	inputState.Value = before + after
	inputState.CursorPosition--
}

// updateInputValidation updates validation message for current input
func (a *Application) updateInputValidation() {
	inputState := a.OperationState.GetInputState()
	if inputState.Action == "clone_url" {
		currentValue := inputState.Value
		if a.NavigationState.GetMode() == ModeInitializeBranches {
			return // No validation in branch mode
		}
		if currentValue == "" {
			inputState.ValidationMsg = ""
		} else if ui.ValidateRemoteURL(currentValue) {
			inputState.ValidationMsg = ""
		} else {
			inputState.ValidationMsg = "Invalid URL format"
		}
	}
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	inputState := app.OperationState.GetInputState()
	switch inputState.Action {
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
	pickerState := app.DialogManager.GetPickerState()
	if pickerState.History == nil || len(pickerState.History.Commits) == 0 {
		return app, nil
	}

	if pickerState.History.SelectedIdx < 0 || pickerState.History.SelectedIdx >= len(pickerState.History.Commits) {
		return app, nil
	}

	selectedCommit := pickerState.History.Commits[pickerState.History.SelectedIdx]
	app.OperationState.GetWorkflowState().PendingRewindCommit = selectedCommit.Hash

	return app, app.showRewindConfirmation(selectedCommit.Hash)
}
