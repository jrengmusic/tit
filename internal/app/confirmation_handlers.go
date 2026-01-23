package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"
)

// ConfirmationType represents different kinds of confirmation dialogs
type ConfirmationType string

const (
	ConfirmNestedRepoInit        ConfirmationType = "nested_repo_init"
	ConfirmForcePush             ConfirmationType = "force_push"
	ConfirmHardReset             ConfirmationType = "hard_reset"
	ConfirmDestructiveOp         ConfirmationType = "destructive_op"
	ConfirmAlert                 ConfirmationType = "alert"
	ConfirmPullMerge             ConfirmationType = "pull_merge"
	ConfirmPullMergeDiverged     ConfirmationType = "pull_merge_diverged"
	ConfirmDirtyPull             ConfirmationType = "dirty_pull"
	ConfirmTimeTravel            ConfirmationType = "time_travel"
	ConfirmTimeTravelReturn      ConfirmationType = "time_travel_return"
	ConfirmTimeTravelMerge       ConfirmationType = "time_travel_merge"
	ConfirmTimeTravelMergeDirty  ConfirmationType = "time_travel_merge_dirty"
	ConfirmTimeTravelReturnDirty ConfirmationType = "time_travel_return_dirty"
	ConfirmRewind                ConfirmationType = "rewind"
)

// ConfirmationAction is a function that handles a confirmed action
type ConfirmationAction func(*Application) (tea.Model, tea.Cmd)

// ConfirmationActionPair pairs a YES handler with its NO handler
// This guarantees that every confirmation type has both handlers registered
type ConfirmationActionPair struct {
	Confirm ConfirmationAction
	Reject  ConfirmationAction
}

// confirmationHandlers maps confirmation types to paired YES/NO handlers
// Using a single paired structure ensures no accidental missing confirm/reject pairs
var confirmationHandlers = map[string]ConfirmationActionPair{
	string(ConfirmNestedRepoInit): {
		Confirm: (*Application).executeConfirmNestedRepoInit,
		Reject:  (*Application).executeRejectNestedRepoInit,
	},
	string(ConfirmForcePush): {
		Confirm: (*Application).executeConfirmForcePush,
		Reject:  (*Application).executeRejectForcePush,
	},
	string(ConfirmHardReset): {
		Confirm: (*Application).executeConfirmHardReset,
		Reject:  (*Application).executeRejectHardReset,
	},
	string(ConfirmAlert): {
		Confirm: (*Application).executeAlert,
		Reject:  (*Application).executeAlert, // Any key dismisses alert
	},
	string(ConfirmDirtyPull): {
		Confirm: (*Application).executeConfirmDirtyPull,
		Reject:  (*Application).executeRejectDirtyPull,
	},
	string(ConfirmPullMerge): {
		Confirm: (*Application).executeConfirmPullMerge,
		Reject:  (*Application).executeRejectPullMerge,
	},
	string(ConfirmPullMergeDiverged): {
		Confirm: (*Application).executeConfirmPullMerge,
		Reject:  (*Application).executeRejectPullMerge,
	},
	string(ConfirmTimeTravel): {
		Confirm: (*Application).executeConfirmTimeTravel,
		Reject:  (*Application).executeRejectTimeTravel,
	},
	string(ConfirmTimeTravelReturn): {
		Confirm: (*Application).executeConfirmTimeTravelReturn,
		Reject:  (*Application).executeRejectTimeTravelReturn,
	},
	string(ConfirmTimeTravelMerge): {
		Confirm: (*Application).executeConfirmTimeTravelMerge,
		Reject:  (*Application).executeRejectTimeTravelMerge,
	},
	string(ConfirmTimeTravelMergeDirty): {
		Confirm: (*Application).executeConfirmTimeTravelMergeDirtyCommit,
		Reject:  (*Application).executeConfirmTimeTravelMergeDirtyDiscard,
	},
	string(ConfirmTimeTravelReturnDirty): {
		Confirm: (*Application).executeConfirmTimeTravelReturnDirtyDiscard,
		Reject:  (*Application).executeRejectTimeTravelReturnDirty,
	},
	"time_travel_return_dirty_choice": {
		Confirm: (*Application).executeConfirmTimeTravelMergeDirtyCommit,   // YES = Merge
		Reject:  (*Application).executeConfirmTimeTravelReturnDirtyDiscard, // NO = Discard
	},
	string(ConfirmRewind): {
		Confirm: (*Application).executeConfirmRewind,
		Reject:  (*Application).executeRejectRewind,
	},
}

