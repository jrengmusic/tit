package app

import (
	"context"
	"os"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// cmdFinalizeTimeTravelMerge finalizes time travel merge after conflict resolution
// Commits resolved conflicts and clears time travel marker file
func (a *Application) cmdFinalizeTimeTravelMerge() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files (should already be staged, but be safe)
		buffer.Append("Staging resolved files...", ui.TypeStatus)
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append("Failed to stage resolved files", ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeTravelMerge,
				Success: false,
				Error:   "Failed to stage resolved files",
			}
		}

		// Check if there are any staged changes
		// (User might have chosen "keep main" for all conflicts, resulting in no changes)
		diffResult := git.Execute("diff", "--cached", "--quiet")
		hasStagedChanges := !diffResult.Success // --quiet returns non-zero if there are differences

		if hasStagedChanges {
			// Commit the merge
			buffer.Append("Committing resolved conflicts...", ui.TypeStatus)
			result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Merge time travel changes (conflicts resolved)")
			if !result.Success {
				buffer.Append("Failed to commit merge", ui.TypeStderr)
				return GitOperationMsg{
					Step:    OpFinalizeTravelMerge,
					Success: false,
					Error:   "Failed to commit merge",
				}
			}
		} else {
			// No changes to commit (user chose to keep main's version for all conflicts)
			buffer.Append("No changes to commit (kept current branch state)", ui.TypeStatus)

			// CRITICAL: Clean up git merge state
			// When no commit is made, MERGE_HEAD file remains and git thinks merge is incomplete
			buffer.Append("Cleaning up merge state...", ui.TypeStatus)
			os.Remove(".git/MERGE_HEAD")
			os.Remove(".git/MERGE_MODE")
			os.Remove(".git/MERGE_MSG")
			os.Remove(".git/AUTO_MERGE")
			buffer.Append("Merge state cleaned (no-op merge)", ui.TypeStatus)
		}

		// Clear time travel marker file
		buffer.Append("Cleaning up time travel marker...", ui.TypeStatus)
		if err := git.ClearTimeTravelInfo(); err != nil {
			buffer.Append("Warning: Failed to clear time travel marker", ui.TypeStderr)
		} else {
			buffer.Append("Time travel marker cleared", ui.TypeStatus)
		}

		buffer.Append("Time travel merge completed successfully!", ui.TypeStatus)
		return GitOperationMsg{
			Step:    OpFinalizeTravelMerge,
			Success: true,
			Output:  "Time travel merge finalized",
		}
	}
}

// cmdFinalizeTimeTravelReturn finalizes time travel return after conflict resolution
// Commits resolved stash conflicts and clears time travel marker file
func (a *Application) cmdFinalizeTimeTravelReturn() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		buffer := ui.GetBuffer()

		// Stage all resolved files (should already be staged, but be safe)
		buffer.Append("Staging resolved files...", ui.TypeStatus)
		result := git.ExecuteWithStreaming(ctx, "add", "-A")
		if !result.Success {
			buffer.Append("Failed to stage resolved files", ui.TypeStderr)
			return GitOperationMsg{
				Step:    OpFinalizeTravelReturn,
				Success: false,
				Error:   "Failed to stage resolved files",
			}
		}

		// Check if there are any staged changes
		diffResult := git.Execute("diff", "--cached", "--quiet")
		hasStagedChanges := !diffResult.Success

		if hasStagedChanges {
			// Commit the resolved stash
			buffer.Append("Committing resolved work-in-progress...", ui.TypeStatus)
			result = git.ExecuteWithStreaming(ctx, "commit", "-m", "Restore work-in-progress (conflicts resolved)")
			if !result.Success {
				buffer.Append("Failed to commit", ui.TypeStderr)
				return GitOperationMsg{
					Step:    OpFinalizeTravelReturn,
					Success: false,
					Error:   "Failed to commit resolved work",
				}
			}
		} else {
			buffer.Append("No changes to commit", ui.TypeStatus)
		}

		// Drop the time travel stash
		buffer.Append("Dropping time travel stash...", ui.TypeStatus)
		git.ExecuteWithStreaming(ctx, "stash", "drop")

		// Clear time travel marker file
		buffer.Append("Cleaning up time travel marker...", ui.TypeStatus)
		if err := git.ClearTimeTravelInfo(); err != nil {
			buffer.Append("Warning: Failed to clear time travel marker", ui.TypeStderr)
		}

		buffer.Append("Time travel return completed successfully!", ui.TypeStatus)
		return GitOperationMsg{
			Step:    OpFinalizeTravelReturn,
			Success: true,
			Output:  "Time travel return finalized",
		}
	}
}
