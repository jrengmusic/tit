package app

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	inputHeight          int // Input height: 1 for single-line, 16 for multiline commit message
	inputAction          string // Action being performed (e.g., "init_location", "canon_branch")
	inputValidationMsg   string // Validation feedback message (empty = valid, shows message otherwise)
	clearConfirmActive   bool   // True when waiting for second ESC to clear input

	// Clone workflow state
	cloneURL             string   // URL to clone from
	clonePath            string   // Path to clone into (cwd or subdir)
	cloneMode            string   // "here" (init+remote+fetch) or "subdir" (git clone)
	cloneBranches        []string // Available branches after clone

	// Remote operation state
	remoteBranchName     string   // Current branch name during remote operations

	// Async operation state
	asyncOperationActive bool // True while git operation (clone, init, etc) is running
	asyncOperationAborted bool // True if user pressed ESC to abort during operation
	isExitAllowed        bool // False during critical operations (pull merge) to prevent premature exit
	previousMode         AppMode // Mode before async operation started (for restoration on ESC)
	previousMenuIndex    int // Menu selection before async (for restoration)

	// Console output state (for clone, init, etc)
	consoleState         ui.ConsoleOutState
	outputBuffer         *ui.OutputBuffer
	consoleAutoScroll    bool // Auto-scroll console to bottom (like old-tit)
	
	// Confirmation dialog state
	confirmationDialog   *ui.ConfirmationDialog
	confirmType          string // Type of confirmation for old-tit compatibility
	confirmContext       map[string]string // Context for old-tit compatibility
	
	// Conflict resolution state
	conflictResolveState *ConflictResolveState
	
	// Dirty operation tracking
	dirtyOperationState *DirtyOperationState // nil when no dirty op in progress
	
	// State display info maps
	workingTreeInfo      map[git.WorkingTree]StateInfo
	timelineInfo         map[git.Timeline]StateInfo
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
        case "all":
            // Reset all workflow states
            a.cloneURL = ""
            a.clonePath = ""
            a.cloneBranches = nil
        }
    }
}


// NewApplication creates a new application instance
func NewApplication(sizing ui.Sizing, theme ui.Theme) *Application {
	// Try to find and cd into git repository
	isRepo, repoPath := git.IsInitializedRepo()
	if !isRepo {
		// Check parent directories
		isRepo, repoPath = git.HasParentRepo()
	}
	
	var gitState *git.State
	if isRepo && repoPath != "" {
		// Found a repo, cd into it and detect state
		if err := os.Chdir(repoPath); err != nil {
			// Can't cd into repo - this is a fatal error
			panic(fmt.Sprintf("Cannot cd into repository at %s: %v", repoPath, err))
		}
		state, err := git.DetectState()
		if err != nil {
			// In a repo but state detection failed - this should not happen
			panic(fmt.Sprintf("Failed to detect git state in repo %s: %v", repoPath, err))
		}
		gitState = state
	} else {
		// Not in a repo - use NotRepo operation state to show init/clone menu
		gitState = &git.State{
			Operation: git.NotRepo,
		}
	}

	// Build state info maps
	workingTreeInfo, timelineInfo := BuildStateInfo(theme)

	app := &Application{
		sizing:               sizing,
		theme:                theme,
		mode:                 ModeMenu,
		gitState:             gitState,
		selectedIndex:        0,
		asyncOperationActive: false,
		asyncOperationAborted: false,
		isExitAllowed:        true, // Allow exit by default (disabled during critical operations)
		consoleState:         ui.NewConsoleOutState(),
		outputBuffer:         ui.GetBuffer(),
		consoleAutoScroll:    true, // Start with auto-scroll enabled
		workingTreeInfo:      workingTreeInfo,
		timelineInfo:         timelineInfo,
	}

	// Build and cache key handler registry once
	app.keyHandlers = app.buildKeyHandlers()

	// Pre-generate menu and load initial hint
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 {
		app.footerHint = menu[0].Hint
	}

	// Register menu shortcuts dynamically
	app.rebuildMenuShortcuts()

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
			text := string(msg.Runes) // Don't trim - preserve formatting
			if len(text) > 0 {
				a.insertTextAtCursor(text)
				a.updateInputValidation()
			}
			return a, nil
		}
		keyStr := msg.String()
		
		// Handle ctrl+j (shift+enter equivalent) for newline in multiline input
		if a.isInputMode() && (keyStr == "ctrl+j" || keyStr == "shift+enter") {
			a.insertTextAtCursor("\n")
			return a, nil
		}
		
		// Look up handler in cached registry
		if modeHandlers, modeExists := a.keyHandlers[a.mode]; modeExists {
			if handler, exists := modeHandlers[keyStr]; exists {
				return handler(a)
			}
		}

		// Handle character input in input modes
		if a.isInputMode() && len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
			a.insertTextAtCursor(keyStr)
			a.updateInputValidation()
			return a, nil
		}

		// Handle backspace in input modes
		if a.isInputMode() && keyStr == "backspace" {
			a.deleteAtCursor()
			a.updateInputValidation()
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
		return a.handleGitOperation(msg)
	}

	return a, nil
}

