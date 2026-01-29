package ui

import (
	"strings"
	"tit/internal"

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
	SyncInProgress   bool // True when timeline sync is running
	SyncFrame        int  // Animation frame for spinner
}

// TimelineSyncSpinner returns spinner frame based on animation frame
func TimelineSyncSpinner(frame int) string {
	spinnerFrames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinnerFrames[frame%len(spinnerFrames)]
}

// TimelineSyncLabel returns timeline label with spinner when syncing
func TimelineSyncLabel(baseEmoji, baseLabel string, syncInProgress bool, frame int) string {
	if syncInProgress {
		return TimelineSyncSpinner(frame) + " Syncing..."
	}
	return baseEmoji + " " + baseLabel
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

	// WorkingTree + Version (2-column layout like upper section)
	wtLabel := TimelineSyncLabel(state.WorkingTreeEmoji, state.WorkingTreeLabel, state.SyncInProgress, state.SyncFrame)
	wtLabelStyled := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(state.WorkingTreeColor)).
		Width(leftWidth).
		Render(wtLabel)

	versionText := internal.AppVersion
	versionStyled := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.DimmedTextColor)).
		Align(lipgloss.Right).
		Width(rightWidth).
		Render(versionText)

	wtVersionLine := lipgloss.JoinHorizontal(lipgloss.Top, wtLabelStyled, versionStyled)
	fullWidthLines = append(fullWidthLines, wtVersionLine)

	// WorkingTree descriptions (indented)
	indent := strings.Repeat(" ", EmojiColumnWidth)
	if state.SyncInProgress {
		// Show clear sync message instead of stale/confusing description
		descLine := lipgloss.NewStyle().
			Width(totalWidth).
			Foreground(lipgloss.Color(theme.DimmedTextColor)).
			Render(indent + "Checking local state...")
		fullWidthLines = append(fullWidthLines, descLine)
	} else {
		// Show actual working tree descriptions
		for _, desc := range state.WorkingTreeDesc {
			descLine := lipgloss.NewStyle().
				Width(totalWidth).
				Foreground(lipgloss.Color(theme.ContentTextColor)).
				Render(indent + desc)
			fullWidthLines = append(fullWidthLines, descLine)
		}
	}

	// Timeline label (with spinner when syncing)
	tlLabel := TimelineSyncLabel(state.TimelineEmoji, state.TimelineLabel, state.SyncInProgress, state.SyncFrame)
	tlLabelLine := lipgloss.NewStyle().
		Width(totalWidth).
		Bold(true).
		Foreground(lipgloss.Color(state.TimelineColor)).
		Render(tlLabel)
	fullWidthLines = append(fullWidthLines, tlLabelLine)

	// Timeline descriptions (indented) - show sync message or actual descriptions
	if state.SyncInProgress {
		// Show clear sync message instead of stale/confusing description
		descLine := lipgloss.NewStyle().
			Width(totalWidth).
			Foreground(lipgloss.Color(theme.DimmedTextColor)).
			Render(indent + "Fetching remote updates...")
		fullWidthLines = append(fullWidthLines, descLine)
	} else {
		// Show actual timeline descriptions
		for _, desc := range state.TimelineDesc {
			descLine := lipgloss.NewStyle().
				Width(totalWidth).
				Foreground(lipgloss.Color(theme.ContentTextColor)).
				Render(indent + desc)
			fullWidthLines = append(fullWidthLines, descLine)
		}
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
