package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

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

// ========================================
// Time Travel Confirmation Handlers
// ========================================

// executeConfirmTimeTravel handles YES response to time travel confirmation
func (a *Application) executeConfirmTimeTravel() (tea.Model, tea.Cmd) {
	// Get commit hash from context BEFORE hiding dialog (Hide() clears context)
	commitHash := a.dialogState.GetContext()["commit_hash"]

	// User confirmed time travel - hide dialog now that we have the hash
	a.dialogState.Hide()

	// Get original branch - CHECK BEFORE modifying Operation state
	// Case 1: Already time traveling (Operation == TimeTraveling) → read from TIT_TIME_TRAVEL file
	// Case 2: Normal operation → get current branch from HEAD
	var originalBranch string
	wasAlreadyTimeTraveling := a.gitState.Operation == git.TimeTraveling

	if wasAlreadyTimeTraveling {
		// Already time traveling - read original branch from marker file
		existingBranch, _, err := git.GetTimeTravelInfo()
		if err != nil {
			a.footerHint = ErrorMessages["failed_get_current_branch"]
			return a, nil
		}
		originalBranch = existingBranch
	} else {
		// Normal operation - get current branch from HEAD
		currentBranchResult := git.Execute("rev-parse", "--abbrev-ref", "HEAD")
		if !currentBranchResult.Success {
			a.footerHint = ErrorMessages["failed_get_current_branch"]
			return a, nil
		}
		originalBranch = strings.TrimSpace(currentBranchResult.Stdout)

		// CRITICAL: If at detached HEAD (originalBranch == "HEAD"), try to get actual branch
		if originalBranch == "HEAD" {
			// Try to get default branch from remote tracking
			defaultBranchResult := git.Execute("symbolic-ref", "refs/remotes/origin/HEAD")
			if defaultBranchResult.Success {
				// Output is like "refs/remotes/origin/main", extract "main"
				parts := strings.Split(strings.TrimSpace(defaultBranchResult.Stdout), "/")
				if len(parts) > 0 {
					originalBranch = parts[len(parts)-1]
				}
			} else {
				// Fallback to "main" (most common default)
				originalBranch = "main"
			}
		}
	}

	// CRITICAL: Set Operation to TimeTraveling AFTER getting original branch
	// This prevents Phase 0 restoration from triggering if app restarts during time travel
	a.gitState.Operation = git.TimeTraveling

	// Check if working tree is dirty
	if a.gitState.WorkingTree == git.Dirty {
		// Handle dirty working tree - stash changes first
		return a.executeTimeTravelWithDirtyTree(originalBranch, commitHash)
	} else {
		// Clean working tree - proceed directly
		return a.executeTimeTravelClean(originalBranch, commitHash)
	}
}

// executeTimeTravelClean handles time travel from clean working tree
func (a *Application) executeTimeTravelClean(originalBranch, commitHash string) (tea.Model, tea.Cmd) {
	// Write time travel info (no stash ID for clean tree)
	err := git.WriteTimeTravelInfo(originalBranch, "")
	if err != nil {
		// Use standardized fatal error logging (PATTERN: Invariant violation)
		LogErrorFatal("Failed to write time travel info", err)
	}

	// Build TimeTravelInfo directly (fail fast if git calls fail)
	// This prevents silent failures and inconsistent state
	// CRITICAL: Validate commitHash is non-empty BEFORE git operations
	if commitHash == "" {
		LogErrorFatal("Failed to get commit subject", fmt.Errorf("commit hash is empty"))
	}
	commitSubject := strings.TrimSpace(git.Execute("log", "-1", "--format=%s", commitHash).Stdout)
	if commitSubject == "" {
		LogErrorFatal("Failed to get commit subject", fmt.Errorf("empty subject for %s", commitHash))
	}

	commitTimeStr := strings.TrimSpace(git.Execute("log", "-1", "--format=%aI", commitHash).Stdout)
	if commitTimeStr == "" {
		LogErrorFatal("Failed to get commit time", fmt.Errorf("empty time for %s", commitHash))
	}

	commitTime, err := time.Parse(time.RFC3339, commitTimeStr)
	if err != nil {
		LogErrorFatal("Failed to parse commit time", err)
	}

	a.timeTravelState.SetInfo(&git.TimeTravelInfo{
		OriginalBranch:  originalBranch,
		OriginalStashID: "",
		CurrentCommit: git.CommitInfo{
			Hash:    commitHash,
			Subject: commitSubject,
			Time:    commitTime,
		},
	})

	// Transition to console to show streaming output
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()

	a.workflowState.PreviousMode = ModeHistory
	a.workflowState.PreviousMenuIndex = 0

	// CRITICAL: Set restoreTimeTravelInitiated = true to prevent restoration check
	// from triggering during this intentional time travel session
	a.timeTravelState.MarkRestoreInitiated()

	// Start time travel checkout operation
	return a, git.ExecuteTimeTravelCheckout(originalBranch, commitHash)
}

