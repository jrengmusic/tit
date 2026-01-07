# History & File(s) History Implementation - START HERE

**Date:** 2026-01-07  
**Status:** 🟢 Analysis Complete, Ready for Phase 1  
**Total Documentation:** ~1,465 lines across 3 documents

---

## 📖 Quick Navigation

You have three comprehensive documents. Here's how to use them:

### 1. **READ FIRST** 👇 (This file)
**HISTORY-START-HERE.md** - You are here
- Overview and document map
- Key takeaways
- 5 clarification questions that need answers

**Time to read:** 5 minutes

---

### 2. **READ SECOND** - High Level Overview
**HISTORY-IMPLEMENTATION-SUMMARY.md** (12 KB, 407 lines)
- Visual summary and architecture overview
- Core problem statement
- 5 key design decisions
- Data flow diagram
- Testing requirements (26 items)
- Timeline estimate (~1 week)

**Time to read:** 15 minutes  
**When to use:** Get a visual understanding before diving into details

---

### 3. **READ THIRD** - Main Reference
**HISTORY-IMPLEMENTATION-PLAN.md** (25 KB, 766 lines)
- **THE MAIN DOCUMENT** - Everything is here
- Complete 9-phase breakdown with acceptance criteria
- All moving parts documented
- Risk assessment with mitigation strategies
- Data structures (Phase 1)
- Git commands required
- UI components, keyboard handlers, menu items
- Time Travel integration details (Phase 7)
- Testing strategy with 26 manual test items

**Time to read:** 45 minutes (reference document, not straight through)  
**When to use:** As primary reference during implementation

---

### 4. **KEEP HANDY** - Quick Lookup
**HISTORY-QUICK-REFERENCE.md** (9.3 KB, 292 lines)
- Quick lookup tables for types, handlers, commands
- File structure map (what to create/modify)
- Phase execution checklist
- Important constants and limits
- Common pitfalls to avoid

**Time to use:** During implementation for quick answers  
**When to use:** Quick reference, not detailed explanation

---

## 🎯 TL;DR - Key Points

### What We're Building
- **History Mode:** Browse last 30 commits with metadata (author, date, message)
- **File(s) History Mode:** Browse file changes across commits with diffs
- **Time Travel:** Enter detached HEAD mode to explore old commits (read-only)
  - Replace old-tit's cherry-pick with time travel per SPEC.md § 9

### Complexity Highlights
1. ⚠️  **Caching System** - Two separate caches, thread-safe preload
2. ⚠️  **Mode Architecture** - Two new AppModes with unique UI
3. ⚠️  **State-Dependent Rendering** - Diffs change based on WorkingTree
4. ⚠️  **Time Travel Integration** - Replaces cherry-pick, new workflow
5. ⚠️  **Cache Invalidation** - Must refresh after commits

### Scale
- **Total code:** ~5,500 lines
- **New files:** 3 (historycache.go, history.go, filehistory.go)
- **Modified files:** 10 (add fields, handlers, menu items)
- **Reused components:** 3 (listpane.go, diffpane.go, theme.go)
- **Breaking changes:** ❌ NONE - All additions only

### Timeline
- Phase 1 (Infrastructure): ~1 day
- Phases 2-6 (Caching + UI + Handlers): ~3-4 days
- Phase 7 (Time Travel): ~1-2 days
- Phases 8-9 (Invalidation + Testing): ~1-2 days
- **Total: ~1 week** (assuming daily progress)

### Testing
- 26 manual test items (no automated tests per old-tit pattern)
- All must pass before completion
- Categories: History mode, File(s) History, Time Travel, Cache, Dirty tree, Edge cases

---

## ❓ FIVE CLARIFICATION QUESTIONS

**IMPORTANT:** Please answer these BEFORE starting Phase 1

### Q1: Cache Pre-Load Limit
**Current recommendation:** 30 commits

