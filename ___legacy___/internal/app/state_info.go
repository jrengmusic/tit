package app

import (
	"fmt"
	"github.com/jrengmusic/tit/internal/git"
	"github.com/jrengmusic/tit/internal/ui"
)

// StateInfo holds display information for a git state
type StateInfo struct {
	Label       string
	Emoji       string
	Color       string
	Description func(ahead, behind int) string
}

// BuildStateInfo creates maps of WorkingTree, Timeline, and Operation states to display info
func BuildStateInfo(theme ui.Theme) (map[git.WorkingTree]StateInfo, map[git.Timeline]StateInfo, map[git.Operation]StateInfo) {
	workingTreeInfo := map[git.WorkingTree]StateInfo{
		git.Clean: {
			Label: "Clean",
			Emoji: "✅",
			Color: theme.StatusClean,
			Description: func(ahead, behind int) string {
				return StateDescriptions["working_tree_clean"]
			},
		},
		git.Dirty: {
			Label: "Dirty",
			Emoji: "📝",
			Color: theme.StatusDirty,
			Description: func(ahead, behind int) string {
				return StateDescriptions["working_tree_dirty"]
			},
		},
	}

	timelineInfo := map[git.Timeline]StateInfo{
		git.InSync: {
			Label: "In sync",
			Emoji: "🔗",
			Color: theme.TimelineSynchronized,
			Description: func(ahead, behind int) string {
				return StateDescriptions["timeline_in_sync"]
			},
		},
		git.Ahead: {
			Label: "Local ahead",
			Emoji: "🌍",
			Color: theme.TimelineLocalAhead,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf(StateDescriptions["timeline_ahead"], ahead)
			},
		},
		git.Behind: {
			Label: "Local behind",
			Emoji: "🪐",
			Color: theme.TimelineLocalBehind,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf(StateDescriptions["timeline_behind"], behind)
			},
		},
		git.Diverged: {
			Label: "Diverged",
			Emoji: "💥",
			Color: theme.TimelineLocalBehind,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf(StateDescriptions["timeline_diverged"], ahead, behind)
			},
		},
	}

	operationInfo := map[git.Operation]StateInfo{
		git.Normal: {
			Label: "READY",
			Emoji: "🟢",
			Color: theme.OperationReady,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_normal"]
			},
		},
		git.NotRepo: {
			Label: "NOT REPO",
			Emoji: "🔴",
			Color: theme.OperationNotRepo,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_not_repo"]
			},
		},
		git.Conflicted: {
			Label: "CONFLICTED",
			Emoji: "⚡",
			Color: theme.OperationConflicted,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_conflicted"]
			},
		},
		git.Merging: {
			Label: "MERGING",
			Emoji: "🔀",
			Color: theme.OperationMerging,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_merging"]
			},
		},
		git.Rebasing: {
			Label: "REBASING",
			Emoji: "🔄",
			Color: theme.OperationRebasing,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_rebasing"]
			},
		},
		git.DirtyOperation: {
			Label: "DIRTY OP",
			Emoji: "⚡",
			Color: theme.OperationDirtyOp,
			Description: func(ahead, behind int) string {
				return StateDescriptions["operation_dirty_op"]
			},
		},
		git.TimeTraveling: {
			Label: "TIME TRAVEL",
			Emoji: "🌀",
			Color: theme.OperationTimeTravel,
			Description: func(ahead, behind int) string {
				// Note: Will be formatted with commit hash and date in renderStateHeader
				return StateDescriptions["operation_time_travel"]
			},
		},
	}

	return workingTreeInfo, timelineInfo, operationInfo
}
