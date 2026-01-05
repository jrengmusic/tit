package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/atotto/clipboard"
	"tit/internal/git"
	"tit/internal/ui"
)

// validateAndProceed is a generic input validation handler.
// It uses a validator function and proceeds with onSuccess if validation passes.
func (a *Application) validateAndProceed(
    validator ui.InputValidator,
    onSuccess func(*Application) (tea.Model, tea.Cmd),
) (tea.Model, tea.Cmd) {
    if valid, msg := validator(a.inputValue); !valid {
        a.footerHint = msg
        return a, nil
    }
    return onSuccess(a)
}

// handleKeyCtrlC handles Ctrl+C globally
// During async operations: shows "operation in progress" message
// Otherwise: prompts for confirmation before quitting
func (a *Application) handleKeyCtrlC(app *Application) (tea.Model, tea.Cmd) {
	// If async operation is running, show "in progress" message
	if a.asyncOperationActive && !a.asyncOperationAborted {
		a.footerHint = "Operation in progress. Please wait for completion."
		return a, nil
	}

	// Standard quit confirmation flow
	if a.quitConfirmActive {
		// Second Ctrl+C - quit immediately
		return a, tea.Quit
	}

	// First Ctrl+C - start confirmation timer and set footer hint
	a.quitConfirmActive = true
	a.quitConfirmTime = time.Now()
	a.footerHint = GetFooterMessageText(MessageCtrlCConfirm)
	return a, tea.Tick(QuitConfirmationTimeout, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// handleKeyESC handles ESC globally
// During async operations: aborts the operation
// In input mode with text: confirms clear (3s timeout)
// Otherwise: returns to previous menu and restores state
func (a *Application) handleKeyESC(app *Application) (tea.Model, tea.Cmd) {
	// If async operation is running: abort it
	if a.asyncOperationActive && !a.asyncOperationAborted {
		a.asyncOperationAborted = true
		a.footerHint = "Aborting operation. Please wait..."
		return a, nil
	}

	// If async operation just aborted: restore previous state
	if a.asyncOperationAborted {
		a.asyncOperationActive = false
		a.asyncOperationAborted = false
		a.mode = a.previousMode
		a.selectedIndex = a.previousMenuIndex
		a.consoleState = ui.NewConsoleOutState()
		a.outputBuffer.Clear()
		a.footerHint = ""
		
		// Regenerate menu if returning to menu mode
		if a.mode == ModeMenu {
			menu := app.GenerateMenu()
			app.menuItems = menu
			if a.previousMenuIndex < len(menu) && len(menu) > 0 {
				app.footerHint = menu[a.previousMenuIndex].Hint
			}
		}
		return a, nil
	}

	// In input mode: handle based on input content
	if a.isInputMode() {
		// If input is empty: back to menu
		if a.inputValue == "" {
			return a.returnToMenu()
		}
		
		// If clear confirm active: clear input and stay
		if a.clearConfirmActive {
			a.inputValue = ""
			a.inputCursorPosition = 0
			a.inputValidationMsg = ""
			a.clearConfirmActive = false
			a.footerHint = ""
			return a, nil
		}
		
		// First ESC with non-empty input: start clear confirmation
		a.clearConfirmActive = true
		a.footerHint = GetFooterMessageText(MessageEscClearConfirm)
		return a, tea.Tick(QuitConfirmationTimeout, func(t time.Time) tea.Msg {
			return ClearTickMsg(t)
		})
	}

	// Normal mode: return to menu
	return a.returnToMenu()
}

// returnToMenu resets state and returns to menu mode
func (a *Application) returnToMenu() (tea.Model, tea.Cmd) {
	a.mode = ModeMenu
	a.selectedIndex = 0
	a.consoleState = ui.NewConsoleOutState()
	a.outputBuffer.Clear()
	a.footerHint = ""
	a.inputValue = ""
	a.inputCursorPosition = 0
	a.inputValidationMsg = ""
	a.clearConfirmActive = false
	
	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 {
		a.footerHint = menu[0].Hint
	}
	
	return a, nil
}

// Init workflow handlers

var initLocationConfig = LocationChoiceConfig{
	PathSetter: func(a *Application, path string) { a.initRepositoryPath = path },
	OnCurrentDir: func(a *Application) (tea.Model, tea.Cmd) {
		// Transition to single branch input mode
		a.transitionTo(ModeTransition{
			Mode:        ModeInput,
			InputPrompt: "Initial branch name:",
			InputAction: "init_branch_name",
			FooterHint:  "Enter branch name (default: main), press Enter to initialize",
		})
		a.inputValue = "main"
		a.inputCursorPosition = len("main")
		return a, nil
	},
	SubdirPrompt: "Repository name:",
	SubdirAction: "init_subdir_name",
}

var cloneLocationConfig = LocationChoiceConfig{
	PathSetter: func(a *Application, path string) { a.clonePath = path },
	OnCurrentDir: func(a *Application) (tea.Model, tea.Cmd) {
		return a.startCloneOperation()
	},
	SubdirPrompt: "Directory name:",
	SubdirAction: "clone_subdir_name",
}

// handleInitLocationSelection handles enter on init location menu
func (a *Application) handleInitLocationSelection(app *Application) (tea.Model, tea.Cmd) {
	// Route to choice 1 or 2 based on selectedIndex
	if app.selectedIndex == 0 {
		return app.handleLocationChoice(1, initLocationConfig)
	} else if app.selectedIndex == 1 {
		return app.handleLocationChoice(2, initLocationConfig)
	}
	return app, nil
}

// handleInitLocationChoice1 handles "1" key (init current directory)
func (a *Application) handleInitLocationChoice1(app *Application) (tea.Model, tea.Cmd) {
	return app.handleLocationChoice(1, initLocationConfig)
}

// handleInitLocationChoice2 handles "2" key (create subdirectory)
func (a *Application) handleInitLocationChoice2(app *Application) (tea.Model, tea.Cmd) {
	return app.handleLocationChoice(2, initLocationConfig)
}

// handleInputSubmitSubdirName handles enter when creating subdirectory for init
func (a *Application) handleInputSubmitSubdirName(app *Application) (tea.Model, tea.Cmd) {
	return app.validateAndProceed(ui.Validators["directory"], func(app *Application) (tea.Model, tea.Cmd) {
		cwd, err := os.Getwd()
		if err != nil {
			app.footerHint = "Failed to get current directory"
			return app, nil
		}

		app.initRepositoryPath = fmt.Sprintf("%s/%s", cwd, app.inputValue)

		// Transition to single branch input mode
		app.transitionTo(ModeTransition{
			Mode:        ModeInput,
			InputPrompt: "Initial branch name:",
			InputAction: "init_branch_name",
			FooterHint:  "Enter branch name (default: main), press Enter to initialize",
		})
		app.inputValue = "main"
		app.inputCursorPosition = len("main")

		return app, nil
	})
}



// handleInitBranchesCancel is now handled by global ESC handler
// Keeping as comment for reference of old behavior
// func (a *Application) handleInitBranchesCancel(app *Application) (tea.Model, tea.Cmd) { ... }

// executeInitWorkflow launches git operations in a worker and returns a command
func (a *Application) executeInitWorkflow(branchName string) tea.Cmd {
	// UI THREAD - Launching worker goroutine for git operations
	repoPath := a.initRepositoryPath

	// Set up async state for console display
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0
	a.mode = ModeClone
	a.consoleState = ui.NewConsoleOutState()
	a.outputBuffer.Clear()
	a.footerHint = "Initializing repository... (ESC to abort)"

	return func() tea.Msg {
		result := git.ExecuteWithStreaming("init", repoPath)
		if !result.Success {
			return GitOperationMsg{Step: "init", Success: false, Error: "git init failed", Path: repoPath}
		}

		// Change to repo directory for branch creation
		if err := os.Chdir(repoPath); err != nil {
			return GitOperationMsg{Step: "init", Success: false, Error: "Failed to change to repository directory", Path: repoPath}
		}

		// Create initial branch
		result = git.ExecuteWithStreaming("checkout", "-b", branchName)
		if !result.Success {
			return GitOperationMsg{Step: "init", Success: false, Error: fmt.Sprintf("Failed to create branch '%s'", branchName), Path: repoPath}
		}

		return GitOperationMsg{
			Step:    "init",
			Success: true,
			Output:  fmt.Sprintf("Repository initialized with branch '%s'", branchName),
			Path:    repoPath,
		}
	}
}

// Paste handler

// handleKeyPaste handles ctrl+v and cmd+v - fast paste from clipboard
// Inserts entire pasted text at cursor position atomically
// Does NOT validate - paste allows any text, validation happens on submit
func (a *Application) handleKeyPaste(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Handle paste in input modes only
	if a.isInputMode() {
		text, err := clipboard.ReadAll()
		if err == nil && len(text) > 0 {
			// Trim whitespace from pasted text
			text = strings.TrimSpace(text)

			// Clamp cursor position to valid range
			if app.inputCursorPosition < 0 {
				app.inputCursorPosition = 0
			}
			if app.inputCursorPosition > len(app.inputValue) {
				app.inputCursorPosition = len(app.inputValue)
			}

			// Insert pasted text at cursor position (atomically, not character by character)
			app.inputValue = app.inputValue[:app.inputCursorPosition] + text + app.inputValue[app.inputCursorPosition:]
			app.inputCursorPosition += len(text)
			
			// Update real-time validation if in clone URL mode
			if app.inputAction == "clone_url" {
				if app.inputValue == "" {
					app.inputValidationMsg = ""
				} else if ui.ValidateRemoteURL(app.inputValue) {
					app.inputValidationMsg = "" // Valid - no error message
				} else {
					app.inputValidationMsg = "Invalid URL format"
				}
			}
		}
	}

	return app, nil
}

// Console output handlers

// handleConsoleUp scrolls console up one line
func (a *Application) handleConsoleUp(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Scroll console output up
	app.consoleState.ScrollUp()
	return app, nil
}

// handleConsoleDown scrolls console down one line
func (a *Application) handleConsoleDown(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Scroll console output down
	app.consoleState.ScrollDown()
	return app, nil
}

// handleConsolePageUp scrolls console up one page
func (a *Application) handleConsolePageUp(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Scroll console up by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollUp()
	}
	return app, nil
}

// handleConsolePageDown scrolls console down one page
func (a *Application) handleConsolePageDown(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Scroll console down by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollDown()
	}
	return app, nil
}

