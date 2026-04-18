package app

import (
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

