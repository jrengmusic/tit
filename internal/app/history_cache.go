package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"
)

// ===== CACHE KEY SCHEMA DEFINITIONS =====
// Centralizes cache key formats and validation
// CONSOLIDATION: Prevents hardcoded key formats scattered throughout code

// Cache type constants
const (
	CacheTypeMetadata = "metadata"
	CacheTypeDiffs    = "diffs"
	CacheTypeFiles    = "files"
)

// DiffCacheKey constructs a diff cache key from hash, filepath, and version
// Schema: "hash:filepath:version"
// Example: "abc123def456:src/main.go:after"
func DiffCacheKey(hash, filepath, version string) string {
	return fmt.Sprintf("%s:%s:%s", hash, filepath, version)
}

// ParseDiffCacheKey parses a diff cache key back into its components
// Returns (hash, filepath, version, error)
func ParseDiffCacheKey(key string) (hash, filepath, version string, err error) {
	parts := strings.Split(key, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid diff cache key format: %s (expected 3 colon-separated parts)", key)
	}
	return parts[0], parts[1], parts[2], nil
}

// PreloadHistoryMetadata loads commit metadata for History mode (async)
// CONTRACT: MANDATORY cache build before showing history
// Returns tea.Cmd that spawns goroutine and sends progress updates
// Call via: return app.cmdPreloadHistoryMetadata()
func (a *Application) cmdPreloadHistoryMetadata() tea.Cmd {
	workerCmd := func() tea.Msg {
		buffer := ui.GetBuffer()

		// Determine git ref to log from
		// During time travel: use original branch to show full history
		// Otherwise: use HEAD (current position)
		ref := ""
		if a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
			originalBranch, _, err := git.GetTimeTravelInfo()
			if err == nil && originalBranch != "" {
				ref = originalBranch
				buffer.Append(fmt.Sprintf("Building history cache from branch %s...", ref), ui.TypeStatus)
			}
		}

		// Fetch list of recent commits (basic info only)
		commits, err := git.FetchRecentCommits(30, ref)
		if err != nil {
			// No commits available - cache will remain empty, mark as loaded
			a.historyCacheMutex.Lock()
			a.cacheMetadata = true
			a.cacheMetadataProgress = 0
			a.cacheMetadataTotal = 0
			a.historyCacheMutex.Unlock()

			buffer.Append("History cache build complete (no commits)", ui.TypeStatus)
			return CacheProgressMsg{
				CacheType: CacheTypeMetadata,
				Current:   0,
				Total:     0,
				Complete:  true,
			}
		}

		// Set total
		a.historyCacheMutex.Lock()
		a.cacheMetadataTotal = len(commits)
		a.cacheMetadataProgress = 0
		a.historyCacheMutex.Unlock()

		// For each commit, fetch full metadata and cache it
		for i, commit := range commits {
			details, err := git.GetCommitDetails(commit.Hash)
			if err != nil {
				// Skip this commit if metadata fetch fails
				continue
			}

			// Thread-safe: lock before writing to cache
			a.historyCacheMutex.Lock()
			a.historyMetadataCache[commit.Hash] = details
			a.cacheMetadataProgress = i + 1
			a.historyCacheMutex.Unlock()
		}

		// Mark cache as complete
		a.historyCacheMutex.Lock()
		a.cacheMetadata = true
		a.historyCacheMutex.Unlock()

		buffer.Append(fmt.Sprintf("History metadata cache complete (%d commits)", len(commits)), ui.TypeStatus)
		return CacheProgressMsg{
			CacheType: CacheTypeMetadata,
			Current:   len(commits),
			Total:     len(commits),
			Complete:  true,
		}
	}

	// Return both the worker AND a refresh ticker to show progress updates
	// The worker goroutine will update cacheMetadataProgress, and refreshCmd will trigger re-renders
	return tea.Batch(
		workerCmd,
		a.cmdRefreshCacheProgress(),
	)
}

