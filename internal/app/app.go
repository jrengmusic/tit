package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// Application represents the main TIT app state
type Application struct {
	width                int
	height               int
	sizing               ui.Sizing
	theme                ui.Theme
	mode                 AppMode // Current application mode
	quitConfirmActive    bool
	quitConfirmTime      time.Time
	footerHint           string // Footer hint/message text
	gitState             *git.State
	selectedIndex        int // Current menu item selection
	menuItems            []MenuItem // Cached menu items
	keyHandlers          map[AppMode]map[string]KeyHandler // Cached key handlers

	// Input mode state
	inputPrompt          string // e.g., "Repository name:"
	inputValue           string
	inputCursorPosition  int // Cursor byte position in inputValue
	inputAction          string // Action being performed (e.g., "init_location", "canon_branch")
	inputValidationMsg   string // Validation feedback message (empty = valid, shows message otherwise)
	clearConfirmActive   bool   // True when waiting for second ESC to clear input

	// Initialization workflow state
	initRepositoryPath   string // Path to repository being initialized
	initCanonBranch      string // Canon branch name chosen during init
	initWorkingBranch    string // Working branch name chosen during init
	initActiveField      string // "canon" or "working" - which field is active in ModeInitializeBranches

	// Clone workflow state
	cloneURL             string   // URL to clone from
	clonePath            string   // Path to clone into (cwd or subdir)
	cloneBranches        []string // Available branches after clone

	// Async operation state
	asyncOperationActive bool // True while git operation (clone, init, etc) is running
	asyncOperationAborted bool // True if user pressed ESC to abort during operation
	previousMode         AppMode // Mode before async operation started (for restoration on ESC)
	previousMenuIndex    int // Menu selection before async (for restoration)

	// Console output state (for clone, init, etc)
	consoleState         ui.ConsoleOutState
	outputBuffer         *ui.OutputBuffer
	
}

// ModeTransition configuration for streamlined mode changes
type ModeTransition struct {
    Mode            AppMode
    InputPrompt     string
    InputAction     string
    FooterHint      string
    ResetFields     []string // Field names to reset: "clone", "init", "all"
}

// transitionTo handles standardized mode transitions and state resets.
func (a *Application) transitionTo(config ModeTransition) {
    a.mode = config.Mode

    // Always reset common input state
    a.selectedIndex = 0
    a.inputValue = ""
    a.inputCursorPosition = 0
    a.inputValidationMsg = ""
	a.clearConfirmActive = false

    // Set new input config from the transition configuration
    if config.InputPrompt != "" {
        a.inputPrompt = config.InputPrompt
    }
    if config.InputAction != "" {
        a.inputAction = config.InputAction
    }
    if config.FooterHint != "" {
        a.footerHint = config.FooterHint
    }

    // Reset workflow-specific fields based on the configuration
    for _, field := range config.ResetFields {
        switch field {
        case "clone":
            a.cloneURL = ""
            a.clonePath = ""
            a.cloneBranches = nil
        case "init":
            a.initRepositoryPath = ""
            a.initCanonBranch = ""
            a.initWorkingBranch = ""
        case "all":
            // Reset all workflow states
            a.cloneURL = ""
            a.clonePath = ""
            a.cloneBranches = nil
            a.initRepositoryPath = ""
            a.initCanonBranch = ""
            a.initWorkingBranch = ""
        }
    }
}


// NewApplication creates a new application instance
func NewApplication(sizing ui.Sizing, theme ui.Theme) *Application {
	// Detect git state (nil if not in repo)
	var gitState *git.State
	if state, err := git.DetectState(); err == nil {
		gitState = state
	} else {
		// Not in a repo: create NotRepo state
		gitState = &git.State{
			Operation: git.NotRepo,
		}
	}

	app := &Application{
		sizing:               sizing,
		theme:                theme,
		mode:                 ModeMenu,
		gitState:             gitState,
		selectedIndex:        0,
		asyncOperationActive: false,
		asyncOperationAborted: false,
		consoleState:         ui.NewConsoleOutState(),
		outputBuffer:         ui.GetBuffer(),
	}

	// Build and cache key handler registry once
	app.keyHandlers = app.buildKeyHandlers()

	// Pre-generate menu and load initial hint
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 {
		app.footerHint = menu[0].Hint
	}

	return app
}

