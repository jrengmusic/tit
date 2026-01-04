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
)

// ModeString returns string representation of AppMode
func (m AppMode) String() string {
	modes := map[AppMode]string{
		ModeMenu:              "menu",
		ModeInput:             "input",
		ModeConsole:           "console",
		ModeConfirmation:      "confirmation",
		ModeHistory:           "history",
		ModeConflictResolve:   "conflict_resolve",
	}

	if s, exists := modes[m]; exists {
		return s
	}
	return "unknown"
}
