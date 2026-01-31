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
	a.timeTravelState.ClearRestore()

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
	if a.isAsyncActive() && a.mode == ModeConsole && !a.timeTravelState.IsRestoreInitiated() && hasMarker {
		// Verify marker still exists right before restoration
		a.timeTravelState.MarkRestoreInitiated()
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
			return a, tea.Tick(CacheRefreshInterval, func(t time.Time) tea.Msg {
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
