package app

import (
	"os"
	"tit/internal/git"
)

// menuNotRepo returns menu for NotRepo state
// Always shows both options; smart dispatch happens in dispatchInit/dispatchClone
func (a *Application) menuNotRepo() []MenuItem {
	return []MenuItem{
		GetMenuItem("init"),
		GetMenuItem("clone"),
	}
}

// detectConflictedOperation determines which operation caused conflicts
func detectConflictedOperation() string {
	if _, err := os.Stat(".git/MERGE_HEAD"); err == nil {
		return "merge"
	}
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return "rebase"
	}
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return "rebase"
	}
	if _, err := os.Stat(".git/CHERRY_PICK_HEAD"); err == nil {
		return "cherry-pick"
	}
	return "unknown"
}

// menuNormal returns menu for Normal operation state
func (a *Application) menuNormal() []MenuItem {
	var items []MenuItem

	items = append(items, a.menuWorkingTree()...)

	items = append(items, a.menuTimeline()...)

	// Separator before History section (if there are items above)
	if len(items) > 0 {
		items = append(items, Item("").Separator().Build())
	}

	// History section (always shown)
	items = append(items, a.menuHistory()...)

	// First-time setup (always at bottom)
	if a.gitState.Remote == git.NoRemote {
		items = append(items,
			Item("").Separator().Build(),
			GetMenuItem("add_remote"),
		)
	}

	return items
}

// menuWorkingTree returns working tree actions
func (a *Application) menuWorkingTree() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	// Only show commit when Dirty - HIDDEN when Clean
	switch a.gitState.WorkingTree {
	case git.Clean:
		return []MenuItem{} // No working tree actions when Clean

	case git.Dirty:
		items := []MenuItem{
			GetMenuItem("commit"),
		}

		// Show "Commit and push" only if remote exists
		if a.gitState.Remote == git.HasRemote {
			items = append(items, GetMenuItem("commit_push"))
		}

		return items
	}

	return []MenuItem{}
}

// menuTimeline returns timeline sync actions
func (a *Application) menuTimeline() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	// No remote → no timeline operations (add_remote shown at bottom of menuNormal)
	if a.gitState.Remote == git.NoRemote {
		return []MenuItem{}
	}

	var items []MenuItem

	switch a.gitState.Timeline {
	case git.InSync:
		// When in sync but have uncommitted changes, allow reset to remote
		if a.gitState.WorkingTree == git.Dirty && a.gitState.Remote == git.HasRemote {
			items = append(items, GetMenuItem("reset_discard_changes"))
		}
		// No other sync actions needed when in sync
		return items

	case git.Ahead:
		// Local ahead → show push ONLY if working tree is Clean
		// cannot push uncommitted changes
		if a.gitState.WorkingTree == git.Clean {
			items = append(items,
				GetMenuItem("push"),
				GetMenuItem("force_push"),
			)
		}

	case git.Behind:
		// Dirty → ONLY dirty pull (don't offer clean pull that would lose work)
		if a.gitState.WorkingTree == git.Dirty {
			items = append(items,
				GetMenuItem("dirty_pull_merge"),
				GetMenuItem("replace_local"), // Destructive option (has confirmation)
			)
		} else {
			// Clean → ONLY clean pull options
			items = append(items,
				GetMenuItem("pull_merge"),
				GetMenuItem("replace_local"),
			)
		}

	case git.Diverged:
		// Dirty → ONLY dirty pull (don't offer clean pull that would lose work)
		if a.gitState.WorkingTree == git.Dirty {
			items = append(items,
				GetMenuItem("dirty_pull_merge"),
				GetMenuItem("force_push"),    // Destructive option (has confirmation)
				GetMenuItem("replace_local"), // Destructive option (has confirmation)
			)
		} else {
			// Clean → ONLY clean pull options
			items = append(items,
				GetMenuItem("pull_merge_diverged"),
				GetMenuItem("force_push"),
				GetMenuItem("replace_local"),
			)
		}
	}

	return items
}

// menuHistory returns history actions
// CONTRACT: Disables menu items and shows progress while cache is building
func (a *Application) menuHistory() []MenuItem {
	return a.getHistoryItemsWithCacheState("history", "file_history")
}
