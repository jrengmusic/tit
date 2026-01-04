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
	ModeInitializeLocation      // Choose: init current dir or create subdir
	ModeInitializeCanonBranch   // Input: canon branch name
	ModeInitializeWorkingBranch // Input: working branch name
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
