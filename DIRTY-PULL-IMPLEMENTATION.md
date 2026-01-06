
## Final TIT Artifact — Dirty Pull (Minimal, Source-Agnostic)

No upstream. No cherry-pick. No merge semantics.

---

# TIT Artifact — Dirty Pull

## 1. Intent

Apply a change-set while preserving local uncommitted work.

---

## 2. Core Invariants

1. Abort restores the **exact original dirty state**
2. Conflict state blocks progress
3. Snapshot is internal and temporary
4. No stash stacking, no pop

---

## 3. Flow Overview

```
[Confirm]
   ↓
[Snapshot]
   ↓
[Apply Change-Set]
   ↓
[Conflict?] ── yes ──▶ [Resolve]
   ↓ no
[Apply Snapshot]
   ↓
[Conflict?] ── yes ──▶ [Resolve]
   ↓ no
[Finalize]
```

---

## 4. Snapshot (Safety Net)

```
ORIGINAL_HEAD = git rev-parse HEAD
git stash push -u -m "TIT DIRTY-PULL SNAPSHOT"
```

Result:

* clean tree
* preserved user state

---

## 5. Apply Change-Set

Abstract operation:

```
apply-change-set
```

Dirty Pull does **not** define how this works.

---

## 6. Conflict Detection (Priority)

After **any** apply step:

Indicators:

* non-zero exit code
* unmerged paths in `git status --porcelain`
* `CONFLICT` in command output

Rule:

> **If conflicts exist, stop and require resolution.**

---

## 7. Conflict Resolution Contract

User must:

* resolve conflicts
* stage resolutions

Completion signal:

```
git status --porcelain
```

→ no unmerged entries

---

## 8. Apply Snapshot

```
git stash apply
```

Same conflict rules apply.

---

## 9. Abort (Universal)

```
git reset --hard ORIGINAL_HEAD
git stash apply
```

---

## 10. Finalize

```
git stash drop
```

---

## 11. Edge Safety

* Abort at any step → original dirty state
* Crash-safe → stash persists
* Idempotent abort

---

## Canonical One-Liner

> **Dirty Pull in TIT is source-agnostic, conflict-first, and rollback-safe.**

---

