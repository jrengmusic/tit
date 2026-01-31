package app

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// UI-related console handlers

// handleConsoleUp scrolls console up one line
func (a *Application) handleConsoleUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	app.consoleState.ScrollUp()
	app.consoleState.SetAutoScroll(false) // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsoleDown scrolls console down one line
func (a *Application) handleConsoleDown(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	app.consoleState.ScrollDown()
	app.consoleState.SetAutoScroll(false) // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsolePageUp scrolls console up one page
func (a *Application) handleConsolePageUp(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	// UI THREAD - Scroll console up by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollUp()
	}
	app.consoleState.SetAutoScroll(false) // Disable auto-scroll on manual scroll
	return app, nil
}

// handleConsolePageDown scrolls console down one page
func (a *Application) handleConsolePageDown(app *Application) (tea.Model, tea.Cmd) {
	// Block scrolling during async operations
	if app.isAsyncActive() {
		return app, nil
	}
	// UI THREAD - Scroll console down by page (10 lines)
	for i := 0; i < 10; i++ {
		app.consoleState.ScrollDown()
	}
	app.consoleState.SetAutoScroll(false) // Disable auto-scroll on manual scroll
	return app, nil
}

// cmdFetchRemote runs git fetch in background to sync remote refs
// Called on startup when HasRemote is detected to ensure timeline accuracy
func cmdFetchRemote() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "fetch", "--quiet")
		err := cmd.Run()
		if err != nil {
			return RemoteFetchMsg{Success: false, Error: err.Error()}
		}
		return RemoteFetchMsg{Success: true}
	}
}
