// Package ui provides user interface components for the TIT application
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConfirmationConfig defines the configuration for a confirmation dialog
type ConfirmationConfig struct {
	Title       string
	Explanation string
	YesLabel    string
	NoLabel     string
	ActionID    string
}

// ButtonSelection represents which button is currently selected
type ButtonSelection string

const (
	ButtonYes ButtonSelection = "yes"
	ButtonNo  ButtonSelection = "no"
)

// ConfirmationDialog represents a confirmation dialog state
type ConfirmationDialog struct {
	Config         ConfirmationConfig
	Width          int
	Theme          *Theme
	Active         bool
	Context        map[string]string
	SelectedButton ButtonSelection
}

// NewConfirmationDialog creates a new confirmation dialog
func NewConfirmationDialog(config ConfirmationConfig, width int, theme *Theme) *ConfirmationDialog {
	return &ConfirmationDialog{
		Config:         config,
		Width:          width,
		Theme:          theme,
		Active:         false,
		Context:        make(map[string]string),
		SelectedButton: ButtonYes, // Default to Yes button
	}
}

// SetContext sets the context for placeholder substitution
func (c *ConfirmationDialog) SetContext(context map[string]string) {
	c.Context = context
}

// SelectYes selects the Yes button
func (c *ConfirmationDialog) SelectYes() {
	c.SelectedButton = ButtonYes
}

// SelectNo selects the No button
func (c *ConfirmationDialog) SelectNo() {
	c.SelectedButton = ButtonNo
}

// ToggleSelection switches between Yes and No buttons
func (c *ConfirmationDialog) ToggleSelection() {
	if c.SelectedButton == ButtonYes {
		c.SelectNo()
	} else {
		c.SelectYes()
	}
}

// GetSelectedButton returns the currently selected button
func (c *ConfirmationDialog) GetSelectedButton() ButtonSelection {
	return c.SelectedButton
}

// ApplyContext applies context placeholders to the config
func (c *ConfirmationDialog) ApplyContext() ConfirmationConfig {
	config := c.Config

	// Apply context to title
	if c.Context != nil {
		config.Title = applyPlaceholders(config.Title, c.Context)
		config.Explanation = applyPlaceholders(config.Explanation, c.Context)
	}

	return config
}

// applyPlaceholders replaces {placeholder} with context values
func applyPlaceholders(text string, context map[string]string) string {
	result := text
	for key, value := range context {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// colorizeCommitHashes finds commit hashes in square brackets and colors them
// Example: "Apply changes from [289794ed]" → colored hash
// Uses AccentTextColor for highlighting
func (c *ConfirmationDialog) colorizeCommitHashes(text string) string {
	hashStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Theme.AccentTextColor)).
		Bold(true)

	// Find patterns like [abc123def] and color the hash inside
	var output strings.Builder
	inBracket := false
	var bracket strings.Builder

	for _, ch := range text {
		if ch == '[' {
			inBracket = true
			bracket.Reset()
		} else if ch == ']' && inBracket {
			// Found closing bracket - color the hash inside
			hash := bracket.String()
			output.WriteString(hashStyle.Render(hash))
			inBracket = false
		} else if inBracket {
			bracket.WriteRune(ch)
		} else {
			output.WriteRune(ch)
		}
	}

	return output.String()
}

// Render renders the confirmation dialog centered within the given height using DynamicSizing
func (c *ConfirmationDialog) Render(height int) string {
	// Always render when in confirmation mode
	config := c.ApplyContext()

	// Button colors from active theme
	// Selected button: MenuSelectionBackground + ButtonSelectedTextColor (dark on bright)
	// Unselected button: InlineBackgroundColor + ContentTextColor
	selectedBg := lipgloss.Color(c.Theme.MenuSelectionBackground)
	selectedFg := lipgloss.Color(c.Theme.ButtonSelectedTextColor)
	unselectedBg := lipgloss.Color(c.Theme.InlineBackgroundColor)
	unselectedFg := lipgloss.Color(c.Theme.ContentTextColor)

	// Dialog width uses c.Width (passed from DynamicSizing.ContentInnerWidth)
	dialogWidth := c.Width - 10 // Leave padding for visual centering

	// Create styles
	dialogStyle := lipgloss.NewStyle().
		Width(dialogWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(c.Theme.BoxBorderColor)).
		Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground)).
		Padding(1, 2).
		Align(lipgloss.Center)

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Theme.ContentTextColor)).
		Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground)).
		Width(dialogWidth - 4). // Account for dialog padding
		Align(lipgloss.Left)

	// Button styles (no borders, use theme colors)
	// Yes button style depends on selection state
	var yesButtonStyle lipgloss.Style
	if c.SelectedButton == ButtonYes {
		yesButtonStyle = lipgloss.NewStyle().
			Foreground(selectedFg).
			Background(selectedBg).
			Bold(true).
			Padding(0, 2)
	} else {
		yesButtonStyle = lipgloss.NewStyle().
			Foreground(unselectedFg).
			Background(unselectedBg).
			Bold(true).
			Padding(0, 2)
	}

	// No button style depends on selection state
	var noButtonStyle lipgloss.Style
	if c.SelectedButton == ButtonNo {
		noButtonStyle = lipgloss.NewStyle().
			Foreground(selectedFg).
			Background(selectedBg).
			Bold(true).
			Padding(0, 2)
	} else {
		noButtonStyle = lipgloss.NewStyle().
			Foreground(unselectedFg).
			Background(unselectedBg).
			Bold(true).
			Padding(0, 2)
	}

	// Build dialog content
	var content strings.Builder

	// Render title with highlighted commit hashes and bold entire title
	styledTitle := c.colorizeCommitHashes(config.Title)
	// Apply bold to entire title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Width(dialogWidth - 4).
		Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground))
	content.WriteString(titleStyle.Render(styledTitle) + "\n")
	content.WriteString("\n")

	// Render explanation with lipgloss auto-wrapping (already applied .Width() to explanationStyle)
	content.WriteString(explanationStyle.Render(config.Explanation) + "\n")

	content.WriteString("\n")
	// Create button layout - ALL CAPS for button labels
	yesButton := yesButtonStyle.Render(strings.ToUpper(config.YesLabel))

	var buttonRow string
	if config.NoLabel == "" {
		// Single button (alert dialog)
		buttonRow = yesButton
	} else {
		// Two buttons (confirmation dialog)
		noButton := noButtonStyle.Render(strings.ToUpper(config.NoLabel))
		// Style the gap with dialog background
		buttonGap := lipgloss.NewStyle().
			Background(lipgloss.Color(c.Theme.ConfirmationDialogBackground)).
			Render("  ")
		buttonRow = lipgloss.JoinHorizontal(lipgloss.Center, yesButton, buttonGap, noButton)
	}

	// Center the button row
	buttonContainer := lipgloss.NewStyle().
		Align(lipgloss.Center)

	content.WriteString(buttonContainer.Render(buttonRow))

	// Build the dialog
	dialog := dialogStyle.Render(content.String())

	// Center the dialog vertically and horizontally within the given height
	// Use lipgloss.Place to center within the content box area
	centeredDialog := lipgloss.Place(
		c.Width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		dialog,
	)

	return centeredDialog
}
