package app

import "tit/internal/ui"

// ConflictResolveState holds state for conflict resolution mode
// Used for dirty pull, cherry-pick conflicts, and other conflict scenarios
type ConflictResolveState struct {
	// Operation tracking
	Operation      string // "dirty_pull", "cherry_pick", "external_conflict"
	CommitHash     string // Hash of commit being applied (for display only)
	IsRebase       bool   // true if using rebase (vs merge) for dirty pull
	StashNeedsDrop bool   // true if stash apply succeeded and we need to drop it later

	// Conflict files - generic N-column model
	Files             []ui.ConflictFileGeneric
	SelectedFileIndex int      // Which file is selected (shared across all top panes)
	FocusedPane       int      // Which pane is focused (cyclic: 0...NumColumns*2-1)
	NumColumns        int      // Number of version columns (2 for cherry-pick, 3 for dirty pull)
	ColumnLabels      []string // Labels for each column (e.g., ["LOCAL", "REMOTE"])
	ScrollOffsets     []int    // Scroll position for each bottom pane (length = NumColumns)
	LineCursors       []int    // Line cursor for each bottom pane (length = NumColumns)
}
