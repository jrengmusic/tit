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

// handleInitLocationSelection handles enter on init location menu
func (a *Application) handleInitLocationSelection(app *Application) (tea.Model, tea.Cmd) {
	// Route to choice 1 or 2 based on selectedIndex
	if app.selectedIndex == 0 {
		return app.handleInitLocationChoice1(app)
	} else if app.selectedIndex == 1 {
		return app.handleInitLocationChoice2(app)
	}
	return app, nil
}

// handleInitLocationChoice1 handles "1" key (init current directory)
func (a *Application) handleInitLocationChoice1(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Setting up input mode for user interaction
	cwd, err := os.Getwd()
	if err != nil {
		app.footerHint = "Failed to get current directory"
		return app, nil
	}

	// Store the repository path
	app.initRepositoryPath = cwd

	// Transition to branch input mode (both canon + working)
	app.mode = ModeInitializeBranches
	app.initCanonBranch = "main"                           // Pre-fill canon with default
	app.initWorkingBranch = "dev"                          // Pre-fill working with default
	app.initActiveField = "working"                        // Start on working branch (canonical is pre-filled)
	app.inputCursorPosition = len("dev")                   // Cursor at end of working branch
	app.footerHint = "Tab to switch fields, Enter to submit, ESC to cancel"

	return app, nil
}

// handleInitLocationChoice2 handles "2" key (create subdirectory)
func (a *Application) handleInitLocationChoice2(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Switching to input mode to ask for directory name
	app.mode = ModeInput
	app.selectedIndex = 0
	app.inputValue = ""
	app.inputCursorPosition = 0
	app.inputPrompt = "Repository name:"
	app.inputAction = "init_subdir_name"
	app.footerHint = "Enter new directory name"

	return app, nil
}

// handleInputSubmitSubdirName handles enter when creating subdirectory
func (a *Application) handleInputSubmitSubdirName(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validating subdirectory name
	if app.inputValue == "" {
		app.footerHint = "Directory name cannot be empty"
		return app, nil
	}

	// Get current working directory and construct repo path
	cwd, err := os.Getwd()
	if err != nil {
		app.footerHint = "Failed to get current directory"
		return app, nil
	}

	app.initRepositoryPath = fmt.Sprintf("%s/%s", cwd, app.inputValue)

	// Transition to branch input mode (both canon + working)
	app.mode = ModeInitializeBranches
	app.initCanonBranch = "main"                    // Pre-fill canon with default
	app.initWorkingBranch = "dev"                   // Pre-fill working with default
	app.initActiveField = "working"                 // Start on working branch
	app.inputCursorPosition = len("dev")            // Cursor at end of working branch
	app.footerHint = "Tab to switch fields, Enter to submit, ESC to cancel"

	return app, nil
}

// handleInitBranchesTab cycles between canon and working branch fields
func (a *Application) handleInitBranchesTab(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Toggle active field
	if app.initActiveField == "working" {
		app.initActiveField = "canon"
		// Move cursor to end of canon branch name
		app.inputCursorPosition = len(app.initCanonBranch)
	} else {
		app.initActiveField = "working"
		// Move cursor to end of working branch name
		app.inputCursorPosition = len(app.initWorkingBranch)
	}
	return app, nil
}

// handleInitBranchesSubmit handles enter - only submit if on working field
func (a *Application) handleInitBranchesSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Enter only submits from working field
	if app.initActiveField != "working" {
		// On canon field, just move to working
		app.initActiveField = "working"
		app.inputCursorPosition = len(app.initWorkingBranch)
		return app, nil
	}

	// Validate working branch name
	if app.initWorkingBranch == "" {
		app.footerHint = "Working branch name cannot be empty"
		return app, nil
	}

	// Validate canon branch name
	if app.initCanonBranch == "" {
		app.footerHint = "Canon branch name cannot be empty"
		return app, nil
	}

	// Execute git operations asynchronously
	return app, app.executeInitWorkflow()
}

// handleInitBranchesLeft moves cursor left in active field
func (a *Application) handleInitBranchesLeft(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Move cursor left in active field
	if app.initActiveField == "canon" {
		if app.inputCursorPosition > 0 {
			app.inputCursorPosition--
		}
	} else if app.initActiveField == "working" {
		if app.inputCursorPosition > 0 {
			app.inputCursorPosition--
		}
	}
	return app, nil
}

// handleInitBranchesRight moves cursor right in active field
func (a *Application) handleInitBranchesRight(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Move cursor right in active field
	if app.initActiveField == "canon" {
		if app.inputCursorPosition < len(app.initCanonBranch) {
			app.inputCursorPosition++
		}
	} else if app.initActiveField == "working" {
		if app.inputCursorPosition < len(app.initWorkingBranch) {
			app.inputCursorPosition++
		}
	}
	return app, nil
}

// handleInitBranchesHome moves cursor to start of active field
func (a *Application) handleInitBranchesHome(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Move cursor to start in active field
	app.inputCursorPosition = 0
	return app, nil
}

// handleInitBranchesEnd moves cursor to end of active field
func (a *Application) handleInitBranchesEnd(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Move cursor to end in active field
	if app.initActiveField == "canon" {
		app.inputCursorPosition = len(app.initCanonBranch)
	} else if app.initActiveField == "working" {
		app.inputCursorPosition = len(app.initWorkingBranch)
	}
	return app, nil
}

// handleInitBranchesCancel is now handled by global ESC handler
// Keeping as comment for reference of old behavior
// func (a *Application) handleInitBranchesCancel(app *Application) (tea.Model, tea.Cmd) { ... }

