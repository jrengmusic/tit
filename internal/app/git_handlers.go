package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Git Operation Result Handlers
// ========================================
// All functions handle GitOperationMsg returned from async operations
// They update state, handle errors, and decide next action

// handleGitOperation dispatches GitOperationMsg to the appropriate handler
func (a *Application) handleGitOperation(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Check for conflicts BEFORE checking Success
	// Conflicts are "failures" but require special handling (conflict resolver UI)
	if msg.ConflictDetected && msg.Step == OpPull {
		// Pull operation with merge conflicts: setup conflict resolver
		a.asyncOperationActive = false
		return a.setupConflictResolverForPull(msg)
	}

	if msg.ConflictDetected && msg.Step == "branch_switch" {
		// Branch switch with conflicts: setup conflict resolver
		a.asyncOperationActive = false
		return a.setupConflictResolverForBranchSwitch(msg)
	}

	// Handle other failures
	if !msg.Success {
		buffer.Append(msg.Error, ui.TypeStderr)
		buffer.Append(GetFooterMessageText(MessageOperationFailed), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		a.asyncOperationActive = false

		// Clean up dirty operation snapshot if this was a dirty operation phase
		if a.dirtyOperationState != nil {
			snapshot := &git.DirtyOperationSnapshot{}
			snapshot.Delete() // Clean up .git/TIT_DIRTY_OP on failure
			a.dirtyOperationState = nil
		}

		// Always refresh git state after failure - the operation may have partially succeeded
		// (e.g., commit with weird exit code but changes were actually committed)
		if state, err := git.DetectState(); err == nil {
			a.gitState = state
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

	case OpAddRemote:
		// Chain: OpAddRemote → OpFetchRemote
		buffer.Append(OutputMessages["fetching_remote"], ui.TypeInfo)
		return a, a.cmdFetchRemote()

	case OpFetchRemote:
		// Fetch complete: set upstream tracking
		buffer.Append(OutputMessages["setting_upstream"], ui.TypeInfo)
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			return a, nil
		}
		a.gitState = state
		return a, a.cmdSetUpstream(a.gitState.CurrentBranch)

	case OpPull:
		// Pull operation succeeded (no conflicts)
		// Reload state and return to menu
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.isExitAllowed = true // Re-enable exit on error
			return a, nil
		}
		a.gitState = state
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.isExitAllowed = true // Re-enable exit after successful pull
		return a, nil

	case OpFinalizePullMerge, OpFinalizeTravelMerge, OpFinalizeTravelReturn:
		// Merge finalization succeeded: reload state and stay in console
		// User must press ESC to return to menu (ensures merge completed before menu reachable)
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.isExitAllowed = true // Re-enable exit on error
			a.mode = ModeConsole
			return a, nil
		}
		a.gitState = state
		buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.isExitAllowed = true // Re-enable exit after successful merge finalization
		a.conflictResolveState = nil
		a.mode = ModeConsole // Stay in console, user presses ESC to return to menu
		return a, nil

	case OpAbortMerge:
		// Merge abort succeeded: reload state and stay in console
		// User must press ESC to return to menu (ensures abort completed before menu reachable)
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.isExitAllowed = true // Re-enable exit on error
			a.mode = ModeConsole
			return a, nil
		}
		a.gitState = state
		buffer.Append(OutputMessages["abort_successful"], ui.TypeStatus)
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.isExitAllowed = true // Re-enable exit after successful abort
		a.conflictResolveState = nil
		a.mode = ModeConsole // Stay in console, user presses ESC to return to menu
		return a, nil

	case OpCommit, OpPush, OpCommitPush:
		// Simple operations: reload state
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			return a, nil
		}
		a.gitState = state

		// CONTRACT: Rebuild cache before showing completion (commit changes history)
		cacheCmd := a.invalidateHistoryCaches()

		a.asyncOperationActive = false

		// NOTE: "Press ESC..." message is appended in handleCacheProgress after cache completes
		// This ensures cache messages appear before "Press ESC to return to menu"

		return a, cacheCmd

	case "branch_switch":
		// Branch switch completed successfully (no conflicts or conflicts resolved)
		// Reload state and return to config menu
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			return a, nil
		}
		a.gitState = state

		// Regenerate menu with new branch state
		menu := a.GenerateMenu()
		a.menuItems = menu
		a.selectedIndex = 0

		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.mode = ModeConsole // Stay in console so user sees the success message

	case "finalize_branch_switch":
		// Branch switch conflicts resolved and finalized
		// Reload state and return to config menu
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.isExitAllowed = true
			a.mode = ModeConsole
			return a, nil
		}
		a.gitState = state

		// Regenerate menu with new branch state
		menu := a.GenerateMenu()
		a.menuItems = menu
		a.selectedIndex = 0

		buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.isExitAllowed = true
		a.conflictResolveState = nil
		a.mode = ModeConsole // Stay in console, user presses ESC to return to menu

	case OpForcePush:
		// Force push completed - reload state, stay in console
		// User presses ESC to return to menu

		// DEBUG: Log before DetectState call
		buffer.Append("[DEBUG] OpForcePush handler: About to call git.DetectState() after force push", ui.TypeDebug)

		state, err := git.DetectState()

		// DEBUG: Log after DetectState call
		if err != nil {
			buffer.Append(fmt.Sprintf("[DEBUG] OpForcePush handler: git.DetectState() returned error: %v", err), ui.TypeDebug)
		} else {
			buffer.Append(fmt.Sprintf("[DEBUG] OpForcePush handler: git.DetectState() completed. WorkingTree=%v, Operation=%v", state.WorkingTree, state.Operation), ui.TypeDebug)
		}

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

	case OpHardReset:
		// Hard reset completed - reload state, stay in console
		// User presses ESC to return to menu
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
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.dirtyOperationState = nil
			return a, nil
		}
		a.gitState = state
		buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
		a.footerHint = GetFooterMessageText(MessageOperationComplete)
		a.asyncOperationActive = false
		a.dirtyOperationState = nil
		a.mode = ModeConsole

	case OpDirtyPullAbort:
		// Abort complete: original state restored
		state, err := git.DetectState()
		if err != nil {
			buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
			a.asyncOperationActive = false
			a.dirtyOperationState = nil
			a.conflictResolveState = nil
			return a, nil
		}
		a.gitState = state
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

