package app

import (
	"fmt"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
)

// renderSetupWelcome renders the welcome step
func (a *Application) renderSetupWelcome() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(a.theme.LabelTextColor)).
		Render("Welcome to TIT")

	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.ContentTextColor)).
		Render("We'll set up SSH authentication for git.\nThis only needs to be done once per machine.")

	button := renderButton("Continue", true, a.theme)

	return lipgloss.JoinVertical(lipgloss.Center, title, "", body, "", "", button)
}

// renderSetupPrerequisites renders the prerequisites check step
func (a *Application) renderSetupPrerequisites() string {
	gitOK := git.DetectGitEnvironment() != git.MissingGit
	sshOK := git.DetectGitEnvironment() != git.MissingSSH

	// Status indicators
	gitStatus := "✗ Git not found"
	if gitOK {
		gitStatus = "✓ Git found"
	}
	sshStatus := "✗ SSH not found"
	if sshOK {
		sshStatus = "✓ SSH found"
	}

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(a.theme.ContentTextColor))

	content := statusStyle.Render(gitStatus + "\n" + sshStatus)

	// Show install instructions if needed
	if !gitOK || !sshOK {
		instructions := "\n\nInstall missing tools:\n" +
			"  macOS:   brew install git openssh\n" +
			"  Linux:   sudo apt install git openssh-client\n" +
			"  Windows: git-scm.com/download"
		content += lipgloss.NewStyle().Foreground(lipgloss.Color(a.theme.DimmedTextColor)).Render(instructions)
	}

	button := "\n\n" + renderButton("Re-check", true, a.theme)

	return content + button
}

// renderSetupEmail renders the email input step
func (a *Application) renderSetupEmail() string {
	textInputState := ui.TextInputState{
		Value:     a.inputState.Value,
		CursorPos: a.inputState.CursorPosition,
		Height:    4,
	}

	// Render input component
	inputContent := ui.RenderTextInput(
		"Email (for SSH key comment):",
		textInputState,
		a.theme,
		a.sizing.ContentInnerWidth,
		textInputState.Height,
	)

	// Add continue button below input
	button := renderButton("Continue", true, a.theme)
	combined := lipgloss.JoinVertical(lipgloss.Center, inputContent, "", button)

	// Center in content area
	contentAreaHeight := a.sizing.TerminalHeight - ui.FooterHeight
	centeredContent := lipgloss.Place(
		a.sizing.TerminalWidth,
		contentAreaHeight,
		lipgloss.Center,
		lipgloss.Center,
		combined,
	)

	footer := a.GetFooterContent()
	return lipgloss.JoinVertical(lipgloss.Left, centeredContent, footer)
}

// renderSetupGenerate renders the key generation step
func (a *Application) renderSetupGenerate() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(a.theme.LabelTextColor)).
		Render("Generating SSH Key")

	// Get console output
	lines := ui.GetBuffer().GetAllLines()

	// Join lines with newlines
	var output string
	for _, line := range lines {
		output += line.Text + "\n"
	}

	// If no output yet, show status
	if output == "" {
		output = "Starting SSH key generation..."
	}

	outputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.ContentTextColor)).
		Width(a.sizing.ContentInnerWidth - 4).
		Height(a.sizing.ContentHeight - 10)

	content := outputStyle.Render(output)

	// Show continue button when complete
	button := ""
	if strings.Contains(output, "✓ Configured") || strings.Contains(output, "✗ Failed") {
		button = "\n" + renderButton("Continue", true, a.theme)
	}

	return lipgloss.JoinVertical(lipgloss.Center, title, "", content+button)
}

// renderSetupDisplayKey renders the public key display step
func (a *Application) renderSetupDisplayKey() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(a.theme.LabelTextColor)).
		Render("SSH Key Generated")

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.ContentTextColor)).
		Render("Your public key has been copied to clipboard")

	// Get public key
	pubKey, err := git.GetPublicKey()
	if err != nil {
		return lipgloss.JoinVertical(lipgloss.Center,
			title, "",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FC704C")).Render(
				fmt.Sprintf("Error reading public key: %s", err.Error())),
		)
	}

	// Copy to clipboard if not already done
	if !a.environmentState.SetupKeyCopied {
		clipboard.WriteAll(pubKey)
		a.environmentState.SetupKeyCopied = true
	}

	// Render key in a box (shorter to make room for button)
	// Use word wrap to handle long SSH keys
	keyBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(a.theme.BoxBorderColor)).
		Width(a.sizing.ContentInnerWidth-4).
		Height(8).
		Padding(1, 2)

	// Wrap the public key to fit in the box
	wrappedKey := lipgloss.NewStyle().
		Width(a.sizing.ContentInnerWidth - 8). // Account for border and padding
		Render(pubKey)

	keyContent := keyBox.Render(wrappedKey)

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.DimmedTextColor)).
		Render(
			"Add this key to your git provider:\n" +
				"  • GitHub:    github.com/settings/ssh/new\n" +
				"  • GitLab:    gitlab.com/-/user_settings/ssh_keys\n" +
				"  • Bitbucket: bitbucket.org/account/settings/ssh-keys")

	button := renderButton("Continue", true, a.theme)

	return lipgloss.JoinVertical(lipgloss.Center,
		title, "",
		subtitle, "",
		keyContent, "",
		instructions, "",
		button)
}

// renderSetupError renders the error step
func (a *Application) renderSetupError() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FC704C")).
		Render("✗ Setup Error")

	errorMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.ContentTextColor)).
		Width(a.sizing.ContentInnerWidth - 4).
		Render(a.environmentState.SetupWizardError)

	body := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.DimmedTextColor)).
		Render("Press ESC to go back and try again.")

	button := renderButton("Go Back", true, a.theme)

	return lipgloss.JoinVertical(lipgloss.Center, title, "", errorMsg, "", body, "", button)
}

// renderSetupComplete renders the completion step
func (a *Application) renderSetupComplete() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(a.theme.LabelTextColor)).
		Render("✓ Setup Complete!")

	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.ContentTextColor)).
		Render("TIT is ready. You can now init or clone repositories.")

	button := renderButton("Continue", true, a.theme)

	return lipgloss.JoinVertical(lipgloss.Center, title, "", message, "", "", button)
}

// renderButton renders a button with optional selected state
// Matches confirmation dialog button style: dark text on bright background when selected
func renderButton(label string, selected bool, theme ui.Theme) string {
	// ALL CAPS for button labels
	label = strings.ToUpper(label)

	style := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	if selected {
		// Selected: dark text on bright teal background (same as confirmation dialog)
		style = style.
			Foreground(lipgloss.Color(theme.ButtonSelectedTextColor)).
			Background(lipgloss.Color(theme.MenuSelectionBackground))
	} else {
		// Unselected: normal text on subtle background
		style = style.
			Foreground(lipgloss.Color(theme.ContentTextColor)).
			Background(lipgloss.Color(theme.InlineBackgroundColor))
	}

	return style.Render(label)
}
