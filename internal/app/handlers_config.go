package app

import (
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Confirmation Mode Handlers
// ========================================

// handleConfirmationLeft moves selection to Yes button
func (a *Application) handleConfirmationLeft(app *Application) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() != nil {
		a.dialogState.GetDialog().SelectYes()
	}
	return a, nil
}

// handleConfirmationRight moves selection to No button
func (a *Application) handleConfirmationRight(app *Application) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() != nil {
		a.dialogState.GetDialog().SelectNo()
	}
	return a, nil
}

// handleConfirmationYes selects Yes button
func (a *Application) handleConfirmationYes(app *Application) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() != nil {
		a.dialogState.GetDialog().SelectYes()
	}
	return a, nil
}

// handleConfirmationNo selects No button
func (a *Application) handleConfirmationNo(app *Application) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() != nil {
		a.dialogState.GetDialog().SelectNo()
	}
	return a, nil
}

// handleConfirmationEnter confirms the current selection
func (a *Application) handleConfirmationEnter(app *Application) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() != nil {
		confirmed := a.dialogState.GetDialog().GetSelectedButton() == ui.ButtonYes
		return a.handleConfirmationResponse(confirmed)
	}
	return a, nil
}

// ========================================
// Config Menu Handlers
// ========================================

// handleConfigMenuEnter handles ENTER key in config menu mode
func (a *Application) handleConfigMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.menuItems) {
		return app, nil
	}
	item := app.menuItems[app.selectedIndex]

	// CONTRACT: Cannot execute separators or disabled items
	if item.Separator || !item.Enabled {
		return app, nil
	}

	// Handle back action - return to main menu
	if item.ID == "config_back" {
		return app.returnToMenu()
	}

	// Handle config menu actions
	return app, app.dispatchAction(item.ID)
}
