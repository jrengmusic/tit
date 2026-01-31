package git

import (
	"fmt"
	"os"

	"tit/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// ExecuteTimeTravelMerge performs a merge of time travel changes back to original branch
// Returns: TimeTravelMergeMsg
func ExecuteTimeTravelMerge(originalBranch, timeTravelHash string) func() tea.Msg {
	return func() tea.Msg {

		// Get current working directory (repo path)
		repoPath, err := os.Getwd()
		if err != nil {
			panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
		}

		// Get original stash hash (if any) from config tracking system
		originalStashHash, hasStash := config.FindStashEntry("time_travel", repoPath)

		// Note: Dirty tree handling is done BEFORE calling this function
		// User either committed changes or discarded them, so tree is now clean

		// Checkout original branch
		checkoutResult := Execute("checkout", originalBranch)
		if !checkoutResult.Success {
			return TimeTravelMergeMsg{
				Success:        false,
				OriginalBranch: originalBranch,
				TimeTravelHash: timeTravelHash,
				Error:          fmt.Sprintf("Failed to checkout original branch: %s", checkoutResult.Stderr),
			}
		}

		// =================================================================
		// PHASE 1: Merge time travel changes FIRST (before restoring stash)
		// This provides cleaner separation - merge conflicts and stash conflicts are handled separately
		// =================================================================
		Log("Merging time travel changes...")

		mergeResult := Execute("merge", timeTravelHash)

		if !mergeResult.Success {
			// Check for conflicts
			conflictFiles, err := ListConflictedFiles()

			if err != nil || len(conflictFiles) == 0 {
				// No conflicts detected - this is a merge error (not conflicts)
				// Don't clear marker file - operation failed
				return TimeTravelMergeMsg{
					Success:        false,
					OriginalBranch: originalBranch,
					TimeTravelHash: timeTravelHash,
					Error:          fmt.Sprintf("Merge failed: %s", mergeResult.Stderr),
				}
			} else {
				// Don't clear marker file yet - conflicts need to be resolved first
				return TimeTravelMergeMsg{
					Success:          false,
					OriginalBranch:   originalBranch,
					TimeTravelHash:   timeTravelHash,
					Error:            "Merge conflicts detected",
					ConflictDetected: true,
					ConflictedFiles:  conflictFiles,
				}
			}
		}
		Log("Merge succeeded")

		// =================================================================
		// PHASE 2: Restore original stash (if any) AFTER merge succeeds
		// =================================================================
		if hasStash {
			Log("Phase 2: Restoring original work-in-progress...")
			Log(fmt.Sprintf("Applying stash %s...", originalStashHash))

			// Apply stash using hash (git stash apply accepts hash)
			applyResult := Execute("stash", "apply", originalStashHash)
			if !applyResult.Success {
				Error(fmt.Sprintf("Stash apply failed: %s", applyResult.Stderr))

				// Check for conflicts
				conflictFiles, err := ListConflictedFiles()
				if err != nil || len(conflictFiles) == 0 {
					// Stash apply failed but not due to conflicts - don't drop the stash
					Error("Stash apply failed (not conflicts) - stash will not be dropped")
					return TimeTravelMergeMsg{
						Success:        false,
						OriginalBranch: originalBranch,
						TimeTravelHash: timeTravelHash,
						Error:          fmt.Sprintf("Failed to restore stash: %s", applyResult.Stderr),
					}
				} else {
					// Stash apply had conflicts - enter conflict resolver
					Error(fmt.Sprintf("Stash apply conflicts detected in %d files", len(conflictFiles)))
					Log("Resolve conflicts to complete stash restoration")
					// Don't clear marker file yet - conflicts need to be resolved first
					return TimeTravelMergeMsg{
						Success:          false,
						OriginalBranch:   originalBranch,
						TimeTravelHash:   timeTravelHash,
						Error:            "Stash apply conflicts detected",
						ConflictDetected: true,
						ConflictedFiles:  conflictFiles,
					}
				}
			}

			Log("Stash applied successfully")

			// Drop stash using stash@{N} reference (git stash drop requires reference, not hash)
			stashRef, found := FindStashRefByHash(originalStashHash)
			if !found {
				// Stash was manually dropped - log and continue
				Log("Stash was manually dropped, skipping drop")
				// Remove from config tracking system
				config.RemoveStashEntry("time_travel", repoPath)
				return TimeTravelMergeMsg{
					Success:        true,
					OriginalBranch: originalBranch,
					TimeTravelHash: timeTravelHash,
					Error:          "",
				}
			}

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
			config.RemoveStashEntry("time_travel", repoPath)
		}

		// =================================================================
		// PHASE 3: Clean up marker file (only after all operations succeed)
		// =================================================================
		Log("Cleaning up time travel marker...")
		if err := ClearTimeTravelInfo(); err != nil {
			Error(fmt.Sprintf("Warning: Failed to clear time travel marker: %v", err))
		}

		// Don't append "successful" message here - handler will append it
		return TimeTravelMergeMsg{
			Success:        true,
			OriginalBranch: originalBranch,
			TimeTravelHash: timeTravelHash,
			Error:          "",
		}
	}
}
