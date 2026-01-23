package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/app"
	"tit/internal/config"
	"tit/internal/ui"
)

func main() {
	// Create default theme files (always succeeds or panics)
	ui.CreateDefaultThemeIfMissing()

	// Load config (creates default if missing, fails fast on errors)
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Load theme from config (fails fast on errors - theme files were just created)
	theme, err := ui.LoadThemeByName(cfg.Appearance.Theme)
	if err != nil {
		panic("Failed to load theme: " + err.Error())
	}

	// Start with default terminal size (will be updated by WindowSizeMsg)
	sizing := ui.CalculateDynamicSizing(80, 40)
	application := app.NewApplication(sizing, theme, cfg)

	tea.NewProgram(application,
		tea.WithAltScreen(),
	).Run()
}