// Clone workflow handlers

// handleCloneURLSubmit validates URL and auto-creates subdir with repo name
func (a *Application) handleCloneURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	return app.validateAndProceed(ui.Validators["url"], func(app *Application) (tea.Model, tea.Cmd) {
		app.cloneURL = app.inputValue
		
		// Extract repo name from URL and set clone path
		repoName := git.ExtractRepoName(app.cloneURL)
		cwd, err := os.Getwd()
		if err != nil {
			app.footerHint = "Failed to get current directory"
			return app, nil
		}
		app.clonePath = fmt.Sprintf("%s/%s", cwd, repoName)
		
		// Start clone operation immediately (no location choice menu)
		return app.startCloneOperation()
	})
}

// handleCloneLocationSelection handles enter on clone location menu
func (a *Application) handleCloneLocationSelection(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex == 0 {
		return app.handleLocationChoice(1, cloneLocationConfig)
	} else if app.selectedIndex == 1 {
		return app.handleLocationChoice(2, cloneLocationConfig)
	}
	return app, nil
}

// handleCloneLocationChoice1 handles "1" key (clone to current directory)
func (a *Application) handleCloneLocationChoice1(app *Application) (tea.Model, tea.Cmd) {
	return app.handleLocationChoice(1, cloneLocationConfig)
}

