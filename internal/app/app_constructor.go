package app

import (
	"fmt"
	"os"

	"github.com/jrengmusic/tit/internal/config"
	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"
)

// newSetupWizardApp creates a minimal Application for the setup wizard
// This bypasses all git state detection since git environment is not ready
func newSetupWizardApp(sizing ui.DynamicSizing, theme ui.Theme, gitEnv git.GitEnvironment) *Application {
	envState := NewEnvironmentState()
	envState.GitEnvironment = gitEnv
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
	app.UIState.Resize(sizing.ContentInnerWidth, sizing.ContentHeight)
	app.UIState.theme = theme
	app.NavigationState.mode = ModeSetupWizard
	// Initialize console state to prevent nil pointer panics
	newConsoleState := NewConsoleState()
	app.OperationState.consoleState = &newConsoleState
	app.OperationState.PermitExit(true)
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
		git.CleanStaleLocks()
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

	// LFS: auto-setup filters if binary available but not yet configured
	envState := NewEnvironmentState()
	if gitState != nil && gitState.LFS {
		if gitState.LFSReady {
			// LFS fully configured — nothing to do
		} else if git.IsLFSBinaryAvailable() {
			result := git.SetupLFSFilters()
			if result.Success {
				git.Log("LFS filters installed automatically")
				gitState.LFSReady = true
			} else {
				git.Warn("Failed to setup LFS filters: " + result.Stderr)
			}
		} else {
			git.Warn("This project uses LFS. Install git-lfs for full functionality.")
		}
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
		// Feature-specific state (standalone)
		environmentState: envState,
		// Infrastructure (standalone)
		cacheManager:  NewCacheManager(),
		appConfig:     cfg,
		activityState: NewActivityState(),
	}

	// Initialize embedded cluster fields
	app.UIState.Resize(sizing.ContentInnerWidth, sizing.ContentHeight)
	app.UIState.theme = theme
	app.NavigationState.mode = ModeMenu
	// Initialize console state to prevent nil pointer panics
	newConsoleState := NewConsoleState()
	app.OperationState.consoleState = &newConsoleState
	app.OperationState.PermitExit(true)

	// Build and cache key handler registry for initial mode
	app.NavigationState.keyHandlers = app.buildKeyHandlers()

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
			app.timeTravelState.info = ttInfo
		}
	}

	// Check for incomplete dirty pull operation (Phase 0)
	// If we're in Conflicted state AND TIT_DIRTY_OP marker exists, we have an incomplete dirty pull
	hasDirtyPullMarker := git.FileExists(".git/TIT_DIRTY_OP") && isRepo
	isConflicted := isRepo && app.gitState.Operation == git.Conflicted

	// If conflicted state detected (with or without dirty pull marker), enter conflict resolver
	if isConflicted {
		// Load the dirty operation state if marker exists
		if hasDirtyPullMarker {
			snapshot := &git.DirtyOperationSnapshot{}
			if err := snapshot.Load(); err == nil {
				// Initialize dirty operation state
				app.dirtyOperationState = &DirtyOperationState{
					Phase:          DirtyPhaseApplyChangeset, // We're in the merge phase
					OriginalBranch: snapshot.OriginalBranch,
					OriginalHead:   snapshot.OriginalHead,
				}
			}
		}
		// Skip menu generation - enter conflict resolver immediately
		app.NavigationState.mode = ModeConsole
		app.UIState.footerHint = "Conflict detected - entering resolver..."

		// Setup conflict resolver immediately
		operationType := "pull_merge" // Default for manual conflicts
		if hasDirtyPullMarker {
			operationType = "dirty_pull_changeset_apply"
		}
		_, _ = app.setupConflictResolver(operationType, []string{"BASE", "LOCAL (yours)", "REMOTE (theirs)"}) // return values ignored: model is always app, no cmd needed at startup

		// No need to generate menu or start cache loading
		return app
	}

	// Handle pre-existing Merging state at startup
	if isRepo && app.gitState.Operation == git.Merging {
		app.NavigationState.mode = ModeMenu
		// Menu will show finalize/abort merge options via menuMerging()
	}

	// Handle pre-existing Rebasing state at startup
	if isRepo && app.gitState.Operation == git.Rebasing {
		if git.HasConflicts() {
			app.NavigationState.mode = ModeConsole
			app.UIState.footerHint = "Rebase conflict detected - entering resolver..."
			_, _ = app.setupConflictResolverForRebase() // return values ignored: model is always app, no cmd needed at startup
			return app
		}
		// No conflicts — show rebase menu (continue/abort) via menuRebasing()
		app.NavigationState.mode = ModeMenu
	}

	// Generate initial menu (Phase 1 of initialization)
	menu := app.GenerateMenu()
	app.NavigationState.ReplaceMenu(menu)

	// Set footer hint from first menu item (default)
	if len(menu) > 0 && !shouldRestore {
		app.UIState.footerHint = menu[0].Hint
	}

	// Set up key handlers for current mode (refreshes shortcuts)
	app.rebuildMenuShortcuts(app.NavigationState.mode)

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
