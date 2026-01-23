# Session 84: Footer Unification (Status Bar → Footer)

**Role:** ANALYST
**Agent:** Amp (Claude Sonnet 4)
**Date:** 2026-01-23

## Problem Statement

With reactive layout, **footer and status bar are the same thing** — both sit at the bottom of the screen showing mode-specific hints. Current implementation has:

1. **Footer** — rendered by `RenderReactiveLayout()` as 1-line bottom section
2. **Status bar** — rendered INSIDE content by each mode (`buildHistoryStatusBar()`, `buildFileHistoryStatusBar()`, etc.)

**Issues:**
- Terminology confusion (two names for same concept)
- Duplicated rendering logic
- Status bar height subtracted from content, then footer adds another line
- Inconsistent naming across codebase (`statusBar`, `StatusBar`, `status_bar`)

## Solution

**Unify everything under FOOTER:**

1. Remove all "status bar" terminology
2. Footer content is mode-driven via `GetFooterContent()`
3. Full-screen modes (History, FileHistory, Console, ConflictResolver) return content WITHOUT embedded status bar
4. `RenderReactiveLayout()` handles footer rendering for ALL modes

---

## Current Architecture (Before)

```
┌─────────────────────────────────────┐
│ HEADER                              │
├─────────────────────────────────────┤
│ CONTENT                             │
│ ┌─────────────────────────────────┐ │
│ │ ... mode content ...            │ │
│ │ STATUS BAR (embedded)           │ │  ← Rendered inside content
│ └─────────────────────────────────┘ │
├─────────────────────────────────────┤
│ FOOTER                              │  ← Separate from status bar
└─────────────────────────────────────┘
```

## Target Architecture (After)

```
┌─────────────────────────────────────┐
│ HEADER                              │
├─────────────────────────────────────┤
│ CONTENT                             │
│ ┌─────────────────────────────────┐ │
│ │ ... mode content ...            │ │  ← No embedded footer
│ │                                 │ │
│ └─────────────────────────────────┘ │
├─────────────────────────────────────┤
│ FOOTER                              │  ← Single source, mode-driven
└─────────────────────────────────────┘
```

---

## Implementation Details

### 1. New Footer Content System

**File:** `internal/app/footer.go` (new)

```go
// GetFooterContent returns the rendered footer for current mode.
// Priority: quitConfirm > clearConfirm > mode-specific hints
// Returns styled string ready for display.
func (a *Application) GetFooterContent() string {
    width := a.sizing.TerminalWidth
    
    // Priority 1: Quit confirmation (Ctrl+C)
    if a.quitConfirmActive {
        return ui.RenderFooterOverride(GetFooterMessageText(MessageCtrlCConfirm), width, &a.theme)
    }
    
    // Priority 2: Clear confirmation (ESC in input)
    if a.clearConfirmActive {
        return ui.RenderFooterOverride(GetFooterMessageText(MessageEscClearConfirm), width, &a.theme)
    }
    
    // Priority 3: Mode-specific hints from SSOT
    hintKey := a.getFooterHintKey()
    
    // Special case: Menu uses MenuItem.Hint (plain text, not shortcuts)
    if hintKey == "menu" {
        if len(a.menuItems) > 0 && a.selectedIndex < len(a.menuItems) {
            return ui.RenderFooterOverride(a.menuItems[a.selectedIndex].Hint, width, &a.theme)
        }
        return ""
    }
    
    // Lookup shortcuts from SSOT and render with styling
    shortcuts := ui.FooterHints[hintKey]
    return ui.RenderFooter(shortcuts, width, &a.theme)
}

// getFooterHintKey returns the SSOT key for current mode/state
func (a *Application) getFooterHintKey() string {
    switch a.mode {
    case ModeMenu:
        return "menu" // Special case: uses MenuItem.Hint
        
    case ModeHistory:
        if a.historyState.PaneFocused {
            return "history_list"
        }
        return "history_details"
        
    case ModeFileHistory:
        return a.getFileHistoryHintKey() // returns filehistory_commits|files|diff|visual
        
    case ModeConsole, ModeClone:
        if a.asyncOperationActive {
            return "console_running"
        }
        return "console_complete"
        
    case ModeConflictResolver:
        return a.getConflictHintKey() // returns conflict_list|diff
        
    case ModeInput:
        if a.inputHeight > 1 {
            return "input_multi"
        }
        return "input_single"
        
    case ModeConfirmation:
        return "confirmation"
        
    default:
        return ""
    }
}
```

### 2. Footer Hints SSOT (Structured Data)

**File:** `internal/ui/footer.go` (renamed from statusbar.go)

Footer hints are **not plain text** — they have styling:
- **Key**: Bold, accent color (`theme.AccentTextColor`)
- **Description**: Plain content color (`theme.ContentTextColor`)
- **Separator**: Dimmed (`theme.DimmedTextColor`)

**SSOT Structure:**

