package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// buildKeyHandlers creates the complete key handler registry for all application modes
func (a *Application) buildKeyHandlers() map[AppMode]map[string]KeyHandler {
	// Global handlers - highest priority, applied to all modes
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash, // Open config menu
		"ctrl+v": a.handleKeyPaste, // Linux/Windows/macOS
		"cmd+v":  a.handleKeyPaste, // macOS cmd+v
		"meta+v": a.handleKeyPaste, // macOS meta (cmd) - Bubble Tea may send this
		"alt+v":  a.handleKeyPaste, // Fallback
	}

	cursorNavMixin := CursorNavigationMixin{}

	// Generic input cursor handlers for single-field inputs
	genericInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string { return a.inputState.Value },
		func(a *Application) int { return a.inputState.CursorPosition },
		func(a *Application, pos int) { a.inputState.CursorPosition = pos },
	)

	// Mode-specific handlers (global merged in after)
	modeHandlers := map[AppMode]map[string]KeyHandler{
		ModeMenu: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			Build(),
		ModeConsole: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
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
		ModeHistory: NewModeHandlers().
			On("up", a.handleHistoryUp).
			On("k", a.handleHistoryUp).
			On("down", a.handleHistoryDown).
			On("j", a.handleHistoryDown).
			On("tab", a.handleHistoryTab).
			On("enter", a.handleHistoryEnter).
			On("ctrl+r", a.handleHistoryRewind).
			On("esc", a.handleHistoryEsc).
			Build(),
		ModeFileHistory: NewModeHandlers().
			On("up", a.handleFileHistoryUp).
			On("down", a.handleFileHistoryDown).
			On("k", a.handleFileHistoryUp).
			On("j", a.handleFileHistoryDown).
			On("tab", a.handleFileHistoryTab).
			On("y", a.handleFileHistoryCopy).
			On("v", a.handleFileHistoryVisualMode).
			On("esc", a.handleFileHistoryEsc).
			Build(),
		ModeConflictResolve: NewModeHandlers().
			On("up", a.handleConflictUp).
			On("k", a.handleConflictUp).
			On("down", a.handleConflictDown).
			On("j", a.handleConflictDown).
			On("tab", a.handleConflictTab).
			On(" ", a.handleConflictSpace). // Space character, not "space"
			On("enter", a.handleConflictEnter).
			Build(),
		ModeClone: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeSelectBranch: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleSelectBranchEnter).
			Build(),
		ModeSetupWizard: NewModeHandlers().
			On("enter", a.handleSetupWizardEnter).
			Build(),
		ModeConfig: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			Build(),
		ModeBranchPicker: NewModeHandlers().
			On("up", a.handleBranchPickerUp).
			On("k", a.handleBranchPickerUp).
			On("down", a.handleBranchPickerDown).
			On("j", a.handleBranchPickerDown).
			On("enter", a.handleBranchPickerEnter).
			Build(),
		ModePreferences: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter).
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
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
func (a *Application) rebuildMenuShortcuts(mode AppMode) {
	if a.keyHandlers[mode] == nil {
		a.keyHandlers[mode] = make(map[string]KeyHandler)
	}

	// Remove old shortcut handlers (keep navigation and enter)
	// We'll rebuild from scratch by first copying base handlers

	var baseHandlers map[string]KeyHandler
	if mode == ModeMenu {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			On(" ", a.handleMenuEnter). // Space as enter alias
			Build()
	} else if mode == ModeConfig {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			On(" ", a.handleConfigMenuEnter). // Space as enter alias
			Build()
	} else if mode == ModePreferences {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter). // Space as enter alias
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
			Build()
	}

	// Start fresh
	newHandlers := make(map[string]KeyHandler)

	// Copy base handlers
	for key, handler := range baseHandlers {
		newHandlers[key] = handler
	}

	// Merge global handlers
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash, // Open config menu
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	// Add global handlers (base handlers take priority, no overrides)
	for key, handler := range globalHandlers {
		if _, exists := baseHandlers[key]; !exists {
			newHandlers[key] = handler
		}
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

	// Replace handlers for specified mode
	a.keyHandlers[mode] = newHandlers
}
