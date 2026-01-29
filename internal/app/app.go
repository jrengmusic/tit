package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Use ui.FileHistoryState and ui.FileHistoryPane (no duplication)

// Application is the central state container for the TIT (Terminal Interface for Time Travel) application.
// It manages all application state, UI rendering, git operations, and user interactions.
//
// The Application follows a strict UI THREAD / WORKER THREAD separation pattern:
// - UI THREAD: Handles rendering, input, and immediate user feedback
// - WORKER THREAD: Executes async git operations and state updates
//
// Key Responsibilities:
// - Maintains current application mode and state
// - Manages git repository state and operations
// - Handles user input and menu navigation
// - Coordinates async operations and UI updates
// - Enforces application invariants and contracts

type Application struct {
	width             int
	height            int
	sizing            ui.DynamicSizing
	theme             ui.Theme
	mode              AppMode // Current application mode
	quitConfirmActive bool
	quitConfirmTime   time.Time
	footerHint        string // Footer hint/message text
	gitState          *git.State
	selectedIndex     int                               // Current menu item selection
	menuItems         []MenuItem                        // Cached menu items
	keyHandlers       map[AppMode]map[string]KeyHandler // Cached key handlers

	// Input mode state
	inputState InputState // Text input field state

	// Workflow state (clone, init, mode restoration)
	workflowState WorkflowState // Transient multi-step workflow state

	// Remote operation state

	// Async operation state
	asyncState AsyncState

	// Console output state (for clone, init, etc)
	consoleState ConsoleState

	// Process cancellation
	cancelContext context.CancelFunc

	// Confirmation dialog state
	dialogState DialogState

	// Conflict resolution state
	conflictResolveState *ConflictResolveState

	// Dirty operation tracking
	dirtyOperationState *DirtyOperationState // nil when no dirty op in progress

	// State display info maps
	workingTreeInfo map[git.WorkingTree]StateInfo
	timelineInfo    map[git.Timeline]StateInfo
	operationInfo   map[git.Operation]StateInfo

	// Picker state (history, file history, branch picker)
	pickerState PickerState

	// Time Travel state
	timeTravelState TimeTravelState

	// Environment state (git detection + setup wizard)
	environmentState EnvironmentState // Git environment and setup wizard state

	// History cache
	cacheManager *CacheManager

	// Config state (Session 86)
	appConfig *config.Config // Loaded from ~/.config/tit/config.toml

	// Preferences state (Session 86)

	// Activity tracking (Session 2 - Lazy auto-update)
	activityState ActivityState
}

// ModeTransition configuration for streamlined mode changes
type ModeTransition struct {
	Mode        AppMode
	InputPrompt string
	InputAction string
	FooterHint  string
	InputHeight int
	ResetFields []string
}

// transitionTo handles standardized mode transitions and state resets.
func (a *Application) transitionTo(config ModeTransition) {
	a.mode = config.Mode

	// Always reset common input state
	a.selectedIndex = 0
	a.inputState.Reset()
	a.inputState.ClearConfirming = false

	// Set new input config from the transition configuration
	if config.InputPrompt != "" {
		a.inputState.Prompt = config.InputPrompt
	}
	if config.InputAction != "" {
		a.inputState.Action = config.InputAction
	}
	if config.FooterHint != "" {
		a.footerHint = config.FooterHint
	}
	if config.InputHeight > 0 {
		a.inputState.Height = config.InputHeight
	} else if config.Mode == ModeInput || config.Mode == ModeCloneURL {
		// Default to single-line input (4 = label + 3-line box)
		a.inputState.Height = 4
	}

	// Reset workflow-specific fields based on the configuration
	for _, field := range config.ResetFields {
		switch field {
		case "clone":
			a.workflowState.ResetClone()
		case "all":
			// Reset all workflow states
			a.workflowState.ResetClone()
		}
	}
}

