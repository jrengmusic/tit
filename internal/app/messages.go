package app

import (
	"time"

	"tit/internal/ui"
)

// TickMsg is a custom message for quit confirmation timeout
type TickMsg time.Time

// ClearTickMsg is a custom message for input clear confirmation timeout
type ClearTickMsg time.Time

// GitOperationMsg represents the result of a git operation
type GitOperationMsg struct {
	Step             string // "init", "clone", "push", "pull", etc.
	Success          bool
	Output           string
	Error            string
	Path             string   // Working directory to change to after operation
	BranchName       string   // Current branch name (for remote operations)
	ConflictDetected bool     // true if merge/rebase conflicts detected
	ConflictedFiles  []string // List of files with conflicts
}

// RestoreTimeTravelMsg signals completion of time travel restoration (Phase 0)
type RestoreTimeTravelMsg struct {
	Success bool
	Error   string
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

// OutputRefreshMsg triggers UI re-render to show updated console output
// Sent periodically during long-running operations to display streaming output
type OutputRefreshMsg struct{}

// RewindMsg represents the result of a git reset --hard operation
type RewindMsg struct {
	Commit  string // hash
	Success bool
	Output  string
	Error   string
}

// RemoteFetchMsg signals completion of background git fetch on startup
type RemoteFetchMsg struct {
	Success bool
	Error   string
}

// CacheProgressMsg reports cache building progress (for UI updates)
type CacheProgressMsg struct {
	CacheType string // "metadata" or "diffs"
	Current   int    // Current item processed
	Total     int    // Total items to process
	Complete  bool   // true when cache is fully built
}

// CacheRefreshTickMsg triggers periodic UI refresh during cache building
// Sent every 100ms to update spinner animation and progress counter
type CacheRefreshTickMsg struct{}

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
	messages := map[FooterMessageType]string{
		MessageNone:                "",
		MessageCtrlCConfirm:        "Press Ctrl+C again to quit (3s timeout)",
		MessageEscClearConfirm:     "Press ESC again to clear input (3s timeout)",
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

// ========================================
// Message Domain: Input (Prompts + Hints)
// ========================================

// InputMessage pairs a prompt with its hint for a single input field
type InputMessage struct {
	Prompt string
	Hint   string
}

// InputMessages centralizes input-related messages by domain
// Replaces old InputPrompts + InputHints maps
var InputMessages = map[string]InputMessage{
	"clone_url": {
		Prompt: "Repository URL:",
		Hint:   "Enter git repository URL (https or git+ssh)",
	},
	"remote_url": {
		Prompt: "Remote URL:",
		Hint:   "Enter git repository URL and press Enter",
	},
	"commit_message": {
		Prompt: "Commit message:",
		Hint:   "Enter message and press Enter",
	},
	"subdir_name": {
		Prompt: "Subdirectory name:",
		Hint:   "Enter new directory name",
	},
	"init_branch_name": {
		Prompt: "Initial branch name:",
		Hint:   "Enter branch name (default: main), press Enter to initialize",
	},
	"init_subdir_name": {
		Prompt: "Subdirectory name:",
		Hint:   "Enter subdirectory name for new repository",
	},
	"dirty_pull_save": {
		Prompt: "Save and continue with dirty pull",
		Hint:   "Save changes before pulling",
	},
	"dirty_pull_discard": {
		Prompt: "Discard changes and pull",
		Hint:   "Discard changes before pulling",
	},
}

// ErrorMessages centralizes error messages
var ErrorMessages = map[string]string{
	"cwd_read_failed":          "Failed to get current directory",
	"operation_failed":         "Operation failed",
	"branch_name_empty":        "Branch name cannot be empty",
	"commit_message_empty":     "Commit message cannot be empty",
	"remote_url_empty":         "Remote URL cannot be empty",
	"remote_already_exists":    "Remote 'origin' already exists",
	"failed_create_dir":        "Failed to create directory: %v",
	"failed_change_dir":        "Failed to change to directory: %v",
	"failed_detect_state":      "Failed to detect git state: %v",
	"failed_cd_into":           "Failed to cd into %s: %v",
	"failed_checkout_branch":   "Failed to checkout branch '%s'",
	"pull_conflicts":           "Merge conflicts occurred",
	"pull_failed":              "Failed to pull",
	"failed_stage_resolved":    "Failed to stage resolved files",
	"failed_commit_merge":      "Failed to commit merge",
	"failed_abort_merge":       "Failed to abort merge",
	"failed_reset_after_abort": "Failed to reset working tree after merge abort",
	"failed_determine_branch":  "Error: Could not determine current branch",
	"failed_fetch_remote":      "Failed to fetch from remote",
	// Dirty pull abort path errors
	"failed_checkout_original_branch":   "Error: Failed to checkout original branch",
	"failed_reset_to_original_head":     "Error: Failed to reset to original HEAD",
	"stash_reapply_failed_but_restored": "Warning: Could not reapply stash, but HEAD restored",
	// Validation errors
	"remote_url_empty_validation":      "Remote URL cannot be empty",
	"remote_already_exists_validation": "Remote 'origin' already exists",

	// Time travel errors
	"time_travel_failed":               "Time travel failed: %s",
	"time_travel_merge_failed":         "Time travel merge failed: %s",
	"time_travel_return_failed":        "Time travel return failed: %s",
	"failed_detect_state_after_travel": "Failed to detect state after time travel: %v",
	"failed_detect_state_after_merge":  "Failed to detect state after time travel merge: %v",
	"failed_detect_state_after_return": "Failed to detect state after time travel return: %v",
	"failed_get_current_branch":        "Failed to get current branch",
	"failed_stash_changes":             "Failed to stash changes",
	"failed_get_stash_list":            "Failed to get stash list",
	"failed_write_time_travel_info":    "Failed to write time travel info: %v",
	"failed_load_time_travel_info":     "Error: %v",

	// Rewind (reset --hard) errors
	"rewind_commit_hash_empty": "Commit hash cannot be empty",
	"rewind_failed":            "Reset failed: %s",
}

// OutputMessages centralizes operation success messages
var OutputMessages = map[string]string{
	"remote_added":                 "Remote added",
	"fetch_completed":              "Fetch completed",
	"pushed_successfully":          "Pushed successfully",
	"pulled_successfully":          "Pulled successfully",
	"initializing_repo":            "Initializing repository...",
	"fetching_remote":              "Fetching from remote...",
	"setting_upstream":             "Setting upstream tracking...",
	"detecting_conflicts":          "Detecting conflict files...",
	"checking_out_branch":          "Checking out branch '%s'...",
	"dirty_pull_snapshot":          "Your changes have been saved",
	"dirty_pull_snapshot_saved":    "Snapshot saved. Starting merge pull...",
	"dirty_pull_merge_succeeded":   "Merge succeeded. Reapplying your changes...",
	"dirty_pull_rebase_succeeded":  "Rebase succeeded. Reapplying your changes...",
	"dirty_pull_merge_started":     "Pulling from remote (merge strategy)...",
	"dirty_pull_reapply":           "Reapplying your saved changes...",
	"dirty_pull_changes_reapplied": "Changes reapplied. Finalizing...",
	"dirty_pull_finalize":          "Finalizing dirty pull operation...",
	"conflict_detection_error":     "Error detecting conflicts: %v",
	"conflict_detection_none":      "Conflict detection succeeded but no conflicts found (continuing)",
	"conflicts_detected_count":     "Conflicts detected in %d file(s)",
	"mark_choices_in_resolver":     "Mark your choices in the resolver (SPACE to select, ENTER to continue)",
	"aborting_dirty_pull":          "Aborting dirty pull...",
	"aborting_merge":               "Aborting merge...",
	"merge_finalized":              "Merge completed successfully",
	"merge_aborted":                "Merge aborted - state restored",
	"abort_successful":             "Successfully aborted by user",
	"force_push_in_progress":       "Force pushing to remote (overwriting remote history)...",
	"fetching_latest":              "Fetching latest from remote...",
	"removing_untracked":           "Removing untracked files and directories...",
	"failed_clean_untracked":       "Warning: Failed to clean untracked files",
	"saving_changes_stash":         "Saving your changes (creating stash)...",
	"discarding_changes":           "Discarding your changes...",
	"changes_saved_stashed":        "Changes saved (stashed)",
	// Dirty pull operation phases
	"changes_discarded":                 "Changes discarded",
	"merge_conflicts_detected":          "Merge conflicts detected",
	"merge_completed":                   "Merge completed",
	"reapplying_changes":                "Reapplying your changes...",
	"stash_apply_conflicts_detected":    "Conflicts detected while reapplying changes",
	"changes_reapplied":                 "Changes reapplied",
	"dirty_pull_finalize_started":       "Finalizing dirty pull operation...",
	"stash_drop_failed_warning":         "Warning: Failed to drop stash (manual cleanup may be needed)",
	"dirty_pull_completed_successfully": "Dirty pull completed successfully",
	"dirty_pull_aborting":               "Aborting dirty pull and restoring original state...",
	"original_state_restored":           "Original state restored",

	// Rewind (reset --hard) operations
	"rewind_resetting": "Resetting to commit %s...",
	"rewind_completed": "Rewind completed successfully",
}

// ========================================
// Message Domain: Confirmation Dialogs
// ========================================

// ConfirmationMessage pairs all confirmation dialog components
type ConfirmationMessage struct {
	Title       string
	Explanation string
	YesLabel    string // Button text for YES/confirm action
	NoLabel     string // Button text for NO/reject action
}

// ConfirmationMessages centralizes all confirmation dialog messages by domain
// Replaces old ConfirmationTitles + ConfirmationExplanations + ConfirmationLabels
var ConfirmationMessages = map[string]ConfirmationMessage{
	"force_push": {
		Title:       "Force Push Confirmation",
		Explanation: "This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
		YesLabel:    "Force push",
		NoLabel:     "Cancel",
	},
	"hard_reset": {
		Title:       "Replace Local Confirmation",
		Explanation: "This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
		YesLabel:    "Reset to remote",
		NoLabel:     "Cancel",
	},
	"dirty_pull": {
		Title:       "Save your changes?",
		Explanation: "You have uncommitted changes. To pull, they must be temporarily saved.\n\nAfter the pull, we'll try to reapply them.\n(This may cause conflicts if the changes overlap.)",
		YesLabel:    "Save changes",
		NoLabel:     "Discard changes",
	},
	"pull_merge": {
		Title:       "Pull from remote?",
		Explanation: "This will merge remote changes into your local branch.\n\nIf both branches modified the same files, conflicts may occur.\nYou'll be able to resolve them interactively.",
		YesLabel:    "Proceed",
		NoLabel:     "Cancel",
	},
	"pull_merge_diverged": {
		Title:       "Pull diverged branches?",
		Explanation: "Your branches have diverged (both have new commits).\n\nThis will merge remote changes into your local branch.\n\nIf both modified the same files, conflicts may occur.",
		YesLabel:    "Proceed",
		NoLabel:     "Cancel",
	},
	"time_travel": {
		Title:       "Time Travel Confirmation",
		Explanation: "%s\n\n%s\n\nExplore in read-only mode?",
		YesLabel:    "Time travel",
		NoLabel:     "Cancel",
	},
	"time_travel_return": {
		Title:       "Return to main without merge?",
		Explanation: "Any changes you made while time traveling will be STASHED (not discarded).\n\nYour original work (if any) will be restored.\n\nUse 'git stash apply stash@{0}' later to restore time travel changes.",
		YesLabel:    "Return to main",
		NoLabel:     "Cancel",
	},
	"time_travel_merge": {
		Title:       "Merge and return to main?",
		Explanation: "This will merge time travel changes back to main.\n\nConflicts may occur if the changes overlap.\n\nNote: Any uncommitted changes will be stashed first, then restored after merge.",
		YesLabel:    "Merge & return",
		NoLabel:     "Cancel",
	},
	"time_travel_merge_dirty": {
		Title:       "Uncommitted Changes",
		Explanation: "You modified files during time travel.\n\nCommit them and merge to main, or discard them?",
		YesLabel:    "Commit & merge",
		NoLabel:     "Discard",
	},
	"time_travel_return_dirty": {
		Title:       "Uncommitted Changes",
		Explanation: "You modified files during time travel.\n\nChanges will be discarded when returning to main.",
		YesLabel:    "Discard & return",
		NoLabel:     "Cancel",
	},
	"rewind": {
		Title:       "DESTRUCTIVE OPERATION",
		Explanation: "This will discard all commits after %s.\nAny uncommitted changes will be lost.\n\nAre you sure you want to continue?",
		YesLabel:    "Rewind",
		NoLabel:     "Cancel",
	},
}

// FooterShortcut represents a single keyboard shortcut hint
// NOTE: This is a duplicate of ui.FooterShortcut for legacy compatibility.
// New code should use ui.FooterShortcut and FooterHintShortcuts.
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
		{Key: "Esc", Desc: "abort"},
	},
	"console_complete": {
		{Key: "Esc", Desc: "back"},
	},

	// Input
	"input_single": {
		{Key: "Enter", Desc: "submit"},
		{Key: "Esc", Desc: "cancel"},
	},
	"input_multi": {
		{Key: "Enter", Desc: "submit"},
		{Key: "Esc", Desc: "cancel"},
	},

	// Confirmation
	"confirmation": {
		{Key: "←→", Desc: "select"},
		{Key: "Enter", Desc: "confirm"},
		{Key: "Esc", Desc: "cancel"},
	},
}