// handleCloneLocationChoice2 handles "2" key (create subdirectory)
func (a *Application) handleCloneLocationChoice2(app *Application) (tea.Model, tea.Cmd) {
	return app.handleLocationChoice(2, cloneLocationConfig)
}

// handleInputSubmitCloneSubdirName handles enter when creating subdirectory for clone
func (a *Application) handleInputSubmitCloneSubdirName(app *Application) (tea.Model, tea.Cmd) {
	return app.validateAndProceed(ui.Validators["directory"], func(app *Application) (tea.Model, tea.Cmd) {
		cwd, err := os.Getwd()
		if err != nil {
			app.footerHint = "Failed to get current directory"
			return app, nil
		}
		app.clonePath = fmt.Sprintf("%s/%s", cwd, app.inputValue)
		return app.startCloneOperation()
	})
}

// startCloneOperation sets up async state and executes clone
func (a *Application) startCloneOperation() (tea.Model, tea.Cmd) {
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0
	a.mode = ModeClone
	a.consoleState = ui.NewConsoleOutState()
	a.outputBuffer.Clear()
	a.footerHint = "Cloning repository... (ESC to abort)"

	return a, a.executeCloneWorkflow()
}

// executeCloneWorkflow launches git clone in a worker and returns a command
func (a *Application) executeCloneWorkflow() tea.Cmd {
	cloneURL := a.cloneURL
	clonePath := a.clonePath
	cloneToCwd := false

	cwd, _ := os.Getwd()
	if clonePath == "" || clonePath == cwd {
		cloneToCwd = true
	}

	return func() tea.Msg {
		effectivePath := cwd // Default to current working directory
		
		if cloneToCwd {
			// Clone to cwd: use git init + remote add + pull (works with hidden files)
			result := git.ExecuteWithStreaming("init")
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git init failed", Path: effectivePath}
			}

			// Create .gitignore for common garbage files BEFORE checkout
			if err := git.CreateDefaultGitignore(); err != nil {
				return GitOperationMsg{Step: "clone", Success: false, Error: fmt.Sprintf("Failed to create .gitignore: %v", err), Path: effectivePath}
			}

			result = git.ExecuteWithStreaming("remote", "add", "origin", cloneURL)
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git remote add failed", Path: effectivePath}
			}

			result = git.ExecuteWithStreaming("fetch", "--all", "--progress")
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git fetch failed", Path: effectivePath}
			}

			// Set remote HEAD to auto-detect default branch
			result = git.ExecuteWithStreaming("remote", "set-head", "origin", "-a")
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git remote set-head failed", Path: effectivePath}
			}

			// Get default branch from remote
			defaultBranch := git.GetRemoteDefaultBranch()
			if defaultBranch == "" {
				defaultBranch = "main"
			}

			// Checkout default branch (untracked files are now ignored)
			result = git.ExecuteWithStreaming("checkout", defaultBranch)
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git checkout failed", Path: effectivePath}
			}
		} else {
			// Clone to subdir: standard git clone
			effectivePath = clonePath
			result := git.ExecuteWithStreaming("clone", "--progress", cloneURL, clonePath)
			if !result.Success {
				return GitOperationMsg{
					Step:    "clone",
					Success: false,
					Error:   fmt.Sprintf("Clone failed with exit code %d", result.ExitCode),
					Path:    effectivePath,
				}
			}
		}

		return GitOperationMsg{
			Step:    "clone",
			Success: true,
			Path:    effectivePath,
		}
	}
}

