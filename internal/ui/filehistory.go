package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FileInfo is an alias for git.FileInfo to avoid import cycles in UI
// This is the same type as git.FileInfo but defined in UI package
type FileInfo struct {
	Path   string // File path
	Status string // M, A, D, R, etc.
}

// FileHistoryPane represents which pane is focused in file(s) history mode
type FileHistoryPane int

const (
	PaneCommits FileHistoryPane = iota
	PaneFiles
	PaneDiff
)

// FileHistoryState represents the state of the file(s) history browser
type FileHistoryState struct {
	Commits           []CommitInfo  // List of recent commits
	Files             []FileInfo    // Files in selected commit
	SelectedCommitIdx int          // Currently selected commit (0-indexed)
	SelectedFileIdx   int          // Currently selected file (0-indexed)
	FocusedPane       FileHistoryPane // Which pane has focus
	CommitsScrollOff  int          // Scroll offset for commits list
	FilesScrollOff    int          // Scroll offset for files list
	DiffScrollOff     int          // Scroll offset for diff pane
}

// RenderFileHistorySplitPane renders the file(s) history split-pane view (3 columns side-by-side)
// Returns content exactly `width` chars wide and `height - 2` lines tall (for outer border)
// Parameters:
//   - state: interface{} (FileHistoryState from app package)
//   - theme: Theme (colors, styles)
//   - width, height: Terminal dimensions
//
// Returns: String representation of the rendered pane
func RenderFileHistorySplitPane(state interface{}, theme Theme, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Type assert to FileHistoryState
	fileHistoryState, ok := state.(*FileHistoryState)
	if !ok || fileHistoryState == nil {
		return "Error: invalid file history state"
	}

	// Calculate pane height based on desired visible items
	// We want ~15 visible commits in the list
	// List structure: border(2) + title(1) + separator(1) + items(15) = 19 lines
	desiredVisibleItems := 15
	paneHeight := desiredVisibleItems + 4 // +4 for title, separator, and borders

	// Calculate column widths based on CONTENT NEEDS
	// Divide width equally among 3 panes
	commitPaneWidth := (width / 3)
	filesPaneWidth := (width / 3)
	diffPaneWidth := width - commitPaneWidth - filesPaneWidth // Remaining width for diff

	// Ensure minimum widths
	if commitPaneWidth < 20 {
		commitPaneWidth = 20
	}
	if filesPaneWidth < 20 {
		filesPaneWidth = 20
	}
	if diffPaneWidth < 30 {
		diffPaneWidth = 30
	}

	// Render all three columns at same height
	commitsPaneContent := renderFileHistoryCommitsPane(fileHistoryState, theme, commitPaneWidth, paneHeight)
	filesPaneContent := renderFileHistoryFilesPane(fileHistoryState, theme, filesPaneWidth, paneHeight)
	diffPaneContent := renderFileHistoryDiffPane(fileHistoryState, theme, diffPaneWidth, paneHeight)

	// Join columns horizontally (side-by-side)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, commitsPaneContent, filesPaneContent, diffPaneContent)

	// Build status bar
	statusBar := buildFileHistoryStatusBar(fileHistoryState.FocusedPane, width, theme)

	// Stack: mainRow + statusBar
	// Total height will be (height - 3) + 1 = height - 2 (correct for outer wrapper)
	return mainRow + "\n" + statusBar
}