// LegacyFooterHints maintains backward compatibility for string-based hints
var LegacyFooterHints = map[string]string{
	"mark_all_files":         "Mark all files with SPACE before continuing",
	"resolve_conflicts_help": "Resolve %d conflicted file(s) - SPACE to mark, ENTER to continue, ESC to abort",
	"error_writing_file":     "Error writing %s: %v",
	"error_staging_file":     "Error staging %s: %s",
	"already_marked_column":  "Already marked in this column",
	"marked_file_column":     "Marked: %s → %s",
	"visual_mode_active":     "-- VISUAL --",
	"copy_success":           "✓ Copied to clipboard",
	"copy_failed":            "✗ Copy failed",
	"no_commits_yet":         "No commits yet. Create an initial commit to enable sync actions.",

	// Time travel messages
	"time_travel_success":        "Time travel successful. Press ESC to return to menu.",
	"time_travel_merge_success":  "Time travel merge successful. Press ESC to return to menu.",
	"time_travel_return_success": "Time travel return successful. Press ESC to return to menu.",
	"time_traveling_status":      "Time traveling... (ESC to abort)",
	"restoration_complete":       "Restoration complete. Press ESC to continue.",
	"restoring_time_travel":      "Restoring from incomplete time travel session...",
	"step_1_discarding":          "Step 1: Discarding changes from time travel...",
	"step_2_returning":           "Step 2: Returning to %s...",
	"step_3_restoring_work":      "Step 3: Restoring original uncommitted work...",
	"step_4_cleaning_marker":     "Step 4: Cleaning up time travel marker...",
	"original_work_restored":     "Original work restored",
	"restoration_error":          "Restoration error: %s",
	"marker_corrupted":           "Marker file corrupted. Cleaned up.",
	"warning_discard_changes":    "Warning: Could not discard working tree changes",
	"warning_remove_untracked":   "Warning: Could not remove untracked files",
	"error_checkout_branch":      "Error: Failed to checkout %s",
	"warning_restore_work":       "Warning: Could not restore original work (may have been lost)",
	"warning_cleanup_stash":      "Warning: Could not clean up stash entry",
	"warning_remove_marker":      "Warning: Could not remove marker: %v",
	"error_detect_state":         "Error detecting state after restoration: %v",
}

