# TIT Refactoring - Quick Start Guide

**Priority 1 Projects** for immediate impact. Each takes <2 hours.

---

## 🔧 Project 1: Consolidate Status Bar Builders (HIGH IMPACT)

**Problem:** 4 identical status bar builder functions (history, filehistory, conflictresolver)

**Files Affected:**
- `internal/ui/history.go:158` - `buildHistoryStatusBar()`
- `internal/ui/filehistory.go:218` - `buildFileHistoryStatusBar()`
- `internal/ui/filehistory.go:259` - `buildDiffStatusBar()`
- `internal/ui/conflictresolver.go:182` - `buildGenericConflictStatusBar()`

**Solution:**

1. **Create template builder** (statusbar.go)
```go
type StatusBarParts struct {
    Left   string
    Center string
    Right  string
}

func BuildCustomStatusBar(parts StatusBarParts, width int, theme *Theme) string {
    config := StatusBarConfig{
        Parts:    []string{parts.Left, parts.Center, parts.Right},
        Width:    width,
        Centered: true,
        Theme:    theme,
    }
    return BuildStatusBar(config)
}
```

2. **Replace all 4 functions** with:
```go
// history.go
func buildHistoryStatusBar(paneFocused bool, width int, theme Theme) string {
    parts := StatusBarParts{
        Left:   // build left part
        Center: // build center part
        Right:  // build right part
    }
    return BuildCustomStatusBar(parts, width, &theme)
}
```

3. **Delete old status bar code** - reduce 300+ lines of duplication

**Benefit:** Single place to modify status bar styling globally

**Estimated Time:** 45 min  
**Lines Reduced:** ~300

---

## 🎨 Project 2: Extract Shortcut Style Helpers (EASY WIN)

**Problem:** This pattern repeats 4+ times:
```go
shortcutStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color(theme.AccentTextColor)).
    Bold(true)
```

**Files Affected:**
- `internal/ui/history.go:159`
- `internal/ui/filehistory.go:219`
- `internal/ui/conflictresolver.go:183`
- And more...

**Solution:**

Add to `internal/ui/theme.go`:
```go
// Add these methods to Theme struct
func (t *Theme) ShortcutStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(lipgloss.Color(t.AccentTextColor)).
        Bold(true)
}

func (t *Theme) DescriptionStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(lipgloss.Color(t.ContentTextColor))
}

func (t *Theme) HashHighlightStyle() lipgloss.Style {
    return lipgloss.NewStyle().
        Foreground(lipgloss.Color(t.AccentTextColor)).
        Bold(true)
}
```

**Replace all occurrences:**
```go
// Before:
shortcutStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color(theme.AccentTextColor)).
    Bold(true)

// After:
shortcutStyle := theme.ShortcutStyle()
```

**Benefit:** Single source for all style definitions; easy to modify all shortcuts at once

**Estimated Time:** 30 min  
**Lines Reduced:** ~20

---

## 🎯 Project 3: Centralize Operation Step Constants (SSOT)

**Problem:** Operation names hardcoded as strings throughout codebase

**Files Affected:**
- `internal/app/operationsteps.go` (partial)
- `internal/app/operations.go` (uses strings like "init", "clone", "push")
- `internal/app/githandlers.go`

**Current State:**
```go
// operationsteps.go defines SOME steps but not all
const (
    StepInitRepository = "init_repository"
    StepCreateBranch   = "create_branch"
    StepCommit         = "commit"
    // ... but "init", "clone", "push" are hardcoded elsewhere
)
```

**Solution:**

Expand operationsteps.go:
```go
// Operation step names (for GitOperationMsg.Step field)
const (
    OpInit       = "init"
    OpClone      = "clone"
    OpCommit     = "commit"
    OpCommitPush = "commit_push"
    OpPush       = "push"
    OpPull       = "pull"
    OpAddRemote  = "add_remote"
    OpFetchRemote = "fetch_remote"
    // ... all operations
)

// Cache types (for CacheProgressMsg.CacheType field)
const (
    CacheTypeMetadata = "metadata"
    CacheTypeDiffs    = "diffs"
)

// Time travel operations
const (
    TimeTravelCheckout = "time_travel_checkout"
    TimeTravelMerge    = "time_travel_merge"
    TimeTravelReturn   = "time_travel_return"
)
```

**Replace all hardcoded strings:**
```go
// Before:
return GitOperationMsg{Step: "init", Success: true}

// After:
return GitOperationMsg{Step: OpInit, Success: true}

// Before:
buffer.Append(fmt.Sprintf("Building history cache from branch %s...", ref), ui.TypeStatus)

// After:
return CacheProgressMsg{CacheType: CacheTypeMetadata, Current: i, Total: len(commits)}
```

**Benefit:** No typos in operation names; easy to find all usages of one operation

**Estimated Time:** 45 min  
**Coverage:** ~30 string replacements

---

## 🔐 Project 4: Document Cache Key Schema (SSOT)

**Problem:** Cache key formats scattered, no validation
```go
// fileHistoryDiffCache has implicit schema: "hash:path:version"
// But nowhere documented or validated
key := fmt.Sprintf("%s:%s:%s", hash, filepath, version)  // ← magic!
```

**Files Affected:**
- `internal/app/historycache.go` (builds cache)
- `internal/app/dispatchers.go` (reads cache)
- No single schema definition

**Solution:**

