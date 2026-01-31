package app

import (
	"tit/internal"
	"tit/internal/ui"
)

// FooterMessageType enum for different footer message states
type FooterMessageType int

const (
	MessageNone FooterMessageType = iota
	MessageCtrlCConfirm
	MessageEscClearConfirm
	MessageInit
	MessageClone
	MessageCommit
	MessagePush
	MessagePull
	MessageAddRemote
	MessageResolveConflicts
	MessageOperationComplete
	MessageOperationFailed
	MessageOperationInProgress
	MessageOperationAborting
	MessageExitBlocked
)

// GetFooterMessageText returns display text for a message type
func GetFooterMessageText(msgType FooterMessageType) string {
	timeoutStr := internal.QuitConfirmTimeout.String()
	messages := map[FooterMessageType]string{
		MessageNone:                "",
		MessageCtrlCConfirm:        "Press Ctrl+C again to quit (" + timeoutStr + " timeout)",
		MessageEscClearConfirm:     "Press ESC again to clear input (" + timeoutStr + " timeout)",
		MessageInit:                "Initializing repository... (ESC to abort)",
		MessageClone:               "Cloning repository... (ESC to abort)",
		MessageCommit:              "Committing changes... (ESC to abort)",
		MessagePush:                "Pushing to remote... (ESC to abort)",
		MessagePull:                "Pulling from remote... (ESC to abort)",
		MessageAddRemote:           "Adding remote and fetching... (ESC to abort)",
		MessageResolveConflicts:    "Resolving conflicts...",
		MessageOperationComplete:   "Press ESC to return to menu",
		MessageOperationFailed:     "Failed. Press ESC to return.",
		MessageOperationInProgress: "Operation in progress. Please wait for completion.",
		MessageOperationAborting:   "Aborting operation. Please wait...",
		MessageExitBlocked:         "Exit blocked. Operation must complete or be aborted first.",
	}

	if msg, exists := messages[msgType]; exists {
		return msg
	}
	return ""
}

// FooterShortcut represents a single keyboard shortcut hint
type FooterShortcut struct {
	Key  string // e.g., "↑↓", "Enter", "Esc"
	Desc string // e.g., "navigate", "select", "back"
}

// FooterHintShortcuts defines all mode-specific footer shortcuts (SSOT)
// Key = mode identifier, Value = list of shortcuts
var FooterHintShortcuts = map[string][]ui.FooterShortcut{
	// History mode
	"history_list": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Enter", Desc: "time travel"},
		{Key: "Tab", Desc: "details"},
		{Key: "Esc", Desc: "back"},
	},
	"history_details": {
		{Key: "↑↓", Desc: "scroll"},
		{Key: "Tab", Desc: "list"},
		{Key: "Esc", Desc: "back"},
	},

	// File History mode
	"filehistory_commits": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Tab", Desc: "files"},
		{Key: "Esc", Desc: "back"},
	},
	"filehistory_files": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Tab", Desc: "diff"},
		{Key: "Esc", Desc: "back"},
	},
	"filehistory_diff": {
		{Key: "↑↓", Desc: "scroll"},
		{Key: "v", Desc: "visual"},
		{Key: "Tab", Desc: "commits"},
		{Key: "Esc", Desc: "back"},
	},
	"filehistory_visual": {
		{Key: "↑↓", Desc: "extend"},
		{Key: "y", Desc: "yank"},
		{Key: "Esc", Desc: "cancel"},
	},

	// Conflict Resolver
	"conflict_list": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Space", Desc: "toggle"},
		{Key: "Tab", Desc: "diff"},
		{Key: "Enter", Desc: "resolve"},
	},
	"conflict_diff": {
		{Key: "↑↓", Desc: "scroll"},
		{Key: "Tab", Desc: "list"},
		{Key: "Esc", Desc: "back"},
	},

	// Console
	"console_running": {
		{Key: "↑↓", Desc: "scroll"},
		{Key: "Esc", Desc: "abort"},
	},
	"console_complete": {
		{Key: "↑↓", Desc: "scroll"},
		{Key: "Esc", Desc: "back"},
	},

	// Input
	"input_empty": {
		{Key: "Enter", Desc: "submit"},
		{Key: "Esc", Desc: "back"},
	},
	"input_filled": {
		{Key: "Enter", Desc: "submit"},
		{Key: "Esc", Desc: "clear"},
	},

	// Confirmation
	"confirmation": {
		{Key: "←→", Desc: "select"},
		{Key: "Enter", Desc: "confirm"},
		{Key: "Esc", Desc: "cancel"},
	},

	// Branch Picker
	"branch_picker": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Enter", Desc: "switch"},
		{Key: "Esc", Desc: "cancel"},
	},

	// Preferences
	"preferences": {
		{Key: "↑↓", Desc: "navigate"},
		{Key: "Space", Desc: "toggle"},
		{Key: "=/-", Desc: "±1min"},
		{Key: "+/_", Desc: "±10min"},
		{Key: "Esc", Desc: "back"},
	},
}

// TimelineSyncMessages centralizes timeline sync footer hint messages (SSOT)
var TimelineSyncMessages = map[string]string{
	"sync_completed":              "Auto-update sync completed",
	"sync_failed":                 "Sync failed: %s",
	"auto_update_enabled":         "Auto-update enabled",
	"auto_update_disabled":        "Auto-update disabled",
	"auto_update_enabled_syncing": "Auto-update enabled - syncing...",
}
