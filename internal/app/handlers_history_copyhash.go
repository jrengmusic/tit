package app

import (
	"github.com/atotto/clipboard"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleHistoryCopyHashEnter activates CopyHashMode in the History list pane
func (a *Application) handleHistoryCopyHashEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil || !app.pickerState.History.PaneFocused {
		return app, nil
	}
	if len(app.pickerState.History.Commits) == 0 {
		return app, nil
	}

	app.pickerState.History.CopyHashMode = true
	app.pickerState.History.CopyHashFull = false
	return app, nil
}

// handleHistoryCopyHashFullEnter activates CopyHashMode with full hash copy (Y)
func (a *Application) handleHistoryCopyHashFullEnter(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil || !app.pickerState.History.PaneFocused {
		return app, nil
	}
	if len(app.pickerState.History.Commits) == 0 {
		return app, nil
	}

	app.pickerState.History.CopyHashMode = true
	app.pickerState.History.CopyHashFull = true
	return app, nil
}

// handleHistoryCopyHashEsc exits CopyHashMode without copying
func (a *Application) handleHistoryCopyHashEsc(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History != nil && app.pickerState.History.CopyHashMode {
		app.pickerState.History.CopyHashMode = false
		app.footerHint = ""
		return app, nil
	}
	return app.returnToMenu()
}

// handleHistoryCopyHashKeypress processes a keypress while CopyHashMode is active.
// Returns (handled bool, model tea.Model, cmd tea.Cmd).
// handled = true means the key was consumed and no further dispatch is needed.
func (a *Application) handleHistoryCopyHashKeypress(keyStr string) (bool, tea.Model, tea.Cmd) {
	if a.pickerState.History == nil || !a.pickerState.History.CopyHashMode {
		return false, a, nil
	}

	// Let ESC through to global handler (handleKeyESC has CopyHashMode check)
	// Block all other multi-char keys (up, down, tab, enter, ctrl+r, etc.)
	if len(keyStr) != 1 {
		if keyStr == "esc" {
			return false, a, nil
		}
		return true, a, nil
	}

	ch := rune(keyStr[0])

	// Spacebar cycles visible window by CopyHashMaxVisible, wrapping to top
	if ch == ' ' {
		commitCount := len(a.pickerState.History.Commits)
		nextIdx := a.pickerState.History.SelectedIdx + ui.CopyHashMaxVisible
		if nextIdx >= commitCount {
			nextIdx = 0
		}
		a.pickerState.History.SelectedIdx = nextIdx
		return true, a, nil
	}

	isHexChar := (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')
	if !isHexChar {
		return true, a, nil
	}

	// Compute copy-hash keys for current page — must match renderer's computation
	pageStart := (a.pickerState.History.SelectedIdx / ui.CopyHashMaxVisible) * ui.CopyHashMaxVisible
	keys := ui.ComputeCopyHashKeys(a.pickerState.History.Commits, pageStart, ui.CopyHashMaxVisible)

	// Find the commit whose key matches the pressed char
	matchedCommitIdx := -1
	for _, k := range keys {
		if k.Char == ch {
			matchedCommitIdx = k.CommitIdx
			break
		}
	}

	if matchedCommitIdx < 0 || matchedCommitIdx >= len(a.pickerState.History.Commits) {
		// Key pressed but no match — stay in CopyHashMode, ignore
		return true, a, nil
	}

	fullHash := a.pickerState.History.Commits[matchedCommitIdx].Hash
	hash := ui.ShortenHash(fullHash)
	if a.pickerState.History.CopyHashFull {
		hash = fullHash
	}
	if err := clipboard.WriteAll(hash); err == nil {
		a.footerHint = ConsoleMessages["copy_success"]
	} else {
		a.footerHint = ConsoleMessages["copy_failed"]
	}

	a.pickerState.History.CopyHashMode = false
	return true, a, nil
}