// ========================================
// Confirmation Result Handler
// ========================================

// handleConfirmationResponse routes confirmation YES/NO responses to appropriate handlers
func (a *Application) handleConfirmationResponse(confirmed bool) (tea.Model, tea.Cmd) {
	if a.confirmationDialog == nil {
		// No active confirmation dialog
		return a.returnToMenu()
	}

	confirmType := a.confirmationDialog.Config.ActionID
	actions, ok := confirmationHandlers[confirmType]
	if !ok {
		// No handler registered for this type - return to menu
		a.confirmationDialog = nil
		return a.returnToMenu()
	}

	var handler ConfirmationAction
	if confirmed {
		handler = actions.Confirm
	} else {
		handler = actions.Reject
	}

	if handler == nil {
		// Paired handler is nil - return to menu (shouldn't happen with paired structure)
		a.confirmationDialog = nil
		return a.returnToMenu()
	}

	return handler(a)
}

// ========================================
// Confirmation Action Methods
// ========================================

// executeConfirmNestedRepoInit handles YES response to nested repo warning
func (a *Application) executeConfirmNestedRepoInit() (tea.Model, tea.Cmd) {
	// User confirmed they want to init in a nested repo
	// Return to previous mode (ModeInitializeLocation)
	a.confirmationDialog = nil
	a.mode = ModeInitializeLocation
	return a, nil
}

