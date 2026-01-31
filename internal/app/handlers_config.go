package app

import (
	"context"
	"fmt"
	"strings"

	"tit/internal/git"
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

// handleConfigAddRemoteURLSubmit handles URL input from config add remote menu
func (a *Application) handleConfigAddRemoteURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	url := app.inputState.Value
	if url == "" {
		app.footerHint = "URL cannot be empty"
		return app, nil
	}

	if !ui.ValidateRemoteURL(url) {
		app.footerHint = ui.GetRemoteURLError()
		return app, nil
	}

	result := git.Execute("remote", "get-url", "origin")
	if result.Success {
		app.footerHint = ErrorMessages["remote_already_exists_validation"]
		return app, nil
	}

	app.startAsyncOp()
	app.workflowState.PreviousMode = ModeConfig
	app.workflowState.PreviousMenuIndex = app.selectedIndex
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	return app, a.cmdConfigAddRemote(url)
}

// handleConfigSwitchRemoteURLSubmit handles URL input from config switch remote menu
func (a *Application) handleConfigSwitchRemoteURLSubmit(app *Application) (tea.Model, tea.Cmd) {
	url := app.inputState.Value
	if url == "" {
		app.footerHint = "URL cannot be empty"
		return app, nil
	}

	if !ui.ValidateRemoteURL(url) {
		app.footerHint = ui.GetRemoteURLError()
		return app, nil
	}

	result := git.Execute("remote", "get-url", "origin")
	if !result.Success {
		app.footerHint = "No remote configured to switch"
		return app, nil
	}

	app.startAsyncOp()
	app.workflowState.PreviousMode = ModeConfig
	app.workflowState.PreviousMenuIndex = app.selectedIndex
	app.mode = ModeConsole
	app.consoleState.Reset()
	app.inputState.Value = ""

	return app, a.cmdConfigSwitchRemote(url)
}

// cmdConfigAddRemote adds a new remote from config menu
func (a *Application) cmdConfigAddRemote(url string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "add", "origin", url)
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_add_remote",
				Success: false,
				Error:   "Failed to add remote",
			}
		}

		return GitOperationMsg{
			Step:    "config_add_remote",
			Success: true,
			Output:  "Remote added successfully",
		}
	}
}

// cmdConfigSwitchRemote updates an existing remote URL
func (a *Application) cmdConfigSwitchRemote(url string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "set-url", "origin", url)
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_switch_remote",
				Success: false,
				Error:   "Failed to update remote URL",
			}
		}

		return GitOperationMsg{
			Step:    "config_switch_remote",
			Success: true,
			Output:  "Remote URL updated successfully",
		}
	}
}

// cmdConfigRemoveRemote removes the origin remote
func (a *Application) cmdConfigRemoveRemote() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, "remote", "remove", "origin")
		if !result.Success {
			return GitOperationMsg{
				Step:    "config_remove_remote",
				Success: false,
				Error:   "Failed to remove remote",
			}
		}

		return GitOperationMsg{
			Step:    "config_remove_remote",
			Success: true,
			Output:  "Remote removed successfully",
		}
	}
}

// ========================================
// Branch Picker Mode Handlers (SSOT: matches history navigation pattern)
// ========================================

// handleBranchPickerUp handles UP/K navigation in branch picker (list pane)
func (a *Application) handleBranchPickerUp(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.BranchPicker == nil || len(app.pickerState.BranchPicker.Branches) == 0 {
		return app, nil
	}

	if app.pickerState.BranchPicker.SelectedIdx > 0 {
		app.pickerState.BranchPicker.SelectedIdx--
	}
	return app, nil
}

// handleBranchPickerDown handles DOWN/J navigation in branch picker (list pane)
func (a *Application) handleBranchPickerDown(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.BranchPicker == nil || len(app.pickerState.BranchPicker.Branches) == 0 {
		return app, nil
	}

	if app.pickerState.BranchPicker.SelectedIdx < len(app.pickerState.BranchPicker.Branches)-1 {
		app.pickerState.BranchPicker.SelectedIdx++
	}
	return app, nil
}

// handleBranchPickerEnter handles ENTER key to switch to selected branch (with dirty tree handling)
func (a *Application) handleBranchPickerEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.BranchPicker == nil || app.pickerState.BranchPicker.SelectedIdx < 0 || app.pickerState.BranchPicker.SelectedIdx >= len(app.pickerState.BranchPicker.Branches) {
		return app, nil
	}

	selectedBranch := app.pickerState.BranchPicker.Branches[app.pickerState.BranchPicker.SelectedIdx]

	// Check if we're returning from manual detached (TimeTraveling mode)
	isReturnFromDetached := app.workflowState.IsReturnToBranch

	if isReturnFromDetached {
		// For return from manual detached: check dirty tree and show confirmation
		app.workflowState.IsReturnToBranch = false                 // Reset flag
		app.workflowState.ReturnToBranchName = selectedBranch.Name // Store target branch

		if app.workflowState.ReturnToBranchDirtyTree {
			// Show stash/discard confirmation immediately
			app.mode = ModeConfirmation
			dialog := ui.NewConfirmationDialog(
				ui.ConfirmationConfig{
					Title:       fmt.Sprintf("Return to %s with uncommitted changes", selectedBranch.Name),
					Explanation: "You have changes during time travel. Choose action:\n(Press ESC to cancel)",
					YesLabel:    "Stash changes",
					NoLabel:     "Discard changes",
					ActionID:    "time_travel_return_dirty_choice",
				},
				app.sizing.ContentInnerWidth,
				&app.theme,
			)
			app.dialogState.Show(dialog, nil)
			dialog.SelectNo()
			return app, nil
		}

		// Clean tree - switch directly
		return app, a.cmdSwitchBranch(selectedBranch.Name)
	}

	// If already on this branch, just go back to config menu
	if selectedBranch.IsCurrent {
		app.workflowState.PreviousMode = ModeMenu // Config always returns to menu
		app.mode = ModeConfig
		app.selectedIndex = 0
		configMenu := app.GenerateConfigMenu()
		app.menuItems = configMenu
		if len(configMenu) > 0 {
			app.footerHint = configMenu[0].Hint
		}
		app.rebuildMenuShortcuts(ModeConfig)
		return app, nil
	}

	// Check if working tree is clean before switching branches
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	if hasDirtyTree {
		// Show confirmation dialog for dirty tree
		app.mode = ModeConfirmation
		dialogContext := map[string]string{
			"targetBranch": selectedBranch.Name,
		}
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       fmt.Sprintf("Switch to %s with uncommitted changes?", selectedBranch.Name),
				Explanation: "You have uncommitted changes. Choose action:\n(ESC to cancel)",
				YesLabel:    "Stash changes",
				NoLabel:     "Discard changes",
				ActionID:    "branch_switch_dirty",
			},
			app.sizing.ContentInnerWidth,
			&app.theme,
		)
		app.dialogState.Show(dialog, dialogContext)
		dialog.SelectNo()
		return app, nil
	}

	// Clean tree - perform branch switch directly
	return app, a.cmdSwitchBranch(selectedBranch.Name)
}

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
