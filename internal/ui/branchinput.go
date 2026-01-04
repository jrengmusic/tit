package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBranchInputs renders two branch name inputs (canon + working) stacked within Content height
// Fits within ContentHeight and ContentInnerWidth constraints
// Returns text that fits exactly within Content box
// activeField determines which field has cursor and highlight
func RenderBranchInputs(
	canonPrompt string,
	canonValue string,
	canonCursorPos int,
	workingPrompt string,
	workingValue string,
	workingCursorPos int,
	activeField string, // "canon" or "working" - which field is currently active
	theme Theme,
) string {
	// Available space: ContentHeight (24) - border overhead (2)
	// Each input: prompt (1) + blank (1) + input box (min 3)
	// Total: 2 inputs * (1 + 1 + 3) = 10 lines, leaves 14 lines free for spacing
	
	maxWidth := ContentInnerWidth
	totalHeight := ContentHeight - 2

	// Determine border colors based on which field is active
	canonBorderColor := theme.BorderSecondaryColor
	workingBorderColor := theme.BorderSecondaryColor
	if activeField == "canon" {
		canonBorderColor = theme.SecondaryTextColor // Highlight active field
	}
	if activeField == "working" {
		workingBorderColor = theme.SecondaryTextColor
	}

	// Render canon branch input with height
	// Only show caret if this is the active field
	canonInput := renderCompactTextInput(
		canonPrompt,
		canonValue,
		canonCursorPos,
		canonBorderColor,
		theme,
		maxWidth,
		5, // 5 lines: prompt (1) + blank (1) + box with 3 lines content
		activeField == "canon", // Show caret only if active
	)

	// Render working branch input with height
	// Only show caret if this is the active field
	workingInput := renderCompactTextInput(
		workingPrompt,
		workingValue,
		workingCursorPos,
		workingBorderColor,
		theme,
		maxWidth,
		5,
		activeField == "working", // Show caret only if active
	)

	// Stack vertically with spacing
	// Canon (3) + spacing (3) + Working (3) + padding to fill ContentHeight
	spacer := strings.Repeat(" ", maxWidth)

	combined := lipgloss.JoinVertical(
		lipgloss.Left,
		canonInput,
		spacer,
		spacer,
		spacer,
		workingInput,
	)

	// Ensure exact dimensions
	lines := strings.Split(combined, "\n")

	// Pad each line to maxWidth
	for i := range lines {
		lineWidth := lipgloss.Width(lines[i])
		if lineWidth < maxWidth {
			lines[i] = lines[i] + strings.Repeat(" ", maxWidth-lineWidth)
		}
	}

	// Pad to exactly totalHeight
	for len(lines) < totalHeight {
		lines = append(lines, strings.Repeat(" ", maxWidth))
	}

	// Truncate if too long
	if len(lines) > totalHeight {
		lines = lines[:totalHeight]
	}

	return strings.Join(lines, "\n")
}

// renderCompactTextInput renders a single text input field
// Used by RenderBranchInputs for two-input layout
// Only shows caret if showCaret=true (for active field)
func renderCompactTextInput(
	prompt string,
	value string,
	cursorPos int,
	borderColor string,
	theme Theme,
	maxWidth int,
	totalHeight int,
	showCaret bool, // Only draw caret in active field
) string {
	// Label style
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.SecondaryTextColor)).
		Bold(true)

	styledLabel := labelStyle.Render(prompt)
	labelWidth := lipgloss.Width(styledLabel)
	label := styledLabel
	if labelWidth < maxWidth {
		label = styledLabel + strings.Repeat(" ", maxWidth-labelWidth)
	}

	// Validate cursor position
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(value) {
		cursorPos = len(value)
	}

	// Build display text - only show caret if active
	displayText := value
	if showCaret {
		displayText = value[:cursorPos] + "█" + value[cursorPos:]
	}

	// Wrap to fit (account for border 2 chars)
	wrapWidth := maxWidth - 2
	if wrapWidth < 1 {
		wrapWidth = 1
	}

	// Wrap text for display
	wrappedText := lipgloss.NewStyle().Width(wrapWidth).Render(displayText)
	allLines := strings.Split(wrappedText, "\n")

	// For box display: use first line only (compact single-line input)
	// Structure: label (1) + blank (1) + box with 1 line = 3 lines total
	// If we want 5 lines total, we need 3 lines in the box (with borders = 5 total)
	
	boxContentHeight := totalHeight - 4 // totalHeight - (label + blank + border top/bottom)
	if boxContentHeight < 1 {
		boxContentHeight = 1
	}

	// Collect visible lines (with scrolling to keep caret visible if needed)
	var visibleLines []string
	if showCaret {
		// Find caret line
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
		lineWidth := lipgloss.Width(visibleLines[i])
		if lineWidth < wrapWidth {
			visibleLines[i] = visibleLines[i] + strings.Repeat(" ", wrapWidth-lineWidth)
		}
	}

	// Pad to fill boxContentHeight
	for len(visibleLines) < boxContentHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", wrapWidth))
	}

	boxContent := strings.Join(visibleLines, "\n")

	// Inner box
	boxStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Foreground(lipgloss.Color(theme.PrimaryTextColor))

	box := boxStyle.Render(boxContent)

	// Combine: label (1) + blank (1) + box
	blankLine := strings.Repeat(" ", maxWidth)
	combined := lipgloss.JoinVertical(lipgloss.Left, label, blankLine, box)

	// Ensure exact totalHeight and width
	lines := strings.Split(combined, "\n")

	for i := range lines {
		lineWidth := lipgloss.Width(lines[i])
		if lineWidth < maxWidth {
			lines[i] = lines[i] + strings.Repeat(" ", maxWidth-lineWidth)
		}
	}

	// Pad or truncate to exact totalHeight
	for len(lines) < totalHeight {
		lines = append(lines, strings.Repeat(" ", maxWidth))
	}
	if len(lines) > totalHeight {
		lines = lines[:totalHeight]
	}

	return strings.Join(lines, "\n")
}