```go
// FooterShortcut represents a single keyboard shortcut hint
type FooterShortcut struct {
    Key  string // e.g., "↑↓", "Enter", "Esc"
    Desc string // e.g., "navigate", "select", "back"
}

// FooterHints defines all mode-specific footer shortcuts (SSOT)
// Key = mode identifier, Value = list of shortcuts
var FooterHints = map[string][]FooterShortcut{
    // History mode
    "history_list": {
        {Key: "↑↓", Desc: "navigate"},
        {Key: "Enter", Desc: "time travel"},
        {Key: "Tab", Desc: "details"},
        {Key: "Esc", Desc: "back"},
    },
    "history_details": {
        {Key: "↑↓", Desc: "scroll"},
        {Key: "Tab", Desc: "list"},
        {Key: "Esc", Desc: "back"},
    },
    
    // File History mode
    "filehistory_commits": {
        {Key: "↑↓", Desc: "navigate"},
        {Key: "Tab", Desc: "files"},
        {Key: "Esc", Desc: "back"},
    },
    "filehistory_files": {
        {Key: "↑↓", Desc: "navigate"},
        {Key: "Tab", Desc: "diff"},
        {Key: "Esc", Desc: "back"},
    },
    "filehistory_diff": {
        {Key: "↑↓", Desc: "scroll"},
        {Key: "v", Desc: "visual"},
        {Key: "Tab", Desc: "commits"},
        {Key: "Esc", Desc: "back"},
    },
    "filehistory_visual": {
        {Key: "↑↓", Desc: "extend"},
        {Key: "y", Desc: "yank"},
        {Key: "Esc", Desc: "cancel"},
    },
    
    // Conflict Resolver
    "conflict_list": {
        {Key: "↑↓", Desc: "navigate"},
        {Key: "Space", Desc: "toggle"},
        {Key: "Tab", Desc: "diff"},
        {Key: "Enter", Desc: "resolve"},
    },
    "conflict_diff": {
        {Key: "↑↓", Desc: "scroll"},
        {Key: "Tab", Desc: "list"},
        {Key: "Esc", Desc: "back"},
    },
    
    // Console
    "console_running": {
        {Key: "Esc", Desc: "abort"},
    },
    "console_complete": {
        {Key: "Esc", Desc: "back"},
    },
    
    // Input
    "input_single": {
        {Key: "Enter", Desc: "submit"},
        {Key: "Esc", Desc: "cancel"},
    },
    "input_multi": {
        {Key: "Ctrl+Enter", Desc: "submit"},
        {Key: "Esc", Desc: "cancel"},
    },
    
    // Confirmation
    "confirmation": {
        {Key: "←→", Desc: "select"},
        {Key: "Enter", Desc: "confirm"},
        {Key: "Esc", Desc: "cancel"},
    },
    
    // Menu (hint comes from MenuItem.Hint, not here)
    // Config, Preferences, BranchPicker will be added in Session 86
}
```

### 3. Footer Rendering (Styled)

**File:** `internal/ui/footer.go`

```go
// FooterStyles provides style objects for footer rendering
type FooterStyles struct {
    KeyStyle  lipgloss.Style // Bold, accent color
    DescStyle lipgloss.Style // Plain, content color
    SepStyle  lipgloss.Style // Dimmed separator
}

// NewFooterStyles creates footer styles from theme
func NewFooterStyles(theme *Theme) FooterStyles {
    return FooterStyles{
        KeyStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.AccentTextColor)).
            Bold(true),
        DescStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.ContentTextColor)),
        SepStyle: lipgloss.NewStyle().
            Foreground(lipgloss.Color(theme.DimmedTextColor)),
    }
}

// RenderFooter renders footer shortcuts with proper styling
func RenderFooter(shortcuts []FooterShortcut, width int, theme *Theme) string {
    if theme == nil || len(shortcuts) == 0 {
        return ""
    }
    
    styles := NewFooterStyles(theme)
    
    // Build styled parts: "Key desc"
    var parts []string
    for _, sc := range shortcuts {
        part := styles.KeyStyle.Render(sc.Key) + styles.DescStyle.Render(" "+sc.Desc)
        parts = append(parts, part)
    }
    
    // Join with styled separator
    sep := styles.SepStyle.Render("  ·  ")
    content := strings.Join(parts, sep)
    
    // Center in terminal width
    return lipgloss.NewStyle().
        Width(width).
        Align(lipgloss.Center).
        Render(content)
}

// RenderFooterOverride renders override message (e.g., Ctrl+C confirm)
func RenderFooterOverride(message string, width int, theme *Theme) string {
    style := lipgloss.NewStyle().
        Foreground(lipgloss.Color(theme.FooterTextColor)).
        Width(width).
        Align(lipgloss.Center)
    return style.Render(message)
}
```

### 4. Remove Embedded Status Bars

**Files to modify:**

| File | Remove | Keep |
|------|--------|------|
| `ui/statusbar.go` | Rename to `ui/footer.go` | `BuildFooter()`, `FooterConfig`, `FooterStyles` |
| `ui/history.go` | `buildHistoryStatusBar()` | `RenderHistorySplitPane()` without status bar |
| `ui/filehistory.go` | `buildFileHistoryStatusBar()`, `buildDiffStatusBar()` | Content rendering only |
| `ui/console.go` | `buildConsoleStatusBar()` | Content rendering only |
| `ui/conflictresolver.go` | `buildGenericConflictStatusBar()` | Content rendering only |

