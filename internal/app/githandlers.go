package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// ========================================
// Git Operation Result Handlers
// ========================================
// All functions handle GitOperationMsg returned from async operations
// They update state, handle errors, and decide next action

// handleGitOperation dispatches GitOperationMsg to the appropriate handler
func (a *Application) handleGitOperation(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Handle failure
	if !msg.Success {
		buffer.Append(msg.Error, ui.TypeStderr)
		buffer.Append(GetFooterMessageText(MessageOperationFailed), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		a.asyncOperationActive = false
		return a, nil
	}

	// Operation succeeded
	if msg.Output != "" {
		buffer.Append(msg.Output, ui.TypeStatus)
	}

	// Handle step-specific post-processing and chaining
	switch msg.Step {
	case "init", "clone", "checkout":
		// Init/clone/checkout: reload state, keep console visible
		// User presses ESC to return to menu
		if msg.Path != "" {
			// Change to the path if specified
			if err := os.Chdir(msg.Path); err != nil {
				buffer.Append(fmt.Sprintf(ErrorMessages["failed_cd_into"], msg.Path, err), ui.TypeStderr)
				a.asyncOperationActive = false
				return a, nil
			}
		}
		
		// Detect new state after init/clone/checkout
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			return a, nil
		}
		a.gitState = state
		
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.mode = ModeConsole

	case "add_remote":
		// Chain: add_remote → fetch_remote
		buffer.Append("Fetching from remote...", ui.TypeInfo)
		return a, a.cmdFetchRemote()

	case "fetch_remote":
		// Fetch complete: set upstream tracking
		buffer.Append("Setting upstream tracking...", ui.TypeInfo)
		a.gitState, _ = git.DetectState()
		return a, a.cmdSetUpstream(a.gitState.CurrentBranch)

	case "commit", "push", "pull":
		// Simple operations: reload state
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false

	case "force_push":
		// Force push completed - reload state, stay in console  
		// User presses ESC to return to menu
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.mode = ModeConsole

	case "hard_reset":
		// Hard reset completed - reload state, stay in console
		// User presses ESC to return to menu
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.mode = ModeConsole

	case "dirty_pull_snapshot":
		// Phase 1 complete: snapshot saved, changes stashed/discarded
		// Next: merge or rebase
		buffer.Append(OutputMessages["dirty_pull_snapshot_saved"], ui.TypeInfo)
		a.dirtyOperationState.SetPhase("apply_changeset")
		return a, a.cmdDirtyPullMerge()

	case "dirty_pull_merge":
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

	case "dirty_pull_rebase":
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

	case "dirty_pull_apply_snapshot":
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

	case "dirty_pull_finalize":
		// Operation complete: cleanup stash and snapshot file
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.dirtyOperationState = nil
		a.mode = ModeConsole

	case "dirty_pull_abort":
		// Abort complete: original state restored
		a.gitState, _ = git.DetectState()
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.dirtyOperationState = nil
		a.conflictResolveState = nil
		a.mode = ModeConsole

	case "cancel":
		// User cancelled any confirmation dialog
		a.confirmationDialog = nil
		return a.returnToMenu()

	default:
		// Default: just cleanup
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
	}

	return a, nil
}

// setupConflictResolverForDirtyPull initializes conflict resolver UI for dirty pull operations
// Reads conflicted files from git status and populates 3-way conflict versions
func (a *Application) setupConflictResolverForDirtyPull(msg GitOperationMsg, conflictPhase string) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()
	
	// Get list of conflicted files from git status
	conflictFiles, err := git.ListConflictedFiles()
	if err != nil {
		buffer.Append(fmt.Sprintf(OutputMessages["conflict_detection_error"], err), ui.TypeStderr)
		a.asyncOperationActive = false
		a.footerHint = ErrorMessages["operation_failed"]
		return a, nil
	}
	
	if len(conflictFiles) == 0 {
		// No conflicts found - continue operation
		buffer.Append(OutputMessages["conflict_detection_none"], ui.TypeInfo)
		a.asyncOperationActive = false
		return a, nil
	}
	
	// Read 3-way versions for each conflicted file
	resolveState := &ConflictResolveState{
		Files:               make([]ui.ConflictFileGeneric, 0, len(conflictFiles)),
		SelectedFileIndex:   0,
		FocusedPane:         0,
		NumColumns:          3, // Base, Local, Remote
		LineCursors:         make([]int, 3),
		Operation:           "dirty_pull_" + conflictPhase,
		DiffPane:            nil,
	}
	
	for _, filePath := range conflictFiles {
		// Get base version (stage 1)
		baseContent, err := git.ShowConflictVersion(filePath, 1)
		if err != nil {
			baseContent = fmt.Sprintf("Error reading base: %v", err)
		}
		
		// Get local version (stage 2)
		localContent, err := git.ShowConflictVersion(filePath, 2)
		if err != nil {
			localContent = fmt.Sprintf("Error reading local: %v", err)
		}
		
		// Get remote version (stage 3)
		remoteContent, err := git.ShowConflictVersion(filePath, 3)
		if err != nil {
			remoteContent = fmt.Sprintf("Error reading remote: %v", err)
		}
		
		// Build conflict file entry
		conflictFile := ui.ConflictFileGeneric{
			Path: filePath,
			Versions: []string{
				baseContent,
				localContent,
				remoteContent,
			},
			Chosen: -1, // Not yet marked
		}
		resolveState.Files = append(resolveState.Files, conflictFile)
	}
	
	// Store conflict state and transition to resolver UI
	a.conflictResolveState = resolveState
	a.asyncOperationActive = false
	a.mode = ModeConflictResolve
	a.footerHint = fmt.Sprintf(FooterHints["resolve_conflicts_help"], len(conflictFiles))
	
	buffer.Append(fmt.Sprintf(OutputMessages["conflicts_detected_count"], len(conflictFiles)), ui.TypeInfo)
	buffer.Append(OutputMessages["mark_choices_in_resolver"], ui.TypeInfo)
	
	return a, nil
}