// View renders the current view
func (a *Application) View() string {
	var contentText string
	
	// Render based on current mode
	switch a.mode {
	case ModeMenu:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme, ui.ContentHeight)
	
	case ModeConsole, ModeClone:
		// Console output (both during and after operation)
		contentText = ui.RenderConsoleOutput(
			&a.consoleState,
			a.outputBuffer,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight,
			a.asyncOperationActive && !a.asyncOperationAborted,
			a.asyncOperationAborted,
			a.consoleAutoScroll,
		)
	
	case ModeConfirmation:
		// Confirmation dialog (centered in content area)
		if a.confirmationDialog != nil {
			contentText = a.confirmationDialog.Render()
		} else {
			// Fallback if no dialog - return to menu
			a.mode = ModeMenu
			contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme, ui.ContentHeight)
		}
	
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
			Height:     a.inputHeight, // Use configured height
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
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuCloneLocation()), a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInitializeLocation:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuInitializeLocation()), a.selectedIndex, a.theme, ui.ContentHeight)

	case ModeHistory:
		panic("ModeHistory: not yet implemented")
	case ModeConflictResolve:
		// Render conflict resolution UI using generic N-column view
		if a.conflictResolveState == nil {
			contentText = "No conflict state initialized"
		} else {
			contentText = ui.RenderConflictResolveGeneric(
				a.conflictResolveState.Files,
				a.conflictResolveState.SelectedFileIndex,
				a.conflictResolveState.FocusedPane,
				a.conflictResolveState.NumColumns,
				a.conflictResolveState.ColumnLabels,
				a.conflictResolveState.ScrollOffsets,
				a.conflictResolveState.LineCursors,
				ui.ContentInnerWidth,
				ui.ContentHeight,
				a.theme,
			)
		}
	default:
		panic(fmt.Sprintf("Unknown app mode: %v", a.mode))
	}
	
	// Get current branch from git state
	currentBranch := ""
	if a.gitState != nil {
		currentBranch = a.gitState.CurrentBranch
	}

	// Get current working directory
	cwd, _ := os.Getwd()

	return ui.RenderLayout(a.sizing, contentText, a.width, a.height, a.theme, currentBranch, cwd, a.gitState, a)
}

// Init initializes the application
func (a *Application) Init() tea.Cmd {
	return tea.EnableBracketedPaste
}

// GetFooterHint returns the footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}

// updateFooterHintFromMenu updates footer with hint of currently selected menu item
func (a *Application) updateFooterHintFromMenu() {
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.menuItems) {
		if !a.menuItems[a.selectedIndex].Separator {
			a.footerHint = a.menuItems[a.selectedIndex].Hint
		}
	}
}

// GetGitState returns the current git state
func (a *Application) GetGitState() interface{} {
	return a.gitState
}

