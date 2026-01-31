package app

// WorkflowState manages transient state for multi-step workflows (clone, init).
// All fields reset when workflow completes or is cancelled.
type WorkflowState struct {
	// Clone workflow
	CloneURL      string
	ClonePath     string
	CloneMode     string   // "here" or "subdir"
	CloneBranches []string // Available branches after clone

	// Mode restoration (for ESC handling)
	PreviousMode      AppMode
	PreviousMenuIndex int

	// Pending operations
	PendingRewindCommit string

	// Return to branch from manual detached
	ReturnToBranchDirtyTree bool   // Track if working tree was dirty when entering picker
	IsReturnToBranch        bool   // True when entering picker for return-from-detached
	ReturnToBranchName      string // Target branch for return-from-detached
}

// NewWorkflowState creates a new WorkflowState with defaults.
func NewWorkflowState() WorkflowState {
	return WorkflowState{
		CloneMode:         "here",
		PreviousMode:      ModeMenu,
		PreviousMenuIndex: 0,
	}
}

// ResetClone clears all clone-related state.
func (w *WorkflowState) ResetClone() {
	w.CloneURL = ""
	w.ClonePath = ""
	w.CloneMode = "here"
	w.CloneBranches = nil
}

// SaveMode stores current mode and index for ESC restoration.
func (w *WorkflowState) SaveMode(mode AppMode, index int) {
	w.PreviousMode = mode
	w.PreviousMenuIndex = index
}

// RestoreMode returns the saved mode and index.
func (w *WorkflowState) RestoreMode() (AppMode, int) {
	return w.PreviousMode, w.PreviousMenuIndex
}

// SetPendingRewind stores a commit hash for rewind operation.
func (w *WorkflowState) SetPendingRewind(commit string) {
	w.PendingRewindCommit = commit
}

// GetPendingRewind returns the pending rewind commit (empty if none).
func (w *WorkflowState) GetPendingRewind() string {
	return w.PendingRewindCommit
}

// ClearPendingRewind removes the pending rewind commit.
func (w *WorkflowState) ClearPendingRewind() {
	w.PendingRewindCommit = ""
}
