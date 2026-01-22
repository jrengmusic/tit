package app

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"tit/internal/git"
	"tit/internal/ui"
)

// Use ui.FileHistoryState and ui.FileHistoryPane (no duplication)

// Application represents the main TIT app state
type Application struct {
	width             int
	height            int
	sizing            ui.Sizing
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
	inputPrompt         string // e.g., "Repository name:"
	inputValue          string
	inputCursorPosition int    // Cursor byte position in inputValue
	inputHeight         int    // Input height: 1 for single-line, 16 for multiline commit message
	inputAction         string // Action being performed (e.g., "init_location", "canon_branch")
	inputValidationMsg  string // Validation feedback message (empty = valid, shows message otherwise)
	clearConfirmActive  bool   // True when waiting for second ESC to clear input

	// Clone workflow state
	cloneURL      string   // URL to clone from
	clonePath     string   // Path to clone into (cwd or subdir)
	cloneMode     string   // "here" (init+remote+fetch) or "subdir" (git clone)
	cloneBranches []string // Available branches after clone

	// Remote operation state
	remoteBranchName string // Current branch name during remote operations

	// Async operation state
	asyncOperationActive  bool    // True while git operation (clone, init, etc) is running
	asyncOperationAborted bool    // True if user pressed ESC to abort during operation
	isExitAllowed         bool    // False during critical operations (pull merge) to prevent premature exit
	previousMode          AppMode // Mode before async operation started (for restoration on ESC)
	previousMenuIndex     int     // Menu selection before async (for restoration)

	// Rewind operation state
	pendingRewindCommit string // Commit hash for pending rewind operation

	// Console output state (for clone, init, etc)
	consoleState      ui.ConsoleOutState
	outputBuffer      *ui.OutputBuffer
	consoleAutoScroll bool // Auto-scroll console to bottom (like old-tit)

	// Confirmation dialog state
	confirmationDialog *ui.ConfirmationDialog
	confirmType        string            // Type of confirmation for old-tit compatibility
	confirmContext     map[string]string // Context for old-tit compatibility

	// Conflict resolution state
	conflictResolveState *ConflictResolveState

	// Dirty operation tracking
	dirtyOperationState *DirtyOperationState // nil when no dirty op in progress

	// State display info maps
	workingTreeInfo map[git.WorkingTree]StateInfo
	timelineInfo    map[git.Timeline]StateInfo
	operationInfo   map[git.Operation]StateInfo

	// History mode state
	historyState *ui.HistoryState

	// File(s) History mode state
	fileHistoryState *ui.FileHistoryState

	// Time Travel state
	timeTravelInfo             *git.TimeTravelInfo // Non-nil only when Operation = TimeTraveling
	restoreTimeTravelInitiated bool                // True once restoration has been started

	// Git Environment state (5th axis - checked before all other state)
	gitEnvironment  git.GitEnvironment // Ready, NeedsSetup, MissingGit, MissingSSH
	setupWizardStep SetupWizardStep    // Current step in setup wizard
	setupEmail      string             // Email for SSH key generation
	setupKeyCopied  bool               // True once public key copied to clipboard

	// Cache fields (Phase 2)
	historyMetadataCache  map[string]*git.CommitDetails // hash → commit metadata
	fileHistoryDiffCache  map[string]string             // hash:path:version → diff content
	fileHistoryFilesCache map[string][]git.FileInfo     // hash → file list

	// Cache status flags (CONTRACT: MANDATORY precomputation, no on-the-fly)
	cacheLoadingStarted bool // Guard against re-preloading
	cacheMetadata       bool // true when history metadata cache populated
	cacheDiffs          bool // true when file(s) history diffs cache populated

	// Cache progress tracking (for UI feedback during build)
	cacheMetadataProgress int // Current commit processed for metadata
	cacheMetadataTotal    int // Total commits to process
	cacheDiffsProgress    int // Current commit processed for diffs
	cacheDiffsTotal       int // Total commits to process
	cacheAnimationFrame   int // Animation frame for loading spinner

	// Mutexes for thread-safe cache access
	historyCacheMutex     sync.Mutex
	fileHistoryCacheMutex sync.Mutex
	diffCacheMutex        sync.Mutex
}

// ModeTransition configuration for streamlined mode changes
type ModeTransition struct {
	Mode        AppMode
	InputPrompt string
	InputAction string
	FooterHint  string
	ResetFields []string // Field names to reset: "clone", "init", "all"
}

// transitionTo handles standardized mode transitions and state resets.
func (a *Application) transitionTo(config ModeTransition) {
	a.mode = config.Mode

	// Always reset common input state
	a.selectedIndex = 0
	a.inputValue = ""
	a.inputCursorPosition = 0
	a.inputValidationMsg = ""
	a.clearConfirmActive = false

	// Set new input config from the transition configuration
	if config.InputPrompt != "" {
		a.inputPrompt = config.InputPrompt
	}
	if config.InputAction != "" {
		a.inputAction = config.InputAction
	}
	if config.FooterHint != "" {
		a.footerHint = config.FooterHint
	}

	// Reset workflow-specific fields based on the configuration
	for _, field := range config.ResetFields {
		switch field {
		case "clone":
			a.cloneURL = ""
			a.clonePath = ""
			a.cloneBranches = nil
		case "all":
			// Reset all workflow states
			a.cloneURL = ""
			a.clonePath = ""
			a.cloneBranches = nil
		}
	}
}

