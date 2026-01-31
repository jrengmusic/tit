package app

import (
	"context"
	"strings"

	"tit/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// Git pull operations: merge, rebase

// cmdPullMergeWorkflow launches git pull (merge) in a worker and returns a command
func (a *Application) cmdPullMergeWorkflow() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming(ctx, "pull")
		if !result.Success {
			// Check if conflict occurred
			if strings.Contains(result.Stderr, "CONFLICT") || strings.Contains(result.Stdout, "CONFLICT") {
				return GitOperationMsg{
					Step:    "pull_merge",
					Success: false,
					Error:   "Merge conflict detected - resolve manually",
				}
			}
			return GitOperationMsg{
				Step:    "pull_merge",
				Success: false,
				Error:   "Failed to pull from remote",
			}
		}

		return GitOperationMsg{
			Step:    "pull_merge",
			Success: true,
			Output:  "Pull completed successfully",
		}
	}
}

// cmdPullRebaseWorkflow launches git pull --rebase in a worker and returns a command
func (a *Application) cmdPullRebaseWorkflow() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming(ctx, "pull", "--rebase")
		if !result.Success {
			// Check if conflict occurred
			if strings.Contains(result.Stderr, "CONFLICT") || strings.Contains(result.Stdout, "CONFLICT") {
				return GitOperationMsg{
					Step:    "pull_rebase",
					Success: false,
					Error:   "Rebase conflict detected - resolve manually",
				}
			}
			return GitOperationMsg{
				Step:    "pull_rebase",
				Success: false,
				Error:   "Failed to pull from remote",
			}
		}

		return GitOperationMsg{
			Step:    "pull_rebase",
			Success: true,
			Output:  "Pull completed successfully",
		}
	}
}
