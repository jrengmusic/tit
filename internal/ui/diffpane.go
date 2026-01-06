package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ========================================
// DiffPane: Reusable diff pane component
// ========================================
// Renders diff content with cursor, scrolling, and visual mode
// Data-agnostic: accepts any string diff content
// Used by: ConflictResolve

// CalculateDiffPaneHeight calculates the height of the diff pane
// given the total available height for the multi-pane layout.
// totalHeight: the height parameter passed to render functions
func CalculateDiffPaneHeight(totalHeight int) int {
	contentHeightAvailable := totalHeight - 5 // outer border (2) + blank line (1) + status bar (1) + gap (1)
	topRowHeight := contentHeightAvailable / 3
	bottomRowHeight := contentHeightAvailable - topRowHeight

	// Adjustments (same as in render functions)
	if topRowHeight > 8 {
		topRowHeight = topRowHeight - 1
		bottomRowHeight = bottomRowHeight + 1
	}
	if topRowHeight < 8 {
		topRowHeight = 8
	}
	if bottomRowHeight < 10 {
		bottomRowHeight = 10
	}

	return bottomRowHeight
}

// DiffLine represents a parsed line from diff output
type DiffLine struct {
	LineNum  int    // Line number in file (0 if removed line)
	Marker   string // "+", "-", or " " (space for context)
	Code     string // The actual code content
	LineType string // "added", "removed", or "context"
}

// DiffPaneState holds the complete state of a diff pane
type DiffPaneState struct {
	ScrollOffset     int
	LineCursor       int
	VisualModeActive bool
	VisualModeStart  int
}

// DiffPane is the reusable diff pane component
type DiffPane struct {
	// State - directly modifiable by parent (simple struct)
	ScrollOffset     int
	LineCursor       int
	VisualModeActive bool
	VisualModeStart  int

	// Configuration
	Theme *Theme
}

// NewDiffPane creates a new DiffPane component
func NewDiffPane(theme *Theme) *DiffPane {
	return &DiffPane{
		ScrollOffset:     0,
		LineCursor:       0,
		VisualModeActive: false,
		VisualModeStart:  0,
		Theme:            theme,
	}
}

// Reset clears all state for a new diff
func (dp *DiffPane) Reset() {
	dp.ScrollOffset = 0
	dp.LineCursor = 0
	dp.VisualModeActive = false
	dp.VisualModeStart = 0
}

// GetState returns a snapshot of current state
func (dp *DiffPane) GetState() DiffPaneState {
	return DiffPaneState{
		ScrollOffset:     dp.ScrollOffset,
		LineCursor:       dp.LineCursor,
		VisualModeActive: dp.VisualModeActive,
		VisualModeStart:  dp.VisualModeStart,
	}
}

// ========================================
// Parsing & Line Counting
// ========================================

// CountDisplayableLines counts lines that will be rendered (skips headers/metadata)
// Used for cursor bounds checking
func (dp *DiffPane) CountDisplayableLines(diffContent string) int {
	if diffContent == "" {
		return 0
	}

	count := 0
	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		// Skip @@ headers
		if strings.HasPrefix(line, "@@") {
			continue
		}
		// Skip file headers (---, +++)
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}
		// Skip diff metadata lines (diff --git, index, new file mode, etc.)
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "new file mode") || strings.HasPrefix(line, "deleted file mode") ||
			strings.HasPrefix(line, "old mode") || strings.HasPrefix(line, "new mode") {
			continue
		}
		// Count actual content lines
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || (line != "" && !strings.HasPrefix(line, "\\")) {
			count++
		}
	}
	return count
}