// handleSelectBranchEnter handles selecting canon branch from list
func (a *Application) handleSelectBranchEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.cloneBranches) {
		return app, nil
	}

	selectedBranch := app.cloneBranches[app.selectedIndex]

	// Save canon branch to config
	cfg := git.RepoConfig{}
	cfg.Repo.Initialized = true
	cfg.Repo.RepositoryPath, _ = os.Getwd()
	cfg.Repo.CanonBranch = selectedBranch
	cfg.Repo.LastWorkingBranch = "dev" // Default working branch
	git.SaveRepoConfig(cfg)

	// Reload git state with new config
	if state, err := git.DetectState(); err == nil {
		app.gitState = state
	}

	app.footerHint = fmt.Sprintf("Canon branch set to: %s", selectedBranch)

	// Clean up clone state and return to menu
	app.cloneBranches = nil
	app.cloneURL = ""
	app.clonePath = ""
	app.mode = ModeMenu
	app.selectedIndex = 0
	menu := app.GenerateMenu()
	app.menuItems = menu

	return app, nil
}

// Commit workflow handlers

// handleCommitSubmit validates commit message and executes commit
func (a *Application) handleCommitSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate commit message
	message := app.inputValue
	if message == "" {
		app.footerHint = "Commit message cannot be empty"
		return app, nil
	}

	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState = ui.NewConsoleOutState()
	app.outputBuffer.Clear()
	app.footerHint = GetFooterMessageText(MessageCommit)

	// Execute commit asynchronously
	return app, app.executeCommitWorkflow(message)
}

