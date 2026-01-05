package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// ========================================
// Async Operation Commands
// ========================================
// All cmd* functions return tea.Cmd that execute async git operations
// They capture state in closures to avoid race conditions
// Each returns a GitOperationMsg when complete

// cmdInit executes `git init` and creates initial branch
func (a *Application) cmdInit(branchName string) tea.Cmd {
	name := branchName // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git init
		result := git.ExecuteWithStreaming("init")
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to initialize repository",
			}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming("checkout", "-b", name)
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to create branch",
			}
		}

		return GitOperationMsg{
			Step:    "init",
			Success: true,
			Output:  fmt.Sprintf("Repository initialized with branch '%s'", name),
		}
	}
}

// cmdClone executes `git clone` with streaming output
func (a *Application) cmdClone(url, targetPath string) tea.Cmd {
	u := url             // Capture in closure
	path := targetPath
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Run git clone with streaming output
		result := git.ExecuteWithStreaming("clone", u, path)
		if !result.Success {
			return GitOperationMsg{
				Step:    "clone",
				Success: false,
				Error:   "Failed to clone repository",
			}
		}

		// Change to cloned directory if needed
		if path != "." {
			if err := os.Chdir(path); err != nil {
				return GitOperationMsg{
					Step:    "clone",
					Success: false,
					Error:   fmt.Sprintf("Failed to change directory: %v", err),
				}
			}
		}

		// Verify .git exists
		cwd, _ := os.Getwd()
		gitDir := filepath.Join(cwd, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			return GitOperationMsg{
				Step:    "clone",
				Success: false,
				Error:   fmt.Sprintf("Clone completed but .git not found"),
			}
		}

		return GitOperationMsg{
			Step:    "clone",
			Success: true,
			Output:  "Repository cloned successfully",
		}
	}
}

// cmdAddRemote adds a remote repository
func (a *Application) cmdAddRemote(url string) tea.Cmd {
	u := url // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Add remote
		result := git.ExecuteWithStreaming("remote", "add", "origin", u)
		if !result.Success {
			return GitOperationMsg{
				Step:    "add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		// Fetch from remote
		result = git.ExecuteWithStreaming("fetch", "--all")
		if !result.Success {
			return GitOperationMsg{
				Step:    "add_remote",
				Success: false,
				Error:   "Failed to fetch from remote",
			}
		}

		// Set upstream tracking
		result = git.SetUpstreamTracking()
		if !result.Success {
			// Non-fatal: remote was added and fetched
			return GitOperationMsg{
				Step:    "add_remote",
				Success: true,
				Output:  "Remote added and fetched (tracking not set, may be detached HEAD)",
			}
		}

		return GitOperationMsg{
			Step:    "add_remote",
			Success: true,
			Output:  "Remote added, fetched, and tracking configured",
		}
	}
}

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
				Step:    "commit",
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Commit
		result = git.ExecuteWithStreaming("commit", "-m", msg)
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit",
				Success: false,
				Error:   "Failed to commit",
			}
		}

		return GitOperationMsg{
			Step:    "commit",
			Success: true,
			Output:  "Changes committed successfully",
		}
	}
}

// cmdPush pushes current branch to remote
func (a *Application) cmdPush() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Push to upstream
		result := git.ExecuteWithStreaming("push")
		if !result.Success {
			return GitOperationMsg{
				Step:    "push",
				Success: false,
				Error:   "Failed to push",
			}
		}

		return GitOperationMsg{
			Step:    "push",
			Success: true,
			Output:  "Pushed successfully",
		}
	}
}

// cmdPull pulls from remote (merge)
func (a *Application) cmdPull() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Pull (merge by default)
		result := git.ExecuteWithStreaming("pull")
		if !result.Success {
			// Check if conflict
			if strings.Contains(result.Stderr, "conflict") || strings.Contains(result.Stderr, "CONFLICT") {
				return GitOperationMsg{
					Step:    "pull",
					Success: false,
					Error:   "Merge conflicts occurred",
				}
			}
			return GitOperationMsg{
				Step:    "pull",
				Success: false,
				Error:   "Failed to pull",
			}
		}

		return GitOperationMsg{
			Step:    "pull",
			Success: true,
			Output:  "Pulled successfully",
		}
	}
}

// cmdPullRebase pulls from remote (rebase)
func (a *Application) cmdPullRebase() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Pull with rebase
		result := git.ExecuteWithStreaming("pull", "--rebase")
		if !result.Success {
			// Check if conflict
			if strings.Contains(result.Stderr, "conflict") || strings.Contains(result.Stderr, "CONFLICT") {
				return GitOperationMsg{
					Step:    "pull",
					Success: false,
					Error:   "Rebase conflicts occurred",
				}
			}
			return GitOperationMsg{
				Step:    "pull",
				Success: false,
				Error:   "Failed to pull with rebase",
			}
		}

		return GitOperationMsg{
			Step:    "pull",
			Success: true,
			Output:  "Pulled with rebase successfully",
		}
	}
}
