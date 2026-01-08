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
	Step              string // "init", "clone", "push", "pull", etc.
	Success           bool
	Output            string
	Error             string
	Path              string // Working directory to change to after operation
	BranchName        string // Current branch name (for remote operations)
	ConflictDetected  bool   // true if merge/rebase conflicts detected
	ConflictedFiles   []string // List of files with conflicts
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
		MessageExitBlocked:         "Exit blocked. Operation must complete or be aborted first.",
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
	"pull_conflicts":         "Merge conflicts occurred",
	"pull_failed":            "Failed to pull",
	"failed_stage_resolved":  "Failed to stage resolved files",
	"failed_commit_merge":    "Failed to commit merge",
	"failed_abort_merge":     "Failed to abort merge",
	"failed_reset_after_abort": "Failed to reset working tree after merge abort",
	"failed_determine_branch": "Error: Could not determine current branch",
	"failed_fetch_remote":    "Failed to fetch from remote",
	// Dirty pull abort path errors
	"failed_checkout_original_branch": "Error: Failed to checkout original branch",
	"failed_reset_to_original_head":   "Error: Failed to reset to original HEAD",
	"stash_reapply_failed_but_restored": "Warning: Could not reapply stash, but HEAD restored",
	// Validation errors
	"remote_url_empty_validation": "Remote URL cannot be empty",
	"remote_already_exists_validation": "Remote 'origin' already exists",
	
	// Time travel errors
	"time_travel_failed":                    "Time travel failed: %s",
	"time_travel_merge_failed":              "Time travel merge failed: %s",
	"time_travel_return_failed":             "Time travel return failed: %s",
	"failed_detect_state_after_travel":      "Failed to detect state after time travel: %v",
	"failed_detect_state_after_merge":       "Failed to detect state after time travel merge: %v",
	"failed_detect_state_after_return":      "Failed to detect state after time travel return: %v",
	"failed_get_current_branch":             "Failed to get current branch",
	"failed_stash_changes":                  "Failed to stash changes",
	"failed_get_stash_list":                 "Failed to get stash list",
	"failed_write_time_travel_info":         "Failed to write time travel info: %v",
	"failed_load_time_travel_info":          "Error: %v",
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
	"detecting_conflicts":  "Detecting conflict files...",
	"checking_out_branch":  "Checking out branch '%s'...",
	"dirty_pull_snapshot":  "Your changes have been saved",
	"dirty_pull_snapshot_saved": "Snapshot saved. Starting merge pull...",
	"dirty_pull_merge_succeeded": "Merge succeeded. Reapplying your changes...",
	"dirty_pull_rebase_succeeded": "Rebase succeeded. Reapplying your changes...",
	"dirty_pull_merge_started": "Pulling from remote (merge strategy)...",
	"dirty_pull_reapply":   "Reapplying your saved changes...",
	"dirty_pull_changes_reapplied": "Changes reapplied. Finalizing...",
	"dirty_pull_finalize":  "Finalizing dirty pull operation...",
	"conflict_detection_error": "Error detecting conflicts: %v",
	"conflict_detection_none": "Conflict detection succeeded but no conflicts found (continuing)",
	"conflicts_detected_count": "Conflicts detected in %d file(s)",
	"mark_choices_in_resolver": "Mark your choices in the resolver (SPACE to select, ENTER to continue)",
	"aborting_dirty_pull": "Aborting dirty pull...",
	"aborting_merge": "Aborting merge...",
	"merge_finalized":        "Merge completed successfully",
	"merge_aborted":          "Merge aborted - state restored",
	"abort_successful":       "Successfully aborted by user",
	"force_push_in_progress": "Force pushing to remote (overwriting remote history)...",
	"fetching_latest":        "Fetching latest from remote...",
	"removing_untracked":     "Removing untracked files and directories...",
	"failed_clean_untracked": "Warning: Failed to clean untracked files",
	"saving_changes_stash":   "Saving your changes (creating stash)...",
	"discarding_changes":     "Discarding your changes...",
	"changes_saved_stashed":  "Changes saved (stashed)",
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
	"force_push":        "Force Push Confirmation",
	"hard_reset":        "Replace Local Confirmation",
	"dirty_pull":        "Save your changes?",
	"pull_merge":        "Pull from remote?",
	"pull_merge_diverged": "Pull diverged branches?",
	"time_travel":       "Time Travel Confirmation",
}

