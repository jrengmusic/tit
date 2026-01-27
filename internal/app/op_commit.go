package app

import (
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdCommit stages all changes and creates a commit
func (a *Application) cmdCommit(message string) tea.Cmd {
	msg := message // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Stage all changes
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommit,
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommit,
				Success: false,
				Error:   "Failed to commit",
			}
		}

		return GitOperationMsg{
			Step:    OpCommit,
			Success: true,
			Output:  "Changes committed successfully",
		}
	}
}

// cmdCommitPush stages, commits, and pushes in one operation
func (a *Application) cmdCommitPush(message string) tea.Cmd {
	msg := message // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Stage all changes
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to commit",
			}
		}

		// Push
		result = git.ExecuteWithStreaming("push")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpCommitPush,
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    OpCommitPush,
			Success: true,
			Output:  "Committed and pushed successfully",
		}
	}
}
