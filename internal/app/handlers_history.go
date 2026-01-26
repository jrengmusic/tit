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
	if app.historyState == nil {
		return app, nil
	}

	if app.historyState.PaneFocused { // List pane focused
		if app.historyState.SelectedIdx > 0 {
			app.historyState.SelectedIdx--
			// Reset details cursor when switching commits
			app.historyState.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		if app.historyState.DetailsLineCursor > 0 {
			app.historyState.DetailsLineCursor--
		}
	}
	return app, nil
}

// handleHistoryDown navigates down in history mode
func (a *Application) handleHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil {
		return app, nil
	}

	if app.historyState.PaneFocused { // List pane focused
		if app.historyState.SelectedIdx < len(app.historyState.Commits)-1 {
			app.historyState.SelectedIdx++
			// Reset details cursor when switching commits
			app.historyState.DetailsLineCursor = 0
		}
	} else { // Details pane focused - move line cursor
		// Get total lines in selected commit's details
		if app.historyState.SelectedIdx >= 0 && app.historyState.SelectedIdx < len(app.historyState.Commits) {
			commit := app.historyState.Commits[app.historyState.SelectedIdx]

			// Build details lines (must match renderHistoryDetailsPane logic)
			var totalLines int
			totalLines += 2                                       // "Author:" and "Date:" lines
			totalLines += 1                                       // Empty line separator
			totalLines += strings.Count(commit.Subject, "\n") + 1 // Commit subject lines

			// Only increment if not at the last line
			if app.historyState.DetailsLineCursor < totalLines-1 {
				app.historyState.DetailsLineCursor++
			}
		}
	}
	return app, nil
}

// handleHistoryTab switches focus between panes in history mode
func (a *Application) handleHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil {
		return app, nil
	}

	app.historyState.PaneFocused = !app.historyState.PaneFocused
	return app, nil
}

// handleHistoryEsc returns to menu from history mode
func (a *Application) handleHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	return app.returnToMenu()
}

// handleHistoryEnter handles ENTER key in history mode (Phase 7: Time Travel)
func (a *Application) handleHistoryEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.historyState == nil || app.historyState.SelectedIdx < 0 {
		return app, nil
	}

	// Get selected commit
	commit := app.historyState.Commits[app.historyState.SelectedIdx]

	// Show time travel confirmation dialog
	app.mode = ModeConfirmation
	app.confirmType = "time_travel"
	app.confirmContext = map[string]string{
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
	app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)

	return app, nil
}

// File History Mode Handlers

// updateFileHistoryDiff looks up and sets the current diff content in state
// Called whenever commit or file selection changes
// Direct cache lookup: hash:path:version → diff content
func (a *Application) updateFileHistoryDiff() {
	if a.fileHistoryState == nil {
		return
	}

	// Check bounds - need both commit and file selected
	if len(a.fileHistoryState.Commits) == 0 || a.fileHistoryState.SelectedCommitIdx >= len(a.fileHistoryState.Commits) {
		a.fileHistoryState.DiffContent = ""
		return
	}

	if len(a.fileHistoryState.Files) == 0 || a.fileHistoryState.SelectedFileIdx >= len(a.fileHistoryState.Files) {
		a.fileHistoryState.DiffContent = ""
		return
	}

	// Get selected commit and file
	commit := a.fileHistoryState.Commits[a.fileHistoryState.SelectedCommitIdx]
	file := a.fileHistoryState.Files[a.fileHistoryState.SelectedFileIdx]

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
	a.diffCacheMutex.Lock()
	diffContent, exists := a.fileHistoryDiffCache[cacheKey]
	a.diffCacheMutex.Unlock()

	if exists && diffContent != "" {
		a.fileHistoryState.DiffContent = diffContent
	} else {
		// Not cached yet (can happen if commit has >100 files)
		a.fileHistoryState.DiffContent = ""
	}
}