// ConfirmationExplanations centralizes confirmation dialog explanations
var ConfirmationExplanations = map[string]string{
	"force_push": "This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
	"hard_reset": "This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
	"dirty_pull": "You have uncommitted changes. To pull, they must be temporarily saved.\n\nAfter the pull, we'll try to reapply them.\n(This may cause conflicts if the changes overlap.)",
	"pull_merge": "This will merge remote changes into your local branch.\n\nIf both branches modified the same files, conflicts may occur.\nYou'll be able to resolve them interactively.",
	"pull_merge_diverged": "Your branches have diverged (both have new commits).\n\nThis will merge remote changes into your local branch.\n\nIf both modified the same files, conflicts may occur.",
	"time_travel": "%s\n\n%s\n\nExplore in read-only mode?",
}

// ConfirmationLabels centralizes confirmation dialog button labels by action
var ConfirmationLabels = map[string][2]string{
	"force_push":        {"Force push", "Cancel"},
	"hard_reset":        {"Reset to remote", "Cancel"},
	"dirty_pull":        {"Save changes", "Discard changes"},
	"pull_merge":        {"Proceed", "Cancel"},
	"pull_merge_diverged": {"Proceed", "Cancel"},
	"time_travel":       {"Time travel", "Cancel"},
}

// FooterHints centralizes footer hint messages
var FooterHints = map[string]string{
	"mark_all_files":         "Mark all files with SPACE before continuing",
	"resolve_conflicts_help": "Resolve %d conflicted file(s) - SPACE to mark, ENTER to continue, ESC to abort",
	"error_writing_file":     "Error writing %s: %v",
	"error_staging_file":     "Error staging %s: %s",
	"already_marked_column":  "Already marked in this column",
	"marked_file_column":     "Marked: %s → %s",
	"visual_mode_active":     "-- VISUAL --",
	"copy_success":           "✓ Copied to clipboard",
	"copy_failed":            "✗ Copy failed",
	
	// Time travel messages
	"time_travel_success":           "Time travel successful. Press ESC to return to menu.",
	"time_travel_merge_success":     "Time travel merge successful. Press ESC to return to menu.",
	"time_travel_return_success":    "Time travel return successful. Press ESC to return to menu.",
	"time_traveling_status":         "Time traveling... (ESC to abort)",
	"restoration_complete":          "Restoration complete. Press ESC to continue.",
	"restoring_time_travel":         "Restoring from incomplete time travel session...",
	"step_1_discarding":             "Step 1: Discarding changes from time travel...",
	"step_2_returning":              "Step 2: Returning to %s...",
	"step_3_restoring_work":         "Step 3: Restoring original uncommitted work...",
	"step_4_cleaning_marker":        "Step 4: Cleaning up time travel marker...",
	"original_work_restored":        "Original work restored",
	"restoration_error":             "Restoration error: %s",
	"marker_corrupted":              "Marker file corrupted. Cleaned up.",
	"warning_discard_changes":       "Warning: Could not discard working tree changes",
	"warning_remove_untracked":      "Warning: Could not remove untracked files",
	"error_checkout_branch":         "Error: Failed to checkout %s",
	"warning_restore_work":          "Warning: Could not restore original work (may have been lost)",
	"warning_cleanup_stash":         "Warning: Could not clean up stash entry",
	"warning_remove_marker":         "Warning: Could not remove marker: %v",
	"error_detect_state":            "Error detecting state after restoration: %v",
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
	"working_tree_clean":  "Your files match the remote.",
	"working_tree_dirty":  "You have uncommitted changes.",
	"timeline_in_sync":    "Local and remote are in sync.",
	"timeline_ahead":      "You have %d unsynced commit(s).",
	"timeline_behind":     "The remote has %d new commit(s).",
	"timeline_diverged":   "Both have new commits. Ahead %d, Behind %d.",
	"operation_normal":    "Repository ready for operations.",
	"operation_not_repo":  "Not a git repository.",
	"operation_conflicted": "Conflicts must be resolved.",
	"operation_merging":   "Merge in progress.",
	"operation_rebasing":  "Rebase in progress.",
	"operation_dirty_op":  "Operation interrupted by uncommitted changes.",
	"operation_time_travel": "Exploring commit %s from %s.",
}