// executeRejectNestedRepoInit handles NO response to nested repo warning
func (a *Application) executeRejectNestedRepoInit() (tea.Model, tea.Cmd) {
	// User cancelled - abort init, return to menu
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeConfirmForcePush handles YES response to force push confirmation
func (a *Application) executeConfirmForcePush() (tea.Model, tea.Cmd) {
	// User confirmed force push
	// Initiate async push --force operation
	a.confirmationDialog = nil
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdForcePush()
}

// executeRejectForcePush handles NO response to force push confirmation
func (a *Application) executeRejectForcePush() (tea.Model, tea.Cmd) {
	// User cancelled force push
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeConfirmHardReset handles YES response to hard reset confirmation
func (a *Application) executeConfirmHardReset() (tea.Model, tea.Cmd) {
	// User confirmed hard reset
	a.confirmationDialog = nil
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdHardReset()
}

// executeRejectHardReset handles NO response to hard reset confirmation
func (a *Application) executeRejectHardReset() (tea.Model, tea.Cmd) {
	// User cancelled hard reset
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// executeAlert handles alert dialog dismissal (any response)
func (a *Application) executeAlert() (tea.Model, tea.Cmd) {
	// Alert dialogs are dismissed with any key press
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// ========================================
// Confirmation Dialog Creation
// ========================================

// showConfirmation displays a confirmation dialog with the given config
// Sets confirmation mode and dialog state
func (a *Application) showConfirmation(config ui.ConfirmationConfig) {
	a.confirmationDialog = ui.NewConfirmationDialog(
		config,
		a.sizing.ContentInnerWidth,
		&a.theme,
	)
	a.mode = ModeConfirmation
}

// prepareAsyncOperation consolidates the common async operation setup pattern
// Reduces 10+ lines of duplicate code across confirmation handlers
func (a *Application) prepareAsyncOperation(hint string) {
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = hint
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0
}

// getOriginalBranchForTimeTravel retrieves the original branch for time travel operations
// Checks config stash entry first, then falls back to time travel marker file
func (a *Application) getOriginalBranchForTimeTravel() string {
	repoPath, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
	}

	entry, found := config.GetStashEntry("time_travel", repoPath)
	if found {
		return entry.OriginalBranch
	}

	originalBranch, _, err := git.GetTimeTravelInfo()
	if err != nil || originalBranch == "" {
		panic("FATAL: Cannot determine original branch")
	}
	return originalBranch
}

// discardWorkingTreeChanges resets working tree to HEAD and cleans untracked files
// Returns error if reset fails; clean failures are non-fatal (warning set in footer)
func (a *Application) discardWorkingTreeChanges() error {
	resetResult := git.Execute("reset", "--hard", "HEAD")
	if !resetResult.Success {
		return fmt.Errorf("failed to discard changes: %s", resetResult.Stderr)
	}

	cleanResult := git.Execute("clean", "-fd")
	if !cleanResult.Success {
		a.footerHint = "Warning: Failed to clean untracked files"
	}
	return nil
}

// showConfirmationFromMessage displays a confirmation dialog by looking up SSOT messages
// Consolidates the common pattern: lookup title/explanation/labels → build config → show dialog
// Usage: a.showConfirmationFromMessage(ConfirmForcePush, "")
// Or with custom args: a.showConfirmationFromMessage(ConfirmTimeTravel, fmt.Sprintf("...%s", arg))
func (a *Application) showConfirmationFromMessage(confirmType ConfirmationType, customExplanation string) {
	actionID := string(confirmType)

	// Build config using SSOT messages
	msg := ConfirmationMessages[actionID]
	config := ui.ConfirmationConfig{
		Title:    msg.Title,
		ActionID: actionID,
	}

	// Use custom explanation if provided, otherwise use SSOT
	if customExplanation != "" {
		config.Explanation = customExplanation
	} else {
		config.Explanation = msg.Explanation
	}

	// Load labels (YES/NO button text)
	config.YesLabel = msg.YesLabel
	config.NoLabel = msg.NoLabel

	a.showConfirmation(config)
}

// showRewindConfirmation displays rewind confirmation dialog
// showRewindConfirmation displays destructive rewind confirmation
// Uses ConfirmationMessages map for centralized message management
func (a *Application) showRewindConfirmation(commitHash string) tea.Cmd {
	msg := ConfirmationMessages["rewind"]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: fmt.Sprintf(msg.Explanation, ui.ShortenHash(commitHash)),
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    string(ConfirmRewind),
	}
	a.showConfirmation(config)
	return nil
}

// showNestedRepoWarning displays warning when initializing in a nested repo
func (a *Application) showNestedRepoWarning(path string) {
	config := ui.ConfirmationConfig{
		Title:       "Nested Repository Detected",
		Explanation: fmt.Sprintf("The directory '%s' is inside an existing git repository. Initialize anyway?", path),
		YesLabel:    "Yes, continue",
		NoLabel:     "No, cancel",
		ActionID:    string(ConfirmNestedRepoInit),
	}
	a.showConfirmation(config)
}

// showForcePushWarning displays warning when attempting force push
func (a *Application) showForcePushWarning(branchName string) {
	config := ui.ConfirmationConfig{
		Title:       "Force Push Confirmation",
		Explanation: fmt.Sprintf("This will force push '%s' to remote. Any remote changes will be overwritten. Continue?", branchName),
		YesLabel:    "Force push",
		NoLabel:     "Cancel",
		ActionID:    string(ConfirmForcePush),
	}
	a.showConfirmation(config)
}

// showHardResetWarning displays warning when attempting hard reset
func (a *Application) showHardResetWarning() {
	config := ui.ConfirmationConfig{
		Title:       "Hard Reset Confirmation",
		Explanation: "This will discard all uncommitted changes. This cannot be undone. Continue?",
		YesLabel:    "Reset",
		NoLabel:     "Cancel",
		ActionID:    string(ConfirmHardReset),
	}
	a.showConfirmation(config)
}

// showAlert displays a simple alert dialog (informational, dismisses with any key)
func (a *Application) showAlert(title, explanation string) {
	config := ui.ConfirmationConfig{
		Title:       title,
		Explanation: explanation,
		YesLabel:    "OK",
		NoLabel:     "", // Single button alert
		ActionID:    string(ConfirmAlert),
	}
	a.showConfirmation(config)
}

// ========================================
// Dirty Pull Confirmation Handlers
// ========================================

// executeConfirmDirtyPull handles YES response to dirty pull confirmation (Save changes)
func (a *Application) executeConfirmDirtyPull() (tea.Model, tea.Cmd) {
	// User confirmed to save changes and proceed with dirty pull
	a.confirmationDialog = nil

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", true) // true = preserve changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))

	// Start the operation chain - Phase 1: Snapshot
	return a, a.cmdDirtyPullSnapshot(true)
}

// executeRejectDirtyPull handles NO response to dirty pull confirmation (Discard changes)
func (a *Application) executeRejectDirtyPull() (tea.Model, tea.Cmd) {
	// User chose to discard changes and pull
	a.confirmationDialog = nil

	// Create operation state - merge strategy only
	a.dirtyOperationState = NewDirtyOperationState("dirty_pull_merge", false) // false = discard changes
	a.dirtyOperationState.PullStrategy = "merge"

	// Transition to console to show streaming output
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))

	// Start the operation chain - Phase 1: Snapshot (will discard instead of stash)
	return a, a.cmdDirtyPullSnapshot(false)
}