// newSetupWizardApp creates a minimal Application for the setup wizard
// This bypasses all git state detection since git environment is not ready
func newSetupWizardApp(sizing ui.Sizing, theme ui.Theme, gitEnv git.GitEnvironment) *Application {
	app := &Application{
		sizing:          sizing,
		theme:           theme,
		mode:            ModeSetupWizard,
		gitEnvironment:  gitEnv,
		setupWizardStep: SetupStepWelcome,
		isExitAllowed:   true,
		consoleState:    ui.NewConsoleOutState(),
		outputBuffer:    ui.GetBuffer(),
	}
	app.keyHandlers = app.buildKeyHandlers()
	return app
}

// NewApplication creates a new application instance
func NewApplication(sizing ui.Sizing, theme ui.Theme) *Application {
	// PRIORITY 0: Check git environment BEFORE anything else
	// If git/ssh not available or SSH key missing, show setup wizard
	gitEnv := git.DetectGitEnvironment()
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
			// Can't cd into repo - this is a fatal error
			panic(fmt.Sprintf("Cannot cd into repository at %s: %v", repoPath, err))
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
		sizing:                sizing,
		theme:                 theme,
		mode:                  ModeMenu,
		gitState:              gitState,
		selectedIndex:         0,
		asyncOperationActive:  false,
		asyncOperationAborted: false,
		isExitAllowed:         true, // Allow exit by default (disabled during critical operations)
		consoleState:          ui.NewConsoleOutState(),
		outputBuffer:          ui.GetBuffer(),
		consoleAutoScroll:     true, // Start with auto-scroll enabled
		workingTreeInfo:       workingTreeInfo,
		timelineInfo:          timelineInfo,
		operationInfo:         operationInfo,
		historyState: &ui.HistoryState{
			Commits:           make([]ui.CommitInfo, 0),
			SelectedIdx:       0,
			PaneFocused:       true, // Start with list pane focused
			DetailsLineCursor: 0,
			DetailsScrollOff:  0,
		},
		fileHistoryState: &ui.FileHistoryState{
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
		// Initialize cache fields (Phase 2)
		historyMetadataCache:  make(map[string]*git.CommitDetails),
		fileHistoryDiffCache:  make(map[string]string),
		fileHistoryFilesCache: make(map[string][]git.FileInfo),
		cacheLoadingStarted:   false,
		cacheMetadata:         false,
		cacheDiffs:            false,
	}

	// Build and cache key handler registry once
	app.keyHandlers = app.buildKeyHandlers()

	// CRITICAL: Check for incomplete time travel restoration (Phase 0)
	// If TIT exited while time traveling, restore immediately
	// BUT: only if we're NOT currently in TimeTraveling state (which means we're actively traveling, not incomplete)
	hasTimeTravelMarker := git.FileExists(".git/TIT_TIME_TRAVEL") && isRepo
	shouldRestore := hasTimeTravelMarker && app.gitState.Operation != git.TimeTraveling

	// DEBUG: Log restoration decision to file
	markerPath := ".git/TIT_TIME_TRAVEL"
	markerStat, markerErr := os.Stat(markerPath)
	debugMsg := fmt.Sprintf("[INIT] hasTimeTravelMarker=%v, isRepo=%v, Operation=%v, shouldRestore=%v, cwd=%s, marker_stat=%v, marker_err=%v\n",
		hasTimeTravelMarker, isRepo, app.gitState.Operation, shouldRestore, os.Getenv("PWD"), markerStat != nil, markerErr)
	os.WriteFile("/tmp/tit-init-debug.log", []byte(debugMsg), 0644)

	// Load timeTravelInfo if actively time traveling
	// CRITICAL: If TimeTraveling but can't load info, that's CORRUPT STATE
	// Force restoration immediately to recover
	if app.gitState.Operation == git.TimeTraveling && hasTimeTravelMarker {
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			// CORRUPT STATE: TimeTraveling but can't load info
			// Force restoration to recover
			shouldRestore = true
			debugMsg += fmt.Sprintf("[CORRUPT] TimeTraveling but LoadTimeTravelInfo failed: %v\n", err)
			os.WriteFile("/tmp/tit-init-debug.log", []byte(debugMsg), 0644)
		} else {
			app.timeTravelInfo = ttInfo
		}
	}

	if shouldRestore {
		// Show console and perform restoration
		app.mode = ModeConsole
		app.asyncOperationActive = true
		app.previousMode = ModeMenu
		app.footerHint = "Restoring from incomplete time travel session..."
	}

	// Pre-generate menu and load initial hint (for post-restoration)
	menu := app.GenerateMenu()
	app.menuItems = menu
	if len(menu) > 0 && !shouldRestore {
		// Only set hint if not restoring
		if app.gitState != nil && app.gitState.Remote == git.HasRemote && app.gitState.Timeline == "" && app.gitState.CurrentHash == "" {
			app.footerHint = FooterHints["no_commits_yet"]
		} else {
			app.footerHint = menu[0].Hint
		}
	}

	// Register menu shortcuts dynamically
	app.rebuildMenuShortcuts()

	// Start pre-loading caches (CONTRACT: MANDATORY on startup)
	// Cache building will be triggered in Init() via tea.Cmd
	// Read-only operations, safe for any git state
	if !shouldRestore {
		app.cacheLoadingStarted = true
		// Cache build started in Init() method via tea.Batch
	}

	// If restoration needed, set up the async operation
	if shouldRestore {
		// Will be executed via Update() on first render
		app.asyncOperationActive = true
	}

	return app
}

