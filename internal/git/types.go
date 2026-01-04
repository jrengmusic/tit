package git

type WorkingTree string

const (
	Clean    WorkingTree = "Clean"
	Modified WorkingTree = "Modified"
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
	NotRepo    Operation = "NotRepo"
	Normal     Operation = "Normal"
	Conflicted Operation = "Conflicted"
	Merging    Operation = "Merging"
	Rebasing   Operation = "Rebasing"
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
