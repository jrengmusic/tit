package app

// DirtyOperationState tracks the state of an active dirty pull/merge/timetravel operation
// Used by the Application to coordinate between snapshot, apply, conflict resolution, and finalize phases
type DirtyOperationState struct {
	// Operation type and phase tracking
	OperationType   string // "dirty_pull_merge", "dirty_pull_rebase", "dirty_merge", "dirty_timetravel"
	Phase           string // "snapshot", "apply_changeset", "apply_snapshot", "finalizing"
	ConflictPhase   string // "changeset" or "snapshot_reapply" (only set if conflicts occur)
	PreserveChanges bool   // true if user chose "Save changes", false if "Discard"

	// Original state (from snapshot file)
	OriginalBranch string
	OriginalHead   string

	// Operation parameters
	PullStrategy string // "merge" or "rebase" (for dirty pull)
	RemoteName   string // e.g., "origin"
	RemoteBranch string // e.g., "main"

	// Conflict information (populated if conflicts detected)
	ConflictDetectedAt string   // which phase: "changeset_apply", "snapshot_reapply"
	ConflictFiles      []string // list of conflicted file paths

	// Cleanup flags
	StashNeedsDrop bool // true if git stash apply succeeded in snapshot reapply phase

	// Progress tracking (for UI hints)
	LastMessage string // last operation message for footer
}

// NewDirtyOperationState creates a fresh state for a dirty operation
func NewDirtyOperationState(operationType string, preserveChanges bool) *DirtyOperationState {
	return &DirtyOperationState{
		OperationType:   operationType,
		Phase:           "snapshot",
		ConflictPhase:   "",
		PreserveChanges: preserveChanges,
		RemoteName:      "origin",
		StashNeedsDrop:  false,
		LastMessage:     "",
	}
}

// String returns a human-readable description of the operation state
func (d *DirtyOperationState) String() string {
	status := d.OperationType + ": " + d.Phase
	if d.ConflictPhase != "" {
		status += " (conflicts in " + d.ConflictPhase + ")"
	}
	if !d.PreserveChanges {
		status += " (changes discarded)"
	}
	return status
}

// SetPhase updates the operation phase and clears conflict phase
func (d *DirtyOperationState) SetPhase(newPhase string) {
	d.Phase = newPhase
	d.ConflictPhase = "" // Clear conflict phase when moving to next phase
}

// MarkConflictDetected sets the phase where conflicts were detected
func (d *DirtyOperationState) MarkConflictDetected(phase string, conflictedFiles []string) {
	d.ConflictPhase = phase
	d.ConflictFiles = conflictedFiles
}

// ShouldStashDrop returns true if we need to drop the stash after cleanup
func (d *DirtyOperationState) ShouldStashDrop() bool {
	return d.PreserveChanges && d.Phase == "finalizing"
}
