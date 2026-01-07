package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderTextPane renders scrollable text in a fixed-size box
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

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	if totalLines == 0 {
		return renderEmptyPane(width, height, isActive, theme), 0
	}

	// Available space
	interiorHeight := height - 2
	contentWidth := width - 4

	// Clamp cursor
	if lineCursor < 0 {
		lineCursor = 0
	}
	if lineCursor >= totalLines {
		lineCursor = totalLines - 1
	}

	// Scroll window with 4-line margin (interiorHeight is physical lines)
	scrollWindow := interiorHeight - 4
	if scrollWindow < 1 {
		scrollWindow = 1
	}

	// Don't scroll if all lines fit
	if totalLines <= scrollWindow {
		scrollOffset = 0
	} else {
		// Scroll to keep cursor in window
		if lineCursor < scrollOffset {
			scrollOffset = lineCursor
		}
		if lineCursor >= scrollOffset+scrollWindow {
			scrollOffset = lineCursor - scrollWindow + 1
		}

		// Clamp scroll
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		maxScroll := totalLines - scrollWindow
		if maxScroll < 0 {
			maxScroll = 0
		}
		if scrollOffset > maxScroll {
			scrollOffset = maxScroll
		}
	}

	// Calculate widths
	lineNumWidth := 0
	textWidth := contentWidth
	if showLineNumbers {
		lineNumWidth = 4
		textWidth = contentWidth - lineNumWidth - 1
	}

	// Render all lines from scrollOffset - MaxHeight will clip
	var renderedLines []string

	for i := scrollOffset; i < totalLines; i++ {
		line := ""

		// Line number
		if showLineNumbers {
			num := ""
			if i >= 0 && i < totalLines {
				num = fmt.Sprintf("%d", i+1)
			}
			lineNumCol := lipgloss.NewStyle().
				Width(lineNumWidth).
				Align(lipgloss.Right).
				Foreground(lipgloss.Color(theme.DimmedTextColor)).
				Render(num)
			line = lineNumCol + " "
		}

		// Text - apply width to all lines
		text := lines[i]
		isCursor := (i == lineCursor) && isActive

		if isCursor {
			line += lipgloss.NewStyle().
				Width(textWidth).
				Foreground(lipgloss.Color(theme.MainBackgroundColor)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true).
				Render(text)
		} else {
			line += lipgloss.NewStyle().
				Width(textWidth).
				Foreground(lipgloss.Color(theme.ContentTextColor)).
				Render(text)
		}

		renderedLines = append(renderedLines, line)
	}

	contentText := strings.Join(renderedLines, "\n")

	// Border color
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	// Nested box pattern from Session 52:
	// Inner box MaxHeight(height) - expands fully
	// Outer box Height(height) + Border + Padding - naturally trims
	innerBox := lipgloss.NewStyle().
		Width(contentWidth).
		MaxHeight(height).
		Render(contentText)

	outerBox := lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render(innerBox)

	return outerBox, scrollOffset
}

func renderEmptyPane(width, height int, isActive bool, theme *Theme) string {
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	return lipgloss.NewStyle().
		Width(width - 2).
		Height(height).
		MaxHeight(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render("")
}
