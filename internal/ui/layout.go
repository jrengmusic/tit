package ui

import (
	"embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"tit/internal/banner"
)

//go:embed assets/tit-logo.svg
var logoFS embed.FS

// RenderBannerDynamic renders banner with dynamic dimensions
func RenderBannerDynamic(width, height int) string {
	logoData, err := logoFS.ReadFile("assets/tit-logo.svg")
	if err != nil {
		return strings.Repeat(" ", width) + "\n" +
			strings.Repeat(" ", width) + "\n" +
			strings.Repeat(" ", width)
	}

	svgString := string(logoData)

	canvasWidth := width * 2
	canvasHeight := height * 4

	brailleArray := banner.SvgToBrailleArray(svgString, canvasWidth, canvasHeight)

	var output strings.Builder

	for _, row := range brailleArray {
		for _, bc := range row {
			hex := banner.RGBToHex(bc.Color.R, bc.Color.G, bc.Color.B)
			styledChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(hex)).
				Render(string(bc.Char))
			output.WriteString(styledChar)
		}
		output.WriteString("\n")
	}

	return output.String()
}

// RenderContentDynamic renders content with dynamic sizing
func RenderContentDynamic(sizing DynamicSizing, theme Theme, text string) string {
	innerHeight := sizing.ContentHeight - 2
	padded := PadTextToHeight(text, innerHeight)

	return RenderBox(BoxConfig{
		Content:     padded,
		InnerWidth:  sizing.ContentInnerWidth,
		InnerHeight: innerHeight,
		BorderColor: theme.BoxBorderColor,
		TextColor:   theme.ContentTextColor,
		Theme:       theme,
	})
}

// RenderReactiveLayout combines header/content/footer into full-terminal reactive layout
func RenderReactiveLayout(sizing DynamicSizing, theme Theme, header, content, footer string) string {
	// Too small guard
	if sizing.CheckIsTooSmall() {
		return renderTooSmallMessage(sizing.TerminalWidth, sizing.TerminalHeight)
	}

	contentHeight := sizing.TerminalHeight - HeaderHeight - FooterHeight

	// Header: fixed height, top-aligned
	headerSection := lipgloss.NewStyle().
		Width(sizing.TerminalWidth).
		Height(HeaderHeight).
		AlignVertical(lipgloss.Top).
		Render(header)

	// Content: fills middle space, centered
	contentSection := lipgloss.NewStyle().
		Width(sizing.TerminalWidth).
		Height(contentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)

	// Footer: single line, centered
	footerSection := lipgloss.NewStyle().
		Width(sizing.TerminalWidth).
		Height(FooterHeight).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color(theme.FooterTextColor)).
		Render(footer)

	// Join sections
	combined := lipgloss.JoinVertical(lipgloss.Left, headerSection, contentSection, footerSection)

	// Place in exact terminal dimensions - footer sticks to bottom
	return lipgloss.Place(
		sizing.TerminalWidth,
		sizing.TerminalHeight,
		lipgloss.Left,
		lipgloss.Bottom,
		combined,
	)
}

func renderTooSmallMessage(w, h int) string {
	msg := "Terminal too small.\nResize to >= 70×20."
	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(msg)
}