// executeInitWorkflow launches git operations in a worker and returns a command
func (a *Application) executeInitWorkflow() tea.Cmd {
	// UI THREAD - Launching worker goroutine for git operations
	return func() tea.Msg {
		// WORKER THREAD - Execute all git operations
		if err := git.InitializeRepository(a.initRepositoryPath); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to initialize repository: %v", err),
			}
		}

		// Change to repository directory for branch operations
		originalCwd, _ := os.Getwd()
		if err := os.Chdir(a.initRepositoryPath); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to change directory: %v", err),
			}
		}
		defer os.Chdir(originalCwd)

		// Create canon branch
		if err := git.CreateBranch(a.initCanonBranch); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to create canon branch: %v", err),
			}
		}

		// Create working branch
		if err := git.CreateBranch(a.initWorkingBranch); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to create working branch: %v", err),
			}
		}

		// Save repository config
		cfg := git.RepoConfig{}
		cfg.Repo.Initialized = true
		cfg.Repo.RepositoryPath = a.initRepositoryPath
		cfg.Repo.CanonBranch = a.initCanonBranch
		cfg.Repo.LastWorkingBranch = a.initWorkingBranch

		if err := git.SaveRepoConfig(cfg); err != nil {
			return GitOperationMsg{
				Step:    "init",
				Success: false,
				Error:   fmt.Sprintf("Failed to save repository config: %v", err),
			}
		}

		return GitOperationMsg{
			Step:    "init",
			Success: true,
			Output:  fmt.Sprintf("Repository initialized: %s (canon: %s, working: %s)", a.initRepositoryPath, a.initCanonBranch, a.initWorkingBranch),
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

// handleCloneURLSubmit validates URL and moves to location choice
func (a *Application) handleCloneURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	if app.inputValue == "" {
		app.footerHint = "Repository URL cannot be empty"
		return app, nil
	}

	if !ui.ValidateRemoteURL(app.inputValue) {
		app.footerHint = "Invalid URL format. Try: git@github.com:user/repo.git or https://github.com/user/repo.git"
		return app, nil
	}

	// Store URL and move to location choice
	app.cloneURL = app.inputValue
	app.mode = ModeCloneLocation
	app.selectedIndex = 0
	app.inputValue = ""
	app.inputCursorPosition = 0
	app.inputValidationMsg = ""
	app.footerHint = "Choose where to clone the repository"
	return app, nil
}

// handleCloneLocationSelection handles enter on clone location menu
func (a *Application) handleCloneLocationSelection(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex == 0 {
		return app.handleCloneLocationChoice1(app)
	} else if app.selectedIndex == 1 {
		return app.handleCloneLocationChoice2(app)
	}
	return app, nil
}

// handleCloneLocationChoice1 handles "1" key (clone to current directory)
func (a *Application) handleCloneLocationChoice1(app *Application) (tea.Model, tea.Cmd) {
	cwd, err := os.Getwd()
	if err != nil {
		app.footerHint = "Failed to get current directory"
		return app, nil
	}

	app.clonePath = cwd
	return app.startCloneOperation()
}

// handleCloneLocationChoice2 handles "2" key (create subdirectory)
func (a *Application) handleCloneLocationChoice2(app *Application) (tea.Model, tea.Cmd) {
	app.mode = ModeInput
	app.inputValue = ""
	app.inputCursorPosition = 0
	app.inputPrompt = "Directory name:"
	app.inputAction = "clone_subdir_name"
	app.footerHint = "Enter new directory name"
	return app, nil
}

// handleInputSubmitCloneSubdirName handles enter when creating subdirectory for clone
func (a *Application) handleInputSubmitCloneSubdirName(app *Application) (tea.Model, tea.Cmd) {
	if app.inputValue == "" {
		app.footerHint = "Directory name cannot be empty"
		return app, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		app.footerHint = "Failed to get current directory"
		return app, nil
	}

	app.clonePath = fmt.Sprintf("%s/%s", cwd, app.inputValue)
	return app.startCloneOperation()
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
		if cloneToCwd {
			// Clone to cwd: use git init + remote add + pull (works with hidden files)
			result := git.ExecuteWithStreaming("init")
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git init failed"}
			}

			result = git.ExecuteWithStreaming("remote", "add", "origin", cloneURL)
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git remote add failed"}
			}

			result = git.ExecuteWithStreaming("fetch", "--all", "--progress")
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git fetch failed"}
			}

			// Get default branch from remote
			defaultBranch := git.GetRemoteDefaultBranch()
			if defaultBranch == "" {
				defaultBranch = "main"
			}

			result = git.ExecuteWithStreaming("checkout", defaultBranch)
			if !result.Success {
				return GitOperationMsg{Step: "clone", Success: false, Error: "git checkout failed"}
			}
		} else {
			// Clone to subdir: standard git clone
			result := git.ExecuteWithStreaming("clone", "--progress", cloneURL, clonePath)
			if !result.Success {
				return GitOperationMsg{
					Step:    "clone",
					Success: false,
					Error:   fmt.Sprintf("Clone failed with exit code %d", result.ExitCode),
				}
			}
		}

		return GitOperationMsg{
			Step:    "clone",
			Success: true,
		}
	}
}

// handleSelectBranchEnter handles selecting canon branch from list
func (a *Application) handleSelectBranchEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.cloneBranches) {
		return app, nil
	}
	
	selectedBranch := app.cloneBranches[app.selectedIndex]
	
	// TODO: Save canon branch to config
	app.footerHint = fmt.Sprintf("Canon branch set to: %s", selectedBranch)
	
	// Return to menu
	app.cloneBranches = nil
	app.cloneURL = ""
	app.clonePath = ""
	app.mode = ModeMenu
	app.selectedIndex = 0
	menu := app.GenerateMenu()
	app.menuItems = menu
	
	return app, nil
}