// ParseDiffLines parses diff content into structured DiffLine objects
// Skips headers and metadata, extracts line numbers and code
// Used by: rendering, copy functionality
func (dp *DiffPane) ParseDiffLines(diffContent string) []DiffLine {
	if diffContent == "" {
		return []DiffLine{}
	}

	diffLines := strings.Split(diffContent, "\n")

	var formattedLines []DiffLine
	var lineNum int

	for _, line := range diffLines {
		// Parse @@ headers to extract starting line number
		if strings.HasPrefix(line, "@@") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				newPart := parts[2]
				if strings.HasPrefix(newPart, "+") {
					numPart := strings.TrimPrefix(newPart, "+")
					if idx := strings.Index(numPart, ","); idx >= 0 {
						numPart = numPart[:idx]
					}
					if num, err := strconv.Atoi(numPart); err == nil {
						lineNum = num
					}
				}
			}
			// Skip @@ header - don't show it
			continue
		}

		// Skip file headers (---, +++)
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}

		// Skip diff metadata lines (diff --git, index, new file mode, etc.)
		if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "new file mode") || strings.HasPrefix(line, "deleted file mode") ||
			strings.HasPrefix(line, "old mode") || strings.HasPrefix(line, "new mode") {
			continue
		}

		// Parse diff content lines
		if strings.HasPrefix(line, "+") {
			// Added line
			content := line[1:] // Skip the + prefix
			formattedLines = append(formattedLines, DiffLine{
				LineNum:  lineNum,
				Marker:   "+",
				Code:     content,
				LineType: "added",
			})
			lineNum++
		} else if strings.HasPrefix(line, "-") {
			// Removed line (don't increment line number)
			content := line[1:] // Skip the - prefix
			formattedLines = append(formattedLines, DiffLine{
				LineNum:  0,
				Marker:   "-",
				Code:     content,
				LineType: "removed",
			})
		} else if line != "" {
			// Context line (unchanged)
			formattedLines = append(formattedLines, DiffLine{
				LineNum:  lineNum,
				Marker:   " ",
				Code:     line,
				LineType: "context",
			})
			lineNum++
		}
	}

	return formattedLines
}

// ========================================
// Navigation
// ========================================

// MoveCursorUp moves cursor up one line with smart scrolling
// diffContent is needed for bounds checking
func (dp *DiffPane) MoveCursorUp(diffContent string) {
	if dp.LineCursor > 0 {
		dp.LineCursor--
	}
	// Smart scroll: keep cursor in view
	if dp.LineCursor < dp.ScrollOffset {
		dp.ScrollOffset = dp.LineCursor
	}
}

// MoveCursorDown moves cursor down one line with smart scrolling and bounds checking
// diffContent is needed for bounds checking
// diffPaneHeight is full height of diff pane (including borders and status)
func (dp *DiffPane) MoveCursorDown(diffContent string, diffPaneHeight int) {
	lineCount := dp.CountDisplayableLines(diffContent)

	// Move cursor down with bounds check
	if dp.LineCursor < lineCount-1 {
		dp.LineCursor++
	}

	// Calculate visible lines (same as in Render(): height - 2 for borders only)
	visibleLines := diffPaneHeight - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Smart scroll: keep cursor in view
	// When cursor reaches the bottom of visible area, scroll down
	if dp.LineCursor >= dp.ScrollOffset+visibleLines {
		dp.ScrollOffset = dp.LineCursor - visibleLines + 1
	}
}

// ToggleVisualMode enters/exits visual selection mode
func (dp *DiffPane) ToggleVisualMode() {
	if !dp.VisualModeActive {
		// Enter visual mode: start selection at current cursor position
		dp.VisualModeActive = true
		dp.VisualModeStart = dp.LineCursor
	} else {
		// Exit visual mode
		dp.VisualModeActive = false
	}
}

// ========================================
// Selection & Copy
// ========================================

