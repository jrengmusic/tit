package app

import (
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// handleCacheProgress handles cache building progress updates

func (a *Application) handleCacheProgress(msg CacheProgressMsg) (tea.Model, tea.Cmd) {
	// Cache progress received - regenerate menu to show updated progress
	// Menu generator will read progress fields and show disabled state with progress

	// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
	if a.mode == ModeMenu {
		menu := a.GenerateMenu()
		a.menuItems = menu
		if !a.quitConfirmActive && len(menu) > 0 && a.selectedIndex < len(menu) {
			a.footerHint = menu[a.selectedIndex].Hint
		}
	}

	if msg.Complete {
		// Cache complete - rebuild menu shortcuts to enable items (only in ModeMenu)
		if a.mode == ModeMenu {
			a.rebuildMenuShortcuts(ModeMenu)
		}

		// Check if BOTH caches are now complete (for time travel success message)
		metadataReady := a.cacheManager.IsMetadataReady()
		diffsReady := a.cacheManager.IsDiffsReady()

		// If both caches complete AND in console mode after async operation finished,
		// show "Press ESC to return to menu" message
		if metadataReady && diffsReady && a.mode == ModeConsole && !a.isAsyncActive() {
			buffer := ui.GetBuffer()

			// Check if this is time travel mode (handled separately)
			if a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
				buffer.Append(ConsoleMessages["time_travel_success"], ui.TypeStatus)
			} else {
				// Regular operation (commit, push, etc.) - show completion message
				buffer.Append(GetFooterMessageText(MessageOperationComplete), ui.TypeInfo)
			}
		}
	}

	return a, nil
}

// handleCacheRefreshTick handles periodic cache progress refresh
// Regenerates menu to show updated progress and re-schedules if caches not complete

func (a *Application) handleCacheRefreshTick() (tea.Model, tea.Cmd) {
	// Check if both caches are complete
	metadataComplete := a.cacheManager.IsMetadataReady()
	diffsComplete := a.cacheManager.IsDiffsReady()

	// If both complete, stop ticking
	if metadataComplete && diffsComplete {
		return a, nil
	}

	// Advance animation frame
	a.cacheManager.IncrementAnimationFrame()

	// Only regenerate menu if we're in ModeMenu (don't interfere with other modes)
	if a.mode == ModeMenu {
		menu := a.GenerateMenu()
		a.menuItems = menu
		if !a.quitConfirmActive && len(menu) > 0 && a.selectedIndex < len(menu) {
			a.footerHint = menu[a.selectedIndex].Hint
		}
	}

	// Re-schedule another tick
	return a, a.cmdRefreshCacheProgress()
}

// updateFooterHintFromMenu updates footer with hint of currently selected menu item
// Skips update if app-level message is active (quitConfirmActive)

func (a *Application) updateFooterHintFromMenu() {
	if a.quitConfirmActive {
		return
	}
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.menuItems) {
		if !a.menuItems[a.selectedIndex].Separator {
			a.footerHint = a.menuItems[a.selectedIndex].Hint
		}
	}
}

// GetGitState returns the current git state

func (a *Application) GetGitState() interface{} {
	return a.gitState
}

// RenderStateHeader renders the full git state header (5 rows) using lipgloss
// Row 1: CWD (left) | OPERATION (right)
// RenderStateHeader renders the state header per REACTIVE-LAYOUT-PLAN.md
// 2-column layout: 80/20 split
// LEFT (80%): CWD, Remote, WorkingTree, Timeline
// RIGHT (20%): Operation, Branch
