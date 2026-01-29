package app

import (
	"strings"
	"time"

	"tit/internal/git"
	"tit/internal/ui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
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
	if valid, msg := validator(a.inputState.Value); !valid {
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
	if !a.canExit() {
		a.footerHint = GetFooterMessageText(MessageExitBlocked)
		return a, nil
	}

	// If async operation is running, show "in progress" message
	if a.isAsyncActive() && !a.isAsyncAborted() {
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
		return a, tea.Tick(QuitConfirmationTimeout, func(t time.Time) tea.Msg {
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
		if app.inputState.CursorPosition < 0 {
			app.inputState.CursorPosition = 0
		}
		if app.inputState.CursorPosition > len(app.inputState.Value) {
			app.inputState.CursorPosition = len(app.inputState.Value)
		}

		// Insert pasted text at cursor position (atomically, not character by character)
		app.inputState.Value = app.inputState.Value[:app.inputState.CursorPosition] + text + app.inputState.Value[app.inputState.CursorPosition:]
		app.inputState.CursorPosition += len(text)

		// Update real-time validation if in clone URL mode
		if app.inputState.Action == "clone_url" {
			if app.inputState.Value == "" {
				app.inputState.ValidationMsg = ""
			} else if ui.ValidateRemoteURL(app.inputState.Value) {
				app.inputState.ValidationMsg = "" // Valid - no error message
			} else {
				app.inputState.ValidationMsg = "Invalid URL format"
			}
		}
	}

	return app, nil
}

// handleKeySlash opens config menu when "/" is pressed in menu mode
func (a *Application) handleKeySlash(app *Application) (tea.Model, tea.Cmd) {
	if app.mode == ModeMenu {
		app.workflowState.PreviousMode = app.mode               // Track previous mode (Menu)
		app.workflowState.PreviousMenuIndex = app.selectedIndex // Track previous selection!
		app.mode = ModeConfig
		app.selectedIndex = 0
		configMenu := app.GenerateConfigMenu()
		app.menuItems = configMenu
		if len(configMenu) > 0 {
			app.footerHint = configMenu[0].Hint
		}
		app.rebuildMenuShortcuts(ModeConfig)
	}
	return app, nil
}
