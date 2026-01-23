package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// PreferencesPane renders the preferences editor (3 rows: auto-update, interval, theme)
func RenderPreferencesPane(selectedRow int, config *Config, width int, height int) string {
	// Placeholder config type for rendering
	if config == nil {
		return "No config loaded"
	}

	const rowHeight = 1
	const padding = 2

	// Build rows
	rows := []string{
		renderAutoUpdateRow(selectedRow == 0, config.AutoUpdate.Enabled),
		renderIntervalRow(selectedRow == 1, config.AutoUpdate.IntervalMinutes),
		renderThemeRow(selectedRow == 2, config.Appearance.Theme),
	}

	// Render with spacing
	content := lipgloss.JoinVertical(lipgloss.Left, rows...)

	// Add padding
	styled := lipgloss.NewStyle().
		Width(width-padding*2).
		Padding(1, 1).
		Render(content)

	return styled
}

// Config is a placeholder for the config type
type Config struct {
	AutoUpdate AutoUpdateConfig
	Appearance AppearanceConfig
}

type AutoUpdateConfig struct {
	Enabled         bool
	IntervalMinutes int
}

type AppearanceConfig struct {
	Theme string
}

func renderAutoUpdateRow(selected bool, enabled bool) string {
	status := "OFF"
	if enabled {
		status = "ON"
	}

	label := fmt.Sprintf("▸ Auto-update Enabled    %s", status)
	if selected {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Render(label)
	}
	return label
}

func renderIntervalRow(selected bool, minutes int) string {
	label := fmt.Sprintf("  Auto-update Interval   %d min", minutes)
	if selected {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Render(label)
	}
	return label
}

func renderThemeRow(selected bool, theme string) string {
	label := fmt.Sprintf("  Theme                  %s", theme)
	if selected {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Render(label)
	}
	return label
}
