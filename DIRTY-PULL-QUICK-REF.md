# Dirty Pull — Quick Reference Card

## Components Created ✅

```
internal/git/dirtyop.go
├── DirtyOperationSnapshot struct
├── Save(branch, head) error
├── Load() error
├── Delete() error
├── IsDirtyOperationActive() bool
├── ReadSnapshotState() (branch, hash, error)
└── CleanupSnapshot() error

internal/app/dirtystate.go
├── DirtyOperationState struct
├── NewDirtyOperationState(opType, preserve) *State
├── SetPhase(newPhase)
├── MarkConflictDetected(phase, files)
└── ShouldStashDrop() bool

internal/app/operations.go
├── cmdDirtyPullSnapshot(preserve) tea.Cmd          [Phase 1]
├── cmdDirtyPullMerge() tea.Cmd                       [Phase 2a]
├── cmdDirtyPullRebase() tea.Cmd                      [Phase 2b]
├── cmdDirtyPullApplySnapshot() tea.Cmd               [Phase 3]
├── cmdDirtyPullFinalize() tea.Cmd                    [Phase 4]
└── cmdAbortDirtyPull() tea.Cmd                       [Abort anywhere]
```

## Operation Phases

| Phase | Command | What It Does | Success → | Fail → |
|-------|---------|--------------|-----------|--------|
| 1 | `cmdDirtyPullSnapshot()` | Save branch/HEAD, stash/discard changes | Phase 2 | Abort |
| 2a | `cmdDirtyPullMerge()` | `git pull` | Phase 3 | Conflict Resolver |
| 2b | `cmdDirtyPullRebase()` | `git pull --rebase` | Phase 3 | Conflict Resolver |
| 3 | `cmdDirtyPullApplySnapshot()` | `git stash apply` | Phase 4 | Conflict Resolver |
| 4 | `cmdDirtyPullFinalize()` | `git stash drop` + cleanup | Menu | (non-fatal) |
| ✗ | `cmdAbortDirtyPull()` | Reset branch/HEAD, reapply stash | Menu | Error (rare) |

## Files Being Modified

| File | Change | Line Range | Purpose |
|------|--------|-----------|---------|
| `internal/git/types.go` | Add `DirtyOperation` const | +1 line | State enum |
| `internal/git/state.go` | Add `detectDirtyOperation()` + check in `DetectState()` | +5 lines | State detection |
| `internal/app/menu.go` | Add dirty pull item in `menuTimeline()` Behind case | +15 lines | Menu visibility |
| `internal/app/dispatchers.go` | Add `dispatchDirtyPull()` | +10 lines | Menu → Confirmation |
| `internal/app/confirmationhandlers.go` | Add `handleDirtyPullConfirm()` | +20 lines | Confirmation → Operation |
| `internal/app/app.go` | Add `dirtyOperationState` field, update Update() | +10 lines | State tracking |
| `internal/app/handlers.go` | Route confirmation action to dirty pull handler | +3 lines | Confirmation routing |
| `internal/app/githandlers.go` | Add `handleGitOperationMsg()` for operation chain | +60 lines | Operation routing |
| `internal/app/conflicthandlers.go` | Implement `handleConflictEnter()` + `handleConflictEsc()` | +40 lines | Conflict handling |
| `internal/git/execute.go` | Add `ReadConflictFiles()` helper | +50 lines | Conflict file reading |

## Test Scenarios (titest.sh)

```bash
Scenario 0: Reset to fresh
Scenario 1: Pull with conflicts (clean tree)
Scenario 2: Dirty pull - merge with conflicts        ← TEST THIS
Scenario 3: Dirty pull - rebase with conflicts       ← TEST THIS
Scenario 4: Dirty pull - stash apply conflicts       ← TEST THIS
Scenario 5: Dirty pull - clean pull (no conflicts)   ← TEST THIS
```

## Key Git Commands Used

