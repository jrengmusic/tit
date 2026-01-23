package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const EmojiColumnWidth = 3

type HeaderState struct {
	CurrentDirectory string
	RemoteURL        string
	RemoteColor      string
	OperationEmoji   string
	OperationLabel   string
	OperationColor   string
	BranchEmoji      string
	BranchLabel      string
	BranchColor      string
	WorkingTreeEmoji string
	WorkingTreeLabel string
	WorkingTreeDesc  []string
	WorkingTreeColor string
	TimelineEmoji    string
	TimelineLabel    string
	TimelineDesc     []string
	TimelineColor    string
}

// RenderHeaderInfo renders header info section (9 content rows + 2 padding = 11 lines)
// 2-column section (80/20): CWD + Remote (left) | Operation + Branch (right)
// Full-width section: Separator, WorkingTree, Timeline
func RenderHeaderInfo(sizing DynamicSizing, theme Theme, state HeaderState) string {
	totalWidth := sizing.HeaderInnerWidth
	leftWidth := int(float64(totalWidth) * 0.8)
	rightWidth := totalWidth - leftWidth

	// === 2-COLUMN SECTION ===
	// LEFT COLUMN (80%)
	var leftLines []string

	// Row 1: CWD
	cwdLine := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.LabelTextColor)).
		Render("📁 " + state.CurrentDirectory)
	leftLines = append(leftLines, cwdLine)

	// Row 2: Remote URL
	remoteLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color(state.RemoteColor)).
		Render(state.RemoteURL)
	leftLines = append(leftLines, remoteLine)

	leftColumn := lipgloss.NewStyle().
		Width(leftWidth).
		Render(strings.Join(leftLines, "\n"))

	// RIGHT COLUMN (20%)
	var rightLines []string

	// Row 1: Operation status
	opLine := lipgloss.NewStyle().
		Width(rightWidth).
		Bold(true).
		Foreground(lipgloss.Color(state.OperationColor)).
		Align(lipgloss.Right).
		Render(state.OperationEmoji + " " + state.OperationLabel)
	rightLines = append(rightLines, opLine)

	// Row 2: Branch
	branchLine := lipgloss.NewStyle().
		Width(rightWidth).
		Bold(true).
		Foreground(lipgloss.Color(state.BranchColor)).
		Align(lipgloss.Right).
		Render(state.BranchEmoji + " " + state.BranchLabel)
	rightLines = append(rightLines, branchLine)

	rightColumn := lipgloss.NewStyle().
		AlignVertical(lipgloss.Top).
		Render(strings.Join(rightLines, "\n"))

	// Join 2-column section
	twoColumnSection := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// === FULL-WIDTH SECTION ===
	var fullWidthLines []string

	// Separator
	separatorLine := lipgloss.NewStyle().
		Width(totalWidth).
		Foreground(lipgloss.Color(theme.SeparatorColor)).
		Render(strings.Repeat("─", totalWidth))
	fullWidthLines = append(fullWidthLines, separatorLine)

	// WorkingTree label
	wtLabelLine := lipgloss.NewStyle().
		Width(totalWidth).
		Bold(true).
		Foreground(lipgloss.Color(state.WorkingTreeColor)).
		Render(state.WorkingTreeEmoji + " " + state.WorkingTreeLabel)
	fullWidthLines = append(fullWidthLines, wtLabelLine)

	// WorkingTree descriptions (indented)
	indent := strings.Repeat(" ", EmojiColumnWidth)
	for _, desc := range state.WorkingTreeDesc {
		descLine := lipgloss.NewStyle().
			Width(totalWidth).
			Foreground(lipgloss.Color(theme.ContentTextColor)).
			Render(indent + desc)
		fullWidthLines = append(fullWidthLines, descLine)
	}

	// Timeline label
	tlLabelLine := lipgloss.NewStyle().
		Width(totalWidth).
		Bold(true).
		Foreground(lipgloss.Color(state.TimelineColor)).
		Render(state.TimelineEmoji + " " + state.TimelineLabel)
	fullWidthLines = append(fullWidthLines, tlLabelLine)

	// Timeline descriptions (indented)
	for _, desc := range state.TimelineDesc {
		descLine := lipgloss.NewStyle().
			Width(totalWidth).
			Foreground(lipgloss.Color(theme.ContentTextColor)).
			Render(indent + desc)
		fullWidthLines = append(fullWidthLines, descLine)
	}

	fullWidthSection := strings.Join(fullWidthLines, "\n")

	// Combine sections vertically
	result := lipgloss.JoinVertical(lipgloss.Left, twoColumnSection, fullWidthSection)

	return result
}

// RenderHeader renders header with margins
func RenderHeader(sizing DynamicSizing, theme Theme, info string) string {
	marginStyle := lipgloss.NewStyle().
		PaddingLeft(HorizontalMargin).
		PaddingRight(HorizontalMargin)

	infoStyled := lipgloss.NewStyle().
		Width(sizing.HeaderInnerWidth).
		AlignVertical(lipgloss.Top).
		Render(info)

	return marginStyle.Render(infoStyled)
}
