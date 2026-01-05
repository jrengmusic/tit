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

		// Create .gitignore and commit it so tree is clean
		if err := git.CreateDefaultGitignore(); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to create .gitignore: %v", err),
			}
		}

		result = git.ExecuteWithStreaming("add", ".gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to add .gitignore",
			}
		}

		result = git.ExecuteWithStreaming("commit", "-m", "Initialize repository with .gitignore")
		if !result.Success {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   "Failed to commit .gitignore",
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

// cmdAddRemote adds a remote repository (step 1 of 3-step chain)
func (a *Application) cmdAddRemote(url string) tea.Cmd {
	u := url // Capture in closure
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Get current branch BEFORE adding remote (works even with zero commits)
		// Use symbolic-ref instead of rev-parse: works when HEAD exists but has no commits
		branchResult := git.Execute("symbolic-ref", "--short", "HEAD")
		branchName := ""
		if branchResult.Success {
		branchName = strings.TrimSpace(branchResult.Stdout)
		}

		// Add remote
		result := git.ExecuteWithStreaming("remote", "add", "origin", u)
		if !result.Success {
			return GitOperationMsg{
				Step:    "add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:       "add_remote",
			Success:    true,
			Output:     "Remote added",
			BranchName: branchName,
		}
	}
}

// cmdFetchRemote fetches from origin (step 2 of 3-step chain)
func (a *Application) cmdFetchRemote() tea.Cmd {
	return func() tea.Msg {
		result := git.ExecuteWithStreaming("fetch", "--all")
		if !result.Success {
			return GitOperationMsg{
				Step:    "fetch_remote",
				Success: false,
				Error:   "Failed to fetch from remote",
			}
		}

		return GitOperationMsg{
			Step:    "fetch_remote",
			Success: true,
			Output:  "Fetch completed",
		}
	}
}

// cmdSetUpstream sets upstream tracking (step 3 of 3-step chain)
// Takes branchName from previous step to avoid querying git again
func (a *Application) cmdSetUpstream(branchName string) tea.Cmd {
	branch := branchName // Capture in closure
	return func() tea.Msg {
		result := git.SetUpstreamTrackingWithBranch(branch)
		if !result.Success {
			// Non-fatal: tracking setup failed but remote is ready
			return GitOperationMsg{
				Step:    "set_upstream",
				Success: true, // Consider success because remote is added
				Output:  "Remote configured (upstream tracking could not be set - may be detached HEAD)",
			}
		}

		return GitOperationMsg{
			Step:    "set_upstream",
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

// cmdInitSubdirectory transitions directly to subdirectory input mode
// Used when CWD is not empty (can't init here, must use subdir)
func (a *Application) cmdInitSubdirectory() tea.Cmd {
	return func() tea.Msg {
		a.mode = ModeInput
		a.inputPrompt = InputPrompts["init_subdir_name"]
		a.inputAction = "init_subdir_name"
		a.footerHint = InputHints["init_subdir_name"]
		a.inputValue = ""
		a.inputCursorPosition = 0
		return nil
	}
}


