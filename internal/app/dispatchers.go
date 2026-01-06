package app

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/ui"
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
		"force_push":         a.dispatchForcePush,
		"pull_merge":         a.dispatchPullMerge,
		"pull_rebase":        a.dispatchPullRebase,
		"replace_local":      a.dispatchReplaceLocal,
		"resolve_conflicts":  a.dispatchResolveConflicts,
		"abort_operation":    a.dispatchAbortOperation,
		"continue_operation": a.dispatchContinueOperation,
		"history":            a.dispatchHistory,
		"test_conflict":      a.dispatchTestConflict, // DEBUG: Test conflict resolver
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

// dispatchForcePush shows confirmation dialog for destructive action (like old-tit)
func (a *Application) dispatchForcePush(app *Application) tea.Cmd {
	// Show confirmation dialog for destructive action
	app.mode = ModeConfirmation
	app.confirmType = "force_push"
	app.confirmContext = map[string]string{}
	
	// Create the confirmation dialog
	config := ui.ConfirmationConfig{
		Title:       "Force Push Confirmation",
		Explanation: "This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
		YesLabel:    "Force push",
		NoLabel:     "Cancel",
		ActionID:    "force_push",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)
	
	return nil
}

// dispatchReplaceLocal shows confirmation dialog for destructive action (like old-tit)  
func (a *Application) dispatchReplaceLocal(app *Application) tea.Cmd {
	// Show confirmation dialog for destructive action  
	app.mode = ModeConfirmation
	app.confirmType = "hard_reset"
	app.confirmContext = map[string]string{}
	
	// Create the confirmation dialog
	config := ui.ConfirmationConfig{
		Title:       "Replace Local Confirmation", 
		Explanation: "This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
		YesLabel:    "Reset to remote",
		NoLabel:     "Cancel",
		ActionID:    "hard_reset",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)
	
	return nil
}

// dispatchHistory shows commit history
func (a *Application) dispatchHistory(app *Application) tea.Cmd {
	// TODO: Implement
	return nil
}

// dispatchTestConflict enters conflict resolve mode with mock data (DEBUG ONLY)
func (a *Application) dispatchTestConflict(app *Application) tea.Cmd {
	// Create mock conflict state with 2 columns (LOCAL vs REMOTE)
	app.conflictResolveState = &ConflictResolveState{
		Operation:  "test_conflict",
		CommitHash: "abc1234",
		IsRebase:   false,
		
		// Mock files in conflict
		Files: []ui.ConflictFileGeneric{
			{
				Path:   "src/main.go",
				Chosen: 0, // Default to LOCAL
				Versions: []string{
					"package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"LOCAL version\")\n}\n",
					"package main\n\nimport \"log\"\n\nfunc main() {\n\tlog.Println(\"REMOTE version\")\n}\n",
				},
			},
			{
				Path:   "README.md",
				Chosen: 1, // Default to REMOTE
				Versions: []string{
					"# Project\n\nLocal README content\nwith some changes.\n",
					"# Project\n\nRemote README content\nwith different changes.\n",
				},
			},
			{
				Path:   "config.yaml",
				Chosen: 0,
				Versions: []string{
					"version: 1.0.0\nmode: development\n",
					"version: 2.0.0\nmode: production\n",
				},
			},
		},
		
		// UI state
		SelectedFileIndex: 0,
		FocusedPane:       0, // Start with first file list focused
		NumColumns:        2,
		ColumnLabels:      []string{"LOCAL", "REMOTE"},
		ScrollOffsets:     []int{0, 0},
		LineCursors:       []int{0, 0},
		
		// Create diff pane (for future use)
		DiffPane: ui.NewDiffPane(&app.theme),
	}
	
	app.mode = ModeConflictResolve
	app.footerHint = "Test conflict mode - use TAB to cycle panes, SPACE to mark choices"
	
	return nil
}
