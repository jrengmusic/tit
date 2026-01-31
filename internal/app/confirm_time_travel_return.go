package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Time Travel Return Confirmation
// ========================================

// executeConfirmTimeTravelReturn handles YES response to return-to-main confirmation
func (a *Application) executeConfirmTimeTravelReturn() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Check if there's a stash entry that needs validation
	repoPath, _ := os.Getwd()
	stashHash, hasStash := config.FindStashEntry("time_travel", repoPath)

	// If stash exists, verify it still exists in git
	if hasStash && !git.StashExists(stashHash) {
		// Stash was manually dropped - show confirmation dialog
		a.mode = ModeConfirmation
		shortHash := git.ShortenHash(stashHash)
		dialogContext := map[string]string{
			"stash_hash":     shortHash,
			"originalBranch": a.getOriginalBranchForTimeTravel(),
		}
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       "Stash Not Found",
				Explanation: fmt.Sprintf("Original stash %s was manually dropped. Continue without restoring stash?", shortHash),
				YesLabel:    "Continue",
				NoLabel:     "Cancel",
				ActionID:    "confirm_stale_stash_continue",
			},
			a.sizing.ContentInnerWidth,
			&a.theme,
		)
		a.dialogState.Show(dialog, dialogContext)
		return a, nil
	}

	// Transition to console to show streaming output (consistent with other git operations)
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

// executeRejectTimeTravelReturn handles NO response to return-to-main confirmation
func (a *Application) executeRejectTimeTravelReturn() (tea.Model, tea.Cmd) {
	// User cancelled return
	a.dialogState.Hide()
	a.mode = ModeMenu
	return a, a.startAutoUpdate()
}

// ========================================
// Time Travel Merge Confirmation
// ========================================

// executeConfirmTimeTravelMerge handles YES response to merge-and-return confirmation
func (a *Application) executeConfirmTimeTravelMerge() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Check if there's a stash entry that needs validation
	repoPath, _ := os.Getwd()
	stashHash, hasStash := config.FindStashEntry("time_travel", repoPath)

	// If stash exists, verify it still exists in git
	if hasStash && !git.StashExists(stashHash) {
		// Stash was manually dropped - show confirmation dialog
		a.mode = ModeConfirmation
		shortHash := git.ShortenHash(stashHash)
		dialogContext := map[string]string{
			"stash_hash":     shortHash,
			"originalBranch": a.getOriginalBranchForTimeTravel(),
		}
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       "Stash Not Found",
				Explanation: fmt.Sprintf("Original stash %s was manually dropped. Continue without restoring stash?", shortHash),
				YesLabel:    "Continue",
				NoLabel:     "Cancel",
				ActionID:    "confirm_stale_stash_merge_continue",
			},
			a.sizing.ContentInnerWidth,
			&a.theme,
		)
		a.dialogState.Show(dialog, dialogContext)
		return a, nil
	}

	// Get current commit hash
	result := git.Execute("rev-parse", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		return a, nil
	}

	timeTravelHash := strings.TrimSpace(result.Stdout)

	// Transition to console to show streaming output (consistent with other git operations)
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

// executeRejectTimeTravelMerge handles NO response to merge-and-return confirmation
func (a *Application) executeRejectTimeTravelMerge() (tea.Model, tea.Cmd) {
	// User cancelled merge
	a.dialogState.Hide()
	a.mode = ModeMenu
	return a, a.startAutoUpdate()
}

// executeConfirmTimeTravelMergeDirtyStash handles "Stash changes" choice
// Stashes changes, then proceeds with merge, then applies stash back
func (a *Application) executeConfirmTimeTravelMergeDirtyStash() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()

	// Get target branch from workflow state
	targetBranch := a.workflowState.ReturnToBranchName
	if targetBranch == "" {
		targetBranch = a.getOriginalBranchForTimeTravel()
	}

	// Get current commit hash (detached HEAD)
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
	a.footerHint = "Stashing and merging back..."
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	a.timeTravelState.MarkRestoreInitiated()

	// Stash changes with message
	stashResult := git.Execute("stash", "push", "-u", "-m", "TIT time-travel merge")
	if !stashResult.Success {
		a.footerHint = fmt.Sprintf("Failed to stash changes: %s", stashResult.Stderr)
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Get stash hash - first line is our stash
	stashListResult := git.Execute("stash", "list")
	if !stashListResult.Success || stashListResult.Stdout == "" {
		a.footerHint = "Failed to get stash list"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	stashLines := strings.Split(stashListResult.Stdout, "\n")
	if len(stashLines) == 0 || stashLines[0] == "" {
		a.footerHint = "No stash entries found"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	// Parse the stash reference from first line (format: "stash@{0}: WIP on <branch>: <hash> <subject>")
	stashEntry := strings.TrimSpace(stashLines[0])
	stashParts := strings.Split(stashEntry, ":")
	if len(stashParts) < 2 {
		a.footerHint = "Failed to parse stash entry"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	// Extract the stash reference (first part before ":")
	stashRef := strings.TrimSpace(stashParts[0])

	// Convert stash reference to actual commit hash
	hashResult := git.Execute("rev-parse", stashRef)
	if !hashResult.Success {
		a.footerHint = "Failed to convert stash reference to hash"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	stashHash := strings.TrimSpace(hashResult.Stdout)

	// Get current working directory (absolute repo path)
	repoPath, err := os.Getwd()
	if err != nil {
		a.footerHint = "Failed to get current directory"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Check if entry already exists (handles stale state from previous operations)
	if _, exists := config.GetStashEntry("time_travel", repoPath); exists {
		// Clean up stale entry first
		config.RemoveStashEntry("time_travel", repoPath)
	}

	// Store stash hash in config for later retrieval
	config.AddStashEntry("time_travel", stashHash, repoPath, targetBranch, timeTravelHash)

	// Execute merge
	return a, git.ExecuteTimeTravelMerge(targetBranch, timeTravelHash)
}
