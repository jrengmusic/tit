package app

import (
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTimeTravelCheckout handles git.TimeTravelCheckoutMsg
func (a *Application) handleTimeTravelCheckout(msg git.TimeTravelCheckoutMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	if !msg.Success {
		buffer.Append(fmt.Sprintf(ErrorMessages["time_travel_failed"], msg.Error), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)

		// Try to cleanup time travel info file
		git.ClearTimeTravelInfo()

		// Return to history mode
		a.mode = ModeHistory
		return a, nil
	}

	// Time travel successful - reload git state
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state_after_travel"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)

		// Try to cleanup time travel info file
		git.ClearTimeTravelInfo()

		// Return to history mode
		a.mode = ModeHistory
		return a, nil
	}

	a.endAsyncOp()
	a.setExitAllowed(true)

	// CONTRACT: Rebuild cache for new detached HEAD state (history always ready)
	// Don't show messages here - cache functions will show progress
	// Final "Time travel successful" message shown after cache completes
	a.cacheManager.SetLoadingStarted(true)
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
		a.endAsyncOp()
		a.setExitAllowed(true)

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
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state_after_merge"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	a.endAsyncOp()
	a.setExitAllowed(true)

	// CONTRACT: ALWAYS rebuild cache when exiting time travel (merge or return)
	// Cache was built from detached HEAD during time travel, need full branch history
	var cacheCmd tea.Cmd
	if a.gitState.Operation == git.Normal {
		a.cacheManager.SetLoadingStarted(true)
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
		a.endAsyncOp()
		a.setExitAllowed(true)

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
		a.endAsyncOp()
		a.setExitAllowed(true)

		// Return to menu
		a.mode = ModeMenu
		return a, a.startAutoUpdate()
	}

	// CRITICAL: Assign detected state to a.gitState (was local variable only!)
	a.gitState = state

	// CONTRACT: ALWAYS rebuild cache when exiting time travel (merge or return)
	// Cache was built from detached HEAD during time travel, need full branch history
	var cacheCmd tea.Cmd
	if state.Operation == git.Normal {
		a.cacheManager.SetLoadingStarted(true)
		cacheCmd = a.invalidateHistoryCaches()
	}

	// NOTE: "Press ESC..." message is appended in handleCacheProgress after cache completes

	return a, cacheCmd
}

// handleFinalizeTravelMerge handles OpFinalizeTravelMerge
func (a *Application) handleFinalizeTravelMerge(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Merge finalization succeeded: reload state and stay in console
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true)
	a.conflictResolveState = nil
	a.mode = ModeConsole

	return a, nil
}

// handleFinalizeTravelReturn handles OpFinalizeTravelReturn
func (a *Application) handleFinalizeTravelReturn(msg GitOperationMsg) (tea.Model, tea.Cmd) {
	buffer := ui.GetBuffer()

	// Merge finalization succeeded: reload state and stay in console
	if err := a.reloadGitState(); err != nil {
		buffer.Append(fmt.Sprintf(ErrorMessages["failed_detect_state"], err), ui.TypeStderr)
		a.endAsyncOp()
		a.setExitAllowed(true)
		a.mode = ModeConsole
		return a, nil
	}

	buffer.Append(OutputMessages["merge_finalized"], ui.TypeStatus)
	buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)
	a.endAsyncOp()
	a.setExitAllowed(true)
	a.conflictResolveState = nil
	a.mode = ModeConsole

	return a, nil
}
