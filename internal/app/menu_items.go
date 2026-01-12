package app

// MenuItems is the single source of truth for all menu item definitions
// Every menu item ID, shortcut, emoji, label, and hint is defined here
// This ensures no conflicts and makes auditing easy
var MenuItems = map[string]MenuItem{
	// NotRepo state
	"init": {
		ID:       "init",
		Shortcut: "i",
		Emoji:    "🔨",
		Label:    "Initialize repository",
		Hint:     "Create a new git repository",
		Enabled:  true,
	},
	"clone": {
		ID:       "clone",
		Shortcut: "c",
		Emoji:    "📥",
		Label:    "Clone repository",
		Hint:     "Clone an existing repository from remote URL",
		Enabled:  true,
	},

	// Conflicted state
	"resolve_conflicts": {
		ID:       "resolve_conflicts",
		Shortcut: "r",
		Emoji:    "🔧",
		Label:    "Resolve conflicts",
		Hint:     "Open conflict resolution UI (3-way view)",
		Enabled:  true,
	},
	"abort_operation": {
		ID:       "abort_operation",
		Shortcut: "a",
		Emoji:    "💥",
		Label:    "Abort operation",
		Hint:     "Cancel operation and return to previous state",
		Enabled:  true,
	},

	// Operation state (Merging/Rebasing)
	"continue_operation": {
		ID:       "continue_operation",
		Shortcut: "c",
		Emoji:    "⏩",
		Label:    "Continue operation",
		Hint:     "Resume the operation in progress",
		Enabled:  true,
	},

	// Working tree (Normal state, Dirty)
	"commit": {
		ID:       "commit",
		Shortcut: "m",
		Emoji:    "📝",
		Label:    "Commit changes",
		Hint:     "Create a new commit with staged changes",
		Enabled:  true,
	},
	"commit_push": {
		ID:       "commit_push",
		Shortcut: "t",
		Emoji:    "🚀",
		Label:    "Commit and push",
		Hint:     "Stage, commit, and push changes in one action",
		Enabled:  true,
	},

	// Timeline: InSync
	"reset_discard_changes": {
		ID:       "reset_discard_changes",
		Shortcut: "r",
		Emoji:    "💥",
		Label:    "Reset to remote (discard changes)",
		Hint:     "💥 DESTRUCTIVE: Discard uncommitted changes, reset to remote state",
		Enabled:  true,
	},

	// Timeline: Ahead
	"push": {
		ID:       "push",
		Shortcut: "h",
		Emoji:    "📤",
		Label:    "Push to remote",
		Hint:     "Send local commits to remote branch",
		Enabled:  true,
	},
	"force_push": {
		ID:       "force_push",
		Shortcut: "f",
		Emoji:    "💥",
		Label:    "Replace remote (force push)",
		Hint:     "💥 DESTRUCTIVE: Overwrite remote branch with local commits",
		Enabled:  true,
	},

	// Timeline: Behind
	"dirty_pull_merge": {
		ID:       "dirty_pull_merge",
		Shortcut: "d",
		Emoji:    "🔺",
		Label:    "Pull (save changes)",
		Hint:     "Save WIP, pull remote, reapply changes (may conflict)",
		Enabled:  true,
	},
	"pull_merge": {
		ID:       "pull_merge",
		Shortcut: "p",
		Emoji:    "📥",
		Label:    "Pull (fetch + merge)",
		Hint:     "Fetch latest from remote and merge into local branch",
		Enabled:  true,
	},
	"replace_local": {
		ID:       "replace_local",
		Shortcut: "x",
		Emoji:    "💥",
		Label:    "Replace local (hard reset)",
		Hint:     "💥 DESTRUCTIVE: Discard local commits, match remote exactly",
		Enabled:  true,
	},

	// Timeline: Diverged
	"pull_merge_diverged": {
		ID:       "pull_merge_diverged",
		Shortcut: "p",
		Emoji:    "📥",
		Label:    "Pull (merge)",
		Hint:     "Fetch remote and merge diverged branches",
		Enabled:  true,
	},

	// History
	"history": {
		ID:       "history",
		Shortcut: "l",
		Emoji:    "📜",
		Label:    "History",
		Hint:     "View and navigate through commit history",
		Enabled:  true,
	},
	"file_history": {
		ID:       "file_history",
		Shortcut: "g",
		Emoji:    "📄",
		Label:    "File(s) history",
		Hint:     "View history of specific files",
		Enabled:  true,
	},

	// Remote
	"add_remote": {
		ID:       "add_remote",
		Shortcut: "e",
		Emoji:    "🔗",
		Label:    "Add remote",
		Hint:     "Configure a remote repository URL",
		Enabled:  true,
	},

	// Time traveling
	"time_travel_history": {
		ID:       "time_travel_history",
		Shortcut: "l",
		Emoji:    "🕒",
		Label:    "History",
		Hint:     "View commit history while time traveling",
		Enabled:  true,
	},
	"time_travel_files_history": {
		ID:       "time_travel_files_history",
		Shortcut: "g",
		Emoji:    "📄",
		Label:    "File(s) history",
		Hint:     "Browse file changes and diffs",
		Enabled:  true,
	},
	"time_travel_merge": {
		ID:       "time_travel_merge",
		Shortcut: "m",
		Emoji:    "📦",
		Label:    "Merge back",
		Hint:     "Merge changes back to original branch",
		Enabled:  true,
	},
	"time_travel_return": {
		ID:       "time_travel_return",
		Shortcut: "r",
		Emoji:    "🔙",
		Label:    "Return",
		Hint:     "Return without merging changes",
		Enabled:  true,
	},

	// Dirty operation (when stashed operation is in progress)
	"view_operation_status": {
		ID:       "view_operation_status",
		Shortcut: "v",
		Emoji:    "🔄",
		Label:    "View operation status",
		Hint:     "Show details of the operation in progress",
		Enabled:  true,
	},

	// Init/Clone location
	"init_here": {
		ID:       "init_here",
		Shortcut: "1",
		Emoji:    "📍",
		Label:    "Initialize current directory",
		Hint:     "Create repository here",
		Enabled:  true,
	},
	"init_subdir": {
		ID:       "init_subdir",
		Shortcut: "2",
		Emoji:    "📁",
		Label:    "Create subdirectory",
		Hint:     "Create new folder and initialize there",
		Enabled:  true,
	},
	"clone_here": {
		ID:       "clone_here",
		Shortcut: "1",
		Emoji:    "📍",
		Label:    "Clone to current directory",
		Hint:     "Clone repository here",
		Enabled:  true,
	},
	"clone_subdir": {
		ID:       "clone_subdir",
		Shortcut: "2",
		Emoji:    "📁",
		Label:    "Create subdirectory",
		Hint:     "Create new folder and clone there",
		Enabled:  true,
	},
}

// GetMenuItem retrieves a menu item by ID from the SSOT map
func GetMenuItem(id string) MenuItem {
	if item, exists := MenuItems[id]; exists {
		return item
	}
	panic("MenuItem not found in SSOT: " + id)
}

// ShortcutMap builds a reverse lookup: shortcut → item ID
// Detects conflicts at build time
func ShortcutMap() map[string][]string {
	conflicts := make(map[string][]string)
	for id, item := range MenuItems {
		if item.Shortcut != "" {
			conflicts[item.Shortcut] = append(conflicts[item.Shortcut], id)
		}
	}
	return conflicts
}
