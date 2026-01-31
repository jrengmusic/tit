package app

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
	// Timeline sync errors
	"timeline_sync_no_remote": "no remote configured",

	// Upstream errors
	"cannot_set_upstream_detached": "Skipped: cannot set upstream in detached HEAD state",

	// Working tree status
	"working_tree_clean": "Nothing to commit (working tree clean)",
}
