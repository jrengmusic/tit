package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ConfirmationType represents different kinds of confirmation dialogs
type ConfirmationType string

const (
	ConfirmNestedRepoInit        ConfirmationType = "nested_repo_init"
	ConfirmForcePush             ConfirmationType = "force_push"
	ConfirmHardReset             ConfirmationType = "hard_reset"
	ConfirmDestructiveOp         ConfirmationType = "destructive_op"
	ConfirmAlert                 ConfirmationType = "alert"
	ConfirmPullMerge             ConfirmationType = "pull_merge"
	ConfirmPullMergeDiverged     ConfirmationType = "pull_merge_diverged"
	ConfirmDirtyPull             ConfirmationType = "dirty_pull"
	ConfirmTimeTravel            ConfirmationType = "time_travel"
	ConfirmTimeTravelReturn      ConfirmationType = "time_travel_return"
	ConfirmTimeTravelMerge       ConfirmationType = "time_travel_merge"
	ConfirmTimeTravelMergeDirty  ConfirmationType = "time_travel_merge_dirty"
	ConfirmTimeTravelReturnDirty ConfirmationType = "time_travel_return_dirty"
	ConfirmRewind                ConfirmationType = "rewind"
	ConfirmBranchSwitchClean     ConfirmationType = "branch_switch_clean"
	ConfirmBranchSwitchDirty     ConfirmationType = "branch_switch_dirty"
)

// ConfirmationAction is a function that handles a confirmed action
type ConfirmationAction func(*Application) (tea.Model, tea.Cmd)

// ConfirmationActionPair pairs a YES handler with its NO handler
// This guarantees that every confirmation type has both handlers registered
type ConfirmationActionPair struct {
	Confirm ConfirmationAction
	Reject  ConfirmationAction
}
