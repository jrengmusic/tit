package app

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
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
				a.footerHint = FooterHints["already_marked_column"]
				return a, nil
			}

			// Mark this column as chosen (radio button - switches from other column)
			file.Chosen = focusedPane
			columnLabel := a.conflictResolveState.ColumnLabels[focusedPane]
			a.footerHint = fmt.Sprintf(FooterHints["marked_file_column"], file.Path, columnLabel)
		}
	}

	return a, nil
}

// handleConflictEnter applies user's file choices and continues the operation
// Routes based on the current operation type
func (a *Application) handleConflictEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve || app.conflictResolveState == nil {
		return app, nil
	}

	// Check if all files have been marked
	allMarked := true
	for _, file := range app.conflictResolveState.Files {
		if file.Chosen < 0 {
			allMarked = false
			break
		}
	}

	if !allMarked {
		app.footerHint = FooterHints["mark_all_files"]
		return app, nil
	}

	// Apply the chosen versions to all files
	for _, file := range app.conflictResolveState.Files {
		if file.Chosen < 0 || file.Chosen >= len(file.Versions) {
			continue // Skip invalid choice
		}

		// Write chosen version to file
		chosenContent := file.Versions[file.Chosen]
		if err := os.WriteFile(file.Path, []byte(chosenContent), 0644); err != nil {
			app.footerHint = fmt.Sprintf(FooterHints["error_writing_file"], file.Path, err)
			return app, nil
		}

		// Stage the resolved file
		result := git.Execute("add", file.Path)
		if !result.Success {
			app.footerHint = fmt.Sprintf(FooterHints["error_staging_file"], file.Path, result.Stderr)
			return app, nil
		}
	}

	// Route based on operation type
	if app.conflictResolveState.Operation == "pull_merge" {
		// Regular pull with merge conflicts: finalize merge
		// Transition to console to show finalization operation
		app.asyncOperationActive = true
		app.mode = ModeConsole
		app.outputBuffer.Clear()
		app.consoleState.Reset()
		return app, app.cmdFinalizePullMerge()
	} else if app.conflictResolveState.Operation == "dirty_pull_changeset_apply" {
		// Dirty pull merge conflicts resolved: commit merge before reapplying stash
		// Must finalize the merge commit before proceeding to stash apply
		app.asyncOperationActive = true
		app.mode = ModeConsole
		app.dirtyOperationState.SetPhase("finalize_merge")
		return app, app.cmdFinalizeDirtyPullMerge()
	} else if app.conflictResolveState.Operation == "dirty_pull_snapshot_reapply" {
		// Continue to finalize
		app.asyncOperationActive = true
		app.mode = ModeConsole
		app.dirtyOperationState.SetPhase("finalizing")
		return app, app.cmdDirtyPullFinalize()
	}

	// Default: return to menu
	return app.returnToMenu()
}

// handleConflictEsc exits conflict resolution and aborts the operation
// Routes to proper abort based on operation type
// CRITICAL: Always returns to Console to show abort operation completing
// User must press ESC again in console to return to menu
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

	// Route abort based on operation type
	if app.conflictResolveState != nil {
		if app.conflictResolveState.Operation == "pull_merge" {
			// Abort pull merge: transition to Console, run git merge --abort
			// User will see abort operation complete, then press ESC to return to menu
			app.asyncOperationActive = true
			app.mode = ModeConsole
			app.outputBuffer.Clear()
			app.consoleState.Reset()
			ui.GetBuffer().Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			return app, app.cmdAbortMerge()
		} else if strings.HasPrefix(app.conflictResolveState.Operation, "dirty_pull_") {
			// Abort dirty pull: transition to Console, restore original state
			// User will see abort operation complete, then press ESC to return to menu
			if app.dirtyOperationState != nil {
				app.asyncOperationActive = true
				app.mode = ModeConsole
				app.outputBuffer.Clear()
				app.consoleState.Reset()
				ui.GetBuffer().Append(OutputMessages["aborting_dirty_pull"], ui.TypeInfo)
				return app, app.cmdAbortDirtyPull()
			}
		}
	}

	// Default: return to menu (should not reach here normally)
	return app.returnToMenu()
}
