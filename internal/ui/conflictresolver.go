package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ========================================
// Conflict Resolution Data Structures
// ========================================

// ConflictFile represents a file in conflict (2-column version)
type ConflictFile struct {
	Path      string
	KeepLocal bool // true=keep LOCAL, false=keep REMOTE
}

// ConflictFileGeneric represents a file in conflict with N-way choice
type ConflictFileGeneric struct {
	Path     string
	Versions []string // Content for each column
	Chosen   int      // Which column is chosen (0-based)
}

// ========================================
// N-Column Generic Parallel View
// ========================================

// RenderConflictResolveGeneric renders the generic N-column parallel view
// Top row: N columns showing file lists with checkboxes
// Bottom row: N columns showing content for selected file
// Returns content exactly `width` chars wide and `height - 1` lines tall (footer handled externally)
func RenderConflictResolveGeneric(
	files []ConflictFileGeneric,
	selectedFileIndex int,
	focusedPane int, // 0...numCols-1 = top columns, numCols...numCols*2-1 = bottom columns
	numColumns int,
	columnLabels []string,
	scrollOffsets []int,
	lineCursors []int,
	width int,
	height int,
	theme Theme,
) string {
	if width <= 0 || height <= 0 || numColumns == 0 {
		return ""
	}

	// Layout: topRow + bottomRow = height - 1 (footer takes 1 line)
	totalPaneHeight := height
	topRowHeight := totalPaneHeight / 3
	bottomRowHeight := totalPaneHeight - topRowHeight - 5

	// Adjust: add 2 to top row, reduce from bottom row
	topRowHeight += 2
	bottomRowHeight -= 2

	// Calculate column widths - no gaps, borders touch
	// For 2 columns: width=76, each gets 38 chars (including their own borders)
	baseColumnWidth := width / numColumns
	remainder := width % numColumns

	// Render top row: N file list columns using ListPane
	var topRowLines []string
	for col := 0; col < numColumns; col++ {
		isActive := (focusedPane == col)
		label := ""
		if col < len(columnLabels) {
			label = columnLabels[col]
			// Colorize hash in title (e.g., "COMMIT abc1234")
			label = colorizeIncomingPaneTitle(label, &theme)
		}

		// Calculate column width: base + 1 if we have remainder
		columnWidth := baseColumnWidth
		if col >= numColumns-remainder {
			columnWidth++
		}

		// Use ListPane for file list
		listPane := NewListPane(label, &theme)
		listItems := convertFilesToListItems(files, selectedFileIndex, col, &theme)

		visibleLines := topRowHeight - 4
		if visibleLines < 1 {
			visibleLines = 1
		}

		// Adjust scroll to keep selected file visible
		listPane.AdjustScroll(selectedFileIndex, visibleLines)

		paneRendered := listPane.Render(listItems, columnWidth, topRowHeight, isActive, col, numColumns)
		topRowLines = append(topRowLines, paneRendered)
	}

	// Join top row columns - lipgloss will place them side-by-side with borders touching
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, topRowLines...)

	// Render bottom row: N content columns
	var bottomRowLines []string
	for col := 0; col < numColumns; col++ {
		isActive := (focusedPane == numColumns+col)
		scrollOffset := 0
		if col < len(scrollOffsets) {
			scrollOffset = scrollOffsets[col]
		}
		lineCursor := 0
		if col < len(lineCursors) {
			lineCursor = lineCursors[col]
		}
		content := ""
		if selectedFileIndex >= 0 && selectedFileIndex < len(files) {
			if col < len(files[selectedFileIndex].Versions) {
				content = files[selectedFileIndex].Versions[col]
			}
		}

		// Calculate column width: base + 1 if we have remainder
		columnWidth := baseColumnWidth
		if col >= numColumns-remainder {
			columnWidth++
		}

		// Render content column with cursor using SSOT TextPane
		// No visual mode in conflict resolver
		paneRendered, newScrollOffset := RenderTextPane(content, columnWidth, bottomRowHeight, lineCursor, scrollOffset, true, isActive, false, &theme, false, 0)

		// Update scroll offset in array
		if col < len(scrollOffsets) {
			scrollOffsets[col] = newScrollOffset
		}

		bottomRowLines = append(bottomRowLines, paneRendered)
	}

	// Join bottom row columns - lipgloss will place them side-by-side with borders touching
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, bottomRowLines...)

	// Stack: topRow + bottomRow (footer handled externally)
	return topRow + "\n" + bottomRow
}

// ========================================
// Helper Functions
// ========================================

// convertFilesToListItems converts conflict files to ListItem format for ListPane
func convertFilesToListItems(files []ConflictFileGeneric, selectedFileIndex int, columnIndex int, theme *Theme) []ListItem {
	var items []ListItem
	for i, file := range files {
		// Checkbox as attribute
		checkbox := "[ ]"
		checkboxColor := theme.DimmedTextColor
		if file.Chosen == columnIndex {
			checkbox = "[✓]"
			checkboxColor = theme.AccentTextColor
		}

		// File path as content
		items = append(items, ListItem{
			AttributeText:  checkbox,
			AttributeColor: checkboxColor,
			ContentText:    file.Path,
			ContentColor:   theme.ContentTextColor,
			ContentBold:    false,
			IsSelected:     (i == selectedFileIndex),
		})
	}
	return items
}

// colorizeIncomingPaneTitle colors the hash portion of "COMMIT ABC1234" using AccentTextColor
func colorizeIncomingPaneTitle(title string, theme *Theme) string {
	// Look for pattern: "COMMIT <hash>"
	parts := strings.Fields(title)
	if len(parts) >= 2 && parts[0] == "COMMIT" {
		prefix := parts[0]
		hash := parts[1]

		// Style hash with AccentTextColor and bold
		hashStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.AccentTextColor)).
			Bold(true)

		return prefix + " " + hashStyle.Render(hash)
	}

	// If pattern doesn't match, return as-is
	return title
}
