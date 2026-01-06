package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Conflict Resolution Mode Handlers
// ========================================

// handleConflictUp navigates up in file list or scrolls content pane
func (a *Application) handleConflictUp(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	numCols := app.conflictResolveState.NumColumns
	focusedPane := app.conflictResolveState.FocusedPane

	if focusedPane < numCols {
		// Top row (file list) - navigate up in file list
		if app.conflictResolveState.SelectedFileIndex > 0 {
			app.conflictResolveState.SelectedFileIndex--
		}
	} else {
		// Bottom row (content pane) - move cursor up
		bottomPaneIdx := focusedPane - numCols
		if bottomPaneIdx >= 0 && bottomPaneIdx < len(app.conflictResolveState.LineCursors) {
			if app.conflictResolveState.LineCursors[bottomPaneIdx] > 0 {
				app.conflictResolveState.LineCursors[bottomPaneIdx]--
			}
		}
	}

	return app, nil
}

// handleConflictDown navigates down in file list or scrolls content pane
func (a *Application) handleConflictDown(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	numCols := app.conflictResolveState.NumColumns
	focusedPane := app.conflictResolveState.FocusedPane

	if focusedPane < numCols {
		// Top row (file list) - navigate down in file list
		if app.conflictResolveState.SelectedFileIndex < len(app.conflictResolveState.Files)-1 {
			app.conflictResolveState.SelectedFileIndex++
		}
	} else {
		// Bottom row (content pane) - move cursor down
		bottomPaneIdx := focusedPane - numCols
		if bottomPaneIdx >= 0 && bottomPaneIdx < len(app.conflictResolveState.LineCursors) {
			// Get content for this pane
			selectedFileIdx := app.conflictResolveState.SelectedFileIndex
			if selectedFileIdx >= 0 && selectedFileIdx < len(app.conflictResolveState.Files) {
				if bottomPaneIdx < len(app.conflictResolveState.Files[selectedFileIdx].Versions) {
					content := app.conflictResolveState.Files[selectedFileIdx].Versions[bottomPaneIdx]
					totalLines := len(strings.Split(content, "\n"))
					if app.conflictResolveState.LineCursors[bottomPaneIdx] < totalLines-1 {
						app.conflictResolveState.LineCursors[bottomPaneIdx]++
					}
				}
			}
		}
	}

	return app, nil
}

// handleConflictTab cycles focus through all panes (top columns + bottom columns)
func (a *Application) handleConflictTab(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	numCols := app.conflictResolveState.NumColumns
	totalPanes := numCols * 2 // Top row + bottom row

	// Cycle: 0 → 1 → 2 → ... → totalPanes-1 → 0
	app.conflictResolveState.FocusedPane = (app.conflictResolveState.FocusedPane + 1) % totalPanes

	return app, nil
}

// handleConflictSpace marks the selected file in the focused column
func (a *Application) handleConflictSpace(app *Application) (tea.Model, tea.Cmd) {
	if a.mode != ModeConflictResolve || a.conflictResolveState == nil {
		return a, nil
	}

	numCols := a.conflictResolveState.NumColumns
	focusedPane := a.conflictResolveState.FocusedPane

	// Only mark when focused on top row (file list columns)
	if focusedPane < numCols {
		fileIdx := a.conflictResolveState.SelectedFileIndex
		if fileIdx >= 0 && fileIdx < len(a.conflictResolveState.Files) {
			file := &a.conflictResolveState.Files[fileIdx]

			// Radio button behavior: if already chosen in this column, do nothing
			if file.Chosen == focusedPane {
				a.footerHint = "Already marked in this column"
				return a, nil
			}

			// Mark this column as chosen (radio button - switches from other column)
			file.Chosen = focusedPane
			a.footerHint = fmt.Sprintf("Marked: %s → column %d", file.Path, focusedPane)
		}
	}

	return a, nil
}

// handleConflictEnter applies user's file choices and continues the operation
// For now, this is a stub that will be implemented with dirty pull
func (a *Application) handleConflictEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	// TODO: Implement conflict resolution logic when dirty pull is integrated
	// For now, just return to menu
	app.footerHint = "Conflict resolution: ENTER handler (not yet implemented)"
	return app.returnToMenu()
}

// handleConflictEsc exits conflict resolution and aborts the operation
// For now, this is a stub that will be implemented with dirty pull
func (a *Application) handleConflictEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve {
		return app, nil
	}

	// Check if visual mode is active in diff pane
	if app.conflictResolveState != nil && 
	   app.conflictResolveState.DiffPane != nil && 
	   app.conflictResolveState.DiffPane.VisualModeActive {
		// Exit visual mode, stay in ConflictResolve
		app.conflictResolveState.DiffPane.VisualModeActive = false
		return app, nil
	}

	// TODO: Implement abort logic when dirty pull is integrated
	// For now, just return to menu
	app.footerHint = "Conflict resolution aborted (stub)"
	return app.returnToMenu()
}