// ========================================
// Pull Merge Confirmation Handlers
// ========================================

// executeConfirmPullMerge handles YES response to pull merge confirmation
func (a *Application) executeConfirmPullMerge() (tea.Model, tea.Cmd) {
	// User confirmed to proceed with pull merge
	a.confirmationDialog = nil

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.isExitAllowed = false // Block Ctrl+C until operation completes or is aborted
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// Start pull operation with merge strategy (--no-rebase)
	return a, a.cmdPull()
}

// executeRejectPullMerge handles NO response to pull merge confirmation
func (a *Application) executeRejectPullMerge() (tea.Model, tea.Cmd) {
	// User cancelled pull merge
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// ========================================
// Time Travel Confirmation Handlers
// ========================================

// executeConfirmTimeTravel handles YES response to time travel confirmation
func (a *Application) executeConfirmTimeTravel() (tea.Model, tea.Cmd) {
	// User confirmed time travel
	a.confirmationDialog = nil
	buffer := ui.GetBuffer()
	buffer.Append("[DEBUG] executeConfirmTimeTravel called", ui.TypeStatus)

	// Get commit hash from context
	commitHash := a.confirmContext["commit_hash"]
	buffer.Append(fmt.Sprintf("[DEBUG] commitHash=%s", commitHash), ui.TypeStatus)

	// Get original branch - CHECK BEFORE modifying Operation state
	// Case 1: Already time traveling (Operation == TimeTraveling) → read from TIT_TIME_TRAVEL file
	// Case 2: Normal operation → get current branch from HEAD
	var originalBranch string
	wasAlreadyTimeTraveling := a.gitState.Operation == git.TimeTraveling

	if wasAlreadyTimeTraveling {
		// Already time traveling - read original branch from marker file
		existingBranch, _, err := git.GetTimeTravelInfo()
		if err != nil {
			a.footerHint = ErrorMessages["failed_get_current_branch"]
			buffer.Append("[DEBUG] Failed to read TIT_TIME_TRAVEL file", ui.TypeStderr)
			return a, nil
		}
		originalBranch = existingBranch
		buffer.Append(fmt.Sprintf("[DEBUG] Already time traveling, originalBranch from file=%s", originalBranch), ui.TypeStatus)
	} else {
		// Normal operation - get current branch from HEAD
		currentBranchResult := git.Execute("rev-parse", "--abbrev-ref", "HEAD")
		if !currentBranchResult.Success {
			a.footerHint = ErrorMessages["failed_get_current_branch"]
			buffer.Append("[DEBUG] Failed to get current branch", ui.TypeStderr)
			return a, nil
		}
		originalBranch = strings.TrimSpace(currentBranchResult.Stdout)
		buffer.Append(fmt.Sprintf("[DEBUG] Normal operation, originalBranch from HEAD=%s", originalBranch), ui.TypeStatus)

		// CRITICAL: If at detached HEAD (originalBranch == "HEAD"), try to get actual branch
		if originalBranch == "HEAD" {
			// Try to get default branch from remote tracking
			defaultBranchResult := git.Execute("symbolic-ref", "refs/remotes/origin/HEAD")
			if defaultBranchResult.Success {
				// Output is like "refs/remotes/origin/main", extract "main"
				parts := strings.Split(strings.TrimSpace(defaultBranchResult.Stdout), "/")
				if len(parts) > 0 {
					originalBranch = parts[len(parts)-1]
					buffer.Append(fmt.Sprintf("[DEBUG] Detached HEAD, using remote tracking branch: %s", originalBranch), ui.TypeStatus)
				}
			} else {
				// Fallback to "main" (most common default)
				originalBranch = "main"
				buffer.Append(fmt.Sprintf("[DEBUG] Detached HEAD, using fallback: %s", originalBranch), ui.TypeStatus)
			}
		}
	}

	// CRITICAL: Set Operation to TimeTraveling AFTER getting original branch
	// This prevents Phase 0 restoration from triggering if app restarts during time travel
	a.gitState.Operation = git.TimeTraveling
	buffer.Append("[DEBUG] Set Operation=TimeTraveling to prevent restoration loop", ui.TypeStatus)

	buffer.Append(fmt.Sprintf("[DEBUG] Final originalBranch=%s, isDirty=%v", originalBranch, a.gitState.WorkingTree == git.Dirty), ui.TypeStatus)

	// Check if working tree is dirty
	if a.gitState.WorkingTree == git.Dirty {
		buffer.Append("[DEBUG] Dirty tree - calling executeTimeTravelWithDirtyTree", ui.TypeStatus)
		// Handle dirty working tree - stash changes first
		return a.executeTimeTravelWithDirtyTree(originalBranch, commitHash)
	} else {
		buffer.Append("[DEBUG] Clean tree - calling executeTimeTravelClean", ui.TypeStatus)
		// Clean working tree - proceed directly
		return a.executeTimeTravelClean(originalBranch, commitHash)
	}
}

// executeTimeTravelClean handles time travel from clean working tree
func (a *Application) executeTimeTravelClean(originalBranch, commitHash string) (tea.Model, tea.Cmd) {
	// Write time travel info (no stash ID for clean tree)
	err := git.WriteTimeTravelInfo(originalBranch, "")
	if err != nil {
		// Use standardized fatal error logging (PATTERN: Invariant violation)
		LogErrorFatal("Failed to write time travel info", err)
	}

	// Build TimeTravelInfo directly (fail fast if git calls fail)
	// This prevents silent failures and inconsistent state
	commitSubject := strings.TrimSpace(git.Execute("log", "-1", "--format=%s", commitHash).Stdout)
	if commitSubject == "" {
		LogErrorFatal("Failed to get commit subject", fmt.Errorf("empty subject for %s", commitHash))
	}

	commitTimeStr := strings.TrimSpace(git.Execute("log", "-1", "--format=%aI", commitHash).Stdout)
	if commitTimeStr == "" {
		LogErrorFatal("Failed to get commit time", fmt.Errorf("empty time for %s", commitHash))
	}

	commitTime, err := time.Parse(time.RFC3339, commitTimeStr)
	if err != nil {
		LogErrorFatal("Failed to parse commit time", err)
	}

	a.timeTravelInfo = &git.TimeTravelInfo{
		OriginalBranch:  originalBranch,
		OriginalStashID: "",
		CurrentCommit: git.CommitInfo{
			Hash:    commitHash,
			Subject: commitSubject,
			Time:    commitTime,
		},
	}

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = LegacyFooterHints["time_traveling_status"]
	a.previousMode = ModeHistory
	a.previousMenuIndex = 0

	// CRITICAL: Set restoreTimeTravelInitiated = true to prevent restoration check
	// from triggering during this intentional time travel session
	a.restoreTimeTravelInitiated = true

	// Start time travel checkout operation
	return a, git.ExecuteTimeTravelCheckout(originalBranch, commitHash)
}

// executeTimeTravelWithDirtyTree handles time travel from dirty working tree
func (a *Application) executeTimeTravelWithDirtyTree(originalBranch, commitHash string) (tea.Model, tea.Cmd) {
	// Transition to console to show streaming output (MUST happen BEFORE buffer operations)
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = LegacyFooterHints["time_traveling_status"]
	a.previousMode = ModeHistory
	a.previousMenuIndex = 0
	a.restoreTimeTravelInitiated = true

	buffer := ui.GetBuffer()
	buffer.Append("[DEBUG] executeTimeTravelWithDirtyTree called", ui.TypeStatus)

	// Stash changes first
	buffer.Append("Stashing changes...", ui.TypeStatus)
	stashResult := git.Execute("stash", "push", "-u", "-m", "TIT_TIME_TRAVEL")
	if !stashResult.Success {
		buffer.Append(fmt.Sprintf("Failed to stash changes: %s", stashResult.Stderr), ui.TypeStderr)
		a.footerHint = ErrorMessages["failed_stash_changes"]
		a.asyncOperationActive = false
		return a, nil
	}

	buffer.Append("Stash created successfully", ui.TypeStatus)

	// Get stash ID
	stashListResult := git.Execute("stash", "list")
	if !stashListResult.Success {
		buffer.Append(fmt.Sprintf("Failed to get stash list: %s", stashListResult.Stderr), ui.TypeStderr)
		a.footerHint = ErrorMessages["failed_get_stash_list"]
		a.asyncOperationActive = false
		return a, nil
	}

	// Find the stash we just created (should be stash@{0})
	stashRef := ""
	lines := strings.Split(stashListResult.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "TIT_TIME_TRAVEL") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				// Remove trailing colon from stash reference (e.g., "stash@{0}:" → "stash@{0}")
				stashRef = strings.TrimSuffix(parts[0], ":")
				buffer.Append(fmt.Sprintf("[DEBUG] Found stash reference: %s", stashRef), ui.TypeStatus)
				break
			}
		}
	}

	if stashRef == "" {
		buffer.Append("[DEBUG] WARNING: No stash found with TIT_TIME_TRAVEL message!", ui.TypeStderr)
	}

	// Convert stash reference to SHA hash (stable, doesn't shift like stash@{N})
	// This prevents bugs when stash indices shift due to new stashes being created
	if stashRef == "" {
		panic("FATAL: No stash reference found after creating stash. This should never happen.")
	}

	buffer.Append(fmt.Sprintf("[DEBUG] Converting %s to hash...", stashRef), ui.TypeStatus)
	hashResult := git.Execute("rev-parse", stashRef)
	if !hashResult.Success {
		panic(fmt.Sprintf("FATAL: Failed to convert stash reference to hash: %s", hashResult.Stderr))
	}

	stashHash := strings.TrimSpace(hashResult.Stdout)
	buffer.Append(fmt.Sprintf("[DEBUG] Converted to hash: %s", stashHash), ui.TypeStatus)

	// Get current working directory (absolute repo path)
	repoPath, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to get current working directory: %v", err))
	}

	// Add stash entry to config tracking system
	buffer.Append(fmt.Sprintf("[DEBUG] Adding stash entry - operation: time_travel, repo: %s, hash: %s", repoPath, stashHash), ui.TypeStatus)
	config.AddStashEntry("time_travel", stashHash, repoPath, originalBranch, commitHash)

	// Build TimeTravelInfo directly from commit hash (fail fast if git calls fail)
	// This prevents silent failures and inconsistent state
	commitSubject := strings.TrimSpace(git.Execute("log", "-1", "--format=%s", commitHash).Stdout)
	if commitSubject == "" {
		panic(fmt.Sprintf("FATAL: Failed to get commit subject for %s", commitHash))
	}

	commitTimeStr := strings.TrimSpace(git.Execute("log", "-1", "--format=%aI", commitHash).Stdout)
	if commitTimeStr == "" {
		panic(fmt.Sprintf("FATAL: Failed to get commit time for %s", commitHash))
	}

	commitTime, err := time.Parse(time.RFC3339, commitTimeStr)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to parse commit time: %v", err))
	}

	a.timeTravelInfo = &git.TimeTravelInfo{
		OriginalBranch:  originalBranch,
		OriginalStashID: stashHash,
		CurrentCommit: git.CommitInfo{
			Hash:    commitHash,
			Subject: commitSubject,
			Time:    commitTime,
		},
	}

	// Start time travel checkout operation (console already set up at function start)
	return a, git.ExecuteTimeTravelCheckout(originalBranch, commitHash)
}

