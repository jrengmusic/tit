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

	for i, row := range brailleArray {
		for _, bc := range row {
			hex := banner.RGBToHex(bc.Color.R, bc.Color.G, bc.Color.B)
			styledChar := lipgloss.NewStyle().
				Foreground(lipgloss.Color(hex)).
				Render(string(bc.Char))
			output.WriteString(styledChar)
		}
		if i < len(brailleArray)-1 {
			output.WriteString("\n")
		}
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

	contentHeight := sizing.TerminalHeight - HeaderHeight - FooterHeight - 1 // -1 for terminal rendering

	// Header: stick to top, exact height
	headerSection := lipgloss.Place(
		sizing.TerminalWidth,
		HeaderHeight,
		lipgloss.Left,
		lipgloss.Top,
		header,
	)

	// Content: fills middle space
	contentSection := lipgloss.NewStyle().
		Width(sizing.TerminalWidth).
		Height(contentHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(content)

	// Footer: stick to bottom, exact height
	footerSection := lipgloss.Place(
		sizing.TerminalWidth,
		FooterHeight,
		lipgloss.Left,
		lipgloss.Top,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.FooterTextColor)).
			Render(footer),
	)

	// Join sections vertically - no wrapping Place
	return lipgloss.JoinVertical(lipgloss.Left, headerSection, contentSection, footerSection)
}

// RenderTextInputFullScreen renders text input centered with footer at bottom
func RenderTextInputFullScreen(
	sizing DynamicSizing,
	theme Theme,
	prompt string,
	state TextInputState,
	footer string,
) string {
	if sizing.CheckIsTooSmall() {
		return renderTooSmallMessage(sizing.TerminalWidth, sizing.TerminalHeight)
	}

	inputContent := RenderTextInput(
		prompt,
		state,
		theme,
		sizing.ContentInnerWidth,
		state.Height,
	)

	contentAreaHeight := sizing.TerminalHeight - FooterHeight
	centeredContent := lipgloss.Place(
		sizing.TerminalWidth,
		contentAreaHeight,
		lipgloss.Center,
		lipgloss.Center,
		inputContent,
	)

	return lipgloss.JoinVertical(lipgloss.Left, centeredContent, footer)
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
