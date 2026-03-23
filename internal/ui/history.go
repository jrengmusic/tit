package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"tit/internal/git"
)

// CommitInfo is an alias for git.CommitInfo to avoid import cycles in UI
type CommitInfo = git.CommitInfo

// HistoryState represents the state of the history browser
type HistoryState struct {
	Commits           []CommitInfo // List of recent commits
	SelectedIdx       int          // Currently selected commit (0-indexed)
	PaneFocused       bool         // true = list pane, false = details pane
	DetailsLineCursor int          // Line cursor position in details pane
	DetailsScrollOff  int          // Scroll offset for details pane
	CopyHashMode      bool         // True when copy-hash-by-char mode is active
	CopyHashFull      bool         // True = copy full hash (Y), false = copy short hash (y)
}

// CopyHashKey represents a flash label for a visible commit
type CopyHashKey struct {
	Char      rune // The unique character to press
	CharPos   int  // Position within the shortened hash (0-6)
	CommitIdx int  // Index into Commits slice
}

// CopyHashShortHashLen is the number of characters in a shortened hash label
const CopyHashShortHashLen = 7

// CopyHashMaxVisible is the maximum number of visible commits processed for copy-hash keys
const CopyHashMaxVisible = 10

// ComputeCopyHashKeys computes unique key labels for visible commits.
// For each visible commit, finds the first char in its shortened hash
// that is unique among all visible commits at that position.
func ComputeCopyHashKeys(commits []CommitInfo, startIdx, count int) []CopyHashKey {
	if count > CopyHashMaxVisible {
		count = CopyHashMaxVisible
	}

	end := startIdx + count
	if end > len(commits) {
		end = len(commits)
	}
	if startIdx >= end {
		return nil
	}

	visible := commits[startIdx:end]
	keys := make([]CopyHashKey, 0, len(visible))

	for i, commit := range visible {
		short := ShortenHash(commit.Hash)
		assignedPos := -1

		for pos := 0; pos < len(short); pos++ {
			candidate := rune(short[pos])
			unique := true
			for j, other := range visible {
				if j == i {
					continue
				}
				otherShort := ShortenHash(other.Hash)
				if pos < len(otherShort) && rune(otherShort[pos]) == candidate {
					unique = false
					break
				}
			}
			if unique {
				assignedPos = pos
				break
			}
		}

		if assignedPos >= 0 {
			keys = append(keys, CopyHashKey{
				Char:      rune(short[assignedPos]),
				CharPos:   assignedPos,
				CommitIdx: startIdx + i,
			})
		}
	}

	return keys
}

// RenderHistorySplitPane renders the history split-pane view (2 columns side-by-side)
// Returns content exactly `width` chars wide and `height - 1` lines tall (footer handled externally)
// Parameters:
//   - state: interface{} (HistoryState from app package)
//   - theme: Theme (colors, styles)
//   - width, height: Terminal dimensions
//
// Returns: String representation of the rendered pane
func RenderHistorySplitPane(state interface{}, theme Theme, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Type assert to HistoryState
	historyState, ok := state.(*HistoryState)
	if !ok || historyState == nil {
		return "Error: invalid history state"
	}

	// Calculate pane height from terminal height (footer + padding)
	paneHeight := height - SplitPaneHeightOffset

	// Calculate column widths based on CONTENT NEEDS
	// Commits list: "07-Jan 02:11 957f977" = 20 chars + border(2) + padding(2)
	listPaneWidth := CommitListPaneWidth
	detailsPaneWidth := width - listPaneWidth // Remaining width for details

	// Render both columns at same height
	listPaneContent := renderHistoryListPane(historyState, theme, listPaneWidth, paneHeight)
	detailsPaneContent := renderHistoryDetailsPane(historyState, theme, detailsPaneWidth, paneHeight)

	// Join columns horizontally (side-by-side)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, listPaneContent, detailsPaneContent)

	return mainRow
}

// buildCommitListItems creates ListItems from commit data for rendering in list panes.
// copyHashKeys, when non-empty, overlays flash labels for copy-hash mode.
func buildCommitListItems(commits []CommitInfo, selectedIdx int, theme Theme, copyHashKeys []CopyHashKey) []ListItem {
	// Build a lookup from commit index → key for O(1) access
	keyByCommitIdx := make(map[int]CopyHashKey, len(copyHashKeys))
	for _, k := range copyHashKeys {
		keyByCommitIdx[k.CommitIdx] = k
	}

	items := make([]ListItem, len(commits))
	for i, commit := range commits {
		item := ListItem{
			AttributeText:  commit.Time.Format("02-Jan 15:04"),
			AttributeColor: theme.DimmedTextColor,
			ContentText:    ShortenHash(commit.Hash),
			ContentColor:   theme.ContentTextColor,
			ContentBold:    false,
			IsSelected:     i == selectedIdx && len(copyHashKeys) == 0,
		}

		if key, ok := keyByCommitIdx[i]; ok {
			item.CopyHashChar    = key.Char
			item.CopyHashCharPos = key.CharPos
			item.CopyHashFg     = theme.CopyHashLabelForeground
			item.CopyHashBg     = theme.CopyHashLabelBackground
		}

		items[i] = item
	}
	return items
}

// renderHistoryListPane renders the list pane with commit list (matches Conflict Resolver)
func renderHistoryListPane(state *HistoryState, theme Theme, width, height int) string {
	// Create list pane for commits
	listPane := NewListPane("Commits", &theme)

	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll FIRST so ScrollOffset is correct before computing hash keys
	listPane.AdjustScroll(state.SelectedIdx, visibleLines)

	// Compute copy-hash flash keys for visible window when mode is active
	var copyHashKeys []CopyHashKey
	if state.CopyHashMode {
		copyHashKeys = ComputeCopyHashKeys(state.Commits, listPane.ScrollOffset, visibleLines)
	}

	// Build list items from actual commits
	items := buildCommitListItems(state.Commits, state.SelectedIdx, theme, copyHashKeys)

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