// executeTimeTravelWithDirtyTree handles time travel from dirty working tree
func (a *Application) executeTimeTravelWithDirtyTree(originalBranch, commitHash string) (tea.Model, tea.Cmd) {
	// Transition to console to show streaming output (MUST happen BEFORE buffer operations)
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()

	a.workflowState.PreviousMode = ModeHistory
	a.workflowState.PreviousMenuIndex = 0
	a.timeTravelState.MarkRestoreInitiated()

	buffer := ui.GetBuffer()

	// Stash changes first
	buffer.Append("Stashing changes...", ui.TypeStatus)
	stashResult := git.Execute("stash", "push", "-u", "-m", "TIT_TIME_TRAVEL")
	if !stashResult.Success {
		buffer.Append(fmt.Sprintf("Failed to stash changes: %s", stashResult.Stderr), ui.TypeStderr)
		a.footerHint = ErrorMessages["failed_stash_changes"]
		a.endAsyncOp()
		return a, nil
	}

	buffer.Append("Stash created successfully", ui.TypeStatus)

	// Get stash ID
	stashListResult := git.Execute("stash", "list")
	if !stashListResult.Success {
		buffer.Append(fmt.Sprintf("Failed to get stash list: %s", stashListResult.Stderr), ui.TypeStderr)
		a.footerHint = ErrorMessages["failed_get_stash_list"]
		a.endAsyncOp()
		return a, nil
	}

	// Find the stash we just created (should be stash@{0})
	stashRef := ""
	lines := strings.Split(stashListResult.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "TIT_TIME_TRAVEL") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				// Remove trailing colon from stash reference (e.g., "stash@{0}:" → "stash@{0}")
				stashRef = strings.TrimSuffix(parts[0], ":")
				buffer.Append(fmt.Sprintf("[DEBUG] Found stash reference: %s", stashRef), ui.TypeStatus)
				break
			}
		}
	}

	if stashRef == "" {
	}

	// Convert stash reference to SHA hash (stable, doesn't shift like stash@{N})
	// This prevents bugs when stash indices shift due to new stashes being created
	if stashRef == "" {
		panic("FATAL: No stash reference found after creating stash. This should never happen.")
	}

	hashResult := git.Execute("rev-parse", stashRef)
	if !hashResult.Success {
		panic(fmt.Sprintf("FATAL: Failed to convert stash reference to hash: %s", hashResult.Stderr))
	}

	stashHash := strings.TrimSpace(hashResult.Stdout)

	// Get current working directory (absolute repo path)
	repoPath, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
	}

	// Check if entry already exists (handles stale state from previous TIT session)
	if _, exists := config.GetStashEntry("time_travel", repoPath); exists {
		// Clean up stale entry first
		config.RemoveStashEntry("time_travel", repoPath)
	}

	// Add stash entry to config tracking system
	config.AddStashEntry("time_travel", stashHash, repoPath, originalBranch, commitHash)

	// Build TimeTravelInfo directly from commit hash (fail fast if git calls fail)
	// This prevents silent failures and inconsistent state
	commitSubject := strings.TrimSpace(git.Execute("log", "-1", "--format=%s", commitHash).Stdout)
	if commitSubject == "" {
		panic(fmt.Sprintf("FATAL: Failed to get commit subject for %s", commitHash))
	}

	commitTimeStr := strings.TrimSpace(git.Execute("log", "-1", "--format=%aI", commitHash).Stdout)
	if commitTimeStr == "" {
		panic(fmt.Sprintf("FATAL: Failed to get commit time for %s", commitHash))
	}

	commitTime, err := time.Parse(time.RFC3339, commitTimeStr)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to parse commit time: %v", err))
	}

	a.timeTravelState.SetInfo(&git.TimeTravelInfo{
		OriginalBranch:  originalBranch,
		OriginalStashID: stashHash,
		CurrentCommit: git.CommitInfo{
			Hash:    commitHash,
			Subject: commitSubject,
			Time:    commitTime,
		},
	})

	// Start time travel checkout operation (console already set up at function start)
	return a, git.ExecuteTimeTravelCheckout(originalBranch, commitHash)
}

// executeRejectTimeTravel handles NO response to time travel confirmation
func (a *Application) executeRejectTimeTravel() (tea.Model, tea.Cmd) {
	// User cancelled time travel
	a.dialogState.Hide()
	return a.returnToMenu()
}

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

// executeConfirmRewind handles "Rewind" choice
// executeConfirmRewind executes git reset --hard at pending commit
func (a *Application) executeConfirmRewind() (tea.Model, tea.Cmd) {
	if a.workflowState.PendingRewindCommit == "" {
		return a.returnToMenu()
	}

	commitHash := a.workflowState.PendingRewindCommit
	a.workflowState.PendingRewindCommit = "" // Clear after capturing

	// Set up async operation
	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeHistory
	a.workflowState.PreviousMenuIndex = a.pickerState.History.SelectedIdx
	a.mode = ModeConsole
	a.consoleState.Reset()
	ui.GetBuffer().Clear()

	// Start rewind + refresh ticker
	return a, tea.Batch(
		a.executeRewindOperation(commitHash),
		a.cmdRefreshConsole(),
	)
}

// executeRejectRewind handles "Cancel" choice on rewind confirmation
func (a *Application) executeRejectRewind() (tea.Model, tea.Cmd) {
	a.workflowState.PendingRewindCommit = "" // Clear pending commit
	return a.returnToMenu()
}

// executeRewindOperation performs the actual git reset --hard in a worker goroutine
func (a *Application) executeRewindOperation(commitHash string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		_, err := git.ResetHardAtCommit(ctx, commitHash)

		return RewindMsg{
			Commit:  commitHash,
			Success: err == nil,
			Error:   errorOrEmpty(err),
		}
	}
}

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

	// Get stash hash
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
	// First line is our stash, parse the hash
	stashEntry := strings.TrimSpace(stashLines[0])
	stashParts := strings.Split(stashEntry, ":")
	if len(stashParts) < 2 {
		a.footerHint = "Failed to parse stash entry"
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}
	stashHash := strings.TrimSpace(stashParts[0])

	// Store stash hash in config for later retrieval
	repoPath, _ := os.Getwd()
	config.AddStashEntry("time_travel", stashHash, repoPath, targetBranch, timeTravelHash)

	// Execute merge
	return a, git.ExecuteTimeTravelMerge(targetBranch, timeTravelHash)
}
