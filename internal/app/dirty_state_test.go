package app

import "testing"

func TestNewDirtyOperationState_Defaults(t *testing.T) {
	state := NewDirtyOperationState("dirty_pull_merge", true)

	if state.Phase != DirtyPhaseSnapshot {
		t.Errorf("Phase: got %q, want %q", state.Phase, DirtyPhaseSnapshot)
	}
	if state.RemoteName != "origin" {
		t.Errorf("RemoteName: got %q, want %q", state.RemoteName, "origin")
	}
	if state.StashNeedsDrop != false {
		t.Errorf("StashNeedsDrop: got %v, want false", state.StashNeedsDrop)
	}
}

func TestNewDirtyOperationState_PreserveChanges(t *testing.T) {
	tests := []struct {
		preserveChanges bool
	}{
		{true},
		{false},
	}
	for _, tt := range tests {
		state := NewDirtyOperationState("dirty_merge", tt.preserveChanges)
		if state.PreserveChanges != tt.preserveChanges {
			t.Errorf("PreserveChanges: got %v, want %v", state.PreserveChanges, tt.preserveChanges)
		}
	}
}

func TestNewDirtyOperationState_OperationType(t *testing.T) {
	tests := []struct {
		operationType string
	}{
		{"dirty_pull_merge"},
		{"dirty_pull_rebase"},
		{"dirty_merge"},
		{"dirty_timetravel"},
	}
	for _, tt := range tests {
		state := NewDirtyOperationState(tt.operationType, false)
		if state.OperationType != tt.operationType {
			t.Errorf("OperationType: got %q, want %q", state.OperationType, tt.operationType)
		}
	}
}

func TestAdvancePhase_SetsPhase(t *testing.T) {
	state := NewDirtyOperationState("dirty_merge", true)
	state.AdvancePhase(DirtyPhaseApplyChangeset)

	if state.Phase != DirtyPhaseApplyChangeset {
		t.Errorf("Phase: got %q, want %q", state.Phase, DirtyPhaseApplyChangeset)
	}
}

func TestAdvancePhase_ClearsConflictPhase(t *testing.T) {
	state := NewDirtyOperationState("dirty_merge", true)
	state.ConflictPhase = DirtyConflictChangeset

	state.AdvancePhase(DirtyPhaseApplySnapshot)

	if state.ConflictPhase != "" {
		t.Errorf("ConflictPhase: got %q, want empty string", state.ConflictPhase)
	}
}

func TestShouldStashDrop(t *testing.T) {
	tests := []struct {
		name            string
		preserveChanges bool
		phase           DirtyPhase
		want            bool
	}{
		{"preserve=true finalizing", true, DirtyPhaseFinalizing, true},
		{"preserve=false finalizing", false, DirtyPhaseFinalizing, false},
		{"preserve=true snapshot", true, DirtyPhaseSnapshot, false},
		{"preserve=true apply_changeset", true, DirtyPhaseApplyChangeset, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewDirtyOperationState("dirty_pull_merge", tt.preserveChanges)
			state.Phase = tt.phase
			got := state.ShouldStashDrop()
			if got != tt.want {
				t.Errorf("ShouldStashDrop(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirtyOperationState_String(t *testing.T) {
	tests := []struct {
		name          string
		operationType string
		phase         DirtyPhase
		conflictPhase DirtyConflictPhase
		preserve      bool
		want          string
	}{
		{
			name:          "basic",
			operationType: "dirty_merge",
			phase:         DirtyPhaseSnapshot,
			conflictPhase: "",
			preserve:      true,
			want:          "dirty_merge: snapshot",
		},
		{
			name:          "with conflict phase",
			operationType: "dirty_pull_merge",
			phase:         DirtyPhaseApplyChangeset,
			conflictPhase: DirtyConflictChangeset,
			preserve:      true,
			want:          "dirty_pull_merge: apply_changeset (conflicts in changeset_apply)",
		},
		{
			name:          "changes discarded",
			operationType: "dirty_merge",
			phase:         DirtyPhaseFinalizing,
			conflictPhase: "",
			preserve:      false,
			want:          "dirty_merge: finalizing (changes discarded)",
		},
		{
			name:          "conflict and discarded",
			operationType: "dirty_timetravel",
			phase:         DirtyPhaseApplySnapshot,
			conflictPhase: DirtyConflictSnapshotReapply,
			preserve:      false,
			want:          "dirty_timetravel: apply_snapshot (conflicts in snapshot_reapply) (changes discarded)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewDirtyOperationState(tt.operationType, tt.preserve)
			state.Phase = tt.phase
			state.ConflictPhase = tt.conflictPhase
			got := state.String()
			if got != tt.want {
				t.Errorf("String(): got %q, want %q", got, tt.want)
			}
		})
	}
}
