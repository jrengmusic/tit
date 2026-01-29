package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Init workflow handlers

var initLocationConfig = LocationChoiceConfig{
	PathSetter: func(a *Application, path string) {}, // No-op, init always happens in CWD
	OnCurrentDir: func(a *Application) (tea.Model, tea.Cmd) {
		// Transition to single branch input mode
		a.transitionTo(ModeTransition{
			Mode:        ModeInput,
			InputPrompt: "Initial branch name:",
			InputAction: "init_branch_name",
			FooterHint:  "Enter branch name (default: main), press Enter to initialize",
		})
		a.inputState.Value = "main"
		a.inputState.CursorPosition = len("main")
		return a, nil
	},
	SubdirPrompt: "Repository name:",
	SubdirAction: "init_subdir_name",
}

var cloneLocationConfig = LocationChoiceConfig{
	PathSetter: func(a *Application, path string) { a.clonePath = path },
	OnCurrentDir: func(a *Application) (tea.Model, tea.Cmd) {
		// Clone here: clonePath is already set by PathSetter to cwd
		// Ask for URL, then init + remote add + fetch + checkout
		a.cloneMode = "here"
		return a.transitionToCloneURL("clone_here")
	},
	SubdirPrompt: "",
	SubdirAction: "clone_to_subdir",
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
			app.footerHint = ErrorMessages["cwd_read_failed"]
			return app, nil
		}

		subdirPath := fmt.Sprintf("%s/%s", cwd, app.inputState.Value)

		// Create subdirectory
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			app.footerHint = fmt.Sprintf(ErrorMessages["failed_create_dir"], err)
			return app, nil
		}

		// Change to subdirectory
		if err := os.Chdir(subdirPath); err != nil {
			app.footerHint = fmt.Sprintf(ErrorMessages["failed_change_dir"], err)
			return app, nil
		}

		// Ask for branch name before init
		app.transitionTo(ModeTransition{
			Mode:        ModeInput,
			InputPrompt: "Initial branch name:",
			InputAction: "init_branch_name",
			FooterHint:  "Enter branch name (default: main), press Enter to initialize",
		})
		app.inputState.Value = "main"
		app.inputState.CursorPosition = len("main")
		return app, nil
	})
}

// handleInitBranchNameSubmit validates branch name and starts init operation
func (a *Application) handleInitBranchNameSubmit() (tea.Model, tea.Cmd) {
	branchName := strings.TrimSpace(a.inputState.Value)
	if branchName == "" {
		a.footerHint = ErrorMessages["branch_name_empty"]
		return a, nil
	}

	buffer := ui.GetBuffer()
	buffer.Clear()
	buffer.Append(OutputMessages["initializing_repo"], ui.TypeStatus)

	a.mode = ModeConsole
	a.startAsyncOp()
	a.inputState.Value = ""

	return a, a.cmdInit(branchName)
}

// Clone workflow handlers

// transitionToCloneURL transitions to clone URL input with specified action
func (a *Application) transitionToCloneURL(action string) (tea.Model, tea.Cmd) {
	a.transitionTo(ModeTransition{
		Mode:        ModeCloneURL,
		InputPrompt: InputMessages["clone_url"].Prompt,
		InputAction: action,
		FooterHint:  InputMessages["clone_url"].Hint,
		ResetFields: []string{},
	})
	return a, nil
}

// handleCloneURLSubmit validates URL and routes based on input action
func (a *Application) handleCloneURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	return app.validateAndProceed(ui.Validators["url"], func(app *Application) (tea.Model, tea.Cmd) {
		app.cloneURL = app.inputState.Value

		// Route based on how we got here
		if app.inputState.Action == "clone_url" {
			// CWD not empty: start clone to subdir operation
			cwd, err := os.Getwd()
			if err != nil {
				app.footerHint = ErrorMessages["cwd_read_failed"]
				return app, nil
			}

			app.clonePath = cwd // git clone will create subdir automatically
			return app.startCloneOperation()
		}

		// CWD empty: either clone_here or clone_to_subdir
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

// handleSelectBranchEnter handles selecting canon branch from list
func (a *Application) handleSelectBranchEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.cloneBranches) {
		return app, nil
	}

	selectedBranch := app.cloneBranches[app.selectedIndex]

	// Checkout selected branch
	buffer := ui.GetBuffer()
	buffer.Clear()
	buffer.Append(fmt.Sprintf(OutputMessages["checking_out_branch"], selectedBranch), ui.TypeStatus)

	app.mode = ModeConsole
	app.startAsyncOp()

	ctx, cancel := context.WithCancel(context.Background())
	app.cancelContext = cancel

	return app, func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "checkout", selectedBranch)
		if !result.Success {
			return GitOperationMsg{
				Step:    "checkout",
				Success: false,
				Error:   fmt.Sprintf(ErrorMessages["failed_checkout_branch"], selectedBranch),
			}
		}

		return GitOperationMsg{
			Step:    "checkout",
			Success: true,
			Output:  fmt.Sprintf("Checked out branch '%s'", selectedBranch),
		}
	}
}

// Commit workflow handlers

// handleCommitSubmit validates commit message and executes commit
func (a *Application) handleCommitSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate commit message
	message := app.inputState.Value
	if message == "" {
		app.footerHint = ErrorMessages["commit_message_empty"]
		return app, nil
	}

	// Set up async state for console display
	app.startAsyncOp()
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	// Execute commit asynchronously using operations pattern
	return app, app.cmdCommit(message)
}

// handleCommitPushSubmit validates commit message and executes commit+push
func (a *Application) handleCommitPushSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate commit message
	message := app.inputState.Value
	if message == "" {
		app.footerHint = ErrorMessages["commit_message_empty"]
		return app, nil
	}

	// Set up async state for console display
	app.startAsyncOp()
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	// Execute commit+push asynchronously
	return app, app.cmdCommitPush(message)
}

// handleAddRemoteSubmit validates URL and executes add remote + fetch
func (a *Application) handleAddRemoteSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate remote URL
	url := app.inputState.Value
	if url == "" {
		app.footerHint = ErrorMessages["remote_url_empty_validation"]
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
		app.footerHint = ErrorMessages["remote_already_exists_validation"]
		return app, nil
	}

	// Set up async state for console display
	app.startAsyncOp()
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	// Execute add remote + fetch asynchronously using operations pattern
	return app, app.cmdAddRemote(url)
}
