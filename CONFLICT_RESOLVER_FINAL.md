# Conflict Resolver - Final Color & Border Fix

## Issue: Can't Differentiate Active Pane

### Root Cause
Using old-tit's `BorderPrimaryColor` (#2C4144 - dark) and `BorderSecondaryColor` (#8CC9D9 - bright) was too subtle.

### Solution: Revert to Previous Border Strategy
**Unfocused:** `BoxBorderColor` = #8CC9D9 (dolphin - bright baseline)  
**Focused:** `AccentTextColor` = #01C2D2 (caribbeanBlue - brighter accent)

This creates **clear visual contrast** - active pane "pops" with bright cyan.

## Issue: Selection Colors Don't Match Menu

### Root Cause
Using conflict-specific colors instead of menu convention.

### Solution: Use Menu Convention Throughout

**List Pane Selection (Active):**
- Foreground: `MainBackgroundColor` (#090D12 - dark)
- Background: `MenuSelectionBackground` (#7EB8C5 - teal)
- **Matches menu exactly**

**Diff Pane Cursor (Active):**
- Foreground: `MainBackgroundColor` (#090D12 - dark)
- Background: `MenuSelectionBackground` (#7EB8C5 - teal)
- **Same as menu selection**

**Content Pane Cursor (Active):**
- Foreground: `MainBackgroundColor` (#090D12 - dark)
- Background: `MenuSelectionBackground` (#7EB8C5 - teal)
- **Consistent across all components**

## Color Summary

### Borders (Clear Differentiation)
```
Unfocused: #8CC9D9 (dolphin - bright teal)
Focused:   #01C2D2 (caribbeanBlue - brighter cyan)
```

### Selection (Menu Convention)
```
Foreground: #090D12 (bunker - dark)
Background: #7EB8C5 (teal - menu selection)
```

### Pane Titles
```
Color: #8CC9D9 (dolphin)
```

## Changes Applied

**Updated Files:**
- `listpane.go` - Border: BoxBorderColor/AccentTextColor, Selection: menu convention
- `diffpane.go` - Border: BoxBorderColor/AccentTextColor, Cursor: menu convention
- `conflictresolver.go` - Border: BoxBorderColor/AccentTextColor, Cursor: menu convention

**Result:**
- ✅ Active pane clearly visible (bright cyan border)
- ✅ Selection matches menu (dark on teal)
- ✅ Consistent visual language across app

## Testing

Launch `./tit_x64` → press `t` → verify:
1. **Unfocused panes:** Bright teal borders (#8CC9D9)
2. **Focused pane:** Brighter cyan border (#01C2D2) - **clearly different**
3. **Selected item:** Dark text on teal background (matches menu)
4. **TAB navigation:** Border color change is obvious