// RenderStateHeader renders the full git state header (6 rows) using lipgloss
func (a *Application) RenderStateHeader() string {
	cwd, _ := os.Getwd()
	state := a.gitState
	if state == nil || state.Operation == git.NotRepo {
		// Don't render state header if not in a repo
		return ""
	}

	// Guard: Skip rendering if WorkingTree/Timeline are empty (happens during dirty operations)
	// DetectState() returns partial state for dirty operations (only Operation is set)
	if state.WorkingTree == "" || state.Timeline == "" {
		return ""
	}

	// Row 1: CWD with emoji, no truncation, full width
	cwdRow := lipgloss.NewStyle().
		Width(ui.ContentInnerWidth).
		Padding(0, 0).
		Bold(true).
		Foreground(lipgloss.Color(a.theme.LabelTextColor)).
		Render("📁 " + cwd)

	// Row 2: Remote URL
	remoteLabel := "🔌 NO REMOTE"
	remoteColor := a.theme.DimmedTextColor
	if state.Remote == git.HasRemote {
		url := git.GetRemoteURL()
		if url != "" {
			remoteLabel = "🔗 " + url
			remoteColor = a.theme.AccentTextColor
		}
	}
	remoteRow := lipgloss.NewStyle().
		Width(ui.ContentInnerWidth).
		Padding(0, 0).
		Foreground(lipgloss.Color(remoteColor)).
		Render("╚ " + remoteLabel)

	// Row 3: Blank spacer
	spacerRow := ""

	// Row 4: Working tree and timeline status
	wtInfo := a.workingTreeInfo[state.WorkingTree]
	tlInfo := a.timelineInfo[state.Timeline]

	wtLabel := wtInfo.Emoji + " " + wtInfo.Label
	tlLabel := tlInfo.Emoji + " " + tlInfo.Label

	statusRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(ui.ContentInnerWidth/2).
			Bold(true).
			Foreground(lipgloss.Color(wtInfo.Color)).
			Render(wtLabel),
		lipgloss.NewStyle().
			Width(ui.ContentInnerWidth/2).
			Bold(true).
			Foreground(lipgloss.Color(tlInfo.Color)).
			Render(tlLabel),
	)

	// Row 5: Descriptions
	wtDesc := wtInfo.Description(state.CommitsAhead, state.CommitsBehind)
	tlDesc := tlInfo.Description(state.CommitsAhead, state.CommitsBehind)

	descRow := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(ui.ContentInnerWidth/2).
			Render(wtDesc),
		lipgloss.NewStyle().
			Width(ui.ContentInnerWidth/2).
			Render(tlDesc),
	)

	// Row 6: Blank
	blankRow := ""

	// Combine all rows
	headerContent := cwdRow + "\n" + remoteRow + "\n" + spacerRow + "\n" + statusRow + "\n" + descRow + "\n" + blankRow

	return ui.RenderBox(ui.BoxConfig{
		Content:     headerContent,
		InnerWidth:  ui.ContentInnerWidth,
		InnerHeight: ui.HeaderHeight,
		BorderColor: a.theme.BoxBorderColor,
		TextColor:   a.theme.LabelTextColor,
		Theme:       a.theme,
	})
}

// isInputMode checks if current mode accepts text input
func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeCloneURL
}