// PreloadFileHistoryDiffs loads file lists and diffs (async)
// CONTRACT: MANDATORY cache build before showing file history
// Returns tea.Cmd that spawns goroutine and sends progress updates
// Call via: return app.cmdPreloadFileHistoryDiffs()
func (a *Application) cmdPreloadFileHistoryDiffs() tea.Cmd {
	workerCmd := func() tea.Msg {
		buffer := ui.GetBuffer()

		// Determine git ref to log from (same logic as metadata cache)
		ref := ""
		if a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
			originalBranch, _, err := git.GetTimeTravelInfo()
			if err == nil && originalBranch != "" {
				ref = originalBranch
			}
		}

		// Fetch list of recent commits (basic info only)
		commits, err := git.FetchRecentCommits(30, ref)
		if err != nil {
			// No commits available - cache will remain empty, mark as loaded
			a.diffCacheMutex.Lock()
			a.cacheDiffs = true
			a.cacheDiffsProgress = 0
			a.cacheDiffsTotal = 0
			a.diffCacheMutex.Unlock()

			buffer.Append("File history cache build complete (no commits)", ui.TypeStatus)
			return CacheProgressMsg{
				CacheType: CacheTypeDiffs,
				Current:   0,
				Total:     0,
				Complete:  true,
			}
		}

		// Set total
		a.diffCacheMutex.Lock()
		a.cacheDiffsTotal = len(commits)
		a.cacheDiffsProgress = 0
		a.diffCacheMutex.Unlock()

		// For each commit, fetch file list and diffs
		for i, commit := range commits {
			// Get list of files changed in this commit
			files, err := git.GetFilesInCommit(commit.Hash)
			if err != nil {
				// Skip this commit if file list fetch fails
				continue
			}

			// Thread-safe: cache the file list (always cache, lightweight)
			a.diffCacheMutex.Lock()
			a.fileHistoryFilesCache[commit.Hash] = files
			a.diffCacheMutex.Unlock()

			// Skip caching diffs for commits with >100 files (too expensive)
			if len(files) > 100 {
				continue
			}

			// Cache both diff versions for each file in this commit
			for _, file := range files {
				// Version 1: Commit vs parent (Clean state)
				parentDiff, err := git.GetCommitDiff(commit.Hash, file.Path, "parent")
				if err == nil {
					// Thread-safe: cache the diff
					a.diffCacheMutex.Lock()
					key := DiffCacheKey(commit.Hash, file.Path, "parent")
					a.fileHistoryDiffCache[key] = parentDiff
					a.diffCacheMutex.Unlock()
				}

				// Version 2: Commit vs working tree (Modified state)
				wipDiff, err := git.GetCommitDiff(commit.Hash, file.Path, "wip")
				if err == nil {
					// Thread-safe: cache the diff
					a.diffCacheMutex.Lock()
					key := DiffCacheKey(commit.Hash, file.Path, "wip")
					a.fileHistoryDiffCache[key] = wipDiff
					a.diffCacheMutex.Unlock()
				}
			}

			// Update progress after each commit
			a.diffCacheMutex.Lock()
			a.cacheDiffsProgress = i + 1
			a.diffCacheMutex.Unlock()
		}

		// Mark cache as complete
		a.diffCacheMutex.Lock()
		a.cacheDiffs = true
		a.diffCacheMutex.Unlock()

		buffer.Append(fmt.Sprintf("File history cache complete (%d commits)", len(commits)), ui.TypeStatus)
		return CacheProgressMsg{
			CacheType: CacheTypeDiffs,
			Current:   len(commits),
			Total:     len(commits),
			Complete:  true,
		}
	}

	// Return both the worker AND a refresh ticker to show progress updates
	// The worker goroutine will update cacheDiffsProgress, and refreshCmd will trigger re-renders
	return tea.Batch(
		workerCmd,
		a.cmdRefreshCacheProgress(),
	)
}

// InvalidateHistoryCaches clears all caches (call after commits, merges, or time travel)
// CONTRACT: Rebuilds cache immediately, returns tea.Cmd for async execution
// Thread-safe: acquires both mutexes
func (a *Application) invalidateHistoryCaches() tea.Cmd {
	// Clear history metadata cache
	a.historyCacheMutex.Lock()
	a.historyMetadataCache = make(map[string]*git.CommitDetails)
	a.cacheMetadata = false
	a.cacheMetadataProgress = 0
	a.cacheMetadataTotal = 0
	a.historyCacheMutex.Unlock()

	// Clear file history caches
	a.diffCacheMutex.Lock()
	a.fileHistoryDiffCache = make(map[string]string)
	a.fileHistoryFilesCache = make(map[string][]git.FileInfo)
	a.cacheDiffs = false
	a.cacheDiffsProgress = 0
	a.cacheDiffsTotal = 0
	a.diffCacheMutex.Unlock()

	// Reset state structures
	a.historyState = &ui.HistoryState{
		Commits:           make([]ui.CommitInfo, 0),
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}

	a.fileHistoryState = &ui.FileHistoryState{
		Commits:           make([]ui.CommitInfo, 0),
		Files:             make([]ui.FileInfo, 0),
		SelectedCommitIdx: 0,
		SelectedFileIdx:   0,
		FocusedPane:       ui.PaneCommits,
		CommitsScrollOff:  0,
		FilesScrollOff:    0,
		DiffScrollOff:     0,
		DiffLineCursor:    0,
		VisualModeActive:  false,
		VisualModeStart:   0,
	}

	// CONTRACT: Restart loading for ALL states (Normal, TimeTraveling, etc.)
	// Operation guard REMOVED - cache must rebuild regardless of git state
	a.cacheLoadingStarted = true

	// Return batch command to run both caches in parallel
	return tea.Batch(
		a.cmdPreloadHistoryMetadata(),
		a.cmdPreloadFileHistoryDiffs(),
	)
}

// cmdToggleAutoUpdate toggles auto-update setting and updates config
func (a *Application) cmdToggleAutoUpdate() tea.Cmd {
	return func() tea.Msg {
		// Load current config
		cfg, err := config.Load()
		if err != nil {
			return ToggleAutoUpdateMsg{Success: false, Error: err.Error()}
		}

		// Toggle the setting
		newValue := !cfg.AutoUpdate.Enabled
		if err := cfg.SetAutoUpdateEnabled(newValue); err != nil {
			return ToggleAutoUpdateMsg{Success: false, Error: err.Error()}
		}

		// Update timeline sync state based on new setting
		if newValue {
			a.startTimelineSync()
		}

		return ToggleAutoUpdateMsg{Success: true, Enabled: newValue}
	}
}

// ToggleAutoUpdateMsg is sent when auto-update toggle completes
type ToggleAutoUpdateMsg struct {
	Success bool
	Error   string
	Enabled bool
}
