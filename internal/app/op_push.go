package app

import (
	"context"

	"tit/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdPush pushes current branch to remote.
// On rejection (diverged), triggers auto sync flow instead of failing.
func (a *Application) cmdPush() tea.Cmd {
	branch := a.gitState.CurrentBranch
	hasUpstream := a.gitState.LocalBranchOnRemote
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
			// Push rejected - trigger auto sync flow
			return GitOperationMsg{
				Step:    OpPushSyncNeeded,
				Success: true,
			}
		}
		return GitOperationMsg{
			Step:    OpPush,
			Success: true,
		}
	}
}

// cmdForcePush force-pushes current branch (use with caution).
// Uses -u origin <branch> when no upstream tracking is set (first push to bare remote).
func (a *Application) cmdForcePush() tea.Cmd {
	if !a.gitState.LocalBranchOnRemote {
		return a.executeGitOp(OpForcePush, "push", "-u", "origin", a.gitState.CurrentBranch, "--force")
	}
	return a.executeGitOp(OpForcePush, "push", "--force")
}