Should we always pre-load exactly 30 commits, or make it configurable?
- Impact: Determines memory usage & cache time
- Options:
  - [ ] Always 30 (fixed, simple, recommended)
  - [ ] Configurable (flexible, adds complexity)

### Q2: Diff Cache Threshold
**Current recommendation:** Skip >100 files per commit

Should we skip caching diffs for commits with >100 files?
- Impact: Performance & memory trade-off
- Options:
  - [ ] Yes, skip >100 (prevents lag, recommended)
  - [ ] Different number? (specify: ___ files)
  - [ ] Never skip (might lag for massive commits)

### Q3: Time Travel Merge Strategy
**Current recommendation:** Create merge commit

When merging time-travel changes back to original branch, use:
- Options:
  - [ ] Merge commit (safe, preserves history, recommended)
  - [ ] Rebase (cleaner history, complex)
  - [ ] Cherry-pick (isolated, no merge)

### Q4: History Depth
**Current recommendation:** Last 30 commits

Should History show:
- Options:
  - [ ] Last 30 commits (bounded memory, recommended)
  - [ ] All commits (unlimited, more data, slower)
  - [ ] Configurable (flexible, adds complexity)

### Q5: Cache Reload Timing
**Current recommendation:** Immediate (async)

After commit, should caches reload:
- Options:
  - [ ] Immediately in background (responsive, recommended)
  - [ ] On next menu entry (simpler, slightly delayed)
  - [ ] Only on manual refresh (manual, explicit)

---

## ✅ Document Checklist

Before implementing, verify you have all three documents:

- [ ] **HISTORY-IMPLEMENTATION-PLAN.md** (main reference)
  - [ ] Sections 1-5: Problem analysis
  - [ ] Sections 6-8: Data structures, git commands, UI
  - [ ] Sections 9-12: Phase breakdown, integration, testing

- [ ] **HISTORY-QUICK-REFERENCE.md** (quick lookup)
  - [ ] File structure map
  - [ ] Data types reference
  - [ ] Keyboard handlers table
  - [ ] Git commands reference
  - [ ] Phase checklist

- [ ] **HISTORY-IMPLEMENTATION-SUMMARY.md** (visual overview)
  - [ ] Architecture diagram
  - [ ] Data flow diagram
  - [ ] Key design decisions
  - [ ] Testing requirements
  - [ ] Timeline estimate

---

## 🚀 Next Steps

### Before Phase 1 Starts
1. ✅ Read this file (HISTORY-START-HERE.md)
2. ✅ Read HISTORY-IMPLEMENTATION-SUMMARY.md
3. ✅ Answer the 5 clarification questions above
4. ✅ Confirm Phase 1 is ready to implement

### Phase 1: Infrastructure & UI Types (Day 1)
1. Add `CommitInfo`, `CommitDetails`, `FileInfo` to `internal/git/types.go`
2. Add `HistoryState`, `FileHistoryState` fields to `Application` struct
3. Add `ModeHistory`, `ModeFileHistory` to `modes.go`
4. Verify compilation
5. **Code review** before Phase 2

### Phases 2-6: Core Implementation (Days 2-5)
- Each phase follows same pattern: Implementation → Code Review → Testing
- Reference HISTORY-QUICK-REFERENCE.md for file structure & checklist

### Phase 7: Time Travel Integration (Days 6-7)
- Most complex phase
- New menu generator + handlers + git operations
- Strong testing required

### Phases 8-9: Refinement & Testing (Days 8+)
- Cache invalidation after commits
- All 26 manual test items
- Final verification

---

## 💡 Key Design Decisions

All decisions are documented and reasoned in HISTORY-IMPLEMENTATION-PLAN.md § 4:

1. ✅ **Two Separate Caches**
   - History metadata (small, always cached)
   - File(s) diffs (large, selective caching)

2. ✅ **Menu Disabling Until Cache Ready**
   - Menu items show "(Loading...)" until cache available
   - UX clarity: user knows what's happening

