package ui

import (
	"embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"tit/internal/banner"
)

//go:embed assets/tit-logo.svg
var logoFS embed.FS

// RenderBanner renders top banner with braille cherry logo (height = BannerHeight)
func RenderBanner(s Sizing) string {
	// Load SVG logo from embedded assets
	logoData, err := logoFS.ReadFile("assets/tit-logo.svg")
	if err != nil {
		return strings.Repeat(" ", InterfaceWidth) + "\n" +
			strings.Repeat(" ", InterfaceWidth) + "\n" +
			strings.Repeat(" ", InterfaceWidth)
	}

	svgString := string(logoData)

	// Calculate canvas size: each braille char is 2px wide, 4px tall
	// BannerHeight is in terminal lines, so multiply by 4 for pixel height
	canvasWidth := InterfaceWidth * 2
	canvasHeight := BannerHeight * 4

	// Convert SVG to braille array
	brailleArray := banner.SvgToBrailleArray(svgString, canvasWidth, canvasHeight)

	var output strings.Builder

	// Render each row of braille characters
	for _, row := range brailleArray {
		for _, bc := range row {
			// Convert RGB to hex color and apply
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

// RenderHeader renders header section with border (height = HeaderHeight)
// canonBranch and workingBranch are displayed right-aligned at top and bottom
func RenderHeader(s Sizing, theme Theme, canonBranch string, workingBranch string) string {
	// Header layout: 2 lines of branches (right-aligned, all caps, bold)
	
	canonLine := Line{
		Content: StyledContent{
			Text:    strings.ToUpper(canonBranch),
			FgColor: theme.LabelTextColor,
			Bold:    true,
		},
		Alignment: "right",
		Width:     ContentInnerWidth,
	}

	workingLine := Line{
		Content: StyledContent{
			Text:    strings.ToUpper(workingBranch),
			FgColor: theme.LabelTextColor,
			Bold:    true,
		},
		Alignment: "right",
		Width:     ContentInnerWidth,
	}

	// Combine branches
	content := canonLine.Render() + "\n" + workingLine.Render()

	// Pad to fill HeaderHeight-2 (border adds 2 for total)
	padded := PadTextToHeight(content, HeaderHeight-2)

	// Render with border
	return RenderBox(BoxConfig{
		Content:     padded,
		InnerWidth:  ContentInnerWidth,
		InnerHeight: HeaderHeight - 2,
		BorderColor: theme.BoxBorderColor,
		TextColor:   theme.LabelTextColor,
		Theme:       theme,
	})
}

// RenderContent renders main content area with border (height = ContentHeight)
func RenderContent(s Sizing, text string, theme Theme) string {
	padded := PadTextToHeight(text, ContentHeight-2)

	return RenderBox(BoxConfig{
		Content:     padded,
		InnerWidth:  ContentInnerWidth,
		InnerHeight: ContentHeight - 2,
		BorderColor: theme.BoxBorderColor,
		TextColor:   theme.ContentTextColor,
		Theme:       theme,
	})
}

// RenderFooter renders footer section without border (height = FooterHeight)
func RenderFooter(s Sizing, theme Theme, app interface{ GetFooterHint() string }) string {
	content := ""
	
	// Show hint if active
	if hint := app.GetFooterHint(); hint != "" {
		content = hint
	}

	style := lipgloss.NewStyle().
		Width(InterfaceWidth).
		Height(FooterHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Foreground(lipgloss.Color(theme.FooterTextColor))

	return style.Render(content)
}

// RenderLayout combines all 4 sections into centered view (horizontally and vertically)
func RenderLayout(s Sizing, contentText string, termWidth int, termHeight int, theme Theme, canonBranch string, workingBranch string, app interface{ GetFooterHint() string }) string {
	banner := RenderBanner(s)
	header := RenderHeader(s, theme, canonBranch, workingBranch)
	content := RenderContent(s, contentText, theme)
	footer := RenderFooter(s, theme, app)

	// Stack sections vertically
	stack := lipgloss.JoinVertical(
		lipgloss.Top,
		banner,
		header,
		content,
		footer,
	)

	// Use lipgloss to center within terminal
	centeredStyle := lipgloss.NewStyle().
		Width(termWidth).
		Height(termHeight).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center)

	return centeredStyle.Render(stack)
}

