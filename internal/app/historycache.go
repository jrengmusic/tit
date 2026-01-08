package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// PreloadHistoryMetadata loads commit metadata for History mode (async)
// CONTRACT: MANDATORY cache build before showing history
// Returns tea.Cmd that spawns goroutine and sends progress updates
// Call via: return app.cmdPreloadHistoryMetadata()
func (a *Application) cmdPreloadHistoryMetadata() tea.Cmd {
	return func() tea.Msg {
		// Fetch list of recent commits (basic info only)
		commits, err := git.FetchRecentCommits(30)
		if err != nil {
			// No commits available - cache will remain empty, mark as loaded
			a.historyCacheMutex.Lock()
			a.cacheMetadata = true
			a.cacheMetadataProgress = 0
			a.cacheMetadataTotal = 0
			a.historyCacheMutex.Unlock()

			return CacheProgressMsg{
				CacheType: "metadata",
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

		return CacheProgressMsg{
			CacheType: "metadata",
			Current:   len(commits),
			Total:     len(commits),
			Complete:  true,
		}
	}
}

// PreloadFileHistoryDiffs loads file lists and diffs (async)
// CONTRACT: MANDATORY cache build before showing file history
// Returns tea.Cmd that spawns goroutine and sends progress updates
// Call via: return app.cmdPreloadFileHistoryDiffs()
func (a *Application) cmdPreloadFileHistoryDiffs() tea.Cmd {
	return func() tea.Msg {
		// Fetch list of recent commits (basic info only)
		commits, err := git.FetchRecentCommits(30)
		if err != nil {
			// No commits available - cache will remain empty, mark as loaded
			a.diffCacheMutex.Lock()
			a.cacheDiffs = true
			a.cacheDiffsProgress = 0
			a.cacheDiffsTotal = 0
			a.diffCacheMutex.Unlock()

			return CacheProgressMsg{
				CacheType: "diffs",
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
					key := commit.Hash + ":" + file.Path + ":parent"
					a.fileHistoryDiffCache[key] = parentDiff
					a.diffCacheMutex.Unlock()
				}

				// Version 2: Commit vs working tree (Modified state)
				wipDiff, err := git.GetCommitDiff(commit.Hash, file.Path, "wip")
				if err == nil {
					// Thread-safe: cache the diff
					a.diffCacheMutex.Lock()
					key := commit.Hash + ":" + file.Path + ":wip"
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

		return CacheProgressMsg{
			CacheType: "diffs",
			Current:   len(commits),
			Total:     len(commits),
			Complete:  true,
		}
	}
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

	// Restart loading (CONTRACT: MUST rebuild before operation complete)
	a.cacheLoadingStarted = false
	if a.gitState.Operation == git.Normal {
		a.cacheLoadingStarted = true
		// Return batch command to run both caches in parallel
		return tea.Batch(
			a.cmdPreloadHistoryMetadata(),
			a.cmdPreloadFileHistoryDiffs(),
		)
	}

	return nil
}
