package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Console output handlers

// handleConsoleUp scrolls console up one line
func (a *Application) handleConsoleUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	app.consoleState.ScrollUp()
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsoleDown scrolls console down one line
func (a *Application) handleConsoleDown(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	app.consoleState.ScrollDown()
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsolePageUp scrolls console up one page
func (a *Application) handleConsolePageUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	// UI THREAD - Scroll console up by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollUp()
	}
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsolePageDown scrolls console down one page
func (a *Application) handleConsolePageDown(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	// UI THREAD - Scroll console down by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollDown()
	}
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// cmdRefreshConsole sends periodic refresh messages while async operation is active
// This forces UI re-renders to display streaming output in real-time
func (a *Application) cmdRefreshConsole() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return OutputRefreshMsg{}
	})
}

// cmdRefreshCacheProgress sends periodic refresh messages while cache is building
// This forces UI re-renders to show cache building progress counter
// Returns a tea.Cmd that schedules continuous ticks until both caches complete
func (a *Application) cmdRefreshCacheProgress() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return CacheRefreshTickMsg{}
	})
}

// cmdFetchRemote runs git fetch in background to sync remote refs
// Called on startup when HasRemote is detected to ensure timeline accuracy
func cmdFetchRemote() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "fetch", "--quiet")
		err := cmd.Run()
		if err != nil {
			return RemoteFetchMsg{Success: false, Error: err.Error()}
		}
		return RemoteFetchMsg{Success: true}
	}
}

// startCloneOperation sets up async state and executes clone
func (a *Application) startCloneOperation() (tea.Model, tea.Cmd) {
	a.startAsyncOp()
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0
	a.mode = ModeClone
	a.consoleState.Reset()
	a.outputBuffer.Clear()
	a.footerHint = GetFooterMessageText(MessageClone)

	// Return BOTH the clone worker AND periodic refresh ticker
	return a, tea.Batch(
		a.cmdCloneWorkflow(),
		a.cmdRefreshConsole(),
	)
}

// cmdCloneWorkflow launches git clone in a worker and returns a command
func (a *Application) cmdCloneWorkflow() tea.Cmd {
	cloneURL := a.cloneURL
	cloneMode := a.cloneMode

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

// cmdSwitchBranch performs git switch to the target branch
// Handles conflicts if they occur during the switch (files that would be overwritten)
func (a *Application) cmdSwitchBranch(targetBranch string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Execute git switch
		result := git.ExecuteWithStreaming(ctx, "switch", targetBranch)
		if !result.Success {
			// Check if we're in a conflicted state after failed switch
			// This can happen when switching would overwrite local changes
			if msg := a.checkForConflicts("branch_switch", false); msg != nil {
				buffer.Append(fmt.Sprintf("Conflicts detected while switching to %s", targetBranch), ui.TypeWarning)
				msg.BranchName = targetBranch
				msg.Error = fmt.Sprintf("Conflicts switching to %s", targetBranch)
				return *msg
			}

			// Other failure (permissions, invalid branch, etc)
			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   fmt.Sprintf("Failed to switch to %s: %s", targetBranch, result.Stderr),
			}
		}

		// Success - Update() will refresh state automatically
		buffer.Append(fmt.Sprintf("Switched to branch %s", targetBranch), ui.TypeInfo)
		return GitOperationMsg{
			Step:       "branch_switch",
			Success:    true,
			Output:     fmt.Sprintf("Switched to branch %s", targetBranch),
			BranchName: targetBranch,
		}
	}
}

// cmdBranchSwitchWithStash performs: stash → switch → stash apply
func (a *Application) cmdBranchSwitchWithStash(targetBranch string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Step1: Stash changes
		buffer.Append("Stashing changes...", ui.TypeStatus)
		stashResult := git.Execute("stash", "push", "-u")
		if !stashResult.Success {
			buffer.Append(fmt.Sprintf("Failed to stash: %s", stashResult.Stderr), ui.TypeStderr)
			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   "Failed to stash changes",
			}
		}
		buffer.Append("Changes stashed", ui.TypeStatus)

		// Step2: Switch branch
		buffer.Append(fmt.Sprintf("Switching to %s...", targetBranch), ui.TypeStatus)
		switchResult := git.ExecuteWithStreaming(ctx, "switch", targetBranch)
		if !switchResult.Success {
			buffer.Append(fmt.Sprintf("Failed to switch: %s", switchResult.Stderr), ui.TypeStderr)

			// Try to restore stash on failure
			buffer.Append("Restoring stash...", ui.TypeStatus)
			git.Execute("stash", "pop")

			return GitOperationMsg{
				Step:    "branch_switch",
				Success: false,
				Error:   fmt.Sprintf("Failed to switch to %s", targetBranch),
			}
		}
		buffer.Append(fmt.Sprintf("Switched to %s", targetBranch), ui.TypeStatus)

		// Step3: Restore stash
		buffer.Append("Restoring changes...", ui.TypeStatus)
		applyResult := git.Execute("stash", "pop")
		if !applyResult.Success {
			buffer.Append("Warning: Stash apply failed (conflicts or errors)", ui.TypeWarning)
			buffer.Append("Your changes are still in stash (use 'git stash apply')", ui.TypeInfo)
		} else {
			buffer.Append("Changes restored", ui.TypeStatus)
		}

		return GitOperationMsg{
			Step:    "branch_switch",
			Success: true,
			Output:  fmt.Sprintf("Switched to %s", targetBranch),
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
					Output:  "Nothing to commit (working tree clean)",
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