// RestoreFromTimeTravel handles recovery from incomplete time travel sessions (Phase 0)
// Called if TIT detected .git/TIT_TIME_TRAVEL marker on startup
// Returns a tea.Cmd that performs the restoration and shows status
func (a *Application) RestoreFromTimeTravel() tea.Cmd {
	return func() tea.Msg {
		buffer := ui.GetBuffer()
		buffer.Clear()
		buffer.Append(FooterHints["restoring_time_travel"], ui.TypeStatus)

		// Load time travel info
		ttInfo, err := git.LoadTimeTravelInfo()
		if err != nil {
			// Use standardized error logging (PATTERN: ErrorWarn for recovery paths)
			a.LogError(ErrorConfig{
				Level:      ErrorWarn,
				Message:    "Failed to load time travel info",
				InnerError: err,
				BufferLine: fmt.Sprintf(ErrorMessages["failed_load_time_travel_info"], err),
				FooterLine: "Failed to restore time travel state",
			})
			return RestoreTimeTravelMsg{
				Success: false,
				Error:   err.Error(),
			}
		}

		if ttInfo == nil {
			// Marker exists but no info - cleanup and continue
			git.ClearTimeTravelInfo()
			buffer.Append(FooterHints["marker_corrupted"], ui.TypeStatus)
			return RestoreTimeTravelMsg{
				Success: true,
				Error:   "",
			}
		}

		// Step 1: Discard any changes made during time travel
		buffer.Append(FooterHints["step_1_discarding"], ui.TypeStatus)
		// Use reset --hard instead of checkout . (works with uncommitted changes)
		resetResult := git.Execute("reset", "--hard", "HEAD")
		if !resetResult.Success {
			buffer.Append(FooterHints["warning_discard_changes"], ui.TypeStatus)
		}

		cleanResult := git.Execute("clean", "-fd")
		if !cleanResult.Success {
			buffer.Append(FooterHints["warning_remove_untracked"], ui.TypeStatus)
		}

		// Step 2: Return to original branch
		buffer.Append(fmt.Sprintf(FooterHints["step_2_returning"], ttInfo.OriginalBranch), ui.TypeStatus)
		checkoutBranchResult := git.Execute("checkout", ttInfo.OriginalBranch)
		if !checkoutBranchResult.Success {
			buffer.Append(fmt.Sprintf(FooterHints["error_checkout_branch"], ttInfo.OriginalBranch), ui.TypeStderr)
			return RestoreTimeTravelMsg{
				Success: false,
				Error:   "Failed to checkout original branch",
			}
		}

		// Step 3: Restore original stashed work if any
		if ttInfo.OriginalStashID != "" {
			buffer.Append(FooterHints["step_3_restoring_work"], ui.TypeStatus)
			applyResult := git.Execute("stash", "apply", ttInfo.OriginalStashID)
			if !applyResult.Success {
				buffer.Append(FooterHints["warning_restore_work"], ui.TypeStatus)
			} else {
				buffer.Append(FooterHints["original_work_restored"], ui.TypeStatus)
				dropResult := git.Execute("stash", "drop", ttInfo.OriginalStashID)
				if !dropResult.Success {
					buffer.Append(FooterHints["warning_cleanup_stash"], ui.TypeStatus)
				}
			}
		}

		// Step 4: Clean up marker
		buffer.Append(FooterHints["step_4_cleaning_marker"], ui.TypeStatus)
		err = git.ClearTimeTravelInfo()
		if err != nil {
			buffer.Append(fmt.Sprintf(FooterHints["warning_remove_marker"], err), ui.TypeStatus)
		}

		buffer.Append(FooterHints["restoration_complete"], ui.TypeStatus)

		return RestoreTimeTravelMsg{
			Success: true,
			Error:   "",
		}
	}
}

// handleRewind processes the result of a rewind (git reset --hard) operation
// Stays in console until user presses ESC
func (a *Application) handleRewind(msg RewindMsg) (tea.Model, tea.Cmd) {
	a.asyncOperationActive = false
	buffer := ui.GetBuffer()

	if !msg.Success {
		// Rewind failed - show error in console and stay
		buffer.Append(fmt.Sprintf(ErrorMessages["rewind_failed"], msg.Error), ui.TypeStderr)
		a.footerHint = GetFooterMessageText(MessageOperationFailed)
		return a, nil
	}

	// Rewind succeeded - show success message and stay in console
	buffer.Append(OutputMessages["rewind_completed"], ui.TypeStatus)
	a.footerHint = GetFooterMessageText(MessageOperationComplete)

	// Refresh git state for next menu display (when user presses ESC)
	if state, err := git.DetectState(); err == nil {
		a.gitState = state
	}

	return a, nil
}

// handleRestoreTimeTravel processes the result of time travel restoration (Phase 0)
func (a *Application) handleRestoreTimeTravel(msg RestoreTimeTravelMsg) (tea.Model, tea.Cmd) {
	a.asyncOperationActive = false
	a.restoreTimeTravelInitiated = false

	if !msg.Success {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(FooterHints["restoration_error"], msg.Error), ui.TypeStderr)
		a.footerHint = "Press ESC to acknowledge error"
		// Stay in console mode so user can read error
		return a, nil
	}

	// Reload git state after successful restoration
	state, err := git.DetectState()
	if err != nil {
		buffer := ui.GetBuffer()
		buffer.Append(fmt.Sprintf(FooterHints["error_detect_state"], err), ui.TypeStderr)
	} else {
		a.gitState = state
	}

	// Regenerate menu for new state
	menu := a.GenerateMenu()
	a.menuItems = menu
	a.selectedIndex = 0
	if len(menu) > 0 {
		a.footerHint = menu[0].Hint
	}

	// Stay in console mode until user presses ESC to acknowledge
	a.footerHint = "Restoration complete. Press ESC to continue"
	return a, nil
}

