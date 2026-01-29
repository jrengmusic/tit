package app

import (
	"fmt"
	"time"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *Application) handleRewind(msg RewindMsg) (tea.Model, tea.Cmd) {
	a.endAsyncOp()
	buffer := ui.GetBuffer()

	if !msg.Success {
		// Rewind failed - show error in console and stay
		buffer.Append(fmt.Sprintf(ErrorMessages["rewind_failed"], msg.Error), ui.TypeStderr)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		return a, nil
	}

	// Rewind succeeded - show success message and stay in console
	buffer.Append(OutputMessages["rewind_completed"], ui.TypeStatus)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)

	// Refresh git state for next menu display (when user presses ESC)
	a.reloadGitState() // silent failure is acceptable here

	return a, nil
}

// handleRestoreTimeTravel processes the result of time travel restoration (Phase 0)

func (a *Application) handleRestoreTimeTravel(msg RestoreTimeTravelMsg) (tea.Model, tea.Cmd) {
	a.endAsyncOp()
	a.clearTimeTravelRestore()

	if !msg.Success {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(ConsoleMessages["restoration_error"], msg.Error), ui.TypeStderr)
		a.footerHint = "Press ESC to acknowledge error"
		// Stay in console mode so user can read error
		return a, nil
	}

	// Reload git state after successful restoration
	if err := a.reloadGitState(); err != nil {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(ConsoleMessages["error_detect_state"], err), ui.TypeStderr)
	}

	// Regenerate menu for new state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0
	if len(menu) > 0 {
		a.footerHint = menu[0].Hint
	}

	// Stay in console mode until user presses ESC to acknowledge
	a.footerHint = "Restoration complete. Press ESC to continue"
	return a, nil
}

// Package-level counter for Update() calls
var updateCallCount int = 0

// Update handles all messages

func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateCallCount++
	// CRITICAL: If restoration is needed, initiate it on first update (Phase 0)
	hasMarker := git.FileExists(".git/TIT_TIME_TRAVEL")

	// Double-check before calling RestoreFromTimeTravel
	if a.isAsyncActive() && a.mode == ModeConsole && !a.isTimeTravelRestoreInitiated() && hasMarker {
		// Verify marker still exists right before restoration
		a.markTimeTravelRestoreInitiated()
		return a, a.RestoreFromTimeTravel()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.sizing = ui.CalculateDynamicSizing(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		// Handle bracketed paste - entire paste comes as single KeyMsg with Paste=true
		if msg.Paste && a.isInputMode() {
			text := string(msg.Runes) // Don't trim - preserve formatting
			if len(text) > 0 {
				a.insertTextAtCursor(text)
				a.updateInputValidation()
			}
			return a, nil
		}
		keyStr := msg.String()

		// Handle ctrl+j (shift+enter equivalent) for newline in multiline input
		if a.isInputMode() && (keyStr == "ctrl+j" || keyStr == "shift+enter") {
			a.insertTextAtCursor("\n")
			return a, nil
		}

		// Look up handler in cached registry
		if modeHandlers, modeExists := a.keyHandlers[a.mode]; modeExists {
			if handler, exists := modeHandlers[keyStr]; exists {
				return handler(a)
			}
		}

		// Handle character input in input modes
		if a.isInputMode() && len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
			a.insertTextAtCursor(keyStr)
			a.updateInputValidation()
			return a, nil
		}

		// Handle backspace in input modes
		if a.isInputMode() && keyStr == "backspace" {
			a.deleteAtCursor()
			a.updateInputValidation()
			return a, nil
		}

	case TickMsg:
		if a.quitConfirmActive {
			a.quitConfirmActive = false
			a.footerHint = "" // Clear confirmation message
		}

	case ClearTickMsg:
		if a.inputState.ClearConfirming {
			a.inputState.ClearConfirming = false
			a.footerHint = "" // Clear confirmation message
		}

	case OutputRefreshMsg:
		// Force re-render to display updated console output
		// If operation still active, schedule next refresh tick
		if a.isAsyncActive() {
			return a, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return OutputRefreshMsg{}
			})
		}
		// Operation completed, stop sending refresh messages
		return a, nil

	case GitOperationMsg:
		// WORKER THREAD - git operation completed
		return a.handleGitOperation(msg)

	case CacheProgressMsg:
		// Cache building progress update
		return a.handleCacheProgress(msg)

	case CacheRefreshTickMsg:
		// Periodic tick to refresh cache progress UI
		return a.handleCacheRefreshTick()

	case RestoreTimeTravelMsg:
		// Time travel restoration completed (Phase 0)
		return a.handleRestoreTimeTravel(msg)

	case git.TimeTravelCheckoutMsg:
		// Time travel checkout operation completed
		return a.handleTimeTravelCheckout(msg)

	case git.TimeTravelMergeMsg:
		// Time travel merge operation completed
		return a.handleTimeTravelMerge(msg)

	case git.TimeTravelReturnMsg:
		// Time travel return operation completed
		return a.handleTimeTravelReturn(msg)
	case SetupCompleteMsg:
		// SSH key generation completed successfully
		if msg.Step == "generate" {
			a.environmentState.SetWizardStep(SetupStepDisplayKey)
		}
		return a, nil

	case SetupErrorMsg:
		// Error occurred during setup - show error to user
		a.environmentState.SetWizardError(msg.Error)
		a.environmentState.SetWizardStep(SetupStepError)
		return a, nil

	case RewindMsg:
		// WORKER THREAD - rewind operation completed
		return a.handleRewind(msg)

	case RemoteFetchMsg:
		// Background fetch completed - refresh state to update timeline
		if msg.Success {
			if newState, err := git.DetectState(); err == nil {
				a.gitState = newState
				// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
				if a.mode == ModeMenu {
					a.menuItems = a.GenerateMenu()
					a.rebuildMenuShortcuts(ModeMenu)
					a.updateFooterHintFromMenu()
				}
			}
		}
		return a, nil

	case AutoUpdateTickMsg:
		// Periodic auto-update tick
		return a.handleAutoUpdateTick()

	case AutoUpdateAnimationMsg:
		// Auto-update spinner animation frame
		return a.handleAutoUpdateAnimation()

	case AutoUpdateCompleteMsg:
		// Background state detection completed
		return a.handleAutoUpdateComplete(msg.State)

	}

	return a, nil
}

