package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

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