// Package-level counter for Update() calls
var updateCallCount int = 0

// Update handles all messages
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updateCallCount++
	// CRITICAL: If restoration is needed, initiate it on first update (Phase 0)
	hasMarker := git.FileExists(".git/TIT_TIME_TRAVEL")
	markerPath := ".git/TIT_TIME_TRAVEL"
	markerStat, markerErr := os.Stat(markerPath)

	debugLog := fmt.Sprintf("[UPDATE #%d] asyncActive=%v, mode=%v, hasMarker=%v, marker_stat=%v, marker_err=%v\n",
		updateCallCount, a.asyncOperationActive, a.mode, hasMarker, markerStat != nil, markerErr)
	// Append to log file instead of overwriting
	f, _ := os.OpenFile("/tmp/tit-update-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f.WriteString(debugLog)
	f.Close()

	// Log each condition separately
	cond1 := a.asyncOperationActive
	cond2 := a.mode == ModeConsole
	cond3 := !a.restoreTimeTravelInitiated
	cond4 := hasMarker

	condLog := fmt.Sprintf("[CONDITIONS #%d] cond1(asyncActive)=%v, cond2(mode==console)=%v, cond3(!restored)=%v, cond4(hasMarker)=%v, all=%v\n",
		updateCallCount, cond1, cond2, cond3, cond4, cond1 && cond2 && cond3 && cond4)
	// Append to log file
	f2, _ := os.OpenFile("/tmp/tit-conditions-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	f2.WriteString(condLog)
	f2.Close()

	// Double-check before calling RestoreFromTimeTravel
	if cond1 && cond2 && cond3 && cond4 {
		// Verify marker still exists right before restoration
		markerExists := git.FileExists(".git/TIT_TIME_TRAVEL")
		debugLog := fmt.Sprintf("[UPDATE #%d] RESTORE TRIGGERED: asyncActive=%v, mode=%v, hasMarker=%v, markerExists=%v\n",
			updateCallCount, a.asyncOperationActive, a.mode, hasMarker, markerExists)
		// Append to log
		f, _ := os.OpenFile("/tmp/tit-restore-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.WriteString(debugLog)
		f.Close()
		a.restoreTimeTravelInitiated = true
		return a, a.RestoreFromTimeTravel()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case tea.KeyMsg:
		// Handle bracketed paste - entire paste comes as single KeyMsg with Paste=true
		if msg.Paste && a.isInputMode() {
			text := string(msg.Runes) // Don't trim - preserve formatting
			if len(text) > 0 {
				a.insertTextAtCursor(text)
				a.updateInputValidation()
			}
			return a, nil
		}
		keyStr := msg.String()

		// DEBUG: Log key press and mode (for enter and ctrl+enter)
		if keyStr == "enter" || keyStr == "ctrl+enter" || strings.Contains(keyStr, "enter") {
			debugLog := fmt.Sprintf("[KEY] mode=%v, key=%s, hasHandler=%v, handlerExists=%v\n",
				a.mode, keyStr,
				a.keyHandlers[a.mode] != nil,
				a.keyHandlers[a.mode] != nil && a.keyHandlers[a.mode][keyStr] != nil)
			os.WriteFile("/tmp/tit-key-debug.log", []byte(debugLog), 0644)
		}

		// Handle ctrl+j (shift+enter equivalent) for newline in multiline input
		if a.isInputMode() && (keyStr == "ctrl+j" || keyStr == "shift+enter") {
			a.insertTextAtCursor("\n")
			return a, nil
		}

		// Look up handler in cached registry
		if modeHandlers, modeExists := a.keyHandlers[a.mode]; modeExists {
			if handler, exists := modeHandlers[keyStr]; exists {
				return handler(a)
			}
		}

		// Handle character input in input modes
		if a.isInputMode() && len(keyStr) == 1 && keyStr[0] >= 32 && keyStr[0] <= 126 {
			a.insertTextAtCursor(keyStr)
			a.updateInputValidation()
			return a, nil
		}

		// Handle backspace in input modes
		if a.isInputMode() && keyStr == "backspace" {
			a.deleteAtCursor()
			a.updateInputValidation()
			return a, nil
		}

	case TickMsg:
		if a.quitConfirmActive {
			a.quitConfirmActive = false
			a.footerHint = "" // Clear confirmation message
		}

	case ClearTickMsg:
		if a.clearConfirmActive {
			a.clearConfirmActive = false
			a.footerHint = "" // Clear confirmation message
		}

	case OutputRefreshMsg:
		// Force re-render to display updated console output
		// If operation still active, schedule next refresh tick
		if a.asyncOperationActive {
			return a, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
				return OutputRefreshMsg{}
			})
		}
		// Operation completed, stop sending refresh messages
		return a, nil

	case GitOperationMsg:
		// AUDIO THREAD - Worker returned git operation result
		return a.handleGitOperation(msg)

	case CacheProgressMsg:
		// Cache building progress update
		return a.handleCacheProgress(msg)

	case CacheRefreshTickMsg:
		// Periodic tick to refresh cache progress UI
		return a.handleCacheRefreshTick()

	case RestoreTimeTravelMsg:
		// Time travel restoration completed (Phase 0)
		return a.handleRestoreTimeTravel(msg)

	case git.TimeTravelCheckoutMsg:
		// Time travel checkout operation completed
		return a.handleTimeTravelCheckout(msg)

	case git.TimeTravelMergeMsg:
		// Time travel merge operation completed
		return a.handleTimeTravelMerge(msg)

	case git.TimeTravelReturnMsg:
		// Time travel return operation completed
		return a.handleTimeTravelReturn(msg)
	case SetupCompleteMsg:
		// SSH key generation completed successfully
		if msg.Step == "generate" {
			a.setupWizardStep = SetupStepDisplayKey
		}
		return a, nil

	case SetupErrorMsg:
		// Error occurred during setup
		// For now, just log the error and stay on current step
		// TODO: Show error to user in UI
		return a, nil

	case RewindMsg:
		// AUDIO THREAD - Rewind (git reset --hard) operation completed
		return a.handleRewind(msg)

	case RemoteFetchMsg:
		// Background fetch completed - refresh state to update timeline
		if msg.Success {
			if newState, err := git.DetectState(); err == nil {
				a.gitState = newState
				a.menuItems = a.GenerateMenu()
				a.rebuildMenuShortcuts()
				a.updateFooterHintFromMenu()
			}
		}
		return a, nil
	}

	return a, nil
}

