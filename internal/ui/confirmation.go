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

// Render renders the confirmation dialog
func (c *ConfirmationDialog) Render() string {
	// Always render when in confirmation mode
	config := c.ApplyContext()

	// Button colors from active theme
	yesButtonBg := lipgloss.Color(c.Theme.MenuSelectionBackground)
	yesButtonFg := lipgloss.Color(c.Theme.HighlightTextColor)
	noBg := lipgloss.Color(c.Theme.InlineBackgroundColor)
	noFg := lipgloss.Color(c.Theme.ContentTextColor)

	// Create styles
	dialogStyle := lipgloss.NewStyle().
		Width(c.Width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(c.Theme.BoxBorderColor)).
		Padding(1, 2).
		Align(lipgloss.Center)

	explanationStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Theme.ContentTextColor)).
		Width(c.Width - 4).
		Align(lipgloss.Left)

	// Button styles (no borders, use theme colors)
	// Yes button style depends on selection state
	var yesButtonStyle lipgloss.Style
	if c.SelectedButton == ButtonYes {
		yesButtonStyle = lipgloss.NewStyle().
			Foreground(yesButtonFg).
			Background(yesButtonBg).
			Bold(true).
			Padding(0, 2)
	} else {
		yesButtonStyle = lipgloss.NewStyle().
			Foreground(noFg).
			Background(noBg).
			Bold(true).
			Padding(0, 2)
	}

	// No button style depends on selection state
	var noButtonStyle lipgloss.Style
	if c.SelectedButton == ButtonNo {
		noButtonStyle = lipgloss.NewStyle().
			Foreground(yesButtonFg).
			Background(yesButtonBg).
			Bold(true).
			Padding(0, 2)
	} else {
		noButtonStyle = lipgloss.NewStyle().
			Foreground(noFg).
			Background(noBg).
			Bold(true).
			Padding(0, 2)
	}

	// Build dialog content
	var content strings.Builder

	// Render title with highlighted commit hashes and bold entire title
	styledTitle := c.colorizeCommitHashes(config.Title)
	// Apply bold to entire title
	titleStyle := lipgloss.NewStyle().Bold(true)
	content.WriteString(titleStyle.Render(styledTitle) + "\n")
	content.WriteString("\n")

	// Render explanation with lipgloss auto-wrapping (already applied .Width() to explanationStyle)
	content.WriteString(explanationStyle.Render(config.Explanation) + "\n")

	content.WriteString("\n")
	// Create button layout
	yesButton := yesButtonStyle.Render(config.YesLabel)

	var buttonRow string
	if config.NoLabel == "" {
		// Single button (alert dialog)
		buttonRow = yesButton
	} else {
		// Two buttons (confirmation dialog)
		noButton := noButtonStyle.Render(config.NoLabel)
		buttonRow = lipgloss.JoinHorizontal(lipgloss.Center, yesButton, "  ", noButton)
	}

	// Center the button row
	buttonContainer := lipgloss.NewStyle().
		Align(lipgloss.Center)

	content.WriteString(buttonContainer.Render(buttonRow))

	return dialogStyle.Render(content.String())
}