// Update handles all messages
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// Handle bracketed paste - entire paste comes as single KeyMsg with Paste=true
		if msg.Paste && a.isInputMode() {
			text := strings.TrimSpace(string(msg.Runes))
			if len(text) > 0 {
				a.inputValue = a.inputValue[:a.inputCursorPosition] + text + a.inputValue[a.inputCursorPosition:]
				a.inputCursorPosition += len(text)
				
				if a.inputAction == "clone_url" {
					if a.inputValue == "" {
						a.inputValidationMsg = ""
					} else if ui.ValidateRemoteURL(a.inputValue) {
						a.inputValidationMsg = ""
					} else {
						a.inputValidationMsg = "Invalid URL format"
					}
				}
			}
			return a, nil
		}
		keyStr := msg.String()
		
		// Look up handler in cached registry
		if modeHandlers, modeExists := a.keyHandlers[a.mode]; modeExists {
			if handler, exists := modeHandlers[keyStr]; exists {
				return handler(a)
			}
		}

		// Handle character input in input modes
		if a.isInputMode() && len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
			// Insert character at cursor position
			if a.mode == ModeInitializeBranches {
				// Insert into active field
				if a.initActiveField == "canon" {
					a.initCanonBranch = a.initCanonBranch[:a.inputCursorPosition] + keyStr + a.initCanonBranch[a.inputCursorPosition:]
				} else {
					a.initWorkingBranch = a.initWorkingBranch[:a.inputCursorPosition] + keyStr + a.initWorkingBranch[a.inputCursorPosition:]
				}
				a.inputCursorPosition++
			} else {
				// Generic input mode
				a.inputValue = a.inputValue[:a.inputCursorPosition] + keyStr + a.inputValue[a.inputCursorPosition:]
				a.inputCursorPosition++
				
				// Real-time validation for clone URL input
				if a.inputAction == "clone_url" {
					if a.inputValue == "" {
						a.inputValidationMsg = ""
					} else if ui.ValidateRemoteURL(a.inputValue) {
						a.inputValidationMsg = "" // Valid - no error message
					} else {
						a.inputValidationMsg = "Invalid URL format"
					}
				}
			}
			return a, nil
		}

		// Handle backspace in input modes
		if a.isInputMode() && keyStr == "backspace" {
			if a.mode == ModeInitializeBranches {
				// Delete from active field
				if a.inputCursorPosition > 0 {
					if a.initActiveField == "canon" {
						a.initCanonBranch = a.initCanonBranch[:a.inputCursorPosition-1] + a.initCanonBranch[a.inputCursorPosition:]
					} else {
						a.initWorkingBranch = a.initWorkingBranch[:a.inputCursorPosition-1] + a.initWorkingBranch[a.inputCursorPosition:]
					}
					a.inputCursorPosition--
				}
			} else {
				// Generic input mode
				if a.inputCursorPosition > 0 {
					a.inputValue = a.inputValue[:a.inputCursorPosition-1] + a.inputValue[a.inputCursorPosition:]
					a.inputCursorPosition--
					
					// Real-time validation for clone URL input
					if a.inputAction == "clone_url" {
						if a.inputValue == "" {
							a.inputValidationMsg = ""
						} else if ui.ValidateRemoteURL(a.inputValue) {
							a.inputValidationMsg = "" // Valid - no error message
						} else {
							a.inputValidationMsg = "Invalid URL format"
						}
					}
				}
			}
			return a, nil
		}

	case TickMsg:
		if a.quitConfirmActive {
			a.quitConfirmActive = false
			a.footerHint = "" // Clear confirmation message
		}

	case ClearTickMsg:
		if a.clearConfirmActive {
			a.clearConfirmActive = false
			a.footerHint = "" // Clear confirmation message
		}

	case GitOperationMsg:
		// AUDIO THREAD - Worker returned git operation result
		switch msg.Step {
		case "init":
			if msg.Success {
				// Reload git state after successful initialization
				if state, err := git.DetectState(); err == nil {
					a.gitState = state
				}
				
				// Reset init state and return to menu
				a.initRepositoryPath = ""
				a.initCanonBranch = ""
				a.initWorkingBranch = ""
				a.inputValue = ""
				a.inputCursorPosition = 0
				a.inputPrompt = ""
				a.inputAction = ""
				
				// Regenerate menu and switch to menu mode
				a.mode = ModeMenu
				a.selectedIndex = 0
				menu := a.GenerateMenu()
				a.menuItems = menu
				if len(menu) > 0 {
					a.footerHint = menu[0].Hint
				}
				a.footerHint = msg.Output // Show success message
			} else {
				// Show error message, stay in current mode
				a.footerHint = msg.Error
			}
		case "clone":
			if msg.Success {
				// Clone succeeded - STAY in console mode (contract: user dismisses with ESC)
				// DO NOT set asyncOperationActive = false yet

				// Change to cloned directory if subdir was created
				if a.clonePath != "" {
					cwd, _ := os.Getwd()
					if a.clonePath != cwd {
						os.Chdir(a.clonePath)
					}
				}

				// Reload git state after successful clone
				if state, err := git.DetectState(); err == nil {
					a.gitState = state
				}

				// Detect branches for canon selection
				branches, err := git.ListRemoteBranches()
				if err != nil || len(branches) == 0 {
					// Fallback to local branches
					branches, _ = git.ListBranches()
				}

				if len(branches) == 0 {
					// No branches found - set flag to allow ESC to close console
					a.asyncOperationActive = false
					a.footerHint = "Clone completed. No branches detected. Press ESC to continue."
				} else if len(branches) == 1 {
					// Single branch - auto-set as canon and save config
					canonBranch := branches[0]
					cfg := git.RepoConfig{}
					cfg.Repo.Initialized = true
					cfg.Repo.RepositoryPath, _ = os.Getwd()
					cfg.Repo.CanonBranch = canonBranch
					cfg.Repo.LastWorkingBranch = "dev" // Default working branch
					git.SaveRepoConfig(cfg)

					// Reload state with new config
					if state, err := git.DetectState(); err == nil {
						a.gitState = state
					}

					// Set flag to allow ESC to close console
					a.asyncOperationActive = false
					a.footerHint = fmt.Sprintf("Clone completed. Canon branch: %s. Press ESC to continue.", canonBranch)
				} else {
					// Multiple branches - need user selection
					// Keep asyncOperationActive = true to prevent ESC from closing console
					// Instead, transition to branch selection mode
					a.asyncOperationActive = false
					a.cloneBranches = branches
					a.mode = ModeSelectBranch
					a.selectedIndex = 0
					a.footerHint = "Select canon branch"
				}
			} else {
				// Clone failed - allow ESC to close console
				a.asyncOperationActive = false
				a.asyncOperationAborted = false
				a.footerHint = fmt.Sprintf("Clone failed: %s. Press ESC to return.", msg.Error)
			}
		}
	}

	return a, nil
}

