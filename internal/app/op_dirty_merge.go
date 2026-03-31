package app

import (
	"context"
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdDirtyMergeSnapshot creates a git stash and saves the snapshot state
// Phase 1: Capture original branch/HEAD, then git stash push -u
func (a *Application) cmdDirtyMergeSnapshot(preserveChanges bool) tea.Cmd {
	preserve := preserveChanges
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
				Step:    OpDirtyMergeSnapshot,
				Success: false,
				Error:   ErrorMessages["failed_get_current_branch"],
			}
		}
		currentBranch := strings.TrimSpace(branchResult.Stdout)

		// Get current HEAD commit hash
		headResult := git.Execute("rev-parse", "HEAD")
		if !headResult.Success {
			return GitOperationMsg{
				Step:    OpDirtyMergeSnapshot,
				Success: false,
				Error:   "Failed to get current HEAD",
			}
		}
		currentHead := strings.TrimSpace(headResult.Stdout)

		// Save snapshot to .git/TIT_DIRTY_OP
		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Save(currentBranch, currentHead); err != nil {
			return GitOperationMsg{
				Step:    OpDirtyMergeSnapshot,
				Success: false,
				Error:   fmt.Sprintf("Failed to save snapshot: %v", err),
			}
		}

		if preserve {
			result := git.ExecuteWithStreaming(ctx, "stash", "push", "-u", "-m", "TIT DIRTY-MERGE SNAPSHOT")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    OpDirtyMergeSnapshot,
					Success: false,
					Error:   ErrorMessages["failed_stash_changes"],
				}
			}
			buffer.Append(OutputMessages["changes_saved_stashed"], ui.TypeInfo)
		} else {
			result := git.ExecuteWithStreaming(ctx, "reset", "--hard")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    OpDirtyMergeSnapshot,
					Success: false,
					Error:   "Failed to discard changes",
				}
			}
			result = git.ExecuteWithStreaming(ctx, "clean", "-fd")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    OpDirtyMergeSnapshot,
					Success: false,
					Error:   "Failed to clean untracked files",
				}
			}
			buffer.Append(OutputMessages["changes_discarded"], ui.TypeInfo)
		}

		return GitOperationMsg{
			Step:    OpDirtyMergeSnapshot,
			Success: true,
			Output:  "Snapshot created, tree cleaned",
		}
	}
}

// cmdDirtyMerge merges the source branch after snapshot
// Phase 2: After stash/discard, perform the merge
func (a *Application) cmdDirtyMerge() tea.Cmd {
	sourceBranch := a.dirtyOperationState.MergeBranch
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(OutputMessages["dirty_merge_started"]+" (%s)", sourceBranch), ui.TypeInfo)

		result := git.ExecuteWithStreaming(ctx, "merge", sourceBranch)
		if !result.Success {
			if msg := a.checkForConflicts(OpDirtyMerge, true); msg != nil {
				buffer.Append(OutputMessages["merge_conflicts_detected"], ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    OpDirtyMerge,
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["merge_branch_failed"], result.Stderr),
			}
		}

		buffer.Append(OutputMessages["merge_completed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtyMerge,
			Success: true,
			Output:  fmt.Sprintf("Merged %s", sourceBranch),
		}
	}
}

// cmdFinalizeDirtyMerge finalizes the merge commit during dirty merge
// Called after user resolves merge conflicts in conflict resolver
func (a *Application) cmdFinalizeDirtyMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeDirtyMerge,
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeDirtyMerge,
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_committed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpFinalizeDirtyMerge,
			Success: true,
			Output:  "Merge commit created",
		}
	}
}

// cmdDirtyMergeApplySnapshot applies stashed changes back after merge
// Phase 3: After merge succeeds, reapply saved changes
func (a *Application) cmdDirtyMergeApplySnapshot() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["reapplying_changes"], ui.TypeInfo)

		stashListResult := git.Execute("stash", "list")
		if !strings.Contains(stashListResult.Stdout, "TIT DIRTY-MERGE SNAPSHOT") {
			buffer.Append("No stash to apply (changes were discarded)", ui.TypeInfo)
			return GitOperationMsg{
				Step:    OpDirtyMergeApplySnapshot,
				Success: true,
				Output:  "No stashed changes to reapply",
			}
		}

		result := git.ExecuteWithStreaming(ctx, "stash", "apply")
		if !result.Success {
			if msg := a.checkForConflicts(OpDirtyMergeApplySnapshot, true); msg != nil {
				buffer.Append(OutputMessages["stash_apply_conflicts_detected"], ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    OpDirtyMergeApplySnapshot,
				Success: false,
				Error:   "Failed to reapply stash",
			}
		}

		buffer.Append(OutputMessages["changes_reapplied"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtyMergeApplySnapshot,
			Success: true,
			Output:  "Stashed changes reapplied",
		}
	}
}

// cmdDirtyMergeFinalize drops the stash and cleans up the snapshot file
// Phase 4: After all operations succeed, finalize
func (a *Application) cmdDirtyMergeFinalize() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_merge_finalize_started"], ui.TypeInfo)

		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-MERGE SNAPSHOT") {
			result := git.ExecuteWithStreaming(ctx, "stash", "drop")
			if !result.Success {
				buffer.Append(OutputMessages["stash_drop_failed_warning"], ui.TypeWarning)
			}
		}

		if err := git.CleanupSnapshot(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to cleanup snapshot file: %v", err), ui.TypeWarning)
		}

		buffer.Append(OutputMessages["dirty_merge_completed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtyMergeFinalize,
			Success: true,
			Output:  "Dirty merge finalized",
		}
	}
}

// cmdAbortDirtyMerge restores the exact original state
func (a *Application) cmdAbortDirtyMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_merge_aborting"], ui.TypeWarning)

		// Abort merge first if in progress
		state, _ := git.DetectState()
		if state != nil && state.Operation == git.Conflicted {
			buffer.Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			result := git.ExecuteWithStreaming(ctx, "merge", "--abort")
			if !result.Success {
				buffer.Append("Warning: merge abort failed, continuing with restore", ui.TypeWarning)
			}
		}

		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Load(); err != nil {
			return GitOperationMsg{
				Step:    OpDirtyMergeAbort,
				Success: false,
				Error:   fmt.Sprintf("Failed to load snapshot for abort: %v", err),
			}
		}

		result := git.ExecuteWithStreaming(ctx, "checkout", snapshot.OriginalBranch)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_checkout_original_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpDirtyMergeAbort,
				Success: false,
				Error:   fmt.Sprintf("Failed to checkout %s", snapshot.OriginalBranch),
			}
		}

		result = git.ExecuteWithStreaming(ctx, "reset", "--hard", snapshot.OriginalHead)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_to_original_head"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpDirtyMergeAbort,
				Success: false,
				Error:   "Failed to reset to original HEAD",
			}
		}

		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-MERGE SNAPSHOT") {
			result = git.ExecuteWithStreaming(ctx, "stash", "apply")
			if !result.Success {
				buffer.Append(ErrorMessages["stash_reapply_failed_but_restored"], ui.TypeWarning)
			}
			git.ExecuteWithStreaming(ctx, "stash", "drop")
		}

		snapshot.Delete()

		buffer.Append(OutputMessages["original_state_restored"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtyMergeAbort,
			Success: true,
			Output:  "Abort completed, original state restored",
		}
	}
}
