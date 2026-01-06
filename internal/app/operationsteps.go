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
	OpPush      = "push"
	OpForcePush = "force_push"

	// Pull operations
	OpPull                = "pull"
	OpPullMerge           = "pull_merge"
	OpPullRebase          = "pull_rebase"
	OpFinalizePullMerge   = "finalize_pull_merge"
	OpAbortMerge          = "abort_merge"

	// Remote operations
	OpAddRemote  = "add_remote"
	OpFetchRemote = "fetch_remote"
	OpSetUpstream = "set_upstream"
	OpCheckout   = "checkout"

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
)