// menuItemsToMaps converts MenuItem slice to map slice for rendering
// Note: Hint is excluded from maps (displayed in footer instead)
func (a *Application) menuItemsToMaps(items []MenuItem) []map[string]interface{} {
	maps := make([]map[string]interface{}, len(items))
	for i, item := range items {
		maps[i] = map[string]interface{}{
			"ID":        item.ID,
			"Shortcut":  item.Shortcut,
			"Emoji":     item.Emoji,
			"Label":     item.Label,
			"Enabled":   item.Enabled,
			"Separator": item.Separator,
		}
	}
	return maps
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

	// Mode-specific handlers (global merged in after)
	modeHandlers := map[AppMode]map[string]KeyHandler{
		ModeMenu: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			Build(),
		ModeConsole: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
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
		ModeConfirmation: NewModeHandlers().
			On("left", a.handleConfirmationLeft).
			On("right", a.handleConfirmationRight).
			On("h", a.handleConfirmationLeft).
			On("l", a.handleConfirmationRight).
			On("y", a.handleConfirmationYes).
			On("n", a.handleConfirmationNo).
			On("enter", a.handleConfirmationEnter).
			Build(),
		ModeHistory:         NewModeHandlers().Build(),
		ModeConflictResolve: NewModeHandlers().
			On("up", a.handleConflictUp).
			On("down", a.handleConflictDown).
			On("tab", a.handleConflictTab).
			On(" ", a.handleConflictSpace).  // Space character, not "space"
			On("enter", a.handleConflictEnter).
			Build(),
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

// rebuildMenuShortcuts dynamically registers keyboard handlers for all current menu item shortcuts
// Called after GenerateMenu() to ensure shortcuts match current git state
func (a *Application) rebuildMenuShortcuts() {
	if a.keyHandlers[ModeMenu] == nil {
		a.keyHandlers[ModeMenu] = make(map[string]KeyHandler)
	}

	// Remove old shortcut handlers (keep navigation and enter)
	// We'll rebuild from scratch by first copying the base handlers
	baseHandlers := NewModeHandlers().
		WithMenuNav(a).
		On("enter", a.handleMenuEnter).
		Build()

	// Merge global handlers
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	// Start fresh
	newHandlers := make(map[string]KeyHandler)

	// Copy base handlers
	for key, handler := range baseHandlers {
		newHandlers[key] = handler
	}

	// Add global handlers
	for key, handler := range globalHandlers {
		newHandlers[key] = handler
	}

	// Dynamically register shortcuts for current menu items
	for i, item := range a.menuItems {
		if item.Shortcut != "" && item.Enabled && !item.Separator {
			// Capture loop variables in closure
			itemIndex := i
			itemID := item.ID
			itemHint := item.Hint

			// Create handler that selects item and dispatches action
			newHandlers[item.Shortcut] = func(app *Application) (tea.Model, tea.Cmd) {
				app.selectedIndex = itemIndex
				app.footerHint = itemHint
				return app, app.dispatchAction(itemID)
			}
		}
	}

	// Replace ModeMenu handlers
	a.keyHandlers[ModeMenu] = newHandlers
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

// Input mode helpers

// insertTextAtCursor inserts text at current cursor position (UTF-8 safe)
func (a *Application) insertTextAtCursor(text string) {
	// Defensive bounds checking
	valueLen := len(a.inputValue)
	if a.inputCursorPosition < 0 {
		a.inputCursorPosition = 0
	}
	if a.inputCursorPosition > valueLen {
		a.inputCursorPosition = valueLen
	}
	
	// Safe slice operation
	before := a.inputValue[:a.inputCursorPosition]
	after := a.inputValue[a.inputCursorPosition:]
	a.inputValue = before + text + after
	a.inputCursorPosition += len(text)
}

// deleteAtCursor deletes character before cursor (UTF-8 safe)
func (a *Application) deleteAtCursor() {
	valueLen := len(a.inputValue)
	if a.inputCursorPosition <= 0 || valueLen == 0 {
		return
	}
	if a.inputCursorPosition > valueLen {
		a.inputCursorPosition = valueLen
	}
	
	// Safe slice operation
	before := a.inputValue[:a.inputCursorPosition-1]
	after := a.inputValue[a.inputCursorPosition:]
	a.inputValue = before + after
	a.inputCursorPosition--
}

// updateInputValidation updates validation message for current input
func (a *Application) updateInputValidation() {
	if a.inputAction == "clone_url" {
		currentValue := a.inputValue
		if a.mode == ModeInitializeBranches {
			return // No validation in branch mode
		}
		if currentValue == "" {
			a.inputValidationMsg = ""
		} else if ui.ValidateRemoteURL(currentValue) {
			a.inputValidationMsg = ""
		} else {
			a.inputValidationMsg = "Invalid URL format"
		}
	}
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	switch app.inputAction {
	case "init_branch_name":
		return app.handleInitBranchNameSubmit()
	case "init_subdir_name":
		return app.handleInputSubmitSubdirName(app)
	case "add_remote_url":
		return app.handleAddRemoteSubmit(app)
	case "commit_message":
		return app.handleCommitSubmit(app)
	case "commit_push_message":
		return app.handleCommitPushSubmit(app)
	default:
		return app, nil
	}
}


