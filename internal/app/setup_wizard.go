package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
		// Validate email input
		email := strings.TrimSpace(a.inputState.Value)
		if email == "" {
			// Empty email, stay on this step
			return a, nil
		}

		// Simple email validation
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			// Invalid email format, stay on this step
			return a, nil
		}

		// Store email and advance to generate step
		a.setupEmail = email
		a.inputState.Value = ""
		a.setupWizardStep = SetupStepGenerate
		return a, a.cmdGenerateSSHKey()
	case SetupStepGenerate:
		// Will be handled in Phase 7
		a.setupWizardStep = SetupStepDisplayKey
	case SetupStepDisplayKey:
		a.setupWizardStep = SetupStepComplete
	case SetupStepError:
		// Go back to previous step (Generate step)
		a.setupWizardError = ""
		a.setupWizardStep = SetupStepGenerate
		return a, nil
	case SetupStepComplete:
		// Setup complete - transition to normal TIT operation
		a.gitEnvironment = git.Ready

		// Try to find and cd into git repository (same as normal NewApplication)
		isRepo, repoPath := git.IsInitializedRepo()
		if !isRepo {
			// Check parent directories
			isRepo, repoPath = git.HasParentRepo()
		}

		if isRepo && repoPath != "" {
			// Found a repo, cd into it and detect state
			if err := os.Chdir(repoPath); err != nil {
				// cannot cd into repo - set NotRepo state
				a.gitState = &git.State{Operation: git.NotRepo}
			} else {
				// Successfully cded into repo, detect state
				if err := a.reloadGitState(); err != nil {
					a.gitState = &git.State{Operation: git.NotRepo}
				}
			}
		} else {
			// No repo found
			a.gitState = &git.State{Operation: git.NotRepo}
		}

		a.mode = ModeMenu
		a.menuItems = a.GenerateMenu()
		return a, a.startAutoUpdate()
	}
	return a, nil
}

// cmdGenerateSSHKey generates SSH key and configures SSH
func (a *Application) cmdGenerateSSHKey() tea.Cmd {
	email := a.setupEmail

	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()

		// Generate SSH key
		buffer.Append("Generating SSH key...", ui.TypeStatus)
		if err := git.GenerateSSHKey(email); err != nil {
			buffer.Append(fmt.Sprintf("✗ Failed to generate SSH key: %s", err.Error()), ui.TypeStderr)
			return SetupErrorMsg{Step: "keygen", Error: err.Error()}
		}
		buffer.Append("✓ Created ~/.ssh/TIT_id_rsa", ui.TypeStdout)

		// Add key to SSH agent
		buffer.Append("Adding to SSH agent...", ui.TypeStatus)
		if err := git.AddKeyToAgent(); err != nil {
			buffer.Append("⚠ Could not add to agent (add manually)", ui.TypeWarning)
		} else {
			buffer.Append("✓ Added to SSH agent", ui.TypeStdout)
		}

		// Configure SSH
		buffer.Append("Configuring ~/.ssh/config...", ui.TypeStatus)
		if err := git.WriteSSHConfig(); err != nil {
			buffer.Append("⚠ Could not write config (configure manually)", ui.TypeWarning)
		} else {
			buffer.Append("✓ Configured ~/.ssh/config", ui.TypeStdout)
		}

		return SetupCompleteMsg{Step: "generate"}
	}
}

// SetupErrorMsg represents an error during setup
type SetupErrorMsg struct {
	Step  string
	Error string
}

// SetupCompleteMsg represents successful completion of a setup step
type SetupCompleteMsg struct {
	Step string
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
	case SetupStepError:
		content = a.renderSetupError()
	default:
		content = fmt.Sprintf("Unknown setup step: %d", a.setupWizardStep)
	}

	// Center content in the content area
	style := lipgloss.NewStyle().
		Width(a.sizing.ContentInnerWidth).
		Height(a.sizing.ContentHeight).
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
	if !a.setupKeyCopied {
		clipboard.WriteAll(pubKey)
		a.setupKeyCopied = true
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
		Render(a.setupWizardError)

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
