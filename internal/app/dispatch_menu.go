package app

import (
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// dispatchInit starts the repository initialization workflow
func (a *Application) dispatchInit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode: ModeInitializeLocation,
	})
	return nil
}

// dispatchClone starts the clone workflow
func (a *Application) dispatchClone(app *Application) tea.Cmd {
	cwdEmpty := isCwdEmpty()
	if cwdEmpty {
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneLocation,
			ResetFields: []string{"clone"},
		})
	} else {
		app.workflowState.CloneMode = "subdir"
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneURL,
			InputPrompt: InputMessages["clone_url"].Prompt,
			InputAction: "clone_url",
			FooterHint:  InputMessages["clone_url"].Hint,
			ResetFields: []string{"clone"},
		})
	}
	return nil
}

// dispatchAddRemote starts the add remote workflow
func (a *Application) dispatchAddRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["remote_url"].Prompt,
		InputAction: "add_remote_url",
		FooterHint:  InputMessages["remote_url"].Hint,
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommit starts the commit workflow
func (a *Application) dispatchCommit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["commit_message"].Prompt,
		InputAction: "commit_message",
		FooterHint:  InputMessages["commit_message"].Hint,
		InputHeight: app.sizing.TerminalHeight - ui.FooterHeight,
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommitPush starts commit+push workflow
func (a *Application) dispatchCommitPush(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputMessages["commit_message"].Prompt,
		InputAction: "commit_push_message",
		FooterHint:  "Enter commit message (will commit and push)",
		InputHeight: app.sizing.TerminalHeight - ui.FooterHeight,
		ResetFields: []string{},
	})
	return nil
}