```bash
# Snapshot phase
git symbolic-ref --short HEAD           # Get branch name
git rev-parse HEAD                       # Get commit hash
git stash push -u -m "TIT DIRTY-PULL SNAPSHOT"    # Save changes
git reset --hard                         # Discard changes
git clean -fd                            # Clean untracked files

# Apply phases
git pull [--rebase]                      # Pull from remote
git stash apply                          # Reapply stashed changes
git stash drop                           # Drop stash after success

# Abort phase
git checkout <branch>                    # Return to original branch
git reset --hard <commit>                # Return to original HEAD

# Conflict reading
git status --porcelain=v2               # Find unmerged files
git show :1:<file>                       # Base version
git show :2:<file>                       # Local version
git show :3:<file>                       # Remote version
```

## State Machine

```
NotRepo ← (always, not in git repo)
Normal ← (clean repo, no operations)
Conflicted ← (merge/rebase/dirty-op conflicts)
Merging ← (merge in progress, no conflicts)
Rebasing ← (rebase in progress, no conflicts)
DirtyOperation ← (detected by snapshot file present)
```

When DirtyOperation detected:
```
Menu → Only show: Continue / Abort
       (Cannot do anything else until dirty op resolves)
```

## Menu Integration

```
When: Timeline = Behind + WorkingTree = Modified + Operation = Normal

Show: "⚠️ Pull (save changes)" [shortcut: d]

User selects → dispatchDirtyPull()
            → ConfirmationDialog (Save? / Discard? / Cancel)
            → handleDirtyPullConfirm()
            → cmdDirtyPullSnapshot()
            → [continues to pull + conflict resolution]
```

## Error Handling

| Error | Recovery |
|-------|----------|
| Snapshot save fails | Delete snapshot file, return to menu |
| Pull fails (no conflict) | Delete snapshot, return to menu with error |
| Merge conflicts | Show conflict resolver, user resolves |
| Stash apply conflicts | Show conflict resolver, user resolves |
| Abort fails | User must manually `git reset --hard` |
| Stash drop fails | Warn but continue (stash can be cleaned up manually) |

## Conflict Resolver Integration

```
When conflicts detected in Phase 2 or 3:

ConflictResolveState:
- Operation: "dirty_pull_changeset_apply" OR "dirty_pull_snapshot_reapply"
- NumColumns: 3 (LOCAL / REMOTE / SNAPSHOT)
- ColumnLabels: ["LOCAL", "REMOTE", "SNAPSHOT"]

User marks files with SPACE (radio button per file)
Press ENTER → handleConflictEnter()
          → Stage marked files
          → Continue to next phase

Press ESC → handleConflictEsc()
         → Call cmdAbortDirtyPull()
         → Restore original state
         → Return to menu
```

## Testing Checklist

- [ ] Phase 1: DirtyOperation enum + detect in state ✓ (build clean)
- [ ] Phase 2: Menu items appear when Modified + Behind ✓
- [ ] Phase 3: Confirmation dialog shows on menu select ✓
- [ ] Phase 4: Snapshot created, pull begins ✓
- [ ] Phase 5: No conflicts → finalize succeeds ✓
- [ ] Phase 6: Changeset conflicts → resolver shows ✓
- [ ] Phase 7: Mark file, ENTER continues ✓
- [ ] Phase 8: Snapshot apply conflicts → resolver shows ✓
- [ ] Phase 9: Finalize cleans up stash + snapshot ✓
- [ ] Phase 10: ESC during conflict → abort restores state ✓

## Commands to Review

```bash
# Build (verifies clean compile)
./build.sh

# Run with test repo
cd /Users/jreng/Documents/Poems/inf/tit_test_repo
../tit_x64

# Setup test scenario
./titest.sh
# Select scenario (2-5)
```

## Documentation Files

```
DIRTY-PULL-AUDIT.md           ← Component inventory + checklist
DIRTY-PULL-NEXT-PHASES.md     ← Detailed implementation guide
DIRTY-PULL-QUICK-REF.md       ← This file
DIRTY-PULL-IMPLEMENTATION.md  ← Original spec (reference)
SPEC.md                        ← User-facing specification (sections 6.1-6.4)
```

---

**Status:** Foundation complete, ready for Phase 1 wiring.  
**Next:** Implement Phase 1 (state detection).