3. ✅ **State-Dependent Diff Display**
   - Clean working tree → show "commit vs parent"
   - Dirty working tree → show "commit vs WIP"
   - Cache both versions for instant switching

4. ✅ **Time Travel Replaces Cherry-Pick**
   - Old-tit: ENTER = cherry-pick (apply commit)
   - New TIT: ENTER = time travel (explore commit, read-only)
   - Per SPEC.md § 9

5. ✅ **Cache Invalidation After Commits**
   - Old caches cleared after commit succeeds
   - Caches reload async (non-blocking)
   - Menu stays responsive

---

## ⚠️ Critical Constraints

These are non-negotiable design constraints:

### Thread Safety
- ✅ All cache mutations behind mutex
- ✅ Goroutines read-only (immutable snapshots)
- ❌ Never mutate Application from goroutine

### State Consistency
- ✅ gitState always reflects actual git state
- ✅ Menu items disabled until cache ready
- ✅ No invalid state transitions

### Memory Management
- ✅ Skip diff caching for >100-file commits
- ✅ Limit preload to 30 commits
- ✅ No pointers in cache (use values)

### Architecture Compliance
- ✅ Single-threaded UI (Bubble Tea event loop)
- ✅ Worker threads for git ops (goroutines)
- ✅ Immutable message passing
- ✅ Key handler registry pattern

---

## 📋 Success Criteria

At end of Phase 9, all of these must be true:

1. ✅ History mode shows last 30 commits with metadata
2. ✅ File(s) History shows commits, files, diffs (state-dependent)
3. ✅ Time Travel works from History mode (read-only)
4. ✅ Can merge changes back from time-travel mode
5. ✅ Caches refresh after commits
6. ✅ All keyboard shortcuts work
7. ✅ No race conditions or deadlocks
8. ✅ **All 26 manual tests pass** ← Most important
9. ✅ Code follows TIT architecture patterns
10. ✅ No existing functionality broken

---

## 🔗 Document Links

Within `/Users/jreng/Documents/Poems/inf/tit/`:

- **HISTORY-START-HERE.md** (this file) - Entry point
- **HISTORY-IMPLEMENTATION-SUMMARY.md** - Visual overview
- **HISTORY-IMPLEMENTATION-PLAN.md** - Main reference
- **HISTORY-QUICK-REFERENCE.md** - Quick lookup

Related documents:
- **ARCHITECTURE.md** - New TIT architecture patterns
- **SPEC.md** - Specification (especially § 7, 9 for History & Time Travel)
- **CLAUDE.md** - Design guidance

---

## ❓ Questions?

If anything is unclear:

1. Check HISTORY-IMPLEMENTATION-PLAN.md § 12 (Clarifications section)
2. Review HISTORY-QUICK-REFERENCE.md for quick answers
3. Consult SPEC.md § 7 (History) and § 9 (Time Travel)
4. Ask before implementing - better to clarify now than fix later

---

## 🎯 Ready to Proceed?

### Checklist Before Phase 1:

- [ ] Read all 3 documents
- [ ] Answer 5 clarification questions
- [ ] Understand the 5 key design decisions
- [ ] Know the 9 phases and their order
- [ ] Know the 26 manual test items
- [ ] Understand thread-safe caching requirement
- [ ] Ready to implement Phase 1 (git types + app fields)

### When Ready:
**Proceed to Phase 1 Implementation** → See HISTORY-IMPLEMENTATION-PLAN.md § Phase 1

**Confidence Level:** 🟢 Very High
- All moving parts documented
- All decisions reasoned and explained
- Architecture verified against constraints
- Timeline realistic
- Testing strategy clear

---

**Status:** ✅ Analysis Complete
**Phase 1 Ready:** ✅ Yes
**Proceed When:** You've answered the 5 questions

Good luck! 🚀
