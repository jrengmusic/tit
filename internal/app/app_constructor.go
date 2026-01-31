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
		// Embedded state clusters
		UIState:         &UIState{},
		NavigationState: &NavigationState{},
		OperationState:  &OperationState{},
		DialogManager:   &DialogManager{},
		// Feature-specific state (standalone)
		environmentState: envState,
		// Infrastructure (standalone)
		cacheManager:  NewCacheManager(),
		appConfig:     nil,
		activityState: NewActivityState(),
	}
	// Initialize embedded cluster fields
	app.UIState.SetSize(sizing.ContentInnerWidth, sizing.ContentHeight)
	app.UIState.theme = theme
	app.NavigationState.SetMode(ModeSetupWizard)
	// Initialize console state to prevent nil pointer panics
	newConsoleState := NewConsoleState()
	app.OperationState.consoleState = &newConsoleState
	app.OperationState.SetExitAllowed(true)
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
		// Embedded state clusters
		UIState:         &UIState{},
		NavigationState: &NavigationState{},
		OperationState:  &OperationState{},
		DialogManager:   &DialogManager{},
		// Core business logic (standalone)
		gitState:        gitState,
		workingTreeInfo: workingTreeInfo,
		timelineInfo:    timelineInfo,
		operationInfo:   operationInfo,
		// Infrastructure (standalone)
		cacheManager:  NewCacheManager(),
		appConfig:     cfg,
		activityState: NewActivityState(),
	}

	// Initialize embedded cluster fields
	app.UIState.SetSize(sizing.ContentInnerWidth, sizing.ContentHeight)
	app.UIState.theme = theme
	app.NavigationState.SetMode(ModeMenu)
	// Initialize console state to prevent nil pointer panics
	newConsoleState := NewConsoleState()
	app.OperationState.consoleState = &newConsoleState
	app.OperationState.SetExitAllowed(true)

	// Build and cache key handler registry for initial mode
	app.NavigationState.SetKeyHandlers(app.buildKeyHandlers())

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
		app.NavigationState.SetMode(ModeConsole)
		app.OperationState.StartAsyncOp()
		app.OperationState.GetWorkflowState().PreviousMode = ModeMenu
		app.UIState.SetFooterHint("Restoring from incomplete time travel session...")
	}

	// Generate initial menu (Phase 1 of initialization)
	menu := app.GenerateMenu()
	app.NavigationState.SetMenuItems(menu)

	// Set footer hint from first menu item (default)
	if len(menu) > 0 && !shouldRestore {
		app.UIState.SetFooterHint(menu[0].Hint)
	}

	// Set up key handlers for current mode (refreshes shortcuts)
	app.rebuildMenuShortcuts(app.NavigationState.GetMode())

	// Start cache loading if not restoring
	if !shouldRestore {
		app.cacheManager.SetLoadingStarted(true)
	}

	// If restoration needed, start the async restoration operation
	if shouldRestore {
		// Will be executed via Update() on first render
		app.OperationState.StartAsyncOp()
	}

	return app
}

// GetFooterHint returns the current footer hint text
func (a *Application) GetFooterHint() string {
	return a.UIState.GetFooterHint()
}
