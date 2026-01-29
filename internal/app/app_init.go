package app

import (
	"fmt"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *Application) RestoreFromTimeTravel() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()
		buffer.Append(ConsoleMessages["restoring_time_travel"], ui.TypeStatus)

		// Load time travel info
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			// Use standardized error logging (PATTERN: ErrorWarn for recovery paths)
			a.LogError(ErrorConfig{
				Level:      ErrorWarn,
				Message:    "Failed to load time travel info",
				InnerError: err,
				BufferLine: fmt.Sprintf(ErrorMessages["failed_load_time_travel_info"], err),
				FooterLine: "Failed to restore time travel state",
			})
			return RestoreTimeTravelMsg{
				Success: false,
				Error:   err.Error(),
			}
		}

		if ttInfo == nil {
			// Marker exists but no info - cleanup and continue
			git.ClearTimeTravelInfo()
			buffer.Append(ConsoleMessages["marker_corrupted"], ui.TypeStatus)
			return RestoreTimeTravelMsg{
				Success: true,
				Error:   "",
			}
		}

		// Step 1: Discard any changes made during time travel
		buffer.Append(ConsoleMessages["step_1_discarding"], ui.TypeStatus)
		// Use reset --hard instead of checkout . (works with uncommitted changes)
		resetResult := git.Execute("reset", "--hard", "HEAD")
		if !resetResult.Success {
			buffer.Append(ConsoleMessages["warning_discard_changes"], ui.TypeStatus)
		}

		cleanResult := git.Execute("clean", "-fd")
		if !cleanResult.Success {
			buffer.Append(ConsoleMessages["warning_remove_untracked"], ui.TypeStatus)
		}

		// Step 2: Return to original branch
		buffer.Append(fmt.Sprintf("Step 2: Returning to %s...", ttInfo.OriginalBranch), ui.TypeStatus)
		checkoutBranchResult := git.Execute("checkout", ttInfo.OriginalBranch)
		if !checkoutBranchResult.Success {
			buffer.Append(fmt.Sprintf("Error: Failed to checkout %s", ttInfo.OriginalBranch), ui.TypeStderr)
			return RestoreTimeTravelMsg{
				Success: false,
				Error:   "Failed to checkout original branch",
			}
		}

		// Step 3: Restore original stashed work if any
		if ttInfo.OriginalStashID != "" {
			buffer.Append(ConsoleMessages["step_3_restoring_work"], ui.TypeStatus)
			applyResult := git.Execute("stash", "apply", ttInfo.OriginalStashID)
			if !applyResult.Success {
				buffer.Append("Warning: Could not restore original work (may have been lost)", ui.TypeStatus)
			} else {
				buffer.Append(ConsoleMessages["original_work_restored"], ui.TypeStatus)
				dropResult := git.Execute("stash", "drop", ttInfo.OriginalStashID)
				if !dropResult.Success {
					buffer.Append(ConsoleMessages["warning_cleanup_stash"], ui.TypeStatus)
				}
			}
		}

		// Step 4: Clean up marker
		buffer.Append(ConsoleMessages["step_4_cleaning_marker"], ui.TypeStatus)
		err = git.ClearTimeTravelInfo()
		if err != nil {
			buffer.Append(fmt.Sprintf("Warning: Could not remove marker: %v", err), ui.TypeStatus)
		}

		buffer.Append(ConsoleMessages["restoration_complete"], ui.TypeStatus)

		return RestoreTimeTravelMsg{
			Success: true,
			Error:   "",
		}
	}
}

// handleRewind processes the result of a rewind (git reset --hard) operation
// Stays in console until user presses ESC

func (a *Application) Init() tea.Cmd {
	// sizing is already set from NewApplication with default dimensions (80, 40)
	// WindowSizeMsg will update it to actual terminal dimensions
	// Config is already loaded in main.go and passed to NewApplication (fail-fast)

	// CONTRACT: Start cache building immediately on app startup
	// Cache MUST be ready before history menus can be used
	commands := []tea.Cmd{tea.EnableBracketedPaste}

	if a.cacheManager.IsLoadingStarted() {
		commands = append(commands,
			a.cmdPreloadHistoryMetadata(),
			a.cmdPreloadFileHistoryDiffs(),
		)
	}

	// Async fetch remote on startup to ensure timeline accuracy
	// Without this, timeline state uses stale local refs
	// CONTRACT: Only start timeline sync if HasRemote AND AutoUpdate.Enabled
	// Fetch remote if available
	if a.gitState != nil && a.gitState.Remote == git.HasRemote {
		commands = append(commands, cmdFetchRemote())
	}

	// Start auto-update (Phase 2)
	if cmd := a.startAutoUpdate(); cmd != nil {
		commands = append(commands, cmd)
	}

	return tea.Batch(commands...)
}

// GetFooterHint returns the footer hint text
