package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// BranchInfo represents a single branch with metadata
// Used by branch picker to display branch details
type BranchInfo struct {
	Name           string
	IsCurrent      bool
	LastCommitTime time.Time
	LastCommitHash string
	LastCommitSubj string
	Author         string
	TrackingRemote string // e.g., "origin/main", or "" if local only
	Ahead          int
	Behind         int
}

// BranchPickerState represents the state of the branch picker (2-pane split-view)
// Mirrors HistoryState pattern: list pane (left) + details pane (right)
// Uses SSOT: ListPane + TextPane for consistent rendering with history mode
type BranchPickerState struct {
	Branches          []BranchInfo // List of all branches
	SelectedIdx       int          // Currently selected branch (0-indexed)
	PaneFocused       bool         // true = list pane, false = details pane
	ListScrollOffset  int          // Scroll offset for branch list
	DetailsLineCursor int          // Line cursor position in details pane
	DetailsScrollOff  int          // Scroll offset for details pane
}

// RenderBranchPickerSplitPane renders the branch picker split-pane view (2 columns side-by-side)
// Uses SSOT: ListPane (left) + TextPane (right) matching History mode pattern
// Returns content exactly `width` chars wide and `height - 1` lines tall (footer handled externally)
func RenderBranchPickerSplitPane(state interface{}, theme Theme, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Type assert to BranchPickerState
	branchState, ok := state.(*BranchPickerState)
	if !ok || branchState == nil {
		return "Error: invalid branch picker state"
	}

	if len(branchState.Branches) == 0 {
		return "No branches found"
	}

	// Calculate pane height from terminal height (footer takes 1 line)
	paneHeight := height - 3

	// 50/50 split for branch picker (list and details get equal width)
	listPaneWidth := width / 2
	detailsPaneWidth := width - listPaneWidth // Remaining width for details

	// Render both panes at same height using SSOT components (matches history/conflict resolver pattern)
	listPaneContent := renderBranchListPane(branchState, &theme, listPaneWidth, paneHeight)
	detailsPaneContent := renderBranchDetailsPane(branchState, &theme, detailsPaneWidth, paneHeight)

	// Join columns horizontally using lipgloss (SSOT with history + conflict resolver)
	mainRow := lipgloss.JoinHorizontal(lipgloss.Top, listPaneContent, detailsPaneContent)

	return mainRow
}

// buildBranchListItems creates ListItems from branch data
// Shows branch-specific metadata (tracking status, divergence) not commit metadata
func buildBranchListItems(branches []BranchInfo, selectedIdx int, theme *Theme) []ListItem {
	items := make([]ListItem, len(branches))
	for i, branch := range branches {
		// Attribute: tracking status and divergence (branch-specific metadata)
		attrText := ""
		attrColor := theme.DimmedTextColor
		
		if branch.IsCurrent {
			attrText = "● "
			attrColor = theme.AccentTextColor
		} else {
			attrText = "  "
		}
		
		// Show tracking/divergence info
		if branch.TrackingRemote != "" {
			// Has upstream: show divergence if any
			if branch.Ahead > 0 || branch.Behind > 0 {
				attrText += fmt.Sprintf("↑%d ↓%d", branch.Ahead, branch.Behind)
			} else {
				attrText += "synced"
			}
		} else {
			// Local only branch
			attrText += "local"
		}

		items[i] = ListItem{
			AttributeText:  attrText,
			AttributeColor: attrColor,
			ContentText:    branch.Name,
			ContentColor:   theme.ContentTextColor,
			ContentBold:    branch.IsCurrent, // Bold current branch
			IsSelected:     i == selectedIdx,
		}
	}
	return items
}

// renderBranchListPane renders the list pane with branch list using SSOT ListPane
func renderBranchListPane(state *BranchPickerState, theme *Theme, width, height int) string {
	// Create list pane (SSOT with history)
	listPane := NewListPane("Branches", theme)
	listPane.ScrollOffset = state.ListScrollOffset

	// Build list items from actual branches
	items := buildBranchListItems(state.Branches, state.SelectedIdx, theme)

	visibleLines := height - 2
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Adjust scroll to keep selected branch visible
	listPane.AdjustScroll(state.SelectedIdx, visibleLines)
	state.ListScrollOffset = listPane.ScrollOffset

	// Render list pane (active when list pane is focused)
	return listPane.Render(items, width, height, state.PaneFocused, 0, 1)
}

// renderBranchDetailsPane renders the details pane with branch details using SSOT TextPane
// Clearly separates branch metadata from tip commit metadata
func renderBranchDetailsPane(state *BranchPickerState, theme *Theme, width, height int) string {
	// Build content string
	var lines []string

	if len(state.Branches) > 0 && state.SelectedIdx >= 0 && state.SelectedIdx < len(state.Branches) {
		branch := state.Branches[state.SelectedIdx]

		// === BRANCH METADATA ===
		lines = append(lines, "BRANCH")
		lines = append(lines, fmt.Sprintf("  Name: %s", branch.Name))
		
		if branch.IsCurrent {
			lines = append(lines, "  Status: ● Current")
		} else {
			lines = append(lines, "  Status: Not current")
		}
		
		// Tracking/upstream info
		if branch.TrackingRemote != "" {
			trackStr := fmt.Sprintf("  Upstream: %s", branch.TrackingRemote)
			if branch.Ahead > 0 || branch.Behind > 0 {
				trackStr += fmt.Sprintf(" (↑%d ↓%d)", branch.Ahead, branch.Behind)
			} else {
				trackStr += " (synced)"
			}
			lines = append(lines, trackStr)
		} else {
			lines = append(lines, "  Upstream: none (local only)")
		}
		
		lines = append(lines, "")
		
		// === TIP COMMIT ===
		lines = append(lines, "TIP COMMIT")
		lines = append(lines, fmt.Sprintf("  Hash: %s", branch.LastCommitHash))
		lines = append(lines, fmt.Sprintf("  Author: %s", branch.Author))
		lines = append(lines, fmt.Sprintf("  Date: %s", branch.LastCommitTime.Format("Mon, 2 Jan 2006 15:04:05 -0700")))
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  %s", branch.LastCommitSubj))
	} else {
		lines = append(lines, "(no branch selected)")
	}

	content := strings.Join(lines, "\n")

	// Use SSOT TextPane with line cursor (like history + conflict resolver diff pane)
	rendered, newScrollOffset := RenderTextPane(
		content,
		width,
		height,
		state.DetailsLineCursor, // Line cursor for navigation
		state.DetailsScrollOff,  // Current scroll offset
		false,                   // No line numbers
		!state.PaneFocused,      // Active when list is NOT focused
		false,                   // Not diff mode
		theme,
		false, // No visual mode in branch picker
		0,
	)

	// Update scroll offset in state
	state.DetailsScrollOff = newScrollOffset

	return rendered
}
