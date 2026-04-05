package app

import (
	"fmt"
	"os"

	"tit/internal/git"
	"tit/internal/ui"
)

// RenderStateHeader renders the state header per REACTIVE-LAYOUT-PLAN.md
func (a *Application) RenderStateHeader() string {
	state := a.gitState

	if state == nil || state.Operation == git.NotRepo {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		cwd = "unknown" // Graceful fallback
	}

	remoteURL := "🔌 NO REMOTE"
	remoteColor := a.theme.DimmedTextColor
	if state.Remote == git.HasRemote {
		url := git.GetRemoteURL()
		if url != "" {
			remoteURL = "🔗 " + url
			remoteColor = a.theme.AccentTextColor
		}
	}

	wtInfo := a.workingTreeInfo[state.WorkingTree]
	wtDesc := []string{wtInfo.Description(state.CommitsAhead, state.CommitsBehind)}

	// OMP-style: append modified count to status label
	workingTreeLabel := wtInfo.Label
	if state.WorkingTree == git.Dirty && state.ModifiedCount > 0 {
		workingTreeLabel = wtInfo.Label + " ● " + fmt.Sprintf("%d", state.ModifiedCount)
	}

	timelineEmoji := "🔌"
	timelineLabel := "N/A"
	timelineColor := a.theme.DimmedTextColor
	timelineDesc := []string{"No remote configured."}

	if state.Operation == git.TimeTraveling {
		if a.timeTravelState.info != nil {
			shortHash := a.timeTravelState.info.CurrentCommit.Hash
			if len(shortHash) >= 7 {
				shortHash = shortHash[:7]
			}
			timelineEmoji = "📌"
			timelineLabel = "DETACHED @ " + shortHash
			timelineColor = a.theme.OutputWarningColor
			timelineDesc = []string{"Viewing commit from " + a.timeTravelState.info.CurrentCommit.Time.Format("Jan 2, 2006")}
		}
	} else if state.Timeline != "" {
		tlInfo := a.timelineInfo[state.Timeline]
		timelineEmoji = tlInfo.Emoji
		timelineLabel = tlInfo.Label
		timelineColor = tlInfo.Color
		timelineDesc = []string{tlInfo.Description(state.CommitsAhead, state.CommitsBehind)}

		// OMP-style: append ahead/behind arrows with count
		if state.Timeline == git.Ahead && state.CommitsAhead > 0 {
			timelineLabel = tlInfo.Label + " ⬆ " + fmt.Sprintf("%d", state.CommitsAhead)
		} else if state.Timeline == git.Behind && state.CommitsBehind > 0 {
			timelineLabel = tlInfo.Label + " ⬇ " + fmt.Sprintf("%d", state.CommitsBehind)
		} else if state.Timeline == git.Diverged {
			if state.CommitsAhead > 0 || state.CommitsBehind > 0 {
				timelineLabel = tlInfo.Label + " ⬆ " + fmt.Sprintf("%d", state.CommitsAhead) + " ⬇ " + fmt.Sprintf("%d", state.CommitsBehind)
			}
		}
	}

	// Operation status (right column top)
	opInfo := a.operationInfo[state.Operation]

	// Branch name (right column bottom)
	branchName := state.CurrentBranch
	if branchName == "" {
		branchName = "N/A"
	}

	// SSOT: detached HEAD gets special icon and time travel color
	branchEmoji := "🌿"
	branchColor := a.theme.AccentTextColor

	// Manual detached HEAD (not TIT time travel): show DETACHED ops, hash in branch column
	if state.Detached && !state.IsTitTimeTravel {
		opInfo = StateInfo{
			Label: "DETACHED",
			Emoji: "",
			Color: a.theme.OutputWarningColor,
			Description: func(ahead, behind int) string {
				return "Not on a branch. Select 'Return to branch' to continue."
			},
		}
		// Branch column shows HASH with commit icon
		branchEmoji = ""
		branchColor = a.theme.AccentTextColor
		branchName = state.CurrentHash
	} else if state.Detached && state.IsTitTimeTravel {
		// TIT time travel: use normal opInfo, show original branch
		branchEmoji = ""
		branchColor = a.theme.OutputWarningColor
	}

	// OMP-style: use workingTreeLabel from above

	// LFS indicator
	lfsLabel := ""
	lfsColor := a.theme.DimmedTextColor
	if state.LFS {
		if state.LFSReady {
			lfsLabel = "LFS"
			lfsColor = a.theme.AccentTextColor
		} else {
			lfsLabel = "LFS ⚠️"
			lfsColor = a.theme.OutputWarningColor
		}
	}

	headerState := ui.HeaderState{
		CurrentDirectory: cwd,
		RemoteURL:        remoteURL,
		RemoteColor:      remoteColor,
		OperationEmoji:   opInfo.Emoji,
		OperationLabel:   opInfo.Label,
		OperationColor:   opInfo.Color,
		BranchEmoji:      branchEmoji,
		BranchLabel:      branchName,
		BranchColor:      branchColor,
		WorkingTreeEmoji: wtInfo.Emoji,
		WorkingTreeLabel: workingTreeLabel,
		WorkingTreeDesc:  wtDesc,
		WorkingTreeColor: wtInfo.Color,
		TimelineEmoji:    timelineEmoji,
		TimelineLabel:    timelineLabel,
		TimelineDesc:     timelineDesc,
		TimelineColor:    timelineColor,
		SyncInProgress:   a.activityState.IsAutoUpdateInProgress(),
		SyncFrame:        a.activityState.autoUpdateFrame,
		LFSLabel:         lfsLabel,
		LFSColor:         lfsColor,
	}

	info := ui.RenderHeaderInfo(a.sizing, a.theme, headerState)

	return ui.RenderHeader(a.sizing, a.theme, info)
}

// isInputMode checks if current mode accepts text input

func (a *Application) isInputMode() bool {
	return a.mode == ModeInput ||
		a.mode == ModeCloneURL ||
		(a.mode == ModeSetupWizard && a.environmentState.SetupWizardStep == SetupStepEmail)
}

// buildKeyHandlers builds the complete handler registry for all modes
// Global handlers take priority and are merged into each mode
