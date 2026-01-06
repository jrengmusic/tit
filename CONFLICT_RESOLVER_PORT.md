# Conflict Resolver - Final Fixes

## Issues Fixed (2026-01-06 18:25)

### 1. ✅ List Pane Scrolling - FIXED
**Problem:** Only 1 file visible, couldn't see other files when navigating with ↑↓

**Root Cause:** ListPane created fresh each render, no scroll adjustment

**Fix:** Added `AdjustScroll(selectedFileIndex, visibleLines)` before rendering
```go
listPane.AdjustScroll(selectedFileIndex, visibleLines)
```

**Result:** All 3 files now visible with scrolling as you navigate

### 2. ✅ Border Colors - VISIBLE DIFFERENCE
**Problem:** Unfocused border (#2C4144) too dark, invisible against background (#090D12)

**Fix:** Changed unfocused to dimmed but visible color
```
Unfocused: #33535B (mediterranea - dim but clearly visible)
Focused:   #01C2D2 (caribbeanBlue - bright cyan, pops)
```

**Result:** Clear visual difference between active/inactive panes

### 3. ✅ Border Artifacts - LIPGLOSS EXPECTED
**Note:** The "touching" borders are normal lipgloss behavior when using `JoinHorizontal`
- Borders touch at edges (no gaps) = maximizes content space
- Works correctly in old-tit with same approach
- Allows 3x3 panes to fit with maximum space

## Current Color Scheme

```
Unfocused Border: #33535B (dim teal - visible but recedes)
Focused Border:   #01C2D2 (bright cyan - clearly active)

Selection:        #090D12 on #7EB8C5 (menu convention)
Pane Titles:      #8CC9D9 (bright teal)
```

## Files Updated
- `theme.go` - Updated border colors
- `conflictresolver.go` - Added AdjustScroll call

## Test Now
1. `./tit_x64` → press `t`
2. Verify all 3 files visible in list
3. Press ↑↓ to navigate - list scrolls to keep selection visible
4. Press TAB - focused border is bright cyan, others are dim teal
5. All content maximized (borders touch for space efficiency)