// View renders the current view
func (a *Application) View() string {
	var contentText string

	// Render based on current mode
	switch a.mode {
	case ModeMenu:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme, ui.ContentHeight)

	case ModeConsole, ModeClone:
		// Console output (both during and after operation)
		contentText = ui.RenderConsoleOutput(
			&a.consoleState,
			a.outputBuffer,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight,
			a.asyncOperationActive && !a.asyncOperationAborted,
			a.asyncOperationAborted,
			a.consoleAutoScroll,
		)

	case ModeConfirmation:
		// Confirmation dialog (centered in content area)
		if a.confirmationDialog != nil {
			contentText = a.confirmationDialog.Render()
		} else {
			// Fallback if no dialog - return to menu
			a.mode = ModeMenu
			contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuItems), a.selectedIndex, a.theme, ui.ContentHeight)
		}

	case ModeSelectBranch:
		// Dynamic menu from cloneBranches
		items := make([]map[string]interface{}, len(a.cloneBranches))
		for i, branch := range a.cloneBranches {
			items[i] = map[string]interface{}{
				"ID":        branch,
				"Shortcut":  "",
				"Emoji":     "🌿",
				"Label":     branch,
				"Hint":      fmt.Sprintf("Set %s as canon branch", branch),
				"Enabled":   true,
				"Separator": false,
			}
		}
		contentText = ui.RenderMenuWithHeight(items, a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInput:
		textInputState := ui.TextInputState{
			Value:     a.inputValue,
			CursorPos: a.inputCursorPosition,
			Height:    a.inputHeight, // Use configured height
		}

		// Render text input with optional validation message
		inputContent := ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)

		// Append validation message if present
		if a.inputValidationMsg != "" {
			inputContent += "\n\n" + a.inputValidationMsg
		}

		contentText = inputContent
	case ModeCloneURL:
		textInputState := ui.TextInputState{
			Value:     a.inputValue,
			CursorPos: a.inputCursorPosition,
			Height:    1,
		}

		inputContent := ui.RenderTextInput(
			a.inputPrompt,
			textInputState,
			a.theme,
			ui.ContentInnerWidth,
			ui.ContentHeight-2,
		)

		if a.inputValidationMsg != "" {
			inputContent += "\n\n" + a.inputValidationMsg
		}

		contentText = inputContent
	case ModeCloneLocation:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuCloneLocation()), a.selectedIndex, a.theme, ui.ContentHeight)
	case ModeInitializeLocation:
		contentText = ui.RenderMenuWithHeight(a.menuItemsToMaps(a.menuInitializeLocation()), a.selectedIndex, a.theme, ui.ContentHeight)

	case ModeHistory:
		// Render history split-pane view
		if a.historyState == nil {
			contentText = "History state not initialized"
		} else {
			contentText = ui.RenderHistorySplitPane(
				a.historyState,
				a.theme,
				ui.ContentInnerWidth,
				ui.ContentHeight, // RenderHistorySplitPane accounts for outer border internally
			)
		}
	case ModeFileHistory:
		// Render file(s) history split-pane view
		if a.fileHistoryState == nil {
			contentText = "File history state not initialized"
		} else {
			contentText = ui.RenderFileHistorySplitPane(
				a.fileHistoryState,
				a.theme,
				ui.ContentInnerWidth,
				ui.ContentHeight, // RenderFileHistorySplitPane accounts for outer border internally
			)
		}
	case ModeConflictResolve:
		// Render conflict resolution UI using generic N-column view
		if a.conflictResolveState == nil {
			contentText = "No conflict state initialized"
		} else {
			contentText = ui.RenderConflictResolveGeneric(
				a.conflictResolveState.Files,
				a.conflictResolveState.SelectedFileIndex,
				a.conflictResolveState.FocusedPane,
				a.conflictResolveState.NumColumns,
				a.conflictResolveState.ColumnLabels,
				a.conflictResolveState.ScrollOffsets,
				a.conflictResolveState.LineCursors,
				ui.ContentInnerWidth,
				ui.ContentHeight,
				a.theme,
			)
		}
	case ModeSetupWizard:
		// Git environment setup wizard
		contentText = a.renderSetupWizard()
	default:
		panic(fmt.Sprintf("Unknown app mode: %v", a.mode))
	}

	// Get current branch from git state
	currentBranch := ""
	if a.gitState != nil {
		currentBranch = a.gitState.CurrentBranch
	}

	// Get current working directory
	cwd, _ := os.Getwd()

	return ui.RenderLayout(a.sizing, contentText, a.width, a.height, a.theme, currentBranch, cwd, a.gitState, a)
}

