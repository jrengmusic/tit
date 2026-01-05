package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// ActionHandler is a function type for action dispatchers
type ActionHandler func(*Application) tea.Cmd

// isCwdEmpty checks if current working directory is empty
// Ignores macOS metadata files (.DS_Store)
// Used for smart dispatch in init/clone workflows
func isCwdEmpty() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false // If we can't read dir, assume not empty (safe)
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return false // If we can't read dir, assume not empty (safe)
	}

	// Count entries, ignoring macOS metadata
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if name != ".DS_Store" && name != ".AppleDouble" {
			count++
		}
	}

	return count == 0
}

// dispatchAction routes menu item selections to appropriate handlers
func (a *Application) dispatchAction(actionID string) tea.Cmd {
	actionDispatchers := map[string]ActionHandler{
		"init":               a.dispatchInit,
		"clone":              a.dispatchClone,
		"add_remote":         a.dispatchAddRemote,
		"commit":             a.dispatchCommit,
		"commit_push":        a.dispatchCommitPush,
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
// Smart dispatch: if CWD not empty, skip to subdir initialization
func (a *Application) dispatchInit(app *Application) tea.Cmd {
	// Check if CWD is empty (can only init in empty directories)
	cwdEmpty := isCwdEmpty()

	if !cwdEmpty {
		// CWD not empty: can't init here, must use subdir
		// Auto-dispatch to subdir init (skip location menu)
		return a.cmdInitSubdirectory()
	}

	// CWD is empty: show location choice menu
	app.transitionTo(ModeTransition{
		Mode:        ModeInitializeLocation,
		ResetFields: []string{"init"},
	})
	return nil
}

// dispatchClone starts the clone workflow
// If CWD empty: show location menu first (user chooses mode), then ask URL
// If CWD not empty: ask URL, then clone to subdir directly
func (a *Application) dispatchClone(app *Application) tea.Cmd {
	cwdEmpty := isCwdEmpty()

	if cwdEmpty {
		// CWD empty: show location menu first (user decides: clone here or subdir)
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneLocation,
			ResetFields: []string{"clone"},
		})
	} else {
		// CWD not empty: ask URL, then clone to subdir
		app.cloneMode = "subdir"
		app.transitionTo(ModeTransition{
			Mode:        ModeCloneURL,
			InputPrompt: InputPrompts["clone_url"],
			InputAction: "clone_url",
			FooterHint:  InputHints["clone_url"],
			ResetFields: []string{"clone"},
		})
	}
	return nil
}

// dispatchAddRemote starts the add remote workflow by asking for URL
func (a *Application) dispatchAddRemote(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputPrompts["remote_url"],
		InputAction: "add_remote_url",
		FooterHint:  InputHints["remote_url"],
		ResetFields: []string{},
	})
	return nil
}

// dispatchCommit starts the commit workflow by asking for message
func (a *Application) dispatchCommit(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputPrompts["commit_message"],
		InputAction: "commit_message",
		FooterHint:  InputHints["commit_message"],
		ResetFields: []string{},
	})
	app.inputHeight = 16 // Multiline for commit message
	return nil
}

// dispatchCommitPush starts commit+push workflow by asking for message
func (a *Application) dispatchCommitPush(app *Application) tea.Cmd {
	app.transitionTo(ModeTransition{
		Mode:        ModeInput,
		InputPrompt: InputPrompts["commit_message"],
		InputAction: "commit_push_message",
		FooterHint:  "Enter commit message (will commit and push)",
		ResetFields: []string{},
	})
	app.inputHeight = 16 // Multiline for commit message
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
	app.consoleState.Reset()

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
	app.consoleState.Reset()

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
	app.consoleState.Reset()

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
