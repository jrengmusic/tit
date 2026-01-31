package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/config"
	"tit/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Time Travel Merge Dirty Confirmation
// ========================================

// executeConfirmTimeTravelMergeDirtyCommit handles "Commit & merge" choice
// Auto-commits with generated message and immediately starts merge
func (a *Application) executeConfirmTimeTravelMergeDirtyCommit() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Get current commit hash for auto-generated message
	result := git.Execute("rev-parse", "--short", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	shortHash := strings.TrimSpace(result.Stdout)

	// Auto-generate commit message
	commitMessage := fmt.Sprintf("TIT: Changes from time travel to %s", shortHash)

	// Stage all changes (same as normal commit workflow)
	stageResult := git.Execute("add", "-A")
	if !stageResult.Success {
		a.footerHint = fmt.Sprintf("Failed to stage changes: %s", stageResult.Stderr)
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Commit changes immediately
	commitResult := git.Execute("commit", "-m", commitMessage)
	if !commitResult.Success {
		a.footerHint = fmt.Sprintf("Failed to commit: %s", commitResult.Stderr)
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Tree is now clean and changes committed - proceed directly with merge
	// Get full commit hash for merge operation
	fullHashResult := git.Execute("rev-parse", "HEAD")
	if !fullHashResult.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	timeTravelHash := strings.TrimSpace(fullHashResult.Stdout)

	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	// Marker file still exists during merge, but this is NOT an incomplete session
	a.timeTravelState.MarkRestoreInitiated()

	// Get original branch: prefer workflowState (for manual detached from branch picker),
	// fallback to marker file (for TIT time travel)
	originalBranch := a.workflowState.ReturnToBranchName
	if originalBranch == "" {
		// Fallback to marker file for TIT time travel
		originalBranch = a.getOriginalBranchForTimeTravel()
	}

	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// executeConfirmTimeTravelMergeDirtyDiscard handles "Discard" choice
// Discards uncommitted changes immediately, then proceeds with merge
func (a *Application) executeConfirmTimeTravelMergeDirtyDiscard() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Discard working tree changes
	if err := a.discardWorkingTreeChanges(); err != nil {
		a.footerHint = err.Error()
		a.mode = ModeMenu
		return a, nil
	}

	// Tree is now clean - proceed directly with merge (no second confirmation)
	// Get current commit hash
	result := git.Execute("rev-parse", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	timeTravelHash := strings.TrimSpace(result.Stdout)

	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	a.timeTravelState.MarkRestoreInitiated()

	// Get original branch and execute time travel merge operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// ========================================
// Time Travel Return Dirty Confirmation
// ========================================

// executeConfirmTimeTravelReturnDirtyDiscard handles "Discard & return" choice
// Discards uncommitted changes immediately, then proceeds with return
func (a *Application) executeConfirmTimeTravelReturnDirtyDiscard() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Discard working tree changes
	if err := a.discardWorkingTreeChanges(); err != nil {
		a.footerHint = err.Error()
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Tree is now clean - proceed directly with return (no second confirmation)
	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = "Returning to main..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel return!
	a.timeTravelState.MarkRestoreInitiated()

	// Get original branch and execute time travel return operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelReturn(originalBranch)
}

// ========================================
// Stale Stash Confirmation Handlers
// ========================================

// executeConfirmStaleStashContinue handles YES response to stale stash confirmation
func (a *Application) executeConfirmStaleStashContinue() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Clean up the stale TOML entry
	repoPath, _ := os.Getwd()
	config.RemoveStashEntry("time_travel", repoPath)

	// Proceed with time travel return (now without stash restore)
	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = "Returning to main..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel return!
	a.timeTravelState.MarkRestoreInitiated()

	// Get original branch and execute time travel return operation (stash will be skipped)
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelReturn(originalBranch)
}

// executeRejectStaleStashContinue handles NO response - cancel operation
func (a *Application) executeRejectStaleStashContinue() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmStaleStashMergeContinue handles YES response to stale stash confirmation during merge
func (a *Application) executeConfirmStaleStashMergeContinue() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Clean up the stale TOML entry
	repoPath, _ := os.Getwd()
	config.RemoveStashEntry("time_travel", repoPath)

	// Proceed with time travel merge (now without stash restore)
	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	a.timeTravelState.MarkRestoreInitiated()

	// Get current commit hash
	result := git.Execute("rev-parse", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		return a, nil
	}

	timeTravelHash := strings.TrimSpace(result.Stdout)
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// executeRejectTimeTravelReturnDirty handles "Cancel" choice
func (a *Application) executeRejectTimeTravelReturnDirty() (tea.Model, tea.Cmd) {
	// User cancelled return
	a.dialogState.Hide()
	a.mode = ModeMenu
	return a, a.startAutoUpdate()
}
