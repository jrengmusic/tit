# Menu = Contract: Critical Architectural Fix

**Date:** 2026-01-08  
**Issue:** Menu items were showing options for incomplete git operations (merge/rebase/conflicts), violating the core design principle.

---

## The Contract Principle

**Menu = Contract:** If an action appears in the menu, it MUST succeed.

Never show operations that could:
- Fail due to git state
- Leave repo in a dangling/incomplete state
- Require manual user cleanup

This is TIT's foundational safety guarantee.

---

## Key Insight: Mid-Operation is Not a Menu State

**The critical realization:** Conflicted, Merging, Rebasing, and DirtyOperation are **not valid states for TIT to manage**. These are **pre-existing abnormal conditions** that the user must resolve externally.

TIT is designed to operate only on clean/normal repositories.

---

## Changes Applied

### 1. SPEC.md - Pre-Flight Checks (NEW § 5)

**Added new section before any menu generation:**

If git state detection finds:
- `Conflicted` (merge/rebase/conflict in progress)
- `Merging` (merge in progress, no conflicts)
- `Rebasing` (rebase in progress)
- `DirtyOperation` (stash mid-operation)

→ **TIT refuses to start** and shows fatal error:

```
⚠️ Git repository is in the middle of a [operation] operation

Your repository cannot be managed by TIT while an operation is in progress.
Please complete or abort the operation using standard git commands:

  git merge --continue / --abort
  git rebase --continue / --abort
  git stash pop

[Exit TIT]
```

### 2. SPEC.md - Removed All Mid-Op Menus

**DELETED entire sections:**
- `When Operation = Conflicted` - No menu shown
- `When Operation = Merging/Rebasing (no conflicts)` - No menu shown
- `When Operation = DirtyOperation` - No menu shown

**Renumbered all subsequent sections** (§ 5 → § 18)

### 3. SPEC.md - Updated Design Invariants (§ 17)

- **Item #1:** "Menu = Contract" - Emphasized prevention of failing operations
- **Item #7:** Changed from abort-based to "Automatically managed by TIT"
- **Item #11 (NEW):** "No Dangling States: These operations don't appear in TIT—they're resolved externally"

### 4. ARCHITECTURE.md - Architectural Clarity

**Updated MenuGenerator pattern:**
- Removed `menuConflicted()`, `menuOperation()`, `menuDirtyOperation()`
- Only valid generators: `menuNotRepo()`, `menuNormal()`, `menuTimeTraveling()`
- Added comment: "These states cause startup failure, never reach menuGenerators"

**Added startup check documentation:**
```go
// In app.New() or main():
state := git.DetectState()
if state.Operation == Conflicted || state.Operation == Merging || 
   state.Operation == Rebasing || state.Operation == DirtyOperation {
    showFatalError("Repository in mid-operation state")
    os.Exit(1)
}
// ... proceed with normal TIT initialization
```

---

## User Experience Impact

**Before:**
```
Repository state: Merging
App starts TIT with menu:
- Continue operation
- View details
- (Maybe abort?)
```
❌ Confusing. Is abort even safe?

**After:**
```
Repository state: Merging
App: ⚠️ Cannot start TIT
User: Use standard git commands to complete
$ git merge --continue
$ git merge --abort
(TIT starts after repo is clean)
```
✅ Clear. Zero ambiguity.

---

## Architectural Benefits

1. **Simpler:** No mid-op menu handlers
2. **Safer:** Impossible to show failing operations
3. **Clearer:** Users know to use `git` CLI for unusual states
4. **Robust:** Zero dangling state risk from TIT itself

---

## Related Documentation

- **SPEC.md § 5:** Pre-Flight Checks (new)
- **SPEC.md § 17:** Design Invariants (updated)
- **ARCHITECTURE.md § Menu System:** MenuGenerator pattern (updated)
- **CLAUDE.md § Design Decisions:** Can add "Why no mid-op handling" rationale

---

## Summary

**Strict Design Rule:** TIT only operates on clean/normal repositories. If git is mid-operation, TIT refuses to start. User resolves externally. No exceptions. No dangling states.

This eliminates an entire class of bugs and confusion.
