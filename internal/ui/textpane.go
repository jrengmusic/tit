package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderTextPane renders a scrollable text pane with optional line numbers and cursor
// This is the SSOT for all scrollable text rendering (conflict resolver, history details, etc.)
//
// Parameters:
//   - content: Text to display (newline-separated)
//   - width, height: Dimensions (including border)
//   - lineCursor: Current line cursor position (0-indexed, -1 for no cursor)
//   - scrollOffset: Vertical scroll position (0-indexed)
//   - showLineNumbers: Whether to show line numbers in left gutter
//   - isActive: Whether this pane has focus (affects border color and cursor highlight)
//   - theme: Theme for colors
//
// Returns:
//   - rendered: The rendered pane string
//   - newScrollOffset: Updated scroll offset (auto-adjusted to keep cursor visible)
func RenderTextPane(
	content string,
	width int,
	height int,
	lineCursor int,
	scrollOffset int,
	showLineNumbers bool,
	isActive bool,
	theme *Theme,
) (rendered string, newScrollOffset int) {
	if width <= 0 || height <= 0 {
		return "", scrollOffset
	}

	// Content area inside border
	contentWidth := width - 2
	contentHeight := height  // Will be constrained by MaxHeight in outer box

	if contentWidth <= 0 || contentHeight <= 0 {
		return "", scrollOffset
	}

	// Parse content into lines
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	visibleLines := contentHeight
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Clamp line cursor to content bounds
	if lineCursor >= totalLines {
		lineCursor = totalLines - 1
	}
	if lineCursor < 0 && totalLines > 0 {
		lineCursor = 0
	}

	// Adjust scroll to keep cursor visible
	if lineCursor >= 0 {
		if lineCursor < scrollOffset {
			scrollOffset = lineCursor
		} else if lineCursor >= scrollOffset+visibleLines {
			scrollOffset = lineCursor - visibleLines + 1
		}
	}

	// Clamp scroll offset
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	if scrollOffset > totalLines-visibleLines && totalLines > visibleLines {
		scrollOffset = totalLines - visibleLines
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	start := scrollOffset
	end := start + visibleLines
	if end > totalLines {
		end = totalLines
	}

	// Calculate widths
	var lineNumWidth int
	var codeWidth int
	if showLineNumbers {
		lineNumWidth = 4
		codeWidth = contentWidth - lineNumWidth - 1 // -1 for space separator
		if codeWidth < 1 {
			codeWidth = 1
		}
	} else {
		codeWidth = contentWidth
	}

	// Build content lines
	var contentLines []string
	lineNumberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimmedTextColor))

	for i := start; i < end; i++ {
		isCursorLine := (i == lineCursor) && isActive

		var line string

		// Add line number if enabled
		if showLineNumbers {
			lineNum := i + 1
			lineNumText := lipgloss.NewStyle().
				Width(lineNumWidth).
				Align(lipgloss.Right).
				Render(fmt.Sprintf("%d", lineNum))
			lineNumColumn := lineNumberStyle.Render(lineNumText)
			line = lineNumColumn + " "
		}

		// Style code based on cursor (lipgloss will wrap text naturally)
		code := lines[i]
		var codeStyle lipgloss.Style
		if isCursorLine {
			// Cursor line: dark foreground on teal background (menu convention)
			codeStyle = lipgloss.NewStyle().
				Width(codeWidth).
				Foreground(lipgloss.Color(theme.MainBackgroundColor)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true)
		} else {
			// Normal line
			codeStyle = lipgloss.NewStyle().
				Width(codeWidth).
				Foreground(lipgloss.Color(theme.ContentTextColor))
		}
		wrappedCode := codeStyle.Render(code)
		line += wrappedCode

		contentLines = append(contentLines, line)
	}

	// Pad remaining lines to fill contentHeight
	emptyLine := strings.Repeat(" ", contentWidth)
	for len(contentLines) < visibleLines {
		contentLines = append(contentLines, emptyLine)
	}

	// Join all lines
	contentText := strings.Join(contentLines, "\n")

	// Border color based on focus state
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	// Nested box approach:
	// Inner box: constrain content with MaxHeight (no border)
	contentBox := lipgloss.NewStyle().
		Width(width - 4).  // Account for border(2) + padding(2)
		MaxHeight(contentHeight).  // Use calculated contentHeight
		Render(contentText)
	
	// Outer box: fixed size with border and padding
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(width - 2).
		Height(height).
		Padding(0, 1)

	return boxStyle.Render(contentBox), scrollOffset
}
