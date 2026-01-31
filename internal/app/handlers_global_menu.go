package app

import (
	"time"
	"tit/internal"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyESC handles ESC globally
// SSOT: ESC returns to previousMode (Menu mode quits app)
// Special cases: Conflict resolver (delegate), Console with async (block), Input (confirm clear)
func (a *Application) handleKeyESC(app *Application) (tea.Model, tea.Cmd) {
	// Conflict resolver mode: delegate to conflict-specific handler
	if a.mode == ModeConflictResolve {
		return a.handleConflictEsc(app)
	}

	// Block ESC in console mode while async operation is active
	// ESC aborts the operation and sets abort flag
	if (a.mode == ModeConsole || a.mode == ModeClone) && a.isAsyncActive() {
		// Kill running process (like Ctrl+C)
		if a.cancelContext != nil {
			a.cancelContext()
		}
		a.abortAsyncOp()
		// Print abort message to console using stderr color from theme
		a.consoleState.GetBuffer().Append("", ui.TypeStdout)
		a.consoleState.GetBuffer().Append("Operation aborted by user", ui.TypeStderr)
		a.consoleState.GetBuffer().Append("Press ESC to return to menu", ui.TypeInfo)
		return a, nil
	}

	// If async operation was aborted but completed: restore previous state
	if a.isAsyncAborted() {
		a.endAsyncOp()
		a.clearAsyncAborted()
		a.mode = a.workflowState.PreviousMode
		a.selectedIndex = a.workflowState.PreviousMenuIndex
		a.consoleState.Reset()
		a.consoleState.Clear()
		a.footerHint = ""

		// Regenerate menu if returning to menu mode
		if a.mode == ModeMenu {
			menu := app.GenerateMenu()
			app.menuItems = menu
			if a.workflowState.PreviousMenuIndex < len(menu) && len(menu) > 0 {
				app.footerHint = menu[a.workflowState.PreviousMenuIndex].Hint
			}
			// Rebuild shortcuts for new menu
			app.rebuildMenuShortcuts(ModeMenu)
		}
		return a, nil
	}

	// In input mode: handle based on input content
	if a.isInputMode() {
		// If input is empty: back to menu
		if a.inputState.Value == "" {
			return a.returnToMenu()
		}

		// If clear confirm active: clear input and stay
		if a.inputState.ClearConfirming {
			a.inputState.Value = ""
			a.inputState.CursorPosition = 0
			a.inputState.ValidationMsg = ""
			a.inputState.ClearConfirming = false
			a.footerHint = ""
			return a, nil
		}

		// First ESC with non-empty input: start clear confirmation
		a.inputState.ClearConfirming = true
		a.footerHint = GetFooterMessageText(MessageEscClearConfirm)
		return a, tea.Tick(internal.QuitConfirmTimeout, func(t time.Time) tea.Msg {
			return ClearTickMsg(t)
		})
	}

	// Menu mode: ESC does nothing (quit handled by Ctrl+C)
	if a.mode == ModeMenu {
		return a, nil
	}

	// Confirmation mode: ESC dismisses dialog directly (bypass handler routing)
	if a.mode == ModeConfirmation {
		return a.dismissConfirmationDialog()
	}

	// Console mode after time travel completed: go to time travel menu
	// This handles the case where time travel finishes successfully and user presses ESC
	if (a.mode == ModeConsole || a.mode == ModeClone) && a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
		a.mode = ModeMenu
		a.consoleState.Reset()
		a.consoleState.Clear()
		menu := app.GenerateMenu()
		app.menuItems = menu
		app.selectedIndex = 0
		app.footerHint = menu[0].Hint
		app.rebuildMenuShortcuts(ModeMenu)
		return a, nil
	}

	// All other modes: return to previousMode and regenerate menu
	app.mode = app.workflowState.PreviousMode

	// Regenerate menu based on new mode
	switch app.mode {
	case ModeMenu:
		menu := app.GenerateMenu()
		app.menuItems = menu
		if a.workflowState.PreviousMenuIndex >= 0 && a.workflowState.PreviousMenuIndex < len(menu) {
			app.selectedIndex = a.workflowState.PreviousMenuIndex
			app.footerHint = menu[a.workflowState.PreviousMenuIndex].Hint
		} else {
			app.selectedIndex = 0
			if len(menu) > 0 {
				app.footerHint = menu[0].Hint
			}
		}
		app.rebuildMenuShortcuts(ModeMenu)
	case ModeConfig:
		app.workflowState.PreviousMode = ModeMenu // Config always returns to menu on next ESC
		app.menuItems = app.GenerateConfigMenu()
		app.selectedIndex = 0
		if len(app.menuItems) > 0 {
			app.footerHint = app.menuItems[0].Hint
		}
		app.rebuildMenuShortcuts(ModeConfig)
	case ModePreferences:
		app.menuItems = app.GeneratePreferencesMenu()
		app.selectedIndex = 0
		if len(app.menuItems) > 0 {
			app.footerHint = app.menuItems[0].Hint
		}
		app.rebuildMenuShortcuts(ModePreferences)
	default:
		// History, FileHistory, BranchPicker: keep previousMenuIndex
		app.rebuildMenuShortcuts(app.mode)
	}

	return app, nil
}

// dismissConfirmationDialog dismisses confirmation dialog and returns to previous mode
// Used by ESC key to avoid circular dependency with confirmationHandlers map
func (a *Application) dismissConfirmationDialog() (tea.Model, tea.Cmd) {
	// Reset confirmation state
	a.dialogState.Hide()
	a.mode = a.workflowState.PreviousMode

	// Restore menu state based on previous mode
	switch a.mode {
	case ModeMenu:
		menu := a.GenerateMenu()
		a.menuItems = menu
		if a.workflowState.PreviousMenuIndex >= 0 && a.workflowState.PreviousMenuIndex < len(menu) {
			a.selectedIndex = a.workflowState.PreviousMenuIndex
			if len(menu) > 0 {
				a.footerHint = menu[a.selectedIndex].Hint
			}
		} else {
			a.selectedIndex = 0
			if len(menu) > 0 {
				a.footerHint = menu[0].Hint
			}
		}
		a.rebuildMenuShortcuts(ModeMenu)
	case ModeConfig:
		a.menuItems = a.GenerateConfigMenu()
		a.selectedIndex = 0
		if len(a.menuItems) > 0 {
			a.footerHint = a.menuItems[0].Hint
		}
		a.rebuildMenuShortcuts(ModeConfig)
	}

	return a, nil
}

// returnToMenu resets state and returns to menu mode
func (a *Application) returnToMenu() (tea.Model, tea.Cmd) {
	a.mode = ModeMenu
	a.selectedIndex = 0
	a.consoleState.Reset()
	a.consoleState.Clear()
	a.footerHint = ""
	a.inputState.Value = ""
	a.inputState.CursorPosition = 0
	a.inputState.ValidationMsg = ""
	a.inputState.ClearConfirming = false
	a.setExitAllowed(true) // ALWAYS allow exit when in menu

	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 {
		if a.gitState != nil && a.gitState.Remote == git.HasRemote && a.gitState.Timeline == "" && a.gitState.CurrentHash == "" {

		} else {
			a.footerHint = menu[0].Hint
		}
	}

	// Rebuild shortcuts for new menu
	a.rebuildMenuShortcuts(ModeMenu)

	// Restart auto-update when returning to menu
	return a, a.startAutoUpdate()
}
