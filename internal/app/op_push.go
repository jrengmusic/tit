package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// cmdPush pushes current branch to remote
func (a *Application) cmdPush() tea.Cmd {
	return a.executeGitOp(OpPush, "push")
}

// cmdForcePush force-pushes current branch (use with caution)
func (a *Application) cmdForcePush() tea.Cmd {
	return a.executeGitOp(OpForcePush, "push", "--force")
}
