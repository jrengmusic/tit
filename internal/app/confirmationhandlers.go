package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/ui"
)

// ConfirmationType represents different kinds of confirmation dialogs
type ConfirmationType string

const (
	ConfirmNestedRepoInit    ConfirmationType = "nested_repo_init"
	ConfirmForcePush         ConfirmationType = "force_push"
	ConfirmHardReset         ConfirmationType = "hard_reset"
	ConfirmDestructiveOp     ConfirmationType = "destructive_op"
	ConfirmAlert             ConfirmationType = "alert"
	ConfirmPullMerge         ConfirmationType = "pull_merge"
	ConfirmPullMergeDiverged ConfirmationType = "pull_merge_diverged"
	ConfirmDirtyPull         ConfirmationType = "dirty_pull"
)

// ConfirmationAction is a function that handles a confirmed action
type ConfirmationAction func(*Application) (tea.Model, tea.Cmd)

// confirmationActions maps confirmation types to their YES handlers
var confirmationActions = map[string]ConfirmationAction{
	string(ConfirmNestedRepoInit):    (*Application).executeConfirmNestedRepoInit,
	string(ConfirmForcePush):         (*Application).executeConfirmForcePush,
	string(ConfirmHardReset):         (*Application).executeConfirmHardReset,
	string(ConfirmAlert):             (*Application).executeAlert,
	string(ConfirmDirtyPull):         (*Application).executeConfirmDirtyPull,
	string(ConfirmPullMerge):         (*Application).executeConfirmPullMerge,
	string(ConfirmPullMergeDiverged): (*Application).executeConfirmPullMerge,
}

// confirmationRejectActions maps confirmation types to their NO handlers
var confirmationRejectActions = map[string]ConfirmationAction{
	string(ConfirmNestedRepoInit):    (*Application).executeRejectNestedRepoInit,
	string(ConfirmForcePush):         (*Application).executeRejectForcePush,
	string(ConfirmHardReset):         (*Application).executeRejectHardReset,
	string(ConfirmAlert):             (*Application).executeAlert, // Any key dismisses alert
	string(ConfirmDirtyPull):         (*Application).executeRejectDirtyPull,
	string(ConfirmPullMerge):         (*Application).executeRejectPullMerge,
	string(ConfirmPullMergeDiverged): (*Application).executeRejectPullMerge,
}

// ========================================
// Confirmation Result Handler
// ========================================

// handleConfirmationResponse routes confirmation YES/NO responses to appropriate handlers
func (a *Application) handleConfirmationResponse(confirmed bool) (tea.Model, tea.Cmd) {
	if a.confirmationDialog == nil {
		// No active confirmation dialog
		return a.returnToMenu()
	}

	confirmType := a.confirmationDialog.Config.ActionID
	var handler ConfirmationAction

	if confirmed {
		handler = confirmationActions[confirmType]
	} else {
		handler = confirmationRejectActions[confirmType]
	}

	if handler == nil {
		// No handler registered for this type - return to menu
		a.confirmationDialog = nil
		return a.returnToMenu()
	}

	return handler(a)
}

// ========================================
// Confirmation Action Methods
// ========================================

// executeConfirmNestedRepoInit handles YES response to nested repo warning
func (a *Application) executeConfirmNestedRepoInit() (tea.Model, tea.Cmd) {
	// User confirmed they want to init in a nested repo
	// Return to previous mode (ModeInitializeLocation)
	a.confirmationDialog = nil
	a.mode = ModeInitializeLocation
	return a, nil
}

// executeRejectNestedRepoInit handles NO response to nested repo warning
func (a *Application) executeRejectNestedRepoInit() (tea.Model, tea.Cmd) {
	// User cancelled - abort init, return to menu
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeConfirmForcePush handles YES response to force push confirmation
func (a *Application) executeConfirmForcePush() (tea.Model, tea.Cmd) {
	// User confirmed force push
	// Initiate async push --force operation
	a.confirmationDialog = nil
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)

	return a, a.cmdForcePush()
}

// executeRejectForcePush handles NO response to force push confirmation
func (a *Application) executeRejectForcePush() (tea.Model, tea.Cmd) {
	// User cancelled force push
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeConfirmHardReset handles YES response to hard reset confirmation
func (a *Application) executeConfirmHardReset() (tea.Model, tea.Cmd) {
	// User confirmed hard reset
	a.confirmationDialog = nil
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)

	return a, a.cmdHardReset()
}

