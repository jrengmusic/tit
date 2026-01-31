package app

import (
	"fmt"
	"os"
	"strings"

	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// SetupErrorMsg represents an error during setup
type SetupErrorMsg struct {
	Step  string
	Error string
}

// SetupCompleteMsg represents successful completion of a setup step
type SetupCompleteMsg struct {
	Step string
}

// handleSetupWizardEnter handles ENTER key in setup wizard
func (a *Application) handleSetupWizardEnter(app *Application) (tea.Model, tea.Cmd) {
	switch a.environmentState.SetupWizardStep {
	case SetupStepWelcome:
		a.environmentState.SetupWizardStep = SetupStepPrerequisites
	case SetupStepPrerequisites:
		// Re-check prerequisites
		env := git.DetectGitEnvironment()
		if env == git.MissingGit || env == git.MissingSSH {
			// Still missing, stay on this step
			return a, nil
		}
		// Prerequisites OK, advance to email
		a.environmentState.SetupWizardStep = SetupStepEmail
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
		a.environmentState.SetupEmail = email
		a.inputState.Value = ""
		a.environmentState.SetupWizardStep = SetupStepGenerate
		return a, a.cmdGenerateSSHKey()
	case SetupStepGenerate:
		// Will be handled in Phase 7
		a.environmentState.SetupWizardStep = SetupStepDisplayKey
	case SetupStepDisplayKey:
		a.environmentState.SetupWizardStep = SetupStepComplete
	case SetupStepError:
		// Go back to previous step (Generate step)
		a.environmentState.SetupWizardError = ""
		a.environmentState.SetupWizardStep = SetupStepGenerate
		return a, nil
	case SetupStepComplete:
		// Setup complete - transition to normal TIT operation
		a.environmentState.GitEnvironment = git.Ready

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
	email := a.environmentState.SetupEmail

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
