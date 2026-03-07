package app

import (
	"context"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdPushSyncMerge fetches and merges remote into local, then returns result.
// Called when push was rejected due to divergence.
func (a *Application) cmdPushSyncMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.OperationState.SetCancelContext(cancel)
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Append("Remote has new commits - syncing before push...", ui.TypeInfo)

		result := git.ExecuteWithStreaming(ctx, "pull", "--no-rebase")
		if !result.Success {
			if msg := a.checkForConflicts(OpPushSyncMerge, true); msg != nil {
				buffer.Append("Conflicts detected - opening resolver...", ui.TypeWarning)
				return *msg
			}
			return GitOperationMsg{
				Step:    OpPushSyncMerge,
				Success: false,
				Error:   "Failed to merge remote changes",
			}
		}

		buffer.Append("Remote changes merged - pushing...", ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpPushSyncMerge,
			Success: true,
		}
	}
}

// cmdFinalizePushSyncMerge stages resolved conflicts, commits the merge, then pushes.
// Called after user resolves conflicts in the push sync flow.
func (a *Application) cmdFinalizePushSyncMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.OperationState.SetCancelContext(cancel)
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpFinalizePushSync,
				Success: false,
				Error:   "Failed to stage resolved files",
			}
		}

		result = git.ExecuteWithStreaming(ctx, "commit", "--no-edit")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpFinalizePushSync,
				Success: false,
				Error:   "Failed to commit merge",
			}
		}

		buffer.Append("Merge committed - pushing...", ui.TypeInfo)
		return GitOperationMsg{
			Step:    OpFinalizePushSync,
			Success: true,
		}
	}
}

// cmdPushAfterSync performs the final push after a successful sync merge.
func (a *Application) cmdPushAfterSync() tea.Cmd {
	branch := ""
	hasUpstream := true
	if a.gitState != nil {
		branch = a.gitState.CurrentBranch
		hasUpstream = a.gitState.LocalBranchOnRemote
	}
	ctx, cancel := context.WithCancel(context.Background())
	a.OperationState.SetCancelContext(cancel)
	return func() tea.Msg {
		var result git.CommandResult
		if !hasUpstream {
			result = git.ExecuteWithStreaming(ctx, "push", "-u", "origin", branch)
		} else {
			result = git.ExecuteWithStreaming(ctx, "push")
		}
		if !result.Success {
			return GitOperationMsg{
				Step:    OpPush,
				Success: false,
				Error:   "Failed to push after sync",
			}
		}
		return GitOperationMsg{
			Step:    OpPush,
			Success: true,
		}
	}
}
