package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Branch Switch Confirmation Handlers
// ========================================

// executeConfirmBranchSwitchClean handles YES response to clean tree branch switch
func (a *Application) executeConfirmBranchSwitchClean() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	targetBranch := a.dialogState.GetContext()["targetBranch"]
	if targetBranch == "" {
		return a.returnToMenu()
	}

	// Transition to console to show streaming output
	a.prepareAsyncOperation("Switching branch...")

	// Clean tree - perform branch switch directly
	return a, a.cmdSwitchBranch(targetBranch)
}

// executeRejectBranchSwitch handles NO/Cancel response (clean tree)
func (a *Application) executeRejectBranchSwitch() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Return to branch picker (preserve state)
	a.mode = ModeBranchPicker
	return a, nil
}

// executeConfirmBranchSwitchDirty handles YES response (Stash changes via Dirty Operation Protocol)
func (a *Application) executeConfirmBranchSwitchDirty() (tea.Model, tea.Cmd) {
	targetBranch := a.dialogState.GetContext()["targetBranch"]
	a.dialogState.Hide()

	if targetBranch == "" {
		return a.returnToMenu()
	}

	// Initialize dirty operation state
	a.dirtyOperationState = NewDirtyOperationState("dirty_switch", true)
	a.dirtyOperationState.TargetBranch = targetBranch
	a.dirtyOperationState.OriginalBranch = a.gitState.CurrentBranch

	a.prepareAsyncOperation("Saving changes and switching branch...")
	return a, a.cmdDirtySwitchSnapshot(true)
}

// executeRejectBranchSwitchDirty handles NO response (Discard changes via Dirty Operation Protocol)
func (a *Application) executeRejectBranchSwitchDirty() (tea.Model, tea.Cmd) {
	targetBranch := a.dialogState.GetContext()["targetBranch"]
	a.dialogState.Hide()

	if targetBranch == "" {
		return a.returnToMenu()
	}

	// Initialize dirty operation state (discard mode)
	a.dirtyOperationState = NewDirtyOperationState("dirty_switch", false)
	a.dirtyOperationState.TargetBranch = targetBranch
	a.dirtyOperationState.OriginalBranch = a.gitState.CurrentBranch

	a.prepareAsyncOperation("Discarding changes and switching branch...")
	return a, a.cmdDirtySwitchSnapshot(false)
}
