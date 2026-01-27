package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuItem re-exports the app.MenuItem for UI compatibility
// Note: The actual MenuItem is defined in app/menu.go
// This type alias allows the ui package to use it without circular imports
// In practice, we pass []app.MenuItem to RenderMenuWithHeight

// RenderMenuWithHeight renders menu items centered with 3-column layout (KEY | EMOJI | LABEL)
// items can be []app.MenuItem (passed as interface{})
func RenderMenuWithHeight(items interface{}, selectedIndex int, theme Theme, contentHeight int, contentWidth int) string {
	// Type assertion to handle app.MenuItem
	var menuItems []map[string]interface{}

	switch v := items.(type) {
	case []map[string]interface{}:
		menuItems = v
	default:
		// If it's another type, try reflection or return empty
		return ""
	}

	if len(menuItems) == 0 {
		return ""
	}

	// Column widths for menu box
	keyColWidth := 12
	emojiColWidth := 3
	labelColWidth := 21
	menuBoxWidth := keyColWidth + emojiColWidth + labelColWidth

	// Build styled lines
	var lines []string
	for i, itemMap := range menuItems {
		// Handle separators
		if isSep, ok := itemMap["Separator"].(bool); ok && isSep {
			// Separator spans emoji + label columns only (not shortcut column)
			keyPad := strings.Repeat(" ", keyColWidth)
			sepLine := strings.Repeat("─", emojiColWidth+labelColWidth)
			separator := keyPad + lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.SeparatorColor)).
				Render(sepLine)
			lines = append(lines, separator)
			continue
		}

		emoji, _ := itemMap["Emoji"].(string)
		shortcut, _ := itemMap["Shortcut"].(string)
		shortcutLabel, _ := itemMap["ShortcutLabel"].(string)
		label, _ := itemMap["Label"].(string)
		enabled, _ := itemMap["Enabled"].(bool)

		// Use ShortcutLabel for display if set, otherwise use Shortcut
		shortcutDisplay := shortcut
		if shortcutLabel != "" {
			shortcutDisplay = shortcutLabel
		}

		// Column 1: KEY (right-aligned)
		keyCol := strings.Repeat(" ", keyColWidth-len(shortcutDisplay)) + shortcutDisplay + " "

		// Column 2: EMOJI (center-aligned)
		emojiCol := emoji
		emojiW := lipgloss.Width(emojiCol)
		leftPad := (emojiColWidth - emojiW) / 2
		rightPad := emojiColWidth - emojiW - leftPad
		if leftPad < 0 {
			leftPad = 0
			rightPad = 0
		}
		if rightPad < 0 {
			rightPad = 0
		}
		emojiCol = strings.Repeat(" ", leftPad) + emojiCol + strings.Repeat(" ", rightPad)

		// Column 3: LABEL (left-aligned, truncate if needed)
		labelCol := label
		labelW := lipgloss.Width(labelCol)
		if labelW > labelColWidth {
			labelCol = labelCol[:labelColWidth]
			labelW = labelColWidth
		}
		labelCol = labelCol + strings.Repeat(" ", labelColWidth-labelW)

		// Build styled line
		var styledLine string

		// Check if emoji is a spinner (braille character) for special coloring
		isSpinner := IsSpinnerFrame(emoji)

		if !enabled {
			// Disabled: dimmed text, but spinner gets vivid color
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			if isSpinner {
				emojiStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.SpinnerColor))
			}
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			styledLine = keyStyle.Render(keyCol) + emojiStyle.Render(emojiCol) + labelStyle.Render(labelCol)
		} else if i == selectedIndex {
			// Selected: highlight background
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.AccentTextColor)).
				Bold(true)
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			if isSpinner {
				emojiStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.SpinnerColor))
			}
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.MainBackgroundColor)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true)
			styledLine = keyStyle.Render(keyCol) + emojiStyle.Render(emojiCol) + labelStyle.Render(labelCol)
		} else {
			// Normal
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.AccentTextColor)).
				Bold(true)
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			if isSpinner {
				emojiStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color(theme.SpinnerColor))
			}
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.LabelTextColor))
			styledLine = keyStyle.Render(keyCol) + emojiStyle.Render(emojiCol) + labelStyle.Render(labelCol)
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

	// Horizontal centering using dynamic width
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

// RenderMenuWithBanner renders menu (left column) + banner (right column)
// 50/50 split, both columns centered H/V
func RenderMenuWithBanner(sizing DynamicSizing, items interface{}, selectedIndex int, theme Theme) string {
	// 50/50 split
	leftWidth := sizing.ContentInnerWidth / 2
	rightWidth := sizing.ContentInnerWidth - leftWidth

	// Render menu in left column (centered H/V)
	menuContent := RenderMenuWithHeight(items, selectedIndex, theme, sizing.ContentHeight, leftWidth)

	menuColumn := lipgloss.NewStyle().
		Width(leftWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(menuContent)

	// Render banner in right column (centered H/V)
	banner := RenderBannerDynamic(rightWidth, sizing.ContentHeight)

	bannerColumn := lipgloss.NewStyle().
		Width(rightWidth).
		Height(sizing.ContentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(banner)

	// Join horizontally
	return lipgloss.JoinHorizontal(lipgloss.Top, menuColumn, bannerColumn)
}
