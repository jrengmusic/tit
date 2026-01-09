package app

// AppMode represents the current application mode
type AppMode int

const (
	ModeMenu AppMode = iota
	ModeInput
	ModeConsole
	ModeConfirmation
	ModeHistory
	ModeConflictResolve
	ModeInitializeLocation    // Choose: init current dir or create subdir
	ModeInitializeBranches    // Both canon + working branch inputs (canon pre-filled with "main")
	ModeCloneURL              // Input clone URL
	ModeCloneLocation         // Choose: clone to current dir or create subdir
	ModeClone                 // Clone operation with console output
	ModeSelectBranch          // Dynamic menu to select canon branch from cloned repo
	ModeFileHistory           // File(s) history browser mode
)

// ModeMetadata describes a mode's purpose, rendering behavior, and input handling
type ModeMetadata struct {
	// String representation (e.g., "menu", "history")
	Name string
	// Human-readable description of the mode's purpose
	Description string
	// True if mode handles key input (vs. just rendering)
	AcceptsInput bool
	// True if mode is blocking/async (console, clone operations)
	IsAsync bool
}

// modeDescriptions provides metadata for each AppMode
// Used for documentation, debugging, and UI consistency
var modeDescriptions = map[AppMode]ModeMetadata{
	ModeMenu: {
		Name:         "menu",
		Description:  "Main navigation menu (state-driven): Init/Clone, Commit/Amend, Push/Pull, etc.",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeInput: {
		Name:         "input",
		Description:  "Generic text input (deprecated: being phased out in favor of dedicated modes)",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeConsole: {
		Name:         "console",
		Description:  "Streaming git command output with progress indicator",
		AcceptsInput: true, // Can abort with ESC
		IsAsync:      true,
	},
	ModeConfirmation: {
		Name:         "confirmation",
		Description:  "Yes/No confirmation dialog with title and explanation",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeHistory: {
		Name:         "history",
		Description:  "Commit history browser (timeline view with commit details)",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeConflictResolve: {
		Name:         "conflict_resolve",
		Description:  "3-way merge conflict resolver (base/ours/theirs panes)",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeInitializeLocation: {
		Name:         "init_location",
		Description:  "Menu to choose: init in current directory or create subdirectory",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeInitializeBranches: {
		Name:         "init_branches",
		Description:  "Dual input for canon branch (pre-filled 'main') and working branch names",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeCloneURL: {
		Name:         "clone_url",
		Description:  "Text input for git repository URL with validation",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeCloneLocation: {
		Name:         "clone_location",
		Description:  "Menu to choose: clone to current directory or create subdirectory",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeSelectBranch: {
		Name:         "select_branch",
		Description:  "Menu to select canon branch from cloned repository branches",
		AcceptsInput: true,
		IsAsync:      false,
	},
	ModeFileHistory: {
		Name:         "file_history",
		Description:  "File(s) history browser with before/after diff comparison (dual pane)",
		AcceptsInput: true,
		IsAsync:      false,
	},
}

// GetModeMetadata returns metadata for the given AppMode
// Returns zero struct if mode not found (fail-safe for unknown modes)
func GetModeMetadata(m AppMode) ModeMetadata {
	if meta, exists := modeDescriptions[m]; exists {
		return meta
	}
	return ModeMetadata{Name: "unknown", Description: "Unknown mode"}
}

// String returns string representation of AppMode
func (m AppMode) String() string {
	return GetModeMetadata(m).Name
}
