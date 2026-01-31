package app

import (
	"fmt"
	"os"

	"tit/internal/config"
	"tit/internal/git"
	"tit/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

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
