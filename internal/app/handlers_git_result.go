package app

import (
	"fmt"

	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleGitOperationFailure handles the !msg.Success path: logs error, cleans up dirty state, reloads git state.
func (a *Application) handleGitOperationFailure(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(msg.Error, ui.TypeStderr)
	buffer.Append(GetFooterMessageText(MessageOperationFailed), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationFailed)
	a.EndAsyncOp()

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

// handleGitOperationDefault handles unrecognized steps: logs completion and cleans up.
func (a *Application) handleGitOperationDefault(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	return a, nil
}

// handleCancelDialog handles the "cancel" step: dismisses dialog and returns to menu.
func (a *Application) handleCancelDialog() (tea.Model, tea.Cmd) {
	a.dialogState.Hide()
	return a.returnToMenu()
}

// handlePushSyncMerge handles OpPushSyncMerge: if conflict detected routes to resolver, else pushes.
func (a *Application) handlePushSyncMerge(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForPushSync(msg)
	}
	return a, a.cmdPushAfterSync()
}

// handleDirtyPullSnapshot handles OpDirtyPullSnapshot: Phase 1 complete, snapshot saved.
func (a *Application) handleDirtyPullSnapshot(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(OutputMessages["dirty_pull_snapshot_saved"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplyChangeset)
	return a, a.cmdDirtyPullMerge()
}

// handleDirtyPullMerge handles OpDirtyPullMerge: Phase 2a complete, pull with merge succeeded.
func (a *Application) handleDirtyPullMerge(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtyPull(msg, DirtyConflictChangeset)
	}
	buffer.Append(OutputMessages["dirty_pull_merge_succeeded"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	return a, a.cmdDirtyPullApplySnapshot()
}

// handleFinalizeDirtyPullMerge handles finalize_dirty_pull_merge: Phase 2b complete, merge conflicts resolved.
func (a *Application) handleFinalizeDirtyPullMerge(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(OutputMessages["dirty_pull_merge_succeeded"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	a.conflictResolveState = nil
	return a, a.cmdDirtyPullApplySnapshot()
}

// handleDirtyPullRebase handles OpPullRebase: Phase 2b complete, pull with rebase succeeded.
func (a *Application) handleDirtyPullRebase(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtyPull(msg, DirtyConflictChangeset)
	}
	buffer.Append(OutputMessages["dirty_pull_rebase_succeeded"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	return a, a.cmdDirtyPullApplySnapshot()
}

// handleDirtyPullApplySnapshot handles OpDirtyPullApplySnapshot: Phase 3 complete, stashed changes reapplied.
func (a *Application) handleDirtyPullApplySnapshot(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtyPull(msg, DirtyConflictSnapshotReapply)
	}
	buffer.Append(OutputMessages["dirty_pull_changes_reapplied"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
	return a, a.cmdDirtyPullFinalize()
}

// handleDirtyPullFinalize handles OpDirtyPullFinalize: operation complete, cleanup stash and snapshot.
func (a *Application) handleDirtyPullFinalize(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleDirtyPullAbort handles OpDirtyPullAbort: abort complete, original state restored.
func (a *Application) handleDirtyPullAbort(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleDirtyMergeSnapshot handles OpDirtyMergeSnapshot: snapshot saved, proceed to merge.
func (a *Application) handleDirtyMergeSnapshot(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(OutputMessages["dirty_merge_snapshot_saved"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplyChangeset)
	return a, a.cmdDirtyMerge()
}

// handleDirtyMergeOp handles OpDirtyMerge: merge phase complete.
func (a *Application) handleDirtyMergeOp(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtyMerge(msg, DirtyConflictChangeset)
	}
	buffer.Append(OutputMessages["dirty_merge_succeeded"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	return a, a.cmdDirtyMergeApplySnapshot()
}

// handleFinalizeDirtyMerge handles OpFinalizeDirtyMerge: merge conflicts resolved, proceed to snapshot reapply.
func (a *Application) handleFinalizeDirtyMerge(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(OutputMessages["dirty_merge_succeeded"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	a.conflictResolveState = nil
	return a, a.cmdDirtyMergeApplySnapshot()
}

// handleDirtyMergeApplySnapshot handles OpDirtyMergeApplySnapshot: snapshot reapply phase complete.
func (a *Application) handleDirtyMergeApplySnapshot(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtyMerge(msg, DirtyConflictSnapshotReapply)
	}
	buffer.Append(OutputMessages["dirty_pull_changes_reapplied"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
	return a, a.cmdDirtyMergeFinalize()
}

// handleDirtyMergeFinalize handles OpDirtyMergeFinalize: merge operation complete, cleanup.
func (a *Application) handleDirtyMergeFinalize(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleDirtyMergeAbort handles OpDirtyMergeAbort: abort complete, original state restored.
func (a *Application) handleDirtyMergeAbort(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleDirtySwitchSnapshot handles OpDirtySwitchSnapshot: snapshot saved, proceed to switch.
func (a *Application) handleDirtySwitchSnapshot(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	buffer.Append(OutputMessages["dirty_switch_snapshot_saved"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplyChangeset)
	return a, a.cmdDirtySwitchExecute()
}

// handleDirtySwitchExecuteResult handles OpDirtySwitchExecute: switch complete, proceed to snapshot reapply.
func (a *Application) handleDirtySwitchExecuteResult() (tea.Model, tea.Cmd) {
	a.dirtyOperationState.AdvancePhase(DirtyPhaseApplySnapshot)
	return a, a.cmdDirtySwitchApplySnapshot()
}

// handleDirtySwitchApplySnapshot handles OpDirtySwitchApplySnapshot: snapshot reapply phase complete.
func (a *Application) handleDirtySwitchApplySnapshot(msg GitOperationMsg, buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if msg.ConflictDetected {
		return a.setupConflictResolverForDirtySwitch(msg, DirtyConflictSnapshotReapply)
	}
	buffer.Append(OutputMessages["dirty_pull_changes_reapplied"], ui.TypeInfo)
	a.dirtyOperationState.AdvancePhase(DirtyPhaseFinalizing)
	return a, a.cmdDirtySwitchFinalize()
}

// handleDirtySwitchFinalize handles OpDirtySwitchFinalize: switch operation complete, regenerate menu.
func (a *Application) handleDirtySwitchFinalize(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	// Regenerate menu with new branch state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}

// handleDirtySwitchAbort handles OpDirtySwitchAbort: abort complete, original state restored.
func (a *Application) handleDirtySwitchAbort(buffer *ui.OutputBuffer) (tea.Model, tea.Cmd) {
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.EndAsyncOp()
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		return a, nil
	}
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.EndAsyncOp()
	a.dirtyOperationState = nil
	a.conflictResolveState = nil
	a.mode = ModeConsole
	return a, nil
}