// executeRejectTimeTravel handles NO response to time travel confirmation
func (a *Application) executeRejectTimeTravel() (tea.Model, tea.Cmd) {
	// User cancelled time travel
	a.confirmationDialog = nil
	return a.returnToMenu()
}

// ========================================
// Time Travel Return Confirmation
// ========================================

// executeConfirmTimeTravelReturn handles YES response to return-to-main confirmation
func (a *Application) executeConfirmTimeTravelReturn() (tea.Model, tea.Cmd) {
	a.confirmationDialog = nil

	// Transition to console to show streaming output (consistent with other git operations)
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = "Returning to main..."
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel return!
	a.restoreTimeTravelInitiated = true

	// Get original branch and execute time travel return operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelReturn(originalBranch)
}

// executeRejectTimeTravelReturn handles NO response to return-to-main confirmation
func (a *Application) executeRejectTimeTravelReturn() (tea.Model, tea.Cmd) {
	// User cancelled return
	a.confirmationDialog = nil
	a.mode = ModeMenu
	return a, nil
}

// ========================================
// Time Travel Merge Confirmation
// ========================================

// executeConfirmTimeTravelMerge handles YES response to merge-and-return confirmation
func (a *Application) executeConfirmTimeTravelMerge() (tea.Model, tea.Cmd) {
	a.confirmationDialog = nil

	// Get current commit hash
	result := git.Execute("rev-parse", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		return a, nil
	}

	timeTravelHash := strings.TrimSpace(result.Stdout)

	// Transition to console to show streaming output (consistent with other git operations)
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	a.restoreTimeTravelInitiated = true

	// Get original branch and execute time travel merge operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// executeRejectTimeTravelMerge handles NO response to merge-and-return confirmation
func (a *Application) executeRejectTimeTravelMerge() (tea.Model, tea.Cmd) {
	// User cancelled merge
	a.confirmationDialog = nil
	a.mode = ModeMenu
	return a, nil
}

// ========================================
// Time Travel Merge Dirty Confirmation
// ========================================

// executeConfirmTimeTravelMergeDirtyCommit handles "Commit & merge" choice
// Auto-commits with generated message and immediately starts merge
func (a *Application) executeConfirmTimeTravelMergeDirtyCommit() (tea.Model, tea.Cmd) {
	a.confirmationDialog = nil

	buffer := ui.GetBuffer()
	buffer.Append("[DEBUG] === executeConfirmTimeTravelMergeDirtyCommit START ===", ui.TypeStatus)

	// Get current commit hash for auto-generated message
	result := git.Execute("rev-parse", "--short", "HEAD")
	buffer.Append(fmt.Sprintf("[DEBUG] Get short hash: Success=%v, Stdout=%s", result.Success, result.Stdout), ui.TypeStatus)
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, nil
	}
	shortHash := strings.TrimSpace(result.Stdout)

	// Auto-generate commit message
	commitMessage := fmt.Sprintf("TIT: Changes from time travel to %s", shortHash)
	buffer.Append(fmt.Sprintf("[DEBUG] Auto-generated commit message: %s", commitMessage), ui.TypeStatus)

	// Stage all changes (same as normal commit workflow)
	stageResult := git.Execute("add", "-A")
	buffer.Append(fmt.Sprintf("[DEBUG] Stage result: Success=%v", stageResult.Success), ui.TypeStatus)
	if !stageResult.Success {
		a.footerHint = fmt.Sprintf("Failed to stage changes: %s", stageResult.Stderr)
		a.mode = ModeMenu
		return a, nil
	}

	// Commit changes immediately
	commitResult := git.Execute("commit", "-m", commitMessage)
	buffer.Append(fmt.Sprintf("[DEBUG] Commit result: Success=%v", commitResult.Success), ui.TypeStatus)
	if !commitResult.Success {
		a.footerHint = fmt.Sprintf("Failed to commit: %s", commitResult.Stderr)
		a.mode = ModeMenu
		return a, nil
	}

	// Tree is now clean and changes committed - proceed directly with merge
	// Get full commit hash for merge operation
	fullHashResult := git.Execute("rev-parse", "HEAD")
	if !fullHashResult.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, nil
	}
	timeTravelHash := strings.TrimSpace(fullHashResult.Stdout)

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	// Marker file still exists during merge, but this is NOT an incomplete session
	a.restoreTimeTravelInitiated = true

	// Get original branch and execute time travel merge operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// executeConfirmTimeTravelMergeDirtyDiscard handles "Discard" choice
