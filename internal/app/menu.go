package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tit/internal/git"
	"tit/internal/ui"
)

// MenuItem represents a single menu action or separator
type MenuItem struct {
	ID            string // Unique identifier for the action
	Shortcut      string // Keyboard shortcut (actual key binding, e.g., "}", "{")
	ShortcutLabel string // Display label for shortcut (e.g., "shift + ]"), empty = use Shortcut
	Emoji         string // Leading emoji
	Label         string // Action name
	Hint          string // Plain language hint shown on focus
	Enabled       bool   // Whether this item can be selected
	Separator     bool   // If true, this is a visual separator (non-selectable)
}

// MenuGenerator is a function type that generates menu items
type MenuGenerator func(*Application) []MenuItem

// GenerateMenu produces menu items based on current git state
func (a *Application) GenerateMenu() []MenuItem {
	// Priority 1: Operation State (most restrictive)
	if a.gitState == nil {
		return []MenuItem{}
	}

	menuGenerators := map[git.Operation]MenuGenerator{
		git.NotRepo:       (*Application).menuNotRepo,
		git.Normal:        (*Application).menuNormal,
		git.TimeTraveling: (*Application).menuTimeTraveling,
	}

	if generator, exists := menuGenerators[a.gitState.Operation]; exists {
		return generator(a)
	}

	// Unknown operation: fail fast with clear error
	panic(fmt.Sprintf("Unknown git operation state: %v", a.gitState.Operation))
}

// menuNotRepo returns menu for NotRepo state
// Always shows both options; smart dispatch happens in dispatchInit/dispatchClone
func (a *Application) menuNotRepo() []MenuItem {
	return []MenuItem{
		GetMenuItem("init"),
		GetMenuItem("clone"),
	}
}

// detectConflictedOperation determines which operation caused conflicts
func detectConflictedOperation() string {
	if _, err := os.Stat(".git/MERGE_HEAD"); err == nil {
		return "merge"
	}
	if _, err := os.Stat(".git/rebase-merge"); err == nil {
		return "rebase"
	}
	if _, err := os.Stat(".git/rebase-apply"); err == nil {
		return "rebase"
	}
	if _, err := os.Stat(".git/CHERRY_PICK_HEAD"); err == nil {
		return "cherry-pick"
	}
	return "unknown"
}

// menuNormal returns menu for Normal operation state
func (a *Application) menuNormal() []MenuItem {
	var items []MenuItem

	items = append(items, a.menuWorkingTree()...)

	items = append(items, a.menuTimeline()...)

	// Separator before History section (if there are items above)
	if len(items) > 0 {
		items = append(items, Item("").Separator().Build())
	}

	// History section (always shown)
	items = append(items, a.menuHistory()...)

	// First-time setup (always at bottom)
	if a.gitState.Remote == git.NoRemote {
		items = append(items,
			Item("").Separator().Build(),
			GetMenuItem("add_remote"),
		)
	}

	return items
}

// menuWorkingTree returns working tree actions
func (a *Application) menuWorkingTree() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	// Only show commit when Dirty - HIDDEN when Clean
	switch a.gitState.WorkingTree {
	case git.Clean:
		return []MenuItem{} // No working tree actions when Clean

	case git.Dirty:
		items := []MenuItem{
			GetMenuItem("commit"),
		}

		// Show "Commit and push" only if remote exists
		if a.gitState.Remote == git.HasRemote {
			items = append(items, GetMenuItem("commit_push"))
		}

		return items
	}

	return []MenuItem{}
}

// menuTimeline returns timeline sync actions
func (a *Application) menuTimeline() []MenuItem {
	if a.gitState == nil {
		return []MenuItem{}
	}

	// No remote → no timeline operations (add_remote shown at bottom of menuNormal)
	if a.gitState.Remote == git.NoRemote {
		return []MenuItem{}
	}

	var items []MenuItem

	switch a.gitState.Timeline {
	case git.InSync:
		// When in sync but have uncommitted changes, allow reset to remote
		if a.gitState.WorkingTree == git.Dirty && a.gitState.Remote == git.HasRemote {
			items = append(items, GetMenuItem("reset_discard_changes"))
		}
		// No other sync actions needed when in sync
		return items

	case git.Ahead:
		// Local ahead → show push ONLY if working tree is Clean
		// cannot push uncommitted changes
		if a.gitState.WorkingTree == git.Clean {
			items = append(items,
				GetMenuItem("push"),
				GetMenuItem("force_push"),
			)
		}

	case git.Behind:
		// Dirty → ONLY dirty pull (don't offer clean pull that would lose work)
		if a.gitState.WorkingTree == git.Dirty {
			items = append(items,
				GetMenuItem("dirty_pull_merge"),
				GetMenuItem("replace_local"), // Destructive option (has confirmation)
			)
		} else {
			// Clean → ONLY clean pull options
			items = append(items,
				GetMenuItem("pull_merge"),
				GetMenuItem("replace_local"),
			)
		}

	case git.Diverged:
		// Dirty → ONLY dirty pull (don't offer clean pull that would lose work)
		if a.gitState.WorkingTree == git.Dirty {
			items = append(items,
				GetMenuItem("dirty_pull_merge"),
				GetMenuItem("force_push"),    // Destructive option (has confirmation)
				GetMenuItem("replace_local"), // Destructive option (has confirmation)
			)
		} else {
			// Clean → ONLY clean pull options
			items = append(items,
				GetMenuItem("pull_merge_diverged"),
				GetMenuItem("force_push"),
				GetMenuItem("replace_local"),
			)
		}
	}

	return items
}

