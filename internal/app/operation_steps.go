package app

// OperationStep constants define all async git operation step names
// Used in GitOperationMsg.Step field to track operation progress
// These are internal identifiers - not user-facing text
const (
	// Repository initialization
	OpInit = "init"

	// Clone operations
	OpClone = "clone"

	// Commit operations
	OpCommit     = "commit"
	OpCommitPush = "commit_push"

	// Push operations
	OpPush             = "push"
	OpForcePush        = "force_push"
	OpPushSyncNeeded   = "push_sync_needed"   // Push rejected - need to sync first
	OpPushSyncMerge    = "push_sync_merge"    // Merging remote into local before push
	OpFinalizePushSync = "finalize_push_sync" // Finalizing merge commit then pushing

	// Pull operations
	OpPull              = "pull"
	OpPullMerge         = "pull_merge"
	OpPullRebase        = "pull_rebase"
	OpFinalizePullMerge = "finalize_pull_merge"
	OpAbortMerge        = "abort_merge"

	// Remote operations
	OpAddRemote   = "add_remote"
	OpFetchRemote = "fetch_remote"
	OpSetUpstream = "set_upstream"
	OpCheckout    = "checkout"

	// Reset/discard operations
	OpHardReset = "hard_reset"

	// Dirty pull (save changes before pull) operation phases
	OpDirtyPullSnapshot      = "dirty_pull_snapshot"
	OpDirtyPullMerge         = "dirty_pull_merge"
	OpDirtyPullApplySnapshot = "dirty_pull_apply_snapshot"
	OpDirtyPullFinalize      = "dirty_pull_finalize"
	OpDirtyPullAbort         = "dirty_pull_abort"

	// Input mode state tracking
	OpInputModeSet = "input_mode_set"

	// Time travel operations
	OpTimeTravelCheckout   = "time_travel_checkout"
	OpTimeTravelMerge      = "time_travel_merge"
	OpFinalizeTravelMerge  = "finalize_time_travel_merge"
	OpTimeTravelReturn     = "time_travel_return"
	OpFinalizeTravelReturn = "finalize_time_travel_return"

	// Branch operations
	OpBranchCreate = "branch_create"

	OpMergeBranch         = "merge_branch"
	OpFinalizeBranchMerge = "finalize_branch_merge"

	// Dirty merge (save changes before merge) operation phases
	OpDirtyMergeSnapshot      = "dirty_merge_snapshot"
	OpDirtyMerge              = "dirty_merge"
	OpDirtyMergeApplySnapshot = "dirty_merge_apply_snapshot"
	OpDirtyMergeFinalize      = "dirty_merge_finalize"
	OpDirtyMergeAbort         = "dirty_merge_abort"
	OpFinalizeDirtyMerge      = "finalize_dirty_merge"

	// Rebase operations
	OpRebase         = "rebase"
	OpRebaseContinue = "rebase_continue"
	OpRebaseAbort    = "rebase_abort"

	// Mid-operation recovery menu actions
	OpFinalizeMergeFromMenu = "finalize_merge"
	OpAbortMergeFromMenu    = "abort_merge_from_menu"

	// Dirty branch switch (save changes before switch) operation phases
	OpDirtySwitch          = "dirty_switch"
	OpDirtySwitchSnapshot      = "dirty_switch_snapshot"
	OpDirtySwitchExecute       = "dirty_switch_execute"
	OpDirtySwitchApplySnapshot = "dirty_switch_apply_snapshot"
	OpDirtySwitchFinalize      = "dirty_switch_finalize"
	OpDirtySwitchAbort         = "dirty_switch_abort"
)
