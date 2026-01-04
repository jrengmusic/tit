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
primaryBackground = "#090D12"        # bunker
secondaryBackground = "#1B2A31"      # dark
highlightBackground = "#0D141C"      # corbeau

# Text & Foreground
primaryTextColor = "#4E8C93"          # paradiso
secondaryTextColor = "#8CC9D9"        # dolphin (bright line)
dimmedTextColor = "#33535B"           # mediterranea
accentTextColor = "#01C2D2"           # caribbeanBlue (bright text)
cwdTextColor = "#67DFEF"              # poseidonJr (cyan)
footerTextColor = "#519299"           # lagoon (muted cyan)
mattWhite = "#D1D5DA"                 # off-white
plainGray = "#999999"                 # neutral gray for plain text output

# UI Elements
borderPrimaryColor = "#2C4144"        # littleMermaid (dark line)
borderSecondaryColor = "#8CC9D9"      # dolphin (bright line)

# Status & State
statusClean = "#01C2D2"               # caribbeanBlue
statusModified = "#FC704C"            # preciousPersimmon
timelineSynchronized = "#01C2D2"      # caribbeanBlue
timelineLocalAhead = "#00C8D8"        # blueBikini
timelineLocalBehind = "#F2AB53"       # safflower

# Menu
menuSelectionBackground = "#7EB8C5"    # brighter muted teal (background when highlighted)
`

// ThemeDefinition represents a theme file structure
type ThemeDefinition struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Palette     struct {
		PrimaryBackground   string `toml:"primaryBackground"`
		SecondaryBackground string `toml:"secondaryBackground"`
		HighlightBackground string `toml:"highlightBackground"`
		PrimaryTextColor    string `toml:"primaryTextColor"`
		SecondaryTextColor  string `toml:"secondaryTextColor"`
		DimmedTextColor     string `toml:"dimmedTextColor"`
		AccentTextColor     string `toml:"accentTextColor"`
		CwdTextColor        string `toml:"cwdTextColor"`
		FooterTextColor     string `toml:"footerTextColor"`
		BorderPrimaryColor  string `toml:"borderPrimaryColor"`
		BorderSecondaryColor string `toml:"borderSecondaryColor"`
		StatusClean              string `toml:"statusClean"`
		StatusModified           string `toml:"statusModified"`
		TimelineSynchronized     string `toml:"timelineSynchronized"`
		TimelineLocalAhead       string `toml:"timelineLocalAhead"`
		TimelineLocalBehind      string `toml:"timelineLocalBehind"`
		MenuSelectionBackground  string `toml:"menuSelectionBackground"`
	} `toml:"palette"`
}

// Theme defines all semantic colors from the active theme
type Theme struct {
	PrimaryBackground       string
	SecondaryBackground     string
	HighlightBackground     string
	PrimaryTextColor        string
	SecondaryTextColor      string
	DimmedTextColor         string
	AccentTextColor         string
	CwdTextColor            string
	FooterTextColor         string
	BorderPrimaryColor      string
	BorderSecondaryColor    string
	StatusClean             string
	StatusModified          string
	TimelineSynchronized    string
	TimelineLocalAhead      string
	TimelineLocalBehind     string
	MenuSelectionBackground string
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
		PrimaryBackground:       themeDef.Palette.PrimaryBackground,
		SecondaryBackground:     themeDef.Palette.SecondaryBackground,
		HighlightBackground:     themeDef.Palette.HighlightBackground,
		PrimaryTextColor:        themeDef.Palette.PrimaryTextColor,
		SecondaryTextColor:      themeDef.Palette.SecondaryTextColor,
		DimmedTextColor:         themeDef.Palette.DimmedTextColor,
		AccentTextColor:         themeDef.Palette.AccentTextColor,
		CwdTextColor:            themeDef.Palette.CwdTextColor,
		FooterTextColor:         themeDef.Palette.FooterTextColor,
		BorderPrimaryColor:      themeDef.Palette.BorderPrimaryColor,
		BorderSecondaryColor:    themeDef.Palette.BorderSecondaryColor,
		StatusClean:             themeDef.Palette.StatusClean,
		StatusModified:          themeDef.Palette.StatusModified,
		TimelineSynchronized:    themeDef.Palette.TimelineSynchronized,
		TimelineLocalAhead:      themeDef.Palette.TimelineLocalAhead,
		TimelineLocalBehind:     themeDef.Palette.TimelineLocalBehind,
		MenuSelectionBackground: themeDef.Palette.MenuSelectionBackground,
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