Add helpers to `historycache.go`:
```go
// Cache key format documentation
type CacheKeySchemas struct {
    // Metadata cache: simple commit hash
    // Example: "abc123def456abcdef1234567890abcdef123456"
    // Returns: git.CommitDetails
    
    // Diff cache: hash:filepath:version
    // Example: "abc123def456abcdef1234567890abcdef123456:src/main.go:after"
    // Returns: diff content string
    
    // Files cache: simple commit hash
    // Example: "abc123def456abcdef1234567890abcdef123456"
    // Returns: []git.FileInfo
}

// Builders
func DiffCacheKey(hash, filepath, version string) string {
    return fmt.Sprintf("%s:%s:%s", hash, filepath, version)
}

func FileCacheKey(hash string) string {
    return hash
}

func MetadataCacheKey(hash string) string {
    return hash
}

// Parsers with validation
func ParseDiffCacheKey(key string) (hash, filepath, version string, err error) {
    parts := strings.Split(key, ":")
    if len(parts) != 3 {
        return "", "", "", fmt.Errorf("invalid diff cache key: %s (expected 3 parts)", key)
    }
    return parts[0], parts[1], parts[2], nil
}
```

**Replace hardcoded key construction:**
```go
// Before:
key := fmt.Sprintf("%s:%s:%s", hash, filepath, "before")
diff := a.fileHistoryDiffCache[key]

// After:
key := DiffCacheKey(hash, filepath, "before")
diff := a.fileHistoryDiffCache[key]
```

**Benefit:** Single source for key format; validation prevents typos

**Estimated Time:** 30 min  
**Coverage:** ~15 key building/parsing locations

---

## 📋 Project 5: Pair Confirmation Handlers (SSOT)

**Problem:** confirm/reject handlers in separate maps (easy to miss pairing)

**Files Affected:**
- `internal/app/confirmationhandlers.go:38-75`

**Current State:**
```go
var confirmationActions = map[string]ConfirmationAction{
    "nested_repo_init": { ... handleConfirm ... },
    // ... 10 more
}

var confirmationRejectActions = map[string]ConfirmationAction{
    "nested_repo_init": { ... handleReject ... },
    // ... 10 more
}
```

**Problem:** If you add to one map and forget the other, no compiler error!

**Solution:**

Replace with:
```go
type ConfirmationActionPair struct {
    Confirm ConfirmationAction
    Reject  ConfirmationAction
}

var confirmationHandlers = map[string]ConfirmationActionPair{
    "nested_repo_init": {
        Confirm: ConfirmationAction{
            Handler: (*Application).executeConfirmNestedRepoInit,
        },
        Reject: ConfirmationAction{
            Handler: (*Application).executeRejectNestedRepoInit,
        },
    },
    "force_push": {
        Confirm: ConfirmationAction{
            Handler: (*Application).executeConfirmForcePush,
        },
        Reject: ConfirmationAction{
            Handler: (*Application).executeRejectForcePush,
        },
    },
    // ... all actions
}

// Usage:
actions := confirmationHandlers[actionID]
if confirmed {
    return actions.Confirm.Handler(a)
} else {
    return actions.Reject.Handler(a)
}
```

**Benefit:** Guaranteed pairing; can't accidentally miss confirm/reject pair

**Estimated Time:** 30 min  
**Impact:** 20 lines → 10 lines (code reduction) + safety

---

## 📊 Quick Impact Summary

| Project | Files | Lines Removed | Time | Safety |
|---------|-------|---------------|------|--------|
| 1. Status bars | 4 | ~300 | 45min | High |
| 2. Styles | 6+ | ~20 | 30min | High |
| 3. Constants | 5+ | ~50 | 45min | Very High |
| 4. Cache keys | 2 | ~10 | 30min | Very High |
| 5. Handlers | 1 | ~60 | 30min | Very High |
| **Total** | | **~440** | **2.5h** | |

---

## 🎯 Recommended Order

**Session 1:**
- Project 1: Status bar consolidation (45min)
- Project 2: Shortcut styles (30min)

**Session 2:**
- Project 3: Operation constants (45min)
- Project 4: Cache keys (30min)

**Session 3:**
- Project 5: Confirmation handlers (30min)

**Result:** ~2.5 hours of focused work removes 440 lines of duplication and improves maintainability significantly.

---

## ✅ Verification Checklist

After each project:
- [ ] Build succeeds: `./build.sh`
- [ ] No compiler errors or warnings
- [ ] Existing tests still pass (if applicable)
- [ ] Manual testing: feature still works end-to-end
- [ ] Search for old pattern confirms all replaced
- [ ] Code review: no missed locations

---

## 🔍 Finding All Occurrences

**Project 1 (Status bars):**
```bash
grep -r "buildHistoryStatusBar\|buildFileHistoryStatusBar\|buildDiffStatusBar\|buildGenericConflictStatusBar" internal/
```

**Project 2 (Shortcut styles):**
```bash
grep -r "lipgloss.NewStyle().*AccentTextColor.*Bold" internal/ui/
```

**Project 3 (Constants):**
```bash
grep -r '"init"\|"clone"\|"push"\|"pull"\|"commit"' internal/app/ | grep -v "// " | head -20
```

**Project 4 (Cache keys):**
```bash
grep -r 'fmt.Sprintf("%s:%s:%s"' internal/
```

**Project 5 (Handlers):**
```bash
grep -r "confirmationActions\|confirmationRejectActions" internal/app/
```

---

## 🚀 How to Execute

1. Create feature branch: `git checkout -b refactor/consolidate-patterns`
2. Work through projects one at a time
3. After each project: `./build.sh` + manual test
4. Commit: `git add -A && git commit -m "Refactor: consolidate [project name]"`
5. After all projects: Create PR with summary

**Total effort:** ~2.5 hours of focused work for significant maintainability improvement.