### 5. Update Render Functions

**Before (history.go):**
```go
func RenderHistorySplitPane(..., statusBarOverride string) string {
    // Reserve 1 line for status bar
    paneHeight := height - 2
    
    // ... render content ...
    
    statusBar := buildHistoryStatusBar(...)
    return mainRow + "\n" + statusBar
}
```

**After (history.go):**
```go
func RenderHistorySplitPane(...) string {
    // Full height for content (footer handled externally)
    paneHeight := height
    
    // ... render content ...
    
    return mainRow  // No status bar
}
```

### 6. Update View() in app.go

**Before:**
```go
func (a *Application) View() string {
    switch a.mode {
    case ModeHistory:
        statusOverride := ""
        if a.quitConfirmActive {
            statusOverride = GetFooterMessageText(MessageCtrlCConfirm)
        }
        contentText = ui.RenderHistorySplitPane(..., statusOverride)
    }
    
    return ui.RenderReactiveLayout(sizing, theme, header, contentText, footer)
}
```

**After:**
```go
func (a *Application) View() string {
    switch a.mode {
    case ModeHistory:
        contentText = ui.RenderHistorySplitPane(...)  // No statusOverride param
    }
    
    footer := a.GetFooterContent()  // Unified footer
    return ui.RenderReactiveLayout(sizing, theme, header, contentText, footer)
}
```

---

## Naming Convention (MANDATORY)

**All references must use "footer" terminology:**

| Old Name | New Name |
|----------|----------|
| `statusBar` | `footer` |
| `StatusBar` | `Footer` |
| `statusBarOverride` | (removed — use `GetFooterContent()`) |
| `buildXxxStatusBar()` | (removed — footer is centralized) |
| `StatusBarConfig` | `FooterConfig` |
| `StatusBarStyles` | `FooterStyles` |
| `NewStatusBarStyles()` | `NewFooterStyles()` |
| `BuildStatusBar()` | `BuildFooter()` |

---

## Files to Create

| File | Purpose |
|------|---------|
| `internal/app/footer.go` | `GetFooterContent()`, mode-specific footer logic |

## Files to Rename

| Old | New |
|-----|-----|
| `internal/ui/statusbar.go` | `internal/ui/footer.go` |

## Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/footer.go` | Rename types/functions: StatusBar → Footer |
| `internal/ui/history.go` | Remove `buildHistoryStatusBar()`, update `RenderHistorySplitPane()` signature |
| `internal/ui/filehistory.go` | Remove `buildFileHistoryStatusBar()`, `buildDiffStatusBar()`, update render |
| `internal/ui/console.go` | Remove `buildConsoleStatusBar()`, update `RenderConsoleOutput()` |
| `internal/ui/conflictresolver.go` | Remove `buildGenericConflictStatusBar()`, update render |
| `internal/app/app.go` | Update `View()` to use `GetFooterContent()`, remove statusOverride params |
| `internal/app/messages.go` | Add `FooterHints` SSOT map |

---

## Implementation Phases

### Phase 1: Rename statusbar.go → footer.go
- Rename file
- Rename all types/functions (StatusBar → Footer)
- Update imports in all files

### Phase 2: Create footer.go (app package)
- Implement `GetFooterContent()`
- Implement `getFooterForMode()` with all mode cases
- Add `FooterHints` SSOT to messages.go

### Phase 3: Update History Mode
- Remove `buildHistoryStatusBar()`
- Update `RenderHistorySplitPane()` to not render footer
- Update call sites in app.go

### Phase 4: Update FileHistory Mode
- Remove `buildFileHistoryStatusBar()`, `buildDiffStatusBar()`
- Update `RenderFileHistorySplitPane()`
- Update call sites

### Phase 5: Update Console Mode
- Remove `buildConsoleStatusBar()`
- Update `RenderConsoleOutput()`
- Update call sites

### Phase 6: Update ConflictResolver Mode
- Remove `buildGenericConflictStatusBar()`
- Update render function
- Update call sites

### Phase 7: Unify View()
- Update `View()` to call `GetFooterContent()` once
- Pass footer to `RenderReactiveLayout()`
- Remove all statusOverride parameters

---

## Success Criteria

1. ✅ No "statusBar" or "StatusBar" references in codebase
2. ✅ All modes use unified `GetFooterContent()`
3. ✅ Footer content is mode-driven (SSOT in messages.go)
4. ✅ Ctrl+C confirmation works in all modes
5. ✅ ESC clear confirmation works in input modes
6. ✅ Footer renders correctly in all modes
7. ✅ Clean build with `./build.sh`

---

## Dependencies

- **Prerequisite for:** Session 85 (Timeline Sync), Session 86 (Config Menu)
- **Depends on:** Reactive layout (Session 80) — already implemented

---

**End of Kickoff Plan**

Ready for SCAFFOLDER to implement Phase 1.
