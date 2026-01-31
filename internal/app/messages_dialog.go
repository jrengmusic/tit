package app

// ========================================
// Message Domain: Confirmation Dialogs
// ========================================

// ConfirmationMessage pairs all confirmation dialog components
type ConfirmationMessage struct {
	Title       string
	Explanation string
	YesLabel    string // Button text for YES/confirm action
	NoLabel     string // Button text for NO/reject action
}

// ConfirmationMessages centralizes all confirmation dialog messages by domain
// Replaces old ConfirmationTitles + ConfirmationExplanations + ConfirmationLabels
var ConfirmationMessages = map[string]ConfirmationMessage{
	"force_push": {
		Title:       "Force Push Confirmation",
		Explanation: "This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
		YesLabel:    "Force push",
		NoLabel:     "Cancel",
	},
	"hard_reset": {
		Title:       "Replace Local Confirmation",
		Explanation: "This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
		YesLabel:    "Reset to remote",
		NoLabel:     "Cancel",
	},
	"dirty_pull": {
		Title:       "Save your changes?",
		Explanation: "You have uncommitted changes. To pull, they must be temporarily saved.\n\nAfter the pull, we'll try to reapply them.\n(This may cause conflicts if the changes overlap.)",
		YesLabel:    "Save changes",
		NoLabel:     "Discard changes",
	},
	"pull_merge": {
		Title:       "Pull from remote?",
		Explanation: "This will merge remote changes into your local branch.\n\nIf both branches modified the same files, conflicts may occur.\nYou'll be able to resolve them interactively.",
		YesLabel:    "Proceed",
		NoLabel:     "Cancel",
	},
	"pull_merge_diverged": {
		Title:       "Pull diverged branches?",
		Explanation: "Your branches have diverged (both have new commits).\n\nThis will merge remote changes into your local branch.\n\nIf both modified the same files, conflicts may occur.",
		YesLabel:    "Proceed",
		NoLabel:     "Cancel",
	},
	"branch_switch_clean": {
		Title:       "Switch to {targetBranch}?",
		Explanation: "Current branch: {currentBranch}\nWorking tree: clean\n\nReady to switch?",
		YesLabel:    "Switch",
		NoLabel:     "Cancel",
	},
	"branch_switch_dirty": {
		Title:       "Switch to {targetBranch} with uncommitted changes?",
		Explanation: "Current branch: {currentBranch}\nWorking tree: dirty\n\nYour changes must be saved or discarded before switching.\n\nChoose action:",
		YesLabel:    "Stash changes",
		NoLabel:     "Discard changes",
	},
	"time_travel": {
		Title:       "Time Travel Confirmation",
		Explanation: "%s\n\n%s\n\nExplore in read-only mode?",
		YesLabel:    "Time travel",
		NoLabel:     "Cancel",
	},
	"time_travel_return": {
		Title:       "Return to main without merge?",
		Explanation: "Any changes you made while time traveling will be STASHED (not discarded).\n\nYour original work (if any) will be restored.\n\nUse 'git stash apply stash@{0}' later to restore time travel changes.",
		YesLabel:    "Return to main",
		NoLabel:     "Cancel",
	},
	"time_travel_merge": {
		Title:       "Merge and return to main?",
		Explanation: "This will merge time travel changes back to main.\n\nConflicts may occur if the changes overlap.\n\nNote: Any uncommitted changes will be stashed first, then restored after merge.",
		YesLabel:    "Merge & return",
		NoLabel:     "Cancel",
	},
	"time_travel_merge_dirty": {
		Title:       "Uncommitted Changes",
		Explanation: "You modified files during time travel.\n\nCommit them and merge to main, or discard them?",
		YesLabel:    "Commit & merge",
		NoLabel:     "Discard",
	},
	"time_travel_return_dirty": {
		Title:       "Uncommitted Changes",
		Explanation: "You modified files during time travel.\n\nChanges will be discarded when returning to main.",
		YesLabel:    "Discard & return",
		NoLabel:     "Cancel",
	},
	"rewind": {
		Title:       "DESTRUCTIVE OPERATION",
		Explanation: "This will discard all commits after %s.\nAny uncommitted changes will be lost.\n\nAre you sure you want to continue?",
		YesLabel:    "Rewind",
		NoLabel:     "Cancel",
	},
}

// DialogMessages centralizes dialog box content (titles + explanations)
var DialogMessages = map[string][2]string{
	"nested_repo": {
		"Nested Repository Detected",
		"The directory '%s' is inside another git repository.\n\nThis may cause confusion. Would you like to initialize in a subdirectory instead?",
	},
	"force_push_nested": {
		"Force Push Confirmation",
		"This will force push to remote, overwriting remote history.\n\nAny commits on the remote that you don't have locally will be permanently lost.\n\nContinue?",
	},
	"hard_reset_nested": {
		"Hard Reset Confirmation",
		"This will discard all local changes and commits, resetting to match the remote exactly.\n\nAll uncommitted changes and untracked files will be permanently lost.\n\nContinue?",
	},
}
