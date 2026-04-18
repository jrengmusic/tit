package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Preferences Mode Handlers (menu-style, reuses standard navigation)
// ========================================

// handlePreferencesIncrement increments interval by 1 minute
func (a *Application) handlePreferencesIncrement(app *Application) (tea.Model, tea.Cmd) {
	if a.appConfig == nil || len(a.menuItems) == 0 {
		return app, nil
	}

	// Only adjust if on interval row
	if a.menuItems[a.selectedIndex].ID != "preferences_interval" {
		return app, nil
	}

	newInterval := a.appConfig.AutoUpdate.IntervalMinutes + 1
	if newInterval > 60 {
		newInterval = 60
	}
	if err := a.appConfig.SetAutoUpdateInterval(newInterval); err != nil {
		app.footerHint = fmt.Sprintf("Failed to save config: %v", err)
	}

	return app, nil
}

// handlePreferencesDecrement decrements interval by 1 minute
func (a *Application) handlePreferencesDecrement(app *Application) (tea.Model, tea.Cmd) {
	if a.appConfig == nil || len(a.menuItems) == 0 {
		return app, nil
	}

	// Only adjust if on interval row
	if a.menuItems[a.selectedIndex].ID != "preferences_interval" {
		return app, nil
	}

	newInterval := a.appConfig.AutoUpdate.IntervalMinutes - 1
	if newInterval < 1 {
		newInterval = 1
	}
	if err := a.appConfig.SetAutoUpdateInterval(newInterval); err != nil {
		app.footerHint = fmt.Sprintf("Failed to save config: %v", err)
	}

	return app, nil
}

// handlePreferencesIncrement10 increments interval by 10 minutes
func (a *Application) handlePreferencesIncrement10(app *Application) (tea.Model, tea.Cmd) {
	if a.appConfig == nil || len(a.menuItems) == 0 {
		return app, nil
	}

	if a.menuItems[a.selectedIndex].ID != "preferences_interval" {
		return app, nil
	}

	newInterval := a.appConfig.AutoUpdate.IntervalMinutes + 10
	if newInterval > 60 {
		newInterval = 60
	}
	if err := a.appConfig.SetAutoUpdateInterval(newInterval); err != nil {
		app.footerHint = fmt.Sprintf("Failed to save config: %v", err)
	}

	return app, nil
}

// handlePreferencesDecrement10 decrements interval by 10 minutes
func (a *Application) handlePreferencesDecrement10(app *Application) (tea.Model, tea.Cmd) {
	if a.appConfig == nil || len(a.menuItems) == 0 {
		return app, nil
	}

	if a.menuItems[a.selectedIndex].ID != "preferences_interval" {
		return app, nil
	}

	newInterval := a.appConfig.AutoUpdate.IntervalMinutes - 10
	if newInterval < 1 {
		newInterval = 1
	}
	if err := a.appConfig.SetAutoUpdateInterval(newInterval); err != nil {
		app.footerHint = fmt.Sprintf("Failed to save config: %v", err)
	}

	return app, nil
}

// handlePreferencesEnter handles Enter/Space in preferences mode
// Dispatches action based on selected menu item ID
func (a *Application) handlePreferencesEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.menuItems) {
		return app, nil
	}
	item := app.menuItems[app.selectedIndex]

	// CONTRACT: Cannot execute separators or disabled items
	if item.Separator || !item.Enabled {
		return app, nil
	}

	// Dispatch action for selected preference
	return app, app.dispatchAction(item.ID)
}