// reloadGitState refreshes git state from repository.
// This is SSOT for all git state reloads in the application.
func (a *Application) reloadGitState() error {
	state, err := git.DetectState()
	if err != nil {
		return err
	}
	a.gitState = state
	return nil
}

// checkForConflicts detects if git is in conflicted state after an operation.
// Returns GitOperationMsg if conflicts detected, nil otherwise.
// successFlag: set to true when caller wants to trigger conflict resolver (e.g., dirty pull merge)
// successFlag: set to false for normal conflict detection during operations
func (a *Application) checkForConflicts(step string, successFlag bool) *GitOperationMsg {
	if err := a.reloadGitState(); err != nil {
		return nil
	}
	if a.gitState.Operation == git.Conflicted {
		return &GitOperationMsg{
			Step:             step,
			Success:          successFlag,
			ConflictDetected: true,
			Error:            "Merge conflicts detected",
		}
	}
	return nil
}

// executeGitOp executes a git command and returns appropriate message.
// This is SSOT for git command execution with standard error handling.
func (a *Application) executeGitOp(step string, args ...string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelContext = cancel
	return func() tea.Msg {
		result := git.ExecuteWithStreaming(ctx, args...)
		if !result.Success {
			return GitOperationMsg{
				Step:    step,
				Success: false,
				Error:   result.Stderr,
			}
		}
		return GitOperationMsg{
			Step:    step,
			Success: true,
		}
	}
}

// ========================================
// Async Operation State Helpers
// ========================================
// SSOT for async operation lifecycle.

// startAsyncOp marks an async operation as active.
func (a *Application) startAsyncOp() {
	a.asyncState.Start()
}

// endAsyncOp marks an async operation as complete.
func (a *Application) endAsyncOp() {
	a.asyncState.End()
}

// abortAsyncOp marks an async operation as aborted by user.
func (a *Application) abortAsyncOp() {
	a.asyncState.Abort()
}

// isAsyncActive returns true if an async operation is running.
func (a *Application) isAsyncActive() bool {
	return a.asyncState.IsActive()
}

// isAsyncAborted returns true if current async operation was aborted.
func (a *Application) isAsyncAborted() bool {
	return a.asyncState.IsAborted()
}

// clearAsyncAborted resets the aborted flag.
func (a *Application) clearAsyncAborted() {
	a.asyncState.ClearAborted()
}

// setExitAllowed sets whether exit is allowed during operation.
func (a *Application) setExitAllowed(allowed bool) {
	a.asyncState.SetExitAllowed(allowed)
}

