package git

import (
	"fmt"
	"os"

	"tit/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// ExecuteTimeTravelReturn returns from time travel without merging changes
// Returns: TimeTravelReturnMsg
func ExecuteTimeTravelReturn(originalBranch string) func() tea.Msg {
	return func() tea.Msg {
		Log(fmt.Sprintf("Returning from time travel to %s...", originalBranch))

		// Get current working directory (repo path)
		repoPath, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
		}

		// Get original stash hash (if any) from config tracking system
		originalStashHash, hasStash := config.FindStashEntry("time_travel", repoPath)

		// Note: Dirty tree handling is done BEFORE calling this function
		// User discarded changes, so tree is now clean

		// Checkout original branch
		Log(fmt.Sprintf("Checking out %s...", originalBranch))
		checkoutResult := Execute("checkout", originalBranch)
		if !checkoutResult.Success {
			Error(fmt.Sprintf("Error checking out original branch: %s", checkoutResult.Stderr))
			return TimeTravelReturnMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		Log(fmt.Sprintf("Checked out %s successfully", originalBranch))

		// If we had an original stash from BEFORE time travel, restore it
		if hasStash {
			Log(fmt.Sprintf("Restoring original stash %s...", originalStashHash))

			// Apply stash using hash (git stash apply accepts hash)
			applyResult := Execute("stash", "apply", originalStashHash)
			if !applyResult.Success {
				Error(fmt.Sprintf("Stash apply failed: %s", applyResult.Stderr))

				// Check for conflicts
				conflictFiles, err := ListConflictedFiles()
				if err != nil || len(conflictFiles) == 0 {
					// Stash apply failed but not due to conflicts - don't drop the stash
					Error("Stash apply failed (not conflicts) - stash will not be dropped")
					return TimeTravelReturnMsg{
						Success:        false,
						OriginalBranch: originalBranch,
						Error:          fmt.Sprintf("Failed to restore stash: %s", applyResult.Stderr),
					}
				} else {
					// Stash apply had conflicts - enter conflict resolver
					Error(fmt.Sprintf("Stash apply conflicts detected in %d files", len(conflictFiles)))
					return TimeTravelReturnMsg{
						Success:          false,
						OriginalBranch:   originalBranch,
						Error:            "Stash apply conflicts detected",
						ConflictDetected: true,
						ConflictedFiles:  conflictFiles,
					}
				}
			}

			Log("Stash applied successfully")

			// Drop stash using stash@{N} reference (git stash drop requires reference, not hash)
			Log("[DEBUG] Finding stash reference by hash...")
			stashRef, found := FindStashRefByHash(originalStashHash)
			if !found {
				// Stash was manually dropped - log and continue
				Log("Stash was manually dropped, skipping drop")
				// Remove from config tracking system
				Log("[DEBUG] Removing stash entry from config...")
				config.RemoveStashEntry("time_travel", repoPath)
				Log("[DEBUG] Stash entry removed from config")
				return TimeTravelReturnMsg{
					Success:        true,
					OriginalBranch: originalBranch,
					Error:          "",
				}
			}
			Log(fmt.Sprintf("[DEBUG] Found stash reference: %s", stashRef))

			Log(fmt.Sprintf("Dropping stash %s...", stashRef))
			dropResult := Execute("stash", "drop", stashRef)
			if !dropResult.Success {
				// Don't panic - just warn user
				Error(fmt.Sprintf("Warning: Failed to drop stash %s: %s", stashRef, dropResult.Stderr))
				Log("You may need to manually drop the stash later")
			} else {
				Log("Stash dropped successfully")
			}

			// Remove from config tracking system
			Log("[DEBUG] Removing stash entry from config...")
			config.RemoveStashEntry("time_travel", repoPath)
			Log("[DEBUG] Stash entry removed from config")
		}

		// Clean up time travel marker file (needed for state detection)
		if err := ClearTimeTravelInfo(); err != nil {
			Error(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err))
		}

		// Don't append "successful" message here - handler will append it
		return TimeTravelReturnMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			Error:          "",
		}
	}
}
