package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// dispatchAction routes menu item selections to appropriate handlers
func (a *Application) dispatchAction(actionID string) tea.Cmd {
	actionDispatchers := map[string]ActionHandler{
		"init":                      a.dispatchInit,
		"clone":                     a.dispatchClone,
		"add_remote":                a.dispatchAddRemote,
		"commit":                    a.dispatchCommit,
		"commit_push":               a.dispatchCommitPush,
		"push":                      a.dispatchPush,
		"force_push":                a.dispatchForcePush,
		"pull_merge":                a.dispatchPullMerge,
		"pull_merge_diverged":       a.dispatchPullMerge,
		"dirty_pull_merge":          a.dispatchDirtyPullMerge,
		"replace_local":             a.dispatchReplaceLocal,
		"reset_discard_changes":     a.dispatchResetDiscardChanges,
		"history":                   a.dispatchHistory,
		"file_history":              a.dispatchFileHistory,
		"time_travel_history":       a.dispatchTimeTravelHistory,
		"time_travel_files_history": a.dispatchFileHistory,
		"time_travel_merge":         a.dispatchTimeTravelMerge,
		"time_travel_return":        a.dispatchTimeTravelReturn,
		// Config menu actions
		"config_add_remote":         a.dispatchConfigAddRemote,
		"config_switch_remote":      a.dispatchConfigSwitchRemote,
		"config_remove_remote":      a.dispatchConfigRemoveRemote,
		"config_toggle_auto_update": a.dispatchConfigToggleAutoUpdate,
		"config_switch_branch":      a.dispatchConfigSwitchBranch,
		"config_preferences":        a.dispatchConfigPreferences,
		// Preferences menu actions
		"preferences_auto_update": a.dispatchPreferencesToggleAutoUpdate,
		"preferences_interval":    a.dispatchPreferencesInterval,
		"preferences_theme":       a.dispatchPreferencesCycleTheme,
	}

	if handler, exists := actionDispatchers[actionID]; exists {
		return handler(a)
	}
	return nil
}