// getHistoryItemsWithCacheState returns history items with cache state applied
// CONTRACT: Centralized cache checking - no duplication across menu generators
// Takes item IDs for history and file history, returns items with:
// - Disabled state while building
// - Progress indicators: ⏳ [Building... 12/30]
// - Enabled state when cache ready
func (a *Application) getHistoryItemsWithCacheState(historyID, fileHistoryID string) []MenuItem {
	// Get cache status (thread-safe)
	metadataReady := a.cacheManager.IsMetadataReady()
	metadataProgress, metadataTotal := a.cacheManager.GetMetadataProgress()

	diffsReady := a.cacheManager.IsDiffsReady()
	diffsProgress, diffsTotal := a.cacheManager.GetDiffsProgress()

	items := []MenuItem{}

	// History menu item - CONTRACT: disabled while building, shows progress
	historyItem := GetMenuItem(historyID)
	if !metadataReady {
		historyItem.Enabled = false
		historyItem.Emoji = ui.GetSpinnerFrame(a.cacheManager.GetAnimationFrame())
		if metadataTotal > 0 {
			historyItem.Label = fmt.Sprintf("History %d/%d", metadataProgress, metadataTotal)
		} else {
			historyItem.Label = "History..."
		}
	} else {
		historyItem.Enabled = true
	}
	items = append(items, historyItem)

	// File history menu item - CONTRACT: disabled while building, shows progress
	fileHistoryItem := GetMenuItem(fileHistoryID)
	if !diffsReady {
		fileHistoryItem.Enabled = false
		fileHistoryItem.Emoji = ui.GetSpinnerFrame(a.cacheManager.GetAnimationFrame())
		if diffsTotal > 0 {
			fileHistoryItem.Label = fmt.Sprintf("Files %d/%d", diffsProgress, diffsTotal)
		} else {
			fileHistoryItem.Label = "Files..."
		}
	} else {
		fileHistoryItem.Enabled = true
	}
	items = append(items, fileHistoryItem)

	return items
}

// menuHistory returns history actions
// CONTRACT: Disables menu items and shows progress while cache is building
func (a *Application) menuHistory() []MenuItem {
	return a.getHistoryItemsWithCacheState("history", "file_history")
}

// menuTimeTraveling returns menu for TimeTraveling operation state
// CONTRACT: Uses centralized cache checking (no hardcoded items)
func (a *Application) menuTimeTraveling() []MenuItem {
	// Get original branch from .git/TIT_TIME_TRAVEL file (not from detached HEAD)
	originalBranch := "unknown"
	travelInfoPath := filepath.Join(".git", "TIT_TIME_TRAVEL")
	data, err := os.ReadFile(travelInfoPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			originalBranch = lines[0]
		}
	}

	// Get history items with cache state applied (centralized logic)
	items := a.getHistoryItemsWithCacheState("time_travel_history", "time_travel_files_history")

	// Add single return option (handles both merge and discard via dialog when dirty)
	returnItem := GetMenuItem("time_travel_return")
	returnItem.Label = fmt.Sprintf("Return to %s", originalBranch)
	items = append(items, returnItem)

	return items
}

// menuInitializeLocation returns options for where to initialize repository
func (a *Application) menuInitializeLocation() []MenuItem {
	return []MenuItem{
		GetMenuItem("init_here"),
		GetMenuItem("init_subdir"),
	}
}

// menuCloneLocation returns options for where to clone repository
func (a *Application) menuCloneLocation() []MenuItem {
	return []MenuItem{
		GetMenuItem("clone_here"),
		GetMenuItem("clone_subdir"),
	}
}

// GenerateConfigMenu generates config menu items based on dynamic git state
func (a *Application) GenerateConfigMenu() []MenuItem {
	var items []MenuItem

	// Remote operations (dynamic based on remote state)
	if a.gitState != nil && a.gitState.Remote == git.NoRemote {
		items = append(items, GetMenuItem("config_add_remote"))
	} else {
		items = append(items, GetMenuItem("config_switch_remote"))
	}

	items = append(items, Item("").Separator().Build())

	// Remove remote (only when remote exists)
	if a.gitState != nil && a.gitState.Remote == git.HasRemote {
		items = append(items, GetMenuItem("config_remove_remote"))
	}

	items = append(items, Item("").Separator().Build())

	// Branch switching (always available)
	items = append(items, GetMenuItem("config_switch_branch"))

	// Preferences (always available)
	items = append(items, GetMenuItem("config_preferences"))

	items = append(items, Item("").Separator().Build())

	// Back to main menu
	items = append(items, GetMenuItem("config_back"))

	return items
}

// GeneratePreferencesMenu generates preferences menu items
// Only 3 items - no Back (ESC in footer is sufficient)
func (a *Application) GeneratePreferencesMenu() []MenuItem {
	return []MenuItem{
		GetMenuItem("preferences_auto_update"),
		GetMenuItem("preferences_interval"),
		GetMenuItem("preferences_theme"),
	}
}
