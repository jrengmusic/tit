package app

import (
	"fmt"

	"tit/internal/git"

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

// executeConfirmBranchSwitchDirty handles YES response (Stash changes)
func (a *Application) executeConfirmBranchSwitchDirty() (tea.Model, tea.Cmd) {
	// Get targetBranch from context BEFORE hiding dialog (Hide() clears context)
	targetBranch := a.dialogState.GetContext()["targetBranch"]
	if targetBranch == "" {
		a.dialogState.Hide()
		return a.returnToMenu()
	}

	// User confirmed - hide dialog now that we have the branch name
	a.dialogState.Hide()

	// Transition to console
	a.prepareAsyncOperation("Switching branch with stash...")

	// Execute: stash → switch → stash apply
	return a, a.cmdBranchSwitchWithStash(targetBranch)
}

// executeRejectBranchSwitchDirty handles NO response (Discard changes)
func (a *Application) executeRejectBranchSwitchDirty() (tea.Model, tea.Cmd) {
	// Get targetBranch from context BEFORE hiding dialog (Hide() clears context)
	targetBranch := a.dialogState.GetContext()["targetBranch"]
	if targetBranch == "" {
		a.dialogState.Hide()
		return a.returnToMenu()
	}

	// User confirmed - hide dialog now that we have the branch name
	a.dialogState.Hide()

	// Discard changes first
	resetResult := git.Execute("reset", "--hard", "HEAD")
	if !resetResult.Success {
		a.footerHint = fmt.Sprintf("Failed to discard changes: %s", resetResult.Stderr)
		a.mode = ModeBranchPicker
		return a, nil
	}

	cleanResult := git.Execute("clean", "-fd")
	if !cleanResult.Success {
		a.footerHint = "Warning: Failed to clean untracked files"
	}

	// Transition to console
	a.prepareAsyncOperation(fmt.Sprintf("Switching to %s...", targetBranch))

	// Clean tree now - perform switch
	return a, a.cmdSwitchBranch(targetBranch)
}
