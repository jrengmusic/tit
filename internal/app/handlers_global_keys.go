package app

import (
	"strings"
	"time"
	"tit/internal"
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
	return a, tea.Tick(internal.QuitConfirmTimeout, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
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
