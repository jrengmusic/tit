package app

import (
	"context"
	"fmt"
	"os"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Git workflow commands: clone, commit, push

// startCloneOperation sets up async state and executes clone
func (a *Application) startCloneOperation() (tea.Model, tea.Cmd) {
	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0
	a.mode = ModeClone
	a.consoleState.Reset()
	a.consoleState.Clear()
	a.footerHint = GetFooterMessageText(MessageClone)

	// Return BOTH the clone worker AND periodic refresh ticker
	return a, tea.Batch(
		a.cmdCloneWorkflow(),
		a.cmdRefreshConsole(),
	)
}

// cmdCloneWorkflow launches git clone in a worker and returns a command
func (a *Application) cmdCloneWorkflow() tea.Cmd {
	cloneURL := a.workflowState.CloneURL
	cloneMode := a.workflowState.CloneMode

	cwd, _ := os.Getwd()

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		effectivePath := cwd // Default to current working directory

		if cloneMode == "here" {
			// Clone here: init + remote add + fetch + checkout with tracking
			buffer := ui.GetBuffer()

			// Step 1: git init
			result := git.ExecuteWithStreaming(ctx, "init")
			if !result.Success {
				return GitOperationMsg{Step: OpClone, Success: false, Error: "git init failed", Path: effectivePath}
			}

			// Step 2: git remote add origin <url>
			result = git.ExecuteWithStreaming(ctx, "remote", "add", "origin", cloneURL)
			if !result.Success {
				return GitOperationMsg{Step: OpClone, Success: false, Error: "git remote add failed", Path: effectivePath}
			}

			// Step 3: Query remote default branch BEFORE fetch (using git ls-remote)
			buffer.Append("Querying remote default branch...", ui.TypeStatus)
			defaultBranch, err := git.GetRemoteDefaultBranch()
			if err != nil {
				return GitOperationMsg{
					Step:    OpClone,
					Success: false,
					Error:   fmt.Sprintf("Failed to determine default branch: %v", err),
					Path:    effectivePath,
				}
			}
			buffer.Append(fmt.Sprintf("Remote default branch: %s", defaultBranch), ui.TypeStatus)

			// Step 4: Fetch all refs
			result = git.ExecuteWithStreaming(ctx, "fetch", "--all", "--progress")
			if !result.Success {
				return GitOperationMsg{Step: OpClone, Success: false, Error: "git fetch failed", Path: effectivePath}
			}

			// Step 5: Create and checkout local branch tracking remote
			// This sets up upstream automatically: -t = --track (sets upstream to origin/<branch>)
			result = git.ExecuteWithStreaming(ctx, "checkout", "-t", "origin/"+defaultBranch)
			if !result.Success {
				return GitOperationMsg{
					Step:    OpClone,
					Success: false,
					Error:   fmt.Sprintf("git checkout -t origin/%s failed: unable to create local tracking branch", defaultBranch),
					Path:    effectivePath,
				}
			}
		} else {
			// Clone to subdir: git clone creates subdir with repo name automatically
			// Don't specify a path - git will create it from the repo name
			result := git.ExecuteWithStreaming(ctx, "clone", "--progress", cloneURL)
			if !result.Success {
				return GitOperationMsg{
					Step:    OpClone,
					Success: false,
					Error:   fmt.Sprintf("Clone failed with exit code %d", result.ExitCode),
					Path:    effectivePath,
				}
			}

			// Extract repo name and change to that directory
			repoName := git.ExtractRepoName(cloneURL)
			newPath := fmt.Sprintf("%s/%s", cwd, repoName)
			if err := os.Chdir(newPath); err != nil {
				return GitOperationMsg{
					Step:    "clone",
					Success: false,
					Error:   fmt.Sprintf("Failed to change to cloned directory: %v", err),
					Path:    effectivePath,
				}
			}
			effectivePath = newPath
		}

		return GitOperationMsg{
			Step:    "clone",
			Success: true,
			Path:    effectivePath,
		}
	}
}

// cmdCommitWorkflow launches git commit in a worker and returns a command
func (a *Application) cmdCommitWorkflow(message string) tea.Cmd {
	// UI THREAD - Capturing state before spawning worker
	commitMessage := message

	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		// Stage all changes first
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit",
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Create commit
		result = git.ExecuteWithStreaming(ctx, "commit", "-m", commitMessage)
		if !result.Success {
			// Could be nothing to commit, or actual error
			// Check if working tree is clean
			checkResult := git.Execute("status", "--porcelain")
			if checkResult.Stdout == "" {
				// Nothing to commit - this is OK
				return GitOperationMsg{
					Step:    "commit",
					Success: true,
					Output:  ErrorMessages["working_tree_clean"],
				}
			}
			return GitOperationMsg{
				Step:    "commit",
				Success: false,
				Error:   "Failed to create commit",
			}
		}

		return GitOperationMsg{
			Step:    "commit",
			Success: true,
			Output:  "Commit created successfully",
		}
	}
}

// cmdPushWorkflow launches git push in a worker and returns a command
func (a *Application) cmdPushWorkflow() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming(ctx, "push")
		if !result.Success {
			return GitOperationMsg{
				Step:    "push",
				Success: false,
				Error:   "Failed to push to remote",
			}
		}

		return GitOperationMsg{
			Step:    "push",
			Success: true,
			Output:  "Push completed successfully",
		}
	}
}

// cmdAddRemoteWorkflow just does the add remote step
// Rest handled by three-step chain in githandlers.go (add_remote → fetch_remote → complete)
func (a *Application) cmdAddRemoteWorkflow(remoteURL string) tea.Cmd {
	url := remoteURL
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "add", "origin", url)
		if !result.Success {
			return GitOperationMsg{
				Step:    "add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}
		return GitOperationMsg{
			Step:    "add_remote",
			Success: true,
			Output:  "Remote added",
		}
	}
}
