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
conflictPaneUnfocusedBorder = "#2C4144"       # littleMermaid (OLD-TIT EXACT - unfocused)
conflictPaneFocusedBorder = "#8CC9D9"         # dolphin (OLD-TIT EXACT - focused)

# Selection - Conflict Resolver specific
conflictSelectionForeground = "#090D12"       # bunker (selection text color)
conflictSelectionBackground = "#7EB8C5"       # brighter muted teal (selection background)

# Pane Headers
conflictPaneTitleColor = "#8CC9D9"            # dolphin (pane title text)

# Status & State
statusClean = "#01C2D2"               # caribbeanBlue
statusModified = "#FC704C"            # preciousPersimmon
timelineSynchronized = "#01C2D2"      # caribbeanBlue
timelineLocalAhead = "#00C8D8"        # blueBikini
timelineLocalBehind = "#F2AB53"       # safflower

# Menu
menuSelectionBackground = "#7EB8C5"    # brighter muted teal (background when highlighted)

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
		
		// Conflict Resolver - Borders
		ConflictPaneUnfocusedBorder string `toml:"conflictPaneUnfocusedBorder"`
		ConflictPaneFocusedBorder   string `toml:"conflictPaneFocusedBorder"`
		
		// Conflict Resolver - Selection
		ConflictSelectionForeground string `toml:"conflictSelectionForeground"`
		ConflictSelectionBackground string `toml:"conflictSelectionBackground"`
		
		// Conflict Resolver - Pane Headers
		ConflictPaneTitleColor string `toml:"conflictPaneTitleColor"`

		// Status Colors
		StatusClean    string `toml:"statusClean"`
		StatusModified string `toml:"statusModified"`

		// Timeline Colors
		TimelineSynchronized string `toml:"timelineSynchronized"`
		TimelineLocalAhead   string `toml:"timelineLocalAhead"`
		TimelineLocalBehind  string `toml:"timelineLocalBehind"`

		// UI Elements
		MenuSelectionBackground string `toml:"menuSelectionBackground"`

		// Console Output Colors
		OutputStdoutColor string `toml:"outputStdoutColor"`
		OutputStderrColor string `toml:"outputStderrColor"`
		OutputStatusColor string `toml:"outputStatusColor"`
		OutputWarningColor string `toml:"outputWarningColor"`
		OutputDebugColor  string `toml:"outputDebugColor"`
		OutputInfoColor   string `toml:"outputInfoColor"`
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
	
	// Conflict Resolver - Borders
	ConflictPaneUnfocusedBorder string
	ConflictPaneFocusedBorder   string
	
	// Conflict Resolver - Selection
	ConflictSelectionForeground string
	ConflictSelectionBackground string
	
	// Conflict Resolver - Pane Headers
	ConflictPaneTitleColor string

	// Status Colors
	StatusClean    string
	StatusModified string

	// Timeline Colors
	TimelineSynchronized string
	TimelineLocalAhead   string
	TimelineLocalBehind  string

	// UI Elements
	MenuSelectionBackground string

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
		
		// Conflict Resolver - Borders
		ConflictPaneUnfocusedBorder: themeDef.Palette.ConflictPaneUnfocusedBorder,
		ConflictPaneFocusedBorder:   themeDef.Palette.ConflictPaneFocusedBorder,
		
		// Conflict Resolver - Selection
		ConflictSelectionForeground: themeDef.Palette.ConflictSelectionForeground,
		ConflictSelectionBackground: themeDef.Palette.ConflictSelectionBackground,
		
		// Conflict Resolver - Pane Headers
		ConflictPaneTitleColor: themeDef.Palette.ConflictPaneTitleColor,

		// Status Colors
		StatusClean:    themeDef.Palette.StatusClean,
		StatusModified: themeDef.Palette.StatusModified,

		// Timeline Colors
		TimelineSynchronized: themeDef.Palette.TimelineSynchronized,
		TimelineLocalAhead:   themeDef.Palette.TimelineLocalAhead,
		TimelineLocalBehind:  themeDef.Palette.TimelineLocalBehind,

		// UI Elements
		MenuSelectionBackground: themeDef.Palette.MenuSelectionBackground,

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

// CreateDefaultThemeIfMissing creates the default theme file on first run
func CreateDefaultThemeIfMissing() (string, error) {
	configThemeDir := filepath.Join(getConfigDirectory(), "themes")
	configThemeFile := filepath.Join(configThemeDir, "default.toml")

	if _, err := os.Stat(configThemeFile); err == nil {
		return configThemeFile, nil
	}

	if err := os.MkdirAll(configThemeDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config themes directory: %w", err)
	}

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