// handleTimeTravelCheckout handles git.TimeTravelCheckoutMsg
func (a *Application) handleTimeTravelCheckout(msg git.TimeTravelCheckoutMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	if !msg.Success {
		buffer.Append(fmt.Sprintf(ErrorMessages["time_travel_failed"], msg.Error), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// Try to cleanup time travel info file
		git.ClearTimeTravelInfo()

		// Return to history mode
		a.mode = ModeHistory
		return a, nil
	}

	// Time travel successful - reload git state
	state, err := git.DetectState()
	if err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state_after_travel"], err), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// Try to cleanup time travel info file
		git.ClearTimeTravelInfo()

		// Return to history mode
		a.mode = ModeHistory
		return a, nil
	}

	a.gitState = state
	a.asyncOperationActive = false
	a.isExitAllowed = true

	// CONTRACT: Rebuild cache for new detached HEAD state (history always ready)
	// Don't show messages here - cache functions will show progress
	// Final "Time travel successful" message shown after cache completes
	a.cacheLoadingStarted = true
	cacheCmd := a.invalidateHistoryCaches()

	// STAY IN CONSOLE - Let user see output and press ESC to return to menu
	// Mode remains ModeConsole, git state is now Operation = TimeTraveling
	// ESC handler will detect this and show time travel menu

	return a, cacheCmd
}

