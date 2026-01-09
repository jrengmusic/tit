# TIT Codebase Audit Summary - 2026-01-10

**Comprehensive analysis of refactoring opportunities, SSOT violations, and pattern consolidations**

---

## 📊 At a Glance

**Total Opportunities Identified:** 15  
**No Critical Violations:** ✅ (Architecture is clean)  
**Potential Code Reduction:** ~440 lines  
**Estimated Effort:** 2.5 hours (for Priority 1)  

---

## 🎯 KEY FINDINGS

### Strengths
✅ **Strong SSOT enforcement** - MenuItem, Theme, Message maps well-centralized  
✅ **Clean architecture** - Clear layer separation (app → git → ui)  
✅ **Proper async patterns** - cmd* functions correctly implemented  
✅ **Good naming** - Most functions follow verb-noun convention  
✅ **Fail-fast principle** - Error handling mostly correct  

### Opportunities
🔧 **Status bar code (4 functions)** - 70% duplication, consolidate into template  
🎨 **Style builders (6+ locations)** - Extract lipgloss patterns to theme helpers  
🔐 **Confirmation handlers (2 maps)** - Pair confirm/reject in single map  
📋 **Operation constants (scattered)** - Centralize in operationsteps.go  
💾 **Cache key formats** - Document schema + add validators  
📍 **Navigation** - File organization could be improved (defer)  

---

## 📈 PRIORITY BREAKDOWN

### PRIORITY 1: High Impact, Low Effort (Do This Week)
| Item | Files | Impact | Effort | SSOT |
|------|-------|--------|--------|------|
| **Status bar consolidation** | 4 | ~300 lines removed | 45min | ✅ |
| **Shortcut style helpers** | 6+ | Theme centralization | 30min | ✅ |
| **Operation step constants** | 5+ | Remove string magic | 45min | ✅ |
| **Cache key schema** | 2 | Add validation | 30min | ✅ |

**Total Priority 1:** 150 min = 2.5 hours, removes ~330 lines, no breaking changes

### PRIORITY 2: Medium Impact (Next Sprint)
| Item | Files | Impact | Effort | SSOT |
|------|-------|--------|--------|------|
| **Confirmation handler pairing** | 1 | Safety improvement | 30min | ✅ |
| **Error handling standardization** | 8+ | Consistency | 1h | ✅ |
| **Message domain reorganization** | 1 | Better navigation | 1.5h | ✅ |
| **Mode metadata** | 1 | Documentation | 30min | ✅ |

**Total Priority 2:** 3.5 hours, improves maintainability significantly

### PRIORITY 3: Large Refactor (Defer)
| Item | Files | Impact | Effort | SSOT |
|------|-------|--------|--------|------|
| **Feature-based file organization** | All | Better navigation | 8+ hours | – |
| **Type definition consolidation** | 3 | Reduce duplicates | 2h | ✅ |

**Total Priority 3:** Only do if navigation becomes pain point

---

## 🔍 DETAILED ISSUES

### Issue 1: Status Bar Builders (70% Code Duplication)
**Locations:** history.go:158, filehistory.go:218, filehistory.go:259, conflictresolver.go:182  
**Pattern:** Build `[]string{left, center, right}`, call `BuildStatusBar()`, apply theme  
**Duplication:** Same logic in 4 places  
**Solution:** Extract template factory in statusbar.go  
**Impact:** Remove ~300 lines, single styling source  
**Recommendation:** ✅ Do this first (quick win)

### Issue 2: Shortcut Styles Scattered
**Locations:** history.go:159, filehistory.go:219, conflictresolver.go:183, and more  
**Pattern:** `lipgloss.NewStyle().Foreground(...).Bold(true)`  
**Duplication:** Same style code in 6+ places  
**Solution:** Add theme methods: `theme.ShortcutStyle()`, `theme.DescriptionStyle()`  
**Impact:** 1-line calls instead of 3-line builders  
**Recommendation:** ✅ Do this after status bars

### Issue 3: Operation Step Constants Hardcoded
**Locations:** operations.go, githandlers.go, app.go, historycache.go  
**Problem:** Strings like "init", "clone", "push" hardcoded throughout  
**Example:** `GitOperationMsg{Step: "init"}` (error-prone)  
**Solution:** Use operationsteps.go constants: `GitOperationMsg{Step: OpInit}`  
**Impact:** No string typos, find all usages easily  
**Recommendation:** ✅ Do this for safety

### Issue 4: Cache Key Format Undocumented
**Locations:** historycache.go, dispatchers.go  
**Problem:** Key format (e.g., "hash:path:version") implicit, no validation  
**Risk:** Typos in key construction cause silent cache misses  
**Solution:** Add `DiffCacheKey()`, `ParseDiffCacheKey()` helpers with validation  
**Impact:** Catch typos early, self-documenting code  
**Recommendation:** ✅ Do this for reliability

### Issue 5: Confirmation Handlers Split Across 2 Maps
**Location:** confirmationhandlers.go:38-75  
**Problem:** Confirm/reject handlers in separate maps (confirmationActions, confirmationRejectActions)  
**Risk:** Easy to add to one map and forget the other  
**Solution:** Single paired map: `confirmationHandlers[actionID] = ConfirmationActionPair{Confirm: ..., Reject: ...}`  
**Impact:** Compile-time guarantee of pairing  
**Recommendation:** ✅ Do this for safety