// canExit returns true if exit is allowed.
func (a *Application) canExit() bool {
	return a.asyncState.CanExit()
}

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
		// Not in a repo - use NotRepo operation state to show init/clone menu
		gitState = &git.State{
			Operation: git.NotRepo,
		}
	}

	// CRITICAL: Check for incomplete time travel from previous session (Phase 0)
	// If TIT exited while time traveling, restore to original state
	if git.FileExists(".git/TIT_TIME_TRAVEL") && isRepo {
		// Will be handled after app creation to show status
		// (defer restoration until after UI is ready)
	}

	// Build state info maps
	workingTreeInfo, timelineInfo, operationInfo := BuildStateInfo(theme)

	app := &Application{
		sizing:          sizing,
		theme:           theme,
		mode:            ModeMenu,
		gitState:        gitState,
		selectedIndex:   0,
		asyncState:      AsyncState{exitAllowed: true}, // Allow exit by default (disabled during critical operations)
		workflowState:   NewWorkflowState(),
		consoleState:    NewConsoleState(),
		workingTreeInfo: workingTreeInfo,
		timelineInfo:    timelineInfo,
		operationInfo:   operationInfo,
		pickerState: PickerState{
			History: &ui.HistoryState{
				Commits:           make([]ui.CommitInfo, 0),
				SelectedIdx:       0,
				PaneFocused:       true, // Start with list pane focused
				DetailsLineCursor: 0,
				DetailsScrollOff:  0,
			},
			FileHistory: &ui.FileHistoryState{
				Commits:           make([]ui.CommitInfo, 0),
				Files:             make([]ui.FileInfo, 0),
				SelectedCommitIdx: 0,
				SelectedFileIdx:   0,
				FocusedPane:       ui.PaneCommits, // Start with commits pane focused
				CommitsScrollOff:  0,
				FilesScrollOff:    0,
				DiffScrollOff:     0,
				DiffLineCursor:    0,
				VisualModeActive:  false,
				VisualModeStart:   0,
			},
			BranchPicker: &ui.BranchPickerState{
				Branches:          make([]ui.BranchInfo, 0),
				SelectedIdx:       0,
				PaneFocused:       true, // Start with list pane focused
				ListScrollOffset:  0,
				DetailsLineCursor: 0,
				DetailsScrollOff:  0,
			},
		},
		// Initialize cache manager
		cacheManager: NewCacheManager(),

		// Config state (Session 86) - passed from main.go (fail-fast on load errors)
		appConfig:     cfg,
		activityState: NewActivityState(),
		dialogState:   NewDialogState(),
	}

	// Build and cache key handler registry once
	app.keyHandlers = app.buildKeyHandlers()

	// CRITICAL: Check for incomplete time travel restoration (Phase 0)
	// If TIT exited while time traveling, restore immediately
	// BUT: only if we're NOT currently in TimeTraveling state (which means we're actively traveling, not incomplete)
	hasTimeTravelMarker := git.FileExists(".git/TIT_TIME_TRAVEL") && isRepo
	shouldRestore := hasTimeTravelMarker && app.gitState.Operation != git.TimeTraveling

	// Load timeTravelInfo if actively time traveling
	// CRITICAL: If TimeTraveling but can't load info, that's CORRUPT STATE
	// Force restoration immediately to recover
	if app.gitState.Operation == git.TimeTraveling && hasTimeTravelMarker {
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			// CORRUPT STATE: TimeTraveling but can't load info
			// Force restoration to recover
			shouldRestore = true
		} else {
			app.timeTravelState.SetInfo(ttInfo)
		}
	}

	if shouldRestore {
		// Show console and perform restoration
		app.mode = ModeConsole
		app.startAsyncOp()
		app.workflowState.PreviousMode = ModeMenu
		app.footerHint = "Restoring from incomplete time travel session..."
	}

	// Pre-generate menu and load initial hint (for post-restoration)
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 && !shouldRestore {
		// Only set hint if not restoring
		if app.gitState != nil && app.gitState.Remote == git.HasRemote && app.gitState.Timeline == "" && app.gitState.CurrentHash == "" {

		} else {
			app.footerHint = menu[0].Hint
		}
	}

	// Register menu shortcuts dynamically
	app.rebuildMenuShortcuts(ModeMenu)

	// Start pre-loading caches (CONTRACT: MANDATORY on startup)
	// Cache building will be triggered in Init() via tea.Cmd
	// Read-only operations, safe for any git state
	if !shouldRestore {
		app.cacheManager.SetLoadingStarted(true)
		// Cache build started in Init() method via tea.Batch
	}

	// If restoration needed, set up the async operation
	if shouldRestore {
		// Will be executed via Update() on first render
		app.startAsyncOp()
	}

	return app
}

// RestoreFromTimeTravel handles recovery from incomplete time travel sessions (Phase 0)
// Called if TIT detected .git/TIT_TIME_TRAVEL marker on startup
// Returns a tea.Cmd that performs the restoration and shows status
func (a *Application) GetFooterHint() string {
	return a.footerHint
}

