package app

import (
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

	// Fallback (should never reach)
	return []MenuItem{}
}

// menuNotRepo returns menu for NotRepo state
func (a *Application) menuNotRepo() []MenuItem {
	return []MenuItem{
		{
			ID:       "init",
			Shortcut: "i",
			Emoji:    "🔨",
			Label:    "Initialize repository",
			Hint:     "Create a new git repository in current directory",
			Enabled:  true,
		},
		{
			ID:       "clone",
			Shortcut: "c",
			Emoji:    "📥",
			Label:    "Clone repository",
			Hint:     "Clone an existing repository from remote URL",
			Enabled:  true,
		},
	}
}

// menuConflicted returns menu for Conflicted state
func (a *Application) menuConflicted() []MenuItem {
	operationType := detectConflictedOperation()

	var abortLabel, abortHint string
	switch operationType {
	case "merge":
		abortLabel = "Abort merge"
		abortHint = "Cancel merge and return to pre-merge state"
	case "rebase":
		abortLabel = "Abort rebase"
		abortHint = "Cancel rebase and return to original branch"
	case "cherry-pick":
		abortLabel = "Abort cherry-pick"
		abortHint = "Cancel cherry-pick and discard changes"
	default:
		abortLabel = "Abort operation"
		abortHint = "Cancel operation and return to previous state"
	}

	return []MenuItem{
		{
			ID:       "resolve_conflicts",
			Shortcut: "r",
			Emoji:    "🔧",
			Label:    "Resolve conflicts",
			Hint:     "Open conflict resolution UI (3-way view)",
			Enabled:  true,
		},
		{
			ID:       "abort_operation",
			Shortcut: "a",
			Emoji:    "⛔",
			Label:    abortLabel,
			Hint:     abortHint,
			Enabled:  true,
		},
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
	operationType := "merge"
	if a.gitState.Operation == git.Rebasing {
		operationType = "rebase"
	}

	return []MenuItem{
		{
			ID:       "continue_operation",
			Shortcut: "c",
			Emoji:    "⏩",
			Label:    "Continue " + operationType,
			Hint:     "Resume the operation in progress",
			Enabled:  true,
		},
		{
			ID:       "abort_operation",
			Shortcut: "a",
			Emoji:    "⛔",
			Label:    "Abort " + operationType,
			Hint:     "Stop the operation and return to previous state",
			Enabled:  true,
		},
	}
}

// menuNormal returns menu for Normal operation state
func (a *Application) menuNormal() []MenuItem {
	var items []MenuItem

	// Working Tree section
	items = append(items, a.menuWorkingTree()...)

	// Timeline section
	items = append(items, a.menuTimeline()...)

	// History section
	items = append(items, a.menuHistory()...)

	return items
}

// menuWorkingTree returns working tree actions
func (a *Application) menuWorkingTree() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	isModified := a.gitState.WorkingTree == git.Modified

	return []MenuItem{
		{
			ID:       "commit",
			Shortcut: "m",
			Emoji:    "📝",
			Label:    "Commit changes",
			Hint:     "Create a new commit with staged changes",
			Enabled:  isModified,
		},
	}
}

// menuTimeline returns timeline sync actions
func (a *Application) menuTimeline() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	var items []MenuItem

	switch a.gitState.Timeline {
	case git.InSync:
		items = append(items, MenuItem{
			ID:       "pull_merge",
			Shortcut: "p",
			Emoji:    "📥",
			Label:    "Pull (fetch + merge)",
			Hint:     "Fetch latest from remote and merge into local branch",
			Enabled:  a.gitState.Remote == git.HasRemote,
		})

	case git.Ahead:
		items = append(items, MenuItem{
			ID:       "push",
			Shortcut: "h",
			Emoji:    "📤",
			Label:    "Push to remote",
			Hint:     "Send local commits to remote branch",
			Enabled:  a.gitState.Remote == git.HasRemote,
		})

	case git.Behind:
		items = append(items, MenuItem{
			ID:       "pull_merge",
			Shortcut: "p",
			Emoji:    "📥",
			Label:    "Pull (fetch + merge)",
			Hint:     "Fetch latest from remote and merge into local branch",
			Enabled:  a.gitState.Remote == git.HasRemote,
		})

	case git.Diverged:
		items = append(items, MenuItem{
			ID:       "pull_merge",
			Shortcut: "p",
			Emoji:    "📥",
			Label:    "Pull (merge strategy)",
			Hint:     "Fetch remote and merge diverged branches",
			Enabled:  a.gitState.Remote == git.HasRemote,
		})
		items = append(items, MenuItem{
			ID:       "pull_rebase",
			Shortcut: "r",
			Emoji:    "📥",
			Label:    "Pull (rebase strategy)",
			Hint:     "Fetch remote and rebase local commits on top",
			Enabled:  a.gitState.Remote == git.HasRemote,
		})
	}

	return items
}

// menuHistory returns history actions
func (a *Application) menuHistory() []MenuItem {
	return []MenuItem{
		{
			ID:       "history",
			Shortcut: "l",
			Emoji:    "📜",
			Label:    "Browse commit history",
			Hint:     "View and navigate through commit history",
			Enabled:  true,
		},
	}
}