// Init initializes the application
func (a *Application) Init() tea.Cmd {
	// CONTRACT: Start cache building immediately on app startup
	// Cache MUST be ready before history menus can be used
	commands := []tea.Cmd{tea.EnableBracketedPaste}

	if a.cacheLoadingStarted {
		commands = append(commands,
			a.cmdPreloadHistoryMetadata(),
			a.cmdPreloadFileHistoryDiffs(),
		)
	}

	// Async fetch remote on startup to ensure timeline accuracy
	// Without this, timeline state uses stale local refs
	if a.gitState != nil && a.gitState.Remote == git.HasRemote {
		commands = append(commands, cmdFetchRemote())
	}

	return tea.Batch(commands...)
}

// GetFooterHint returns the footer hint text
func (a *Application) GetFooterHint() string {
	return a.footerHint
}

// handleCacheProgress handles cache building progress updates
func (a *Application) handleCacheProgress(msg CacheProgressMsg) (tea.Model, tea.Cmd) {
	// Cache progress received - regenerate menu to show updated progress
	// Menu generator will read progress fields and show disabled state with progress

	// Always regenerate menu on progress update (menu reads progress fields)
	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 && a.selectedIndex < len(menu) {
		a.footerHint = menu[a.selectedIndex].Hint
	}

	if msg.Complete {
		// Cache complete - rebuild menu shortcuts to enable items
		a.rebuildMenuShortcuts()

		// Check if BOTH caches are now complete (for time travel success message)
		a.historyCacheMutex.Lock()
		metadataReady := a.cacheMetadata
		a.historyCacheMutex.Unlock()

		a.diffCacheMutex.Lock()
		diffsReady := a.cacheDiffs
		a.diffCacheMutex.Unlock()

		// If both caches complete AND in console mode during time travel, show final message
		if metadataReady && diffsReady && a.mode == ModeConsole && a.gitState != nil && a.gitState.Operation == git.TimeTraveling {
			buffer := ui.GetBuffer()
			buffer.Append("Time travel successful. Press ESC to return to menu.", ui.TypeStatus)
			a.footerHint = "Time travel successful. Press ESC to return to menu."
		}
	}

	return a, nil
}

// handleCacheRefreshTick handles periodic cache progress refresh
// Regenerates menu to show updated progress and re-schedules if caches not complete
func (a *Application) handleCacheRefreshTick() (tea.Model, tea.Cmd) {
	// Check if both caches are complete
	a.historyCacheMutex.Lock()
	metadataComplete := a.cacheMetadata
	a.historyCacheMutex.Unlock()

	a.diffCacheMutex.Lock()
	diffsComplete := a.cacheDiffs
	a.diffCacheMutex.Unlock()

	// If both complete, stop ticking
	if metadataComplete && diffsComplete {
		return a, nil
	}

	// Advance animation frame
	a.cacheAnimationFrame++

	// Regenerate menu to show updated progress
	menu := a.GenerateMenu()
	a.menuItems = menu
	if len(menu) > 0 && a.selectedIndex < len(menu) {
		a.footerHint = menu[a.selectedIndex].Hint
	}

	// Re-schedule another tick
	return a, a.cmdRefreshCacheProgress()
}

// updateFooterHintFromMenu updates footer with hint of currently selected menu item
func (a *Application) updateFooterHintFromMenu() {
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.menuItems) {
		if !a.menuItems[a.selectedIndex].Separator {
			a.footerHint = a.menuItems[a.selectedIndex].Hint
		}
	}
}

// GetGitState returns the current git state
func (a *Application) GetGitState() interface{} {
	return a.gitState
}

