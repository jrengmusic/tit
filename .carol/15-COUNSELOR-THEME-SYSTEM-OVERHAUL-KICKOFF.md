# Sprint 15: Theme System Architectural Overhaul - COUNSELOR Kickoff Plan

**Role:** COUNSELOR
**Agent:** Copilot (claude-sonnet-4)
**Date:** 2026-01-27
**Objective:** Replace broken HSL transformation with explicit named color system across 5 themes
**Updated for:** Post-Sprint 14 theme system replacement

---

## Problem Statement

**Current System Issues:**
1. **Broken HSL Transformation**: Lines 170-238 in `theme.go` attempt mathematical color transformations that produce poor results
2. **Naming Convention Violations**: `DefaultThemeTOML` violates **Rule 2** (construct expression without data type) - should be semantic name
3. **Mixed Architecture**: Some themes use named colors, others use direct hex values, creating inconsistency
4. **Poor Seasonal Results**: Mathematical transformation destroys color relationships and semantic meaning

**Target Architecture:**
- **5 Explicit Themes**: GFX (SSOT) + 4 seasonal themes with identical structure
- **Named Colors**: Each theme has 18 curated colors with semantic names from palette files
- **Semantic Hooks**: 42 hooks reference named colors (not hex values)
- **Zero Mathematics**: Complete removal of HSL transformation logic

---

## Scope Analysis

**Massive Architectural Change:**
- **Files Modified**: 1 major file (`theme.go`) + 5 theme files
- **Code Removal**: ~240 lines of HSL transformation logic
- **New Structure**: 5 × (18 named colors + 42 semantic mappings) = 210 explicit assignments
- **Semantic Mappings**: 4 seasonal themes × 42 hooks = 168 expert color assignments

**Risk Level**: HIGH - Core visual system replacement affects entire UI

---

## Architecture Design

### New Theme File Structure

**All 5 Themes Use Identical Structure:**
```toml
name = "Theme Name"
description = "Theme description"

[colors]
# 18 named colors (unique per theme, from palette files)
colorName = "HEXVALUE"
# ... (18 total)

[theme]
# 42 semantic hooks reference color names (not hex)
mainBackgroundColor = "colorName"
# ... (42 total)
```

### Theme Specifications

