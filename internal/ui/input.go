package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// InputFieldState holds the state of an input field
type InputFieldState struct {
	Label      string // Field label (e.g., "Canon:", "Working:")
	Value      string // Current input value
	CursorPos  int    // Cursor position in value
	IsActive   bool   // Whether field is active (shows caret)
	BorderColor string // Color for field border
}

// RenderInputField renders a single input field with label and bordered box
// Returns text that's exactly maxWidth × totalHeight lines
// Typical totalHeight = 4: label (1) + box with 3 lines content (3, including borders)
func RenderInputField(field InputFieldState, maxWidth int, totalHeight int, theme Theme) string {
	// Label line with padding
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.LabelTextColor)).
		Bold(true)

	styledLabel := labelStyle.Render(field.Label)
	labelLine := PadLineToWidth(styledLabel, maxWidth)

	// Calculate content area dimensions
	// totalHeight = label (1) + box content (totalHeight - 1, which includes borders)
	boxContentHeight := totalHeight - 3 // totalHeight - (label + border top/bottom)
	if boxContentHeight < 1 {
		boxContentHeight = 1
	}

	// Validate cursor position
	cursorPos := field.CursorPos
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(field.Value) {
		cursorPos = len(field.Value)
	}

	// Build display text - only show caret if active
	displayText := field.Value
	if field.IsActive {
		displayText = field.Value[:cursorPos] + "█" + field.Value[cursorPos:]
	}

	// Wrap to fit (account for border 2 chars + inner padding 2)
	innerPadding := 1
	wrapWidth := maxWidth - 2 - (innerPadding * 2)
	if wrapWidth < 1 {
		wrapWidth = 1
	}

	// Wrap text for display
	wrappedText := lipgloss.NewStyle().Width(wrapWidth).Render(displayText)
	allLines := strings.Split(wrappedText, "\n")

	// Collect visible lines (with scrolling to keep caret visible if active)
	var visibleLines []string
	if field.IsActive {
		// Find caret line and scroll to keep it visible
		caretLineIndex := 0
		for lineIdx, line := range allLines {
			if strings.Contains(line, "█") {
				caretLineIndex = lineIdx
				break
			}
		}
		// Scroll to keep caret visible
		scrollOffset := 0
		if caretLineIndex >= boxContentHeight {
			scrollOffset = caretLineIndex - boxContentHeight + 1
		}
		for i := scrollOffset; i < scrollOffset+boxContentHeight && i < len(allLines); i++ {
			visibleLines = append(visibleLines, allLines[i])
		}
	} else {
		// Inactive field: show first lines
		for i := 0; i < boxContentHeight && i < len(allLines); i++ {
			visibleLines = append(visibleLines, allLines[i])
		}
	}

	// Pad each line to width
	for i := range visibleLines {
		visibleLines[i] = PadLineToWidth(visibleLines[i], wrapWidth)
	}

	// Pad to fill boxContentHeight
	for len(visibleLines) < boxContentHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", wrapWidth))
	}

	// Add inner padding to each line
	for i := range visibleLines {
		visibleLines[i] = strings.Repeat(" ", innerPadding) + visibleLines[i] + strings.Repeat(" ", innerPadding)
	}

	boxContent := strings.Join(visibleLines, "\n")

	// Render box
	box := RenderBox(BoxConfig{
		Content:     boxContent,
		InnerWidth:  maxWidth - 2,
		InnerHeight: boxContentHeight,
		BorderColor: field.BorderColor,
		TextColor:   theme.ContentTextColor,
		Theme:       theme,
	})

	// Combine label + box
	combined := lipgloss.JoinVertical(lipgloss.Left, labelLine, box)

	// Ensure exact dimensions
	return EnsureExactDimensions(combined, maxWidth, totalHeight)
}