// handleTimeTravelMerge handles git.TimeTravelMergeMsg
func (a *Application) handleTimeTravelMerge(msg git.TimeTravelMergeMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	if !msg.Success {
		buffer.Append(fmt.Sprintf(ErrorMessages["time_travel_merge_failed"], msg.Error), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// If conflicts detected, set up conflict resolver
		if msg.ConflictDetected {
			// Build dynamic labels with commit hash and date
			// Column 1: BASE (common ancestor) - [hash date]
			// Column 2: MAIN (current branch) - [hash date]
			// Column 3: YOUR CHANGES

			// Get merge base (common ancestor)
			mergeBaseResult := git.Execute("merge-base", msg.OriginalBranch, msg.TimeTravelHash)
			baseHash := "???????"
			baseDate := ""
			if mergeBaseResult.Success {
				baseHash = strings.TrimSpace(mergeBaseResult.Stdout)[:7] // Short hash
				dateResult := git.Execute("show", "-s", "--format=%cd", "--date=short", baseHash)
				if dateResult.Success {
					baseDate = strings.TrimSpace(dateResult.Stdout)
				}
			}

			// Get main branch tip
			mainHashResult := git.Execute("rev-parse", msg.OriginalBranch)
			mainHash := "???????"
			mainDate := ""
			if mainHashResult.Success {
				mainHash = strings.TrimSpace(mainHashResult.Stdout)[:7]
				dateResult := git.Execute("show", "-s", "--format=%cd", "--date=short", mainHash)
				if dateResult.Success {
					mainDate = strings.TrimSpace(dateResult.Stdout)
				}
			}

			labels := []string{
				fmt.Sprintf("%s %s", baseHash, baseDate),
				fmt.Sprintf("%s %s", mainHash, mainDate),
				"YOUR CHANGES",
			}

			return a.setupConflictResolver("time_travel_merge", labels)
		}

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Time travel merge successful - reload git state
	state, err := git.DetectState()
	if err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state_after_merge"], err), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	a.gitState = state
	a.asyncOperationActive = false
	a.isExitAllowed = true

	// CONTRACT: ALWAYS rebuild cache when exiting time travel (merge or return)
	// Cache was built from detached HEAD during time travel, need full branch history
	var cacheCmd tea.Cmd
	if state.Operation == git.Normal {
		a.cacheLoadingStarted = true
		cacheCmd = a.invalidateHistoryCaches()
	}

	// NOTE: "Press ESC..." message is appended in handleCacheProgress after cache completes

	return a, cacheCmd
}

// handleTimeTravelReturn handles git.TimeTravelReturnMsg
func (a *Application) handleTimeTravelReturn(msg git.TimeTravelReturnMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	if !msg.Success {
		buffer.Append(fmt.Sprintf(ErrorMessages["time_travel_return_failed"], msg.Error), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// If conflicts detected, set up conflict resolver
		if msg.ConflictDetected {
			// Build dynamic labels
			// For stash apply conflicts, git only uses 2 stages:
			// - Stage 2 (LOCAL): Current main branch state
			// - Stage 3 (REMOTE): Stashed changes
			// No stage 1 (BASE) since stash is a diff, not a merge

			// Get main branch tip
			mainHashResult := git.Execute("rev-parse", msg.OriginalBranch)
			mainHash := "???????"
			mainDate := ""
			if mainHashResult.Success {
				mainHash = strings.TrimSpace(mainHashResult.Stdout)[:7]
				dateResult := git.Execute("show", "-s", "--format=%cd", "--date=short", mainHash)
				if dateResult.Success {
					mainDate = strings.TrimSpace(dateResult.Stdout)
				}
			}

			// Only 2 labels for 2-way conflict (no BASE)
			// setupConflictResolver will auto-detect only stages 2 and 3 exist
			// and use labels[1] and labels[2] (skipping labels[0])
			labels := []string{
				"",                                       // Placeholder for BASE (won't be used)
				fmt.Sprintf("%s %s", mainHash, mainDate), // MAIN branch (stage 2)
				"YOUR CHANGES",                           // Stashed changes (stage 3)
			}

			return a.setupConflictResolver("time_travel_return", labels)
		}

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// Time travel return successful - reload git state
	state, err := git.DetectState()
	if err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state_after_return"], err), ui.TypeStderr)
		a.asyncOperationActive = false
		a.isExitAllowed = true

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// CONTRACT: ALWAYS rebuild cache when exiting time travel (merge or return)
	// Cache was built from detached HEAD during time travel, need full branch history
	var cacheCmd tea.Cmd
	if state.Operation == git.Normal {
		a.cacheLoadingStarted = true
		cacheCmd = a.invalidateHistoryCaches()
	}

	// NOTE: "Press ESC..." message is appended in handleCacheProgress after cache completes

	return a, cacheCmd
}

// setupConflictResolver initializes conflict resolver UI for any conflict-resolving operation
// Parameters:
//   - operation: operation identifier (e.g., "pull_merge", "dirty_pull_changeset_apply", "cherry_pick")
//   - columnLabels: human-readable labels for the 3 columns (e.g., ["BASE", "LOCAL (yours)", "REMOTE (theirs)"])
func (a *Application) setupConflictResolver(operation string, columnLabels []string) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	buffer.Append(OutputMessages["detecting_conflicts"], ui.TypeInfo)

	// Get list of conflicted files from git status
	conflictFiles, err := git.ListConflictedFiles()
	if err != nil {
		buffer.Append(fmt.Sprintf(OutputMessages["conflict_detection_error"], err), ui.TypeStderr)
		a.asyncOperationActive = false
		a.footerHint = ErrorMessages["operation_failed"]
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(fmt.Sprintf("Found %d conflicted file(s)", len(conflictFiles)), ui.TypeInfo)

	if len(conflictFiles) == 0 {
		// No conflicts found - should not happen, but handle gracefully
		buffer.Append(OutputMessages["conflict_detection_none"], ui.TypeInfo)
		a.asyncOperationActive = false
		a.mode = ModeConsole
		return a, nil
	}

	// Detect which stages exist by checking the first conflicted file
	// This determines how many columns we'll show (2-way vs 3-way merge)
	var stagesPresent []int
	var activeLabels []string

	if len(conflictFiles) > 0 {
		// Check which stages exist for first file (all files should have same stage structure)
		testFile := conflictFiles[0]

		// Try stage 1 (BASE)
		if _, err := git.ShowConflictVersion(testFile, 1); err == nil {
			stagesPresent = append(stagesPresent, 1)
		}

		// Try stage 2 (LOCAL)
		if _, err := git.ShowConflictVersion(testFile, 2); err == nil {
			stagesPresent = append(stagesPresent, 2)
		}

		// Try stage 3 (REMOTE)
		if _, err := git.ShowConflictVersion(testFile, 3); err == nil {
			stagesPresent = append(stagesPresent, 3)
		}

		// Build active labels based on which stages exist
		// columnLabels is indexed 0, 1, 2 for BASE, LOCAL, REMOTE
		for _, stage := range stagesPresent {
			labelIdx := stage - 1 // Stage 1->label[0], stage 2->label[1], stage 3->label[2]
			if labelIdx < len(columnLabels) {
				activeLabels = append(activeLabels, columnLabels[labelIdx])
			}
		}
	}

	numColumns := len(stagesPresent)
	if numColumns == 0 {
		// Fallback: assume 3-way merge
		numColumns = 3
		stagesPresent = []int{1, 2, 3}
		activeLabels = columnLabels
	}

	// Read versions for each conflicted file
	resolveState := &ConflictResolveState{
		Files:             make([]ui.ConflictFileGeneric, 0, len(conflictFiles)),
		SelectedFileIndex: 0,
		FocusedPane:       0,
		NumColumns:        numColumns,
		ColumnLabels:      activeLabels,
		ScrollOffsets:     make([]int, numColumns),
		LineCursors:       make([]int, numColumns),
		Operation:         operation,
	}

	for _, filePath := range conflictFiles {
		var versions []string

		// Read only the stages that actually exist
		for _, stage := range stagesPresent {
			content, err := git.ShowConflictVersion(filePath, stage)
			if err != nil {
				content = fmt.Sprintf("Error reading stage %d: %v", stage, err)
			}
			versions = append(versions, content)
		}

		// Build conflict file entry
		conflictFile := ui.ConflictFileGeneric{
			Path:     filePath,
			Versions: versions,
			Chosen:   -1, // Not yet marked
		}
		resolveState.Files = append(resolveState.Files, conflictFile)
	}

	// Store conflict state and transition to resolver UI
	a.conflictResolveState = resolveState
	a.asyncOperationActive = false
	a.mode = ModeConflictResolve
	a.footerHint = fmt.Sprintf(ConsoleMessages["resolve_conflicts_help"], len(conflictFiles))

	buffer.Append(fmt.Sprintf(OutputMessages["conflicts_detected_count"], len(conflictFiles)), ui.TypeInfo)
	buffer.Append(OutputMessages["mark_choices_in_resolver"], ui.TypeInfo)

	return a, nil
}

// setupConflictResolverForPull initializes conflict resolver for pull merge conflicts (convenience wrapper)
func (a *Application) setupConflictResolverForPull(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	return a.setupConflictResolver("pull_merge", []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})
}

// setupConflictResolverForDirtyPull initializes conflict resolver for dirty pull conflicts (convenience wrapper)
func (a *Application) setupConflictResolverForDirtyPull(msg GitOperationMsg, conflictPhase string) (tea.Model, tea.Cmd) {
	operation := "dirty_pull_" + conflictPhase
	return a.setupConflictResolver(operation, []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"})
}

// setupConflictResolverForBranchSwitch initializes conflict resolver for branch switch conflicts (convenience wrapper)
// Conflicts occur when switching would overwrite local changes
func (a *Application) setupConflictResolverForBranchSwitch(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	targetBranch := msg.BranchName
	currentBranch := ""
	if a.gitState != nil {
		currentBranch = a.gitState.CurrentBranch
	}

	// Column labels: BASE, LOCAL (current branch), REMOTE (target branch)
	labels := []string{
		"BASE",
		fmt.Sprintf("LOCAL (%s)", currentBranch),
		fmt.Sprintf("REMOTE (%s)", targetBranch),
	}

	return a.setupConflictResolver("branch_switch", labels)
}
