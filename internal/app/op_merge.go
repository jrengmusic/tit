package app

import (
	"context"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdFinalizePullMerge finalizes a merge by committing staged changes
// Called after user resolves conflicts in conflict resolver for pull merge
func (a *Application) cmdFinalizePullMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizePullMerge,
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the merge
		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizePullMerge,
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_finalized"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpFinalizePullMerge,
			Success: true,
			Output:  OutputMessages["merge_finalized"],
		}
	}
}

// cmdFinalizeBranchSwitch stages and commits resolved branch switch conflicts
// Called after user resolves conflicts in conflict resolver for branch switching
func (a *Application) cmdFinalizeBranchSwitch() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_branch_switch",
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the resolution (no merge commit needed, just stage the resolved files)
		// The branch switch itself is already done, we're just resolving the conflicts
		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Resolved branch switch conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_branch_switch",
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append("Branch switch conflicts resolved and committed", ui.TypeInfo)
		return GitOperationMsg{
			Step:    "finalize_branch_switch",
			Success: true,
			Output:  "Branch switch completed successfully",
		}
	}
}

// cmdAbortMerge aborts an in-progress merge
// Called when user presses ESC in conflict resolver during pull merge
func (a *Application) cmdAbortMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Abort the merge
		result := git.ExecuteWithStreaming(ctx, "merge", "--abort")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_abort_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpAbortMerge,
				Success: false,
				Error:   ErrorMessages["failed_abort_merge"],
			}
		}

		// Reset working tree to remove conflict markers
		result = git.ExecuteWithStreaming(ctx, "reset", "--hard")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_reset_after_abort"], ui.TypeWarning)
			// Non-fatal: merge state is cleared, just working tree has stale markers
		}

		buffer.Append(OutputMessages["merge_aborted"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpAbortMerge,
			Success: true,
			Output:  OutputMessages["merge_aborted"],
		}
	}
}
