package app

import (
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
	keyHandlers          map[AppMode]map[string]func(*Application) (tea.Model, tea.Cmd) // Cached key handlers

	// Input mode state
	inputPrompt          string // e.g., "Repository name:"
	inputValue           string
	inputCursorPosition  int // Cursor byte position in inputValue
	inputAction          string // Action being performed (e.g., "init_location", "canon_branch")

	// Initialization workflow state
	initRepositoryPath   string // Path to repository being initialized
	initCanonBranch      string // Canon branch name chosen during init
	initWorkingBranch    string // Working branch name chosen during init
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
		sizing:        sizing,
		theme:         theme,
		mode:          ModeMenu,
		gitState:      gitState,
		selectedIndex: 0,
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
			a.inputValue = a.inputValue[:a.inputCursorPosition] + keyStr + a.inputValue[a.inputCursorPosition:]
			a.inputCursorPosition++
			return a, nil
		}

		// Handle backspace in input modes
		if a.isInputMode() && keyStr == "backspace" {
			if a.inputCursorPosition > 0 {
				a.inputValue = a.inputValue[:a.inputCursorPosition-1] + a.inputValue[a.inputCursorPosition:]
				a.inputCursorPosition--
			}
			return a, nil
		}

	case TickMsg:
		if a.quitConfirmActive {
			a.quitConfirmActive = false
			a.footerHint = "" // Clear confirmation message
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
		contentText = "[Console mode - TODO]"
	case ModeInput:
		textInputState := ui.TextInputState{
			Value:      a.inputValue,
			CursorPos:  a.inputCursorPosition,
			Height:     1,
		}
		contentText = ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)
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
	case ModeInitializeCanonBranch:
		textInputState := ui.TextInputState{
			Value:      a.inputValue,
			CursorPos:  a.inputCursorPosition,
			Height:     1,
		}
		contentText = ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)
	case ModeInitializeWorkingBranch:
		textInputState := ui.TextInputState{
			Value:      a.inputValue,
			CursorPos:  a.inputCursorPosition,
			Height:     1,
		}
		contentText = ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
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
	
	return ui.RenderLayout(a.sizing, contentText, a.width, a.height, a.theme, a)
}

// Init initializes the application
func (a *Application) Init() tea.Cmd {
	return nil
}

// GetFooterHint returns the footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}

// isInputMode checks if current mode accepts text input
func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeInitializeCanonBranch ||
		a.mode == ModeInitializeWorkingBranch
}

// buildKeyHandlers builds the complete handler registry for all modes
// Global handlers take priority and are merged into each mode
func (a *Application) buildKeyHandlers() map[AppMode]map[string]func(*Application) (tea.Model, tea.Cmd) {
	// Global handlers - highest priority, applied to all modes
	globalHandlers := map[string]func(*Application) (tea.Model, tea.Cmd){
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
	}

	// Mode-specific handlers (global merged in after)
	modeHandlers := map[AppMode]map[string]func(*Application) (tea.Model, tea.Cmd){
		ModeMenu: {
			"up":    a.handleMenuUp,
			"k":     a.handleMenuUp,
			"down":  a.handleMenuDown,
			"j":     a.handleMenuDown,
			"enter": a.handleMenuEnter,
		},
		ModeConsole: {},
		ModeInput: {
			"enter": a.handleInputSubmit,
			"left":  a.handleInputLeft,
			"right": a.handleInputRight,
			"home":  a.handleInputHome,
			"end":   a.handleInputEnd,
		},
		ModeInitializeLocation: {
			"up":    a.handleMenuUp,
			"k":     a.handleMenuUp,
			"down":  a.handleMenuDown,
			"j":     a.handleMenuDown,
			"enter": a.handleInitLocationSelection,
			"1":     a.handleInitLocationChoice1,
			"2":     a.handleInitLocationChoice2,
		},
		ModeInitializeCanonBranch: {
			"enter": a.handleCanonBranchSubmit,
			"left":  a.handleInputLeft,
			"right": a.handleInputRight,
			"home":  a.handleInputHome,
			"end":   a.handleInputEnd,
		},
		ModeInitializeWorkingBranch: {
			"enter": a.handleWorkingBranchSubmit,
			"left":  a.handleInputLeft,
			"right": a.handleInputRight,
			"home":  a.handleInputHome,
			"end":   a.handleInputEnd,
		},
		ModeConfirmation: {},
		ModeHistory: {},
		ModeConflictResolve: {},
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
	// Generic input submit - dispatch based on inputAction
	return app, nil
}

// handleInputLeft moves cursor left
func (a *Application) handleInputLeft(app *Application) (tea.Model, tea.Cmd) {
	if app.inputCursorPosition > 0 {
		app.inputCursorPosition--
	}
	return app, nil
}

// handleInputRight moves cursor right
func (a *Application) handleInputRight(app *Application) (tea.Model, tea.Cmd) {
	if app.inputCursorPosition < len(app.inputValue) {
		app.inputCursorPosition++
	}
	return app, nil
}

// handleInputHome moves cursor to start
func (a *Application) handleInputHome(app *Application) (tea.Model, tea.Cmd) {
	app.inputCursorPosition = 0
	return app, nil
}

// handleInputEnd moves cursor to end
func (a *Application) handleInputEnd(app *Application) (tea.Model, tea.Cmd) {
	app.inputCursorPosition = len(app.inputValue)
	return app, nil
}

// Init workflow handlers

// handleInitLocationSelection handles enter on init location menu
func (a *Application) handleInitLocationSelection(app *Application) (tea.Model, tea.Cmd) {
	// TODO: Implement - dispatch based on selectedIndex
	return app, nil
}

// handleInitLocationChoice1 handles "1" key (init current directory)
func (a *Application) handleInitLocationChoice1(app *Application) (tea.Model, tea.Cmd) {
	// TODO: Implement - init current directory
	return app, nil
}

// handleInitLocationChoice2 handles "2" key (create subdirectory)
func (a *Application) handleInitLocationChoice2(app *Application) (tea.Model, tea.Cmd) {
	// TODO: Implement - ask for repo name
	return app, nil
}

// handleCanonBranchSubmit handles enter on canon branch input
func (a *Application) handleCanonBranchSubmit(app *Application) (tea.Model, tea.Cmd) {
	// TODO: Implement
	return app, nil
}

// handleWorkingBranchSubmit handles enter on working branch input
func (a *Application) handleWorkingBranchSubmit(app *Application) (tea.Model, tea.Cmd) {
	// TODO: Implement
	return app, nil
}
