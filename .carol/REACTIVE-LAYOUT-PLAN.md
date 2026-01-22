# Reactive Layout Implementation Plan

**Author:** ANALYST (Amp Claude Sonnet 4)  
**Date:** 2026-01-22  
**Status:** Ready for SCAFFOLDER execution

---

## Goal

Transform TIT from fixed 80×46 centered layout to reactive full-terminal layout:
- Header fixed at top (10 lines)
- Footer fixed at bottom (2 lines)
- Content fills remaining vertical space
- Banner conditionally renders in header when width permits
- Graceful degradation when terminal too small

---

## Layout Specification

### Terminal Size Thresholds

```
MIN_WIDTH         = 60   // Minimum usable width (header info only)
MIN_WIDTH_BANNER  = 100  // Width threshold to show banner in header
MIN_HEIGHT        = 18   // Minimum usable height (header + footer + 4 content lines)

HEADER_HEIGHT     = 11   // Fixed (9 content lines + 1 top padding + 1 bottom padding)
FOOTER_HEIGHT     = 2    // Fixed
MIN_CONTENT_HEIGHT = 4   // Minimum content area
BANNER_HEIGHT     = 11   // Same as header for alignment
```

### Sizing Formulas

```
TerminalWidth  = tea.WindowSizeMsg.Width
TerminalHeight = tea.WindowSizeMsg.Height

// Validation
IsTooSmall = (TerminalWidth < MIN_WIDTH) || (TerminalHeight < MIN_HEIGHT)
ShowBanner = (TerminalWidth >= MIN_WIDTH_BANNER)

// Dynamic dimensions
ContentHeight = TerminalHeight - HEADER_HEIGHT - FOOTER_HEIGHT
ContentHeight = max(ContentHeight, MIN_CONTENT_HEIGHT)

// Child margins (applied inside each section)
HORIZONTAL_MARGIN = 2  // Left + right padding for children
VERTICAL_MARGIN   = 1  // Top + bottom padding for children (where applicable)

// Usable inner widths
HeaderInnerWidth  = TerminalWidth - (HORIZONTAL_MARGIN * 2)
ContentInnerWidth = TerminalWidth - (HORIZONTAL_MARGIN * 2)
FooterInnerWidth  = TerminalWidth - (HORIZONTAL_MARGIN * 2)

// Header column split (when banner shown)
BANNER_WIDTH      = 30  // Fixed width for braille banner
INFO_WIDTH        = HeaderInnerWidth - BANNER_WIDTH - 2  // -2 for gap between columns

// Header column (no banner)
INFO_WIDTH_FULL   = HeaderInnerWidth
```

### Layout Structure

```
┌─────────────────────────────────────────────────────────────────┐ ← Terminal top
│ HEADER (full width, HEADER_HEIGHT lines)                        │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ INFO COLUMN          │ gap │ BANNER COLUMN │ margin│ │
│ │        │ (2 sub-columns)      │     │ (braille)     │       │ │
│ │        │ vertically centered  │     │ max height    │       │ │
│ └─────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│ CONTENT (full width, ContentHeight lines)                       │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ CONTENT CHILDREN                           │ margin│ │
│ │        │ (menu, input, console, history, etc.)      │       │ │
│ └─────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│ FOOTER (full width, FOOTER_HEIGHT lines)                        │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ margin │ FOOTER CONTENT (hints, status)             │ margin│ │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘ ← Terminal bottom
```

### Header Info Column (9 content rows + 2 padding rows = 11 lines)

```
┌─────────────────────────────────────────────────────────────────┐
│ (1 line top padding)                                            │
│ [📁] /Users/jreng/Documents/Poems/dev/tit                       │  Row 1: CWD
│ [🔗] git@github.com:jrengmusic/tit.git                          │  Row 2: Remote URL
│ ─────────────────────────────────────────────────────           │  Row 3: Separator
│ [💥] DIRTY                                                       │  Row 4: WorkingTree label
│      You have uncommitted changes.                              │  Row 5: WT description
│      Stage and commit to continue.                              │  Row 6: WT description (optional)
│ [⬆️] AHEAD                                                       │  Row 7: Timeline label
│      3 commits ahead of remote.                                 │  Row 8: TL description
│      Push to sync with remote.                                  │  Row 9: TL description (optional)
│ (1 line bottom padding)                                         │
└─────────────────────────────────────────────────────────────────┘
```

