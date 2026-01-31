package app

import (
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

		// CRITICAL: If at detached HEAD (originalBranch == HEADRef), try to get actual branch
		if originalBranch == HEADRef {
			// Try to get default branch from remote tracking
			defaultBranchResult := git.Execute("symbolic-ref", "refs/remotes/origin/HEAD")
			if defaultBranchResult.Success {
				// Output is like "refs/remotes/origin/main", extract "main"
				parts := strings.Split(strings.TrimSpace(defaultBranchResult.Stdout), "/")
				if len(parts) > 0 {
					originalBranch = parts[len(parts)-1]
				}
			} else {
				// Fallback to DefaultBranch (most common default)
				originalBranch = DefaultBranch
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