### Issue 6: Message Maps (11 Separate Maps)
**Location:** messages.go:100-340  
**Problem:** 11 different maps with related messages  
**Example:** "clone" action requires updates in: InputPrompts, InputHints, ErrorMessages, OutputMessages, FooterHints, ...  
**Solution:** Domain-scoped maps or nested structures  
**Impact:** All clone messages in one place  
**Recommendation:** ⏳ Do if maintenance pain increases

### Issue 7: Pane Height Calculations Scattered
**Locations:** history.go:47-51, filehistory.go:66-72, conflictresolver.go:45-57  
**Problem:** Similar height calculations in 3 places, slightly different  
**Risk:** Inconsistent UI if one is updated but others aren't  
**Solution:** Centralize in sizing.go: `CalculatePaneHeights(height, numPanes)`  
**Impact:** Consistency guaranteed  
**Recommendation:** ⏳ Do this sprint

### Issue 8: Type Definitions Duplicated
**Locations:** git/types.go + ui/history.go (CommitInfo), ui/filehistory.go (FileInfo)  
**Problem:** Same types defined twice to avoid import cycles  
**Risk:** If git.CommitInfo changes, ui copy might not  
**Solution:** Use type aliases: `type CommitInfo = git.CommitInfo`  
**Impact:** Single source of truth  
**Recommendation:** ⏳ Do if type changes become frequent

### Issue 9: Error Handling Inconsistent
**Locations:** githandlers.go, confirmationhandlers.go, app.go (scattered)  
**Pattern:** Some errors logged, some panicked, some silent  
**Problem:** No consistent recovery pattern  
**Solution:** Standardize with ErrorConfig{Level, Message, Buffer, Footer}  
**Impact:** Consistent user experience  
**Recommendation:** ⏳ Do next sprint

### Issue 10: Handler Naming Inconsistent
**Examples:** 
- `executeTimeTravelClean()` → should be `cmdTimeTravelClean()` (returns tea.Cmd)
- `executeCloneWorkflow()` → should be `cmdCloneWorkflow()`
- `handleInitBranchesSubmit()` → correct pattern

**Impact:** Code matches documentation  
**Recommendation:** ⏳ Low priority (documentation only)

---

## 📁 FILES TO FOCUS ON

### Top 3 Files to Review
1. **internal/ui/statusbar.go** (42 lines) - Template for consolidation patterns
2. **internal/app/menuitems.go** (260 lines) - Perfect SSOT example to follow
3. **internal/app/messages.go** (340 lines) - See where organization helps

### Files With Most Duplication
1. **internal/ui/history.go** - Status bar, shortcut styles
2. **internal/ui/filehistory.go** - Status bar (2 builders!), shortcut styles
3. **internal/ui/conflictresolver.go** - Status bar, shortcut styles
4. **internal/app/confirmationhandlers.go** - Paired map opportunity

---

## 🚀 ACTION PLAN

### Week 1: Priority 1 (2.5 hours focused work)
- [ ] Session 1: Status bar consolidation + shortcut styles (1.25h)
  - Build: `./build.sh`
  - Test: Manual E2E test of history/filehistory modes
- [ ] Session 2: Operation constants + cache keys (1.25h)
  - Build: `./build.sh`
  - Test: Time travel feature still works
- [ ] Commit: `git add -A && git commit -m "Refactor: consolidate duplicated patterns"`

### Week 2: Priority 2 (3.5 hours)
- [ ] Confirmation handler pairing (30min)
- [ ] Error handling standardization (1h)
- [ ] Consider message domain reorganization (1.5h)
- [ ] Commit in separate PR

### Week 3+: Priority 3 (Defer)
- Only if codebase navigation becomes pain point
- Feature-based file organization (8+ hours)

---

## 📊 BEFORE/AFTER METRICS

### Code Reduction
- **Before:** ~300 lines of duplicate status bar code
- **After:** ~50 lines of template + 4 thin wrappers = **~85% reduction**

### Maintainability
- **Before:** Change status bar style → modify 4 files
- **After:** Change status bar style → modify 1 file

### Safety
- **Before:** Confirmation handler pairing could be missed
- **After:** Pair guaranteed by struct design

### Consistency
- **Before:** 6+ places with shortcut style code
- **After:** 1 place (theme.ShortcutStyle())

---

## ✅ QUALITY GATES

All refactorings must:
- [ ] Build cleanly: `./build.sh`
- [ ] No new warnings/errors
- [ ] Feature-parity verified (E2E testing)
- [ ] Git history clean (one commit per refactor)
- [ ] Code review: no business logic changes

---

## 📝 DOCUMENTATION

See detailed analysis in:
- **CODEBASE-REFACTORING-AUDIT.md** - Complete analysis with code examples
- **REFACTORING-QUICK-START.md** - Step-by-step guide for top 5 projects

---

## 🎓 CONCLUSIONS

### What's Good
The codebase is **well-architected** with clean separation of concerns. SSOT pattern is enforced well (MenuItem, Theme). Async operations properly implemented. Error handling mostly correct.

### What Needs Improvement
**Duplication** is the main issue—4 status bar builders, 6+ shortcut style definitions, scattered constants. **Organization** could be improved (11 message maps). **Naming** mostly consistent but some execute* should be cmd*.

### Bottom Line
**No critical issues.** 2.5 hours of focused consolidation removes 440 lines and significantly improves maintainability. Recommended to do Priority 1 this week, Priority 2 next sprint. Priority 3 only if needed.

---

**Next Step:** Start with Project 1 (Status Bar Consolidation) in REFACTORING-QUICK-START.md

