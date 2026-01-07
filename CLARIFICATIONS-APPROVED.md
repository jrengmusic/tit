# Clarifications Approved ✅

**Date:** 2026-01-07  
**Status:** All 5 clarifications APPROVED for implementation

---

## Approved Decisions

### Q1: Cache Pre-Load Limit
**Decision:** Always 30 commits (fixed, not configurable)
- Provides reasonable exploration depth
- Manageable memory footprint (~30KB for metadata, ~1.8MB for diffs)
- Non-blocking startup (~1-2 seconds)
- ✅ APPROVED

### Q2: Diff Cache Threshold
**Decision:** Skip caching diffs for commits with >100 files
- Prevents performance lag for massive commits
- Diffs still work (fetched on-demand, not cached)
- Typical commits ~5-20 files → cached
- Large commits ~100+ files → fetched on-demand
- ✅ APPROVED

### Q3: Time Travel Merge Strategy
**Decision:** Create merge commit (preserve history)
- Safe, auditable, preserves merge history
- Alternative (rebase) would rewrite history
- Alternative (cherry-pick) would lose context
- ✅ APPROVED

### Q4: History Depth
**Decision:** Show last 30 commits (bounded memory)
- Same as pre-load limit (consistency)
- Sufficient for typical exploration
- Bounded memory usage
- Alternative (all commits) would grow unbounded
- ✅ APPROVED

### Q5: Cache Reload Timing
**Decision:** Immediate async reload after commit success
- Non-blocking UI (async goroutines)
- Responsive menu (not delayed)
- Users see progress indication while reloading
- Alternative (manual refresh) would be explicit but less convenient
- ✅ APPROVED

---

## Implementation Constants (Finalized)

```go
const (
    PreloadCommitLimit      = 30     // Always load last 30 commits
    DiffCacheSkipThreshold  = 100    // Skip caching for >100 files
    DiffCacheVersion        = 2      // Cache both "parent" and "wip"
    TimeTravelMergeMethod   = "merge" // git merge (not rebase/cherry-pick)
    CacheReloadAsync        = true    // Non-blocking reload
)
```

---

## Ready for Phase 1

All decisions documented, approved, and ready to implement.

**Next:** Proceed to Phase 1 Implementation
