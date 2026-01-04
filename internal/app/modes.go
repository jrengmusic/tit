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
