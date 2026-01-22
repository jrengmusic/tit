package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PadLineToWidth pads a line to exact width with spaces
// Note: Only adds padding, doesn't truncate (to preserve styled text)
func PadLineToWidth(line string, width int) string {
	lineWidth := lipgloss.Width(line)
	if lineWidth < width {
		return line + strings.Repeat(" ", width-lineWidth)
	}
	return line
}

// RightAlignLine right-aligns content within width
func RightAlignLine(content string, width int) string {
	contentWidth := lipgloss.Width(content)
	if contentWidth >= width {
		return content
	}
	padding := width - contentWidth
	return strings.Repeat(" ", padding) + content
}

// CenterAlignLine centers content within width
func CenterAlignLine(content string, width int) string {
	contentWidth := lipgloss.Width(content)
	if contentWidth >= width {
		return content
	}
	leftPadding := (width - contentWidth) / 2
	rightPadding := width - contentWidth - leftPadding
	return strings.Repeat(" ", leftPadding) + content + strings.Repeat(" ", rightPadding)
}

// PadAllLinesToWidth ensures all lines are exactly width characters
func PadAllLinesToWidth(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = PadLineToWidth(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

// PadTextToHeight pads text to exact height with empty lines
func PadTextToHeight(text string, height int) string {
	lines := strings.Split(text, "\n")

	// Pad with empty lines to fill height
	for len(lines) < height {
		lines = append(lines, "")
	}

	// Truncate if too many lines
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// EnsureExactDimensions ensures text is exactly width × height
func EnsureExactDimensions(text string, width int, height int) string {
	// First pad to height
	padded := PadTextToHeight(text, height)

	// Then pad each line to width
	return PadAllLinesToWidth(padded, width)
}

// ShortenHash returns first 7 characters of a git hash for display
func ShortenHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
