package app

import (
	"fmt"
	"os"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
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
	ConfirmBranchSwitchClean     ConfirmationType = "branch_switch_clean"
	ConfirmBranchSwitchDirty     ConfirmationType = "branch_switch_dirty"
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
		Confirm: (*Application).executeConfirmTimeTravelMergeDirtyStash,    // YES = Stash
		Reject:  (*Application).executeConfirmTimeTravelReturnDirtyDiscard, // NO = Discard
	},
	"confirm_stale_stash_continue": {
		Confirm: (*Application).executeConfirmStaleStashContinue,
		Reject:  (*Application).executeRejectStaleStashContinue,
	},
	"confirm_stale_stash_merge_continue": {
		Confirm: (*Application).executeConfirmStaleStashMergeContinue,
		Reject:  (*Application).executeRejectStaleStashContinue,
	},
	string(ConfirmRewind): {
		Confirm: (*Application).executeConfirmRewind,
		Reject:  (*Application).executeRejectRewind,
	},
	string(ConfirmBranchSwitchClean): {
		Confirm: (*Application).executeConfirmBranchSwitchClean,
		Reject:  (*Application).executeRejectBranchSwitch,
	},
	string(ConfirmBranchSwitchDirty): {
		Confirm: (*Application).executeConfirmBranchSwitchDirty,
		Reject:  (*Application).executeRejectBranchSwitchDirty,
	},
}

// ========================================
// Confirmation Result Handler
// ========================================

// handleConfirmationResponse routes confirmation YES/NO responses to appropriate handlers
func (a *Application) handleConfirmationResponse(confirmed bool) (tea.Model, tea.Cmd) {
	if a.dialogState.GetDialog() == nil {
		// No active confirmation dialog
		return a.returnToMenu()
	}

	confirmType := a.dialogState.GetDialog().Config.ActionID
	actions, ok := confirmationHandlers[confirmType]
	if !ok {
		// No handler registered for this type - return to menu
		a.dialogState.Hide()
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
		a.dialogState.Hide()
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
	a.dialogState.Hide()
	a.mode = ModeInitializeLocation
	return a, nil
}

// executeRejectNestedRepoInit handles NO response to nested repo warning
func (a *Application) executeRejectNestedRepoInit() (tea.Model, tea.Cmd) {
	// User cancelled - abort init, return to menu
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmForcePush handles YES response to force push confirmation
func (a *Application) executeConfirmForcePush() (tea.Model, tea.Cmd) {
	// User confirmed force push
	// Initiate async push --force operation
	a.dialogState.Hide()
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdForcePush()
}

// executeRejectForcePush handles NO response to force push confirmation
func (a *Application) executeRejectForcePush() (tea.Model, tea.Cmd) {
	// User cancelled force push
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeConfirmHardReset handles YES response to hard reset confirmation
func (a *Application) executeConfirmHardReset() (tea.Model, tea.Cmd) {
	// User confirmed hard reset
	a.dialogState.Hide()
	a.prepareAsyncOperation(GetFooterMessageText(MessageOperationInProgress))
	return a, a.cmdHardReset()
}

// executeRejectHardReset handles NO response to hard reset confirmation
func (a *Application) executeRejectHardReset() (tea.Model, tea.Cmd) {
	// User cancelled hard reset
	a.dialogState.Hide()
	return a.returnToMenu()
}

// executeAlert handles alert dialog dismissal (any response)
func (a *Application) executeAlert() (tea.Model, tea.Cmd) {
	// Alert dialogs are dismissed with any key press
	a.dialogState.Hide()
	return a.returnToMenu()
}

// ========================================
// Confirmation Dialog Creation
// ========================================

// showConfirmation displays a confirmation dialog with the given config
// Sets confirmation mode and dialog state
func (a *Application) showConfirmation(config ui.ConfirmationConfig) {
	dialog := ui.NewConfirmationDialog(
		config,
		a.sizing.ContentInnerWidth,
		&a.theme,
	)
	a.dialogState.Show(dialog, nil)
	a.mode = ModeConfirmation
}

// prepareAsyncOperation consolidates the common async operation setup pattern
// Reduces 10+ lines of duplicate code across confirmation handlers
func (a *Application) prepareAsyncOperation(hint string) {
	a.startAsyncOp()
	a.consoleState.SetAutoScroll(true)
	a.mode = ModeConsole
	a.consoleState.Clear()
	a.consoleState.Reset()
	a.footerHint = hint
	a.workflowState.PreviousMode = ModeMenu
	a.workflowState.PreviousMenuIndex = 0
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
		panic("FATAL: cannot determine original branch")
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
