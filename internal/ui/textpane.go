package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a parsed diff line (for diff mode rendering)
type DiffLine struct {
	LineNum  int    // Line number in file (0 if removed)
	Marker   string // "+", "-", or " "
	Code     string // The code content
	LineType string // "added", "removed", or "context"
}

// parseDiffContent parses diff output into structured DiffLine objects for 3-column rendering
func parseDiffContent(diffContent string) []DiffLine {
	if diffContent == "" {
		return []DiffLine{}
	}

	lines := strings.Split(diffContent, "\n")
	var result []DiffLine
	var lineNum int

	for _, line := range lines {
		// Parse @@ headers
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
			continue
		}

		// Skip metadata
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") ||
			strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "new file mode") || strings.HasPrefix(line, "deleted file mode") ||
			strings.HasPrefix(line, "old mode") || strings.HasPrefix(line, "new mode") {
			continue
		}

		// Parse content lines
		if strings.HasPrefix(line, "+") {
			result = append(result, DiffLine{
				LineNum: lineNum, Marker: "+", Code: line[1:], LineType: "added",
			})
			lineNum++
		} else if strings.HasPrefix(line, "-") {
			result = append(result, DiffLine{
				LineNum: 0, Marker: "-", Code: line[1:], LineType: "removed",
			})
		} else if line != "" {
			result = append(result, DiffLine{
				LineNum: lineNum, Marker: " ", Code: line, LineType: "context",
			})
			lineNum++
		}
	}

	return result
}

