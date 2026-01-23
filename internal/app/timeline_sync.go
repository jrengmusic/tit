package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
)

// Timeline sync constants (SSOT)
const (
	TimelineSyncTickRate = 100 * time.Millisecond // Animation refresh rate
)

// cmdTimelineSync runs git fetch in background and updates timeline
// CONTRACT: Only runs when HasRemote, returns immediately if NoRemote
func (a *Application) cmdTimelineSync() tea.Cmd {
	return func() tea.Msg {
		// Check if remote exists - no-op if no remote
		if a.gitState == nil || a.gitState.Remote != git.HasRemote {
			return TimelineSyncMsg{
				Success:  true,
				Timeline: "",
				Ahead:    0,
				Behind:   0,
			}
		}

		// Run git fetch
		result := git.Execute("fetch", "origin")
		if !result.Success {
			return TimelineSyncMsg{
				Success: false,
				Error:   "fetch failed",
			}
		}

		// Re-detect state to get updated timeline
		newState, err := git.DetectState()
		if err != nil {
			return TimelineSyncMsg{
				Success: false,
				Error:   err.Error(),
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

	// Recently synced - check interval from config
	if !a.timelineSyncLastUpdate.IsZero() {
		interval := 60 * time.Second // default
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
func (a *Application) handleTimelineSyncMsg(msg TimelineSyncMsg) (tea.Model, tea.Cmd) {
	a.timelineSyncInProgress = false
	a.timelineSyncLastUpdate = time.Now()

	if msg.Success {
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
// Only updates UI when in ModeMenu to avoid disrupting other modes
func (a *Application) handleTimelineSyncTickMsg() (tea.Model, tea.Cmd) {
	// Only process ticks when in ModeMenu
	if a.mode != ModeMenu {
		return a, nil
	}

	// Advance animation frame for spinner
	a.timelineSyncFrame++

	// Check if we should run a new sync
	if a.shouldRunTimelineSync() {
		a.timelineSyncInProgress = true
		return a, a.cmdTimelineSync()
	}

	// Schedule next tick if sync still in progress (for animation)
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
