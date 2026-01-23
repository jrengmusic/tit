package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
)

// Timeline sync constants (SSOT)
const (
	TimelineSyncTickRate   = 100 * time.Millisecond // Animation refresh rate
	TimelineSyncInterval   = 60 * time.Second       // Default periodic sync interval
)

// cmdTimelineSync runs git fetch in background and updates timeline
// CONTRACT: Checks remote existence before running; includes stderr on fetch failure
func (a *Application) cmdTimelineSync() tea.Cmd {
	// Capture remote state BEFORE returning closure (avoid stale closure captures)
	hasRemote := a.gitState != nil && a.gitState.Remote == git.HasRemote
	
	return func() tea.Msg {
		// Check if remote exists - fail-fast with explicit error
		if !hasRemote {
			return TimelineSyncMsg{
				Success: false,
				Error:   "no remote configured",
			}
		}

		// Run git fetch
		result := git.Execute("fetch", "origin")
		if !result.Success {
			// FAIL-FAST: Include stderr in error message, not just generic text
			errMsg := "fetch failed"
			if result.Stderr != "" {
				errMsg += ": " + result.Stderr
			}
			return TimelineSyncMsg{
				Success: false,
				Error:   errMsg,
			}
		}

		// Re-detect state to get updated timeline
		newState, err := git.DetectState()
		if err != nil {
			return TimelineSyncMsg{
				Success: false,
				Error:   "failed to detect state after fetch: " + err.Error(),
			}
		}

		return TimelineSyncMsg{
			Success:  true,
			Timeline: newState.Timeline,
			Ahead:    newState.CommitsAhead,
			Behind:   newState.CommitsBehind,
		}
	}
}

// cmdTimelineSyncTicker schedules periodic timeline sync
// CONTRACT: Only schedules next tick if mode == ModeMenu
func (a *Application) cmdTimelineSyncTicker() tea.Cmd {
	return tea.Tick(TimelineSyncTickRate, func(t time.Time) tea.Msg {
		return TimelineSyncTickMsg{}
	})
}

// shouldRunTimelineSync checks if sync should run
// Returns false if: NoRemote, sync in progress, recently synced, or disabled in config
func (a *Application) shouldRunTimelineSync() bool {
	// Check if auto-update is disabled in config
	if a.appConfig != nil && !a.appConfig.AutoUpdate.Enabled {
		return false
	}

	// No remote - no sync needed
	if a.gitState == nil || a.gitState.Remote != git.HasRemote {
		return false
	}

	// Sync already in progress
	if a.timelineSyncInProgress {
		return false
	}

	// Recently synced - check interval from config or use SSOT default
	if !a.timelineSyncLastUpdate.IsZero() {
		interval := TimelineSyncInterval // SSOT: 60 seconds default
		if a.appConfig != nil && a.appConfig.AutoUpdate.IntervalMinutes > 0 {
			interval = time.Duration(a.appConfig.AutoUpdate.IntervalMinutes) * time.Minute
		}
		if time.Since(a.timelineSyncLastUpdate) < interval {
			return false
		}
	}

	return true
}

// handleTimelineSyncMsg processes the result of background timeline sync
// CONTRACT: Only mark last-sync on success; don't update on no-remote error
func (a *Application) handleTimelineSyncMsg(msg TimelineSyncMsg) (tea.Model, tea.Cmd) {
	a.timelineSyncInProgress = false

	if msg.Success {
		// Mark successful sync time (only on actual fetch, not on error)
		a.timelineSyncLastUpdate = time.Now()
		
		// Update git state with new timeline info
		if a.gitState != nil {
			a.gitState.Timeline = msg.Timeline
			a.gitState.CommitsAhead = msg.Ahead
			a.gitState.CommitsBehind = msg.Behind
		}

		// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
		if a.mode == ModeMenu {
			a.menuItems = a.GenerateMenu()
			a.rebuildMenuShortcuts()
			a.updateFooterHintFromMenu()
		}
		
		// Update footer hint in preferences to show sync completed
		if a.mode == ModePreferences {
			a.footerHint = "Auto-update sync completed"
		}
	} else {
		// On error (including no-remote), don't update last-sync timestamp
		// This allows shouldRunTimelineSync to retry sooner on failure
		if msg.Error != "" {
			// Surface error to footer in preferences mode
			if a.mode == ModePreferences {
				a.footerHint = fmt.Sprintf("Sync failed: %s", msg.Error)
			}
		}
	}

	// If in ModeMenu, schedule next periodic sync
	if a.mode == ModeMenu && a.shouldRunTimelineSync() {
		a.timelineSyncInProgress = true
		return a, a.cmdTimelineSync()
	}

	// Schedule ticker for animation updates while in menu mode
	if a.mode == ModeMenu {
		return a, a.cmdTimelineSyncTicker()
	}

	return a, nil
}

// handleTimelineSyncTickMsg handles periodic sync ticks
// Updates animation frame while sync is in progress
func (a *Application) handleTimelineSyncTickMsg() (tea.Model, tea.Cmd) {
	// Advance animation frame for spinner (all modes with headers show this)
	a.timelineSyncFrame++

	// Only schedule new syncs when in ModeMenu (auto-sync only in menu)
	if a.mode == ModeMenu {
		// Check if we should run a new sync
		if a.shouldRunTimelineSync() {
			a.timelineSyncInProgress = true
			return a, a.cmdTimelineSync()
		}
	}

	// Schedule next tick if sync still in progress (for animation in all modes)
	if a.timelineSyncInProgress {
		return a, a.cmdTimelineSyncTicker()
	}

	return a, nil
}

// startTimelineSync initiates a timeline sync if conditions are met
// Called on app startup and when returning to menu mode
func (a *Application) startTimelineSync() {
	if a.gitState != nil && a.gitState.Remote == git.HasRemote {
		a.timelineSyncInProgress = true
		a.timelineSyncFrame = 0
	}
}
