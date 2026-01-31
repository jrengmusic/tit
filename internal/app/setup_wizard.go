package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// renderSetupWizard renders the current setup wizard step
func (a *Application) renderSetupWizard() string {
	// Build content based on current step
	var content string

	switch a.environmentState.SetupWizardStep {
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
		content = fmt.Sprintf("Unknown setup step: %d", a.environmentState.SetupWizardStep)
	}

	// Center content in the content area
	style := lipgloss.NewStyle().
		Width(a.sizing.ContentInnerWidth).
		Height(a.sizing.ContentHeight).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content)
}
