package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrengmusic/tit/internal/app"
	"github.com/jrengmusic/tit/internal/config"
	"github.com/jrengmusic/tit/internal/ui"
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

	opts := []tea.ProgramOption{tea.WithAltScreen()}

	reader, restore := platformInput()
	if reader != nil {
		opts = append(opts, tea.WithInput(reader))
	}
	if restore != nil {
		defer restore()
	}

	tea.NewProgram(application, opts...).Run()
}