// executeRejectHardReset handles NO response to hard reset confirmation
func (a *Application) executeRejectHardReset() (tea.Model, tea.Cmd) {
	// User cancelled hard reset
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeAlert handles alert dialog dismissal (any response)
func (a *Application) executeAlert() (tea.Model, tea.Cmd) {
	// Alert dialogs are dismissed with any key press
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// ========================================
// Confirmation Dialog Creation
// ========================================

// showConfirmation displays a confirmation dialog with the given config
// Sets confirmation mode and dialog state
func (a *Application) showConfirmation(config ui.ConfirmationConfig) {
	a.confirmationDialog = ui.NewConfirmationDialog(
		config,
		ui.ContentInnerWidth,
		&a.theme,
	)
	a.mode = ModeConfirmation
}

// showNestedRepoWarning displays warning when initializing in a nested repo
func (a *Application) showNestedRepoWarning(path string) {
	config := ui.ConfirmationConfig{
		Title:       "Nested Repository Detected",
		Explanation: fmt.Sprintf("The directory '%s' is inside an existing git repository. Initialize anyway?", path),
		YesLabel:    "Yes, continue",
		NoLabel:     "No, cancel",
		ActionID:    string(ConfirmNestedRepoInit),
	}
	a.showConfirmation(config)
}

// showForcePushWarning displays warning when attempting force push
func (a *Application) showForcePushWarning(branchName string) {
	config := ui.ConfirmationConfig{
		Title:       "Force Push Confirmation",
		Explanation: fmt.Sprintf("This will force push '%s' to remote. Any remote changes will be overwritten. Continue?", branchName),
		YesLabel:    "Force push",
		NoLabel:     "Cancel",
		ActionID:    string(ConfirmForcePush),
	}
	a.showConfirmation(config)
}

// showHardResetWarning displays warning when attempting hard reset
func (a *Application) showHardResetWarning() {
	config := ui.ConfirmationConfig{
		Title:       "Hard Reset Confirmation",
		Explanation: "This will discard all uncommitted changes. This cannot be undone. Continue?",
		YesLabel:    "Reset",
		NoLabel:     "Cancel",
		ActionID:    string(ConfirmHardReset),
	}
	a.showConfirmation(config)
}

// showAlert displays a simple alert dialog (informational, dismisses with any key)
func (a *Application) showAlert(title, explanation string) {
	config := ui.ConfirmationConfig{
		Title:       title,
		Explanation: explanation,
		YesLabel:    "OK",
		NoLabel:     "", // Single button alert
		ActionID:    string(ConfirmAlert),
	}
	a.showConfirmation(config)
}

// ========================================
// Dirty Pull Confirmation Handlers
// ========================================

// executeConfirmDirtyPull handles YES response to dirty pull confirmation (Save changes)
func (a *Application) executeConfirmDirtyPull() (tea.Model, tea.Cmd) {
	// User confirmed to save changes and proceed with dirty pull
	a.confirmationDialog = nil

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", true) // true = preserve changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// Start the operation chain - Phase 1: Snapshot
	return a, a.cmdDirtyPullSnapshot(true)
}

// executeRejectDirtyPull handles NO response to dirty pull confirmation (Discard changes)
func (a *Application) executeRejectDirtyPull() (tea.Model, tea.Cmd) {
	// User chose to discard changes and pull
	a.confirmationDialog = nil

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", false) // false = discard changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// Start the operation chain - Phase 1: Snapshot (will discard instead of stash)
	return a, a.cmdDirtyPullSnapshot(false)
}

// ========================================
// Pull Merge Confirmation Handlers
// ========================================

// executeConfirmPullMerge handles YES response to pull merge confirmation
func (a *Application) executeConfirmPullMerge() (tea.Model, tea.Cmd) {
	// User confirmed to proceed with pull merge
	a.confirmationDialog = nil

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// Start pull operation with merge strategy (--no-rebase)
	return a, a.cmdPull()
}

// executeRejectPullMerge handles NO response to pull merge confirmation
func (a *Application) executeRejectPullMerge() (tea.Model, tea.Cmd) {
	// User cancelled pull merge
	a.confirmationDialog = nil
	return a.returnToMenu()
}
