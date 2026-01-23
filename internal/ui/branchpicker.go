package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// BranchPickerState holds state for the branch picker mode
// Mirrors HistoryState pattern with 2-pane layout (list + details)
type BranchPickerState struct {
	SelectedIndex  int // Current selection in branch list
	ScrollOffset   int // Scroll position in branch list
	Branches       []BranchInfo
	LoadingCache   bool // True while building branch metadata cache
	CacheProgress  int  // Current branch processed
	CacheTotal     int  // Total branches to process
	AnimationFrame int  // Animation frame for loading spinner
}

// BranchInfo represents a single branch with metadata
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

// RenderBranchPickerSplitPane renders 2-pane layout (left: list, right: details)
func RenderBranchPickerSplitPane(state *BranchPickerState, width int, height int) string {
	if state == nil || len(state.Branches) == 0 {
		return "No branches found"
	}

	// Split width
	leftPaneWidth := width / 2
	rightPaneWidth := width - leftPaneWidth - 1

	// Render left pane (branch list)
	leftPane := renderBranchList(state, leftPaneWidth, height-3)

	// Render right pane (branch details)
	rightPane := renderBranchDetails(state, rightPaneWidth, height-3)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, " ", rightPane)

	return panes
}

func renderBranchList(state *BranchPickerState, width int, height int) string {
	lines := []string{}
	for i, branch := range state.Branches {
		marker := " "
		if branch.IsCurrent {
			marker = "●"
		}
		line := fmt.Sprintf("%s %s", marker, branch.Name)

		if i == state.SelectedIndex {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("4")).
				Bold(true).
				Width(width).
				Render(line)
		} else {
			line = lipgloss.NewStyle().Width(width).Render(line)
		}
		lines = append(lines, line)

		if len(lines) >= height {
			break
		}
	}

	// Pad to height
	for len(lines) < height {
		lines = append(lines, "")
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderBranchDetails(state *BranchPickerState, width int, height int) string {
	if state.SelectedIndex < 0 || state.SelectedIndex >= len(state.Branches) {
		return ""
	}

	b := state.Branches[state.SelectedIndex]
	timeStr := formatRelativeTime(b.LastCommitTime)

	lines := []string{
		fmt.Sprintf("Branch: %s", b.Name),
		fmt.Sprintf("Last Commit: %s", timeStr),
		fmt.Sprintf("Subject: %s", b.LastCommitSubj),
		fmt.Sprintf("Author: %s", b.Author),
	}

	// Tracking status
	if b.TrackingRemote != "" {
		trackStr := fmt.Sprintf("Tracking: %s", b.TrackingRemote)
		if b.Ahead > 0 || b.Behind > 0 {
			trackStr += fmt.Sprintf(" (↑%d ↓%d)", b.Ahead, b.Behind)
		}
		lines = append(lines, trackStr)
	} else {
		lines = append(lines, "Tracking: local only")
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d min ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}
	return t.Format("2006-01-02")
}
