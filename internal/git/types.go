package git

import "time"

type WorkingTree string

const (
	Clean WorkingTree = "Clean"
	Dirty WorkingTree = "Dirty"
)

type Timeline string

const (
	InSync   Timeline = "InSync"
	Ahead    Timeline = "Ahead"
	Behind   Timeline = "Behind"
	Diverged Timeline = "Diverged"
	// Empty string ("") = Timeline N/A (no remote OR detached HEAD)
)

type Operation string

const (
	NotRepo        Operation = "NotRepo"
	Normal         Operation = "Normal"
	Conflicted     Operation = "Conflicted"
	Merging        Operation = "Merging"
	Rebasing       Operation = "Rebasing"
	DirtyOperation Operation = "DirtyOperation"
	TimeTraveling  Operation = "TimeTraveling"
	Rewinding      Operation = "Rewinding" // Represents active rewind operation (git reset --hard in progress)
)

type Remote string

const (
	NoRemote  Remote = "NoRemote"
	HasRemote Remote = "HasRemote"
)

// GitEnvironment represents the machine's git environment readiness
// This is the 5th axis, checked BEFORE all other state detection
type GitEnvironment int

const (
	Ready      GitEnvironment = iota // git + ssh + key exists
	NeedsSetup                       // git + ssh exist, no SSH key
	MissingGit                       // git not installed
	MissingSSH                       // ssh not installed
)

func (e GitEnvironment) String() string {
	switch e {
	case Ready:
		return "ready"
	case NeedsSetup:
		return "needs_setup"
	case MissingGit:
		return "missing_git"
	case MissingSSH:
		return "missing_ssh"
	default:
		return "unknown"
	}
}

// State represents the complete git state tuple: (WorkingTree, Timeline, Operation, Remote)
type State struct {
	WorkingTree         WorkingTree
	Timeline            Timeline
	Operation           Operation
	Remote              Remote
	CurrentBranch       string
	CurrentHash         string
	RemoteHash          string
	CommitsAhead        int
	CommitsBehind       int
	LocalBranchOnRemote bool // Whether current branch exists on remote
	Detached            bool // HEAD is detached (not on any branch)
}

// CommitInfo contains basic information about a commit (for list display)
type CommitInfo struct {
	Hash    string    // Full commit hash (40 chars)
	Subject string    // Commit message first line
	Time    time.Time // Commit author date
}

// CommitDetails contains full metadata for a commit (for details pane)
type CommitDetails struct {
	Author  string // Author name (e.g., "John Doe")
	Date    string // Formatted date (e.g., "Mon, 7 Jan 2026 04:45:12 +0000")
	Message string // Full commit message (multiline)
}

// FileInfo contains information about a file in a commit
type FileInfo struct {
	Path   string // File path relative to repo root
	Status string // Single character: M, A, D, R, C, T, U
}

// TimeTravelInfo stores metadata for an active time travel session
type TimeTravelInfo struct {
	OriginalBranch  string     // Branch we departed from (e.g., "main")
	OriginalHead    string     // Commit hash before time travel started
	CurrentCommit   CommitInfo // Currently checked-out commit while time traveling
	OriginalStashID string     // If user had dirty tree: stash ID (empty if clean entry)
}