// RenderTextPane renders scrollable text in a fixed-size box
// If isDiff is true, parses diff content and renders 3-column layout (line# + marker + code)
// Otherwise renders plain text with optional line numbers
// Supports visual mode selection in diff mode (v + y)
func RenderTextPane(
	content string,
	width int,
	height int,
	lineCursor int,
	scrollOffset int,
	showLineNumbers bool,
	isActive bool,
	isDiff bool,
	theme *Theme,
	visualModeActive bool,
	visualModeStart int,
) (rendered string, newScrollOffset int) {

	var lines []string
	var diffLines []DiffLine
	var totalLines int
	var isDiffMode bool

	// Parse content based on mode
	if isDiff {
		// Diff mode: parse into structured format for later rendering
		diffLines = parseDiffContent(content)
		totalLines = len(diffLines)
		isDiffMode = true
		showLineNumbers = false // Diff already has line numbers in 3-column format
	} else {
		// Plain text mode
		lines = strings.Split(content, "\n")
		totalLines = len(lines)
		isDiffMode = false
	}

	if totalLines == 0 {
		return renderEmptyPane(width, height, isActive, theme), 0
	}

	// Available space
	interiorHeight := height - 2
	contentWidth := width - 4

	// Clamp cursor
	if lineCursor < 0 {
		lineCursor = 0
	}
	if lineCursor >= totalLines {
		lineCursor = totalLines - 1
	}

	// Scroll window calculation
	var scrollWindow int

	if isDiffMode {
		// Diff mode: no line wrapping, each line = 1 physical line
		// Allow cursor to move beyond visible window before scrolling
		scrollWindow = interiorHeight + 2
	} else {
		// Plain text: conservative margin for line wrapping
		scrollWindow = interiorHeight - 4
	}

	if scrollWindow < 1 {
		scrollWindow = 1
	}

	// Don't scroll if all lines fit
	if totalLines <= scrollWindow {
		scrollOffset = 0
	} else {
		// Scroll to keep cursor in window
		if lineCursor < scrollOffset {
			scrollOffset = lineCursor
		}
		if lineCursor >= scrollOffset+scrollWindow {
			scrollOffset = lineCursor - scrollWindow + 1
		}

		// Clamp scroll
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		maxScroll := totalLines - scrollWindow
		if maxScroll < 0 {
			maxScroll = 0
		}
		if scrollOffset > maxScroll {
			scrollOffset = maxScroll
		}
	}

	// Calculate widths
	lineNumWidth := 0
	textWidth := contentWidth
	if showLineNumbers {
		lineNumWidth = 4
		textWidth = contentWidth - lineNumWidth - 1
	}

	// Render all lines from scrollOffset - MaxHeight will clip
	var renderedLines []string

	for i := scrollOffset; i < totalLines; i++ {
		var line string
		isCursor := (i == lineCursor) && isActive

		if isDiffMode {
			// Diff mode: render 3-column format (line# + marker + code)
			dl := diffLines[i]

			// Column 1: Line number (4 chars, dimmed)
			var lineNumText string
			if dl.LineType == "removed" {
				lineNumText = "    "
			} else {
				lineNumText = fmt.Sprintf("%4d", dl.LineNum)
			}

			// Column 2: Marker (2 chars, colored by diff type)
			markerText := dl.Marker + " "

			// Determine color for code and marker
			var codeColor string
			switch dl.LineType {
			case "added":
				codeColor = theme.DiffAddedLineColor
			case "removed":
				codeColor = theme.DiffRemovedLineColor
			default:
				codeColor = theme.ContentTextColor
			}

			// Check if line is in visual selection
			var isInVisualSelection bool
			if visualModeActive && isActive {
				minLine := visualModeStart
				maxLine := lineCursor
				if minLine > maxLine {
					minLine, maxLine = maxLine, minLine
				}
				isInVisualSelection = i >= minLine && i <= maxLine
			}

			// Column 1: Line number (always dimmed)
			lineNumCol := lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.DimmedTextColor)).
				Width(4).
				Render(lineNumText)

			// Column 2: Marker (colored by diff type)
			markerCol := lipgloss.NewStyle().
				Foreground(lipgloss.Color(codeColor)).
				Width(2).
				Render(markerText)

			// Column 3: Code (with cursor or visual selection)
			var codeCol string
			if isInVisualSelection && visualModeActive {
				// Visual selection highlight
				codeCol = lipgloss.NewStyle().
					Width(contentWidth - 6). // -6 for lineNum(4) + marker(2)
					Foreground(lipgloss.Color(theme.MainBackgroundColor)).
					Background(lipgloss.Color(theme.MenuSelectionBackground)).
					Render(dl.Code)
			} else if isCursor {
				// Cursor highlight (bold)
				codeCol = lipgloss.NewStyle().
					Width(contentWidth - 6). // -6 for lineNum(4) + marker(2)
					Foreground(lipgloss.Color(theme.MainBackgroundColor)).
					Background(lipgloss.Color(theme.MenuSelectionBackground)).
					Bold(true).
					Render(dl.Code)
			} else {
				codeCol = lipgloss.NewStyle().
					Foreground(lipgloss.Color(codeColor)).
					Render(dl.Code)
			}

			line = lineNumCol + markerCol + codeCol
		} else {
			// Plain text mode
			line = ""

			// Line number
			if showLineNumbers {
				num := ""
				if i >= 0 && i < totalLines {
					num = fmt.Sprintf("%d", i+1)
				}
				lineNumCol := lipgloss.NewStyle().
					Width(lineNumWidth).
					Align(lipgloss.Right).
					Foreground(lipgloss.Color(theme.DimmedTextColor)).
					Render(num)
				line = lineNumCol + " "
			}

			// Text - apply width to all lines
			text := lines[i]

			if isCursor {
				line += lipgloss.NewStyle().
					Width(textWidth).
					Foreground(lipgloss.Color(theme.MainBackgroundColor)).
					Background(lipgloss.Color(theme.MenuSelectionBackground)).
					Bold(true).
					Render(text)
			} else {
				line += lipgloss.NewStyle().
					Width(textWidth).
					Foreground(lipgloss.Color(theme.ContentTextColor)).
					Render(text)
			}
		}

		renderedLines = append(renderedLines, line)
	}

	contentText := strings.Join(renderedLines, "\n")

	// Border color
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	// Nested box pattern from Session 52:
	// Inner box MaxHeight(height) - expands fully
	// Outer box Height(height) + Border + Padding - naturally trims
	innerBox := lipgloss.NewStyle().
		Width(contentWidth).
		MaxHeight(height).
		Render(contentText)

	outerBox := lipgloss.NewStyle().
		Width(width-2).
		Height(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render(innerBox)

	return outerBox, scrollOffset
}

func renderEmptyPane(width, height int, isActive bool, theme *Theme) string {
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	return lipgloss.NewStyle().
		Width(width-2).
		Height(height).
		MaxHeight(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render("")
}

// GetSelectedLinesFromDiff returns the lines in visual selection from diff content
// Used by copy/yank functionality
func GetSelectedLinesFromDiff(diffContent string, visualModeStart int, visualModeEnd int) []string {
	diffLines := parseDiffContent(diffContent)
	if len(diffLines) == 0 {
		return []string{}
	}

	// Normalize start/end
	minLine := visualModeStart
	maxLine := visualModeEnd
	if minLine > maxLine {
		minLine, maxLine = maxLine, minLine
	}

	// Bounds check
	if minLine < 0 {
		minLine = 0
	}
	if maxLine >= len(diffLines) {
		maxLine = len(diffLines) - 1
	}

	var selectedLines []string
	for i := minLine; i <= maxLine && i < len(diffLines); i++ {
		dl := diffLines[i]
		// Build line with marker + code (no line number for copying)
		selectedLines = append(selectedLines, dl.Marker+dl.Code)
	}

	return selectedLines
}