// View renders the current view

func (a *Application) handleCacheProgress(msg CacheProgressMsg) (tea.Model, tea.Cmd) {
	// Cache progress received - regenerate menu to show updated progress
	// Menu generator will read progress fields and show disabled state with progress

	// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
	if a.mode == ModeMenu {
		menu := a.GenerateMenu()
		a.menuItems = menu
		if !a.quitConfirmActive && len(menu) > 0 && a.selectedIndex < len(menu) {
			a.footerHint = menu[a.selectedIndex].Hint
		}
	}

	if msg.Complete {
		// Cache complete - rebuild menu shortcuts to enable items (only in ModeMenu)
		if a.mode == ModeMenu {
			a.rebuildMenuShortcuts(ModeMenu)
		}

		// Check if BOTH caches are now complete (for time travel success message)
		metadataReady := a.cacheManager.IsMetadataReady()
		diffsReady := a.cacheManager.IsDiffsReady()

		// If both caches complete AND in console mode after async operation finished,
		// show "Press ESC to return to menu" message
		if metadataReady && diffsReady && a.mode == ModeConsole && !a.isAsyncActive() {
			buffer := ui.GetBuffer()

			// Check if this is time travel mode (handled separately)
			if a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
				buffer.Append(ConsoleMessages["time_travel_success"], ui.TypeStatus)
			} else {
				// Regular operation (commit, push, etc.) - show completion message
				buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
			}
		}
	}

	return a, nil
}

// handleCacheRefreshTick handles periodic cache progress refresh
// Regenerates menu to show updated progress and re-schedules if caches not complete

func (a *Application) handleCacheRefreshTick() (tea.Model, tea.Cmd) {
	// Check if both caches are complete
	metadataComplete := a.cacheManager.IsMetadataReady()
	diffsComplete := a.cacheManager.IsDiffsReady()

	// If both complete, stop ticking
	if metadataComplete && diffsComplete {
		return a, nil
	}

	// Advance animation frame
	a.cacheManager.IncrementAnimationFrame()

	// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
	if a.mode == ModeMenu {
		menu := a.GenerateMenu()
		a.menuItems = menu
		if !a.quitConfirmActive && len(menu) > 0 && a.selectedIndex < len(menu) {
			a.footerHint = menu[a.selectedIndex].Hint
		}
	}

	// Re-schedule another tick
	return a, a.cmdRefreshCacheProgress()
}

// updateFooterHintFromMenu updates footer with hint of currently selected menu item
// Skips update if app-level message is active (quitConfirmActive)

func (a *Application) updateFooterHintFromMenu() {
	if a.quitConfirmActive {
		return
	}
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.menuItems) {
		if !a.menuItems[a.selectedIndex].Separator {
			a.footerHint = a.menuItems[a.selectedIndex].Hint
		}
	}
}

// GetGitState returns the current git state

func (a *Application) GetGitState() interface{} {
	return a.gitState
}

// RenderStateHeader renders the full git state header (5 rows) using lipgloss
// Row 1: CWD (left) | OPERATION (right)
// RenderStateHeader renders the state header per REACTIVE-LAYOUT-PLAN.md
// 2-column layout: 80/20 split
// LEFT (80%): CWD, Remote, WorkingTree, Timeline
// RIGHT (20%): Operation, Branch
