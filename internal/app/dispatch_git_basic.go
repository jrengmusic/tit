package app

import (
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// dispatchPush pushes to remote
func (a *Application) dispatchPush(app *Application) tea.Cmd {
	a.startAsyncOp()
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0
	a.mode = ModeConsole
	a.consoleState.Reset()
	return app.cmdPush()
}

// dispatchPullMerge pulls with merge strategy
func (a *Application) dispatchPullMerge(app *Application) tea.Cmd {
	confirmType := string(ConfirmPullMerge)
	if app.gitState.Timeline == git.Diverged {
		confirmType = string(ConfirmPullMergeDiverged)
	}
	app.workflowState.PreviousMode = app.mode // Track previous mode (Menu)
	app.mode = ModeConfirmation
	msg := ConfirmationMessages[confirmType]
	dialog := ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       msg.Title,
			Explanation: msg.Explanation,
			YesLabel:    msg.YesLabel,
			NoLabel:     msg.NoLabel,
			ActionID:    confirmType,
		},
		a.sizing.ContentInnerWidth,
		&a.theme,
	)
	app.dialogState.Show(dialog, nil)
	dialog.SelectNo()
	return nil
}

// dispatchForcePush shows confirmation dialog for force push
func (a *Application) dispatchForcePush(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode // Track previous mode (Menu)
	app.mode = ModeConfirmation
	app.dialogState.SetContext(map[string]string{})
	msg := ConfirmationMessages["force_push"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "force_push",
	}
	dialog := ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, nil)
	return nil
}

// dispatchReplaceLocal shows confirmation dialog for destructive action
func (a *Application) dispatchReplaceLocal(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.dialogState.SetContext(map[string]string{})
	msg := ConfirmationMessages["hard_reset"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "hard_reset",
	}
	dialog := ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, nil)
	dialog.SelectNo()
	return nil
}

// dispatchResetDiscardChanges shows confirmation dialog for discarding all changes
func (a *Application) dispatchResetDiscardChanges(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.dialogState.SetContext(map[string]string{})

	var confirmType string
	if app.gitState.Remote == git.HasRemote {
		confirmType = "confirm_discard_changes_remote_choice"
	} else {
		confirmType = "confirm_discard_changes_local"
	}

	msg := ConfirmationMessages[confirmType]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    confirmType,
	}
	dialog := ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, nil)

	// For simple discard (Yes/Cancel), select Cancel (No) by default for safety
	// For remote choice (Local/Remote), select Local (Yes) by default
	if confirmType == "confirm_discard_changes_local" {
		dialog.SelectNo()
	} else {
		dialog.SelectYes()
	}
	return nil
}

// dispatchDirtyPullMerge starts the dirty pull confirmation dialog
func (a *Application) dispatchDirtyPullMerge(app *Application) tea.Cmd {
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	app.dialogState.SetContext(map[string]string{})
	msg := ConfirmationMessages["dirty_pull"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "dirty_pull",
	}
	dialog := ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, nil)
	return nil
}
