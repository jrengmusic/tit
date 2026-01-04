package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/app"
	"tit/internal/git"
	"tit/internal/ui"
)

func main() {
	// Initialize theme
	ui.CreateDefaultThemeIfMissing()
	theme, _ := ui.LoadDefaultTheme()

	// Initialize repo config
	git.CreateDefaultRepoConfigIfMissing()

	sizing := ui.CalculateSizing()
	application := app.NewApplication(sizing, theme)

	tea.NewProgram(application, 
		tea.WithAltScreen(),
	).Run()
}
