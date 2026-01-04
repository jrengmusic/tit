package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// KeyHandler is a function type for handling keyboard input
type KeyHandler func(*Application) (tea.Model, tea.Cmd)
