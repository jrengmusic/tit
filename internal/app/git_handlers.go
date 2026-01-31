package app

import (
	"fmt"

	"tit/internal/git"
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
		a.endAsyncOp()
		return a.setupConflictResolverForPull(msg)
	}

	if msg.ConflictDetected && msg.Step == "branch_switch" {
		// Branch switch with conflicts: setup conflict resolver
		a.endAsyncOp()
		return a.setupConflictResolverForBranchSwitch(msg)
	}

	// Handle other failures
	if !msg.Success {
		buffer.Append(msg.Error, ui.TypeStderr)
		buffer.Append(GetFooterMessageText(MessageOperationFailed), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		a.endAsyncOp()

		// Clean up dirty operation snapshot if this was a dirty operation phase
		if a.dirtyOperationState != nil {
			snapshot := &git.DirtyOperationSnapshot{}
			snapshot.Delete() // Clean up .git/TIT_DIRTY_OP on failure
			a.dirtyOperationState = nil
		}

		// Always refresh git state after failure - the operation may have partially succeeded
		// (e.g., commit with weird exit code but changes were actually committed)
		if err := a.reloadGitState(); err != nil {
			// State reload failed, but proceed with cleanup
		}

		return a, nil
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

	case "finalize_branch_switch":
		return a.handleFinalizeBranchSwitch(msg)

	case OpForcePush:
		return a.handleForcePush(msg)

	case OpHardReset:
		return a.handleHardReset(msg)

	case OpDirtyPullSnapshot:
		// Phase 1 complete: snapshot saved, changes stashed/discarded
		// Next: merge or rebase
		buffer.Append(OutputMessages["dirty_pull_snapshot_saved"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("apply_changeset")
		return a, a.cmdDirtyPullMerge()

	case OpDirtyPullMerge:
		// Phase 2a complete: pull with merge succeeded
		// Check for conflicts before proceeding
		if msg.ConflictDetected {
			// Conflicts during merge: setup conflict resolver
			return a.setupConflictResolverForDirtyPull(msg, "changeset_apply")
		}
		// No conflicts: proceed to snapshot reapply
		buffer.Append(OutputMessages["dirty_pull_merge_succeeded"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("apply_snapshot")
		return a, a.cmdDirtyPullApplySnapshot()

	case "finalize_dirty_pull_merge":
		// Phase 2b complete: merge conflicts resolved and committed
		// Now proceed to stash apply
		buffer.Append(OutputMessages["dirty_pull_merge_succeeded"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("apply_snapshot")
		a.conflictResolveState = nil // Clear conflict state
		return a, a.cmdDirtyPullApplySnapshot()

	case OpPullRebase:
		// Phase 2b complete: pull with rebase succeeded
		// Check for conflicts before proceeding
		if msg.ConflictDetected {
			// Conflicts during rebase: setup conflict resolver
			return a.setupConflictResolverForDirtyPull(msg, "changeset_apply")
		}
		// No conflicts: proceed to snapshot reapply
		buffer.Append(OutputMessages["dirty_pull_rebase_succeeded"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("apply_snapshot")
		return a, a.cmdDirtyPullApplySnapshot()

	case OpDirtyPullApplySnapshot:
		// Phase 3 complete: stashed changes reapplied
		// Check for conflicts before finalizing
		if msg.ConflictDetected {
			// Conflicts during snapshot reapply: setup conflict resolver
			return a.setupConflictResolverForDirtyPull(msg, "snapshot_reapply")
		}
		// No conflicts: finalize operation
		buffer.Append(OutputMessages["dirty_pull_changes_reapplied"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("finalizing")
		return a, a.cmdDirtyPullFinalize()

	case OpDirtyPullFinalize:
		// Operation complete: cleanup stash and snapshot file
		if err := a.reloadGitState(); err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.endAsyncOp()
			a.dirtyOperationState = nil
			return a, nil
		}
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.endAsyncOp()
		a.dirtyOperationState = nil
		a.mode = ModeConsole

	case OpDirtyPullAbort:
		// Abort complete: original state restored
		if err := a.reloadGitState(); err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.endAsyncOp()
			a.dirtyOperationState = nil
			a.conflictResolveState = nil
			return a, nil
		}
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.endAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		a.mode = ModeConsole

	case OpFinalizeTravelMerge:
		return a.handleFinalizeTravelMerge(msg)

	case OpFinalizeTravelReturn:
		return a.handleFinalizeTravelReturn(msg)

	case "cancel":
		// User cancelled any confirmation dialog
		a.dialogState.Hide()
		return a.returnToMenu()

	default:
		// Default: just cleanup
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.endAsyncOp()
	}

	return a, nil
}
