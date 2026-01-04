package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// BoxConfig defines a bordered box layout
type BoxConfig struct {
	Content       string // Pre-sized content (should be exact dimensions - 2 for border)
	InnerWidth    int    // Width of content area (excluding border)
	InnerHeight   int    // Height of content area (excluding border)
	BorderColor   string // Theme color for border
	TextColor     string // Theme color for text
	Theme         Theme  // Theme for styling
}

// RenderBox renders a bordered box with content
// Content should already be properly sized to InnerWidth × InnerHeight
// Returns a box that's exactly (InnerWidth + 2) × (InnerHeight + 2) lines
func RenderBox(cfg BoxConfig) string {
	content := EnsureExactDimensions(cfg.Content, cfg.InnerWidth, cfg.InnerHeight)

	style := lipgloss.NewStyle().
		Width(cfg.InnerWidth).
		Foreground(lipgloss.Color(cfg.TextColor)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(cfg.BorderColor))

	return style.Render(content)
}

// StyledContent represents text with styling
type StyledContent struct {
	Text    string
	FgColor string
	Bold    bool
}

// Render returns styled text
func (sc StyledContent) Render() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(sc.FgColor))
	if sc.Bold {
		style = style.Bold(true)
	}
	return style.Render(sc.Text)
}

// Width returns rendered width (accounting for color codes)
func (sc StyledContent) Width() string {
	return sc.Render()
}

// Line represents a single formatted line (left, center, or right aligned)
type Line struct {
	Content   StyledContent
	Alignment string // "left", "center", "right"
	Width     int
}

// Render returns the formatted line at exact width
func (l Line) Render() string {
	rendered := l.Content.Render()
	renderedWidth := lipgloss.Width(rendered)

	if renderedWidth >= l.Width {
		return rendered
	}

	switch l.Alignment {
	case "right":
		return RightAlignLine(rendered, l.Width)
	case "center":
		return CenterAlignLine(rendered, l.Width)
	default: // "left"
		return PadLineToWidth(rendered, l.Width)
	}
}
