package ui

import (
	"strconv"
	"strings"
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
