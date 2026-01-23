# Session 86 CARETAKER Analysis

**Role:** CARETAKER
**Agent:** OpenCode (CLI Agent)
**Date:** 2026-01-23
**Task:** Analysis of SCAFFOLDER work and identification of critical missing pieces

## Summary

SCAFFOLDER has completed literal scaffolding of Sessions 85 and 86 as specified in the kickoff plans. The code compiles cleanly but several critical integration points are missing that prevent the features from working.

## Session 85: Timeline Sync Status

### ✅ COMPLETED
1. Core sync infrastructure (`timeline_sync.go`)
2. Message types (`TimelineSyncMsg`, `TimelineSyncTickMsg`)
3. Application state fields (`timelineSyncInProgress`, `timelineSyncLastUpdate`, `timelineSyncFrame`)
4. Update handlers for sync messages
5. Header visual feedback (spinner during sync)
6. Constants defined (`TimelineSyncInterval`, `TimelineSyncTickRate`)

### ❌ MISSING - CRITICAL
1. **Config integration** - Timeline sync doesn't respect `appConfig.AutoUpdate.Enabled`
2. **Interval configuration** - Uses hardcoded 60s instead of `appConfig.AutoUpdate.IntervalMinutes`

## Session 86: Config Menu Status

### ✅ COMPLETED
1. Config infrastructure (`config.go` with TOML support)
2. Config menu generation (`GenerateConfigMenu()`)
3. "/" shortcut handler to open config menu
4. Menu item definitions for config operations
5. UI components created (`preferences.go`, `branchpicker.go`)
6. Git branch operations (`branch.go`)
7. State structs for preferences and branch picker

### ❌ MISSING - CRITICAL

#### View Method Cases
1. **ModeBranchPicker** - No case in View() method
2. **ModePreferences** - No case in View() method
Result: These modes will panic with "Unknown app mode"

#### Handler Wiring
1. Config menu handlers not fully wired:
   - `config_toggle_auto_update` - handler exists but doesn't toggle
   - `config_switch_branch` - should transition to ModeBranchPicker
   - `config_preferences` - should transition to ModePreferences

#### Integration Points
1. Auto-update toggle doesn't update config
2. Theme switcher not connected to actual theme loading
3. Branch picker navigation (up/down) not implemented
4. Preferences row navigation not implemented
5. Footer hints missing for new modes

## Critical Integration Points to Fix

### Priority 1: Fix Panic Conditions
```go
// In app.go View() method, add cases:
case ModeBranchPicker:
    contentText = ui.RenderBranchPickerSplitPane(a.sizing, a.branchPickerState, a.theme)

case ModePreferences:
    contentText = ui.RenderPreferencesPane(a.sizing, a.preferencesState, a.appConfig, a.theme)
```

### Priority 2: Wire Config to Timeline Sync
```go
// In timeline_sync.go shouldRunTimelineSync():
// Check if auto-update is enabled
if a.appConfig != nil && !a.appConfig.AutoUpdate.Enabled {
    return false
}

// Use configured interval
interval := time.Duration(a.appConfig.AutoUpdate.IntervalMinutes) * time.Minute
```

### Priority 3: Wire Navigation Handlers
- Connect config menu items to proper mode transitions
- Implement up/down navigation for preferences and branch picker
- Add space/+/- handlers for preferences editing

### Priority 4: Add Footer Hints
- ModeConfig: "Enter select · / back to main · Ctrl+C quit"
- ModeBranchPicker: "Enter switch · Esc back · Ctrl+C quit"
- ModePreferences: "Space toggle · +/- adjust · Esc save · Ctrl+C quit"

## Files Requiring CARETAKER Work

### Must Modify
1. `internal/app/app.go` - Add View cases for new modes
2. `internal/app/timeline_sync.go` - Respect config settings
3. `internal/app/handlers.go` - Wire config menu actions
4. `internal/app/footer.go` - Add footer hints for new modes

### Should Review
1. `internal/ui/preferences.go` - Ensure proper state handling
2. `internal/ui/branchpicker.go` - Verify branch list rendering
3. `internal/git/branch.go` - Test git operations work correctly

## Risk Assessment

**HIGH RISK**: Application will panic if user navigates to ModeBranchPicker or ModePreferences
**MEDIUM RISK**: Auto-update config is ignored, timeline sync runs regardless
**LOW RISK**: Missing footer hints and navigation polish

## Recommendation

CARETAKER should focus on:
1. Preventing panics (add View cases)
2. Wiring config to timeline sync
3. Basic navigation working
4. Testing all flows end-to-end

Leave for future session:
- Theme hot-reload
- Advanced branch switching (dirty handling)
- Preference validation edge cases