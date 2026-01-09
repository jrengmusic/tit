package git

// TimeTravelCheckoutMsg represents the result of a time travel checkout operation
type TimeTravelCheckoutMsg struct {
	Success           bool
	OriginalBranch    string // Original branch before time travel
	CommitHash        string // Commit hash being traveled to
	Error             string
}

// TimeTravelMergeMsg represents the result of merging time travel changes back
type TimeTravelMergeMsg struct {
	Success           bool
	OriginalBranch    string // Branch we returned to
	TimeTravelHash    string // Hash of the time travel commit
	Error             string
	ConflictDetected  bool   // true if merge conflicts detected
	ConflictedFiles   []string // List of files with conflicts
}

// TimeTravelReturnMsg represents the result of returning from time travel without merge
type TimeTravelReturnMsg struct {
	Success           bool
	OriginalBranch    string   // Branch we returned to
	Error             string
	ConflictDetected  bool     // true if stash apply conflicts detected
	ConflictedFiles   []string // List of files with conflicts
}