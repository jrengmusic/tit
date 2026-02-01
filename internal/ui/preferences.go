package ui

import (
	"fmt"
	"strings"

	"tit/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// PreferenceRow represents a single preference row with label and value
type PreferenceRow struct {
	Emoji   string
	Label   string
	Value   string
	Enabled bool
}

// BuildPreferenceRows builds preference rows from config
// Returns rows matching menu item order for consistent rendering
func BuildPreferenceRows(cfg *config.Config) []PreferenceRow {
	if cfg == nil {
		return []PreferenceRow{}
	}

	autoUpdateValue := "OFF"
	if cfg.AutoUpdate.Enabled {
		autoUpdateValue = "ON"
	}

	return []PreferenceRow{
		{Emoji: "🔄", Label: "Auto-update", Value: autoUpdateValue, Enabled: true},
		{Emoji: "⏱️", Label: "Update Interval", Value: fmt.Sprintf("%d min", cfg.AutoUpdate.IntervalMinutes), Enabled: true},
		{Emoji: "🎨", Label: "Theme", Value: cfg.Appearance.Theme, Enabled: true},
	}
}

// RenderPreferencesMenu renders preference rows as EMOJI | LABEL | VALUE
// No shortcut column - preferences use navigation only
func RenderPreferencesMenu(rows []PreferenceRow, selectedIndex int, theme Theme, contentHeight int, contentWidth int) string {
	if len(rows) == 0 {
		return ""
	}

	// Column widths (no shortcut column)
	emojiColWidth := 3
	labelColWidth := 18
	valueColWidth := 10

	// Dynamically shrink label if content is too narrow
	if contentWidth < (emojiColWidth + labelColWidth + valueColWidth) {
		labelColWidth = contentWidth - emojiColWidth - valueColWidth
		if labelColWidth < 0 {
			labelColWidth = 0
		}
	}

	menuBoxWidth := emojiColWidth + labelColWidth + valueColWidth

	var lines []string
	for i, row := range rows {
		// Column 1: EMOJI (center-aligned)
		emojiCol := row.Emoji
		emojiW := lipgloss.Width(emojiCol)
		leftPad := (emojiColWidth - emojiW) / 2
		rightPad := emojiColWidth - emojiW - leftPad
		if leftPad < 0 {
			leftPad = 0
		}
		if rightPad < 0 {
			rightPad = 0
		}
		emojiCol = strings.Repeat(" ", leftPad) + emojiCol + strings.Repeat(" ", rightPad)

		// Column 2: LABEL (left-aligned)
		labelCol := row.Label
		labelW := lipgloss.Width(labelCol)
		if labelW > labelColWidth {
			labelCol = labelCol[:labelColWidth]
			labelW = labelColWidth
		}
		labelCol = labelCol + strings.Repeat(" ", labelColWidth-labelW)

		// Column 3: VALUE (right-aligned)
		valueCol := row.Value
		valueW := lipgloss.Width(valueCol)
		if valueW > valueColWidth {
			valueCol = valueCol[:valueColWidth]
			valueW = valueColWidth
		}
		valueCol = strings.Repeat(" ", valueColWidth-valueW) + valueCol

		// Build styled line
		var styledLine string

		if i == selectedIndex {
			// Selected: highlight label+value
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.MainBackgroundColor)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true)
			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.AccentTextColor)).
				Bold(true)
			styledLine = emojiStyle.Render(emojiCol) + labelStyle.Render(labelCol) + valueStyle.Render(valueCol)
		} else {
			// Normal
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			valueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.ContentTextColor))
			styledLine = emojiStyle.Render(emojiCol) + labelStyle.Render(labelCol) + valueStyle.Render(valueCol)
		}

		lines = append(lines, styledLine)
	}

	// Center menu vertically and horizontally
	innerHeight := contentHeight - 2
	menuHeight := len(lines)
	topPad := (innerHeight - menuHeight) / 2
	if topPad < 0 {
		topPad = 0
	}
	bottomPad := innerHeight - menuHeight - topPad
	if bottomPad < 0 {
		bottomPad = 0
	}

	leftPad := (contentWidth - menuBoxWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}

	var result strings.Builder

	// Top padding
	for i := 0; i < topPad; i++ {
		result.WriteString(strings.Repeat(" ", contentWidth))
		if i < topPad-1 || menuHeight > 0 {
			result.WriteString("\n")
		}
	}

	// Menu lines
	for i, line := range lines {
		centeredLine := strings.Repeat(" ", leftPad) + line
		lineWidth := lipgloss.Width(centeredLine)
		if lineWidth < contentWidth {
			centeredLine = centeredLine + strings.Repeat(" ", contentWidth-lineWidth)
		}
		result.WriteString(centeredLine)
		if i < len(lines)-1 || bottomPad > 0 {
			result.WriteString("\n")
		}
	}

	// Bottom padding
	for i := 0; i < bottomPad; i++ {
		result.WriteString(strings.Repeat(" ", contentWidth))
		if i < bottomPad-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// RenderPreferencesWithBanner renders preferences (left) + banner (right)
// 50/50 split, same layout as main menu
func RenderPreferencesWithBanner(cfg *config.Config, selectedIndex int, theme Theme, sizing DynamicSizing) string {
	// 50/50 split
	leftWidth := sizing.ContentInnerWidth / 2
	rightWidth := sizing.ContentInnerWidth - leftWidth

	// Build preference rows from config
	rows := BuildPreferenceRows(cfg)

	// Left column: preferences menu
	menuContent := RenderPreferencesMenu(rows, selectedIndex, theme, sizing.ContentHeight, leftWidth)

	menuColumn := lipgloss.NewStyle().
		Width(leftWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Left).
		AlignVertical(lipgloss.Center).
		Render(menuContent)

	// Right column: banner (same as main menu)
	banner := RenderBannerDynamic(rightWidth, sizing.ContentHeight)

	bannerColumn := lipgloss.NewStyle().
		Width(rightWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(banner)

	return lipgloss.JoinHorizontal(lipgloss.Top, menuColumn, bannerColumn)
}