// executeCommitWorkflow launches git commit in a worker and returns a command
func (a *Application) executeCommitWorkflow(message string) tea.Cmd {
	// UI THREAD - Capturing state before spawning worker
	commitMessage := message

	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		// Stage all changes first
		result := git.ExecuteWithStreaming("add", "-A")
		if !result.Success {
			return GitOperationMsg{
				Step:    "commit",
				Success: false,
				Error:   "Failed to stage changes",
			}
		}

		// Create commit
		result = git.ExecuteWithStreaming("commit", "-m", commitMessage)
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

// executePushWorkflow launches git push in a worker and returns a command
func (a *Application) executePushWorkflow() tea.Cmd {
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming("push")
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

// executePullMergeWorkflow launches git pull (merge) in a worker and returns a command
func (a *Application) executePullMergeWorkflow() tea.Cmd {
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming("pull")
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

// executePullRebaseWorkflow launches git pull --rebase in a worker and returns a command
func (a *Application) executePullRebaseWorkflow() tea.Cmd {
	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		result := git.ExecuteWithStreaming("pull", "--rebase")
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

// handleAddRemoteSubmit validates URL and executes add remote + fetch
func (a *Application) handleAddRemoteSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate remote URL
	url := app.inputValue
	if url == "" {
		app.footerHint = "Remote URL cannot be empty"
		return app, nil
	}

	// Validate URL format
	if !ui.ValidateRemoteURL(url) {
		app.footerHint = ui.GetRemoteURLError()
		return app, nil
	}

	// Check if remote already exists
	result := git.Execute("remote", "get-url", "origin")
	if result.Success {
		app.footerHint = "Remote 'origin' already exists"
		return app, nil
	}

	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState = ui.NewConsoleOutState()
	app.outputBuffer.Clear()
	app.footerHint = GetFooterMessageText(MessageAddRemote)

	// Execute add remote + fetch asynchronously
	return app, app.executeAddRemoteWorkflow(url)
}

// executeAddRemoteWorkflow launches git remote add + fetch in a worker and returns a command
func (a *Application) executeAddRemoteWorkflow(remoteURL string) tea.Cmd {
	// UI THREAD - Capturing URL before spawning worker
	url := remoteURL

	return func() tea.Msg {
		// WORKER THREAD - Never touch Application
		// Add remote
		result := git.ExecuteWithStreaming("remote", "add", "origin", url)
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

		return GitOperationMsg{
			Step:    "add_remote",
			Success: true,
			Output:  "Remote added and fetched successfully",
		}
	}
}
