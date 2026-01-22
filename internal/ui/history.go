package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// CommitInfo contains basic information about a commit (for list display)
// This mirrors git.CommitInfo but is defined here to avoid import cycles
type CommitInfo struct {
	Hash    string    // Full commit hash (40 chars)
	Subject string    // Commit message first line
	Time    time.Time // Commit author date
}

// HistoryState represents the state of the history browser
type HistoryState struct {
	Commits           []CommitInfo // List of recent commits
	SelectedIdx       int          // Currently selected commit (0-indexed)
	PaneFocused       bool         // true = list pane, false = details pane
	DetailsLineCursor int          // Line cursor position in details pane
	DetailsScrollOff  int          // Scroll offset for details pane
}

// RenderHistorySplitPane renders the history split-pane view (2 columns side-by-side)
// Returns content exactly `width` chars wide and `height - 2` lines tall (for outer border)
// Parameters:
//   - state: interface{} (HistoryState from app package)
//   - theme: Theme (colors, styles)
//   - width, height: Terminal dimensions
//
// Returns: String representation of the rendered pane
func RenderHistorySplitPane(state interface{}, theme Theme, width, height int, statusBarOverride string) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Type assert to HistoryState
	historyState, ok := state.(*HistoryState)
	if !ok || historyState == nil {
		return "Error: invalid history state"
	}

	// Calculate pane height from terminal height
	// Reserve 1 line for status bar + 1 for newline separator
	paneHeight := height - 3

	// Calculate column widths based on CONTENT NEEDS
	// Commits list: "07-Jan 02:11 957f977" = 20 chars + border(2) + padding(2) = 24 chars
	listPaneWidth := 24
	detailsPaneWidth := width - listPaneWidth // Remaining width for details

	// Render both columns at same height
	listPaneContent := renderHistoryListPane(historyState, theme, listPaneWidth, paneHeight)
	detailsPaneContent := renderHistoryDetailsPane(historyState, theme, detailsPaneWidth, paneHeight)

	// Join columns horizontally (side-by-side)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, listPaneContent, detailsPaneContent)

	// Build status bar
	statusBar := buildHistoryStatusBar(historyState.PaneFocused, width, theme, statusBarOverride)

	// Stack: mainRow + statusBar
	// Total height will be (height - 3) + 1 = height - 2 (correct for outer wrapper)
	return mainRow + "\n" + statusBar
}

// renderHistoryListPane renders the list pane with commit list (matches Conflict Resolver)
func renderHistoryListPane(state *HistoryState, theme Theme, width, height int) string {
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
			IsSelected:     i == state.SelectedIdx,
		})
	}

	// Calculate visible lines for scrolling (height - border(2))
	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected commit visible (CRITICAL - was missing!)
	listPane.AdjustScroll(state.SelectedIdx, visibleLines)

	// Render list pane (active when list pane is focused)
	// Pass 0, 1 for column positioning (single column layout, treat as col 0 of 1)
	return listPane.Render(items, width, height, state.PaneFocused, 0, 1)
}

// renderHistoryDetailsPane renders the details pane with commit details using SSOT TextPane
func renderHistoryDetailsPane(state *HistoryState, theme Theme, width, height int) string {
	// Build content string
	var lines []string

	// Show commit details if available
	if len(state.Commits) > 0 && state.SelectedIdx >= 0 && state.SelectedIdx < len(state.Commits) {
		commit := state.Commits[state.SelectedIdx]

		// No "Commit: hash" line - redundant (hash already shown in list)
		lines = append(lines, fmt.Sprintf("Author: Unknown"))
		lines = append(lines, fmt.Sprintf("Date:   %s", commit.Time.Format("Mon, 2 Jan 2006 15:04:05 -0700")))
		lines = append(lines, "")

		// Split subject into multiple lines if it contains newlines or is too long
		// This allows proper scrolling through long commit messages
		subjectLines := strings.Split(commit.Subject, "\n")
		lines = append(lines, subjectLines...)
	} else {
		lines = append(lines, "(no commit selected)")
	}

	content := strings.Join(lines, "\n")

	// Use SSOT TextPane with line cursor (like Conflict Resolver diff pane)
	rendered, newScrollOffset := RenderTextPane(
		content,
		width,
		height,
		state.DetailsLineCursor, // Line cursor for better UX
		state.DetailsScrollOff,  // Current scroll offset
		false,                   // No line numbers (not code)
		!state.PaneFocused,      // Active when list is NOT focused
		false,                   // Not diff mode
		&theme,
		false, // No visual mode in history
		0,
	)

	// Update scroll offset in state
	state.DetailsScrollOff = newScrollOffset

	return rendered
}

// buildHistoryStatusBar builds the status bar with keyboard shortcuts (matches Conflict Resolver)
func buildHistoryStatusBar(listPaneFocused bool, width int, theme Theme, overrideMessage string) string {
	styles := NewStatusBarStyles(&theme)

	// Build shortcuts based on focused pane
	var parts []string
	if listPaneFocused {
		// List pane - navigate commits
		parts = []string{
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" navigate"),
		}
	} else {
		// Details pane - scroll content
		parts = []string{
			styles.shortcutStyle.Render("↑↓") + styles.descStyle.Render(" scroll"),
		}
	}
	parts = append(parts,
		styles.shortcutStyle.Render("TAB")+styles.descStyle.Render(" switch pane"),
		styles.shortcutStyle.Render("ENTER")+styles.descStyle.Render(" time travel"),
		styles.shortcutStyle.Render("ESC")+styles.descStyle.Render(" back"),
	)

	return BuildStatusBar(StatusBarConfig{
		Parts:           parts,
		Width:           width,
		Centered:        true,
		Theme:           &theme,
		OverrideMessage: overrideMessage,
	})
}
