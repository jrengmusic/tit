package app

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

// ConsoleMessages centralizes all console output messages
var ConsoleMessages = map[string]string{
	// Time travel restoration
	"restoring_time_travel":    "Restoring from incomplete time travel session...",
	"marker_corrupted":         "Marker file corrupted. Cleaned up.",
	"step_1_discarding":        "Step 1: Discarding changes from time travel...",
	"warning_discard_changes":  "Warning: Could not discard working tree changes",
	"warning_remove_untracked": "Warning: Could not remove untracked files",
	"step_2_returning":         "Step 2: Returning to %s...",
	"error_checkout_branch":    "Error: Failed to checkout %s",
	"step_3_restoring_work":    "Step 3: Restoring original uncommitted work...",
	"warning_restore_work":     "Warning: Could not restore original work (may have been lost)",
	"original_work_restored":   "Original work restored",
	"warning_cleanup_stash":    "Warning: Could not clean up stash entry",
	"step_4_cleaning_marker":   "Step 4: Cleaning up time travel marker...",
	"warning_remove_marker":    "Warning: Could not remove marker: %v",
	"restoration_complete":     "Restoration complete. Press ESC to continue.",
	"restoration_error":        "Restoration error: %s",
	"error_detect_state":       "Error detecting state after restoration: %v",

	// Time travel success
	"time_travel_success":        "Time travel successful. Press ESC to return to menu.",
	"time_travel_merge_success":  "Time travel merge successful. Press ESC to return to menu.",
	"time_travel_return_success": "Time travel return successful. Press ESC to return to menu.",
	"time_traveling_status":      "Time traveling... (ESC to abort)",

	// Conflict resolver
	"resolve_conflicts_help": "Resolve %d conflicted file(s) - SPACE to mark, ENTER to continue, ESC to abort",
	"already_marked_column":  "Already marked in this column",
	"marked_file_column":     "Marked: %s → %s",
	"mark_all_files":         "Mark all files with SPACE before continuing",
	"error_writing_file":     "Error writing %s: %v",
	"error_staging_file":     "Error staging %s: %s",

	// History mode
	"no_commits_yet": "No commits yet. Create an initial commit to enable sync actions.",

	// File history visual mode
	"visual_mode_active": "-- VISUAL --",

	// Clipboard
	"copy_success": "✓ Copied to clipboard",
	"copy_failed":  "✗ Copy failed",
}

// StateDescriptions centralizes git state display descriptions
var StateDescriptions = map[string]string{
	// Working Tree (2 descriptions)
	"working_tree_clean": "No local changes",
	"working_tree_dirty": "Local changes present",

	// Timeline (4 descriptions)
	"timeline_in_sync":  "In sync with remote",
	"timeline_ahead":    "%d commit(s) ahead",
	"timeline_behind":   "%d commit(s) behind",
	"timeline_diverged": "%d↑ %d↓",

	// Operation (7 descriptions)
	"operation_normal":      "Ready",
	"operation_not_repo":    "Not a repository",
	"operation_conflicted":  "Conflicts detected",
	"operation_merging":     "Merge in progress",
	"operation_rebasing":    "Rebase in progress",
	"operation_dirty_op":    "Operation started with local changes",
	"operation_time_travel": "Viewing commit %s (%s)",
}
