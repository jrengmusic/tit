package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"tit/internal/git"
	"tit/internal/ui"
)

// ConfirmationType represents different kinds of confirmation dialogs
type ConfirmationType string

const (
	ConfirmNestedRepoInit     ConfirmationType = "nested_repo_init"
	ConfirmForcePush          ConfirmationType = "force_push"
	ConfirmHardReset          ConfirmationType = "hard_reset"
	ConfirmDestructiveOp      ConfirmationType = "destructive_op"
	ConfirmAlert              ConfirmationType = "alert"
	ConfirmPullMerge          ConfirmationType = "pull_merge"
	ConfirmPullMergeDiverged  ConfirmationType = "pull_merge_diverged"
	ConfirmDirtyPull          ConfirmationType = "dirty_pull"
	ConfirmTimeTravel         ConfirmationType = "time_travel"
	ConfirmTimeTravelReturn   ConfirmationType = "time_travel_return"
	ConfirmTimeTravelMerge    ConfirmationType = "time_travel_merge"
)

// ConfirmationAction is a function that handles a confirmed action
type ConfirmationAction func(*Application) (tea.Model, tea.Cmd)

// confirmationActions maps confirmation types to their YES handlers
var confirmationActions = map[string]ConfirmationAction{
	string(ConfirmNestedRepoInit):     (*Application).executeConfirmNestedRepoInit,
	string(ConfirmForcePush):          (*Application).executeConfirmForcePush,
	string(ConfirmHardReset):          (*Application).executeConfirmHardReset,
	string(ConfirmAlert):              (*Application).executeAlert,
	string(ConfirmDirtyPull):          (*Application).executeConfirmDirtyPull,
	string(ConfirmPullMerge):          (*Application).executeConfirmPullMerge,
	string(ConfirmPullMergeDiverged):  (*Application).executeConfirmPullMerge,
	string(ConfirmTimeTravel):         (*Application).executeConfirmTimeTravel,
	string(ConfirmTimeTravelReturn):   (*Application).executeConfirmTimeTravelReturn,
	string(ConfirmTimeTravelMerge):    (*Application).executeConfirmTimeTravelMerge,
}

// confirmationRejectActions maps confirmation types to their NO handlers
var confirmationRejectActions = map[string]ConfirmationAction{
	string(ConfirmNestedRepoInit):     (*Application).executeRejectNestedRepoInit,
	string(ConfirmForcePush):          (*Application).executeRejectForcePush,
	string(ConfirmHardReset):          (*Application).executeRejectHardReset,
	string(ConfirmAlert):              (*Application).executeAlert, // Any key dismisses alert
	string(ConfirmDirtyPull):          (*Application).executeRejectDirtyPull,
	string(ConfirmPullMerge):          (*Application).executeRejectPullMerge,
	string(ConfirmPullMergeDiverged):  (*Application).executeRejectPullMerge,
	string(ConfirmTimeTravel):         (*Application).executeRejectTimeTravel,
	string(ConfirmTimeTravelReturn):   (*Application).executeRejectTimeTravelReturn,
	string(ConfirmTimeTravelMerge):    (*Application).executeRejectTimeTravelMerge,
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
	var handler ConfirmationAction

	if confirmed {
		handler = confirmationActions[confirmType]
	} else {
		handler = confirmationRejectActions[confirmType]
	}

	if handler == nil {
		// No handler registered for this type - return to menu
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
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)

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
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)

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
		ui.ContentInnerWidth,
		&a.theme,
	)
	a.mode = ModeConfirmation
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
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

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
	a.asyncOperationActive = true
	a.asyncOperationAborted = false
	a.consoleAutoScroll = true
	a.mode = ModeConsole
	a.outputBuffer.Clear()
	a.consoleState.Reset()
	a.footerHint = GetFooterMessageText(MessageOperationInProgress)
	a.previousMode = ModeMenu
	a.previousMenuIndex = 0

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
	
	// CRITICAL: Set Operation to TimeTraveling IMMEDIATELY
	// This prevents Phase 0 restoration from triggering if app restarts during time travel
	a.gitState.Operation = git.TimeTraveling
	buffer.Append("[DEBUG] Set Operation=TimeTraveling to prevent restoration loop", ui.TypeStatus)
	
	// Get commit hash from context
	commitHash := a.confirmContext["commit_hash"]
	buffer.Append(fmt.Sprintf("[DEBUG] commitHash=%s", commitHash), ui.TypeStatus)
	
	// Get current branch (original branch before time travel)
	currentBranchResult := git.Execute("rev-parse", "--abbrev-ref", "HEAD")
	if !currentBranchResult.Success {
		a.footerHint = ErrorMessages["failed_get_current_branch"]
		buffer.Append("[DEBUG] Failed to get current branch", ui.TypeStderr)
		return a, nil
	}
	
	originalBranch := strings.TrimSpace(currentBranchResult.Stdout)
	buffer.Append(fmt.Sprintf("[DEBUG] originalBranch=%s, isDirty=%v", originalBranch, a.gitState.WorkingTree == git.Dirty), ui.TypeStatus)
	
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
		panic(fmt.Sprintf("FATAL: Failed to write time travel info: %v", err))
	}
	
	// Build TimeTravelInfo directly (fail fast if git calls fail)
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
	a.footerHint = FooterHints["time_traveling_status"]
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
	// Stash changes first
	stashResult := git.Execute("stash", "push", "-u", "-m", "TIT_TIME_TRAVEL")
	if !stashResult.Success {
		a.footerHint = ErrorMessages["failed_stash_changes"]
		return a, nil
	}
	
	// Get stash ID
	stashListResult := git.Execute("stash", "list")
	if !stashListResult.Success {
		a.footerHint = ErrorMessages["failed_get_stash_list"]
		return a, nil
	}
	
	// Find the stash we just created (should be stash@{0})
	stashID := ""
	lines := strings.Split(stashListResult.Stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "TIT_TIME_TRAVEL") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				stashID = parts[0]
				break
			}
		}
	}
	
	// Write time travel info with stash ID
	err := git.WriteTimeTravelInfo(originalBranch, stashID)
	if err != nil {
		panic(fmt.Sprintf("FATAL: Failed to write time travel info: %v", err))
	}
	
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
		OriginalStashID: stashID,
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
	a.footerHint = FooterHints["time_traveling_status"]
	a.previousMode = ModeHistory
	a.previousMenuIndex = 0

	// CRITICAL: Set restoreTimeTravelInitiated = true to prevent restoration check
	// from triggering during this intentional time travel session
	a.restoreTimeTravelInitiated = true

	// Start time travel checkout operation
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
	
	// Get original branch from time travel info
	originalBranch, _, err := git.GetTimeTravelInfo()
	if err != nil {
		a.footerHint = "Failed to get time travel info"
		return a, nil
	}
	
	// Execute time travel return operation
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
	
	// Get original branch from time travel info
	originalBranch, _, err := git.GetTimeTravelInfo()
	if err != nil {
		a.footerHint = "Failed to get time travel info"
		return a, nil
	}
	
	// Execute time travel merge operation
	return a, git.ExecuteTimeTravelMerge(originalBranch, timeTravelHash)
}

// executeRejectTimeTravelMerge handles NO response to merge-and-return confirmation
func (a *Application) executeRejectTimeTravelMerge() (tea.Model, tea.Cmd) {
	// User cancelled merge
	a.confirmationDialog = nil
	a.mode = ModeMenu
	return a, nil
}
