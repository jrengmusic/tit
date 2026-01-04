package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TextInputState holds state for text input component
type TextInputState struct {
	Value                 string
	CursorPos             int  // Cursor position (byte index in Value)
	ShowClearConfirmation bool
	Height                int // Number of lines (1 for single-line, 16 for multi-line, etc.)
}

// RenderTextInput renders a text input component with label and inner bordered viewport
// Creates an inner box for text editing that fits within Content bounds
// Returns exactly totalHeight lines
func RenderTextInput(
	prompt string,
	state TextInputState,
	theme Theme,
	maxWidth int,
	totalHeight int,
) string {
	// Label style - pad to maxWidth
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryTextColor)).
		Bold(true)

	// Render label and pad to full width
	styledLabel := labelStyle.Render(prompt)
	labelWidth := lipgloss.Width(styledLabel)
	label := styledLabel
	if labelWidth < maxWidth {
		label = styledLabel + strings.Repeat(" ", maxWidth-labelWidth)
	}

	// Border color based on confirmation state
	borderColor := theme.BorderSecondaryColor
	if state.ShowClearConfirmation {
		borderColor = theme.SecondaryTextColor // Error-like color
	}

	// Calculate box height from totalHeight
	// Structure: label (1) + blank (1) + box (totalHeight - 2)
	// Box content height: box - border (2) = totalHeight - 4
	boxContentHeight := totalHeight - 4
	if boxContentHeight < 1 {
		boxContentHeight = 1
	}

	// Insert caret at cursor position
	if state.CursorPos < 0 {
		state.CursorPos = 0
	}
	if state.CursorPos > len(state.Value) {
		state.CursorPos = len(state.Value)
	}
	textWithCaret := state.Value[:state.CursorPos] + "█" + state.Value[state.CursorPos:]

	// Wrap width accounts for inner box border
	// maxWidth = 76, inner box has border (2), so wrap to 74
	wrapWidth := maxWidth - 2
	if wrapWidth < 1 {
		wrapWidth = 1
	}

	// Let lipgloss wrap the text
	wrappedText := lipgloss.NewStyle().Width(wrapWidth).Render(textWithCaret)
	allLines := strings.Split(wrappedText, "\n")

	// Find which output line contains the caret (█ symbol)
	caretLineIndex := 0
	for lineIdx, line := range allLines {
		if strings.Contains(line, "█") {
			caretLineIndex = lineIdx
			break
		}
	}

	// Calculate scroll offset to keep caret visible
	// If caretLineIndex >= boxContentHeight, scroll so caret is at bottom
	scrollOffset := 0
	if caretLineIndex >= boxContentHeight {
		scrollOffset = caretLineIndex - boxContentHeight + 1
	}

	// Constrain to boxContentHeight (take lines from scrollOffset)
	var visibleLines []string
	for i := scrollOffset; i < scrollOffset+boxContentHeight && i < len(allLines); i++ {
		visibleLines = append(visibleLines, allLines[i])
	}

	// Pad each line to exact width and pad to fill height
	for i := range visibleLines {
		lineWidth := lipgloss.Width(visibleLines[i])
		if lineWidth < wrapWidth {
			visibleLines[i] = visibleLines[i] + strings.Repeat(" ", wrapWidth-lineWidth)
		}
	}
	for len(visibleLines) < boxContentHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", wrapWidth))
	}

	constrainedText := strings.Join(visibleLines, "\n")

	// Inner bordered box - just adds border, NO width/height constraints
	// Content is already sized correctly
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Foreground(lipgloss.Color(theme.PrimaryTextColor))

	box := boxStyle.Render(constrainedText)

	// Combine label (1) + blank (1) + box
	blankLine := strings.Repeat(" ", maxWidth)
	combined := lipgloss.JoinVertical(lipgloss.Left, label, blankLine, box)

	// Ensure exactly totalHeight lines and exact width
	combinedLines := strings.Split(combined, "\n")

	// Pad each line to maxWidth
	for i := range combinedLines {
		lineWidth := lipgloss.Width(combinedLines[i])
		if lineWidth < maxWidth {
			combinedLines[i] = combinedLines[i] + strings.Repeat(" ", maxWidth-lineWidth)
		}
	}

	// Pad to totalHeight
	for len(combinedLines) < totalHeight {
		combinedLines = append(combinedLines, strings.Repeat(" ", maxWidth))
	}

	// Truncate if somehow too long
	if len(combinedLines) > totalHeight {
		combinedLines = combinedLines[:totalHeight]
	}

	return strings.Join(combinedLines, "\n")
}

// GetInputBoxHeight returns the height of the rendered text input component
// height: number of input lines (1 for single-line, 16 for multi-line, etc.)
func GetInputBoxHeight(height int) int {
	if height < 1 {
		height = 1
	}
	// Label (1 line) + blank line (1) + content (height lines)
	// Border is added by RenderContent wrapper
	return 1 + 1 + height
}
