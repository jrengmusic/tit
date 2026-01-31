package app

import (
	"context"
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdDirtyPullSnapshot creates a git stash and saves the snapshot state
// Phase 1: Capture original branch/HEAD, then git stash push -u
func (a *Application) cmdDirtyPullSnapshot(preserveChanges bool) tea.Cmd {
	preserve := preserveChanges // Capture in closure
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		if preserve {
			buffer.Append(OutputMessages["saving_changes_stash"], ui.TypeInfo)
		} else {
			buffer.Append(OutputMessages["discarding_changes"], ui.TypeInfo)
		}

		// Get current branch name
		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		if !branchResult.Success {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   "Failed to get current branch",
			}
		}
		currentBranch := strings.TrimSpace(branchResult.Stdout)

		// Get current HEAD commit hash
		headResult := git.Execute("rev-parse", "HEAD")
		if !headResult.Success {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   "Failed to get current HEAD",
			}
		}
		currentHead := strings.TrimSpace(headResult.Stdout)

		// Save snapshot to .git/TIT_DIRTY_OP
		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Save(currentBranch, currentHead); err != nil {
			return GitOperationMsg{
				Step:    "dirty_pull_snapshot",
				Success: false,
				Error:   fmt.Sprintf("Failed to save snapshot: %v", err),
			}
		}

		// Create stash with uncommitted changes
		if preserve {
			result := git.ExecuteWithStreaming(ctx, "stash", "push", "-u", "-m", "TIT DIRTY-PULL SNAPSHOT")
			if !result.Success {
				snapshot.Delete() // Cleanup on failure
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to stash changes",
				}
			}
			buffer.Append(OutputMessages["changes_saved_stashed"], ui.TypeInfo)
		} else {
			// Discard changes without stash
			result := git.ExecuteWithStreaming(ctx, "reset", "--hard")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to discard changes",
				}
			}
			result = git.ExecuteWithStreaming(ctx, "clean", "-fd")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    "dirty_pull_snapshot",
					Success: false,
					Error:   "Failed to clean untracked files",
				}
			}
			buffer.Append(OutputMessages["changes_discarded"], ui.TypeInfo)
		}

		return GitOperationMsg{
			Step:    "dirty_pull_snapshot",
			Success: true,
			Output:  "Snapshot created, tree cleaned",
		}
	}
}

// cmdDirtyPullApplySnapshot applies the stashed changes back to the tree
// Phase 3: After pull succeeds, reapply saved changes
func (a *Application) cmdDirtyPullApplySnapshot() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["reapplying_changes"], ui.TypeInfo)

		// Check if there's a stash to apply
		stashListResult := git.Execute("stash", "list")
		if !strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			buffer.Append("No stash to apply (changes were discarded)", ui.TypeInfo) // No SSOT entry needed - contextual message
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: true,
				Output:  "No stashed changes to reapply",
			}
		}

		result := git.ExecuteWithStreaming(ctx, "stash", "apply")
		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects stash apply conflicts by checking git state (unmerged files)
			if msg := a.checkForConflicts("dirty_pull_apply_snapshot", true); msg != nil {
				buffer.Append(OutputMessages["stash_apply_conflicts_detected"], ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    "dirty_pull_apply_snapshot",
				Success: false,
				Error:   "Failed to reapply stash",
			}
		}

		buffer.Append(OutputMessages["changes_reapplied"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_apply_snapshot",
			Success: true,
			Output:  "Stashed changes reapplied",
		}
	}
}

// cmdDirtyPullFinalize drops the stash and cleans up the snapshot file
// Phase 4: After all operations succeed, finalize
func (a *Application) cmdDirtyPullFinalize() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_finalize_started"], ui.TypeInfo)

		// Drop the stash (if it exists)
		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			result := git.ExecuteWithStreaming(ctx, "stash", "drop")
			if !result.Success {
				buffer.Append(OutputMessages["stash_drop_failed_warning"], ui.TypeWarning)
				// Continue anyway - snapshot file cleanup is more important
			}
		}

		// Delete the snapshot file
		if err := git.CleanupSnapshot(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to cleanup snapshot file: %v", err), ui.TypeWarning)
			// Non-fatal, but warn user
		}

		buffer.Append(OutputMessages["dirty_pull_completed_successfully"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_finalize",
			Success: true,
			Output:  "Dirty pull finalized",
		}
	}
}

// cmdAbortDirtyPull restores the exact original state
func (a *Application) cmdAbortDirtyPull() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_aborting"], ui.TypeWarning)

		// CRITICAL: Abort merge first if merge is in progress
		state, _ := git.DetectState()
		if state != nil && state.Operation == git.Conflicted {
			buffer.Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			result := git.ExecuteWithStreaming(ctx, "merge", "--abort")
			if !result.Success {
				buffer.Append("Warning: merge abort failed, continuing with restore", ui.TypeWarning)
			}
		}

		// Load snapshot
		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Load(); err != nil {
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   fmt.Sprintf("Failed to load snapshot for abort: %v", err),
			}
		}

		// Checkout original branch
		result := git.ExecuteWithStreaming(ctx, "checkout", snapshot.OriginalBranch)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_checkout_original_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   fmt.Sprintf("Failed to checkout %s", snapshot.OriginalBranch),
			}
		}

		// Reset to original HEAD
		result = git.ExecuteWithStreaming(ctx, "reset", "--hard", snapshot.OriginalHead)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_to_original_head"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "dirty_pull_abort",
				Success: false,
				Error:   "Failed to reset to original HEAD",
			}
		}

		// Reapply stash (if it exists)
		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-PULL SNAPSHOT") {
			result = git.ExecuteWithStreaming(ctx, "stash", "apply")
			if !result.Success {
				buffer.Append(ErrorMessages["stash_reapply_failed_but_restored"], ui.TypeWarning)
			}
			git.ExecuteWithStreaming(ctx, "stash", "drop")
		}

		// Delete the snapshot file
		snapshot.Delete()

		buffer.Append(OutputMessages["original_state_restored"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_abort",
			Success: true,
			Output:  "Abort completed, original state restored",
		}
	}
}
