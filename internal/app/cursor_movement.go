package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// CursorNavigationMixin provides a factory for creating cursor navigation key handlers.
type CursorNavigationMixin struct{}

// CreateHandlers generates a map of key handlers for cursor navigation
// based on the provided getter and setter functions.
func (m *CursorNavigationMixin) CreateHandlers(
	getValue func(*Application) string,
	getCursor func(*Application) int,
	setCursor func(*Application, int),
) map[string]KeyHandler {
	return map[string]KeyHandler{
		"left": func(a *Application) (tea.Model, tea.Cmd) {
			if getCursor(a) > 0 {
				setCursor(a, getCursor(a)-1)
			}
			return a, nil
		},
		"right": func(a *Application) (tea.Model, tea.Cmd) {
			if getCursor(a) < len(getValue(a)) {
				setCursor(a, getCursor(a)+1)
			}
			return a, nil
		},
		"home": func(a *Application) (tea.Model, tea.Cmd) {
			setCursor(a, 0)
			return a, nil
		},
		"end": func(a *Application) (tea.Model, tea.Cmd) {
			setCursor(a, len(getValue(a)))
			return a, nil
		},
	}
}