// renderFileHistoryCommitsPane renders the commits list pane (left)
func renderFileHistoryCommitsPane(state *FileHistoryState, theme Theme, width, height int) string {
	// Create list pane for commits
	listPane := NewListPane("Commits", &theme)

	// Build list items from actual commits
	var items []ListItem
	for i, commit := range state.Commits {
		attributeText := commit.Time.Format("02-Jan 15:04")
		// Show first 7 chars of hash
		hashShort := commit.Hash
		if len(hashShort) > 7 {
			hashShort = hashShort[:7]
		}

		items = append(items, ListItem{
			AttributeText:  attributeText,
			AttributeColor: theme.DimmedTextColor,
			ContentText:    hashShort,
			ContentColor:   theme.ContentTextColor,
			ContentBold:    false,
			IsSelected:     i == state.SelectedCommitIdx,
		})
	}

	// Calculate visible lines for scrolling (height - border(2))
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected commit visible
	listPane.AdjustScroll(state.SelectedCommitIdx, visibleLines)

	// Render list pane (active when commits pane is focused)
	// Pass 0, 3 for column positioning (first of 3 columns)
	return listPane.Render(items, width, height, state.FocusedPane == PaneCommits, 0, 3)
}

// renderFileHistoryFilesPane renders the files list pane (middle)
func renderFileHistoryFilesPane(state *FileHistoryState, theme Theme, width, height int) string {
	// Create list pane for files
	listPane := NewListPane("Files", &theme)

	// Build list items from files in selected commit
	var items []ListItem
	for i, file := range state.Files {
		// Show status indicator and filename
		statusIndicator := " "
		if file.Status == "M" {
			statusIndicator = "✓" // Modified
		} else if file.Status == "A" {
			statusIndicator = "+" // Added
		} else if file.Status == "D" {
			statusIndicator = "-" // Deleted
		} else if file.Status == "R" {
			statusIndicator = "→" // Renamed
		}

		items = append(items, ListItem{
			AttributeText:  statusIndicator,
			AttributeColor: theme.DimmedTextColor,
			ContentText:    file.Path,
			ContentColor:   theme.ContentTextColor,
			ContentBold:    false,
			IsSelected:     i == state.SelectedFileIdx,
		})
	}

	// Calculate visible lines for scrolling (height - border(2))
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected file visible
	listPane.AdjustScroll(state.SelectedFileIdx, visibleLines)

	// Render list pane (active when files pane is focused)
	// Pass 1, 3 for column positioning (second of 3 columns)
	return listPane.Render(items, width, height, state.FocusedPane == PaneFiles, 1, 3)
}

// renderFileHistoryDiffPane renders the diff pane (right)
func renderFileHistoryDiffPane(state *FileHistoryState, theme Theme, width, height int) string {
	// For now, use a placeholder diff content
	// In Phase 6, this will be populated from the diff cache
	diffContent := "@@ -1,10 +1,12 @@\n package main\n\n func main() {\n+  fmt.Println(\"Hello, World!\")\n }\n"

	// Create diff pane
	diffPane := NewDiffPane(&theme)
	diffPane.ScrollOffset = state.DiffScrollOff

	// Render diff pane (active when diff pane is focused)
	isActive := state.FocusedPane == PaneDiff
	return diffPane.Render(diffContent, isActive, width, height)
}

// buildFileHistoryStatusBar builds the status bar with keyboard shortcuts
func buildFileHistoryStatusBar(focusedPane FileHistoryPane, width int, theme Theme) string {
	shortcutStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.AccentTextColor)).
		Bold(true)
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.ContentTextColor))
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.DimmedTextColor))

	// Build shortcuts based on focused pane
	var parts []string
	switch focusedPane {
	case PaneCommits:
		parts = []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" navigate commits"),
		}
	case PaneFiles:
		parts = []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" navigate files"),
		}
	case PaneDiff:
		parts = []string{
			shortcutStyle.Render("↑↓") + descStyle.Render(" scroll diff"),
		}
	}

	parts = append(parts,
		shortcutStyle.Render("TAB") + descStyle.Render(" cycle panes"),
		shortcutStyle.Render("Y") + descStyle.Render(" copy"),
		shortcutStyle.Render("V") + descStyle.Render(" visual"),
		shortcutStyle.Render("ESC") + descStyle.Render(" back"),
	)

	statusText := strings.Join(parts, sepStyle.Render("  │  "))

	// Center and size status bar
	statusStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center)

	return statusStyle.Render(statusText)
}