**1. GFX Theme (SSOT Reference)**
- Renamed from `DefaultThemeTOML` to `GfxTheme` (naming convention compliance)
- Can have >18 colors (it's the reference implementation)
- Existing color scheme preserved, restructured to [colors]/[theme] sections

**2. Spring Theme**
- 18 colors: wildWatermelon → daisyBush (natural progression)
- Mapping: Green colors for positive states, red for negative, natural hierarchy

**3. Summer Theme**
- 18 colors: cyan → astra (electric progression) 
- Mapping: Electric cyan/bright colors for positives, hot reds for negatives, full spectrum energy

**4. Autumn Theme**
- 18 colors: hibiscus → jacaranda (rich progression)
- Mapping: Golden colors for positives, deep reds for negatives, warm earth tones

**5. Winter Theme**
- 18 colors: cloudBurst → melanie (sophisticated progression)
- Mapping: Professional blues for positives, soft pinks for negatives, cool elegance

---

## Implementation Strategy

### Phase 1: Remove Broken HSL System
**File:** `internal/ui/theme.go`

**Remove Components:**
- `HSLColor` struct (lines 14-17)
- `hslToHex` function (lines 20-75)  
- `hexToHSL` function (lines 77-120)
- `SeasonalTheme` struct (lines 140-170)
- `GetSeasonalThemes` function (lines 172-190)
- `generateSeasonalTheme` function (lines 192-238)
- `EnsureFiveThemesExist` HSL calls (lines 254-261)

**Result:** ~240 lines of mathematical transformation removed

### Phase 2: Restructure GFX Theme
**File:** `internal/ui/theme.go`

**Rename and Restructure:**
```go
// OLD: DefaultThemeTOML (violates naming convention)
// NEW: GfxTheme (semantic, compliant)
const GfxTheme = `name = "GFX (Reference)"
description = "TIT color scheme"

[colors]
# Named colors from existing GFX palette
bunker = "#090D12"
dark = "#1B2A31"
# ... (all existing colors with semantic names)

[theme]
# 42 semantic hooks reference color names
mainBackgroundColor = "bunker"
inlineBackgroundColor = "dark"
# ... (all 42 hooks)
`
```

### Phase 3: Create Explicit Seasonal Themes
**Files:** Create 4 new theme constant strings in `theme.go`

**Spring Theme Integration:**
```go
const SpringTheme = `name = "Spring"
description = "Fresh, natural progression"

[colors]
wildWatermelon = "FD5B68"
froly = "F67F78"
# ... (16 more from Spring.txt)

[theme]
mainBackgroundColor = "daisyBush"
contentTextColor = "easternBlue"
statusClean = "emerald"
statusDirty = "wildWatermelon"
# ... (38 more mappings)
`
```

**Similar integration for Summer, Autumn, Winter themes.**

### Phase 4: Update Theme Generation Logic
**File:** `internal/ui/theme.go`

**Replace `EnsureFiveThemesExist`:**
```go
func EnsureFiveThemesExist() error {
    configThemeDir := filepath.Join(getConfigDirectory(), "themes")
    if err := os.MkdirAll(configThemeDir, 0755); err != nil {
        return fmt.Errorf("failed to create themes directory: %w", err)
    }

    // Write 5 explicit themes
    themes := map[string]string{
        "gfx":    GfxTheme,
        "spring": SpringTheme,
        "summer": SummerTheme,
        "autumn": AutumnTheme,
        "winter": WinterTheme,
    }
    
    for name, content := range themes {
        themePath := filepath.Join(configThemeDir, name+".toml")
        if err := os.WriteFile(themePath, []byte(content), 0644); err != nil {
            return fmt.Errorf("failed to write %s theme: %w", name, err)
        }
    }

    return nil
}
```

### Phase 5: Update References
**File:** `internal/ui/theme.go`

**Update Function References:**
- `LoadDefaultTheme` → references "gfx.toml"
- `CreateDefaultThemeIfMissing` → calls updated `EnsureFiveThemesExist`
- Remove all references to `DefaultThemeTOML` → use `GfxTheme`

---

## Complete Color Mappings

### Semantic Hook Categories (42 Total)

**Backgrounds (3):**
- `mainBackgroundColor`, `inlineBackgroundColor`, `selectionBackgroundColor`

**Text Content (6):**
- `contentTextColor`, `labelTextColor`, `dimmedTextColor`, `accentTextColor`, `highlightTextColor`, `terminalTextColor`

**Special Text (2):**
- `cwdTextColor`, `footerTextColor`

**Borders (2):**
- `boxBorderColor`, `separatorColor`

**Confirmation (1):**
- `confirmationDialogBackground`

**Conflict Resolver (5):**
- `conflictPaneUnfocusedBorder`, `conflictPaneFocusedBorder`, `conflictSelectionForeground`, `conflictSelectionBackground`, `conflictPaneTitleColor`

**Status (2):**
- `statusClean`, `statusDirty`

**Timeline (3):**
- `timelineSynchronized`, `timelineLocalAhead`, `timelineLocalBehind`

**Operations (7):**
- `operationReady`, `operationNotRepo`, `operationTimeTravel`, `operationConflicted`, `operationMerging`, `operationRebasing`, `operationDirtyOp`

**UI Elements (2):**
- `menuSelectionBackground`, `buttonSelectedTextColor`

**Animation (1):**
- `spinnerColor`

**Diff (2):**
- `diffAddedLineColor`, `diffRemovedLineColor`

**Console Output (6):**
- `outputStdoutColor`, `outputStderrColor`, `outputStatusColor`, `outputWarningColor`, `outputDebugColor`, `outputInfoColor`

### Expert Color Assignments

**Spring Theme (Natural Fresh):**
- Positive: `emerald`, `feijoa`, `shamrock` (greens)
- Negative: `wildWatermelon`, `froly` (reds)
- Backgrounds: `daisyBush` → `sapphire` → `ceruleanBlue` (dark to light blues)
- Content: `easternBlue`, `lochmara` (readable mid-blues)

**Summer Theme (Electric Vibrant):**
- Positive: `cyan`, `brightTurquoise` (electric)
- Negative: `radicalRed`, `wildStrawberry` (hot)
- Backgrounds: `blueMarguerite` → `havelockBlue` → `violetBlue` (purple depth)
- Content: `pictonBlue`, `cyan` (bright readable)

**Autumn Theme (Rich Warm):**
- Positive: `corn`, `saffronMango` (golds)
- Negative: `guardsmanRed`, `grenadier` (deep reds)
- Backgrounds: `jacaranda` → `mulberryWood` → `roseBudCherry` (purple depth)
- Content: `california`, `tangerine` (warm oranges)

**Winter Theme (Professional Cool):**
- Positive: `havelockBlue`, `chetwodeBlue` (professional blues)
- Negative: `melanie`, `lily` (soft pinks - Winter approach)
- Backgrounds: `cloudBurst` → `sanMarino` → `sanJuan` (blue depth)
- Content: `chetwodeBlue`, `pigeonPost` (readable cool blues)

---

## Files Modified Summary

| File | Change Type | Impact |
|------|-------------|--------|
| `internal/ui/theme.go` | Major Refactor | Remove ~240 lines HSL, add 5 theme constants |
| `~/.config/tit/themes/gfx.toml` | Generated | New structure |
| `~/.config/tit/themes/spring.toml` | Generated | Expert color mapping |
| `~/.config/tit/themes/summer.toml` | Generated | Expert color mapping |
| `~/.config/tit/themes/autumn.toml` | Generated | Expert color mapping |
| `~/.config/tit/themes/winter.toml` | Generated | Expert color mapping |

**Total Impact:** 1 major file refactor + 5 theme regenerations

---

## Testing Strategy

### Phase Testing
1. **Phase 1-2**: Verify GFX theme loads correctly, no visual regressions
2. **Phase 3**: Verify all 5 themes generate without errors
3. **Phase 4**: Test theme switching across all 5 themes
4. **Phase 5**: Comprehensive visual testing across all UI components

### Critical Test Scenarios

**Theme Loading:**
```bash
# Verify all themes generate
./build.sh
# Check theme files created
ls ~/.config/tit/themes/
# Expected: gfx.toml spring.toml summer.toml autumn.toml winter.toml
```

**Visual Verification:**
1. **All 42 semantic elements render correctly in each theme**
2. **No theme shows broken colors or missing mappings**  
3. **Readability maintained across all terminal variations**
4. **Status colors semantically appropriate (green=good, red=bad)**

**Specific Component Tests:**
- **Confirmation dialogs**: Background solid, text readable
- **Status indicators**: Clean vs dirty clearly distinguished
- **Console output**: Error/warning/info colors distinct
- **Conflict resolver**: Focused vs unfocused borders clear
- **Menu selection**: Readable text on selection background

### Acceptance Criteria

1. **Build passes**: `./build.sh` succeeds with no errors
2. **All themes generate**: 5 theme files created in ~/.config/tit/themes/
3. **No HSL code remaining**: All mathematical transformation removed
4. **GFX unchanged**: Reference theme visually identical to current
5. **Seasonal themes readable**: All text/element combinations legible
6. **Semantic correctness**: Positive states use appropriate colors for each theme
7. **Theme switching works**: All 5 themes load without errors
8. **No naming violations**: No `DefaultThemeTOML` references remaining

---

## Risk Mitigation

### High-Risk Elements
- **Complete HSL removal**: Core visual system replacement
- **5 simultaneous theme changes**: Large scope potential for errors
- **Color readability**: 168 color assignments must be tested

### Mitigation Strategies
1. **Backup current theme.go**: Preserve original before changes
2. **Phase implementation**: Test GFX first, then add seasonal themes
3. **Expert color validation**: Subagents provided mathematically sound mappings
4. **Terminal variation testing**: Test across different terminal backgrounds

### Rollback Plan
If issues arise:
1. **Restore original theme.go from backup**
2. **Keep old theme files**: Rename instead of delete during transition
3. **Selective revert**: Can revert individual seasonal themes while keeping GFX improvements

---

## LIFESTAR Compliance

- [x] **L - Lean**: Removes 240 lines of broken code, replaces with explicit definitions
- [x] **I - Immutable**: Deterministic color assignments, no mathematical transformations
- [x] **F - Findable**: All colors in [colors] section, all mappings in [theme] section  
- [x] **E - Explicit**: Named colors with semantic hooks, no hidden transformations
- [x] **S - SSOT**: GFX theme as reference, seasonal themes as explicit variants
- [x] **T - Testable**: Clear acceptance criteria, visual verification possible
- [x] **A - Accessible**: Expert color assignments ensure readability and contrast
- [x] **R - Reviewable**: Clear structure, semantic naming, documented rationale

---

## Completion Criteria

### Technical Completion
- [ ] All HSL transformation code removed
- [ ] GFX theme restructured with named colors
- [ ] 4 seasonal themes implemented with expert mappings
- [ ] Theme generation function updated
- [ ] All 5 theme files generate correctly

### Quality Assurance  
- [ ] Build passes without errors
- [ ] All themes load without broken colors
- [ ] Visual verification across all UI components
- [ ] Theme switching works between all 5 themes
- [ ] Readability maintained in all themes

### Documentation
- [ ] COUNSELOR summary document created
- [ ] Color mapping rationale documented
- [ ] Testing results recorded

---

**This represents a complete architectural transformation of TIT's theme system, replacing mathematical approximations with explicit, expertly-curated color relationships that honor both semantic meaning and visual excellence.**

**Ready for ENGINEER execution.**

Rock 'n Roll!
**JRENG!**