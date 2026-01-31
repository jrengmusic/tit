package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Dirty Pull Confirmation Handlers
// ========================================

// executeConfirmDirtyPull handles YES response to dirty pull confirmation (Save changes)
func (a *Application) executeConfirmDirtyPull() (tea.Model, tea.Cmd) {
	// User confirmed to save changes and proceed with dirty pull
	a.dialogState.Hide()

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", true) // true = preserve changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))

	// Start the operation chain - Phase 1: Snapshot
	return a, a.cmdDirtyPullSnapshot(true)
}

// executeRejectDirtyPull handles NO response to dirty pull confirmation (Discard changes)
func (a *Application) executeRejectDirtyPull() (tea.Model, tea.Cmd) {
	// User chose to discard changes and pull
	a.dialogState.Hide()

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", false) // false = discard changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))

	// Start the operation chain - Phase 1: Snapshot (will discard instead of stash)
	return a, a.cmdDirtyPullSnapshot(false)
}

// ========================================
// Pull Merge Confirmation Handlers
// ========================================

// executeConfirmPullMerge handles YES response to pull merge confirmation
func (a *Application) executeConfirmPullMerge() (tea.Model, tea.Cmd) {
	// User confirmed to proceed with pull merge
	a.dialogState.Hide()

	// Transition to console to show streaming output
	a.setExitAllowed(false) // Block Ctrl+C until operation completes or is aborted
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// Start pull operation with merge strategy (--no-rebase)
	return a, a.cmdPull()
}

// executeRejectPullMerge handles NO response to pull merge confirmation
func (a *Application) executeRejectPullMerge() (tea.Model, tea.Cmd) {
	// User cancelled pull merge
	a.dialogState.Hide()
	return a.returnToMenu()
}
