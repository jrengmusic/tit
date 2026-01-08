package app

import (
	"os"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
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
		"init":                 a.dispatchInit,
		"clone":                a.dispatchClone,
		"add_remote":           a.dispatchAddRemote,
		"commit":               a.dispatchCommit,
		"commit_push":          a.dispatchCommitPush,
		"push":                 a.dispatchPush,
		"force_push":           a.dispatchForcePush,
		"pull_merge":           a.dispatchPullMerge,
		"pull_merge_diverged":  a.dispatchPullMerge,
		"dirty_pull_merge":     a.dispatchDirtyPullMerge,
		"replace_local":        a.dispatchReplaceLocal,
		"resolve_conflicts":    a.dispatchResolveConflicts,
		"abort_operation":      a.dispatchAbortOperation,
		"continue_operation":   a.dispatchContinueOperation,
		"history":              a.dispatchHistory,
		"file_history":         a.dispatchFileHistory,
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
	// Determine confirmation type based on timeline state
	confirmType := string(ConfirmPullMerge)
	if app.gitState.Timeline == git.Diverged {
		confirmType = string(ConfirmPullMergeDiverged)
	}

	// Show confirmation dialog for pull (merge) - may cause conflicts
	app.mode = ModeConfirmation
	labels := ConfirmationLabels[confirmType]
	app.confirmationDialog = ui.NewConfirmationDialog(
		ui.ConfirmationConfig{
			Title:       ConfirmationTitles[confirmType],
			Explanation: ConfirmationExplanations[confirmType],
			YesLabel:    labels[0],
			NoLabel:     labels[1],
			ActionID:    confirmType,
		},
		ui.ContentInnerWidth,
		&app.theme,
	)
	app.confirmationDialog.SelectNo() // Right (Cancel) selected by default
	return nil
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
	
	// Create the confirmation dialog from SSOT
	labels := ConfirmationLabels["force_push"]
	config := ui.ConfirmationConfig{
		Title:       ConfirmationTitles["force_push"],
		Explanation: ConfirmationExplanations["force_push"],
		YesLabel:    labels[0],
		NoLabel:     labels[1],
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
	
	// Create the confirmation dialog from SSOT
	labels := ConfirmationLabels["hard_reset"]
	config := ui.ConfirmationConfig{
		Title:       ConfirmationTitles["hard_reset"], 
		Explanation: ConfirmationExplanations["hard_reset"],
		YesLabel:    labels[0],
		NoLabel:     labels[1],
		ActionID:    "hard_reset",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)
	
	return nil
}

// dispatchHistory shows commit history
func (a *Application) dispatchHistory(app *Application) tea.Cmd {
	app.mode = ModeHistory
	
	// Use cached metadata to build commits list
	app.historyCacheMutex.Lock()
	defer app.historyCacheMutex.Unlock()
	
	var commits []ui.CommitInfo
	
	// Build commits from cache if available (convert git.CommitInfo → ui.CommitInfo)
	for hash, details := range app.historyMetadataCache {
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message, // Full message (not just first line)
			Time:    parseCommitDate(details.Date),
		})
	}
	
	// CRITICAL: Sort commits by time (newest first) - map iteration is unordered!
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})
	
	// Always initialize state (even if commits empty)
	app.historyState = &ui.HistoryState{
		Commits:           commits,
		SelectedIdx:       0,
		PaneFocused:       true, // List pane focused initially
		DetailsLineCursor: 0,    // Start at top of details
		DetailsScrollOff:  0,    // No scroll initially
	}
	
	// Show appropriate hint
	if len(commits) == 0 {
		if app.cacheMetadata {
			app.footerHint = "No commits found in history"
		} else {
			app.footerHint = "Loading commit history..."
		}
	}
	
	return nil
}

// dispatchFileHistory shows file(s) history
func (a *Application) dispatchFileHistory(app *Application) tea.Cmd {
	app.mode = ModeFileHistory
	
	// Use cached data to build file history state
	app.fileHistoryCacheMutex.Lock()
	defer app.fileHistoryCacheMutex.Unlock()
	
	var commits []ui.CommitInfo
	
	// Build commits from cache if available
	for hash, details := range app.historyMetadataCache {
		commits = append(commits, ui.CommitInfo{
			Hash:    hash,
			Subject: details.Message,
			Time:    parseCommitDate(details.Date),
		})
	}
	
	// Sort commits by time (newest first)
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Time.After(commits[j].Time)
	})
	
	// Get files for first commit (if any)
	var files []ui.FileInfo
	if len(commits) > 0 {
		firstCommitHash := commits[0].Hash
		if gitFileList, exists := app.fileHistoryFilesCache[firstCommitHash]; exists {
			// Convert git.FileInfo to ui.FileInfo
			for _, gitFile := range gitFileList {
				files = append(files, ui.FileInfo{
					Path:   gitFile.Path,
					Status: gitFile.Status,
				})
			}
		}
	}
	
	// Convert commits to git.CommitInfo for app state
	var gitCommits []git.CommitInfo
	for _, uiCommit := range commits {
		gitCommits = append(gitCommits, git.CommitInfo{
			Hash:    uiCommit.Hash,
			Subject: uiCommit.Subject,
			Time:    uiCommit.Time,
		})
	}
	
	// Initialize state
	app.fileHistoryState = &ui.FileHistoryState{
		Commits:           commits,
		Files:             files,
		SelectedCommitIdx: 0,
		SelectedFileIdx:   0,
		FocusedPane:       ui.PaneCommits, // Start with commits pane focused
		CommitsScrollOff:  0,
		FilesScrollOff:    0,
		DiffScrollOff:     0,
		DiffLineCursor:    0,
		VisualModeActive:  false,
		VisualModeStart:   0,
	}
	
	// Show appropriate hint
	if len(commits) == 0 {
		if app.cacheDiffs {
			app.footerHint = "No commits found in file history"
		} else {
			app.footerHint = "Loading file history..."
		}
	} else {
		app.footerHint = "File(s) History │ ↑↓ navigate │ TAB cycle panes │ ESC back"
	}
	
	return nil
}

// parseCommitDate parses git commit date format
func parseCommitDate(dateStr string) time.Time {
	// Expected format from git: "Mon, 7 Jan 2026 08:42:00 -0700"
	t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", dateStr)
	if err != nil {
		// Fallback to current time if parsing fails
		return time.Now()
	}
	return t
}

// dispatchDirtyPullMerge starts the dirty pull confirmation dialog
func (a *Application) dispatchDirtyPullMerge(app *Application) tea.Cmd {
	app.mode = ModeConfirmation
	app.confirmType = "dirty_pull"
	app.confirmContext = map[string]string{}

	// Create the confirmation dialog from SSOT
	labels := ConfirmationLabels["dirty_pull"]
	config := ui.ConfirmationConfig{
		Title:       ConfirmationTitles["dirty_pull"],
		Explanation: ConfirmationExplanations["dirty_pull"],
		YesLabel:    labels[0],
		NoLabel:     labels[1],
		ActionID:    "dirty_pull",
	}
	app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)

	return nil
}
