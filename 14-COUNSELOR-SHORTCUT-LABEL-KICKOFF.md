# Sprint 12: Shortcut Label Fix - Kickoff Plan

**Role:** COUNSELOR
**Agent:** Copilot (claude-opus-4.5)
**Date:** 2026-01-26
**Objective:** Fix shift+bracket shortcuts and widen shortcut column

---

## Problem Statement

`shift+]` and `shift+[` shortcuts don't work because:
1. **Bubble Tea sends character, not modifier string** - When user presses Shift+[, terminal sends `{` character (not "shift+[")
2. **Shortcut field used for both binding AND display** - Currently `Shortcut: "shift+]"` is used for key matching (fails) and display (correct)

**Additional issue:**
- Shortcut column too narrow (7 chars) for formatted shortcuts like `"shift + ]"`

---

## Root Cause

Bubble Tea's `msg.String()` behavior:
- Special keys: `shift+tab` → `"shift+tab"` (modifier prefix works)
- Regular keys: `Shift+[` → `{` (terminal sends the shifted character)
- Regular keys: `Shift+]` → `}` (terminal sends the shifted character)

**Result:** Handlers listening for `"shift+]"` never match because Bubble Tea sends `"}"`.

---

## Solution: Separate Shortcut and ShortcutLabel

Add `ShortcutLabel` field to MenuItem:
- `Shortcut` = actual key binding (what Bubble Tea receives): `"}"`, `"{"`
- `ShortcutLabel` = display label (what user sees): `"shift + ]"`, `"shift + ["`

If `ShortcutLabel` is empty, display `Shortcut` (backward compatible).

---

## Phase 1: Update MenuItem Struct

**File:** `internal/app/menu.go`

### 1.1 Add ShortcutLabel field to MenuItem struct:

**Current (lines 13-21):**
```go
type MenuItem struct {
    ID        string // Unique identifier for the action
    Shortcut  string // Keyboard shortcut (single letter from label)
    Emoji     string // Leading emoji
    Label     string // Action name
    Hint      string // Plain language hint shown on focus
    Enabled   bool   // Whether this item can be selected
    Separator bool   // If true, this is a visual separator (non-selectable)
}
```

**Fixed:**
```go
type MenuItem struct {
    ID            string // Unique identifier for the action
    Shortcut      string // Keyboard shortcut (actual key binding, e.g., "}", "{")
    ShortcutLabel string // Display label for shortcut (e.g., "shift + ]"), empty = use Shortcut
    Emoji         string // Leading emoji
    Label         string // Action name
    Hint          string // Plain language hint shown on focus
    Enabled       bool   // Whether this item can be selected
    Separator     bool   // If true, this is a visual separator (non-selectable)
}
```

---

## Phase 2: Update SSOT Menu Items

**File:** `internal/app/menu_items.go`

### 2.1 Update force_push item (lines 62-69):

**Current:**
```go
"force_push": {
    ID:       "force_push",
    Shortcut: "shift+]",
    Emoji:    "...",
    ...
},
```

**Fixed:**
```go
"force_push": {
    ID:            "force_push",
    Shortcut:      "}",           // Actual key (Shift+] produces })
    ShortcutLabel: "shift + ]",   // Display label
    Emoji:         "...",
    ...
},
```

### 2.2 Update dirty_pull_merge item (lines 72-79):

**Current:**
```go
"dirty_pull_merge": {
    ID:       "dirty_pull_merge",
    Shortcut: "shift+[",
    Emoji:    "...",
    ...
},
```

**Fixed:**
```go
"dirty_pull_merge": {
    ID:            "dirty_pull_merge",
    Shortcut:      "{",           // Actual key (Shift+[ produces {)
    ShortcutLabel: "shift + [",   // Display label
    Emoji:         "...",
    ...
},
```

---

## Phase 3: Update Display Logic

**File:** `internal/app/app.go`

### 3.1 Update menuItemsToMaps to use ShortcutLabel for display:

Find `menuItemsToMaps` function and update the Shortcut field mapping:

**Current pattern (somewhere in menuItemsToMaps):**
```go
"Shortcut": item.Shortcut,
```

**Fixed:**
```go
// Use ShortcutLabel for display if set, otherwise use Shortcut
shortcutDisplay := item.Shortcut
if item.ShortcutLabel != "" {
    shortcutDisplay = item.ShortcutLabel
}
...
"Shortcut": shortcutDisplay,
```

---

## Phase 4: Widen Shortcut Column

**File:** `internal/ui/menu.go`

### 4.1 Update keyColWidth from 7 to 10:

**Current (line 33):**
```go
keyColWidth := 7
```

**Fixed:**
```go
keyColWidth := 10
```

**Rationale:** `"shift + ]"` is 9 characters + 1 space separator to label = 10 chars total.

---

## Phase 5: Format Existing Shortcuts (Optional Enhancement)

For visual consistency, format all shortcuts with spaces:

| Current | Formatted |
|---------|-----------|
| `ctrl+r` | `ctrl + r` |
| `[` | `[` |
| `]` | `]` |

**Note:** Single-character shortcuts stay as-is. Only multi-char shortcuts get spaces.

This is **optional** - can be done as follow-up sprint if desired.

---

## Files Modified Summary

| File | Change |
|------|--------|
| `internal/app/menu.go` | Add `ShortcutLabel` field to MenuItem struct |
| `internal/app/menu_items.go` | Update `force_push` and `dirty_pull_merge` with actual key + display label |
| `internal/app/app.go` | Update `menuItemsToMaps` to use `ShortcutLabel` if set |
| `internal/ui/menu.go` | Change `keyColWidth` from 7 to 10 |

---

## Acceptance Criteria

1. **Build passes**: `./build.sh` succeeds with no errors
2. **Shift+] works**: Pressing Shift+] triggers force_push action (shows confirmation)
3. **Shift+[ works**: Pressing Shift+[ triggers dirty_pull_merge action (shows confirmation)
4. **Display correct**: Menu shows `"shift + ]"` and `"shift + ["` (not `}` or `{`)
5. **Column width**: Shortcut column is 10 chars wide (no truncation)
6. **Backward compatible**: Other shortcuts (`[`, `]`, `ctrl+r`, etc.) still work
7. **No regression**: All existing menu shortcuts function correctly

---

## Testing Scenarios

### Scenario 1: Force Push Shortcut
1. Create local commit (ahead of remote)
2. Open TIT, see main menu
3. Press `Shift+]` (should trigger force_push confirmation)
4. Expected: Confirmation dialog appears for force push

### Scenario 2: Dirty Pull Shortcut
1. Create local changes (dirty tree) while behind remote
2. Open TIT, see main menu
3. Press `Shift+[` (should trigger dirty_pull_merge)
4. Expected: Dirty pull flow starts

### Scenario 3: Display Verification
1. Open TIT with ahead state
2. Navigate menu to show force_push item
3. Expected: Shortcut column shows `"shift + ]"` (not `"}"`)

---

## LIFESTAR Compliance

- [x] **L - Lean**: Minimal change (4 files, ~10 lines each)
- [x] **I - Immutable**: Deterministic display logic
- [x] **F - Findable**: All shortcuts in SSOT (menu_items.go)
- [x] **E - Explicit**: Separate fields for binding vs display
- [x] **S - SSOT**: Single definition per item
- [x] **T - Testable**: Clear test scenarios
- [x] **A - Accessible**: Human-readable shortcut labels
- [x] **R - Reviewable**: Consistent pattern across all items

---

**Ready for ENGINEER execution.**

Rock 'n Roll!
**JRENG!**