// View renders the current view
func (a *Application) View() string {
	var contentText string
	
	// Render based on current mode
	switch a.mode {
	case ModeMenu:
		// Convert cached MenuItem to map for rendering
		menuMaps := make([]map[string]interface{}, len(a.menuItems))
		for i, item := range a.menuItems {
			menuMaps[i] = map[string]interface{}{
				"ID":        item.ID,
				"Shortcut":  item.Shortcut,
				"Emoji":     item.Emoji,
				"Label":     item.Label,
				"Hint":      item.Hint,
				"Enabled":   item.Enabled,
				"Separator": item.Separator,
			}
		}
		
		contentText = ui.RenderMenuWithHeight(menuMaps, a.selectedIndex, a.theme, ui.ContentHeight)
	
	case ModeConsole:
		// Rendering clone console output
		contentText = ui.RenderConsoleOutput(
			a.consoleState,
			a.outputBuffer,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight,
			a.asyncOperationActive && !a.asyncOperationAborted,
			a.asyncOperationAborted,
			a.asyncOperationActive && !a.asyncOperationAborted, // autoScroll while operation running
		)
	case ModeClone:
		// Clone operation in progress - show console
		contentText = ui.RenderConsoleOutput(
			a.consoleState,
			a.outputBuffer,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight,
			a.asyncOperationActive && !a.asyncOperationAborted,
			a.asyncOperationAborted,
			a.asyncOperationActive && !a.asyncOperationAborted, // autoScroll while operation running
		)
	case ModeSelectBranch:
		// Dynamic menu from cloneBranches
		items := make([]map[string]interface{}, len(a.cloneBranches))
		for i, branch := range a.cloneBranches {
			items[i] = map[string]interface{}{
				"ID":        branch,
				"Shortcut":  "",
				"Emoji":     "🌿",
				"Label":     branch,
				"Hint":      fmt.Sprintf("Set %s as canon branch", branch),
				"Enabled":   true,
				"Separator": false,
			}
		}
		contentText = ui.RenderMenuWithHeight(items, a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInput:
		textInputState := ui.TextInputState{
			Value:      a.inputValue,
			CursorPos:  a.inputCursorPosition,
			Height:     1,
		}
		
		// Render text input with optional validation message
		inputContent := ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)
		
		// Append validation message if present
		if a.inputValidationMsg != "" {
			inputContent += "\n\n" + a.inputValidationMsg
		}
		
		contentText = inputContent
	case ModeCloneURL:
		textInputState := ui.TextInputState{
			Value:     a.inputValue,
			CursorPos: a.inputCursorPosition,
			Height:    1,
		}
		
		inputContent := ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)
		
		if a.inputValidationMsg != "" {
			inputContent += "\n\n" + a.inputValidationMsg
		}
		
		contentText = inputContent
	case ModeCloneLocation:
		items := []map[string]interface{}{
			{
				"ID":        "clone_here",
				"Shortcut":  "1",
				"Emoji":     "📍",
				"Label":     "Clone to current directory",
				"Hint":      "Clone repository here",
				"Enabled":   true,
				"Separator": false,
			},
			{
				"ID":        "clone_subdir",
				"Shortcut":  "2",
				"Emoji":     "📁",
				"Label":     "Create subdirectory",
				"Hint":      "Create new folder and clone there",
				"Enabled":   true,
				"Separator": false,
			},
		}
		contentText = ui.RenderMenuWithHeight(items, a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInitializeLocation:
		// Show two options: initialize current directory or create subdirectory
		items := []map[string]interface{}{
			{
				"ID":        "init_here",
				"Shortcut":  "1",
				"Emoji":     "📍",
				"Label":     "Initialize current directory",
				"Hint":      "Create repository here",
				"Enabled":   true,
				"Separator": false,
			},
			{
				"ID":        "init_subdir",
				"Shortcut":  "2",
				"Emoji":     "📁",
				"Label":     "Create subdirectory",
				"Hint":      "Create new folder and initialize there",
				"Enabled":   true,
				"Separator": false,
			},
		}
		contentText = ui.RenderMenuWithHeight(items, a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInitializeBranches:
		// Both canon and working branch inputs displayed simultaneously
		// Pass correct cursor position based on active field
		canonCursorPos := 0
		workingCursorPos := 0
		if a.initActiveField == "canon" {
			canonCursorPos = a.inputCursorPosition
		} else {
			workingCursorPos = a.inputCursorPosition
		}

		contentText = ui.RenderBranchInputs(
			"Canon:",
			a.initCanonBranch,
			canonCursorPos,
			"Working:",
			a.initWorkingBranch,
			workingCursorPos,
			a.initActiveField, // "canon" or "working"
			a.theme,
		)
	case ModeConfirmation:
		contentText = "[Confirmation mode - TODO]"
	case ModeHistory:
		contentText = "[History mode - TODO]"
	case ModeConflictResolve:
		contentText = "[Conflict Resolve mode - TODO]"
	default:
		contentText = ""
	}
	
	// Get branch names from git state
	canonBranch := ""
	workingBranch := ""
	if a.gitState != nil {
		canonBranch = a.gitState.CanonBranch
		workingBranch = a.gitState.WorkingBranch
	}

	return ui.RenderLayout(a.sizing, contentText, a.width, a.height, a.theme, canonBranch, workingBranch, a)
}

// Init initializes the application
func (a *Application) Init() tea.Cmd {
	return tea.EnableBracketedPaste
}

// GetFooterHint returns the footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}

// isInputMode checks if current mode accepts text input
func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeInitializeBranches ||
		a.mode == ModeCloneURL
}