// handleCacheProgress handles cache building progress updates
func (a *Application) buildKeyHandlers() map[AppMode]map[string]KeyHandler {
	// Global handlers - highest priority, applied to all modes
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash, // Open config menu
		"ctrl+v": a.handleKeyPaste, // Linux/Windows/macOS
		"cmd+v":  a.handleKeyPaste, // macOS cmd+v
		"meta+v": a.handleKeyPaste, // macOS meta (cmd) - Bubble Tea may send this
		"alt+v":  a.handleKeyPaste, // Fallback
	}

	cursorNavMixin := CursorNavigationMixin{}

	// Generic input cursor handlers for single-field inputs
	genericInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string { return a.inputState.Value },
		func(a *Application) int { return a.inputState.CursorPosition },
		func(a *Application, pos int) { a.inputState.CursorPosition = pos },
	)

	// Mode-specific handlers (global merged in after)
	modeHandlers := map[AppMode]map[string]KeyHandler{
		ModeMenu: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			Build(),
		ModeConsole: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeInput: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleInputSubmit).
			Build(),
		ModeInitializeLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleInitLocationSelection).
			On("1", a.handleInitLocationChoice1).
			On("2", a.handleInitLocationChoice2).
			Build(),
		ModeCloneURL: NewModeHandlers().
			WithCursorNav(genericInputNav).
			On("enter", a.handleCloneURLSubmit).
			Build(),
		ModeCloneLocation: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleCloneLocationSelection).
			On("1", a.handleCloneLocationChoice1).
			On("2", a.handleCloneLocationChoice2).
			Build(),
		ModeConfirmation: NewModeHandlers().
			On("left", a.handleConfirmationLeft).
			On("right", a.handleConfirmationRight).
			On("h", a.handleConfirmationLeft).
			On("l", a.handleConfirmationRight).
			On("y", a.handleConfirmationYes).
			On("n", a.handleConfirmationNo).
			On("enter", a.handleConfirmationEnter).
			Build(),
		ModeHistory: NewModeHandlers().
			On("up", a.handleHistoryUp).
			On("k", a.handleHistoryUp).
			On("down", a.handleHistoryDown).
			On("j", a.handleHistoryDown).
			On("tab", a.handleHistoryTab).
			On("enter", a.handleHistoryEnter).
			On("ctrl+r", a.handleHistoryRewind).
			On("esc", a.handleHistoryEsc).
			Build(),
		ModeFileHistory: NewModeHandlers().
			On("up", a.handleFileHistoryUp).
			On("down", a.handleFileHistoryDown).
			On("k", a.handleFileHistoryUp).
			On("j", a.handleFileHistoryDown).
			On("tab", a.handleFileHistoryTab).
			On("y", a.handleFileHistoryCopy).
			On("v", a.handleFileHistoryVisualMode).
			On("esc", a.handleFileHistoryEsc).
			Build(),
		ModeConflictResolve: NewModeHandlers().
			On("up", a.handleConflictUp).
			On("k", a.handleConflictUp).
			On("down", a.handleConflictDown).
			On("j", a.handleConflictDown).
			On("tab", a.handleConflictTab).
			On(" ", a.handleConflictSpace). // Space character, not "space"
			On("enter", a.handleConflictEnter).
			Build(),
		ModeClone: NewModeHandlers().
			On("up", a.handleConsoleUp).
			On("k", a.handleConsoleUp).
			On("down", a.handleConsoleDown).
			On("j", a.handleConsoleDown).
			On("pageup", a.handleConsolePageUp).
			On("pagedown", a.handleConsolePageDown).
			Build(),
		ModeSelectBranch: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleSelectBranchEnter).
			Build(),
		ModeSetupWizard: NewModeHandlers().
			On("enter", a.handleSetupWizardEnter).
			Build(),
		ModeConfig: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			Build(),
		ModeBranchPicker: NewModeHandlers().
			On("up", a.handleBranchPickerUp).
			On("k", a.handleBranchPickerUp).
			On("down", a.handleBranchPickerDown).
			On("j", a.handleBranchPickerDown).
			On("enter", a.handleBranchPickerEnter).
			Build(),
		ModePreferences: NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter).
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
			Build(),
	}

	// Merge global handlers into each mode (global takes priority)
	for mode := range modeHandlers {
		for key, handler := range globalHandlers {
			modeHandlers[mode][key] = handler
		}
	}

	return modeHandlers
}

