package ui

import (
	"fmt"

	"tit/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// RenderPreferencesPane renders the preferences editor (3 rows: auto-update, interval, theme)
// Uses real config.Config, theme, and dynamic sizing
// Layout: 50/50 split with banner (matches menu layout)
func RenderPreferencesPane(selectedRow int, cfg *config.Config, theme Theme, sizing DynamicSizing) string {
	if cfg == nil {
		return "No config loaded"
	}

	// 50/50 split
	leftWidth := sizing.ContentInnerWidth / 2
	rightWidth := sizing.ContentInnerWidth - leftWidth

	// Build rows with theme + styling
	rows := []string{
		renderAutoUpdateRow(selectedRow == 0, cfg.AutoUpdate.Enabled, theme),
		renderIntervalRow(selectedRow == 1, cfg.AutoUpdate.IntervalMinutes, theme),
		renderThemeRow(selectedRow == 2, cfg.Appearance.Theme, theme),
	}

	// Render preferences content (centered H/V in left column)
	prefsContent := lipgloss.JoinVertical(lipgloss.Left, rows...)

	prefsColumn := lipgloss.NewStyle().
		Width(leftWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(prefsContent)

	// Render banner in right column (centered H/V like menu does)
	banner := RenderBannerDynamic(rightWidth, sizing.ContentHeight)

	bannerColumn := lipgloss.NewStyle().
		Width(rightWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(banner)

	// Join columns horizontally (50/50 split)
	return lipgloss.JoinHorizontal(lipgloss.Top, prefsColumn, bannerColumn)
}

func renderAutoUpdateRow(selected bool, enabled bool, theme Theme) string {
	status := "OFF"
	if enabled {
		status = "ON"
	}

	label := fmt.Sprintf("▸ Auto-update Enabled    %s", status)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ContentTextColor))

	if selected {
		style = style.
			Bold(true).
			Foreground(lipgloss.Color(theme.HighlightTextColor))
	}

	return style.Render(label)
}

func renderIntervalRow(selected bool, minutes int, theme Theme) string {
	label := fmt.Sprintf("  Auto-update Interval   %d min", minutes)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ContentTextColor))

	if selected {
		style = style.
			Bold(true).
			Foreground(lipgloss.Color(theme.HighlightTextColor))
	}

	return style.Render(label)
}

func renderThemeRow(selected bool, theme string, appTheme Theme) string {
	label := fmt.Sprintf("  Theme                  %s", theme)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(appTheme.ContentTextColor))

	if selected {
		style = style.
			Bold(true).
			Foreground(lipgloss.Color(appTheme.HighlightTextColor))
	}

	return style.Render(label)
}
