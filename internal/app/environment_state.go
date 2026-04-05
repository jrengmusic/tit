package app

import "tit/internal/git"

// EnvironmentState manages git environment detection and setup wizard state.
// This is only relevant before main application loop starts.
type EnvironmentState struct {
	GitEnvironment   git.GitEnvironment // Ready, NeedsSetup, MissingGit, MissingSSH
	SetupWizardStep  SetupWizardStep    // Current step in wizard
	SetupWizardError string             // Error message for SetupStepError
	SetupEmail       string             // Email for SSH key generation
	SetupKeyCopied bool // Public key copied to clipboard
}

// NewEnvironmentState creates a new EnvironmentState with defaults.
func NewEnvironmentState() EnvironmentState {
	return EnvironmentState{
		GitEnvironment:  git.Ready,
		SetupWizardStep: SetupStepWelcome,
	}
}

// IsReady returns true if git environment is ready for operation.
func (e *EnvironmentState) IsReady() bool {
	return e.GitEnvironment == git.Ready
}

// NeedsSetup returns true if setup wizard is required.
func (e *EnvironmentState) NeedsSetup() bool {
	return e.GitEnvironment == git.NeedsSetup
}

// MarkKeyCopied marks the SSH key as copied.
func (e *EnvironmentState) MarkKeyCopied() {
	e.SetupKeyCopied = true
}

// IsKeyCopied returns true if the SSH key has been copied.
func (e *EnvironmentState) IsKeyCopied() bool {
	return e.SetupKeyCopied
}
