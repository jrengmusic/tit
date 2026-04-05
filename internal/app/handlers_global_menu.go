package app

import (
	"time"
	"github.com/jrengmusic/tit/internal"
	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyESC handles ESC globally
// SSOT: ESC returns to previousMode (Menu mode quits app)
// Special cases: Conflict resolver (delegate), Console with async (block), Input (confirm clear)
func (a *Application) handleKeyESC(app *Application) (tea.Model, tea.Cmd) {
	if a.mode == ModeHistory && a.pickerState.History != nil && a.pickerState.History.CopyHashMode {
		return a.handleEscCopyHashMode()
	}

	if a.mode == ModeConflictResolve {
		return a.handleConflictEsc(app)
	}

	if (a.mode == ModeConsole || a.mode == ModeClone) && a.IsAsyncActive() {
		return a.handleEscAsyncAbort()
	}

	if a.IsAsyncAborted() {
		return a.handleEscPostAbort(app)
	}

	if a.isInputMode() {
		return a.handleEscInput(app)
	}

	if a.mode == ModeMenu {
		return a, nil
	}

	if a.mode == ModeConfirmation {
		return a.dismissConfirmationDialog()
	}

	if (a.mode == ModeConsole || a.mode == ModeClone) && a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
		return a.handleEscTimeTravelConsole(app)
	}

	return a.handleEscReturnToPrevious(app)
}

// handleEscCopyHashMode exits CopyHash mode, stays in history
func (a *Application) handleEscCopyHashMode() (tea.Model, tea.Cmd) {
	a.pickerState.History.CopyHashMode = false
	a.footerHint = ""
	return a, nil
}

// handleEscAsyncAbort aborts the active async operation and prints abort message to console
func (a *Application) handleEscAsyncAbort() (tea.Model, tea.Cmd) {
	if a.cancelContext != nil {
		a.cancelContext()
	}
	a.AbortAsyncOp()
	ui.GetBuffer().Append("", ui.TypeStdout)
	ui.GetBuffer().Append("Operation aborted by user", ui.TypeStderr)
	ui.GetBuffer().Append("Press ESC to return to menu", ui.TypeInfo)
	return a, nil
}

// handleEscPostAbort restores previous state after an aborted async operation completes
func (a *Application) handleEscPostAbort(app *Application) (tea.Model, tea.Cmd) {
	a.EndAsyncOp()
	a.ClearAsyncAborted()
	a.mode = a.workflowState.PreviousMode
	a.selectedIndex = a.workflowState.PreviousMenuIndex
	a.consoleState.Reset()
	a.footerHint = ""

	if a.mode == ModeMenu {
		menu := app.GenerateMenu()
		app.menuItems = menu
		if a.workflowState.PreviousMenuIndex < len(menu) && len(menu) > 0 {
			app.footerHint = menu[a.workflowState.PreviousMenuIndex].Hint
		}
		app.rebuildMenuShortcuts(ModeMenu)
	}
	return a, nil
}

// handleEscInput handles ESC in input mode: back to menu if empty, confirm clear if non-empty
func (a *Application) handleEscInput(app *Application) (tea.Model, tea.Cmd) {
	if a.inputState.Value == "" {
		return a.returnToMenu()
	}

	if a.inputState.ClearConfirming {
		a.inputState.Value = ""
		a.inputState.CursorPosition = 0
		a.inputState.ValidationMsg = ""
		a.inputState.ClearConfirming = false
		a.footerHint = ""
		return a, nil
	}

	a.inputState.ClearConfirming = true
	a.footerHint = GetFooterMessageText(MessageEscClearConfirm)
	return a, tea.Tick(internal.QuitConfirmTimeout, func(t time.Time) tea.Msg {
		return ClearTickMsg(t)
	})
}

// handleEscTimeTravelConsole exits console after time travel, returns to menu with refreshed history
func (a *Application) handleEscTimeTravelConsole(app *Application) (tea.Model, tea.Cmd) {
	a.mode = ModeMenu
	a.consoleState.Reset()
	menu := app.GenerateMenu()
	app.menuItems = menu
	app.selectedIndex = 0
	app.footerHint = menu[0].Hint
	app.rebuildMenuShortcuts(ModeMenu)
	return a, app.invalidateHistoryCaches()
}

// handleEscReturnToPrevious returns to previousMode and regenerates menu for all other modes
func (a *Application) handleEscReturnToPrevious(app *Application) (tea.Model, tea.Cmd) {
	app.mode = app.workflowState.PreviousMode

	var cmd tea.Cmd

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
		// Rebuild history caches (deferred from operation completion to avoid console noise)
		cmd = app.invalidateHistoryCaches()
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

	return app, cmd
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

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	menuItems := app.NavigationState.menuItems
	if len(menuItems) > 0 {
		startIdx := app.NavigationState.selectedIndex
		newIdx := (startIdx - 1 + len(menuItems)) % len(menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for menuItems[newIdx].Separator || !menuItems[newIdx].Enabled {
			newIdx = (newIdx - 1 + len(menuItems)) % len(menuItems)
			// Prevent infinite loop if all items disabled
			if newIdx == startIdx {
				break
			}
		}
		app.NavigationState.SelectAt(newIdx)
		// Update footer hint
		if newIdx < len(menuItems) {
			app.UIState.footerHint = menuItems[newIdx].Hint
		}
	}
	return app, nil
}

// handleMenuDown moves selection down
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	menuItems := app.NavigationState.menuItems
	if len(menuItems) > 0 {
		startIdx := app.NavigationState.selectedIndex
		newIdx := (startIdx + 1) % len(menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for menuItems[newIdx].Separator || !menuItems[newIdx].Enabled {
			newIdx = (newIdx + 1) % len(menuItems)
			// Prevent infinite loop if all items disabled
			if newIdx == startIdx {
				break
			}
		}
		app.NavigationState.SelectAt(newIdx)
		// Update footer hint
		if newIdx < len(menuItems) {
			app.UIState.footerHint = menuItems[newIdx].Hint
		}
	}
	return app, nil
}

// handleMenuEnter selects current menu item and dispatches action
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	item, ok := app.NavigationState.SelectedItem()
	if ok {
		// CONTRACT: Cannot execute separators or disabled items (cache still building)
		if !item.Separator && item.Enabled {
			// Dispatch action
			return app, app.dispatchAction(item.ID)
		}
	}
	return app, nil
}

// returnToMenu resets state and returns to menu mode
func (a *Application) returnToMenu() (tea.Model, tea.Cmd) {
	a.mode = ModeMenu
	a.selectedIndex = 0
	a.consoleState.Reset()
	a.footerHint = ""
	a.inputState.Value = ""
	a.inputState.CursorPosition = 0
	a.inputState.ValidationMsg = ""
	a.inputState.ClearConfirming = false
	a.PermitExit(true) // ALWAYS allow exit when in menu

	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 {
		isEmptyRemoteRepo := a.gitState != nil && a.gitState.Remote == git.HasRemote && a.gitState.Timeline == "" && a.gitState.CurrentHash == ""
		if !isEmptyRemoteRepo {
			a.footerHint = menu[0].Hint
		}
	}

	// Rebuild shortcuts for new menu
	a.rebuildMenuShortcuts(ModeMenu)

	// Restart auto-update when returning to menu and rebuild history caches
	return a, tea.Batch(a.startAutoUpdate(), a.invalidateHistoryCaches())
}
