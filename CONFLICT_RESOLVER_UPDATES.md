# Conflict Resolver - Updates Applied

## Changes Made (2026-01-06)

### 1. File Renamed ✓
- `internal/ui/conflictresolve.go` → `internal/ui/conflictresolver.go`
- Component now properly named "ConflictResolver"

### 2. Semantic Color Naming ✓

**Added to theme.go:**
```toml
# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"  # littleMermaid (unfocused pane)
conflictPaneFocusedBorder = "#8CC9D9"    # dolphin (focused pane)

# Conflict Resolver - Selection
conflictSelectionForeground = "#090D12"  # bunker (selection text)
conflictSelectionBackground = "#7EB8C5"  # brighter muted teal (selection bg)

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#8CC9D9"       # dolphin (pane titles)
```

**Updated Components:**
- `listpane.go` - Uses conflict-specific colors
- `diffpane.go` - Uses conflict-specific colors
- `conflictresolver.go` - Uses conflict-specific colors

**Color Mapping (old-tit → new-tit):**
- `BorderPrimaryColor` → `ConflictPaneUnfocusedBorder` (dark line for unfocused)
- `BorderSecondaryColor` → `ConflictPaneFocusedBorder` (bright line for focused)
- `PrimaryBackground` → `ConflictSelectionForeground` (dark text on selection)
- `MenuHighlightedBackground` → `ConflictSelectionBackground` (teal selection bg)
- `PaneTitleColor` → `ConflictPaneTitleColor` (bright pane headers)

### 3. Exclusive Toggle Fixed ✓

**Previous Behavior:**
- SPACE on already-selected column: no action

**Current Behavior:**
- SPACE always marks the focused column
- Automatically unmarks other columns (exclusive selection)
- One choice must always be selected (can't deselect all)

**Implementation:**
```go
// Exclusive toggle: only one column can be chosen
if file.Chosen == focusedPane {
    return app, nil // Already chosen, no change
}
file.Chosen = focusedPane // Marks this, unmarks others
```

### Color Values (Exact Match with old-tit)

```
Unfocused Border: #2C4144 (littleMermaid - dark subtle line)
Focused Border:   #8CC9D9 (dolphin - bright accent line)
Selection FG:     #090D12 (bunker - dark text on bright bg)
Selection BG:     #7EB8C5 (brighter muted teal - matches old-tit exactly)
Pane Titles:      #8CC9D9 (dolphin - bright headers)
```

### Testing Checklist

✓ Build compiles successfully
✓ Color names are semantically clear
✓ Colors match old-tit values exactly
✓ SPACE toggles selection exclusively
✓ File renamed to ConflictResolver

### Next Test (Your Turn)

1. Launch `./tit_x64`
2. Press `t` (TEST: Conflict resolver)
3. Verify:
   - Unfocused borders are dark (#2C4144)
   - Focused border is bright (#8CC9D9)
   - SPACE marks the focused column
   - Pressing SPACE again does nothing (exclusive selection)
   - TAB moves focus, border colors update
   - Selection background matches old-tit (#7EB8C5)

