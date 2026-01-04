package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyCtrlC handles Ctrl+C with confirmation
func (a *Application) handleKeyCtrlC(app *Application) (tea.Model, tea.Cmd) {
	if a.quitConfirmActive {
		// Second Ctrl+C - quit immediately
		return a, tea.Quit
	}

	// First Ctrl+C - start confirmation timer and set footer hint
	a.quitConfirmActive = true
	a.quitConfirmTime = time.Now()
	a.footerHint = GetFooterMessageText(MessageCtrlCConfirm)
	return a, tea.Tick(QuitConfirmationTimeout, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
