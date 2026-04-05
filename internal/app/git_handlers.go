package app

import (
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Git Operation Result Handlers
// ========================================
// All functions handle GitOperationMsg returned from async operations
// They update state, handle errors, and decide next action
//
// Handler implementations are split across multiple files by operation type:
// - handlers_init.go: Init, Clone, Checkout
// - handlers_remote.go: AddRemote, FetchRemote
// - handlers_pull.go: Pull, Merge, Rebase, BranchSwitch
// - handlers_commit.go: Commit, Push, ForcePush, HardReset
// - handlers_timetravel.go: Time travel operations
// - handlers_conflict.go: Conflict resolution

// handleGitOperation dispatches GitOperationMsg to the appropriate handler
func (a *Application) handleGitOperation(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Check for conflicts BEFORE checking Success
	// Conflicts are "failures" but require special handling (conflict resolver UI)
	if msg.ConflictDetected && msg.Step == OpPull {
		// Pull operation with merge conflicts: setup conflict resolver
		a.EndAsyncOp()
		return a.setupConflictResolverForPull(msg)
	}

	if msg.ConflictDetected && msg.Step == "branch_switch" {
		// Branch switch with conflicts: setup conflict resolver
		a.EndAsyncOp()
		return a.setupConflictResolverForBranchSwitch(msg)
	}

	if msg.ConflictDetected && msg.Step == OpMergeBranch {
		a.EndAsyncOp()
		return a.setupConflictResolverForBranchMerge(msg)
	}

	// Handle other failures
	if !msg.Success {
		return a.handleGitOperationFailure(msg, buffer)
	}

	// Operation succeeded
	if msg.Output != "" {
		buffer.Append(msg.Output, ui.TypeStatus)
	}

	// Handle step-specific post-processing and chaining
	switch msg.Step {
	case OpInit, OpClone, OpCheckout:
		return a.handleInitCloneCheckout(msg)

	case OpAddRemote:
		return a.handleAddRemote(msg)

	case OpFetchRemote:
		return a.handleFetchRemote(msg)

	case OpPull:
		return a.handlePull(msg)

	case OpFinalizePullMerge:
		return a.handleFinalizePullMerge(msg)

	case OpAbortMerge:
		return a.handleAbortMerge(msg)

	case OpCommit, OpPush, OpCommitPush:
		return a.handleCommitPush(msg)

	case "branch_switch":
		return a.handleBranchSwitch(msg)

	case OpBranchCreate:
		return a.handleBranchSwitch(msg)

	case OpMergeBranch:
		return a.handleMergeBranchResult(msg)

	case OpFinalizeBranchMerge:
		return a.handleFinalizeBranchMerge(msg)

	case "finalize_branch_switch":
		return a.handleFinalizeBranchSwitch(msg)

	case OpForcePush:
		return a.handleForcePush(msg)

	case OpPushSyncNeeded:
		return a, a.cmdPushSyncMerge()

	case OpPushSyncMerge:
		return a.handlePushSyncMerge(msg)

	case OpFinalizePushSync:
		return a, a.cmdPushAfterSync()

	case OpHardReset:
		return a.handleHardReset(msg)

	case OpDirtyPullSnapshot:
		return a.handleDirtyPullSnapshot(buffer)

	case OpDirtyPullMerge:
		return a.handleDirtyPullMerge(msg, buffer)

	case "finalize_dirty_pull_merge":
		return a.handleFinalizeDirtyPullMerge(buffer)

	case OpPullRebase:
		return a.handleDirtyPullRebase(msg, buffer)

	case OpDirtyPullApplySnapshot:
		return a.handleDirtyPullApplySnapshot(msg, buffer)

	case OpDirtyPullFinalize:
		return a.handleDirtyPullFinalize(buffer)

	case OpDirtyPullAbort:
		return a.handleDirtyPullAbort(buffer)

	case OpDirtyMergeSnapshot:
		return a.handleDirtyMergeSnapshot(buffer)

	case OpDirtyMerge:
		return a.handleDirtyMergeOp(msg, buffer)

	case OpFinalizeDirtyMerge:
		return a.handleFinalizeDirtyMerge(buffer)

	case OpDirtyMergeApplySnapshot:
		return a.handleDirtyMergeApplySnapshot(msg, buffer)

	case OpDirtyMergeFinalize:
		return a.handleDirtyMergeFinalize(buffer)

	case OpDirtyMergeAbort:
		return a.handleDirtyMergeAbort(buffer)

	case OpDirtySwitchSnapshot:
		return a.handleDirtySwitchSnapshot(buffer)

	case OpDirtySwitchExecute:
		return a.handleDirtySwitchExecuteResult()

	case OpDirtySwitchApplySnapshot:
		return a.handleDirtySwitchApplySnapshot(msg, buffer)

	case OpDirtySwitchFinalize:
		return a.handleDirtySwitchFinalize(buffer)

	case OpDirtySwitchAbort:
		return a.handleDirtySwitchAbort(buffer)

	case OpRebaseContinue:
		return a.handleRebaseContinue(buffer)

	case OpRebaseAbort:
		return a.handleRebaseAbort(buffer)

	case OpFinalizeMergeFromMenu:
		return a.handleFinalizeMergeFromMenu(buffer)

	case OpFinalizeTravelMerge:
		return a.handleFinalizeTravelMerge(msg)

	case OpFinalizeTravelReturn:
		return a.handleFinalizeTravelReturn(msg)

	case "cancel":
		return a.handleCancelDialog()

	case "config_switch_remote":
		return a.handleAddRemote(msg)

	case "config_add_remote":
		return a.handleAddRemote(msg)

	default:
		return a.handleGitOperationDefault(buffer)
	}
}
