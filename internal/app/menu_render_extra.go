package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tit/internal/git"
	"tit/internal/ui"
)

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

// menuTimeTraveling returns menu for TimeTraveling operation state
// CONTRACT: Uses centralized cache checking (no hardcoded items)
func (a *Application) menuTimeTraveling() []MenuItem {
	// Get original branch (from TIT marker if exists, otherwise detect from git)
	originalBranch := ""
	travelInfoPath := filepath.Join(".git", "TIT_TIME_TRAVEL")
	data, err := os.ReadFile(travelInfoPath)
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(lines) > 0 && lines[0] != "" {
			originalBranch = lines[0]
		}
	}

	// If no TIT marker (manual detached HEAD), detect available branches
	if originalBranch == "" {
		branchResult := git.Execute("branch", "--list")
		if branchResult.Success {
			branches := strings.TrimSpace(branchResult.Stdout)
			if branches != "" {
				branchLines := strings.Split(branches, "\n")
				// Filter out asterisk (current branch marker is absent in detached)
				var validBranches []string
				for _, b := range branchLines {
					b = strings.TrimSpace(strings.TrimPrefix(b, "* "))
					if b != "" {
						validBranches = append(validBranches, b)
					}
				}

				if len(validBranches) == 1 {
					// Single branch - auto-select
					originalBranch = validBranches[0]
				} else if len(validBranches) > 1 {
					// Multiple branches - will use "Return to branch" (triggers picker)
					originalBranch = ""
				}
			}
		}
	}

	// Get history items with cache state applied (centralized logic)
	items := a.getHistoryItemsWithCacheState("time_travel_history", "time_travel_files_history")

	// Add single return option (handles both merge and discard via dialog when dirty)
	returnItem := GetMenuItem("time_travel_return")
	if originalBranch != "" {
		returnItem.Label = fmt.Sprintf("Return to %s", originalBranch)
	} else {
		returnItem.Label = "Return to branch"
	}
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
		items = append(items, Item("").Separator().Build())
	}

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
