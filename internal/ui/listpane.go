package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ListPane represents a reusable list pane component with consistent
// border styling, title rendering, and selection highlighting.
type ListPane struct {
	Title        string
	ScrollOffset int
	Theme        *Theme
}

// ListItem represents a single item in the list with attribute and content parts.
// The attribute (left side) keeps its color regardless of selection.
// The content (right side) gets highlighted when selected.
type ListItem struct {
	AttributeText  string // Left attribute text (date/time, status, checkbox)
	AttributeColor string // Color for attribute (hex color code)
	ContentText    string // Main content text (hash, filename)
	ContentColor   string // Color for content when not selected (hex color code)
	ContentBold    bool   // Whether content should be bold when not selected
	IsSelected     bool   // True if this item is currently selected
}

// NewListPane creates a new ListPane instance with the given title and theme
func NewListPane(title string, theme *Theme) *ListPane {
	return &ListPane{
		Title:        title,
		ScrollOffset: 0,
		Theme:        theme,
	}
}

// Render renders the list pane as a single bordered box string.
// The box will be exactly width x height characters (including borders).
func (lp *ListPane) Render(items []ListItem, width, height int, isActive bool, columnPos int, numColumns int) string {
	// Border color based on focus state (like old-tit)
	borderColor := lp.Theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = lp.Theme.ConflictPaneFocusedBorder
	}

	// Calculate content width accounting for border and padding
	// Box will have borders (2 chars) and padding (2 chars: left+right)
	// So content area = width - 2 (border) - 2 (padding) = width - 4
	contentWidth := width - 4

	if contentWidth <= 0 {
		return ""
	}

	// Build content lines
	var contentLines []string

	// Title line (centered, bold)
	contentLines = append(contentLines, lp.renderTitle(contentWidth))

	// Empty separator line
	contentLines = append(contentLines, "")

	// Calculate visible lines for items
	// height - border - title - separator
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Render items with scrolling
	itemLines := lp.renderItems(items, contentWidth, visibleLines, isActive)
	contentLines = append(contentLines, itemLines...)

	// Join all content lines
	content := strings.Join(contentLines, "\n")

	// Add border with padding - ALL FOUR SIDES like old-tit
	// Use Width/Height to enforce exact size, content is already properly sized
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - 2).
		Height(height).
		Padding(0, 1) // 1 char padding left/right

	return boxStyle.Render(content)
}

// renderTitle renders the centered, bold title line
func (lp *ListPane) renderTitle(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(lp.Theme.ConflictPaneTitleColor)).
		Bold(true)

	styledTitle := titleStyle.Render(lp.Title)
	titleWidth := lipgloss.Width(styledTitle)

	// Center the title
	leftPad := (width - titleWidth) / 2
	rightPad := width - titleWidth - leftPad

	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}

	return strings.Repeat(" ", leftPad) + styledTitle + strings.Repeat(" ", rightPad)
}

// renderItems renders the visible items with scrolling and padding
func (lp *ListPane) renderItems(items []ListItem, width, visibleLines int, isActive bool) []string {
	if len(items) == 0 {
		// Empty state
		emptyMsg := "No items"
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(lp.Theme.DimmedTextColor))

		styledMsg := emptyStyle.Render(emptyMsg)
		msgWidth := lipgloss.Width(styledMsg)

		leftPad := (width - msgWidth) / 2
		rightPad := width - msgWidth - leftPad

		if leftPad < 0 {
			leftPad = 0
		}
		if rightPad < 0 {
			rightPad = 0
		}

		emptyLine := strings.Repeat(" ", leftPad) + styledMsg + strings.Repeat(" ", rightPad)

		// Return empty message + padding to fill visibleLines
		result := []string{emptyLine}
		for len(result) < visibleLines {
			result = append(result, strings.Repeat(" ", width))
		}
		return result
	}

	// Calculate scroll window
	start := lp.ScrollOffset
	end := start + visibleLines

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if end > len(items) {
		end = len(items)
	}
	if start > len(items)-visibleLines && visibleLines < len(items) {
		start = len(items) - visibleLines
	}
	if start < 0 {
		start = 0
	}

	var lines []string

	// Render visible items
	for i := start; i < end; i++ {
		itemLine := lp.renderItem(items[i], width, isActive)
		lines = append(lines, itemLine)
	}

	// Pad remaining lines to fill visibleLines
	for len(lines) < visibleLines {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return lines
}

// renderItem renders a single item line with attribute + content
func (lp *ListPane) renderItem(item ListItem, width int, isActive bool) string {
	// Render attribute with its color (never changes)
	attributeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(item.AttributeColor))
	styledAttribute := attributeStyle.Render(item.AttributeText)

	// Calculate content width (full width minus attribute and space)
	contentWidth := width
	if item.AttributeText != "" {
		contentWidth = width - lipgloss.Width(styledAttribute) - 1 // -1 for space
	}

	// Render content with selection highlighting - use Width() to fill background
	var contentStyle lipgloss.Style
	if item.IsSelected {
		if isActive {
			// Focused selection: match menu convention (dark foreground on teal background)
			contentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(lp.Theme.MainBackgroundColor)).
				Background(lipgloss.Color(lp.Theme.MenuSelectionBackground)).
				Bold(true).
				Width(contentWidth)
		} else {
			// Unfocused selection: bright foreground, no background
			contentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(lp.Theme.AccentTextColor)).
				Bold(true).
				Width(contentWidth)
		}
	} else {
		// Normal item: use caller-specified color and bold setting
		contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(item.ContentColor)).
			Bold(item.ContentBold).
			Width(contentWidth)
	}
	styledContent := contentStyle.Render(item.ContentText)

	// Combine attribute + space + content
	if item.AttributeText != "" {
		return styledAttribute + " " + styledContent
	}
	return styledContent
}

// AdjustScroll adjusts the scroll offset based on the selected index and visible lines
func (lp *ListPane) AdjustScroll(selectedIdx, visibleLines int) {
	if selectedIdx < lp.ScrollOffset {
		lp.ScrollOffset = selectedIdx
	} else if selectedIdx >= lp.ScrollOffset+visibleLines {
		lp.ScrollOffset = selectedIdx - visibleLines + 1
	}
}