**Emoji Column Pattern (Reusable Component):**
```
┌──────┬────────────────────────────────────────────────────────┐
│ EMOJI│ CONTENT                                                 │
│ (4ch)│ (remaining width, can wrap to multiple lines)          │
├──────┼────────────────────────────────────────────────────────┤
│ [📁] │ /Users/jreng/Documents/Poems/dev/tit                   │
│ [💥] │ DIRTY                                                   │
│      │     You have uncommitted changes.                      │
│      │     Stage and commit to continue.                      │
└──────┴────────────────────────────────────────────────────────┘

EMOJI_COLUMN_WIDTH = 4  // "[E] " pattern (emoji + space)
```

**Reusable InfoRow Component:**
```go
type InfoRow struct {
    Emoji       string   // Single emoji character
    Label       string   // Primary text (bold, colored)
    Description []string // Optional wrapped lines (dimmed)
    LabelColor  string   // Color for label
}

// Returns []string (1 line for label, N lines for description)
func (r InfoRow) Render(contentWidth int) []string
```

**State Rendering Examples:**

```
WorkingTree = Clean:
[✓] CLEAN
    Working tree is clean.

WorkingTree = Modified:
[💥] DIRTY
    You have uncommitted changes.

Timeline = InSync:
[📡] SYNC
    Local and remote are in sync.

Timeline = Ahead:
[⬆️] AHEAD +3
    3 commits ahead of remote.
    Push to sync.

Timeline = Behind:
[⬇️] BEHIND -2
    2 commits behind remote.
    Pull to update.

Timeline = Diverged:
[🔀] DIVERGED +3/-2
    Local and remote have diverged.
    Pull and merge, or rebase.

Remote = NoRemote:
[🔌] NO REMOTE
    No remote configured.

Operation = TimeTraveling:
[📌] DETACHED @ abc1234
    Viewing commit from Jan 15, 2026.
```

### Too Small Screen

When `IsTooSmall == true`, render centered message:
```
Terminal too small.
Resize to at least 60×16.
```

---

## Implementation Phases

### Phase 1: Refactor Sizing SSOT

**File:** `internal/ui/sizing.go`

**Changes:**
1. Replace fixed constants with threshold constants
2. Create `DynamicSizing` struct that holds computed values
3. Add `CalculateDynamicSizing(termWidth, termHeight int) DynamicSizing` function
4. Add `IsTooSmall()` and `ShowBanner()` methods

**New Code Structure:**
```go
package ui

// Threshold constants (SSOT)
const (
    MinWidth       = 60
    MinWidthBanner = 100
    MinHeight      = 16
    HeaderHeight   = 10
    FooterHeight   = 2
    MinContentHeight = 4
    HorizontalMargin = 2
    BannerWidth    = 30
)

// DynamicSizing holds computed layout dimensions
type DynamicSizing struct {
    TerminalWidth   int
    TerminalHeight  int
    ContentHeight   int
    ContentInnerWidth int
    HeaderInnerWidth  int
    FooterInnerWidth  int
    InfoColumnWidth   int
    ShowBanner        bool
    IsTooSmall        bool
}

// CalculateDynamicSizing computes all dimensions from terminal size
func CalculateDynamicSizing(termWidth, termHeight int) DynamicSizing {
    // Implementation here
}
```

**Validation Test:**
1. Build: `./build.sh`
2. Run binary, resize terminal to various sizes
3. Add temporary debug output showing computed values
4. Verify: ContentHeight changes when terminal height changes
5. Verify: IsTooSmall triggers below 60×16
6. Verify: ShowBanner triggers at width ≥100

---

### Phase 2: Update Application State

**File:** `internal/app/app.go`

**Changes:**
1. Replace `sizing Sizing` field with `sizing DynamicSizing`
2. Update `WindowSizeMsg` handler to recalculate sizing
3. Store sizing in app state for all render functions to access

**Code Changes:**
```go
// In Application struct
sizing ui.DynamicSizing  // Was: sizing ui.Sizing

// In Update(), WindowSizeMsg case:
case tea.WindowSizeMsg:
    a.width = msg.Width
    a.height = msg.Height
    a.sizing = ui.CalculateDynamicSizing(msg.Width, msg.Height)
    return a, nil

// In Init():
a.sizing = ui.CalculateDynamicSizing(a.width, a.height)
```

