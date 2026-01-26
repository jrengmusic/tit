# Sprint 11: Confirmation Dialog Background Fix - Kickoff Plan

**Role:** COUNSELOR
**Agent:** Copilot (claude-opus-4.5)
**Date:** 2026-01-26
**Objective:** Fix confirmation dialog text background not filling content area

---

## Problem Statement

The confirmation dialog's text content (title and explanation) does not fill the entire dialog area with the background color. The terminal background shows through where text doesn't cover.

**Screenshot shows:**
- "Pull from remote?" title - background stops at text end
- Explanation text - background stops at text end
- Right side of dialog shows terminal gradient instead of solid `#112130`

---

## Root Cause

**Lipgloss does NOT inherit background colors from parent containers.**

Current flow:
1. `explanationStyle` renders text with **no background** (line 158-161)
2. `titleStyle` renders text with **no background** (line 202)
3. `dialogStyle` wraps content with background (line 150-156)

The `dialogStyle` background only fills **padding area**, not the actual text content. Each text style needs its own explicit `Background()`.

---

## Fix

**File:** `internal/ui/confirmation.go`

### Change 1: Add background to explanationStyle (lines 158-161)

**Current:**
```go
explanationStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color(c.Theme.ContentTextColor)).
    Width(dialogWidth - 4). // Account for dialog padding
    Align(lipgloss.Left)
```

**Fixed:**
```go
explanationStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color(c.Theme.ContentTextColor)).
    Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground)).
    Width(dialogWidth - 4). // Account for dialog padding
    Align(lipgloss.Left)
```

### Change 2: Add background to titleStyle (line 202)

**Current:**
```go
titleStyle := lipgloss.NewStyle().Bold(true)
```

**Fixed:**
```go
titleStyle := lipgloss.NewStyle().
    Bold(true).
    Width(dialogWidth - 4).
    Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground))
```

**Note:** Added `Width(dialogWidth - 4)` to titleStyle so background extends full width, matching explanationStyle.

---

## Files Modified Summary

| File | Change |
|------|--------|
| `internal/ui/confirmation.go` | Add Background() to explanationStyle (line 159) and titleStyle (line 202) |

---

## Acceptance Criteria

1. **Build passes**: `./build.sh` succeeds with no errors
2. **Dialog background solid**: Entire dialog content area fills with `#112130` (ConfirmationDialogBackground)
3. **No terminal bleed-through**: No gradient/terminal background visible inside dialog
4. **Title full width**: Title background extends to dialog padding edge
5. **Explanation full width**: Explanation background extends to dialog padding edge
6. **Buttons unchanged**: Button styling remains correct (already have backgrounds)

---

## LIFESTAR Compliance

- [x] **L - Lean**: Minimal fix (2 lines changed)
- [x] **E - Explicit**: Background color explicitly set, not relying on inheritance
- [x] **S - SSOT**: Uses existing `c.Theme.ConfirmationDialogBackground` from theme
- [x] **R - Reviewable**: Clear, obvious fix

---

**Ready for ENGINEER execution.**

Rock 'n Roll!
**JRENG!**
