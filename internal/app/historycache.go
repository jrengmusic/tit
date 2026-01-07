package app

import (
	"tit/internal/git"
	"tit/internal/ui"
)

// PreloadHistoryMetadata loads commit metadata for History mode (async)
// Fetches author, date, message for last 30 commits
// Non-blocking: runs in background goroutine, updates cache as data arrives
// Call via: go app.preloadHistoryMetadata()
func (a *Application) preloadHistoryMetadata() {
	// Fetch list of recent commits (basic info only)
	commits, err := git.FetchRecentCommits(30)
	if err != nil {
		// No commits available - cache will remain empty, mark as loaded
		a.historyCacheMutex.Lock()
		a.cacheMetadata = true
		a.historyCacheMutex.Unlock()
		return
	}

	// For each commit, fetch full metadata and cache it
	for _, commit := range commits {
		details, err := git.GetCommitDetails(commit.Hash)
		if err != nil {
			// Skip this commit if metadata fetch fails
			continue
		}

		// Thread-safe: lock before writing to cache
		a.historyCacheMutex.Lock()
		a.historyMetadataCache[commit.Hash] = details
		a.historyCacheMutex.Unlock()
	}

	// Mark cache as complete
	a.historyCacheMutex.Lock()
	a.cacheMetadata = true
	a.historyCacheMutex.Unlock()
}

// PreloadFileHistoryDiffs loads file lists and diffs (async)
// Fetches files and diffs for last 30 commits
// Skips diff caching for commits with >100 files (performance optimization)
// Non-blocking: runs in background goroutine
// Call via: go app.preloadFileHistoryDiffs()
func (a *Application) preloadFileHistoryDiffs() {
	// Fetch list of recent commits (basic info only)
	commits, err := git.FetchRecentCommits(30)
	if err != nil {
		// No commits available - cache will remain empty, mark as loaded
		a.diffCacheMutex.Lock()
		a.cacheDiffs = true
		a.diffCacheMutex.Unlock()
		return
	}

	// For each commit, fetch file list and diffs
	for _, commit := range commits {
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
	}

	// Mark cache as complete
	a.diffCacheMutex.Lock()
	a.cacheDiffs = true
	a.diffCacheMutex.Unlock()
}

// InvalidateHistoryCaches clears all caches (call after commits, merges, or time travel)
// Thread-safe: acquires both mutexes
func (a *Application) invalidateHistoryCaches() {
	// Clear history metadata cache
	a.historyCacheMutex.Lock()
	a.historyMetadataCache = make(map[string]*git.CommitDetails)
	a.cacheMetadata = false
	a.historyCacheMutex.Unlock()

	// Clear file history caches
	a.diffCacheMutex.Lock()
	a.fileHistoryDiffCache = make(map[string]string)
	a.fileHistoryFilesCache = make(map[string][]git.FileInfo)
	a.cacheDiffs = false
	a.diffCacheMutex.Unlock()

	// Reset state structures
	a.historyState = &ui.HistoryState{
		Commits:           make([]ui.CommitInfo, 0),
		SelectedIdx:       0,
		PaneFocused:       true,
		DetailsLineCursor: 0,
		DetailsScrollOff:  0,
	}

	a.fileHistoryState = &FileHistoryState{
		Commits:           make([]git.CommitInfo, 0),
		Files:             make([]ui.FileInfo, 0),
		SelectedCommitIdx: 0,
		SelectedFileIdx:   0,
		FocusedPane:       PaneCommits,
		CommitsScrollOff:  0,
		FilesScrollOff:    0,
		DiffScrollOff:     0,
	}

	// Restart loading (non-blocking, async goroutines)
	a.cacheLoadingStarted = false
	if a.gitState.Operation == git.Normal {
		a.cacheLoadingStarted = true
		go a.preloadHistoryMetadata()
		go a.preloadFileHistoryDiffs()
	}
}
