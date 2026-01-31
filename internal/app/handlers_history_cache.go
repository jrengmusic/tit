package app

import (
	"fmt"
	"strings"

	"tit/internal/ui"

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