// handleFileHistoryUp navigates up in file(s) history mode
func (a *Application) handleFileHistoryUp(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	switch app.fileHistoryState.FocusedPane {
	case ui.PaneCommits:
		// Navigate up in commits list
		if app.fileHistoryState.SelectedCommitIdx > 0 {
			app.fileHistoryState.SelectedCommitIdx--
			// Reset file selection when switching commits
			app.fileHistoryState.SelectedFileIdx = 0
			// Update files for new commit
			if app.fileHistoryState.SelectedCommitIdx >= 0 && app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits) {
				commitHash := app.fileHistoryState.Commits[app.fileHistoryState.SelectedCommitIdx].Hash
				if gitFileList, exists := app.fileHistoryFilesCache[commitHash]; exists {
					app.fileHistoryState.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate up in files list
		if app.fileHistoryState.SelectedFileIdx > 0 {
			app.fileHistoryState.SelectedFileIdx--
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate up in diff pane (move cursor up)
		if app.fileHistoryState.DiffLineCursor > 0 {
			app.fileHistoryState.DiffLineCursor--
		}
	}
	return app, nil
}

// handleFileHistoryDown navigates down in file(s) history mode
func (a *Application) handleFileHistoryDown(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	switch app.fileHistoryState.FocusedPane {
	case ui.PaneCommits:
		// Navigate down in commits list
		if app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits)-1 {
			app.fileHistoryState.SelectedCommitIdx++
			// Reset file selection when switching commits
			app.fileHistoryState.SelectedFileIdx = 0
			// Update files for new commit
			if app.fileHistoryState.SelectedCommitIdx >= 0 && app.fileHistoryState.SelectedCommitIdx < len(app.fileHistoryState.Commits) {
				commitHash := app.fileHistoryState.Commits[app.fileHistoryState.SelectedCommitIdx].Hash
				if gitFileList, exists := app.fileHistoryFilesCache[commitHash]; exists {
					app.fileHistoryState.Files = convertGitFilesToUIFileInfo(gitFileList)
				}
			}
			// Update diff for new commit (file selection was reset to 0, so first file diff is shown)
			a.updateFileHistoryDiff()
		}
	case ui.PaneFiles:
		// Navigate down in files list
		if app.fileHistoryState.SelectedFileIdx < len(app.fileHistoryState.Files)-1 {
			app.fileHistoryState.SelectedFileIdx++
			// Update diff for newly selected file
			a.updateFileHistoryDiff()
		}
	case ui.PaneDiff:
		// Navigate down in diff pane (move cursor down)
		app.fileHistoryState.DiffLineCursor++
	}
	return app, nil
}

// handleFileHistoryTab switches focus between panes in file(s) history mode
func (a *Application) handleFileHistoryTab(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil {
		return app, nil
	}

	// Cycle through panes: Commits → Files → Diff → Commits
	app.fileHistoryState.FocusedPane = (app.fileHistoryState.FocusedPane + 1) % 3
	return app, nil
}

// handleFileHistoryCopy handles copy action in file(s) history mode
// Copies selected lines from diff pane to clipboard
func (a *Application) handleFileHistoryCopy(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState == nil || app.fileHistoryState.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Get selected lines based on visual mode
	var linesToCopy []string
	if app.fileHistoryState.VisualModeActive {
		// Visual mode: copy selected range
		linesToCopy = ui.GetSelectedLinesFromDiff(app.fileHistoryState.DiffContent, app.fileHistoryState.VisualModeStart, app.fileHistoryState.DiffLineCursor)
		// Exit visual mode after copy
		app.fileHistoryState.VisualModeActive = false
	} else {
		// Normal mode: copy current line
		linesToCopy = ui.GetSelectedLinesFromDiff(app.fileHistoryState.DiffContent, app.fileHistoryState.DiffLineCursor, app.fileHistoryState.DiffLineCursor)
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
	if app.fileHistoryState == nil || app.fileHistoryState.FocusedPane != ui.PaneDiff {
		return app, nil
	}

	// Toggle visual mode
	if app.fileHistoryState.VisualModeActive {
		// Already in visual mode - exit
		app.fileHistoryState.VisualModeActive = false
		app.footerHint = ""
	} else {
		// Enter visual mode from current cursor
		app.fileHistoryState.VisualModeActive = true
		app.fileHistoryState.VisualModeStart = app.fileHistoryState.DiffLineCursor
		app.footerHint = ConsoleMessages["visual_mode_active"]
	}

	return app, nil
}

// handleFileHistoryEsc handles ESC in file(s) history mode
// If in visual mode, exit visual mode. Otherwise, return to menu.
func (a *Application) handleFileHistoryEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.fileHistoryState != nil && app.fileHistoryState.VisualModeActive {
		// Exit visual mode, stay in file history
		app.fileHistoryState.VisualModeActive = false
		app.footerHint = ""
		return app, nil
	}
	// Not in visual mode, return to menu
	return app.returnToMenu()
}
