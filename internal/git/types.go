package git

import "time"

type WorkingTree string

const (
	Clean WorkingTree = "Clean"
	Dirty WorkingTree = "Dirty"
)

type Timeline string

const (
	InSync           Timeline = "InSync"
	Ahead            Timeline = "Ahead"
	Behind           Timeline = "Behind"
	Diverged         Timeline = "Diverged"
	TimelineNoRemote Timeline = "NoRemote"
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
)

type Remote string

const (
	NoRemote  Remote = "NoRemote"
	HasRemote Remote = "HasRemote"
)

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
