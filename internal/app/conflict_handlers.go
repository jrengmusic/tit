package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

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

			if file.Chosen != focusedPane {
				file.Chosen = focusedPane
				columnLabel := a.conflictResolveState.ColumnLabels[focusedPane]
				a.footerHint = fmt.Sprintf(ConsoleMessages["marked_file_column"], file.Path, columnLabel)
			}
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
			app.footerHint = fmt.Sprintf(ConsoleMessages["error_writing_file"], file.Path, err)
			return app, nil
		}

		// Stage the resolved file
		result := git.Execute("add", file.Path)
		if !result.Success {
			app.footerHint = fmt.Sprintf(ConsoleMessages["error_staging_file"], file.Path, result.Stderr)
			return app, nil
		}
	}

	// Route based on operation type
	switch app.conflictResolveState.Operation {
	case "pull_merge":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizePullMerge()
	case "branch_switch":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizeBranchSwitch()
	case "dirty_pull_changeset_apply":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizeMerge)
		return app, app.cmdFinalizeDirtyPullMerge()
	case "dirty_pull_snapshot_reapply":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
		return app, app.cmdDirtyPullFinalize()
	case "time_travel_merge":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizeTimeTravelMerge()
	case "time_travel_return":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizeTimeTravelReturn()
	case OpPushSyncMerge:
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizePushSyncMerge()
	case OpMergeBranch:
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdFinalizeBranchMerge()
	case OpRebase:
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.consoleState.Reset()
		return app, app.cmdRebaseContinue()
	case "dirty_merge_changeset_apply":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizeMerge)
		return app, app.cmdFinalizeDirtyMerge()
	case "dirty_merge_snapshot_reapply":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
		return app, app.cmdDirtyMergeFinalize()
	case "dirty_switch_snapshot_reapply":
		app.StartAsyncOp()
		app.mode = ModeConsole
		app.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
		return app, app.cmdDirtySwitchFinalize()
	default:
		return app.returnToMenu()
	}
}

// handleConflictEsc exits conflict resolution and aborts the operation.
// CRITICAL: Always returns to Console to show abort operation completing.
func (a *Application) handleConflictEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.mode != ModeConflictResolve {
		return app, nil
	}

	// Route abort based on operation type
	if app.conflictResolveState != nil {
		switch app.conflictResolveState.Operation {
		case "pull_merge":
			app.StartAsyncOp()
			app.mode = ModeConsole
			app.consoleState.Reset()
			ui.GetBuffer().Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			return app, app.cmdAbortMerge()
		case OpPushSyncMerge:
			app.StartAsyncOp()
			app.mode = ModeConsole
			app.consoleState.Reset()
			ui.GetBuffer().Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			return app, app.cmdAbortMerge()
		case OpMergeBranch:
			app.StartAsyncOp()
			app.mode = ModeConsole
			app.consoleState.Reset()
			ui.GetBuffer().Append(OutputMessages["aborting_merge"], ui.TypeInfo)
			return app, app.cmdAbortMerge()
		case OpRebase:
			app.StartAsyncOp()
			app.mode = ModeConsole
			app.consoleState.Reset()
			ui.GetBuffer().Append(OutputMessages["aborting_rebase"], ui.TypeInfo)
			return app, app.cmdRebaseAbort()
		default:
			if strings.HasPrefix(app.conflictResolveState.Operation, "dirty_pull_") {
				if app.dirtyOperationState != nil {
					app.StartAsyncOp()
					app.mode = ModeConsole
					app.consoleState.Reset()
					ui.GetBuffer().Append(OutputMessages["aborting_dirty_pull"], ui.TypeInfo)
					return app, app.cmdAbortDirtyPull()
				}
			} else if strings.HasPrefix(app.conflictResolveState.Operation, "dirty_merge_") {
				if app.dirtyOperationState != nil {
					app.StartAsyncOp()
					app.mode = ModeConsole
					app.consoleState.Reset()
					ui.GetBuffer().Append(OutputMessages["dirty_merge_aborting"], ui.TypeInfo)
					return app, app.cmdAbortDirtyMerge()
				}
			} else if strings.HasPrefix(app.conflictResolveState.Operation, "dirty_switch_") {
				if app.dirtyOperationState != nil {
					app.StartAsyncOp()
					app.mode = ModeConsole
					app.consoleState.Reset()
					ui.GetBuffer().Append(OutputMessages["dirty_switch_aborting"], ui.TypeInfo)
					return app, app.cmdAbortDirtySwitch()
				}
			}
		}
	}

	return app.returnToMenu()
}