// GetSelectedLines returns the lines in current selection (code only, no metadata)
// If visual mode: returns range from VisualModeStart to LineCursor
// If not visual mode: returns current line only
func (dp *DiffPane) GetSelectedLines(diffContent string) []string {
	formattedLines := dp.ParseDiffLines(diffContent)

	// Determine selection range
	var startLine, endLine int
	if dp.VisualModeActive {
		// Copy visual selection
		startLine = dp.VisualModeStart
		endLine = dp.LineCursor
		if startLine > endLine {
			startLine, endLine = endLine, startLine
		}
	} else {
		// Copy current line only
		startLine = dp.LineCursor
		endLine = dp.LineCursor
	}

	// Collect lines - extract CODE ONLY (strip metadata)
	var linesToCopy []string
	for i := startLine; i <= endLine && i < len(formattedLines); i++ {
		if i >= 0 {
			// Copy only code content, no line numbers or markers
			linesToCopy = append(linesToCopy, formattedLines[i].Code)
		}
	}

	return linesToCopy
}

// ========================================
// Rendering
// ========================================

// Render renders the complete diff pane with borders
// Returns the rendered string (multiple lines)
func (dp *DiffPane) Render(diffContent string, isActive bool, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Border color - darker for unfocused (will use BoxBorderColor for focused later)
	borderColor := dp.Theme.ConflictPaneUnfocusedBorder // dim but visible

	// DiffPane renders WITH border, outputs exactly `height` lines
	// Status bar is rendered separately by caller
	// lipgloss Height() includes border, so actual content area = height - 2
	contentWidth := width - 2
	contentHeight := height - 2 // For visible lines calculation

	if contentWidth <= 0 || contentHeight <= 0 {
		return ""
	}

	// Build title (centered)
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(dp.Theme.ConflictPaneTitleColor)).
		Bold(true)
	titleText := "DIFF"
	styledTitle := titleStyle.Render(titleText)

	// Center title
	titleWidth := lipgloss.Width(styledTitle)
	leftPad := (contentWidth - titleWidth) / 2
	rightPad := contentWidth - titleWidth - leftPad
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}
	title := strings.Repeat(" ", leftPad) + styledTitle + strings.Repeat(" ", rightPad)

	// Build content lines (without borders)
	var contentLines []string

	// Add title as first line
	contentLines = append(contentLines, title)

	// Add blank line after title
	contentLines = append(contentLines, strings.Repeat(" ", contentWidth))

	// Parse diff lines
	formattedLines := dp.ParseDiffLines(diffContent)

	// Color styles for diff lines
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(dp.Theme.ContentTextColor))

	// Marker colors (follow diff status colors)
	addedMarkerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00"))
	removedMarkerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	contextMarkerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(dp.Theme.DimmedTextColor))

	// Line number color (muted/dim)
	lineNumberStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(dp.Theme.DimmedTextColor))

	// Calculate visible lines (contentHeight - title (1) - blank (1))
	visibleLines := contentHeight - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	start := dp.ScrollOffset
	end := start + visibleLines

	// Ensure we don't try to show beyond available lines
	if start > len(formattedLines) {
		start = len(formattedLines) - visibleLines
	}
	if start < 0 {
		start = 0
	}

	end = start + visibleLines
	if end > len(formattedLines) {
		end = len(formattedLines)
	}

	// Render diff lines with lipgloss wrapping, counting OUTPUT lines
	codeWidth := contentWidth - 6 // 4 (line num) + 2 (marker)
	if codeWidth < 1 {
		codeWidth = 1
	}

	outputLinesCount := 0
	diffLineIdx := start

	for diffLineIdx < len(formattedLines) && outputLinesCount < visibleLines {
		diffLine := formattedLines[diffLineIdx]

		// Check if this is the cursor line
		isCursorLine := (diffLineIdx == dp.LineCursor) && isActive

		// Check if this line is in visual selection
		var isInVisualSelection bool
		if dp.VisualModeActive && isActive {
			minLine := dp.VisualModeStart
			maxLine := dp.LineCursor
			if minLine > maxLine {
				minLine, maxLine = maxLine, minLine
			}
			isInVisualSelection = diffLineIdx >= minLine && diffLineIdx <= maxLine
		}

		// Column 1: Line number (4 chars, right-aligned)
		var lineNumText string
		if diffLine.LineType == "removed" || diffLine.LineNum == 0 {
			lineNumText = "    "
		} else {
			lineNumText = fmt.Sprintf("%4d", diffLine.LineNum)
		}
		lineNumColumn := lineNumberStyle.Width(4).Render(lineNumText)

		// Column 2: Marker (2 chars: marker + space)
		var markerStyle lipgloss.Style
		if diffLine.LineType == "added" {
			markerStyle = addedMarkerStyle
		} else if diffLine.LineType == "removed" {
			markerStyle = removedMarkerStyle
		} else {
			markerStyle = contextMarkerStyle
		}
		markerColumn := markerStyle.Width(2).Render(diffLine.Marker + " ")

		// Column 3: Code (lipgloss wraps automatically with Width())
		var codeStyle lipgloss.Style
		selectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(dp.Theme.MainBackgroundColor)).
			Background(lipgloss.Color(dp.Theme.MenuSelectionBackground))

		if isInVisualSelection && dp.VisualModeActive {
			codeStyle = selectionStyle
		} else if isCursorLine {
			codeStyle = selectionStyle
		} else {
			if diffLine.LineType == "added" {
				codeStyle = addedStyle
			} else if diffLine.LineType == "removed" {
				codeStyle = removedStyle
			} else {
				codeStyle = contextStyle
			}
		}
		codeColumn := codeStyle.Width(codeWidth).Render(diffLine.Code)

		// Join columns horizontally (lipgloss wraps code automatically)
		lineContent := lipgloss.JoinHorizontal(lipgloss.Top, lineNumColumn, markerColumn, codeColumn)

		// Count how many output lines this produced
		renderedLines := strings.Split(lineContent, "\n")

		// Add lines until we reach visibleLines limit
		for _, rl := range renderedLines {
			if outputLinesCount >= visibleLines {
				break
			}
			contentLines = append(contentLines, rl)
			outputLinesCount++
		}

		diffLineIdx++
	}

	// Fill remaining lines to reach visibleLines
	for outputLinesCount < visibleLines {
		contentLines = append(contentLines, strings.Repeat(" ", contentWidth))
		outputLinesCount++
	}

	// Join all content lines
	content := strings.Join(contentLines, "\n")

	// Add simple border with lipgloss
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(contentWidth).
		Height(height)

	return boxStyle.Render(content)
}

