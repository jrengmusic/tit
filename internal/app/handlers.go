package app

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// errorOrEmpty returns error string if err != nil, else empty string
// Used in message structures where Error field must not be nil
func errorOrEmpty(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

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

// convertGitFilesToUIFileInfo converts git.FileInfo to ui.FileInfo for state management
// Used when populating file lists from git operations
func convertGitFilesToUIFileInfo(gitFiles []git.FileInfo) []ui.FileInfo {
	converted := make([]ui.FileInfo, len(gitFiles))
	for i, gitFile := range gitFiles {
		converted[i] = ui.FileInfo{
			Path:   gitFile.Path,
			Status: gitFile.Status,
		}
	}
	return converted
}

// handleKeyCtrlC handles Ctrl+C globally
// During async operations: shows "operation in progress" message
// During critical operations (!isExitAllowed): blocks exit to prevent broken git state
// Otherwise: prompts for confirmation before quitting
func (a *Application) handleKeyCtrlC(app *Application) (tea.Model, tea.Cmd) {
	// Block exit during critical operations (e.g., pull merge with potential conflicts)
	if !a.isExitAllowed {
		a.footerHint = GetFooterMessageText(MessageExitBlocked)
		return a, nil
	}

	// If async operation is running, show "in progress" message
	if a.asyncOperationActive && !a.asyncOperationAborted {
		a.footerHint = GetFooterMessageText(MessageOperationInProgress)
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
// In console mode with async operation: blocked (prevents exiting before operation completes)
// In conflict resolver mode: delegates to handleConflictEsc (abort operation)
// During async operations: aborts the operation
// In input mode with text: confirms clear (3s timeout)
// Otherwise: returns to previous menu and restores state
func (a *Application) handleKeyESC(app *Application) (tea.Model, tea.Cmd) {
	// Conflict resolver mode: delegate to conflict-specific handler
	if a.mode == ModeConflictResolve {
		return a.handleConflictEsc(app)
	}

	// Block ESC in console mode while async operation is active
	// This prevents user from exiting before critical operations complete (e.g., git merge --abort)
	if (a.mode == ModeConsole || a.mode == ModeClone) && a.asyncOperationActive {
		a.footerHint = GetFooterMessageText(MessageOperationInProgress)
		return a, nil
	}

	// If async operation was aborted but completed: restore previous state
	if a.asyncOperationAborted {
		a.asyncOperationActive = false
		a.asyncOperationAborted = false
		a.mode = a.previousMode
		a.selectedIndex = a.previousMenuIndex
		a.consoleState.Reset()
		a.outputBuffer.Clear()
		a.footerHint = ""

		// Regenerate menu if returning to menu mode
		if a.mode == ModeMenu {
			menu := app.GenerateMenu()
			app.menuItems = menu
			if a.previousMenuIndex < len(menu) && len(menu) > 0 {
				app.footerHint = menu[a.previousMenuIndex].Hint
			}
			// Rebuild shortcuts for new menu
			app.rebuildMenuShortcuts()
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
	a.consoleState.Reset()
	a.outputBuffer.Clear()
	a.footerHint = ""
	a.inputValue = ""
	a.inputCursorPosition = 0
	a.inputValidationMsg = ""
	a.clearConfirmActive = false
	a.isExitAllowed = true // ALWAYS allow exit when in menu

	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 {
		if a.gitState != nil && a.gitState.Remote == git.HasRemote && a.gitState.Timeline == "" && a.gitState.CurrentHash == "" {
			a.footerHint = FooterHints["no_commits_yet"]
		} else {
			a.footerHint = menu[0].Hint
		}
	}

	// Rebuild shortcuts for new menu
	a.rebuildMenuShortcuts()

	return a, nil
}

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

		subdirPath := fmt.Sprintf("%s/%s", cwd, app.inputValue)

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

		// Set up console for streaming output
		buffer := ui.GetBuffer()
		buffer.Clear()
		buffer.Append(OutputMessages["initializing_repo"], ui.TypeStatus)

		app.mode = ModeConsole
		app.asyncOperationActive = true
		app.inputValue = ""

		// Use cmdInit to create repo with .gitignore
		return app, app.cmdInit("main")
	})
}

// handleInitBranchesCancel is now handled by global ESC handler
// Keeping as comment for reference of old behavior
// func (a *Application) handleInitBranchesCancel(app *Application) (tea.Model, tea.Cmd) { ... }

// handleInitBranchNameSubmit validates branch name and starts init operation
func (a *Application) handleInitBranchNameSubmit() (tea.Model, tea.Cmd) {
	branchName := strings.TrimSpace(a.inputValue)
	if branchName == "" {
		a.footerHint = ErrorMessages["branch_name_empty"]
		return a, nil
	}

	buffer := ui.GetBuffer()
	buffer.Clear()
	buffer.Append(OutputMessages["initializing_repo"], ui.TypeStatus)

	a.mode = ModeConsole
	a.asyncOperationActive = true
	a.inputValue = ""

	return a, a.cmdInit(branchName)
}

// Paste handler

// handleKeyPaste handles ctrl+v and cmd+v - fast paste from clipboard
// Inserts entire pasted text at cursor position atomically
// Does NOT validate - paste allows any text, validation happens on submit
func (a *Application) handleKeyPaste(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Handle paste in input modes only
	if a.isInputMode() {
		text, err := clipboard.ReadAll()
		if err != nil {
			// Clipboard read failed - silently ignore and continue
			// (user may have cancelled, or clipboard unavailable)
			return app, nil
		}
		if len(text) == 0 {
			return app, nil
		}

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

	return app, nil
}

// Console output handlers

// handleConsoleUp scrolls console up one line
func (a *Application) handleConsoleUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.asyncOperationActive {
		return app, nil
	}
	app.consoleState.ScrollUp()
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsoleDown scrolls console down one line
func (a *Application) handleConsoleDown(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.asyncOperationActive {
		return app, nil
	}
	app.consoleState.ScrollDown()
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsolePageUp scrolls console up one page
func (a *Application) handleConsolePageUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.asyncOperationActive {
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
	if app.asyncOperationActive {
		return app, nil
	}
	// UI THREAD - Scroll console down by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollDown()
	}
	app.consoleAutoScroll = false // Disable auto-scroll on manual scroll
	return app, nil
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
		app.cloneURL = app.inputValue

		// Route based on how we got here
		if app.inputAction == "clone_url" {
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
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
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

	return func() tea.Msg {
		effectivePath := cwd // Default to current working directory

		if cloneMode == "here" {
			// Clone here: init + remote add + fetch + checkout with tracking
			buffer := ui.GetBuffer()

			// Step 1: git init
			result := git.ExecuteWithStreaming("init")
			if !result.Success {
				return GitOperationMsg{Step: OpClone, Success: false, Error: "git init failed", Path: effectivePath}
			}

			// Step 2: git remote add origin <url>
			result = git.ExecuteWithStreaming("remote", "add", "origin", cloneURL)
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
			result = git.ExecuteWithStreaming("fetch", "--all", "--progress")
			if !result.Success {
				return GitOperationMsg{Step: OpClone, Success: false, Error: "git fetch failed", Path: effectivePath}
			}

			// Step 5: Create and checkout local branch tracking remote
			// This sets up upstream automatically: -t = --track (sets upstream to origin/<branch>)
			result = git.ExecuteWithStreaming("checkout", "-t", "origin/"+defaultBranch)
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
			result := git.ExecuteWithStreaming("clone", "--progress", cloneURL)
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
	app.asyncOperationActive = true

	return app, func() tea.Msg {
		result := git.ExecuteWithStreaming("checkout", selectedBranch)
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
	message := app.inputValue
	if message == "" {
		app.footerHint = ErrorMessages["commit_message_empty"]
		return app, nil
	}

	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputValue = ""

	// Execute commit asynchronously using operations pattern
	return app, app.cmdCommit(message)
}

// handleCommitPushSubmit validates commit message and executes commit+push
func (a *Application) handleCommitPushSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Validate commit message
	message := app.inputValue
	if message == "" {
		app.footerHint = ErrorMessages["commit_message_empty"]
		return app, nil
	}

	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputValue = ""

	// Execute commit+push asynchronously
	return app, app.cmdCommitPush(message)
}

// cmdCommitWorkflow launches git commit in a worker and returns a command
func (a *Application) cmdCommitWorkflow(message string) tea.Cmd {
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

// cmdPushWorkflow launches git push in a worker and returns a command
func (a *Application) cmdPushWorkflow() tea.Cmd {
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

// cmdPullMergeWorkflow launches git pull (merge) in a worker and returns a command
func (a *Application) cmdPullMergeWorkflow() tea.Cmd {
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

// cmdPullRebaseWorkflow launches git pull --rebase in a worker and returns a command
func (a *Application) cmdPullRebaseWorkflow() tea.Cmd {
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
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputValue = ""

	// Execute add remote + fetch asynchronously using operations pattern
	return app, app.cmdAddRemote(url)
}

// cmdAddRemoteWorkflow just does the add remote step
// Rest handled by three-step chain in githandlers.go (add_remote → fetch_remote → complete)
func (a *Application) cmdAddRemoteWorkflow(remoteURL string) tea.Cmd {
	url := remoteURL
	return func() tea.Msg {
		result := git.ExecuteWithStreaming("remote", "add", "origin", url)
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

// ========================================
// Confirmation Mode Handlers
// ========================================

// handleConfirmationLeft moves selection to Yes button
func (a *Application) handleConfirmationLeft(app *Application) (tea.Model, tea.Cmd) {
	if a.confirmationDialog != nil {
		a.confirmationDialog.SelectYes()
	}
	return a, nil
}

// handleConfirmationRight moves selection to No button
func (a *Application) handleConfirmationRight(app *Application) (tea.Model, tea.Cmd) {
	if a.confirmationDialog != nil {
		a.confirmationDialog.SelectNo()
	}
	return a, nil
}

// handleConfirmationYes selects Yes button
func (a *Application) handleConfirmationYes(app *Application) (tea.Model, tea.Cmd) {
	if a.confirmationDialog != nil {
		a.confirmationDialog.SelectYes()
	}
	return a, nil
}

// handleConfirmationNo selects No button
func (a *Application) handleConfirmationNo(app *Application) (tea.Model, tea.Cmd) {
	if a.confirmationDialog != nil {
		a.confirmationDialog.SelectNo()
	}
	return a, nil
}

// handleConfirmationEnter confirms the current selection
func (a *Application) handleConfirmationEnter(app *Application) (tea.Model, tea.Cmd) {
	if a.confirmationDialog != nil {
		confirmed := a.confirmationDialog.GetSelectedButton() == ui.ButtonYes
		return a.handleConfirmationResponse(confirmed)
	}
	return a, nil
}

// History Mode Handlers

// handleHistoryUp navigates up in history mode
func (a *Application) handleHistoryUp(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil {
		return app, nil
	}

	if app.historyState.PaneFocused { // List pane focused
		if app.historyState.SelectedIdx > 0 {
			app.historyState.SelectedIdx--
			// Reset details cursor when switching commits
			app.historyState.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		if app.historyState.DetailsLineCursor > 0 {
			app.historyState.DetailsLineCursor--
		}
	}
	return app, nil
}

// handleHistoryDown navigates down in history mode
func (a *Application) handleHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil {
		return app, nil
	}

	if app.historyState.PaneFocused { // List pane focused
		if app.historyState.SelectedIdx < len(app.historyState.Commits)-1 {
			app.historyState.SelectedIdx++
			// Reset details cursor when switching commits
			app.historyState.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		// Get total lines in selected commit's details
		if app.historyState.SelectedIdx >= 0 && app.historyState.SelectedIdx < len(app.historyState.Commits) {
			commit := app.historyState.Commits[app.historyState.SelectedIdx]

			// Build details lines (must match renderHistoryDetailsPane logic)
			var totalLines int
			totalLines += 2                                       // "Author:" and "Date:" lines
			totalLines += 1                                       // Empty line separator
			totalLines += strings.Count(commit.Subject, "\n") + 1 // Commit subject lines

			// Only increment if not at the last line
			if app.historyState.DetailsLineCursor < totalLines-1 {
				app.historyState.DetailsLineCursor++
			}
		}
	}
	return app, nil
}

// handleHistoryTab switches focus between panes in history mode
func (a *Application) handleHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil {
		return app, nil
	}

	app.historyState.PaneFocused = !app.historyState.PaneFocused
	return app, nil
}

// handleHistoryEsc returns to menu from history mode
func (a *Application) handleHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	return app.returnToMenu()
}

// handleHistoryEnter handles ENTER key in history mode (Phase 7: Time Travel)
func (a *Application) handleHistoryEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil || app.historyState.SelectedIdx < 0 {
		return app, nil
	}

	// Get selected commit
	commit := app.historyState.Commits[app.historyState.SelectedIdx]

	// Show time travel confirmation dialog
	app.mode = ModeConfirmation
	app.confirmType = "time_travel"
	app.confirmContext = map[string]string{
		"commit_hash":    commit.Hash,
		"commit_subject": commit.Subject,
	}

	// Create confirmation dialog using SSOT
	// Format: hash (first 7 chars) on first line, subject on second line
	shortHash := ui.ShortenHash(commit.Hash)

	// Extract only first line of commit message (subject)
	subject := commit.Subject
	if idx := strings.Index(subject, "\n"); idx >= 0 {
		subject = subject[:idx]
	}

	msg := ConfirmationMessages["time_travel"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: fmt.Sprintf(msg.Explanation, shortHash, subject),
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "time_travel",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)

	return app, nil
}

// updateFileHistoryDiff looks up and sets the current diff content in state
// Called whenever commit or file selection changes
// Direct cache lookup: hash:path:version → diff content
func (a *Application) updateFileHistoryDiff() {
	if a.fileHistoryState == nil {
		return
	}

	// Check bounds - need both commit and file selected
	if len(a.fileHistoryState.Commits) == 0 || a.fileHistoryState.SelectedCommitIdx >= len(a.fileHistoryState.Commits) {
		a.fileHistoryState.DiffContent = ""
		return
	}

	if len(a.fileHistoryState.Files) == 0 || a.fileHistoryState.SelectedFileIdx >= len(a.fileHistoryState.Files) {
		a.fileHistoryState.DiffContent = ""
		return
	}

	// Get selected commit and file
	commit := a.fileHistoryState.Commits[a.fileHistoryState.SelectedCommitIdx]
	file := a.fileHistoryState.Files[a.fileHistoryState.SelectedFileIdx]

	// Determine version based on current working tree state (same logic as old-tit)
	// If working tree is Dirty (has unstaged changes) → use "wip" diff (commit vs working tree)
	// Otherwise → use "parent" diff (commit vs parent commit)
	version := "parent"
	if a.gitState.WorkingTree == git.Dirty {
		version = "wip"
	}

	// Build cache key using SSOT (prevents hardcoded formats)
	cacheKey := DiffCacheKey(commit.Hash, file.Path, version)

	// Direct cache lookup (thread-safe)
	a.diffCacheMutex.Lock()
	diffContent, exists := a.fileHistoryDiffCache[cacheKey]
	a.diffCacheMutex.Unlock()

	if exists && diffContent != "" {
		a.fileHistoryState.DiffContent = diffContent
	} else {
		// Not cached yet (can happen if commit has >100 files)
		a.fileHistoryState.DiffContent = ""
	}
}

// handleFileHistoryUp navigates up in file(s) history mode
func (a *Application) handleFileHistoryUp(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	switch app.fileHistoryState.FocusedPane {
	case ui.PaneCommits:
		// Navigate up in commits list
		if app.fileHistoryState.SelectedCommitIdx > 0 {
			app.fileHistoryState.SelectedCommitIdx--
			// Reset file selection when switching commits
			app.fileHistoryState.SelectedFileIdx = 0
			// Update files for new commit
			if app.fileHistoryState.SelectedCommitIdx >= 0 && app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits) {
				commitHash := app.fileHistoryState.Commits[app.fileHistoryState.SelectedCommitIdx].Hash
				if gitFileList, exists := app.fileHistoryFilesCache[commitHash]; exists {
					app.fileHistoryState.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate up in files list
		if app.fileHistoryState.SelectedFileIdx > 0 {
			app.fileHistoryState.SelectedFileIdx--
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate up in diff pane (move cursor up)
		if app.fileHistoryState.DiffLineCursor > 0 {
			app.fileHistoryState.DiffLineCursor--
		}
	}
	return app, nil
}

// handleFileHistoryDown navigates down in file(s) history mode
func (a *Application) handleFileHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	switch app.fileHistoryState.FocusedPane {
	case ui.PaneCommits:
		// Navigate down in commits list
		if app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits)-1 {
			app.fileHistoryState.SelectedCommitIdx++
			// Reset file selection when switching commits
			app.fileHistoryState.SelectedFileIdx = 0
			// Update files for new commit
			if app.fileHistoryState.SelectedCommitIdx >= 0 && app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits) {
				commitHash := app.fileHistoryState.Commits[app.fileHistoryState.SelectedCommitIdx].Hash
				if gitFileList, exists := app.fileHistoryFilesCache[commitHash]; exists {
					app.fileHistoryState.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate down in files list
		if app.fileHistoryState.SelectedFileIdx < len(app.fileHistoryState.Files)-1 {
			app.fileHistoryState.SelectedFileIdx++
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate down in diff pane (move cursor down)
		app.fileHistoryState.DiffLineCursor++
	}
	return app, nil
}

// handleFileHistoryTab switches focus between panes in file(s) history mode
func (a *Application) handleFileHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	// Cycle through panes: Commits → Files → Diff → Commits
	app.fileHistoryState.FocusedPane = (app.fileHistoryState.FocusedPane + 1) % 3
	return app, nil
}

// handleFileHistoryCopy handles copy action in file(s) history mode
// Copies selected lines from diff pane to clipboard
func (a *Application) handleFileHistoryCopy(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil || app.fileHistoryState.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Get selected lines based on visual mode
	var linesToCopy []string
	if app.fileHistoryState.VisualModeActive {
		// Visual mode: copy selected range
		linesToCopy = ui.GetSelectedLinesFromDiff(app.fileHistoryState.DiffContent, app.fileHistoryState.VisualModeStart, app.fileHistoryState.DiffLineCursor)
		// Exit visual mode after copy
		app.fileHistoryState.VisualModeActive = false
	} else {
		// Normal mode: copy current line
		linesToCopy = ui.GetSelectedLinesFromDiff(app.fileHistoryState.DiffContent, app.fileHistoryState.DiffLineCursor, app.fileHistoryState.DiffLineCursor)
	}

	// Copy to clipboard if we have lines
	if len(linesToCopy) > 0 {
		textToCopy := strings.Join(linesToCopy, "\n")
		if err := clipboard.WriteAll(textToCopy); err == nil {
			app.footerHint = FooterHints["copy_success"]
		} else {
			app.footerHint = FooterHints["copy_failed"]
		}
	}

	return app, nil
}

// handleFileHistoryVisualMode handles visual mode toggle in file(s) history mode
// Toggles visual selection mode, starting from current cursor position
func (a *Application) handleFileHistoryVisualMode(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil || app.fileHistoryState.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Toggle visual mode
	if app.fileHistoryState.VisualModeActive {
		// Already in visual mode - exit
		app.fileHistoryState.VisualModeActive = false
		app.footerHint = ""
	} else {
		// Enter visual mode from current cursor
		app.fileHistoryState.VisualModeActive = true
		app.fileHistoryState.VisualModeStart = app.fileHistoryState.DiffLineCursor
		app.footerHint = FooterHints["visual_mode_active"]
	}

	return app, nil
}

// handleFileHistoryEsc handles ESC in file(s) history mode
// If in visual mode, exit visual mode. Otherwise, return to menu.
func (a *Application) handleFileHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState != nil && app.fileHistoryState.VisualModeActive {
		// Exit visual mode, stay in file history
		app.fileHistoryState.VisualModeActive = false
		app.footerHint = ""
		return app, nil
	}
	// Not in visual mode, return to menu
	return app.returnToMenu()
}
