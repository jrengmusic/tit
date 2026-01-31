package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Console output handlers - Common infrastructure
// Git-related handlers: handlers_console_git.go
// UI-related handlers: handlers_console_ui.go

// cmdRefreshConsole sends periodic refresh messages while async operation is active
// This forces UI re-renders to display streaming output in real-time
func (a *Application) cmdRefreshConsole() tea.Cmd {
	return tea.Tick(CacheRefreshInterval, func(t time.Time) tea.Msg {
		return OutputRefreshMsg{}
	})
}

// cmdRefreshCacheProgress sends periodic refresh messages while cache is building
// This forces UI re-renders to show cache building progress counter
// Returns a tea.Cmd that schedules continuous ticks until both caches complete
func (a *Application) cmdRefreshCacheProgress() tea.Cmd {
	return tea.Tick(CacheRefreshInterval, func(t time.Time) tea.Msg {
		return CacheRefreshTickMsg{}
	})
}