// RenderStateHeader renders the full git state header (5 rows) using lipgloss
// Row 1: CWD (left) | OPERATION (right)
// Row 2: REMOTE (left) | BRANCH (right)
// Row 3: Separator line
// Row 4: WORKING TREE (left) | TIMELINE (right) - 2 columns, 2 rows each
// Row 5: WT Description (left) | TL Description (right)
func (a *Application) RenderStateHeader() string {
	cwd, _ := os.Getwd()
	state := a.gitState

	if state == nil || state.Operation == git.NotRepo {
		// Don't render state header if not in a repo
		return ""
	}

	// Guard: Skip rendering if WorkingTree is empty (happens during dirty operations)
	// DetectState() returns partial state for dirty operations (only Operation is set)
	if state.WorkingTree == "" {
		return ""
	}

	// Right column: fixed 10 chars for short labels (READY, main)
	// Left column: remaining width for cwd + remote
	rightWidth := 10
	leftWidth := ui.ContentInnerWidth - rightWidth
	halfWidth := ui.ContentInnerWidth / 2

	// Row 1: CWD (left) | OPERATION (right)
	cwdLabel := "📁 " + cwd
	opInfo := a.operationInfo[state.Operation]
	opLabel := opInfo.Emoji + " " + opInfo.Label

	// Show operation only if not Normal (special state)
	opColor := a.theme.DimmedTextColor
	if state.Operation != git.Normal {
		opColor = opInfo.Color
	}

	row1 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(leftWidth).
			Bold(true).
			Foreground(lipgloss.Color(a.theme.LabelTextColor)).
			Render(cwdLabel),
		lipgloss.NewStyle().
			Width(rightWidth).
			Bold(true).
			Foreground(lipgloss.Color(opColor)).
			Align(lipgloss.Right).
			Render(opLabel),
	)

	// Row 2: REMOTE (left) | BRANCH (right)
	remoteLabel := "🔌 NO REMOTE"
	remoteColor := a.theme.DimmedTextColor
	if state.Remote == git.HasRemote {
		url := git.GetRemoteURL()
		if url != "" {
			remoteLabel = "🔗 " + url
			remoteColor = a.theme.AccentTextColor
		}
	}

	branchLabel := "🌿 " + state.CurrentBranch
	branchColor := a.theme.LabelTextColor

	// Special case: Detached HEAD (either time traveling or manual checkout)
	if state.Operation == git.TimeTraveling || state.Detached {
		branchLabel = "🔀 DETACHED"
		branchColor = a.theme.OutputWarningColor
	}

	row2 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(leftWidth).
			Foreground(lipgloss.Color(remoteColor)).
			Render(remoteLabel),
		lipgloss.NewStyle().
			Width(rightWidth).
			Bold(true).
			Foreground(lipgloss.Color(branchColor)).
			Align(lipgloss.Right).
			Render(branchLabel),
	)

	// Row 3: Separator line (horizontal rule)
	separatorLine := lipgloss.NewStyle().
		Width(ui.ContentInnerWidth).
		Foreground(lipgloss.Color(a.theme.BoxBorderColor)).
		Render(strings.Repeat("─", ui.ContentInnerWidth))

	// Row 4: WORKING TREE (left) | TIMELINE or Commit info (right)
	// 2 columns, equal width, left aligned
	wtInfo := a.workingTreeInfo[state.WorkingTree]
	wtLabel := wtInfo.Emoji + " " + wtInfo.Label

	var rightLabel string
	var rightColor string

	if state.Operation == git.TimeTraveling {
		// Show commit hash instead of timeline
		// INVARIANT: TimeTraveling MUST have valid timeTravelInfo (enforced at init)
		if a.timeTravelInfo == nil {
			panic("INVARIANT VIOLATION: Operation=TimeTraveling but timeTravelInfo=nil")
		}
		if len(a.timeTravelInfo.CurrentCommit.Hash) < 7 {
			panic(fmt.Sprintf("INVARIANT VIOLATION: TimeTraveling commit hash too short: '%s'", a.timeTravelInfo.CurrentCommit.Hash))
		}

		shortHash := a.timeTravelInfo.CurrentCommit.Hash[:7]
		rightLabel = "📌 Commit: " + shortHash
		rightColor = a.theme.AccentTextColor
	} else {
		// Show timeline status (if applicable)
		if state.Timeline == "" {
			// Timeline N/A (no remote, no comparison possible)
			rightLabel = "🔌 N/A"
			rightColor = a.theme.DimmedTextColor
		} else {
			tlInfo := a.timelineInfo[state.Timeline]
			rightLabel = tlInfo.Emoji + " " + tlInfo.Label
			rightColor = tlInfo.Color
		}
	}

	row4 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(halfWidth).
			Bold(true).
			Foreground(lipgloss.Color(wtInfo.Color)).
			Render(wtLabel),
		lipgloss.NewStyle().
			Width(halfWidth).
			Bold(true).
			Foreground(lipgloss.Color(rightColor)).
			Render(rightLabel),
	)

	// Row 5: Descriptions (2 columns, equal width, left aligned)
	wtDesc := wtInfo.Description(state.CommitsAhead, state.CommitsBehind)

	var rightDesc string
	if state.Operation == git.TimeTraveling {
		// Show commit date
		// INVARIANT: TimeTraveling MUST have valid timeTravelInfo (enforced at init)
		if a.timeTravelInfo == nil {
			panic("INVARIANT VIOLATION: Operation=TimeTraveling but timeTravelInfo=nil in description")
		}
		if a.timeTravelInfo.CurrentCommit.Time.IsZero() {
			panic("INVARIANT VIOLATION: TimeTraveling commit time is zero")
		}

		rightDesc = a.timeTravelInfo.CurrentCommit.Time.Format("Mon, 2 Jan 2006 15:04:05")
	} else {
		// Show timeline description (if applicable)
		if state.Timeline == "" {
			rightDesc = "No remote configured."
		} else {
			tlInfo := a.timelineInfo[state.Timeline]
			rightDesc = tlInfo.Description(state.CommitsAhead, state.CommitsBehind)
		}
	}

	row5 := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().
			Width(halfWidth).
			Foreground(lipgloss.Color(a.theme.ContentTextColor)).
			Render(wtDesc),
		lipgloss.NewStyle().
			Width(halfWidth).
			Foreground(lipgloss.Color(a.theme.ContentTextColor)).
			Render(rightDesc),
	)

	// Combine all rows
	headerContent := row1 + "\n" + row2 + "\n" + separatorLine + "\n" + row4 + "\n" + row5

	return ui.RenderBox(ui.BoxConfig{
		Content:     headerContent,
		InnerWidth:  ui.ContentInnerWidth,
		InnerHeight: ui.HeaderHeight,
		BorderColor: a.theme.BoxBorderColor,
		TextColor:   a.theme.LabelTextColor,
		Theme:       a.theme,
	})
}

