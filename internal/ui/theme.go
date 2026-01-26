package ui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// HSLColor represents a color in HSL space
type HSLColor struct {
	H, S, L float64
}

// hslToHex converts HSL to hex color string
func hslToHex(h, s, l float64) string {
	// Normalize hue to 0-360
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}

	// Convert HSL to RGB
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	var r, g, b float64

	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	// Add m and convert to 0-255 range
	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	// Clamp values
	if r < 0 {
		r = 0
	}
	if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	}
	if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	}
	if b > 255 {
		b = 255
	}

	return fmt.Sprintf("#%02X%02X%02X", int(r), int(g), int(b))
}

// hexToHSL converts hex color to HSL
func hexToHSL(hex string) (float64, float64, float64) {
	// Remove # if present
	hex = strings.TrimPrefix(hex, "#")

	// Parse RGB
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)

	// Normalize to 0-1
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	h, s, l := 0.0, 0.0, (max+min)/2.0

	if max == min {
		h, s = 0.0, 0.0 // achromatic
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / d
			if gf < bf {
				h += 6
			}
		case gf:
			h = (bf-rf)/d + 2
		case bf:
			h = (rf-gf)/d + 4
		}
		h /= 6
	}

	return h * 360, s, l
}

// adjustColorHue shifts a hex color by the given hue degrees and adjusts lightness
func adjustColorHue(hex string, hueShift float64, lightnessMultiplier float64) string {
	h, s, l := hexToHSL(hex)
	h += hueShift
	l *= lightnessMultiplier
	if l > 1.0 {
		l = 1.0
	}
	if l < 0.0 {
		l = 0.0
	}
	return hslToHex(h, s, l)
}

// SeasonalTheme defines a seasonal color variation
type SeasonalTheme struct {
	Name        string
	Description string
	HueShift    float64 // degrees to shift hue
	Lightness   float64 // lightness multiplier (0.8-1.0)
	Saturation  float64 // saturation multiplier (0.8-1.2)
}

// GetSeasonalThemes returns the 4 seasonal theme definitions
func GetSeasonalThemes() []SeasonalTheme {
	return []SeasonalTheme{
		{
			Name:        "spring",
			Description: "Fresh spring greens with vibrant energy",
			HueShift:    60,   // Green hues
			Lightness:   0.95, // Bright and fresh
			Saturation:  1.1,  // More vibrant
		},
		{
			Name:        "summer",
			Description: "Warm summer blues and bright sunshine",
			HueShift:    30,  // Blue-cyan hues
			Lightness:   1.0, // Full brightness
			Saturation:  1.2, // Most saturated
		},
		{
			Name:        "autumn",
			Description: "Rich autumn oranges and warm earth tones",
			HueShift:    -60,  // Orange-red hues
			Lightness:   0.85, // Warmer, less bright
			Saturation:  1.0,  // Natural saturation
		},
		{
			Name:        "winter",
			Description: "Cool winter purples with subtle elegance",
			HueShift:    120, // Purple-magenta hues
			Lightness:   0.8, // Dimmer for winter mood
			Saturation:  0.9, // Slightly muted
		},
	}
}

// generateSeasonalTheme creates a theme variant from the base GFX theme
func generateSeasonalTheme(baseTheme string, seasonal SeasonalTheme) string {
	lines := strings.Split(baseTheme, "\n")
	result := make([]string, 0, len(lines))

	// Update name and description
	for i, line := range lines {
		if strings.HasPrefix(line, "name = ") {
			result = append(result, fmt.Sprintf(`name = "%s"`, strings.Title(seasonal.Name)))
		} else if strings.HasPrefix(line, "description = ") {
			result = append(result, fmt.Sprintf(`description = "%s"`, seasonal.Description))
		} else if strings.Contains(line, " = \"#") && strings.Contains(line, "\"") {
			// This is a color line - extract and transform the hex color
			parts := strings.Split(line, " = \"")
			if len(parts) == 2 {
				colorPart := strings.Split(parts[1], "\"")[0]
				if strings.HasPrefix(colorPart, "#") && len(colorPart) == 7 {
					// Transform the color
					h, s, l := hexToHSL(colorPart)
					h += seasonal.HueShift
					s *= seasonal.Saturation
					l *= seasonal.Lightness

					// Clamp values
					if s > 1.0 {
						s = 1.0
					}
					if s < 0.0 {
						s = 0.0
					}
					if l > 1.0 {
						l = 1.0
					}
					if l < 0.0 {
						l = 0.0
					}

					newColor := hslToHex(h, s, l)
					newLine := strings.Replace(line, colorPart, newColor, 1)
					result = append(result, newLine)
				} else {
					result = append(result, line)
				}
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}

		// Skip the rest if we're at the end
		if i >= len(lines)-1 {
			break
		}
	}

	return strings.Join(result, "\n")
}

// EnsureFiveThemesExist creates/regenerates all 5 themes at startup
func EnsureFiveThemesExist() error {
	configThemeDir := filepath.Join(getConfigDirectory(), "themes")
	if err := os.MkdirAll(configThemeDir, 0755); err != nil {
		return fmt.Errorf("failed to create themes directory: %w", err)
	}

	// Always regenerate gfx.toml from SSOT
	gfxPath := filepath.Join(configThemeDir, "gfx.toml")
	if err := os.WriteFile(gfxPath, []byte(DefaultThemeTOML), 0644); err != nil {
		return fmt.Errorf("failed to write gfx theme: %w", err)
	}

	// Generate seasonal themes
	seasonalThemes := GetSeasonalThemes()
	for _, seasonal := range seasonalThemes {
		themeContent := generateSeasonalTheme(DefaultThemeTOML, seasonal)
		themePath := filepath.Join(configThemeDir, seasonal.Name+".toml")
		if err := os.WriteFile(themePath, []byte(themeContent), 0644); err != nil {
			return fmt.Errorf("failed to write %s theme: %w", seasonal.Name, err)
		}
	}

	return nil
}

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

# Confirmation Dialog
confirmationDialogBackground = "#112130"        # trappedDarkness (dialog box background)
	conflictPaneUnfocusedBorder = "#2C4144"
	conflictPaneFocusedBorder = "#8CC9D9"

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

	# Diff Colors (muted/desaturated for readability)
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
// SSOT: Always regenerates from DefaultThemeTOML to ensure all colors are current
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
