package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/app"
	"tit/internal/ui"
)

func main() {
	// Initialize theme
	ui.CreateDefaultThemeIfMissing()
	theme, _ := ui.LoadDefaultTheme()

	// Start with default terminal size (will be updated by WindowSizeMsg)
	sizing := ui.CalculateDynamicSizing(80, 40)
	application := app.NewApplication(sizing, theme)

	tea.NewProgram(application,
		tea.WithAltScreen(),
	).Run()
}
