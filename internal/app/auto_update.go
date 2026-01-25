package app

import (
	"fmt"
	"os"
	"time"

	"tit/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// startAutoUpdate initiates background state updates if enabled
func (a *Application) startAutoUpdate() tea.Cmd {
	// Only start if enabled in config
	if a.appConfig == nil || !a.appConfig.AutoUpdate.Enabled {
		return nil
	}

	// Only run in menu mode
	if a.mode != ModeMenu {
		return nil
	}

	// Schedule first tick
	return a.scheduleAutoUpdateTick()
}

// scheduleAutoUpdateTick schedules the next auto-update check
func (a *Application) scheduleAutoUpdateTick() tea.Cmd {
	// Config is guaranteed to have IntervalMinutes (see config.Load() fallbacks)
	interval := time.Duration(a.appConfig.AutoUpdate.IntervalMinutes) * time.Minute

	return tea.Tick(interval, func(time.Time) tea.Msg {
		return AutoUpdateTickMsg{}
	})
}

// handleAutoUpdateTick processes the auto-update tick
func (a *Application) handleAutoUpdateTick() (tea.Model, tea.Cmd) {
	// Only process if still in menu mode
	if a.mode != ModeMenu {
		return a, nil
	}

	// Only process if auto-update still enabled
	if a.appConfig == nil || !a.appConfig.AutoUpdate.Enabled {
		return a, nil
	}

	// Skip if user recently navigated menu (lazy update)
	if time.Since(a.lastMenuActivity) < a.menuActivityTimeout {
		return a, a.scheduleAutoUpdateTick()
	}

	// Skip if auto-update already in progress
	if a.autoUpdateInProgress {
		return a, nil
	}

	// Start auto-update: set in progress and run
	a.autoUpdateInProgress = true
	a.autoUpdateFrame = 0
	return a, tea.Batch(
		a.cmdAutoUpdate(),
		a.scheduleAutoUpdateAnimation(), // Start spinner animation
		a.scheduleAutoUpdateTick(),      // Schedule next check tick
	)
}

// cmdAutoUpdate performs full state detection and UI update
func (a *Application) cmdAutoUpdate() tea.Cmd {
	return func() tea.Msg {
		// DEBUG: Log auto-update trigger
		os.Stderr.WriteString("[DEBUG] Auto-update: Starting auto-update cycle\n")

		// If has remote, fetch first (optional)
		if a.gitState != nil && a.gitState.Remote == git.HasRemote {
			os.Stderr.WriteString("[DEBUG] Auto-update: Fetching from origin\n")
			git.Execute("fetch", "origin")
			// Ignore errors - just detect state as-is
		}

		// Detect full state (all 5 axes)
		os.Stderr.WriteString("[DEBUG] Auto-update: Calling git.DetectState()\n")
		newState, err := git.DetectState()
		if err != nil {
			os.Stderr.WriteString(fmt.Sprintf("[DEBUG] Auto-update: git.DetectState() failed: %v\n", err))
			// Silently ignore - don't interrupt user
			return nil
		}

		os.Stderr.WriteString(fmt.Sprintf("[DEBUG] Auto-update: git.DetectState() completed. WorkingTree=%v\n", newState.WorkingTree))

		return AutoUpdateCompleteMsg{State: newState}
	}
}

// handleAutoUpdateComplete updates UI with new state
func (a *Application) handleAutoUpdateComplete(state *git.State) (tea.Model, tea.Cmd) {
	// Clear in-progress flag
	a.autoUpdateInProgress = false

	if state == nil {
		return a, nil
	}

	// Update git state
	a.gitState = state

	// Regenerate menu (full rebuild, like at launch)
	if a.mode == ModeMenu {
		oldSelectedIndex := a.selectedIndex
		oldMenuLen := len(a.menuItems)

		// Full menu regeneration
		a.menuItems = a.GenerateMenu()
		a.rebuildMenuShortcuts()

		// Preserve selection if possible
		if oldMenuLen == len(a.menuItems) {
			a.selectedIndex = oldSelectedIndex
		} else if a.selectedIndex >= len(a.menuItems) {
			a.selectedIndex = len(a.menuItems) - 1
		}

		// Update footer hint
		a.updateFooterHintFromMenu()
	}

	return a, nil
}

// scheduleAutoUpdateAnimation schedules spinner animation frames during auto-update
func (a *Application) scheduleAutoUpdateAnimation() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return AutoUpdateAnimationMsg{}
	})
}

// handleAutoUpdateAnimation advances spinner animation frame
func (a *Application) handleAutoUpdateAnimation() (tea.Model, tea.Cmd) {
	// Only update frame if still in progress
	if !a.autoUpdateInProgress {
		return a, nil
	}

	// Advance animation frame
	a.autoUpdateFrame++

	// Schedule next animation frame while in progress
	if a.autoUpdateInProgress {
		return a, a.scheduleAutoUpdateAnimation()
	}

	return a, nil
}
