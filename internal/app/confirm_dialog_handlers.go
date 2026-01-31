package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// confirmationHandlers maps confirmation types to paired YES/NO handlers
// Using a single paired structure ensures no accidental missing confirm/reject pairs
var confirmationHandlers = map[string]ConfirmationActionPair{
	string(ConfirmNestedRepoInit): {
		Confirm: (*Application).executeConfirmNestedRepoInit,
		Reject:  (*Application).executeRejectNestedRepoInit,
	},
	string(ConfirmForcePush): {
		Confirm: (*Application).executeConfirmForcePush,
		Reject:  (*Application).executeRejectForcePush,
	},
	string(ConfirmHardReset): {
		Confirm: (*Application).executeConfirmHardReset,
		Reject:  (*Application).executeRejectHardReset,
	},
	string(ConfirmAlert): {
		Confirm: (*Application).executeAlert,
		Reject:  (*Application).executeAlert, // Any key dismisses alert
	},
	string(ConfirmDirtyPull): {
		Confirm: (*Application).executeConfirmDirtyPull,
		Reject:  (*Application).executeRejectDirtyPull,
	},
	string(ConfirmPullMerge): {
		Confirm: (*Application).executeConfirmPullMerge,
		Reject:  (*Application).executeRejectPullMerge,
	},
	string(ConfirmPullMergeDiverged): {
		Confirm: (*Application).executeConfirmPullMerge,
		Reject:  (*Application).executeRejectPullMerge,
	},
	string(ConfirmTimeTravel): {
		Confirm: (*Application).executeConfirmTimeTravel,
		Reject:  (*Application).executeRejectTimeTravel,
	},
	string(ConfirmTimeTravelReturn): {
		Confirm: (*Application).executeConfirmTimeTravelReturn,
		Reject:  (*Application).executeRejectTimeTravelReturn,
	},
	string(ConfirmTimeTravelMerge): {
		Confirm: (*Application).executeConfirmTimeTravelMerge,
		Reject:  (*Application).executeRejectTimeTravelMerge,
	},
	string(ConfirmTimeTravelMergeDirty): {
		Confirm: (*Application).executeConfirmTimeTravelMergeDirtyCommit,
		Reject:  (*Application).executeConfirmTimeTravelMergeDirtyDiscard,
	},
	string(ConfirmTimeTravelReturnDirty): {
		Confirm: (*Application).executeConfirmTimeTravelReturnDirtyDiscard,
		Reject:  (*Application).executeRejectTimeTravelReturnDirty,
	},
	"time_travel_return_dirty_choice": {
		Confirm: (*Application).executeConfirmTimeTravelMergeDirtyStash,    // YES = Stash
		Reject:  (*Application).executeConfirmTimeTravelReturnDirtyDiscard, // NO = Discard
	},
	"confirm_stale_stash_continue": {
		Confirm: (*Application).executeConfirmStaleStashContinue,
		Reject:  (*Application).executeRejectStaleStashContinue,
	},
	"confirm_stale_stash_merge_continue": {
		Confirm: (*Application).executeConfirmStaleStashMergeContinue,
		Reject:  (*Application).executeRejectStaleStashContinue,
	},
	string(ConfirmRewind): {
		Confirm: (*Application).executeConfirmRewind,
		Reject:  (*Application).executeRejectRewind,
	},
	string(ConfirmBranchSwitchClean): {
		Confirm: (*Application).executeConfirmBranchSwitchClean,
		Reject:  (*Application).executeRejectBranchSwitch,
	},
	string(ConfirmBranchSwitchDirty): {
		Confirm: (*Application).executeConfirmBranchSwitchDirty,
		Reject:  (*Application).executeRejectBranchSwitchDirty,
	},
}

// handleConfirmationResponse routes confirmation YES/NO responses to appropriate handlers
func (a *Application) handleConfirmationResponse(confirmed bool) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() == nil {
		// No active confirmation dialog
		return a.returnToMenu()
	}

	confirmType := a.dialogState.GetDialog().Config.ActionID
	actions, ok := confirmationHandlers[confirmType]
	if !ok {
		// No handler registered for this type - return to menu
		a.dialogState.Hide()
		return a.returnToMenu()
	}

	var handler ConfirmationAction
	if confirmed {
		handler = actions.Confirm
	} else {
		handler = actions.Reject
	}

	if handler == nil {
		// Paired handler is nil - return to menu (shouldn't happen with paired structure)
		a.dialogState.Hide()
		return a.returnToMenu()
	}

	return handler(a)
}

// executeConfirmNestedRepoInit handles YES response to nested repo warning
func (a *Application) executeConfirmNestedRepoInit() (tea.Model, tea.Cmd) {
	// User confirmed they want to init in a nested repo
	// Return to previous mode (ModeInitializeLocation)
	a.dialogState.Hide()
	a.mode = ModeInitializeLocation
	return a, nil
}

// executeRejectNestedRepoInit handles NO response to nested repo warning
func (a *Application) executeRejectNestedRepoInit() (tea.Model, tea.Cmd) {
	// User cancelled - abort init, return to menu
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmForcePush handles YES response to force push confirmation
func (a *Application) executeConfirmForcePush() (tea.Model, tea.Cmd) {
	// User confirmed force push
	// Initiate async push --force operation
	a.dialogState.Hide()
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdForcePush()
}

// executeRejectForcePush handles NO response to force push confirmation
func (a *Application) executeRejectForcePush() (tea.Model, tea.Cmd) {
	// User cancelled force push
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmHardReset handles YES response to hard reset confirmation
func (a *Application) executeConfirmHardReset() (tea.Model, tea.Cmd) {
	// User confirmed hard reset
	a.dialogState.Hide()
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdHardReset()
}

// executeRejectHardReset handles NO response to hard reset confirmation
func (a *Application) executeRejectHardReset() (tea.Model, tea.Cmd) {
	// User cancelled hard reset
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeAlert handles alert dialog dismissal (any response)
func (a *Application) executeAlert() (tea.Model, tea.Cmd) {
	// Alert dialogs are dismissed with any key press
	a.dialogState.Hide()
	return a.returnToMenu()
}
