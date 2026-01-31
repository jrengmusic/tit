package app

import (
	"time"

	"tit/internal/git"
)

// TickMsg is a custom message for quit confirmation timeout
type TickMsg time.Time

// ClearTickMsg is a custom message for input clear confirmation timeout
type ClearTickMsg time.Time

// GitOperationMsg represents the result of a git operation
type GitOperationMsg struct {
	Step             string // "init", "clone", "push", "pull", etc.
	Success          bool
	Output           string
	Error            string
	Path             string   // Working directory to change to after operation
	BranchName       string   // Current branch name (for remote operations)
	ConflictDetected bool     // true if merge/rebase conflicts detected
	ConflictedFiles  []string // List of files with conflicts
}

// RestoreTimeTravelMsg signals completion of time travel restoration (Phase 0)
type RestoreTimeTravelMsg struct {
	Success bool
	Error   string
}

// GitOperationCompleteMsg signals that a git operation completed
type GitOperationCompleteMsg struct {
	Success bool
	Output  string
	Error   string
}

// InputSubmittedMsg signals that user submitted input
type InputSubmittedMsg struct {
	Action string
	Value  string
}

// OutputRefreshMsg triggers UI re-render to show updated console output
// Sent periodically during long-running operations to display streaming output
type OutputRefreshMsg struct{}

// RewindMsg represents the result of a git reset --hard operation
type RewindMsg struct {
	Commit  string // hash
	Success bool
	Output  string
	Error   string
}

// RemoteFetchMsg signals completion of background git fetch on startup
type RemoteFetchMsg struct {
	Success bool
	Error   string
}

// AutoUpdateTickMsg triggers periodic full state update
type AutoUpdateTickMsg struct{}

// AutoUpdateAnimationMsg triggers spinner animation during auto-update
type AutoUpdateAnimationMsg struct{}

// AutoUpdateCompleteMsg signals completion of background state detection
type AutoUpdateCompleteMsg struct {
	State *git.State
}

// CacheProgressMsg reports cache building progress (for UI updates)
type CacheProgressMsg struct {
	CacheType string // "metadata" or "diffs"
	Current   int    // Current item processed
	Total     int    // Total items to process
	Complete  bool   // true when cache is fully built
}

// CacheRefreshTickMsg triggers periodic UI refresh during cache building
// Sent every 100ms to update spinner animation and progress counter
type CacheRefreshTickMsg struct{}

// ========================================
// Git Constants
// ========================================

const (
	DefaultBranch = "main"
	HEADRef       = "HEAD"
)

// ========================================
// Message Domain: Input (Prompts + Hints)
// ========================================

// InputMessage pairs a prompt with its hint for a single input field
type InputMessage struct {
	Prompt string
	Hint   string
}

// InputMessages centralizes input-related messages by domain
// Replaces old InputPrompts + InputHints maps
var InputMessages = map[string]InputMessage{
	"clone_url": {
		Prompt: "Repository URL:",
		Hint:   "Enter git repository URL (https or git+ssh)",
	},
	"remote_url": {
		Prompt: "Remote URL:",
		Hint:   "Enter git repository URL and press Enter",
	},
	"commit_message": {
		Prompt: "Commit message:",
		Hint:   "Enter message and press Enter",
	},
	"subdir_name": {
		Prompt: "Subdirectory name:",
		Hint:   "Enter new directory name",
	},
	"init_branch_name": {
		Prompt: "Initial branch name:",
		Hint:   "Enter branch name (default: main), press Enter to initialize",
	},
	"init_subdir_name": {
		Prompt: "Subdirectory name:",
		Hint:   "Enter subdirectory name for new repository",
	},
	"dirty_pull_save": {
		Prompt: "Save and continue with dirty pull",
		Hint:   "Save changes before pulling",
	},
	"dirty_pull_discard": {
		Prompt: "Discard changes and pull",
		Hint:   "Discard changes before pulling",
	},
}
