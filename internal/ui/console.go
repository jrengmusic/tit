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

// ScrollToBottom positions viewport to show the last line at the bottom
func (s *ConsoleOutState) ScrollToBottom(totalOutputLines int) {
	// Set scroll offset so last lines are visible
	// maxScroll = total output lines - content height
	// We don't have contentHeight here, so set a large number and renderer will clamp
	s.ScrollOffset = max(0, totalOutputLines-1)
}

// ScrollUp moves the viewport up by one line
func (s *ConsoleOutState) ScrollUp() {
	s.ScrollOffset--
}

// ScrollDown moves the viewport down by one line
func (s *ConsoleOutState) ScrollDown() {
	s.ScrollOffset++
}

// RenderConsoleOutput renders the console output panel with scrolling
// Pattern: pre-size content exactly, pad to dimensions, border wraps pre-sized content
// Returns exactly maxWidth x height output (matches TextInput pattern)
func RenderConsoleOutput(
	state *ConsoleOutState,
	buffer *OutputBuffer,
	palette Theme,
	maxWidth int,
	totalHeight int,
	operationInProgress bool,
	abortConfirmActive bool,
	autoScroll bool,
) string {
	// SSOT: Console structure (no inner box border, outer border from RenderLayout)
	// maxWidth = 76 (ContentInnerWidth)
	// totalHeight = ContentHeight (26)
	// 
	// Structure pre-outer-border:
	//   title (1) + blank (1) + content (?) + blank (1) + status (1) = totalHeight - 2 (for outer border)
	//   
	// So content height = (totalHeight - 2) - 4 = totalHeight - 6
	
	contentLines := totalHeight - 6  // Actual output lines visible
	wrapWidth := maxWidth           // No inner box, use full width
	
	state.LinesPerPage = contentLines

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

	// Step 1: BUILD ALL OUTPUT LINES (after wrapping at wrapWidth)
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

	// Step 2: Calculate scroll bounds
	totalOutputLines := len(allOutputLines)
	maxScroll := totalOutputLines - contentLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	
	// Store maxScroll in state
	state.MaxScroll = maxScroll

	// Auto-scroll: force ScrollOffset to maxScroll
	if autoScroll {
		state.ScrollOffset = state.MaxScroll
	}
	
	// Always clamp to valid range
	if state.ScrollOffset < 0 {
		state.ScrollOffset = 0
	}
	if state.ScrollOffset > state.MaxScroll {
		state.ScrollOffset = state.MaxScroll
	}

	// Step 3: Extract visible window
	start := state.ScrollOffset
	end := start + contentLines
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

	// Pad to exactly contentLines
	emptyLine := strings.Repeat(" ", wrapWidth)
	for len(visibleLines) < contentLines {
		visibleLines = append(visibleLines, emptyLine)
	}

	// Content without inner box border - just joined lines
	// (outer Content border from RenderLayout is sufficient)
	contentBox := strings.Join(visibleLines, "\n")

	// Build title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.OutputInfoColor)).
		Bold(true)

	titleText := fmt.Sprintf("OUTPUT [offset=%d max=%d auto=%v lines=%d]", 
		state.ScrollOffset, state.MaxScroll, autoScroll, totalOutputLines)
	title := titleStyle.Render(titleText)
	titleWidth := lipgloss.Width(title)
	if titleWidth < maxWidth {
		title = title + strings.Repeat(" ", maxWidth-titleWidth)
	}

	// Build status bar
	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.AccentTextColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.LabelTextColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette.DimmedTextColor))

	atBottom := state.ScrollOffset >= maxScroll
	remainingLines := totalOutputLines - (state.ScrollOffset + contentLines)
	if remainingLines < 0 {
		remainingLines = 0
	}

	var statusLeft string
	if abortConfirmActive {
		parts := []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" scroll"),
			shortcutStyle.Render("ESC") + descStyle.Render(" back to menu"),
		}
		statusLeft = strings.Join(parts, sepStyle.Render("  │  "))
	} else if operationInProgress {
		parts := []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" scroll"),
			shortcutStyle.Render("ESC") + descStyle.Render(" abort"),
		}
		statusLeft = strings.Join(parts, sepStyle.Render("  │  "))
	} else {
		parts := []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" scroll"),
			shortcutStyle.Render("ESC") + descStyle.Render(" back to menu"),
		}
		statusLeft = strings.Join(parts, sepStyle.Render("  │  "))
	}

	var statusRight string
	if atBottom {
		statusRight = descStyle.Render("(at bottom)")
	} else if remainingLines > 0 {
		statusRight = sepStyle.Render("↓ ") + descStyle.Render(fmt.Sprintf("%d more lines", remainingLines))
	} else {
		statusRight = descStyle.Render("(can scroll up)")
	}

	statusLeftWidth := lipgloss.Width(statusLeft)
	statusRightWidth := lipgloss.Width(statusRight)
	statusPadding := maxWidth - statusLeftWidth - statusRightWidth
	if statusPadding < 0 {
		statusPadding = 0
	}
	statusBar := statusLeft + strings.Repeat(" ", statusPadding) + statusRight

	// Pad title and status to maxWidth
	if lipgloss.Width(statusBar) < maxWidth {
		statusBar = statusBar + strings.Repeat(" ", maxWidth-lipgloss.Width(statusBar))
	}

	// Build blank line
	blankLine := strings.Repeat(" ", maxWidth)

	// Combine: title + blank + contentBox + blank + status
	panel := lipgloss.JoinVertical(lipgloss.Left,
		title,
		blankLine,
		contentBox,
		blankLine,
		statusBar,
	)

	// Pre-size panel to exact dimensions (CRITICAL - matching TextInput pattern)
	panelLines := strings.Split(panel, "\n")

	// Pad each line to maxWidth
	for i := range panelLines {
		lineWidth := lipgloss.Width(panelLines[i])
		if lineWidth < maxWidth {
			panelLines[i] = panelLines[i] + strings.Repeat(" ", maxWidth-lineWidth)
		}
	}

	// Pad to exact height (totalHeight - 2 for outer border)
	panelHeight := totalHeight - 2
	panelBlankLine := strings.Repeat(" ", maxWidth)
	for len(panelLines) < panelHeight {
		panelLines = append(panelLines, panelBlankLine)
	}

	if len(panelLines) > panelHeight {
		panelLines = panelLines[:panelHeight]
	}

	// Return pre-sized panel WITHOUT outer border
	// RenderLayout will add the outer Content border
	return strings.Join(panelLines, "\n")
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
