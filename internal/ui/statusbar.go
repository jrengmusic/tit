package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarConfig holds the configuration for building a status bar
type StatusBarConfig struct {
	Parts           []string // The status bar parts to join
	Width           int      // Terminal width for centering
	Centered        bool     // Whether to center the status bar
	Theme           *Theme   // Theme for colors
	OverrideMessage string   // Optional override message (e.g., Ctrl+C timeout)
}

// BuildStatusBar constructs a status bar with consistent styling across the app
// Handles joining parts, centering, and width management
// If OverrideMessage is set, it replaces the entire status bar
func BuildStatusBar(config StatusBarConfig) string {
	if config.Theme == nil {
		return ""
	}

	// If override message is set, show it instead of normal status bar
	// Use footer text color from SSOT (same as footer hint styling)
	if config.OverrideMessage != "" {
		overrideStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(config.Theme.FooterTextColor))

		styledMessage := overrideStyle.Render(config.OverrideMessage)

		if config.Centered {
			statusStyle := lipgloss.NewStyle().
				Width(config.Width).
				Align(lipgloss.Center)
			return statusStyle.Render(styledMessage)
		}
		return styledMessage
	}

	// Create style objects for consistent rendering
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.DimmedTextColor))

	// Join parts with separator
	statusText := strings.Join(config.Parts, sepStyle.Render("  │  "))

	// Apply centering if requested
	if config.Centered {
		// Use lipgloss for centering - it handles ANSI codes correctly
		statusStyle := lipgloss.NewStyle().
			Width(config.Width).
			Align(lipgloss.Center)
		return statusStyle.Render(statusText)
	}

	// Return as-is if not centered (left-aligned)
	return statusText
}

// StatusBarStyles provides commonly used style objects for status bars
// CONSOLIDATION: Extracted from multiple buildStatusBar functions to prevent duplication
type StatusBarStyles struct {
	shortcutStyle lipgloss.Style
	descStyle     lipgloss.Style
	visualStyle   lipgloss.Style
}

// NewStatusBarStyles creates a new StatusBarStyles instance using theme colors
func NewStatusBarStyles(theme *Theme) StatusBarStyles {
	return StatusBarStyles{
		shortcutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.AccentTextColor)).
			Bold(true),
		descStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ContentTextColor)),
		visualStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.MainBackgroundColor)).
			Background(lipgloss.Color(theme.AccentTextColor)).
			Bold(true),
	}
}