// isInputMode checks if current mode accepts text input
func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeCloneURL ||
		(a.mode == ModeSetupWizard && a.setupWizardStep == SetupStepEmail)
}

// menuItemsToMaps converts MenuItem slice to map slice for rendering
// Note: Hint is excluded from maps (displayed in footer instead)
func (a *Application) menuItemsToMaps(items []MenuItem) []map[string]interface{} {
	maps := make([]map[string]interface{}, len(items))
	for i, item := range items {
		maps[i] = map[string]interface{}{
			"ID":        item.ID,
			"Shortcut":  item.Shortcut,
			"Emoji":     item.Emoji,
			"Label":     item.Label,
			"Enabled":   item.Enabled,
			"Separator": item.Separator,
		}
	}
	return maps
}

// buildKeyHandlers builds the complete handler registry for all modes
// Global handlers take priority and are merged into each mode
func (a *Application) buildKeyHandlers() map[AppMode]map[string]KeyHandler {
	// Global handlers - highest priority, applied to all modes
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"ctrl+v": a.handleKeyPaste, // Linux/Windows/macOS
		"cmd+v":  a.handleKeyPaste, // macOS cmd+v
		"meta+v": a.handleKeyPaste, // macOS meta (cmd) - Bubble Tea may send this
		"alt+v":  a.handleKeyPaste, // Fallback
	}

	cursorNavMixin := CursorNavigationMixin{}

	// Generic input cursor handlers for single-field inputs
	genericInputNav := cursorNavMixin.CreateHandlers(
		func(a *Application) string { return a.inputValue },
		func(a *Application) int { return a.inputCursorPosition },
		func(a *Application, pos int) { a.inputCursorPosition = pos },
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
func (a *Application) rebuildMenuShortcuts() {
	if a.keyHandlers[ModeMenu] == nil {
		a.keyHandlers[ModeMenu] = make(map[string]KeyHandler)
	}

	// Remove old shortcut handlers (keep navigation and enter)
	// We'll rebuild from scratch by first copying the base handlers
	baseHandlers := NewModeHandlers().
		WithMenuNav(a).
		On("enter", a.handleMenuEnter).
		Build()

	// Merge global handlers
	globalHandlers := map[string]KeyHandler{
		"ctrl+c": a.handleKeyCtrlC,
		"q":      a.handleKeyCtrlC,
		"esc":    a.handleKeyESC,
		"ctrl+v": a.handleKeyPaste,
		"cmd+v":  a.handleKeyPaste,
		"meta+v": a.handleKeyPaste,
		"alt+v":  a.handleKeyPaste,
	}

	// Start fresh
	newHandlers := make(map[string]KeyHandler)

	// Copy base handlers
	for key, handler := range baseHandlers {
		newHandlers[key] = handler
	}

	// Add global handlers
	for key, handler := range globalHandlers {
		newHandlers[key] = handler
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

	// Replace ModeMenu handlers
	a.keyHandlers[ModeMenu] = newHandlers
}

// handleMenuUp moves selection up
func (a *Application) handleMenuUp(app *Application) (tea.Model, tea.Cmd) {
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
	valueLen := len(a.inputValue)
	if a.inputCursorPosition < 0 {
		a.inputCursorPosition = 0
	}
	if a.inputCursorPosition > valueLen {
		a.inputCursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputValue[:a.inputCursorPosition]
	after := a.inputValue[a.inputCursorPosition:]
	a.inputValue = before + text + after
	a.inputCursorPosition += len(text)
}

// deleteAtCursor deletes character before cursor (UTF-8 safe)
func (a *Application) deleteAtCursor() {
	valueLen := len(a.inputValue)
	if a.inputCursorPosition <= 0 || valueLen == 0 {
		return
	}
	if a.inputCursorPosition > valueLen {
		a.inputCursorPosition = valueLen
	}

	// Safe slice operation
	before := a.inputValue[:a.inputCursorPosition-1]
	after := a.inputValue[a.inputCursorPosition:]
	a.inputValue = before + after
	a.inputCursorPosition--
}

// updateInputValidation updates validation message for current input
func (a *Application) updateInputValidation() {
	if a.inputAction == "clone_url" {
		currentValue := a.inputValue
		if a.mode == ModeInitializeBranches {
			return // No validation in branch mode
		}
		if currentValue == "" {
			a.inputValidationMsg = ""
		} else if ui.ValidateRemoteURL(currentValue) {
			a.inputValidationMsg = ""
		} else {
			a.inputValidationMsg = "Invalid URL format"
		}
	}
}

// Input mode handlers

// handleInputSubmit handles enter in generic input mode
func (a *Application) handleInputSubmit(app *Application) (tea.Model, tea.Cmd) {
	// UI THREAD - Route input submission based on action type
	switch app.inputAction {
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
	if app.historyState == nil || len(app.historyState.Commits) == 0 {
		return app, nil
	}

	if app.historyState.SelectedIdx < 0 || app.historyState.SelectedIdx >= len(app.historyState.Commits) {
		return app, nil
	}

	selectedCommit := app.historyState.Commits[app.historyState.SelectedIdx]
	app.pendingRewindCommit = selectedCommit.Hash

	return app, app.showRewindConfirmation(selectedCommit.Hash)
}