**Validation Test:**
1. Build: `./build.sh`
2. App should compile without errors
3. Resize terminal → app should not crash
4. Existing functionality still works (may look broken, that's OK)

---

### Phase 3: Create New Layout Renderer

**File:** `internal/ui/layout.go`

**Changes:**
1. Create `RenderReactiveLayout()` function
2. Handle "too small" case first
3. Render Header/Content/Footer as edge-to-edge sections
4. Remove old `RenderLayout()` (or rename to `RenderLayoutLegacy()`)

**New Function Signature:**
```go
func RenderReactiveLayout(
    sizing DynamicSizing,
    theme Theme,
    headerContent string,  // Pre-rendered header content
    mainContent string,    // Pre-rendered main content
    footerContent string,  // Pre-rendered footer content
) string
```

**Implementation:**
```go
func RenderReactiveLayout(sizing DynamicSizing, theme Theme, header, content, footer string) string {
    // Too small guard
    if sizing.IsTooSmall {
        return renderTooSmallMessage(sizing.TerminalWidth, sizing.TerminalHeight)
    }
    
    // Header section (full width, HeaderHeight)
    headerSection := lipgloss.NewStyle().
        Width(sizing.TerminalWidth).
        Height(HeaderHeight).
        Render(header)
    
    // Content section (full width, ContentHeight)
    contentSection := lipgloss.NewStyle().
        Width(sizing.TerminalWidth).
        Height(sizing.ContentHeight).
        Render(content)
    
    // Footer section (full width, FooterHeight)
    footerSection := lipgloss.NewStyle().
        Width(sizing.TerminalWidth).
        Height(FooterHeight).
        Render(footer)
    
    // Stack vertically
    return lipgloss.JoinVertical(lipgloss.Top, headerSection, contentSection, footerSection)
}

func renderTooSmallMessage(w, h int) string {
    msg := "Terminal too small.\nResize to at least 60×16."
    return lipgloss.NewStyle().
        Width(w).
        Height(h).
        Align(lipgloss.Center).
        AlignVertical(lipgloss.Center).
        Render(msg)
}
```

**Validation Test:**
1. Build: `./build.sh`
2. Resize terminal below 60×16 → see "too small" message
3. Resize above threshold → see stacked sections (may be unstyled)

---

### Phase 4: Create InfoRow Component

**File:** `internal/ui/inforow.go` (NEW)

**Purpose:** Reusable component for emoji-prefixed rows with optional description lines.

**Implementation:**
```go
package ui

import (
    "strings"
    "github.com/charmbracelet/lipgloss"
)

const (
    EmojiColumnWidth = 4  // "[E] " pattern
)

// InfoRow represents a single info row with emoji, label, and optional descriptions
type InfoRow struct {
    Emoji       string   // Single emoji character (wide emoji only!)
    Label       string   // Primary text (bold, colored)
    Description []string // Optional continuation lines (dimmed)
    LabelColor  string   // Color for label text
    DescColor   string   // Color for description text
}

// Render returns the row as multiple lines
// Line 1: "[E] LABEL"
// Line 2+: "    description..." (indented to align with label)
func (r InfoRow) Render(contentWidth int, theme Theme) []string {
    var lines []string
    
    // Emoji column (fixed width)
    emojiCol := lipgloss.NewStyle().
        Width(EmojiColumnWidth).
        Render(r.Emoji + " ")
    
    // Label (remaining width)
    labelWidth := contentWidth - EmojiColumnWidth
    labelStyled := lipgloss.NewStyle().
        Width(labelWidth).
        Bold(true).
        Foreground(lipgloss.Color(r.LabelColor)).
        Render(r.Label)
    
    lines = append(lines, emojiCol + labelStyled)
    
    // Description lines (indented)
    indent := strings.Repeat(" ", EmojiColumnWidth)
    descStyle := lipgloss.NewStyle().
        Width(labelWidth).
        Foreground(lipgloss.Color(r.DescColor))
    
    for _, desc := range r.Description {
        lines = append(lines, indent + descStyle.Render(desc))
    }
    
    return lines
}

// RenderSeparator returns a horizontal line
func RenderSeparator(width int, color string) string {
    return lipgloss.NewStyle().
        Foreground(lipgloss.Color(color)).
        Render(strings.Repeat("─", width))
}
```

**Validation Test:**
1. Build: `./build.sh`
2. Create test usage in header rendering
3. Verify emoji alignment (4-char column)
4. Verify description indentation matches

---

### Phase 5: Refactor Header Rendering

**File:** `internal/ui/header.go` (NEW)

**Changes:**
1. Create `RenderHeaderInfo()` using InfoRow components
2. Create `RenderHeader()` that combines info + optional banner
3. Apply 1-line top/bottom padding for 11-line total

**Implementation:**
```go
package ui

import (
    "github.com/charmbracelet/lipgloss"
)

// RenderHeaderInfo builds the 9-line info column from git state
func RenderHeaderInfo(sizing DynamicSizing, theme Theme, state HeaderState) string {
    var allLines []string
    contentWidth := sizing.InfoColumnWidth
    
    // Row 1: CWD
    cwdRow := InfoRow{
        Emoji:      "📁",
        Label:      state.CWD,
        LabelColor: theme.LabelTextColor,
    }
    allLines = append(allLines, cwdRow.Render(contentWidth, theme)...)
    
    // Row 2: Remote
    remoteRow := InfoRow{
        Emoji:      state.RemoteEmoji,
        Label:      state.RemoteLabel,
        LabelColor: state.RemoteColor,
    }
    allLines = append(allLines, remoteRow.Render(contentWidth, theme)...)
    
    // Row 3: Separator
    allLines = append(allLines, RenderSeparator(contentWidth, theme.BoxBorderColor))
    
    // Rows 4-6: WorkingTree
    wtRow := InfoRow{
        Emoji:       state.WorkingTreeEmoji,
        Label:       state.WorkingTreeLabel,
        Description: state.WorkingTreeDesc,
        LabelColor:  state.WorkingTreeColor,
        DescColor:   theme.ContentTextColor,
    }
    allLines = append(allLines, wtRow.Render(contentWidth, theme)...)
    
    // Rows 7-9: Timeline
    tlRow := InfoRow{
        Emoji:       state.TimelineEmoji,
        Label:       state.TimelineLabel,
        Description: state.TimelineDesc,
        LabelColor:  state.TimelineColor,
        DescColor:   theme.ContentTextColor,
    }
    allLines = append(allLines, tlRow.Render(contentWidth, theme)...)
    
    return strings.Join(allLines, "\n")
}

// RenderHeader renders complete header with optional banner
func RenderHeader(sizing DynamicSizing, theme Theme, info string, banner string) string {
    // Add 1-line padding top and bottom
    paddedInfo := "\n" + info + "\n"
    
    marginStyle := lipgloss.NewStyle().
        PaddingLeft(HorizontalMargin).
        PaddingRight(HorizontalMargin)
    
    if !sizing.ShowBanner {
        // Info only, full width
        infoStyled := lipgloss.NewStyle().
            Width(sizing.HeaderInnerWidth).
            Height(HeaderHeight).
            Render(paddedInfo)
        return marginStyle.Render(infoStyled)
    }
    
    // Info + Banner side by side
    infoColumn := lipgloss.NewStyle().
        Width(sizing.InfoColumnWidth).
        Height(HeaderHeight).
        Render(paddedInfo)
    
    bannerColumn := lipgloss.NewStyle().
        Width(BannerWidth).
        Height(HeaderHeight).
        Render(banner)
    
    joined := lipgloss.JoinHorizontal(lipgloss.Top, infoColumn, "  ", bannerColumn)
    return marginStyle.Render(joined)
}

// HeaderState holds pre-computed state for header rendering
type HeaderState struct {
    CWD               string
    RemoteEmoji       string
    RemoteLabel       string
    RemoteColor       string
    WorkingTreeEmoji  string
    WorkingTreeLabel  string
    WorkingTreeDesc   []string
    WorkingTreeColor  string
    TimelineEmoji     string
    TimelineLabel     string
    TimelineDesc      []string
    TimelineColor     string
}
```

**Validation Test:**
1. Build: `./build.sh`
2. Width < 100 → header shows info only (9 lines + padding)
3. Width ≥ 100 → header shows info + banner side by side
4. All rows align correctly with emoji column

---

### Phase 6: Refactor Banner Rendering

**File:** `internal/ui/layout.go`

**Changes:**
1. Update `RenderBanner()` to accept dynamic dimensions
2. Banner height = BANNER_HEIGHT (11 lines)
3. Banner width = BannerWidth (30 chars)
4. Scale SVG→braille to fit these dimensions

**Updated Signature:**
```go
func RenderBanner(width, height int) string
```

**Validation Test:**
1. Build: `./build.sh`
2. Width ≥ 100 → banner renders in header right column
3. Banner fits within 30×11 character area
4. Braille logo visible and proportional

---

### Phase 7: Refactor Content Rendering

**File:** `internal/ui/layout.go`

**Changes:**
1. Update `RenderContent()` to use dynamic sizing
2. Apply horizontal margins
3. Use `sizing.ContentHeight` instead of fixed constant

**Updated Signature:**
```go
func RenderContent(sizing DynamicSizing, theme Theme, content string) string
```

**Validation Test:**
1. Build: `./build.sh`
2. Resize terminal height → content area grows/shrinks
3. Content respects horizontal margins
4. Menu items still navigable

---

### Phase 8: Update All Content Renderers

**Files to update:**
- `internal/ui/menu.go` - use dynamic ContentHeight
- `internal/ui/console.go` - use dynamic dimensions
- `internal/ui/confirmation.go` - use dynamic dimensions
- `internal/ui/history.go` - use dynamic dimensions
- `internal/ui/filehistory.go` - use dynamic dimensions
- `internal/ui/textinput.go` - use dynamic dimensions
- `internal/ui/branchinput.go` - use dynamic dimensions
- `internal/ui/input.go` - use dynamic dimensions

**Pattern for each file:**
1. Add `sizing DynamicSizing` parameter to render functions
2. Replace `ContentHeight` constant with `sizing.ContentHeight`
3. Replace `ContentInnerWidth` constant with `sizing.ContentInnerWidth`

**Validation Test (per file):**
1. Build: `./build.sh`
2. Test the specific mode (menu, console, history, etc.)
3. Resize terminal → content adapts correctly
4. No overflow or clipping

---

### Phase 9: Update Application View()

**File:** `internal/app/app.go`

**Changes:**
1. Replace `RenderLayout()` call with `RenderReactiveLayout()`
2. Pass `a.sizing` to all render functions
3. Update `RenderStateHeader()` to use dynamic sizing

**Validation Test:**
1. Build: `./build.sh`
2. All modes render correctly
3. Resize terminal → entire app adapts
4. All functionality preserved

---

### Phase 10: Cleanup

**Changes:**
1. Remove old `Sizing` struct from `sizing.go`
2. Remove old `CalculateSizing()` function
3. Remove old `RenderLayout()` function
4. Remove unused constants
5. Update any remaining references

**Validation Test:**
1. Build: `./build.sh`
2. No compiler warnings about unused code
3. Full app functionality test:
   - Menu navigation
   - All input modes
   - Console output
   - History browser
   - File history
   - Confirmation dialogs
   - Conflict resolver
4. Resize at each mode → no crashes, no overflow

---

## File Change Summary

| File | Action | Complexity |
|------|--------|------------|
| `internal/ui/sizing.go` | Rewrite | Medium |
| `internal/ui/inforow.go` | **NEW** | Low |
| `internal/ui/header.go` | **NEW** | Medium |
| `internal/ui/layout.go` | Major refactor | High |
| `internal/app/app.go` | Update sizing usage + remove old RenderStateHeader | Medium |
| `internal/ui/menu.go` | Add sizing param | Low |
| `internal/ui/console.go` | Add sizing param | Low |
| `internal/ui/confirmation.go` | Add sizing param | Low |
| `internal/ui/history.go` | Add sizing param | Low |
| `internal/ui/filehistory.go` | Add sizing param | Low |
| `internal/ui/textinput.go` | Add sizing param | Low |
| `internal/ui/branchinput.go` | Add sizing param | Low |
| `internal/ui/input.go` | Add sizing param | Low |

**Total: 13 files (2 new, 11 modified)**

---

## Risk Mitigation

1. **Phase-by-phase execution** - Each phase is independently testable
2. **Compile after each change** - Catch errors early
3. **Keep old code until validated** - Can revert if needed
4. **Test all modes** - Menu, input, console, history, dialogs

---

## Success Criteria

- [ ] Terminal resize triggers layout recalculation
- [ ] Content height adjusts to terminal height
- [ ] Banner appears/disappears at width threshold
- [ ] Too-small message shows below minimum size
- [ ] All existing modes work correctly
- [ ] No hardcoded 80, 24, 46, or 76 remaining in layout code

---

**End of Plan**

Ready for SCAFFOLDER execution.