// RenderStatusBar renders the keyboard hints status bar
// Shows different hints in visual mode vs normal mode
func (dp *DiffPane) RenderStatusBar(width int) string {
	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(dp.Theme.AccentTextColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(dp.Theme.ContentTextColor))

	var statusBar string

	if dp.VisualModeActive {
		// VISUAL mode: simplified, left-aligned
		visualStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(dp.Theme.MainBackgroundColor)).
			Background(lipgloss.Color(dp.Theme.AccentTextColor)).
			Bold(true)

		parts := []string{
			visualStyle.Render("VISUAL"),
			shortcutStyle.Render("↑↓") + descStyle.Render(" select"),
			shortcutStyle.Render("Y") + descStyle.Render(" copy"),
			shortcutStyle.Render("ESC") + descStyle.Render(" back"),
		}
		statusBar = strings.Join(parts, descStyle.Render("  "))
		return statusBar
	}

	// NORMAL mode: full shortcuts, centered
	parts := []string{
		shortcutStyle.Render("↑↓") + descStyle.Render(" scroll"),
		shortcutStyle.Render("TAB") + descStyle.Render(" cycle"),
		shortcutStyle.Render("ESC") + descStyle.Render(" back"),
		shortcutStyle.Render("V") + descStyle.Render(" visual"),
		shortcutStyle.Render("Y") + descStyle.Render(" copy"),
	}

	statusBar = strings.Join(parts, descStyle.Render("  "))

	// Center the status bar
	statusWidth := lipgloss.Width(statusBar)
	if statusWidth > width {
		return statusBar
	}

	leftPad := (width - statusWidth) / 2
	rightPad := width - statusWidth - leftPad

	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}

	statusBar = strings.Repeat(" ", leftPad) + statusBar + strings.Repeat(" ", rightPad)

	return statusBar
}
