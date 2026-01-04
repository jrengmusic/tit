package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderBranchInputs renders two branch name inputs (canon + working) stacked within Content height
// activeField determines which field is highlighted (border color changes)
// Returns text that fits exactly within ContentHeight-2 bounds
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
	// Layout: title (1) + padding (1) + canon (4) + padding (1) + working (4) + fill
	maxWidth := ContentInnerWidth
	totalHeight := ContentHeight - 2

	// Centered title
	titleLine := Line{
		Content: StyledContent{
			Text:    "SET BRANCH",
			FgColor: theme.LabelTextColor,
			Bold:    true,
		},
		Alignment: "center",
		Width:     maxWidth,
	}

	// Determine border colors based on which field is active
	canonBorderColor := theme.BoxBorderColor
	workingBorderColor := theme.BoxBorderColor
	if activeField == "canon" {
		canonBorderColor = theme.LabelTextColor
	}
	if activeField == "working" {
		workingBorderColor = theme.LabelTextColor
	}

	// Render both input fields
	canonInput := RenderInputField(
		InputFieldState{
			Label:       canonPrompt,
			Value:       canonValue,
			CursorPos:   canonCursorPos,
			IsActive:    activeField == "canon",
			BorderColor: canonBorderColor,
		},
		maxWidth,
		4, // 4 lines: label (1) + box (3)
		theme,
	)

	workingInput := RenderInputField(
		InputFieldState{
			Label:       workingPrompt,
			Value:       workingValue,
			CursorPos:   workingCursorPos,
			IsActive:    activeField == "working",
			BorderColor: workingBorderColor,
		},
		maxWidth,
		4,
		theme,
	)

	// Stack: title + spacer + canon + spacer + working + fill
	spacer := strings.Repeat(" ", maxWidth)
	combined := lipgloss.JoinVertical(
		lipgloss.Left,
		titleLine.Render(),
		spacer,
		canonInput,
		spacer,
		workingInput,
	)

	// Ensure exact dimensions
	return EnsureExactDimensions(combined, maxWidth, totalHeight)
}