// rebuildMenuShortcuts dynamically registers keyboard handlers for all current menu item shortcuts
// Called after GenerateMenu() to ensure shortcuts match current git state
func (a *Application) rebuildMenuShortcuts(mode AppMode) {
	if a.keyHandlers[mode] == nil {
		a.keyHandlers[mode] = make(map[string]KeyHandler)
	}

	// Remove old shortcut handlers (keep navigation and enter)
	// We'll rebuild from scratch by first copying base handlers
	var baseHandlers map[string]KeyHandler
	if mode == ModeMenu {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleMenuEnter).
			On(" ", a.handleMenuEnter). // Space as enter alias
			Build()
	} else if mode == ModeConfig {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handleConfigMenuEnter).
			On(" ", a.handleConfigMenuEnter). // Space as enter alias
			Build()
	} else if mode == ModePreferences {
		baseHandlers = NewModeHandlers().
			WithMenuNav(a).
			On("enter", a.handlePreferencesEnter).
			On(" ", a.handlePreferencesEnter). // Space as enter alias
			On("=", a.handlePreferencesIncrement).
			On("-", a.handlePreferencesDecrement).
			On("+", a.handlePreferencesIncrement10).
			On("_", a.handlePreferencesDecrement10).
			Build()
	}

	// Start fresh
	newHandlers := make(map[string]KeyHandler)

	// Copy base handlers
	for key, handler := range baseHandlers {
		newHandlers[key] = handler
	}

	// Merge global handlers
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"/":      a.handleKeySlash, // Open config menu
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	// Add global handlers (base handlers take priority, no overrides)
	for key, handler := range globalHandlers {
		if _, exists := baseHandlers[key]; !exists {
			newHandlers[key] = handler
		}
	}

	// Dynamically register shortcuts for current menu items
	for i, item := range a.menuItems {
		if item.Shortcut != "" && item.Enabled && !item.Separator {
			// Capture loop variables in closure
			itemIndex := i
			itemID := item.ID
			itemHint := item.Hint

			// Create handler that selects item and dispatches action
			newHandlers[item.Shortcut] = func(app *Application) (tea.Model, tea.Cmd) {
				app.selectedIndex = itemIndex
				app.footerHint = itemHint
				return app, app.dispatchAction(itemID)
			}
		}
	}

	// Replace handlers for specified mode
	a.keyHandlers[mode] = newHandlers
}

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if len(app.menuItems) > 0 {
		startIdx := app.selectedIndex
		app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for app.menuItems[app.selectedIndex].Separator || !app.menuItems[app.selectedIndex].Enabled {
			app.selectedIndex = (app.selectedIndex - 1 + len(app.menuItems)) % len(app.menuItems)
			// Prevent infinite loop if all items disabled
			if app.selectedIndex == startIdx {
				break
			}
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuDown moves selection down
func (a *Application) handleMenuDown(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if len(app.menuItems) > 0 {
		startIdx := app.selectedIndex
		app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
		// Skip separators and disabled items (CONTRACT: disabled items not selectable)
		for app.menuItems[app.selectedIndex].Separator || !app.menuItems[app.selectedIndex].Enabled {
			app.selectedIndex = (app.selectedIndex + 1) % len(app.menuItems)
			// Prevent infinite loop if all items disabled
			if app.selectedIndex == startIdx {
				break
			}
		}
		// Update footer hint
		if app.selectedIndex < len(app.menuItems) {
			app.footerHint = app.menuItems[app.selectedIndex].Hint
		}
	}
	return app, nil
}

// handleMenuEnter selects current menu item and dispatches action
func (a *Application) handleMenuEnter(app *Application) (tea.Model, tea.Cmd) {
	app.activityState.MarkActivity() // Track menu activity
	if app.selectedIndex < 0 || app.selectedIndex >= len(app.menuItems) {
		return app, nil
	}
	item := app.menuItems[app.selectedIndex]

	// CONTRACT: Cannot execute separators or disabled items (cache still building)
	if item.Separator || !item.Enabled {
		return app, nil
	}

	// Dispatch action
	return app, app.dispatchAction(item.ID)
}

// Input mode helpers

// insertTextAtCursor inserts text at current cursor position (UTF-8 safe)
func (a *Application) insertTextAtCursor(text string) {
	// Defensive bounds checking
	valueLen := len(a.inputState.Value)
	if a.inputState.CursorPosition < 0 {
		a.inputState.CursorPosition = 0
	}
	if a.inputState.CursorPosition > valueLen {
		a.inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputState.Value[:a.inputState.CursorPosition]
	after := a.inputState.Value[a.inputState.CursorPosition:]
	a.inputState.Value = before + text + after
	a.inputState.CursorPosition += len(text)
}

// deleteAtCursor deletes character before cursor (UTF-8 safe)
func (a *Application) deleteAtCursor() {
	valueLen := len(a.inputState.Value)
	if a.inputState.CursorPosition <= 0 || valueLen == 0 {
		return
	}
	if a.inputState.CursorPosition > valueLen {
		a.inputState.CursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputState.Value[:a.inputState.CursorPosition-1]
	after := a.inputState.Value[a.inputState.CursorPosition:]
	a.inputState.Value = before + after
	a.inputState.CursorPosition--
}

// updateInputValidation updates validation message for current input
func (a *Application) updateInputValidation() {
	if a.inputState.Action == "clone_url" {
		currentValue := a.inputState.Value
		if a.mode == ModeInitializeBranches {
			return // No validation in branch mode
		}
		if currentValue == "" {
			a.inputState.ValidationMsg = ""
		} else if ui.ValidateRemoteURL(currentValue) {
			a.inputState.ValidationMsg = ""
		} else {
			a.inputState.ValidationMsg = "Invalid URL format"
		}
	}
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	switch app.inputState.Action {
	case "init_branch_name":
		return app.handleInitBranchNameSubmit()
	case "init_subdir_name":
		return app.handleInputSubmitSubdirName(app)
	case "add_remote_url":
		return app.handleAddRemoteSubmit(app)
	case "commit_message":
		return app.handleCommitSubmit(app)
	case "commit_push_message":
		return app.handleCommitPushSubmit(app)
	default:
		return app, nil
	}
}

// handleHistoryRewind handles Ctrl+ENTER in history browser to initiate rewind
func (a *Application) handleHistoryRewind(app *Application) (tea.Model, tea.Cmd) {
	if app.pickerState.History == nil || len(app.pickerState.History.Commits) == 0 {
		return app, nil
	}

	if app.pickerState.History.SelectedIdx < 0 || app.pickerState.History.SelectedIdx >= len(app.pickerState.History.Commits) {
		return app, nil
	}

	selectedCommit := app.pickerState.History.Commits[app.pickerState.History.SelectedIdx]
	app.workflowState.PendingRewindCommit = selectedCommit.Hash

	return app, app.showRewindConfirmation(selectedCommit.Hash)
}

// Workflow state delegation
func (a *Application) resetCloneWorkflow() {
	a.workflowState.ResetClone()
}

func (a *Application) saveCurrentMode() {
	a.workflowState.SaveMode(a.mode, a.selectedIndex)
}

func (a *Application) restorePreviousMode() (AppMode, int) {
	return a.workflowState.RestoreMode()
}

func (a *Application) setPendingRewind(commit string) {
	a.workflowState.SetPendingRewind(commit)
}

func (a *Application) getPendingRewind() string {
	return a.workflowState.GetPendingRewind()
}

func (a *Application) clearPendingRewind() {
	a.workflowState.ClearPendingRewind()
}

// Environment state delegation
func (a *Application) isEnvironmentReady() bool {
	return a.environmentState.IsReady()
}

func (a *Application) needsEnvironmentSetup() bool {
	return a.environmentState.NeedsSetup()
}

func (a *Application) setEnvironment(env git.GitEnvironment) {
	a.environmentState.SetEnvironment(env)
}

func (a *Application) getSetupWizardStep() SetupWizardStep {
	return a.environmentState.SetupWizardStep
}

func (a *Application) setSetupWizardStep(step SetupWizardStep) {
	a.environmentState.SetWizardStep(step)
}

func (a *Application) getSetupWizardError() string {
	return a.environmentState.SetupWizardError
}

func (a *Application) setSetupWizardError(err string) {
	a.environmentState.SetWizardError(err)
}

func (a *Application) getSetupEmail() string {
	return a.environmentState.GetEmail()
}

func (a *Application) setSetupEmail(email string) {
	a.environmentState.SetEmail(email)
}

func (a *Application) markSetupKeyCopied() {
	a.environmentState.MarkKeyCopied()
}

func (a *Application) isSetupKeyCopied() bool {
	return a.environmentState.IsKeyCopied()
}

// Picker state delegation
func (a *Application) getHistoryState() *ui.HistoryState {
	return a.pickerState.GetHistory()
}

func (a *Application) setHistoryState(state *ui.HistoryState) {
	a.pickerState.SetHistory(state)
}

func (a *Application) resetHistoryState() {
	a.pickerState.ResetHistory()
}

func (a *Application) getFileHistoryState() *ui.FileHistoryState {
	return a.pickerState.GetFileHistory()
}

func (a *Application) setFileHistoryState(state *ui.FileHistoryState) {
	a.pickerState.SetFileHistory(state)
}

func (a *Application) resetFileHistoryState() {
	a.pickerState.ResetFileHistory()
}

func (a *Application) getBranchPickerState() *ui.BranchPickerState {
	return a.pickerState.GetBranchPicker()
}

func (a *Application) setBranchPickerState(state *ui.BranchPickerState) {
	a.pickerState.SetBranchPicker(state)
}

func (a *Application) resetBranchPickerState() {
	a.pickerState.ResetBranchPicker()
}

func (a *Application) resetAllPickerStates() {
	a.pickerState.ResetAll()
}

// Console state delegation
func (a *Application) getConsoleBuffer() *ui.OutputBuffer {
	return a.consoleState.GetBuffer()
}

func (a *Application) clearConsoleBuffer() {
	a.consoleState.Clear()
}

func (a *Application) scrollConsoleUp() {
	a.consoleState.ScrollUp()
}

func (a *Application) scrollConsoleDown() {
	a.consoleState.ScrollDown()
}

func (a *Application) pageConsoleUp() {
	a.consoleState.PageUp()
}

func (a *Application) pageConsoleDown() {
	a.consoleState.PageDown()
}

func (a *Application) toggleConsoleAutoScroll() {
	a.consoleState.ToggleAutoScroll()
}

func (a *Application) isConsoleAutoScroll() bool {
	return a.consoleState.IsAutoScroll()
}

func (a *Application) getConsoleState() ui.ConsoleOutState {
	return a.consoleState.GetState()
}

func (a *Application) setConsoleScrollOffset(offset int) {
	a.consoleState.SetScrollOffset(offset)
}

// Activity state delegation
func (a *Application) markMenuActivity() {
	a.activityState.MarkActivity()
}

func (a *Application) isMenuInactive() bool {
	return a.activityState.IsInactive()
}

func (a *Application) setMenuActivityTimeout(timeout time.Duration) {
	a.activityState.SetActivityTimeout(timeout)
}

// Activity state delegation (getters)
func (a *Application) getActivityTimeout() time.Duration {
	return a.activityState.GetActivityTimeout()
}

// Activity state delegation (getters for auto_update.go)
