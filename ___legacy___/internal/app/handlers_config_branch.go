package app

import (
	"fmt"
	"strings"

	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Branch Picker Mode Handlers (SSOT: matches history navigation pattern)
// ========================================

// handleBranchPickerUp handles UP/K navigation in branch picker (list pane)
func (a *Application) handleBranchPickerUp(app *Application) (tea.Model, tea.Cmd) {
	picker := app.pickerState.BranchPicker
	if picker != nil && len(picker.Branches) > 0 && picker.SelectedIdx > 0 {
		picker.SelectedIdx--
	}
	return app, nil
}

// handleBranchPickerDown handles DOWN/J navigation in branch picker (list pane)
func (a *Application) handleBranchPickerDown(app *Application) (tea.Model, tea.Cmd) {
	picker := app.pickerState.BranchPicker
	if picker != nil && len(picker.Branches) > 0 && picker.SelectedIdx < len(picker.Branches)-1 {
		picker.SelectedIdx++
	}
	return app, nil
}

// handleBranchPickerEnter handles ENTER key to switch to selected branch (with dirty tree handling)
func (a *Application) handleBranchPickerEnter(app *Application) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var model tea.Model = app

	picker := app.pickerState.BranchPicker
	valid := picker != nil && picker.SelectedIdx >= 0 && picker.SelectedIdx < len(picker.Branches)

	if valid {
		selectedBranch := picker.Branches[picker.SelectedIdx]
		isReturnFromDetached := app.workflowState.IsReturnToBranch

		switch {
		case isReturnFromDetached && app.workflowState.ReturnToBranchDirtyTree:
			app.workflowState.IsReturnToBranch = false
			app.workflowState.ReturnToBranchName = selectedBranch.Name
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

		case isReturnFromDetached:
			app.workflowState.IsReturnToBranch = false
			app.workflowState.ReturnToBranchName = selectedBranch.Name
			cmd = a.cmdSwitchBranch(selectedBranch.Name)

		case app.workflowState.BranchPickerPurpose == BranchPickerPurposeMerge:
			app.workflowState.BranchPickerPurpose = ""
			model, cmd = app.handleMergeBranchSelection(selectedBranch.Name)

		case selectedBranch.IsCurrent:
			app.workflowState.PreviousMode = ModeMenu
			app.mode = ModeConfig
			app.selectedIndex = 0
			configMenu := app.GenerateConfigMenu()
			app.menuItems = configMenu
			if len(configMenu) > 0 {
				app.footerHint = configMenu[0].Hint
			}
			app.rebuildMenuShortcuts(ModeConfig)

		default:
			statusResult := git.Execute("status", "--porcelain")
			hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""
			if hasDirtyTree {
				app.mode = ModeConfirmation
				dialogContext := map[string]string{"targetBranch": selectedBranch.Name}
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
			} else {
				cmd = a.cmdSwitchBranch(selectedBranch.Name)
			}
		}
	}

	return model, cmd
}

// handleBranchPickerAdd handles "a" key in branch picker — opens new-branch name input
func (a *Application) handleBranchPickerAdd(app *Application) (tea.Model, tea.Cmd) {
	app.workflowState.BranchPickerReturnAfterCreate = true
	app.workflowState.PreviousMode = ModeBranchPicker
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "New branch name:",
		InputAction: "new_branch_name",
		FooterHint:  "Enter branch name, press Enter to create and switch",
	})
	return app, nil
}

// handleBranchPickerMerge handles "m" key in branch picker — merges selected branch into current
func (a *Application) handleBranchPickerMerge(app *Application) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var model tea.Model = app

	picker := app.pickerState.BranchPicker
	valid := picker != nil && picker.SelectedIdx >= 0 && picker.SelectedIdx < len(picker.Branches)

	if valid {
		sel := picker.Branches[picker.SelectedIdx]
		if !sel.IsCurrent {
			app.workflowState.BranchPickerPurpose = ""
			model, cmd = app.handleMergeBranchSelection(sel.Name)
		}
	}

	return model, cmd
}

// handleBranchPickerDelete handles "x" key in branch picker — confirms then deletes selected branch
func (a *Application) handleBranchPickerDelete(app *Application) (tea.Model, tea.Cmd) {
	picker := app.pickerState.BranchPicker
	valid := picker != nil && picker.SelectedIdx >= 0 && picker.SelectedIdx < len(picker.Branches)

	if valid {
		sel := picker.Branches[picker.SelectedIdx]
		if !sel.IsCurrent {
			app.workflowState.PreviousMode = ModeBranchPicker
			app.mode = ModeConfirmation
			dialogContext := map[string]string{"targetBranch": sel.Name}
			dialog := ui.NewConfirmationDialog(
				ui.ConfirmationConfig{
					Title:       fmt.Sprintf("Delete branch %s?", sel.Name),
					Explanation: "This cannot be undone. Force delete will drop unmerged commits on this branch.",
					YesLabel:    "Delete",
					NoLabel:     "Cancel",
					ActionID:    "branch_delete",
				},
				app.sizing.ContentInnerWidth,
				&app.theme,
			)
			app.dialogState.Show(dialog, dialogContext)
			dialog.SelectNo()
		}
	}

	return app, nil
}

// refreshBranchPicker reloads branch list and updates picker state.
// If selectName is non-empty, the matching branch is focused; otherwise the old index is clamped.
func (a *Application) refreshBranchPicker(selectName string) error {
	branches, err := git.ListBranchesWithDetails()

	if err == nil {
		uiBranches := make([]ui.BranchInfo, len(branches))
		for i, b := range branches {
			uiBranches[i] = ui.BranchInfo{
				Name:           b.Name,
				IsCurrent:      b.IsCurrent,
				LastCommitTime: b.LastCommitTime,
				LastCommitHash: b.LastCommitHash,
				LastCommitSubj: b.LastCommitSubj,
				Author:         b.Author,
				TrackingRemote: b.TrackingRemote,
				Ahead:          b.Ahead,
				Behind:         b.Behind,
			}
		}

		a.pickerState.BranchPicker.Branches = uiBranches

		newIdx := a.pickerState.BranchPicker.SelectedIdx
		if selectName != "" {
			for i, b := range uiBranches {
				if b.Name == selectName {
					newIdx = i
					break
				}
			}
		}

		maxIdx := len(uiBranches) - 1
		if newIdx > maxIdx {
			newIdx = maxIdx
		}
		if newIdx < 0 {
			newIdx = 0
		}
		a.pickerState.BranchPicker.SelectedIdx = newIdx
	}

	return err
}
