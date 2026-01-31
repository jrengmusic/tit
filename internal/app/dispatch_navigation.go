package app

import (
	"fmt"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// dispatchReturnToBranchPicker enters branch picker for return to branch (manual detached)
func (a *Application) dispatchReturnToBranchPicker(app *Application, hasDirtyTree bool) tea.Cmd {
	// Load branches into the branch picker state
	branches, err := git.ListBranchesWithDetails()
	if err != nil {
		app.footerHint = fmt.Sprintf("Failed to load branches: %v", err)
		return nil
	}

	// Convert git.BranchDetails to ui.BranchInfo
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

	// Store dirty tree state for after branch selection
	app.workflowState.ReturnToBranchDirtyTree = hasDirtyTree
	app.workflowState.IsReturnToBranch = true // Mark this as return-from-detached

	// Initialize branch picker state
	app.pickerState.BranchPicker = &ui.BranchPickerState{
		Branches:          uiBranches,
		SelectedIdx:       0,
		PaneFocused:       true,
		ListScrollOffset:  0,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}

	// Switch to branch picker mode
	app.workflowState.PreviousMode = app.mode
	app.mode = ModeBranchPicker
	app.footerHint = "↑/↓ Navigate • Tab: Switch panes • Enter: Return to branch • ESC: Cancel"
	return nil
}