// buildKeyHandlers builds the complete handler registry for all modes
// Global handlers take priority and are merged into each mode
func (a *Application) buildKeyHandlers() map[AppMode]map[string]KeyHandler {
	// Global handlers - highest priority, applied to all modes
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"ctrl+v": a.handleKeyPaste,  // Linux/Windows/macOS
		"cmd+v":  a.handleKeyPaste,  // macOS cmd+v
		"meta+v": a.handleKeyPaste,  // macOS meta (cmd) - Bubble Tea may send this
		"alt+v":  a.handleKeyPaste,  // Fallback
	}

	cursorNavMixin := CursorNavigationMixin{}

	// Generic input cursor handlers for single-field inputs
	genericInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string { return a.inputValue },
		func(a *Application) int { return a.inputCursorPosition },
		func(a *Application, pos int) { a.inputCursorPosition = pos },
	)

	// Branch input cursor handlers for dual-field input
	branchInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string {
			if a.initActiveField == "canon" {
				return a.initCanonBranch
			}
			return a.initWorkingBranch
		},
		func(a *Application) int { return a.inputCursorPosition },
		func(a *Application, pos int) { a.inputCursorPosition = pos },
	)

	// Mode-specific handlers (global merged in after)
	modeHandlers := map[AppMode]map[string]KeyHandler{
		ModeMenu: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			Build(),
		ModeConsole: NewModeHandlers().Build(),
		ModeInput: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleInputSubmit).
			Build(),
		ModeInitializeLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleInitLocationSelection).
			On("1", a.handleInitLocationChoice1).
			On("2", a.handleInitLocationChoice2).
			Build(),
		ModeInitializeBranches: NewModeHandlers().
			WithCursorNav(branchInputNav).
			On("tab", a.handleInitBranchesTab).
			On("enter", a.handleInitBranchesSubmit).
			Build(),
		ModeCloneURL: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleCloneURLSubmit).
			Build(),
		ModeCloneLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleCloneLocationSelection).
			On("1", a.handleCloneLocationChoice1).
			On("2", a.handleCloneLocationChoice2).
			Build(),
		ModeConfirmation:    NewModeHandlers().Build(),
		ModeHistory:         NewModeHandlers().Build(),
		ModeConflictResolve: NewModeHandlers().Build(),
		ModeClone: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeSelectBranch: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleSelectBranchEnter).
			Build(),
	}

	// Merge global handlers into each mode (global takes priority)
	for mode := range modeHandlers {
		for key, handler := range globalHandlers {
			modeHandlers[mode][key] = handler
		}
	}

	return modeHandlers
}

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
	if len(app.menuItems) > 0 {
		app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
		// Skip separators
		for app.selectedIndex >= 0 && app.menuItems[app.selectedIndex].Separator {
			app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuDown moves selection down
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) {
	if len(app.menuItems) > 0 {
		app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
		// Skip separators
		for app.selectedIndex < len(app.menuItems) && app.menuItems[app.selectedIndex].Separator {
			app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuEnter selects current menu item and dispatches action
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.menuItems) {
		return app, nil
	}
	item := app.menuItems[app.selectedIndex]
	if item.Separator || !item.Enabled {
		return app, nil
	}
	
	// Dispatch action
	return app, app.dispatchAction(item.ID)
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	switch app.inputAction {
	case "init_subdir_name":
		return app.handleInputSubmitSubdirName(app)
	case "clone_subdir_name":
		return app.handleInputSubmitCloneSubdirName(app)
	default:
		return app, nil
	}
}


