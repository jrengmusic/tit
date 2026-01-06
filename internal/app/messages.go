package app

import (
	"time"
)

// TickMsg is a custom message for quit confirmation timeout
type TickMsg time.Time

// ClearTickMsg is a custom message for input clear confirmation timeout
type ClearTickMsg time.Time

// GitOperationMsg represents the result of a git operation
type GitOperationMsg struct {
	Step       string // "init", "clone", "push", "pull", etc.
	Success    bool
	Output     string
	Error      string
	Path       string // Working directory to change to after operation
	BranchName string // Current branch name (for remote operations)
}

// GitOperationCompleteMsg signals that a git operation completed
type GitOperationCompleteMsg struct {
	Success bool
	Output  string
	Error   string
}

// InputSubmittedMsg signals that user submitted input
type InputSubmittedMsg struct {
	Action string
	Value  string
}

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
)

// GetFooterMessageText returns display text for a message type
func GetFooterMessageText(msgType FooterMessageType) string {
	messages := map[FooterMessageType]string{
		MessageNone:             "",
		MessageCtrlCConfirm:     "Press Ctrl+C again to quit (3s timeout)",
		MessageEscClearConfirm:  "Press ESC again to clear input (3s timeout)",
		MessageInit:             "Initializing repository... (ESC to abort)",
		MessageClone:            "Cloning repository... (ESC to abort)",
		MessageCommit:           "Committing changes... (ESC to abort)",
		MessagePush:             "Pushing to remote... (ESC to abort)",
		MessagePull:             "Pulling from remote... (ESC to abort)",
		MessageAddRemote:        "Adding remote and fetching... (ESC to abort)",
		MessageResolveConflicts: "Resolving conflicts...",
		MessageOperationComplete: "Press ESC to return to menu",
		MessageOperationFailed:   "Failed. Press ESC to return.",
		MessageOperationInProgress: "Operation in progress. Please wait for completion.",
		MessageOperationAborting:   "Aborting operation. Please wait...",
	}

	if msg, exists := messages[msgType]; exists {
		return msg
	}
	return ""
}

// ========================================
// Input Prompts & Hints (SSOT)
// ========================================

// InputPrompts centralizes all user-facing input prompts
var InputPrompts = map[string]string{
	"clone_url":        "Repository URL:",
	"remote_url":       "Remote URL:",
	"commit_message":   "Commit message:",
	"subdir_name":      "Subdirectory name:",
	"init_branch_name": "Initial branch name:",
	"init_subdir_name": "Subdirectory name:",
	"dirty_pull_save":  "Save and continue with dirty pull",
	"dirty_pull_discard": "Discard changes and pull",
}

// InputHints centralizes all user-facing input hints
var InputHints = map[string]string{
	"clone_url":        "Enter git repository URL (https or git+ssh)",
	"remote_url":       "Enter git repository URL and press Enter",
	"commit_message":   "Enter message and press Enter",
	"subdir_name":      "Enter new directory name",
	"init_branch_name": "Enter branch name (default: main), press Enter to initialize",
	"init_subdir_name": "Enter subdirectory name for new repository",
	"dirty_pull_save":  "Save changes before pulling",
	"dirty_pull_discard": "Discard changes before pulling",
}

// ErrorMessages centralizes error messages
var ErrorMessages = map[string]string{
	"cwd_read_failed":        "Failed to get current directory",
	"operation_failed":       "Operation failed",
	"branch_name_empty":      "Branch name cannot be empty",
	"commit_message_empty":   "Commit message cannot be empty",
	"remote_url_empty":       "Remote URL cannot be empty",
	"remote_already_exists":  "Remote 'origin' already exists",
	"failed_create_dir":      "Failed to create directory: %v",
	"failed_change_dir":      "Failed to change to directory: %v",
	"failed_detect_state":    "Failed to detect git state: %v",
	"failed_cd_into":         "Failed to cd into %s: %v",
	"failed_checkout_branch": "Failed to checkout branch '%s'",
}

// OutputMessages centralizes operation success messages
var OutputMessages = map[string]string{
	"remote_added":         "Remote added",
	"fetch_completed":      "Fetch completed",
	"pushed_successfully":  "Pushed successfully",
	"pulled_successfully":  "Pulled successfully",
	"initializing_repo":    "Initializing repository...",
	"fetching_remote":      "Fetching from remote...",
	"setting_upstream":     "Setting upstream tracking...",
	"checking_out_branch":  "Checking out branch '%s'...",
	"dirty_pull_snapshot":  "Your changes have been saved",
	"dirty_pull_merge_started": "Pulling from remote (merge strategy)...",
	"dirty_pull_reapply":   "Reapplying your saved changes...",
	"dirty_pull_finalize":  "Finalizing dirty pull operation...",
}

// ButtonLabels centralizes confirmation dialog button text
var ButtonLabels = map[string]string{
	"continue":   "Yes, continue",
	"cancel":     "No, cancel",
	"force_push": "Force push",
	"reset":      "Reset",
	"ok":         "OK",
}

// ConfirmationTitles centralizes confirmation dialog titles
var ConfirmationTitles = map[string]string{
	"force_push":   "Force Push Confirmation",
	"hard_reset":   "Replace Local Confirmation",
	"dirty_pull":   "Save your changes?",
}

// ConfirmationExplanations centralizes confirmation dialog explanations
var ConfirmationExplanations = map[string]string{
	"force_push": "This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
	"hard_reset": "This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
	"dirty_pull": "You have uncommitted changes. To pull, they must be temporarily saved.\n\nAfter the pull, we'll try to reapply them.\n(This may cause conflicts if the changes overlap.)",
}

// ConfirmationLabels centralizes confirmation dialog button labels by action
var ConfirmationLabels = map[string][2]string{
	"force_push": {"Force push", "Cancel"},
	"hard_reset": {"Reset to remote", "Cancel"},
	"dirty_pull": {"Save changes", "Discard changes"},
}
