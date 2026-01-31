package app

import (
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

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
