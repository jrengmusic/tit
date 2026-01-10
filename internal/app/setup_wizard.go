package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tit/internal/git"
	"tit/internal/ui"
)

// handleSetupWizardEnter handles ENTER key in setup wizard
func (a *Application) handleSetupWizardEnter(app *Application) (tea.Model, tea.Cmd) {
	switch a.setupWizardStep {
	case SetupStepWelcome:
		a.setupWizardStep = SetupStepPrerequisites
	case SetupStepPrerequisites:
		// Re-check prerequisites
		env := git.DetectGitEnvironment()
		if env == git.MissingGit || env == git.MissingSSH {
			// Still missing, stay on this step
			return a, nil
		}
		// Prerequisites OK, advance to email
		a.setupWizardStep = SetupStepEmail
	case SetupStepEmail:
		// Will be handled in Phase 6
		a.setupWizardStep = SetupStepGenerate
	case SetupStepGenerate:
		// Will be handled in Phase 7
		a.setupWizardStep = SetupStepDisplayKey
	case SetupStepDisplayKey:
		a.setupWizardStep = SetupStepComplete
	case SetupStepComplete:
		// Setup complete - transition to normal TIT
		// For now, just quit (will be implemented properly later)
		return a, tea.Quit
	}
	return a, nil
}

// renderSetupWizard renders the current setup wizard step
func (a *Application) renderSetupWizard() string {
	// Build content based on current step
	var content string

	switch a.setupWizardStep {
	case SetupStepWelcome:
		content = a.renderSetupWelcome()
	case SetupStepPrerequisites:
		content = a.renderSetupPrerequisites()
	case SetupStepEmail:
		content = a.renderSetupEmail()
	case SetupStepGenerate:
		content = a.renderSetupGenerate()
	case SetupStepDisplayKey:
		content = a.renderSetupDisplayKey()
	case SetupStepComplete:
		content = a.renderSetupComplete()
	default:
		content = fmt.Sprintf("Unknown setup step: %d", a.setupWizardStep)
	}

	// Center content in the content area
	style := lipgloss.NewStyle().
		Width(ui.ContentInnerWidth).
		Height(ui.ContentHeight).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content)
}

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

// renderButton renders a button with optional selected state
func renderButton(label string, selected bool, theme ui.Theme) string {
	style := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	if selected {
		style = style.
			Foreground(lipgloss.Color(theme.HighlightTextColor)).
			Background(lipgloss.Color(theme.SelectionBackgroundColor))
	} else {
		style = style.
			Foreground(lipgloss.Color(theme.DimmedTextColor))
	}

	return style.Render(label)
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

// renderSetupEmail renders the email input step (placeholder)
func (a *Application) renderSetupEmail() string {
	return "Step 3: Email input (placeholder)"
}

// renderSetupGenerate renders the key generation step (placeholder)
func (a *Application) renderSetupGenerate() string {
	return "Step 4: Generating SSH key (placeholder)"
}

// renderSetupDisplayKey renders the public key display step (placeholder)
func (a *Application) renderSetupDisplayKey() string {
	return "Step 5: Display public key (placeholder)"
}

// renderSetupComplete renders the completion step (placeholder)
func (a *Application) renderSetupComplete() string {
	return "Step 6: Setup complete (placeholder)"
}
