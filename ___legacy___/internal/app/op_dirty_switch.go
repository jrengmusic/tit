package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdDirtySwitchSnapshot creates a git stash and saves the snapshot state
// Phase 1: Capture original branch/HEAD, then stash
func (a *Application) cmdDirtySwitchSnapshot(preserveChanges bool) tea.Cmd {
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

		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		if !branchResult.Success {
			return GitOperationMsg{
				Step:    OpDirtySwitchSnapshot,
				Success: false,
				Error:   ErrorMessages["failed_get_current_branch"],
			}
		}
		currentBranch := strings.TrimSpace(branchResult.Stdout)

		headResult := git.Execute("rev-parse", "HEAD")
		if !headResult.Success {
			return GitOperationMsg{
				Step:    OpDirtySwitchSnapshot,
				Success: false,
				Error:   "Failed to get current HEAD",
			}
		}
		currentHead := strings.TrimSpace(headResult.Stdout)

		snapshot := &git.DirtyOperationSnapshot{}
		if err := snapshot.Save(currentBranch, currentHead); err != nil {
			return GitOperationMsg{
				Step:    OpDirtySwitchSnapshot,
				Success: false,
				Error:   fmt.Sprintf("Failed to save snapshot: %v", err),
			}
		}

		if preserve {
			result := git.ExecuteWithStreaming(ctx, "stash", "push", "-u", "-m", "TIT DIRTY-SWITCH SNAPSHOT")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    OpDirtySwitchSnapshot,
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
					Step:    OpDirtySwitchSnapshot,
					Success: false,
					Error:   "Failed to discard changes",
				}
			}
			result = git.ExecuteWithStreaming(ctx, "clean", "-fd")
			if !result.Success {
				snapshot.Delete()
				return GitOperationMsg{
					Step:    OpDirtySwitchSnapshot,
					Success: false,
					Error:   "Failed to clean untracked files",
				}
			}
			buffer.Append(OutputMessages["changes_discarded"], ui.TypeInfo)
		}

		return GitOperationMsg{
			Step:    OpDirtySwitchSnapshot,
			Success: true,
			Output:  "Snapshot created, tree cleaned",
		}
	}
}

// cmdDirtySwitchExecute switches to the target branch after snapshot
// Phase 2: Tree is clean from stash/discard, perform the switch
func (a *Application) cmdDirtySwitchExecute() tea.Cmd {
	targetBranch := a.dirtyOperationState.TargetBranch
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(OutputMessages["dirty_switch_started"]+" (%s)", targetBranch), ui.TypeInfo)

		result := git.ExecuteWithStreaming(ctx, "switch", targetBranch)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpDirtySwitchExecute,
				Success: false,
				Error:   fmt.Sprintf("Failed to switch to %s: %s", targetBranch, result.Stderr),
			}
		}

		buffer.Append(fmt.Sprintf("Switched to %s", targetBranch), ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtySwitchExecute,
			Success: true,
			Output:  fmt.Sprintf("Switched to %s", targetBranch),
		}
	}
}

// cmdDirtySwitchApplySnapshot applies stashed changes on the new branch
// Phase 3: After switch succeeds, reapply saved changes
func (a *Application) cmdDirtySwitchApplySnapshot() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		stashListResult := git.Execute("stash", "list")
		if !strings.Contains(stashListResult.Stdout, "TIT DIRTY-SWITCH SNAPSHOT") {
			buffer.Append("No stash to apply (changes were discarded)", ui.TypeInfo)
			return GitOperationMsg{
				Step:    OpDirtySwitchApplySnapshot,
				Success: true,
				Output:  "No stashed changes to reapply",
			}
		}

		result := git.ExecuteWithStreaming(ctx, "stash", "apply")
		if !result.Success {
			if msg := a.checkForConflicts(OpDirtySwitchApplySnapshot, true); msg != nil {
				buffer.Append(OutputMessages["stash_apply_conflicts_detected"], ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    OpDirtySwitchApplySnapshot,
				Success: false,
				Error:   "Failed to reapply stash",
			}
		}

		buffer.Append(OutputMessages["changes_reapplied"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtySwitchApplySnapshot,
			Success: true,
			Output:  "Stashed changes reapplied",
		}
	}
}

// cmdDirtySwitchFinalize drops the stash and cleans up snapshot
// Phase 4: All phases complete, cleanup
func (a *Application) cmdDirtySwitchFinalize() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_switch_finalize_started"], ui.TypeInfo)

		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-SWITCH SNAPSHOT") {
			result := git.ExecuteWithStreaming(ctx, "stash", "drop")
			if !result.Success {
				buffer.Append(OutputMessages["stash_drop_failed_warning"], ui.TypeWarning)
			}
		}

		if err := git.CleanupSnapshot(); err != nil {
			buffer.Append(fmt.Sprintf("Warning: Failed to cleanup snapshot file: %v", err), ui.TypeWarning)
		}

		return GitOperationMsg{
			Step:    OpDirtySwitchFinalize,
			Success: true,
			Output:  "",
		}
	}
}

// cmdAbortDirtySwitch restores the exact original state
func (a *Application) cmdAbortDirtySwitch() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_switch_aborting"], ui.TypeWarning)

		// CRITICAL: Abort merge first if merge is in progress
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
				Step:    OpDirtySwitchAbort,
				Success: false,
				Error:   fmt.Sprintf("Failed to load snapshot for abort: %v", err),
			}
		}

		result := git.ExecuteWithStreaming(ctx, "checkout", snapshot.OriginalBranch)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_checkout_original_branch"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpDirtySwitchAbort,
				Success: false,
				Error:   fmt.Sprintf("Failed to checkout %s", snapshot.OriginalBranch),
			}
		}

		result = git.ExecuteWithStreaming(ctx, "reset", "--hard", snapshot.OriginalHead)
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_to_original_head"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpDirtySwitchAbort,
				Success: false,
				Error:   "Failed to reset to original HEAD",
			}
		}

		stashListResult := git.Execute("stash", "list")
		if strings.Contains(stashListResult.Stdout, "TIT DIRTY-SWITCH SNAPSHOT") {
			result = git.ExecuteWithStreaming(ctx, "stash", "apply")
			if !result.Success {
				buffer.Append(ErrorMessages["stash_reapply_failed_but_restored"], ui.TypeWarning)
			}
			git.ExecuteWithStreaming(ctx, "stash", "drop")
		}

		snapshot.Delete()

		buffer.Append(OutputMessages["original_state_restored"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpDirtySwitchAbort,
			Success: true,
			Output:  "Abort completed, original state restored",
		}
	}
}
