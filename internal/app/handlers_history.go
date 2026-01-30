package app

import (
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

// History Mode Handlers

// handleHistoryUp navigates up in history mode
func (a *Application) handleHistoryUp(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil {
		return app, nil
	}

	if app.pickerState.History.PaneFocused { // List pane focused
		if app.pickerState.History.SelectedIdx > 0 {
			app.pickerState.History.SelectedIdx--
			// Reset details cursor when switching commits
			app.pickerState.History.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		if app.pickerState.History.DetailsLineCursor > 0 {
			app.pickerState.History.DetailsLineCursor--
		}
	}
	return app, nil
}

// handleHistoryDown navigates down in history mode
func (a *Application) handleHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil {
		return app, nil
	}

	if app.pickerState.History.PaneFocused { // List pane focused
		if app.pickerState.History.SelectedIdx < len(app.pickerState.History.Commits)-1 {
			app.pickerState.History.SelectedIdx++
			// Reset details cursor when switching commits
			app.pickerState.History.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		// Get total lines in selected commit's details
		if app.pickerState.History.SelectedIdx >= 0 && app.pickerState.History.SelectedIdx < len(app.pickerState.History.Commits) {
			commit := app.pickerState.History.Commits[app.pickerState.History.SelectedIdx]

			// Build details lines (must match renderHistoryDetailsPane logic)
			var totalLines int
			totalLines += 2                                       // "Author:" and "Date:" lines
			totalLines += 1                                       // Empty line separator
			totalLines += strings.Count(commit.Subject, "\n") + 1 // Commit subject lines

			// Only increment if not at the last line
			if app.pickerState.History.DetailsLineCursor < totalLines-1 {
				app.pickerState.History.DetailsLineCursor++
			}
		}
	}
	return app, nil
}

// handleHistoryTab switches focus between panes in history mode
func (a *Application) handleHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil {
		return app, nil
	}

	app.pickerState.History.PaneFocused = !app.pickerState.History.PaneFocused
	return app, nil
}

// handleHistoryEsc returns to menu from history mode
func (a *Application) handleHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	return app.returnToMenu()
}

// handleHistoryEnter handles ENTER key in history mode (Phase 7: Time Travel)
func (a *Application) handleHistoryEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil || app.pickerState.History.SelectedIdx < 0 {
		return app, nil
	}

	// Get selected commit
	commit := app.pickerState.History.Commits[app.pickerState.History.SelectedIdx]

	// Show time travel confirmation dialog with context
	app.mode = ModeConfirmation
	dialogContext := map[string]string{
		"commit_hash":    commit.Hash,
		"commit_subject": commit.Subject,
	}

	// Create confirmation dialog using SSOT
	// Format: hash (first 7 chars) on first line, subject on second line
	shortHash := ui.ShortenHash(commit.Hash)

	// Extract only first line of commit message (subject)
	subject := commit.Subject
	if idx := strings.Index(subject, "\n"); idx >= 0 {
		subject = subject[:idx]
	}

	msg := ConfirmationMessages["time_travel"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: fmt.Sprintf(msg.Explanation, shortHash, subject),
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    "time_travel",
	}
	dialog := ui.NewConfirmationDialog(config, app.sizing.ContentInnerWidth, &app.theme)
	app.dialogState.Show(dialog, dialogContext)

	return app, nil
}

// File History Mode Handlers

// updateFileHistoryDiff looks up and sets the current diff content in state
// Called whenever commit or file selection changes
// Direct cache lookup: hash:path:version → diff content
func (a *Application) updateFileHistoryDiff() {
	if a.pickerState.FileHistory == nil {
		return
	}

	// Check bounds - need both commit and file selected
	if len(a.pickerState.FileHistory.Commits) == 0 || a.pickerState.FileHistory.SelectedCommitIdx >= len(a.pickerState.FileHistory.Commits) {
		a.pickerState.FileHistory.DiffContent = ""
		return
	}

	if len(a.pickerState.FileHistory.Files) == 0 || a.pickerState.FileHistory.SelectedFileIdx >= len(a.pickerState.FileHistory.Files) {
		a.pickerState.FileHistory.DiffContent = ""
		return
	}

	// Get selected commit and file
	commit := a.pickerState.FileHistory.Commits[a.pickerState.FileHistory.SelectedCommitIdx]
	file := a.pickerState.FileHistory.Files[a.pickerState.FileHistory.SelectedFileIdx]

	// Determine version based on current working tree state
	// If working tree is Dirty (has unstaged changes) → use "wip" diff (commit vs working tree)
	// Otherwise → use "parent" diff (commit vs parent commit)
	version := "parent"
	if a.gitState.WorkingTree == git.Dirty {
		version = "wip"
	}

	// Build cache key using SSOT (prevents hardcoded formats)
	cacheKey := DiffCacheKey(commit.Hash, file.Path, version)

	// Direct cache lookup (thread-safe)
	diffContent, exists := a.cacheManager.GetDiff(cacheKey)

	if exists && diffContent != "" {
		a.pickerState.FileHistory.DiffContent = diffContent
	} else {
		// Not cached yet (can happen if commit has >100 files)
		a.pickerState.FileHistory.DiffContent = ""
	}
}