// DialogMessages centralizes dialog box content (titles + explanations)
var DialogMessages = map[string][2]string{
	"nested_repo": {
		"Nested Repository Detected",
		"The directory '%s' is inside another git repository.\n\nThis may cause confusion. Would you like to initialize in a subdirectory instead?",
	},
	"force_push_nested": {
		"Force Push Confirmation",
		"This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
	},
	"hard_reset_nested": {
		"Hard Reset Confirmation",
		"This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
	},
}

// StateDescriptions centralizes git state display descriptions
var StateDescriptions = map[string]string{
	"working_tree_clean":    "Your files match the remote.",
	"working_tree_dirty":    "You have uncommitted changes.",
	"timeline_in_sync":      "Local and remote are in sync.",
	"timeline_ahead":        "You have %d unsynced commit(s).",
	"timeline_behind":       "The remote has %d new commit(s).",
	"timeline_diverged":     "Both have new commits. Ahead %d, Behind %d.",
	"operation_normal":      "Repository ready for operations.",
	"operation_not_repo":    "Not a git repository.",
	"operation_conflicted":  "Conflicts must be resolved.",
	"operation_merging":     "Merge in progress.",
	"operation_rebasing":    "Rebase in progress.",
	"operation_dirty_op":    "Operation interrupted by uncommitted changes.",
	"operation_time_travel": "Exploring commit %s from %s.",
}
