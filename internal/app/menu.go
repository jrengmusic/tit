package app

import (
	"fmt"
	"os"
	"tit/internal/git"
)

// MenuItem represents a single menu action or separator
type MenuItem struct {
	ID        string // Unique identifier for the action
	Shortcut  string // Keyboard shortcut (single letter from label)
	Emoji     string // Leading emoji
	Label     string // Action name
	Hint      string // Plain language hint shown on focus
	Enabled   bool   // Whether this item can be selected
	Separator bool   // If true, this is a visual separator (non-selectable)
}

// MenuGenerator is a function type that generates menu items
type MenuGenerator func(*Application) []MenuItem

// GenerateMenu produces menu items based on current git state
func (a *Application) GenerateMenu() []MenuItem {
	// Priority 1: Operation State (most restrictive)
	if a.gitState == nil {
		return []MenuItem{}
	}

	menuGenerators := map[git.Operation]MenuGenerator{
		git.NotRepo:    (*Application).menuNotRepo,
		git.Conflicted: (*Application).menuConflicted,
		git.Merging:    (*Application).menuOperation,
		git.Rebasing:   (*Application).menuOperation,
		git.Normal:     (*Application).menuNormal,
	}

	if generator, exists := menuGenerators[a.gitState.Operation]; exists {
		return generator(a)
	}

	// Unknown operation: fail fast with clear error
	panic(fmt.Sprintf("Unknown git operation state: %v", a.gitState.Operation))
}

// menuNotRepo returns menu for NotRepo state
// Always shows both options; smart dispatch happens in dispatchInit/dispatchClone
func (a *Application) menuNotRepo() []MenuItem {
	return []MenuItem{
		GetMenuItem("init"),
		GetMenuItem("clone"),
	}
}

// menuConflicted returns menu for Conflicted state
func (a *Application) menuConflicted() []MenuItem {
	return []MenuItem{
		GetMenuItem("resolve_conflicts"),
		GetMenuItem("abort_operation"),
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

// menuOperation returns menu for Merging/Rebasing (no conflicts)
func (a *Application) menuOperation() []MenuItem {
	return []MenuItem{
		GetMenuItem("continue_operation"),
		GetMenuItem("abort_operation"),
	}
}

// menuNormal returns menu for Normal operation state
func (a *Application) menuNormal() []MenuItem {
	var items []MenuItem

	// Working Tree section
	items = append(items, a.menuWorkingTree()...)

	// Timeline section
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
		// Can't push uncommitted changes
		if a.gitState.WorkingTree == git.Clean {
			items = append(items,
				GetMenuItem("push"),
				GetMenuItem("force_push"),
			)
		}

	case git.Behind:
		// If Dirty, show dirty pull first
		if a.gitState.WorkingTree == git.Dirty {
			items = append(items, GetMenuItem("dirty_pull_merge"))
			// Add separator between dirty pull and clean pull
			items = append(items, Item("").Separator().Build())
		}

		// Show clean pull options
		items = append(items,
			GetMenuItem("pull_merge"),
			GetMenuItem("replace_local"),
		)

	case git.Diverged:
		// Diverged → show merge and destructive options (always available)
		items = append(items,
			GetMenuItem("pull_merge_diverged"),
			GetMenuItem("force_push"),
			GetMenuItem("replace_local"),
		)
	}

	return items
}

// menuHistory returns history actions
func (a *Application) menuHistory() []MenuItem {
	return []MenuItem{
		GetMenuItem("history"),
		GetMenuItem("file_history"),
	}
}

// menuInitializeLocation returns options for where to initialize repository
func (a *Application) menuInitializeLocation() []MenuItem {
	return []MenuItem{
		GetMenuItem("init_here"),
		GetMenuItem("init_subdir"),
	}
}

// menuCloneLocation returns options for where to clone repository
func (a *Application) menuCloneLocation() []MenuItem {
	return []MenuItem{
		GetMenuItem("clone_here"),
		GetMenuItem("clone_subdir"),
	}
}