// Discards uncommitted changes immediately, then proceeds with merge
func (a *Application) executeConfirmTimeTravelMergeDirtyDiscard() (tea.Model, tea.Cmd) {
	a.confirmationDialog = nil

	// Discard working tree changes
	if err := a.discardWorkingTreeChanges(); err != nil {
		a.footerHint = err.Error()
		a.mode = ModeMenu
		return a, nil
	}

	// Tree is now clean - proceed directly with merge (no second confirmation)
	// Get current commit hash
	result := git.Execute("rev-parse", "HEAD")
	if !result.Success {
		a.footerHint = "Failed to get current commit"
		a.mode = ModeMenu
		return a, nil
	}

	timeTravelHash := strings.TrimSpace(result.Stdout)

	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = "Merging back to main..."
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel merge!
	a.restoreTimeTravelInitiated = true

	// Get original branch and execute time travel merge operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// ========================================
// Time Travel Return Dirty Confirmation
// ========================================

// executeConfirmTimeTravelReturnDirtyDiscard handles "Discard & return" choice
// Discards uncommitted changes immediately, then proceeds with return
func (a *Application) executeConfirmTimeTravelReturnDirtyDiscard() (tea.Model, tea.Cmd) {
	a.confirmationDialog = nil

	// Discard working tree changes
	if err := a.discardWorkingTreeChanges(); err != nil {
		a.footerHint = err.Error()
		a.mode = ModeMenu
		return a, nil
	}

	// Tree is now clean - proceed directly with return (no second confirmation)
	// Transition to console to show streaming output
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = "Returning to main..."
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

	// CRITICAL: Prevent restoration check from triggering during time travel return!
	a.restoreTimeTravelInitiated = true

	// Get original branch and execute time travel return operation
	originalBranch := a.getOriginalBranchForTimeTravel()
	return a, git.ExecuteTimeTravelReturn(originalBranch)
}

// executeRejectTimeTravelReturnDirty handles "Cancel" choice
func (a *Application) executeRejectTimeTravelReturnDirty() (tea.Model, tea.Cmd) {
	// User cancelled return
	a.confirmationDialog = nil
	a.mode = ModeMenu
	return a, nil
}

// executeConfirmRewind handles "Rewind" choice
// executeConfirmRewind executes git reset --hard at pending commit
func (a *Application) executeConfirmRewind() (tea.Model, tea.Cmd) {
	if a.pendingRewindCommit == "" {
		return a.returnToMenu()
	}

	commitHash := a.pendingRewindCommit
	a.pendingRewindCommit = "" // Clear after capturing

	// Set up async operation
	a.asyncOperationActive = true
	a.previousMode = ModeHistory
	a.previousMenuIndex = a.historyState.SelectedIdx
	a.mode = ModeConsole
	a.consoleState.Reset()
	ui.GetBuffer().Clear()

	// Start rewind + refresh ticker
	return a, tea.Batch(
		a.executeRewindOperation(commitHash),
		a.cmdRefreshConsole(),
	)
}

// executeRejectRewind handles "Cancel" choice on rewind confirmation
func (a *Application) executeRejectRewind() (tea.Model, tea.Cmd) {
	a.pendingRewindCommit = "" // Clear pending commit
	return a.returnToMenu()
}

// executeRewindOperation performs the actual git reset --hard in a worker goroutine
func (a *Application) executeRewindOperation(commitHash string) tea.Cmd {
	return func() tea.Msg {
		_, err := git.ResetHardAtCommit(commitHash)

		return RewindMsg{
			Commit:  commitHash,
			Success: err == nil,
			Error:   errorOrEmpty(err),
		}
	}
}
