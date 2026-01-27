package app

import (
	"fmt"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdInit executes `git init`, creates initial branch, and commits .gitignore
func (a *Application) cmdInit(branchName string) tea.Cmd {
	name := branchName // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git init
		result := git.ExecuteWithStreaming("init")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to initialize repository",
			}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming("checkout", "-b", name)
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

		result = git.ExecuteWithStreaming("add", ".gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to add .gitignore",
			}
		}

		result = git.ExecuteWithStreaming("commit", "-m", "Initialize repository with .gitignore")
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

// cmdInitSubdirectory initializes a git repository in a subdirectory
func (a *Application) cmdInitSubdirectory() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git init in subdirectory
		result := git.ExecuteWithStreaming("init")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to initialize repository in subdirectory",
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

		result = git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		result = git.ExecuteWithStreaming("commit", "-m", "Initialize subdirectory repository")
		if !result.Success {
			return GitOperationMsg{
				Step:    OpInit,
				Success: false,
				Error:   "Failed to commit initial changes",
			}
		}

		return GitOperationMsg{
			Step:    OpInit,
			Success: true,
			Output:  "Repository initialized in subdirectory",
		}
	}
}
