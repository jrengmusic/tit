package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// EnsureFiveThemesExist creates/regenerates all 5 themes at startup
func EnsureFiveThemesExist() error {
	configThemeDir := filepath.Join(getConfigDirectory(), "themes")
	if err := os.MkdirAll(configThemeDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

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

// GfxTheme is the default TIT theme - all other themes derive from this reference
const GfxTheme = `name = "GFX"
description = "TIT default theme - reference for all other themes"

[palette]
# Backgrounds
mainBackgroundColor = "#090D12"       # bunker
inlineBackgroundColor = "#1B2A31"     # dark
selectionBackgroundColor = "#0D141C"  # corbeau

# Text - Content & Body
contentTextColor = "#4E8C93"          # paradiso
labelTextColor = "#8CC9D9"            # dolphin
dimmedTextColor = "#33535B"           # mediterranea
accentTextColor = "#01C2D2"           # caribbeanBlue
highlightTextColor = "#D1D5DA"        # off-white
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#67DFEF"              # poseidonJr
footerTextColor = "#519299"           # lagoon

# Borders
boxBorderColor = "#8CC9D9"            # dolphin
separatorColor = "#1B2A31"            # dark

# Confirmation Dialog
confirmationDialogBackground = "#112130"  # trappedDarkness

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#8CC9D9"

# Conflict Resolver - Selection
conflictSelectionForeground = "#090D12"  # bunker
conflictSelectionBackground = "#7EB8C5"  # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#8CC9D9"       # dolphin

# Status Colors
statusClean = "#01C2D2"               # caribbeanBlue
statusDirty = "#FC704C"               # preciousPersimmon

# Timeline Colors
timelineSynchronized = "#01C2D2"      # caribbeanBlue
timelineLocalAhead = "#00C8D8"        # blueBikini
timelineLocalBehind = "#F2AB53"       # safflower

# Operation Colors
operationReady = "#4ECB71"            # emerald green
operationNotRepo = "#FC704C"          # preciousPersimmon
operationTimeTravel = "#F2AB53"       # safflower
operationConflicted = "#FC704C"       # preciousPersimmon
operationMerging = "#00C8D8"          # blueBikini
operationRebasing = "#00C8D8"         # blueBikini
operationDirtyOp = "#FC704C"          # preciousPersimmon

# UI Elements / Buttons
menuSelectionBackground = "#7EB8C5"   # brighter muted teal
buttonSelectedTextColor = "#0D1418"   # dark text

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#5A9C7A"        # muted green
diffRemovedLineColor = "#B07070"      # muted red

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#FC704C"         # preciousPersimmon
outputStatusColor = "#01C2D2"         # caribbeanBlue
outputWarningColor = "#F2AB53"        # safflower
outputDebugColor = "#33535B"          # mediterranea
outputInfoColor = "#01C2D2"           # caribbeanBlue
`

// SpringTheme is a spring-themed color palette with greens and vibrant energy
const SpringTheme = `name = "Spring"
description = "Fresh spring greens with vibrant energy"

[palette]
# Backgrounds - sapphire → ceruleanBlue → sapphire gradient
mainBackgroundColor = "#323B9E"       # sapphire (main background)
inlineBackgroundColor = "#0972BB"     # easternBlue (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - green colors for positive, red for negative
contentTextColor = "#179CA8"          # easternBlue - neutral readable
labelTextColor = "#90D88D"            # feijoa (labels)
dimmedTextColor = "#C8E189"           # yellowGreen (dimmed)
accentTextColor = "#FEEA85"           # salomie - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#FEEA85"              # salomie - bright yellow accent
footerTextColor = "#58C9BA"           # downy - muted descriptions

# Borders
boxBorderColor = "#90D88D"            # feijoa
separatorColor = "#0972BB"            # easternBlue

# Confirmation Dialog
confirmationDialogBackground = "#244DA8"  # ceruleanBlue

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#90D88D"

# Conflict Resolver - Selection
conflictSelectionForeground = "#323B9E"  # sapphire
conflictSelectionBackground = "#7EB8C5"  # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#90D88D"       # feijoa

# Status Colors - green for clean, red for dirty
statusClean = "#5BCF90"               # emerald - vibrant positive
statusDirty = "#FD5B68"               # wildWatermelon (dirty = red)

# Timeline Colors
timelineSynchronized = "#4ECB71"      # emerald (synced)
timelineLocalAhead = "#5BCF90"        # emerald (ahead)
timelineLocalBehind = "#F67F78"       # froly (behind)

# Operation Colors - green for positive operations
operationReady = "#4ECB71"            # emerald (ready)
operationNotRepo = "#FD5B68"          # wildWatermelon (not repo)
operationTimeTravel = "#F19A84"       # apricot (time travel)
operationConflicted = "#FD5B68"       # wildWatermelon (conflicted)
operationMerging = "#5BCF90"          # emerald (merging)
operationRebasing = "#5BCF90"         # emerald (rebasing)
operationDirtyOp = "#FD5B68"          # wildWatermelon (dirty)

# UI Elements
menuSelectionBackground = "#5BCF90"   # emerald - natural green
buttonSelectedTextColor = "#3F2894"   # daisyBush - dark contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#5BCF90"        # emerald (added)
diffRemovedLineColor = "#FD5B68"      # wildWatermelon (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#FD5B68"         # wildWatermelon
outputStatusColor = "#4ECB71"         # emerald
outputWarningColor = "#F67F78"        # froly
outputDebugColor = "#C8E189"          # yellowGreen
outputInfoColor = "#37CB9F"           # shamrock
`

// SummerTheme is a summer-themed color palette with electric blues and bright sunshine
const SummerTheme = `name = "Summer"
description = "Warm summer blues and bright sunshine"

[palette]
# Backgrounds - blueMarguerite → havelockBlue → violetBlue
mainBackgroundColor = "#000000"       # black (main background)
inlineBackgroundColor = "#4D88D1"     # havelockBlue (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - electric cyan/bright for positives, hot reds for negatives
contentTextColor = "#3CA7E0"          # violetBlue - readable neutral
labelTextColor = "#19E5FF"            # cyan (labels)
dimmedTextColor = "#5E68C1"           # indigo (dimmed)
accentTextColor = "#FFBF16"           # lightningYellow - electric shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#FFBF16"              # lightningYellow - electric accent
footerTextColor = "#8667BF"           # blueMarguerite - muted descriptions

# Borders
boxBorderColor = "#19E5FF"            # cyan
separatorColor = "#4D88D1"            # havelockBlue

# Confirmation Dialog
confirmationDialogBackground = "#2BC6F0"  # pictonBlue

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#19E5FF"

# Conflict Resolver - Selection
conflictSelectionForeground = "#3CA7E0"   # pictonBlue
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#19E5FF"        # cyan

# Status Colors - electric cyan for clean, hot red for dirty
statusClean = "#19E5FF"               # cyan - electric positive
statusDirty = "#FF3469"               # radicalRed (dirty)

# Timeline Colors
timelineSynchronized = "#00FFFF"      # electric cyan (synced)
timelineLocalAhead = "#19E5FF"        # cyan (ahead)
timelineLocalBehind = "#FF9700"       # pizazz (behind)

# Operation Colors - electric colors for positive ops
operationReady = "#00FFFF"            # electric cyan (ready)
operationNotRepo = "#FF3469"          # radicalRed (not repo)
operationTimeTravel = "#FFBF16"       # lightningYellow (time travel)
operationConflicted = "#FF3469"       # radicalRed (conflicted)
operationMerging = "#19E5FF"          # cyan (merging)
operationRebasing = "#19E5FF"         # cyan (rebasing)
operationDirtyOp = "#FF3469"          # radicalRed (dirty)

# UI Elements
menuSelectionBackground = "#FE62B9"   # hotPink - electric highlight
buttonSelectedTextColor = "#8667BF"   # blueMarguerite - dark contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#19E5FF"        # cyan (added)
diffRemovedLineColor = "#FF3469"      # radicalRed (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#FF3469"         # radicalRed
outputStatusColor = "#00FFFF"         # electric cyan
outputWarningColor = "#FF9700"        # pizazz
outputDebugColor = "#5E68C1"          # indigo
outputInfoColor = "#2BC6F0"           # pictonBlue
`

// AutumnTheme is an autumn-themed color palette with rich golds and warm earth tones
const AutumnTheme = `name = "Autumn"
description = "Rich autumn oranges and warm earth tones"

[palette]
# Backgrounds - jacaranda → mulberryWood → roseBudCherry
mainBackgroundColor = "#3E0338"       # jacaranda (main background)
inlineBackgroundColor = "#5E063E"     # mulberryWood (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - gold colors for positive, deep reds for negative
contentTextColor = "#E78C79"          # apricot - warm readable
labelTextColor = "#F9C94D"            # saffronMango (labels)
dimmedTextColor = "#F09D06"           # tulipTree (dimmed)
accentTextColor = "#F5BB09"           # corn - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#F5BB09"              # corn - golden bright
footerTextColor = "#CD5861"           # chestnutRose - muted descriptions

# Borders
boxBorderColor = "#F9C94D"            # saffronMango
separatorColor = "#5E063E"            # mulberryWood

# Confirmation Dialog
confirmationDialogBackground = "#7D0E36"  # roseBudCherry

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#F9C94D"

# Conflict Resolver - Selection
conflictSelectionForeground = "#3E0338"   # jacaranda
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#F9C94D"        # saffronMango

# Status Colors - gold for clean, deep red for dirty
statusClean = "#F5BB09"               # corn (clean = gold)
statusDirty = "#DC3003"               # grenadier (dirty = deep red)

# Timeline Colors
timelineSynchronized = "#F5BB09"      # corn (synced)
timelineLocalAhead = "#F9C94D"        # saffronMango (ahead)
timelineLocalBehind = "#E85C03"       # trinidad (behind)

# Operation Colors - gold colors for positive ops
operationReady = "#F5BB09"            # corn (ready)
operationNotRepo = "#DC3003"          # grenadier (not repo)
operationTimeTravel = "#F2AB53"       # safflower (time travel)
operationConflicted = "#DC3003"       # grenadier (conflicted)
operationMerging = "#F5BB09"          # corn (merging)
operationRebasing = "#F5BB09"         # corn (rebasing)
operationDirtyOp = "#DC3003"          # grenadier (dirty)

# UI Elements
menuSelectionBackground = "#F1AE37"   # tulipTree - golden harvest highlight
buttonSelectedTextColor = "#3E0338"   # jacaranda - darkest contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#F5BB09"        # corn (added)
diffRemovedLineColor = "#DC3003"      # grenadier (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#DC3003"         # grenadier
outputStatusColor = "#F5BB09"         # corn
outputWarningColor = "#E85C03"        # trinidad
outputDebugColor = "#F09D06"          # tulipTree
outputInfoColor = "#F48C06"           # tangerine
`

// WinterTheme is a winter-themed color palette with professional blues and subtle elegance
const WinterTheme = `name = "Winter"
description = "Cool winter purples with subtle elegance"

[palette]
# Backgrounds - cloudBurst → sanJuan → sanMarino
mainBackgroundColor = "#233253"       # cloudBurst (main background)
inlineBackgroundColor = "#334676"     # sanJuan (secondary areas)
selectionBackgroundColor = "#090D12"  # bunker (highlight areas)

# Text - Content & Body - professional blues for positive, soft pinks for negative
contentTextColor = "#CAD0E6"          # cyanGray - cool readable
labelTextColor = "#7F95D6"            # chetwodeBlue (labels)
dimmedTextColor = "#9BA9D0"           # rockBlue (dimmed)
accentTextColor = "#F6F5FA"           # whisper - bright shortcuts
highlightTextColor = "#D1D5DA"        # off-white (highlights)
terminalTextColor = "#999999"         # neutral gray

# Special Text
cwdTextColor = "#F6F5FA"              # whisper - bright white
footerTextColor = "#9BA9D0"           # rockBlue - muted descriptions

# Borders
boxBorderColor = "#7F95D6"            # chetwodeBlue
separatorColor = "#334676"            # sanJuan

# Confirmation Dialog
confirmationDialogBackground = "#233253"  # cloudBurst

# Conflict Resolver - Borders
conflictPaneUnfocusedBorder = "#2C4144"
conflictPaneFocusedBorder = "#7F95D6"

# Conflict Resolver - Selection
conflictSelectionForeground = "#233253"   # cloudBurst
conflictSelectionBackground = "#7EB8C5"   # brighter muted teal

# Conflict Resolver - Pane Headers
conflictPaneTitleColor = "#7F95D6"        # chetwodeBlue

# Status Colors - professional blue for clean, soft pink for dirty
statusClean = "#6281DC"               # havelockBlue - professional positive
statusDirty = "#E0BACF"               # melanie (dirty = soft pink)

# Timeline Colors
timelineSynchronized = "#435A98"      # sanMarino (synced)
timelineLocalAhead = "#6281DC"        # havelockBlue (ahead)
timelineLocalBehind = "#CEBAC5"       # lily (behind)

# Operation Colors - professional blue colors for positive ops
operationReady = "#435A98"            # sanMarino (ready)
operationNotRepo = "#E0BACF"          # melanie (not repo)
operationTimeTravel = "#CEBAC5"       # lily (time travel)
operationConflicted = "#E0BACF"       # melanie (conflicted)
operationMerging = "#6281DC"          # havelockBlue (merging)
operationRebasing = "#6281DC"         # havelockBlue (rebasing)
operationDirtyOp = "#E0BACF"          # melanie (dirty)

# UI Elements
menuSelectionBackground = "#7F95D6"   # chetwodeBlue - professional blue accent
buttonSelectedTextColor = "#F6F5FA"   # whisper - light contrast

# Animation
spinnerColor = "#00FFFF"              # electric cyan

# Diff Colors
diffAddedLineColor = "#6281DC"        # havelockBlue (added)
diffRemovedLineColor = "#E0BACF"      # melanie (removed)

# Console Output Colors
outputStdoutColor = "#999999"         # neutral gray
outputStderrColor = "#E0BACF"         # melanie
outputStatusColor = "#435A98"         # sanMarino
outputWarningColor = "#CEBAC5"        # lily
outputDebugColor = "#9BA9D0"          # rockBlue
outputInfoColor = "#435A98"           # sanMarino
`

// ThemeDefinition represents a theme file structure
type ThemeDefinition struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Palette     struct {
		// Backgrounds
		MainBackgroundColor      string `toml:"mainBackgroundColor"`
		InlineBackgroundColor    string `toml:"inlineBackgroundColor"`
		SelectionBackgroundColor string `toml:"selectionBackgroundColor"`

		// Text - Content & Body
		ContentTextColor   string `toml:"contentTextColor"`
		LabelTextColor     string `toml:"labelTextColor"`
		DimmedTextColor    string `toml:"dimmedTextColor"`
		AccentTextColor    string `toml:"accentTextColor"`
		HighlightTextColor string `toml:"highlightTextColor"`
		TerminalTextColor  string `toml:"terminalTextColor"`

		// Special Text
		CwdTextColor    string `toml:"cwdTextColor"`
		FooterTextColor string `toml:"footerTextColor"`

		// Borders
		BoxBorderColor string `toml:"boxBorderColor"`
		SeparatorColor string `toml:"separatorColor"`

		// Confirmation Dialog
		ConfirmationDialogBackground string `toml:"confirmationDialogBackground"`

		// Conflict Resolver - Borders
		ConflictPaneUnfocusedBorder string `toml:"conflictPaneUnfocusedBorder"`
		ConflictPaneFocusedBorder   string `toml:"conflictPaneFocusedBorder"`

		// Conflict Resolver - Selection
		ConflictSelectionForeground string `toml:"conflictSelectionForeground"`
		ConflictSelectionBackground string `toml:"conflictSelectionBackground"`

		// Conflict Resolver - Pane Headers
		ConflictPaneTitleColor string `toml:"conflictPaneTitleColor"`

		// Status Colors
		StatusClean string `toml:"statusClean"`
		StatusDirty string `toml:"statusDirty"`

		// Timeline Colors
		TimelineSynchronized string `toml:"timelineSynchronized"`
		TimelineLocalAhead   string `toml:"timelineLocalAhead"`
		TimelineLocalBehind  string `toml:"timelineLocalBehind"`

		// Operation Colors
		OperationReady      string `toml:"operationReady"`
		OperationNotRepo    string `toml:"operationNotRepo"`
		OperationTimeTravel string `toml:"operationTimeTravel"`
		OperationConflicted string `toml:"operationConflicted"`
		OperationMerging    string `toml:"operationMerging"`
		OperationRebasing   string `toml:"operationRebasing"`
		OperationDirtyOp    string `toml:"operationDirtyOp"`

		// UI Elements / Buttons
		MenuSelectionBackground string `toml:"menuSelectionBackground"`
		ButtonSelectedTextColor string `toml:"buttonSelectedTextColor"`

		// Animation
		SpinnerColor string `toml:"spinnerColor"`

		// Diff Colors
		DiffAddedLineColor   string `toml:"diffAddedLineColor"`
		DiffRemovedLineColor string `toml:"diffRemovedLineColor"`

		// Console Output Colors
		OutputStdoutColor  string `toml:"outputStdoutColor"`
		OutputStderrColor  string `toml:"outputStderrColor"`
		OutputStatusColor  string `toml:"outputStatusColor"`
		OutputWarningColor string `toml:"outputWarningColor"`
		OutputDebugColor   string `toml:"outputDebugColor"`
		OutputInfoColor    string `toml:"outputInfoColor"`
	} `toml:"palette"`
}

// Theme defines all semantic colors from the active theme
type Theme struct {
	// Backgrounds
	MainBackgroundColor      string
	InlineBackgroundColor    string
	SelectionBackgroundColor string

	// Text - Content & Body
	ContentTextColor   string
	LabelTextColor     string
	DimmedTextColor    string
	AccentTextColor    string
	HighlightTextColor string
	TerminalTextColor  string

	// Special Text
	CwdTextColor    string
	FooterTextColor string

	// Borders
	BoxBorderColor string
	SeparatorColor string

	// Confirmation Dialog
	ConfirmationDialogBackground string

	// Conflict Resolver - Borders
	ConflictPaneUnfocusedBorder string
	ConflictPaneFocusedBorder   string

	// Conflict Resolver - Selection
	ConflictSelectionForeground string
	ConflictSelectionBackground string

	// Conflict Resolver - Pane Headers
	ConflictPaneTitleColor string

	// Status Colors
	StatusClean string
	StatusDirty string

	// Timeline Colors
	TimelineSynchronized string
	TimelineLocalAhead   string
	TimelineLocalBehind  string

	// Operation Colors
	OperationReady      string
	OperationNotRepo    string
	OperationTimeTravel string
	OperationConflicted string
	OperationMerging    string
	OperationRebasing   string
	OperationDirtyOp    string

	// UI Elements / Buttons
	MenuSelectionBackground string
	ButtonSelectedTextColor string

	// Animation
	SpinnerColor string

	// Diff Colors
	DiffAddedLineColor   string
	DiffRemovedLineColor string

	// Console Output Colors
	OutputStdoutColor  string
	OutputStderrColor  string
	OutputStatusColor  string
	OutputWarningColor string
	OutputDebugColor   string
	OutputInfoColor    string
}

// LoadTheme loads a theme from a TOML file
func LoadTheme(themeFilePath string) (Theme, error) {
	fileData, err := os.ReadFile(themeFilePath)
	if err != nil {
		return Theme{}, fmt.Errorf("failed to read theme file: %w", err)
	}

	var themeDef ThemeDefinition
	if err := toml.Unmarshal(fileData, &themeDef); err != nil {
		return Theme{}, fmt.Errorf("failed to parse theme file: %w", err)
	}

	theme := Theme{
		// Backgrounds
		MainBackgroundColor:      themeDef.Palette.MainBackgroundColor,
		InlineBackgroundColor:    themeDef.Palette.InlineBackgroundColor,
		SelectionBackgroundColor: themeDef.Palette.SelectionBackgroundColor,

		// Text - Content & Body
		ContentTextColor:   themeDef.Palette.ContentTextColor,
		LabelTextColor:     themeDef.Palette.LabelTextColor,
		DimmedTextColor:    themeDef.Palette.DimmedTextColor,
		AccentTextColor:    themeDef.Palette.AccentTextColor,
		HighlightTextColor: themeDef.Palette.HighlightTextColor,
		TerminalTextColor:  themeDef.Palette.TerminalTextColor,

		// Special Text
		CwdTextColor:    themeDef.Palette.CwdTextColor,
		FooterTextColor: themeDef.Palette.FooterTextColor,

		// Borders
		BoxBorderColor: themeDef.Palette.BoxBorderColor,
		SeparatorColor: themeDef.Palette.SeparatorColor,

		// Confirmation Dialog
		ConfirmationDialogBackground: themeDef.Palette.ConfirmationDialogBackground,

		// Conflict Resolver - Borders
		ConflictPaneUnfocusedBorder: themeDef.Palette.ConflictPaneUnfocusedBorder,
		ConflictPaneFocusedBorder:   themeDef.Palette.ConflictPaneFocusedBorder,

		// Conflict Resolver - Selection
		ConflictSelectionForeground: themeDef.Palette.ConflictSelectionForeground,
		ConflictSelectionBackground: themeDef.Palette.ConflictSelectionBackground,

		// Conflict Resolver - Pane Headers
		ConflictPaneTitleColor: themeDef.Palette.ConflictPaneTitleColor,

		// Status Colors
		StatusClean: themeDef.Palette.StatusClean,
		StatusDirty: themeDef.Palette.StatusDirty,

		// Timeline Colors
		TimelineSynchronized: themeDef.Palette.TimelineSynchronized,
		TimelineLocalAhead:   themeDef.Palette.TimelineLocalAhead,
		TimelineLocalBehind:  themeDef.Palette.TimelineLocalBehind,

		// Operation Colors
		OperationReady:      themeDef.Palette.OperationReady,
		OperationNotRepo:    themeDef.Palette.OperationNotRepo,
		OperationTimeTravel: themeDef.Palette.OperationTimeTravel,
		OperationConflicted: themeDef.Palette.OperationConflicted,
		OperationMerging:    themeDef.Palette.OperationMerging,
		OperationRebasing:   themeDef.Palette.OperationRebasing,
		OperationDirtyOp:    themeDef.Palette.OperationDirtyOp,

		// UI Elements / Buttons
		MenuSelectionBackground: themeDef.Palette.MenuSelectionBackground,
		ButtonSelectedTextColor: themeDef.Palette.ButtonSelectedTextColor,

		// Animation
		SpinnerColor: themeDef.Palette.SpinnerColor,

		// Diff Colors
		DiffAddedLineColor:   themeDef.Palette.DiffAddedLineColor,
		DiffRemovedLineColor: themeDef.Palette.DiffRemovedLineColor,

		// Console Output Colors
		OutputStdoutColor:  themeDef.Palette.OutputStdoutColor,
		OutputStderrColor:  themeDef.Palette.OutputStderrColor,
		OutputStatusColor:  themeDef.Palette.OutputStatusColor,
		OutputWarningColor: themeDef.Palette.OutputWarningColor,
		OutputDebugColor:   themeDef.Palette.OutputDebugColor,
		OutputInfoColor:    themeDef.Palette.OutputInfoColor,
	}

	return theme, nil
}

// CreateDefaultThemeIfMissing creates or regenerates all 5 themes (gfx + 4 seasons)
// SSOT: Always regenerates from GfxTheme to ensure all colors are current
func CreateDefaultThemeIfMissing() (string, error) {
	return "", EnsureFiveThemesExist()
}

// LoadDefaultTheme loads the default gfx theme
func LoadDefaultTheme() (Theme, error) {
	themeFile := filepath.Join(getConfigDirectory(), "themes", "gfx.toml")
	return LoadTheme(themeFile)
}

// getConfigDirectory returns the TIT config directory
func getConfigDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".tit"
	}
	return filepath.Join(home, ".config", "tit")
}

// DiscoverAvailableThemes returns a list of available theme names
func DiscoverAvailableThemes() ([]string, error) {
	themesDir := filepath.Join(getConfigDirectory(), "themes")

	files, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read themes directory: %w", err)
	}

	var themes []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".toml" {
			themeName := file.Name()[:len(file.Name())-5] // Remove .toml extension
			themes = append(themes, themeName)
		}
	}

	return themes, nil
}

// LoadThemeByName loads a theme by name from the themes directory
func LoadThemeByName(themeName string) (Theme, error) {
	themeFile := filepath.Join(getConfigDirectory(), "themes", themeName+".toml")
	return LoadTheme(themeFile)
}

// GetNextTheme cycles to the next available theme
func GetNextTheme(currentTheme string) (string, error) {
	themes, err := DiscoverAvailableThemes()
	if err != nil {
		return "", err
	}

	if len(themes) == 0 {
		return "", fmt.Errorf("no themes found")
	}

	// Find current theme index
	currentIndex := -1
	for i, theme := range themes {
		if theme == currentTheme {
			currentIndex = i
			break
		}
	}

	// Cycle to next (or first if current not found)
	nextIndex := (currentIndex + 1) % len(themes)
	return themes[nextIndex], nil
}
