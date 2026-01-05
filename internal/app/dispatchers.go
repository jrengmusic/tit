package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/ui"
)

// ActionHandler is a function type for action dispatchers
type ActionHandler func(*Application) tea.Cmd

// dispatchAction routes menu item selections to appropriate handlers
func (a *Application) dispatchAction(actionID string) tea.Cmd {
	actionDispatchers := map[string]ActionHandler{
		"init":               a.dispatchInit,
		"clone":              a.dispatchClone,
		"add_remote":         a.dispatchAddRemote,
		"commit":             a.dispatchCommit,
		"push":               a.dispatchPush,
		"pull_merge":         a.dispatchPullMerge,
		"pull_rebase":        a.dispatchPullRebase,
		"resolve_conflicts":  a.dispatchResolveConflicts,
		"abort_operation":    a.dispatchAbortOperation,
		"continue_operation": a.dispatchContinueOperation,
		"history":            a.dispatchHistory,
	}

	if handler, exists := actionDispatchers[actionID]; exists {
		return handler(a)
	}
	return nil
}

// dispatchInit starts the repository initialization workflow
func (a *Application) dispatchInit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInitializeLocation,
		ResetFields: []string{"init"},
	})
	return nil
}

// dispatchClone starts the clone workflow by asking for repository URL
func (a *Application) dispatchClone(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeCloneURL,
		InputPrompt: "Repository URL:",
		InputAction: "clone_url",
		FooterHint:  "Enter git repository URL (https or git+ssh)",
		ResetFields: []string{"clone"},
	})
	return nil
}

// dispatchAddRemote starts the add remote workflow by asking for URL
func (a *Application) dispatchAddRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "Remote URL:",
		InputAction: "add_remote_url",
		FooterHint:  "Enter git repository URL and press Enter",
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommit starts the commit workflow by asking for message
func (a *Application) dispatchCommit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: "Commit message:",
		InputAction: "commit_message",
		FooterHint:  "Enter message and press Enter",
		ResetFields: []string{},
	})
	return nil
}

// dispatchPush pushes to remote
func (a *Application) dispatchPush(app *Application) tea.Cmd {
	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState = ui.NewConsoleOutState()

	// Execute push asynchronously using operations pattern
	return app.cmdPush()
}

// dispatchPullMerge pulls with merge strategy
func (a *Application) dispatchPullMerge(app *Application) tea.Cmd {
	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState = ui.NewConsoleOutState()

	// Execute pull with merge asynchronously using operations pattern
	return app.cmdPull()
}

// dispatchPullRebase pulls with rebase strategy
func (a *Application) dispatchPullRebase(app *Application) tea.Cmd {
	// Set up async state for console display
	app.asyncOperationActive = true
	app.asyncOperationAborted = false
	app.previousMode = ModeMenu
	app.previousMenuIndex = 0
	app.mode = ModeConsole
	app.consoleState = ui.NewConsoleOutState()

	// Execute pull with rebase asynchronously using operations pattern
	return app.cmdPullRebase()
}

// dispatchResolveConflicts opens conflict resolution UI
func (a *Application) dispatchResolveConflicts(app *Application) tea.Cmd {
	// TODO: Implement
	return nil
}

// dispatchAbortOperation aborts current operation
func (a *Application) dispatchAbortOperation(app *Application) tea.Cmd {
	// TODO: Implement
	return nil
}

// dispatchContinueOperation continues current operation
func (a *Application) dispatchContinueOperation(app *Application) tea.Cmd {
	// TODO: Implement
	return nil
}

// dispatchHistory shows commit history
func (a *Application) dispatchHistory(app *Application) tea.Cmd {
	// TODO: Implement
	return nil
}
