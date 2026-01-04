package app

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
)

// handleKeyCtrlC handles Ctrl+C with confirmation
func (a *Application) handleKeyCtrlC(app *Application) (tea.Model, tea.Cmd) {
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

// handleInitBranchesCancel exits init workflow and returns to menu
func (a *Application) handleInitBranchesCancel(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Cancel init and return to menu
	app.mode = ModeMenu
	app.selectedIndex = 0
	app.initRepositoryPath = ""
	app.initCanonBranch = ""
	app.initWorkingBranch = ""
	app.inputCursorPosition = 0
	app.initActiveField = ""
	app.footerHint = ""

	// Regenerate menu
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 {
		app.footerHint = menu[0].Hint
	}

	return app, nil
}

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
