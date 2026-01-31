package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ========================================
// Confirmation Dispatcher
// ========================================

// handleConfirmationAction dispatches to the appropriate confirmation handler
func (a *Application) handleConfirmationAction(actionID string) (tea.Model, tea.Cmd) {
	switch actionID {
	// Dirty pull confirmations
	case "confirm_dirty_pull":
		return a.executeConfirmDirtyPull()
	case "reject_dirty_pull":
		return a.executeRejectDirtyPull()
	case "confirm_pull_merge":
		return a.executeConfirmPullMerge()
	case "reject_pull_merge":
		return a.executeRejectPullMerge()

	// Time travel confirmations
	case "confirm_time_travel":
		return a.executeConfirmTimeTravel()
	case "reject_time_travel":
		return a.executeRejectTimeTravel()
	case "confirm_time_travel_return":
		return a.executeConfirmTimeTravelReturn()
	case "reject_time_travel_return":
		return a.executeRejectTimeTravelReturn()
	case "confirm_time_travel_merge":
		return a.executeConfirmTimeTravelMerge()
	case "reject_time_travel_merge":
		return a.executeRejectTimeTravelMerge()
	case "confirm_time_travel_merge_dirty_commit":
		return a.executeConfirmTimeTravelMergeDirtyCommit()
	case "confirm_time_travel_merge_dirty_discard":
		return a.executeConfirmTimeTravelMergeDirtyDiscard()
	case "confirm_time_travel_merge_dirty_stash":
		return a.executeConfirmTimeTravelMergeDirtyStash()
	case "confirm_time_travel_return_dirty_discard":
		return a.executeConfirmTimeTravelReturnDirtyDiscard()
	case "confirm_stale_stash_continue":
		return a.executeConfirmStaleStashContinue()
	case "reject_stale_stash_continue":
		return a.executeRejectStaleStashContinue()
	case "confirm_stale_stash_merge_continue":
		return a.executeConfirmStaleStashMergeContinue()
	case "reject_time_travel_return_dirty":
		return a.executeRejectTimeTravelReturnDirty()

	// Rewind confirmations
	case "confirm_rewind":
		return a.executeConfirmRewind()
	case "reject_rewind":
		return a.executeRejectRewind()

	// Branch switch confirmations
	case "confirm_branch_switch_clean":
		return a.executeConfirmBranchSwitchClean()
	case "reject_branch_switch":
		return a.executeRejectBranchSwitch()
	case "confirm_branch_switch_dirty":
		return a.executeConfirmBranchSwitchDirty()
	case "reject_branch_switch_dirty":
		return a.executeRejectBranchSwitchDirty()

	default:
		// Unknown action - hide dialog and return to menu
		a.dialogState.Hide()
		return a.returnToMenu()
	}
}
