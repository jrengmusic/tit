package app

import (
	"fmt"
	"os"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"
)

// newSetupWizardApp creates a minimal Application for the setup wizard
// This bypasses all git state detection since git environment is not ready
func newSetupWizardApp(sizing ui.DynamicSizing, theme ui.Theme, gitEnv git.GitEnvironment) *Application {
	envState := NewEnvironmentState()
	envState.SetEnvironment(gitEnv)
	app := &Application{
		sizing:           sizing,
		theme:            theme,
		mode:             ModeSetupWizard,
		environmentState: envState,
		asyncState:       AsyncState{exitAllowed: true},
		consoleState:     NewConsoleState(),
		dialogState:      NewDialogState(),
	}
	app.keyHandlers = app.buildKeyHandlers()
	return app
}

// NewApplication creates a new application instance
func NewApplication(sizing ui.DynamicSizing, theme ui.Theme, cfg *config.Config) *Application {
	// PRIORITY 0: Check git environment BEFORE anything else
	// If git/ssh not available or SSH key missing, show setup wizard
	gitEnv := git.DetectGitEnvironment()
	InitGitLogger()
	if gitEnv != git.Ready {
		return newSetupWizardApp(sizing, theme, gitEnv)
	}

	// Try to find and cd into git repository
	isRepo, repoPath := git.IsInitializedRepo()
	if !isRepo {
		// Check parent directories
		isRepo, repoPath = git.HasParentRepo()
	}

	var gitState *git.State
	if isRepo && repoPath != "" {
		// Found a repo, cd into it and detect state
		if err := os.Chdir(repoPath); err != nil {
			// cannot cd into repo - this is a fatal error
			panic(fmt.Sprintf("cannot cd into repository at %s: %v", repoPath, err))
		}
		state, err := git.DetectState()
		if err != nil {
			// In a repo but state detection failed - this should not happen
			panic(fmt.Sprintf("Failed to detect git state in repo %s: %v", repoPath, err))
		}
		gitState = state
	} else {
		// Not in a git repository - use NotRepo operation
		gitState = &git.State{Operation: git.NotRepo}
	}

	// Build state info maps for display
	workingTreeInfo, timelineInfo, operationInfo := BuildStateInfo(theme)

	// Initialize application with detected state
	app := &Application{
		sizing:          sizing,
		theme:           theme,
		mode:            ModeMenu, // Default to menu mode
		gitState:        gitState,
		selectedIndex:   0,
		asyncState:      AsyncState{exitAllowed: true}, // Allow exit during initial setup
		workflowState:   NewWorkflowState(),
		consoleState:    NewConsoleState(),
		workingTreeInfo: workingTreeInfo,
		timelineInfo:    timelineInfo,
		operationInfo:   operationInfo,
		pickerState: PickerState{
			History: &ui.HistoryState{
				Commits:     make([]ui.CommitInfo, 0),
				SelectedIdx: 0,
				PaneFocused: true,
			},
			FileHistory: &ui.FileHistoryState{
				Commits:           make([]ui.CommitInfo, 0),
				Files:             make([]ui.FileInfo, 0),
				SelectedCommitIdx: 0,
				SelectedFileIdx:   0,
				FocusedPane:       ui.PaneCommits,
			},
			BranchPicker: &ui.BranchPickerState{
				Branches:    make([]ui.BranchInfo, 0),
				SelectedIdx: 0,
				PaneFocused: true,
			},
		},
		cacheManager:  NewCacheManager(),
		appConfig:     cfg,
		activityState: NewActivityState(),
		dialogState:   NewDialogState(),
	}

	// Build and cache key handler registry for initial mode
	app.keyHandlers = app.buildKeyHandlers()

	// Check for incomplete time travel restoration (Phase 0)
	// If we're in TimeTraveling mode, TIT marker should exist
	// If marker exists but we're not in TimeTraveling mode, restoration was interrupted
	hasTimeTravelMarker := git.FileExists(".git/TIT_TIME_TRAVEL") && isRepo
	shouldRestore := hasTimeTravelMarker && app.gitState.Operation != git.TimeTraveling

	// Load time travel info if available
	if app.gitState.Operation == git.TimeTraveling && hasTimeTravelMarker {
		// Verify the marker is valid - if not, trigger restoration
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			// Marker exists but info is corrupted/invalid - trigger restoration
			shouldRestore = true
		} else {
			// Valid time travel info - restore it
			app.timeTravelState.SetInfo(ttInfo)
		}
	}

	// If restoration needed, switch to console mode and prepare for async operation
	if shouldRestore {
		app.mode = ModeConsole
		app.startAsyncOp()
		app.workflowState.PreviousMode = ModeMenu
		app.footerHint = "Restoring from incomplete time travel session..."
	}

	// Generate initial menu (Phase 1 of initialization)
	menu := app.GenerateMenu()
	app.menuItems = menu

	// Set footer hint from first menu item (default)
	if len(menu) > 0 && !shouldRestore {
		app.footerHint = menu[0].Hint
	}

	// Set up key handlers for current mode (refreshes shortcuts)
	app.rebuildMenuShortcuts(app.mode)

	// Start cache loading if not restoring
	if !shouldRestore {
		app.cacheManager.SetLoadingStarted(true)
	}

	// If restoration needed, start the async restoration operation
	if shouldRestore {
		// Will be executed via Update() on first render
		app.startAsyncOp()
	}

	return app
}

// GetFooterHint returns the current footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}
