package app

import (
	"sync"
	"tit/internal/git"
)

// CacheManager handles async metadata and diff caching for history modes.
// Thread-safe: All operations protected by mutexes.
//
// LOCK ORDER (CRITICAL - prevents deadlock):
//  1. historyMutex (first)
//  2. diffMutex (second)
//
// NEVER acquire diffMutex before historyMutex.
type CacheManager struct {
	// Data caches
	metadataCache map[string]*git.CommitDetails // hash → commit metadata
	diffCache     map[string]string             // hash:path:version → diff content
	filesCache    map[string][]git.FileInfo     // hash → file list

	// Status flags
	loadingStarted bool
	metadataReady  bool
	diffsReady     bool

	// Progress tracking
	metadataProgress int
	metadataTotal    int
	diffsProgress    int
	diffsTotal       int
	animationFrame   int

	// Thread safety (LOCK ORDER: historyMutex → diffMutex)
	historyMutex sync.Mutex
	diffMutex    sync.Mutex
}

// NewCacheManager creates initialized CacheManager.
func NewCacheManager() *CacheManager {
	return &CacheManager{
		metadataCache: make(map[string]*git.CommitDetails),
		diffCache:     make(map[string]string),
		filesCache:    make(map[string][]git.FileInfo),
	}
}

// --- Status Methods ---

// IsLoadingStarted returns whether cache loading has been initiated.
func (c *CacheManager) IsLoadingStarted() bool {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	return c.loadingStarted
}

// SetLoadingStarted marks cache loading as initiated.
func (c *CacheManager) SetLoadingStarted(started bool) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.loadingStarted = started
}

// IsMetadataReady returns whether metadata cache is ready.
func (c *CacheManager) IsMetadataReady() bool {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	return c.metadataReady
}

// SetMetadataReady marks metadata cache as ready.
func (c *CacheManager) SetMetadataReady(ready bool) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataReady = ready
}

// IsDiffsReady returns whether diffs cache is ready.
func (c *CacheManager) IsDiffsReady() bool {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	return c.diffsReady
}

// SetDiffsReady marks diffs cache as ready.
func (c *CacheManager) SetDiffsReady(ready bool) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsReady = ready
}

// --- Progress Methods ---

// GetMetadataProgress returns metadata loading progress.
func (c *CacheManager) GetMetadataProgress() (progress, total int) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	return c.metadataProgress, c.metadataTotal
}

// SetMetadataProgress updates metadata loading progress.
func (c *CacheManager) SetMetadataProgress(progress, total int) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataProgress = progress
	c.metadataTotal = total
}

// GetDiffsProgress returns diffs loading progress.
func (c *CacheManager) GetDiffsProgress() (progress, total int) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	return c.diffsProgress, c.diffsTotal
}

// SetDiffsProgress updates diffs loading progress.
func (c *CacheManager) SetDiffsProgress(progress, total int) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsProgress = progress
	c.diffsTotal = total
}

// GetAnimationFrame returns current animation frame.
func (c *CacheManager) GetAnimationFrame() int {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	return c.animationFrame
}

// IncrementAnimationFrame advances animation frame.
func (c *CacheManager) IncrementAnimationFrame() int {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.animationFrame++
	return c.animationFrame
}

// --- Metadata Cache Methods ---

// GetMetadata returns cached commit details.
func (c *CacheManager) GetMetadata(hash string) (*git.CommitDetails, bool) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	details, ok := c.metadataCache[hash]
	return details, ok
}

// SetMetadata stores commit details in cache.
func (c *CacheManager) SetMetadata(hash string, details *git.CommitDetails) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataCache[hash] = details
}

// GetAllMetadata returns copy of all cached metadata (for iteration).
func (c *CacheManager) GetAllMetadata() map[string]*git.CommitDetails {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	copy := make(map[string]*git.CommitDetails)
	for k, v := range c.metadataCache {
		copy[k] = v
	}
	return copy
}

// --- Diff Cache Methods ---

// GetDiff returns cached diff content.
func (c *CacheManager) GetDiff(key string) (string, bool) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	diff, ok := c.diffCache[key]
	return diff, ok
}

// SetDiff stores diff content in cache.
func (c *CacheManager) SetDiff(key string, diff string) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffCache[key] = diff
}

// --- Files Cache Methods ---

// GetFiles returns cached file list for commit.
func (c *CacheManager) GetFiles(hash string) ([]git.FileInfo, bool) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	files, ok := c.filesCache[hash]
	return files, ok
}

// SetFiles stores file list in cache.
func (c *CacheManager) SetFiles(hash string, files []git.FileInfo) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.filesCache[hash] = files
}

// --- Invalidation ---

// Invalidate clears all caches and resets state.
// IMPORTANT: Acquires locks in correct order (history → diff).
func (c *CacheManager) Invalidate() {
	// Lock order: historyMutex first, then diffMutex
	c.historyMutex.Lock()
	c.metadataCache = make(map[string]*git.CommitDetails)
	c.metadataReady = false
	c.metadataProgress = 0
	c.metadataTotal = 0
	c.loadingStarted = false
	c.animationFrame = 0
	c.historyMutex.Unlock()

	c.diffMutex.Lock()
	c.diffCache = make(map[string]string)
	c.filesCache = make(map[string][]git.FileInfo)
	c.diffsReady = false
	c.diffsProgress = 0
	c.diffsTotal = 0
	c.diffMutex.Unlock()
}

// --- Bulk Operations (for cache building) ---

// InitMetadataLoading prepares for metadata cache building.
func (c *CacheManager) InitMetadataLoading(total int) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataReady = false
	c.metadataProgress = 0
	c.metadataTotal = total
}

// InitDiffsLoading prepares for diffs cache building.
func (c *CacheManager) InitDiffsLoading(total int) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsReady = false
	c.diffsProgress = 0
	c.diffsTotal = total
}

// UpdateMetadataProgress updates progress during cache building.
func (c *CacheManager) UpdateMetadataProgress(progress int) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataProgress = progress
}

// UpdateDiffsProgress updates progress during cache building.
func (c *CacheManager) UpdateDiffsProgress(progress int) {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsProgress = progress
}

// FinalizeMetadata marks metadata cache as complete.
func (c *CacheManager) FinalizeMetadata() {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataReady = true
}

// FinalizeDiffs marks diffs cache as complete.
func (c *CacheManager) FinalizeDiffs() {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsReady = true
}

// InvalidateMetadata clears only metadata cache.
func (c *CacheManager) InvalidateMetadata() {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataCache = make(map[string]*git.CommitDetails)
	c.metadataReady = false
	c.metadataProgress = 0
	c.loadingStarted = false
}

// InvalidateDiffs clears only diffs and files caches.
func (c *CacheManager) InvalidateDiffs() {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffCache = make(map[string]string)
	c.filesCache = make(map[string][]git.FileInfo)
	c.diffsReady = false
	c.diffsProgress = 0
}

// ResetMetadataProgress resets metadata progress counters.
func (c *CacheManager) ResetMetadataProgress() {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()
	c.metadataProgress = 0
	c.metadataTotal = 0
}

// ResetDiffsProgress resets diffs progress counters.
func (c *CacheManager) ResetDiffsProgress() {
	c.diffMutex.Lock()
	defer c.diffMutex.Unlock()
	c.diffsProgress = 0
	c.diffsTotal = 0
}
