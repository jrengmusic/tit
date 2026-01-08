package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarConfig holds the configuration for building a status bar
type StatusBarConfig struct {
	Parts      []string // The status bar parts to join
	Width      int      // Terminal width for centering
	Centered   bool     // Whether to center the status bar
	Theme      *Theme   // Theme for colors
}

// BuildStatusBar constructs a status bar with consistent styling across the app
// Handles joining parts, centering, and width management
func BuildStatusBar(config StatusBarConfig) string {
	if config.Theme == nil {
		return ""
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
