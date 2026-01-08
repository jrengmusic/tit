# Color Organization & Theme System

All colors in TIT are centralized in the theme system. No hardcoded colors in code.

**Naming Rule:** Color names describe WHERE/WHAT they're used, not what they look like.

## Theme Structure

### Backgrounds
- `MainBackgroundColor` - Main app background (#090D12)
- `InlineBackgroundColor` - Secondary areas (#1B2A31)
- `SelectionBackgroundColor` - Highlight/selection areas (#0D141C)

### Text Colors - Content & Body
- `ContentTextColor` - Body text in boxes, content areas (#4E8C93)
- `LabelTextColor` - Labels, headers, bright UI lines (#8CC9D9)
- `DimmedTextColor` - Disabled/muted elements (#33535B)
- `AccentTextColor` - Keyboard shortcuts, warnings (#01C2D2)
- `HighlightTextColor` - Bright contrast text (#D1D5DA)
- `TerminalTextColor` - Command output, terminal text (#999999)

### Special Text Colors
- `CwdTextColor` - Current working directory (#67DFEF)
- `FooterTextColor` - Footer hints (#519299)

### Borders
- `BoxBorderColor` - Borders for all boxes (header, content, inputs) (#8CC9D9)

### Status Colors
- `StatusClean` - Clean working tree (#01C2D2)
- `StatusModified` - Modified files (#FC704C)

### Timeline Colors
- `TimelineSynchronized` - In sync with remote (#01C2D2)
- `TimelineLocalAhead` - Local ahead of remote (#00C8D8)
- `TimelineLocalBehind` - Local behind remote (#F2AB53)

### UI Elements
- `MenuSelectionBackground` - Menu selection highlight (#7EB8C5)

### Diff Colors
- `DiffAddedLineColor` - Added lines in diffs, history view (#5A9C7A - muted green)
- `DiffRemovedLineColor` - Removed lines in diffs, history view (#B07070 - muted red/burgundy)

## Usage Pattern

All components receive `theme Theme` parameter and reference colors as:
```go
Foreground(lipgloss.Color(theme.SecondaryTextColor))
Background(lipgloss.Color(theme.MenuSelectionBackground))
```

## Adding New Colors

1. Add to `DefaultThemeTOML` TOML string
2. Add field to `ThemeDefinition.Palette` struct
3. Add field to `Theme` struct with category comment
4. Update `LoadTheme()` to map the new field
5. Use in code via `theme.FieldName`

## Theme File Location
`~/.config/tit/themes/default.toml`

All changes are user-configurable without code changes.
