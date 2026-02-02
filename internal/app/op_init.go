package app

import (
	"context"
	"fmt"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdInit executes `git init`, creates initial branch, and commits .gitignore
func (a *Application) cmdInit(branchName string) tea.Cmd {
	name := branchName // Capture in closure
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git init
		result := git.ExecuteWithStreaming(ctx, "init")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to initialize repository",
			}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming(ctx, "checkout", "-b", name)
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to create branch",
			}
		}

		// Create .gitignore and commit it so tree is clean
		if err := git.CreateDefaultGitignore(); err != nil {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   fmt.Sprintf("Failed to create .gitignore: %v", err),
			}
		}

		result = git.ExecuteWithStreaming(ctx, "add", ".gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to add .gitignore",
			}
		}

		result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Initialize repository with .gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to commit .gitignore",
			}
		}

		return GitOperationMsg{
			Step:    OpInit,
			Success: true,
			Output:  fmt.Sprintf("Repository initialized with branch '%s'", name),
		}
	}
}

