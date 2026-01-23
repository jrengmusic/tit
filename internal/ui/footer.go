package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FooterConfig holds the configuration for building a footer
type FooterConfig struct {
	Parts           []string // The footer parts to join
	Width           int      // Terminal width for centering
	Centered        bool     // Whether to center the footer
	Theme           *Theme   // Theme for colors
	OverrideMessage string   // Optional override message (e.g., Ctrl+C timeout)
}

// BuildFooter constructs a footer with consistent styling across the app
// Handles joining parts, centering, and width management
// If OverrideMessage is set, it replaces the entire footer
func BuildFooter(config FooterConfig) string {
	if config.Theme == nil {
		return ""
	}

	// If override message is set, show it instead of normal footer
	// Use footer text color from SSOT (same as footer hint styling)
	if config.OverrideMessage != "" {
		overrideStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(config.Theme.FooterTextColor))

		styledMessage := overrideStyle.Render(config.OverrideMessage)

		if config.Centered {
			footerStyle := lipgloss.NewStyle().
				Width(config.Width).
				Align(lipgloss.Center)
			return footerStyle.Render(styledMessage)
		}
		return styledMessage
	}

	// Create style objects for consistent rendering
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Theme.DimmedTextColor))

	// Join parts with separator
	footerText := strings.Join(config.Parts, sepStyle.Render("  │  "))

	// Apply centering if requested
	if config.Centered {
		// Use lipgloss for centering - it handles ANSI codes correctly
		footerStyle := lipgloss.NewStyle().
			Width(config.Width).
			Align(lipgloss.Center)
		return footerStyle.Render(footerText)
	}

	// Return as-is if not centered (left-aligned)
	return footerText
}

// FooterStyles provides commonly used style objects for footers
// CONSOLIDATION: Extracted from multiple buildFooter functions to prevent duplication
type FooterStyles struct {
	shortcutStyle lipgloss.Style
	descStyle     lipgloss.Style
	visualStyle   lipgloss.Style
	sepStyle      lipgloss.Style
}

// NewFooterStyles creates a new FooterStyles instance using theme colors
func NewFooterStyles(theme *Theme) FooterStyles {
	return FooterStyles{
		shortcutStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.AccentTextColor)).
			Bold(true),
		descStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.ContentTextColor)),
		visualStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.MainBackgroundColor)).
			Background(lipgloss.Color(theme.AccentTextColor)).
			Bold(true),
		sepStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.DimmedTextColor)),
	}
}

// FooterShortcut represents a single keyboard shortcut hint
type FooterShortcut struct {
	Key  string // e.g., "↑↓", "Enter", "Esc"
	Desc string // e.g., "navigate", "select", "back"
}

// RenderFooter renders footer shortcuts with optional right-side content
// If rightContent is provided: shortcuts on left, rightContent on right
// If rightContent is empty: shortcuts centered
func RenderFooter(shortcuts []FooterShortcut, width int, theme *Theme, rightContent string) string {
	if theme == nil {
		return ""
	}

	styles := NewFooterStyles(theme)

	// If rightContent provided, render left shortcuts + right content
	if rightContent != "" {
		// Build styled parts: "Key desc"
		var leftParts []string
		for _, sc := range shortcuts {
			part := styles.shortcutStyle.Render(sc.Key) + styles.descStyle.Render(" "+sc.Desc)
			leftParts = append(leftParts, part)
		}
		leftJoined := strings.Join(leftParts, styles.sepStyle.Render("  │  "))

		// Style right content
		rightStyled := styles.descStyle.Render(rightContent)

		// Calculate spacing
		leftWidth := lipgloss.Width(leftJoined)
		rightWidth := lipgloss.Width(rightStyled)
		padding := width - leftWidth - rightWidth
		if padding < 0 {
			padding = 0
		}

		return leftJoined + strings.Repeat(" ", padding) + rightStyled
	}

	// No rightContent: center shortcuts
	var parts []string
	for _, sc := range shortcuts {
		part := styles.shortcutStyle.Render(sc.Key) + styles.descStyle.Render(" "+sc.Desc)
		parts = append(parts, part)
	}

	sep := styles.sepStyle.Render("  ·  ")
	content := strings.Join(parts, sep)

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(content)
}

// RenderFooterOverride renders override message (e.g., Ctrl+C confirm)
func RenderFooterOverride(message string, width int, theme *Theme) string {
	if message == "" {
		return ""
	}
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.FooterTextColor)).
		Width(width).
		Align(lipgloss.Center)
	return style.Render(message)
}
