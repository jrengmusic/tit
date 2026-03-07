package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// cmdPush pushes current branch to remote.
// Uses -u origin <branch> when no upstream tracking is set (first push to bare remote).
func (a *Application) cmdPush() tea.Cmd {
	if !a.gitState.LocalBranchOnRemote {
		return a.executeGitOp(OpPush, "push", "-u", "origin", a.gitState.CurrentBranch)
	}
	return a.executeGitOp(OpPush, "push")
}

// cmdForcePush force-pushes current branch (use with caution).
// Uses -u origin <branch> when no upstream tracking is set (first push to bare remote).
func (a *Application) cmdForcePush() tea.Cmd {
	if !a.gitState.LocalBranchOnRemote {
		return a.executeGitOp(OpForcePush, "push", "-u", "origin", a.gitState.CurrentBranch, "--force")
	}
	return a.executeGitOp(OpForcePush, "push", "--force")
}
