package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// DefaultThemeTOML is the embedded default theme content
const DefaultThemeTOML = `name = "Default (TIT)"
description = "TIT color scheme"

[palette]
mainBackgroundColor = "#090D12"       # bunker (main app background)
inlineBackgroundColor = "#1B2A31"     # dark (secondary areas)
selectionBackgroundColor = "#0D141C"  # corbeau (highlight areas)

# Text - Content & Body
contentTextColor = "#4E8C93"           # paradiso (body text in boxes)
labelTextColor = "#8CC9D9"             # dolphin (labels, headers, borders)
dimmedTextColor = "#33535B"            # mediterranea (disabled/muted)
accentTextColor = "#01C2D2"            # caribbeanBlue (keyboard shortcuts)
highlightTextColor = "#D1D5DA"         # off-white (bright contrast text)
terminalTextColor = "#999999"          # neutral gray (command output)

# Special Text
cwdTextColor = "#67DFEF"               # poseidonJr (current working directory)
footerTextColor = "#519299"            # lagoon (footer hints)

# Borders - Conflict Resolver specific
boxBorderColor = "#8CC9D9"                    # dolphin (borders for all boxes)
separatorColor = "#1B2A31"                    # dark (separator lines)
conflictPaneUnfocusedBorder = "#2C4144"       # littleMermaid (OLD-TIT EXACT - unfocused)
conflictPaneFocusedBorder = "#8CC9D9"         # dolphin (OLD-TIT EXACT - focused)

# Selection - Conflict Resolver specific
conflictSelectionForeground = "#090D12"       # bunker (selection text color)
conflictSelectionBackground = "#7EB8C5"       # brighter muted teal (selection background)

# Pane Headers
conflictPaneTitleColor = "#8CC9D9"            # dolphin (pane title text)

# Status & State
statusClean = "#01C2D2"               # caribbeanBlue
statusDirty = "#FC704C"            # preciousPersimmon
timelineSynchronized = "#01C2D2"      # caribbeanBlue
timelineLocalAhead = "#00C8D8"        # blueBikini
timelineLocalBehind = "#F2AB53"       # safflower

# Operation Colors
operationReady = "#4ECB71"            # emerald green (ready state)
operationNotRepo = "#FC704C"          # preciousPersimmon (error/not repo)
operationTimeTravel = "#F2AB53"       # safflower (time travel - warm orange)
operationConflicted = "#FC704C"       # preciousPersimmon (conflicts)
operationMerging = "#00C8D8"          # blueBikini (merge in progress)
operationRebasing = "#00C8D8"         # blueBikini (rebase in progress)
operationDirtyOp = "#FC704C"          # preciousPersimmon (dirty operation)

# Menu / Buttons
menuSelectionBackground = "#7EB8C5"    # brighter muted teal (background when highlighted)
buttonSelectedTextColor = "#0D1418"    # dark text on bright button background

# Animation
spinnerColor = "#00FFFF"               # electric cyan (vivid loading spinner)

# Diff Colors (muted/desaturated for readability - old-tit exact)
diffAddedLineColor = "#5A9C7A"          # muted green (added lines in diff)
diffRemovedLineColor = "#B07070"        # muted red/burgundy (removed lines in diff)

# Console Output (semantic colors for different output types)
outputStdoutColor = "#999999"           # TerminalTextColor - regular command output
outputStderrColor = "#FC704C"           # ErrorTextColor - stderr/error messages
outputStatusColor = "#01C2D2"           # SuccessTextColor - status/success messages
outputWarningColor = "#F2AB53"          # WarningTextColor - warning messages
outputDebugColor = "#33535B"            # DimmedTextColor - debug/info messages
outputInfoColor = "#01C2D2"             # InfoTextColor - TIT-generated info
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

// CreateDefaultThemeIfMissing creates or regenerates the default theme file
// SSOT: DefaultThemeTOML is regenerated on every launch to ensure all colors are current
func CreateDefaultThemeIfMissing() (string, error) {
	configThemeDir := filepath.Join(getConfigDirectory(), "themes")
	configThemeFile := filepath.Join(configThemeDir, "default.toml")

	if err := os.MkdirAll(configThemeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config themes directory: %w", err)
	}

	// Always regenerate from DefaultThemeTOML SSOT to ensure latest colors
	if err := os.WriteFile(configThemeFile, []byte(DefaultThemeTOML), 0644); err != nil {
		return "", fmt.Errorf("failed to write default theme: %w", err)
	}

	return configThemeFile, nil
}

// LoadDefaultTheme loads the default theme
func LoadDefaultTheme() (Theme, error) {
	themeFile := filepath.Join(getConfigDirectory(), "themes", "default.toml")

	if _, err := os.Stat(themeFile); err != nil {
		return Theme{}, fmt.Errorf("theme file not found: %s", themeFile)
	}

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
