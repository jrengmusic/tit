package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuItem re-exports the app.MenuItem for UI compatibility
// Note: The actual MenuItem is defined in app/menu.go
// This type alias allows the ui package to use it without circular imports
// In practice, we pass []app.MenuItem to RenderMenuWithHeight

// RenderMenu renders menu items (uses default content height)
func RenderMenu(items interface{}) string {
	return RenderMenuWithHeight(items, 0, Theme{}, 24)
}

// RenderMenuWithSelection renders menu items with selection highlight
func RenderMenuWithSelection(items interface{}, selectedIndex int, theme Theme) string {
	return RenderMenuWithHeight(items, selectedIndex, theme, 24)
}

// RenderMenuWithHeight renders menu items centered with 3-column layout (EMOJI | KEY | LABEL)
// items can be []app.MenuItem (passed as interface{})
func RenderMenuWithHeight(items interface{}, selectedIndex int, theme Theme, contentHeight int) string {
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
	emojiColWidth := 3
	keyColWidth := 3
	labelColWidth := 42
	menuBoxWidth := emojiColWidth + keyColWidth + labelColWidth

	// Build styled lines
	var lines []string
	for i, itemMap := range menuItems {
		// Handle separators
		if isSep, ok := itemMap["Separator"].(bool); ok && isSep {
			sepLine := strings.Repeat("─", menuBoxWidth)
			separator := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor)).
				Render(sepLine)
			lines = append(lines, separator)
			continue
		}

		emoji, _ := itemMap["Emoji"].(string)
		shortcut, _ := itemMap["Shortcut"].(string)
		label, _ := itemMap["Label"].(string)
		enabled, _ := itemMap["Enabled"].(bool)

		// Column 1: EMOJI (center-aligned)
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

		// Column 2: KEY (left-aligned)
		keyCol := shortcut + strings.Repeat(" ", keyColWidth-1)

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

		if !enabled {
			// Disabled: dimmed
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor))
			styledLine = emojiStyle.Render(emojiCol) + keyStyle.Render(keyCol) + labelStyle.Render(labelCol)
		} else if i == selectedIndex {
			// Selected: highlight background
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.SecondaryTextColor))
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.AccentTextColor)).
				Bold(true)
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.PrimaryBackground)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true)
			styledLine = emojiStyle.Render(emojiCol) + keyStyle.Render(keyCol) + labelStyle.Render(labelCol)
		} else {
			// Normal
			emojiStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.SecondaryTextColor))
			keyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.AccentTextColor)).
				Bold(true)
			labelStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.SecondaryTextColor))
			styledLine = emojiStyle.Render(emojiCol) + keyStyle.Render(keyCol) + labelStyle.Render(labelCol)
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

	// Horizontal centering
	contentWidth := 80
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
