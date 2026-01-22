package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConsoleOutState holds the scrolling state for console output
type ConsoleOutState struct {
	ScrollOffset int
	LinesPerPage int
	MaxScroll    int // Cached max scroll position
}

// NewConsoleOutState creates a new console output state with default values
func NewConsoleOutState() ConsoleOutState {
	return ConsoleOutState{
		ScrollOffset: 0,
		LinesPerPage: 18, // Default for content area
	}
}

// Reset resets the scroll state
func (s *ConsoleOutState) Reset() {
	s.ScrollOffset = 0
	s.MaxScroll = 0
}

// ScrollUp moves the viewport up by one line
func (s *ConsoleOutState) ScrollUp() {
	if s.ScrollOffset > 0 {
		s.ScrollOffset--
	}
}

// ScrollDown moves the viewport down by one line
func (s *ConsoleOutState) ScrollDown() {
	if s.ScrollOffset < s.MaxScroll {
		s.ScrollOffset++
	}
}

// RenderConsoleOutputFullScreen renders console output for full-screen mode (no header/footer)
// Takes terminal dimensions directly, returns content that occupies full terminal
// Pattern matches RenderHistorySplitPane: content + status bar at bottom
func RenderConsoleOutputFullScreen(
	state *ConsoleOutState,
	buffer *OutputBuffer,
	palette Theme,
	termWidth int,
	termHeight int,
	operationInProgress bool,
	abortConfirmActive bool,
	autoScroll bool,
	statusBarOverride string,
) string {
	if termWidth <= 0 || termHeight <= 0 {
		return ""
	}

	// Calculate console height: reserve 1 line for status bar + 1 for newline separator
	// Plus outer border means we need height - 2 for the content area
	consoleHeight := termHeight - 2

	// Content lines available (title + blank + content + blank = 4 lines used)
	// Status bar takes 1 line, so content gets the rest
	titleHeight := 2
	statusHeight := 1
	contentHeight := consoleHeight - titleHeight - statusHeight

	if contentHeight < 1 {
		contentHeight = 1
	}

	wrapWidth := termWidth

	state.LinesPerPage = contentHeight

	// Color mapping function (semantic colors from new theme)
	getColor := func(lineType OutputLineType) string {
		switch lineType {
		case TypeStdout:
			return palette.OutputStdoutColor
		case TypeStderr:
			return palette.OutputStderrColor
		case TypeCommand:
			return palette.OutputStdoutColor
		case TypeStatus:
			return palette.OutputStatusColor
		case TypeWarning:
			return palette.OutputWarningColor
		case TypeDebug:
			return palette.OutputDebugColor
		case TypeInfo:
			return palette.OutputInfoColor
		default:
			return palette.OutputStdoutColor
		}
	}

	totalBufferLines := buffer.GetLineCount()

	var allOutputLines []string

	if totalBufferLines == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.DimmedTextColor)).
			Italic(true)
		allOutputLines = append(allOutputLines, emptyStyle.Render("(no output yet)"))
	} else {
		displayLines := buffer.GetLines(0, totalBufferLines)
		for _, line := range displayLines {
			lineStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(getColor(line.Type)))

			formatted := fmt.Sprintf("[%s] %s", line.Time, line.Text)
			renderedLine := lineStyle.Width(wrapWidth).Render(formatted)
			renderedLines := strings.Split(renderedLine, "\n")
			allOutputLines = append(allOutputLines, renderedLines...)
		}
	}

	// Calculate scroll bounds
	totalOutputLines := len(allOutputLines)
	maxScroll := totalOutputLines - contentHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	state.MaxScroll = maxScroll

	// Auto-scroll: if enabled, stay at bottom
	if autoScroll {
		state.ScrollOffset = maxScroll
	} else {
		// Manual scroll: clamp to valid range
		if state.ScrollOffset > maxScroll {
			state.ScrollOffset = maxScroll
		}
		if state.ScrollOffset < 0 {
			state.ScrollOffset = 0
		}
	}

	scrollOffset := int(state.ScrollOffset)

	// Extract visible window
	start := scrollOffset
	end := start + contentHeight
	if start < 0 {
		start = 0
	}
	if end > totalOutputLines {
		end = totalOutputLines
	}

	var visibleLines []string
	for i := start; i < end; i++ {
		visibleLines = append(visibleLines, allOutputLines[i])
	}

	// Pad each visible line to wrapWidth
	for i := range visibleLines {
		lineWidth := lipgloss.Width(visibleLines[i])
		if lineWidth < wrapWidth {
			visibleLines[i] = visibleLines[i] + strings.Repeat(" ", wrapWidth-lineWidth)
		}
	}

	// Pad to exactly contentHeight
	emptyLine := strings.Repeat(" ", wrapWidth)
	for len(visibleLines) < contentHeight {
		visibleLines = append(visibleLines, emptyLine)
	}

	// Content without inner box border
	contentBox := strings.Join(visibleLines, "\n")

	// Build title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.OutputInfoColor)).
		Bold(true)

	titleText := "OUTPUT"
	title := titleStyle.Render(titleText)
	titleWidth := lipgloss.Width(title)
	if titleWidth < wrapWidth {
		title = title + strings.Repeat(" ", wrapWidth-titleWidth)
	}

	// Build blank line
	blankLine := strings.Repeat(" ", wrapWidth)

	// Combine: title + blank + contentBox
	panel := lipgloss.JoinVertical(lipgloss.Left,
		title,
		blankLine,
		contentBox,
	)

	// Pad panel to exact height (consoleHeight - statusHeight for status bar)
	panelLines := strings.Split(panel, "\n")
	for len(panelLines) < consoleHeight-statusHeight {
		panelLines = append(panelLines, blankLine)
	}
	if len(panelLines) > consoleHeight-statusHeight {
		panelLines = panelLines[:consoleHeight-statusHeight]
	}
	panel = strings.Join(panelLines, "\n")

	// Build centered status bar at bottom
	statusBar := buildConsoleStatusBar(termWidth, palette, operationInProgress, abortConfirmActive, statusBarOverride)

	// Return: panel + newline + statusBar (total height = consoleHeight + 1 for newline = termHeight - 1)
	// Actually, we want exactly termHeight - 1 for outer border, so:
	// panel (consoleHeight - 1) + "\n" + statusBar (1 line) = termHeight lines total
	return panel + "\n" + statusBar
}

// buildConsoleStatusBar builds a centered status bar for console output
func buildConsoleStatusBar(width int, palette Theme, operationInProgress bool, abortConfirmActive bool, overrideMessage string) string {
	styles := NewStatusBarStyles(&palette)

	// If override message is set, use it
	if overrideMessage != "" {
		return BuildStatusBar(StatusBarConfig{
			Width:           width,
			Centered:        true,
			Theme:           &palette,
			OverrideMessage: overrideMessage,
		})
	}

	// Build shortcuts based on state
	var parts []string
	if abortConfirmActive {
		parts = []string{
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" scroll"),
			styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" back to menu"),
		}
	} else if operationInProgress {
		parts = []string{
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" scroll"),
			styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" abort"),
		}
	} else {
		parts = []string{
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" scroll"),
			styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" back to menu"),
		}
	}

	// Build scroll indicator (at bottom or more lines)
	// This requires access to scroll state, which we don't have here
	// For now, just return the shortcuts centered
	return BuildStatusBar(StatusBarConfig{
		Parts:    parts,
		Width:    width,
		Centered: true,
		Theme:    &palette,
	})
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
