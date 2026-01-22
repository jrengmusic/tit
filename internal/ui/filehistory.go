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
	Commits           []CommitInfo    // List of recent commits
	Files             []FileInfo      // Files in selected commit
	SelectedCommitIdx int             // Currently selected commit (0-indexed)
	SelectedFileIdx   int             // Currently selected file (0-indexed)
	FocusedPane       FileHistoryPane // Which pane has focus
	CommitsScrollOff  int             // Scroll offset for commits list
	FilesScrollOff    int             // Scroll offset for files list
	DiffScrollOff     int             // Scroll offset for diff pane
	DiffLineCursor    int             // Line cursor for diff pane (for TextPane)
	DiffContent       string          // Current diff content (populated by handlers on file/commit selection)
	VisualModeActive  bool            // True when visual mode is active (for selecting lines)
	VisualModeStart   int             // Starting line of visual selection
}

// RenderFileHistorySplitPane renders the file(s) history split-pane view (3-pane layout)
// Layout: Top row (Commits + Files side-by-side), Bottom row (Diff full-width)
// Returns content exactly `width` chars wide and `height - 2` lines tall (for outer border)
// Parameters:
//   - state: interface{} (FileHistoryState from app package)
//   - theme: Theme (colors, styles)
//   - width, height: Terminal dimensions
//
// Returns: String representation of the rendered pane
func RenderFileHistorySplitPane(state interface{}, theme Theme, width, height int, statusBarOverride string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Type assert to FileHistoryState
	fileHistoryState, ok := state.(*FileHistoryState)
	if !ok || fileHistoryState == nil {
		return "Error: invalid file history state"
	}

	// Calculate dimensions from terminal height
	// Reserve 1 line for status bar + 1 for newline separator
	totalPaneHeight := height - 5
	topRowHeight := totalPaneHeight / 3
	bottomRowHeight := totalPaneHeight - topRowHeight

	// Calculate column widths for top row (2 columns: Commits + Files)
	// No gaps, borders touch directly
	commitPaneWidth := 24                     // Fixed width for commits (same as History mode)
	filesPaneWidth := width - commitPaneWidth // Remaining width for files

	// Render top row panes (Commits + Files)
	commitsPaneContent := renderFileHistoryCommitsPane(fileHistoryState, theme, commitPaneWidth, topRowHeight)
	filesPaneContent := renderFileHistoryFilesPane(fileHistoryState, theme, filesPaneWidth, topRowHeight)

	// Join top row columns - lipgloss will place them side-by-side with borders touching
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, commitsPaneContent, filesPaneContent)

	// Render bottom row (Diff pane - full width, single column)
	bottomRow := renderFileHistoryDiffPane(fileHistoryState, theme, width, bottomRowHeight)

	 
	var statusBar string
	if fileHistoryState.FocusedPane == PaneDiff {
		statusBar = buildDiffStatusBar(fileHistoryState.VisualModeActive, width, theme, statusBarOverride)
	} else {
		statusBar = buildFileHistoryStatusBar(fileHistoryState.FocusedPane, width, theme, statusBarOverride)
	}

	// Stack everything with no gaps: topRow + bottomRow + statusBar
	return topRow + "\n" + bottomRow + "\n" + statusBar
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

	 
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected commit visible
	listPane.AdjustScroll(state.SelectedCommitIdx, visibleLines)

	// Render list pane (active when commits pane is focused)
	// Pass 0, 1 for column params (not used in 2-row layout)
	return listPane.Render(items, width, height, state.FocusedPane == PaneCommits, 0, 1)
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

	 
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected file visible
	listPane.AdjustScroll(state.SelectedFileIdx, visibleLines)

	// Render list pane (active when files pane is focused)
	// Pass 0, 1 for column params (not used in 2-row layout)
	return listPane.Render(items, width, height, state.FocusedPane == PaneFiles, 0, 1)
}

// renderFileHistoryDiffPane renders the diff pane with 3-column layout (line# + marker + code)
func renderFileHistoryDiffPane(state *FileHistoryState, theme Theme, width, height int) string {
	// Get diff content from state (populated by handlers on file/commit selection)
	// If no diff yet, show placeholder
	diffContent := state.DiffContent
	if diffContent == "" {
		diffContent = "(no diff available)"
	}

	// Use RenderTextPane with isDiff=true for proper 3-column diff layout
	// This reuses the proven scrolling and height calculations from Session 52
	isActive := state.FocusedPane == PaneDiff

	rendered, newScrollOffset := RenderTextPane(
		diffContent,
		width,
		height,
		state.DiffLineCursor,
		state.DiffScrollOff,
		false, // showLineNumbers - diff has its own line# column
		isActive,
		true, // isDiff - enable 3-column diff parsing and styling
		&theme,
		state.VisualModeActive,
		state.VisualModeStart,
	)

	// Update scroll offset
	state.DiffScrollOff = newScrollOffset

	return rendered
}

// buildFileHistoryStatusBar builds the status bar for file history mode
func buildFileHistoryStatusBar(focusedPane FileHistoryPane, width int, theme Theme, overrideMessage string) string {
	styles := NewStatusBarStyles(&theme)

	// Status bar shortcuts
	parts := []string{
		styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" navigate"),
		styles.shortcutStyle.Render("TAB") + styles.descStyle.Render(" cycle panes"),
		styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" back"),
	}

	return BuildStatusBar(StatusBarConfig{
		Parts:    parts,
		Width:    width,
		Centered: true,
		Theme:    &theme,
	})
}

// buildDiffStatusBar builds the status bar for diff pane (when focused)
// Shows different hints in visual mode vs normal mode (matches old-tit)
func buildDiffStatusBar(visualModeActive bool, width int, theme Theme, overrideMessage string) string {
	styles := NewStatusBarStyles(&theme)

	// Handle override message for visual mode
	if overrideMessage != "" {
		return overrideMessage
	}

	if visualModeActive {
		// VISUAL mode: simplified, left-aligned
		parts := []string{
			styles.visualStyle.Render("VISUAL"),
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" select"),
			styles.shortcutStyle.Render("Y") + styles.descStyle.Render(" copy"),
			styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" back"),
		}
		// For visual mode, use custom join (no separator styling)
		return strings.Join(parts, styles.descStyle.Render("  ")) // Left-aligned, no padding
	}

	// NORMAL mode: full shortcuts, centered
	parts := []string{
		styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" scroll"),
		styles.shortcutStyle.Render("TAB") + styles.descStyle.Render(" cycle"),
		styles.shortcutStyle.Render("ESC") + styles.descStyle.Render(" back"),
		styles.shortcutStyle.Render("V") + styles.descStyle.Render(" visual"),
		styles.shortcutStyle.Render("Y") + styles.descStyle.Render(" copy"),
	}

	return BuildStatusBar(StatusBarConfig{
		Parts:           parts,
		Width:           width,
		Centered:        true,
		Theme:           &theme,
		OverrideMessage: overrideMessage,
	})
}