// handleFileHistoryUp navigates up in file(s) history mode
func (a *Application) handleFileHistoryUp(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory == nil {
		return app, nil
	}

	switch app.pickerState.FileHistory.FocusedPane {
	case ui.PaneCommits:
		// Navigate up in commits list
		if app.pickerState.FileHistory.SelectedCommitIdx > 0 {
			app.pickerState.FileHistory.SelectedCommitIdx--
			// Reset file selection when switching commits
			app.pickerState.FileHistory.SelectedFileIdx = 0
			// Update files for new commit
			if app.pickerState.FileHistory.SelectedCommitIdx >= 0 && app.pickerState.FileHistory.SelectedCommitIdx < len(app.pickerState.FileHistory.Commits) {
				commitHash := app.pickerState.FileHistory.Commits[app.pickerState.FileHistory.SelectedCommitIdx].Hash
				if gitFileList, exists := app.cacheManager.GetFiles(commitHash); exists {
					app.pickerState.FileHistory.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate up in files list
		if app.pickerState.FileHistory.SelectedFileIdx > 0 {
			app.pickerState.FileHistory.SelectedFileIdx--
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate up in diff pane (move cursor up)
		if app.pickerState.FileHistory.DiffLineCursor > 0 {
			app.pickerState.FileHistory.DiffLineCursor--
		}
	}
	return app, nil
}

// handleFileHistoryDown navigates down in file(s) history mode
func (a *Application) handleFileHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory == nil {
		return app, nil
	}

	switch app.pickerState.FileHistory.FocusedPane {
	case ui.PaneCommits:
		// Navigate down in commits list
		if app.pickerState.FileHistory.SelectedCommitIdx < len(app.pickerState.FileHistory.Commits)-1 {
			app.pickerState.FileHistory.SelectedCommitIdx++
			// Reset file selection when switching commits
			app.pickerState.FileHistory.SelectedFileIdx = 0
			// Update files for new commit
			if app.pickerState.FileHistory.SelectedCommitIdx >= 0 && app.pickerState.FileHistory.SelectedCommitIdx < len(app.pickerState.FileHistory.Commits) {
				commitHash := app.pickerState.FileHistory.Commits[app.pickerState.FileHistory.SelectedCommitIdx].Hash
				if gitFileList, exists := app.cacheManager.GetFiles(commitHash); exists {
					app.pickerState.FileHistory.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate down in files list
		if app.pickerState.FileHistory.SelectedFileIdx < len(app.pickerState.FileHistory.Files)-1 {
			app.pickerState.FileHistory.SelectedFileIdx++
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate down in diff pane (move cursor down)
		app.pickerState.FileHistory.DiffLineCursor++
	}
	return app, nil
}

// handleFileHistoryTab switches focus between panes in file(s) history mode
func (a *Application) handleFileHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory == nil {
		return app, nil
	}

	// Cycle through panes: Commits → Files → Diff → Commits
	app.pickerState.FileHistory.FocusedPane = (app.pickerState.FileHistory.FocusedPane + 1) % 3
	return app, nil
}

// handleFileHistoryCopy handles copy action in file(s) history mode
// Copies selected lines from diff pane to clipboard
func (a *Application) handleFileHistoryCopy(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory == nil || app.pickerState.FileHistory.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Get selected lines based on visual mode
	var linesToCopy []string
	if app.pickerState.FileHistory.VisualModeActive {
		// Visual mode: copy selected range
		linesToCopy = ui.GetSelectedLinesFromDiff(app.pickerState.FileHistory.DiffContent, app.pickerState.FileHistory.VisualModeStart, app.pickerState.FileHistory.DiffLineCursor)
		// Exit visual mode after copy
		app.pickerState.FileHistory.VisualModeActive = false
	} else {
		// Normal mode: copy current line
		linesToCopy = ui.GetSelectedLinesFromDiff(app.pickerState.FileHistory.DiffContent, app.pickerState.FileHistory.DiffLineCursor, app.pickerState.FileHistory.DiffLineCursor)
	}

	// Copy to clipboard if we have lines
	if len(linesToCopy) > 0 {
		textToCopy := strings.Join(linesToCopy, "\n")
		if err := clipboard.WriteAll(textToCopy); err == nil {
			app.footerHint = ConsoleMessages["copy_success"]
		} else {
			app.footerHint = ConsoleMessages["copy_failed"]
		}
	}

	return app, nil
}

// handleFileHistoryVisualMode handles visual mode toggle in file(s) history mode
// Toggles visual selection mode, starting from current cursor position
func (a *Application) handleFileHistoryVisualMode(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory == nil || app.pickerState.FileHistory.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Toggle visual mode
	if app.pickerState.FileHistory.VisualModeActive {
		// Already in visual mode - exit
		app.pickerState.FileHistory.VisualModeActive = false
		app.footerHint = ""
	} else {
		// Enter visual mode from current cursor
		app.pickerState.FileHistory.VisualModeActive = true
		app.pickerState.FileHistory.VisualModeStart = app.pickerState.FileHistory.DiffLineCursor
		app.footerHint = ConsoleMessages["visual_mode_active"]
	}

	return app, nil
}

// handleFileHistoryEsc handles ESC in file(s) history mode
// If in visual mode, exit visual mode. Otherwise, return to menu.
func (a *Application) handleFileHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.FileHistory != nil && app.pickerState.FileHistory.VisualModeActive {
		// Exit visual mode, stay in file history
		app.pickerState.FileHistory.VisualModeActive = false
		app.footerHint = ""
		return app, nil
	}
	// Not in visual mode, return to menu
	return app.returnToMenu()
}
