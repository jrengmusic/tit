package app

import (
	"context"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdDirtyPullMerge pulls from remote using merge strategy
// Phase 2: After snapshot, pull remote changes
func (a *Application) cmdDirtyPullMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append(OutputMessages["dirty_pull_merge_started"], ui.TypeInfo)

		result := git.ExecuteWithStreaming(ctx, "pull", "--no-rebase")
		if !result.Success {
			// Check if we're in a conflicted state (more reliable than parsing stderr)
			// This detects merge conflicts by checking git state (.git/MERGE_HEAD + unmerged files)
			if msg := a.checkForConflicts("dirty_pull_merge", true); msg != nil {
				buffer.Append(OutputMessages["merge_conflicts_detected"], ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    "dirty_pull_merge",
				Success: false,
				Error:   "Failed to pull",
			}
		}

		buffer.Append(OutputMessages["merge_completed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "dirty_pull_merge",
			Success: true,
			Output:  "Remote changes merged",
		}
	}
}

// cmdFinalizeDirtyPullMerge finalizes the merge commit during dirty pull
func (a *Application) cmdFinalizeDirtyPullMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_stage_resolved"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_dirty_pull_merge",
				Success: false,
				Error:   ErrorMessages["failed_stage_resolved"],
			}
		}

		// Commit the merge
		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Merge resolved conflicts")
		if !result.Success {
			buffer.Append(ErrorMessages["failed_commit_merge"], ui.TypeStderr)
			return GitOperationMsg{
				Step:    "finalize_dirty_pull_merge",
				Success: false,
				Error:   ErrorMessages["failed_commit_merge"],
			}
		}

		buffer.Append(OutputMessages["merge_committed"], ui.TypeInfo)
		return GitOperationMsg{
			Step:    "finalize_dirty_pull_merge",
			Success: true,
			Output:  "Merge commit created",
		}
	}
}
