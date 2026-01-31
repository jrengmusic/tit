package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// dispatchTimeTravelMerge handles the "Merge back" action during time travel
func (a *Application) dispatchTimeTravelMerge(app *Application) tea.Cmd {
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	var confirmType ConfirmationType
	if hasDirtyTree {
		confirmType = ConfirmTimeTravelMergeDirty
	} else {
		confirmType = ConfirmTimeTravelMerge
	}

	app.workflowState.PreviousMode = app.mode
	app.mode = ModeConfirmation
	msg := ConfirmationMessages[string(confirmType)]
	dialog := ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       msg.Title,
			Explanation: msg.Explanation,
			YesLabel:    msg.YesLabel,
			NoLabel:     msg.NoLabel,
			ActionID:    string(confirmType),
		},
		a.sizing.ContentInnerWidth,
		&a.theme,
	)
	app.dialogState.Show(dialog, nil)
	dialog.SelectNo()
	return nil
}

// dispatchTimeTravelReturn handles the "Return without merge" action during time travel
func (a *Application) dispatchTimeTravelReturn(app *Application) tea.Cmd {
	statusResult := git.Execute("status", "--porcelain")
	hasDirtyTree := statusResult.Success && strings.TrimSpace(statusResult.Stdout) != ""

	// Check if we have original branch from TIT marker
	originalBranch := ""
	travelInfoPath := filepath.Join(".git", "TIT_TIME_TRAVEL")
	data, err := os.ReadFile(travelInfoPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			originalBranch = lines[0]
		}
	}

	// If no original branch (manual detached with multiple branches), show branch picker
	if originalBranch == "" {
		return app.dispatchReturnToBranchPicker(app, hasDirtyTree)
	}

	if hasDirtyTree {
		app.workflowState.PreviousMode = app.mode
		app.mode = ModeConfirmation
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       fmt.Sprintf("Return to %s with uncommitted changes", originalBranch),
				Explanation: "You have changes during time travel. Choose action:\n(Press ESC to cancel)",
				YesLabel:    "Merge changes",
				NoLabel:     "Discard changes",
				ActionID:    "time_travel_return_dirty_choice",
			},
			a.sizing.ContentInnerWidth,
			&a.theme,
		)
		app.dialogState.Show(dialog, nil)
		dialog.SelectNo()
	} else {
		app.workflowState.PreviousMode = app.mode
		app.mode = ModeConfirmation
		msg := ConfirmationMessages[string(ConfirmTimeTravelReturn)]
		msg.Title = fmt.Sprintf("Return to %s", originalBranch)
		dialog := ui.NewConfirmationDialog(
			ui.ConfirmationConfig{
				Title:       msg.Title,
				Explanation: msg.Explanation,
				YesLabel:    msg.YesLabel,
				NoLabel:     msg.NoLabel,
				ActionID:    string(ConfirmTimeTravelReturn),
			},
			a.sizing.ContentInnerWidth,
			&a.theme,
		)
		app.dialogState.Show(dialog, nil)
		dialog.SelectNo()
	}
	return nil
}
