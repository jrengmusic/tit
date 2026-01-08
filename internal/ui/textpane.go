package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DiffLine represents a parsed diff line with structured format
type DiffLine struct {
	LineNum  int    // Line number in file (0 if removed)
	Marker   string // "+", "-", or " "
	Code     string // The code content
	LineType string // "added", "removed", or "context"
}

// parseDiffLines parses diff output into structured DiffLine objects
func parseDiffLines(diffContent string) []DiffLine {
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

// RenderDiffPane renders diff with 3-column layout (line# + marker + code)
func RenderDiffPane(
	diffContent string,
	width int,
	height int,
	lineCursor int,
	scrollOffset int,
	isActive bool,
	theme *Theme,
) (rendered string, newScrollOffset int) {

	if width <= 0 || height <= 0 {
		return "", 0
	}

	// Border color
	borderColor := theme.ConflictPaneUnfocusedBorder
	if isActive {
		borderColor = theme.ConflictPaneFocusedBorder
	}

	contentWidth := width - 2
	contentHeight := height - 2

	if contentWidth <= 0 || contentHeight <= 0 {
		return "", 0
	}

	// Parse diff lines
	diffLines := parseDiffLines(diffContent)

	// Color styles
	addedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DiffAddedLineColor))
	removedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DiffRemovedLineColor))
	contextStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.ContentTextColor))
	lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.DimmedTextColor))
	selectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.MainBackgroundColor)).
		Background(lipgloss.Color(theme.MenuSelectionBackground))

	// Clamp cursor
	if lineCursor < 0 {
		lineCursor = 0
	}
	if lineCursor >= len(diffLines) {
		lineCursor = len(diffLines) - 1
	}

	// Scroll window
	scrollWindow := contentHeight - 4
	if scrollWindow < 1 {
		scrollWindow = 1
	}

	if len(diffLines) <= scrollWindow {
		scrollOffset = 0
	} else {
		if lineCursor < scrollOffset {
			scrollOffset = lineCursor
		}
		if lineCursor >= scrollOffset+scrollWindow {
			scrollOffset = lineCursor - scrollWindow + 1
		}
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		maxScroll := len(diffLines) - scrollWindow
		if maxScroll < 0 {
			maxScroll = 0
		}
		if scrollOffset > maxScroll {
			scrollOffset = maxScroll
		}
	}

	// Column widths: 4 (linenum) + 2 (marker) + remaining (code)
	codeWidth := contentWidth - 6
	if codeWidth < 1 {
		codeWidth = 1
	}

	// Render lines
	var contentLines []string

	start := scrollOffset
	end := start + scrollWindow
	if end > len(diffLines) {
		end = len(diffLines)
	}

	for i := start; i < end; i++ {
		dl := diffLines[i]
		isCursor := (i == lineCursor) && isActive

		// Column 1: Line number (4 chars)
		var lineNumText string
		if dl.LineType == "removed" {
			lineNumText = "    "
		} else {
			lineNumText = fmt.Sprintf("%4d", dl.LineNum)
		}
		lineNumCol := lineNumStyle.Width(4).Render(lineNumText)

		// Column 2: Marker (2 chars)
		markerCol := lineNumStyle.Width(2).Render(dl.Marker + " ")

		// Column 3: Code
		var codeStyle lipgloss.Style
		if isCursor {
			codeStyle = selectionStyle
		} else {
			switch dl.LineType {
			case "added":
				codeStyle = addedStyle
			case "removed":
				codeStyle = removedStyle
			default:
				codeStyle = contextStyle
			}
		}
		codeCol := codeStyle.Width(codeWidth).Render(dl.Code)

		// Join columns
		line := lipgloss.JoinHorizontal(lipgloss.Top, lineNumCol, markerCol, codeCol)
		contentLines = append(contentLines, line)
	}

	// Pad to height
	for len(contentLines) < scrollWindow {
		contentLines = append(contentLines, strings.Repeat(" ", contentWidth))
	}

	content := strings.Join(contentLines, "\n")

	// Apply border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(contentWidth).
		Height(height)

	return boxStyle.Render(content), scrollOffset
}

// RenderTextPane renders scrollable text in a fixed-size box
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
) (rendered string, newScrollOffset int) {

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

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

	// Scroll window with 4-line margin (interiorHeight is physical lines)
	scrollWindow := interiorHeight - 4
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
		line := ""

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
		isCursor := (i == lineCursor) && isActive

		if isCursor {
			line += lipgloss.NewStyle().
				Width(textWidth).
				Foreground(lipgloss.Color(theme.MainBackgroundColor)).
				Background(lipgloss.Color(theme.MenuSelectionBackground)).
				Bold(true).
				Render(text)
		} else {
			// Diff mode: color +/- lines from theme
			var textColor string
			if isDiff {
				if strings.HasPrefix(text, "+") {
					textColor = theme.DiffAddedLineColor // Green for added
				} else if strings.HasPrefix(text, "-") {
					textColor = theme.DiffRemovedLineColor // Red for removed
				} else {
					textColor = theme.ContentTextColor
				}
			} else {
				textColor = theme.ContentTextColor
			}

			line += lipgloss.NewStyle().
				Width(textWidth).
				Foreground(lipgloss.Color(textColor)).
				Render(text)
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
		Width(width - 2).
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
		Width(width - 2).
		Height(height).
		MaxHeight(height).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1).
		Render("")
}
