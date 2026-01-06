package app

import (
	"fmt"
	"tit/internal/git"
	"tit/internal/ui"
)

// StateInfo holds display information for a git state
type StateInfo struct {
	Label       string
	Emoji       string
	Color       string
	Description func(ahead, behind int) string
}

// BuildStateInfo creates maps of WorkingTree and Timeline states to display info
func BuildStateInfo(theme ui.Theme) (map[git.WorkingTree]StateInfo, map[git.Timeline]StateInfo) {
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
		git.TimelineNoRemote: {
			Label: "No remote",
			Emoji: "🔌",
			Color: theme.FooterTextColor,
			Description: func(ahead, behind int) string {
				return StateDescriptions["timeline_no_remote"]
			},
		},
		git.InSync: {
			Label: "Sync",
			Emoji: "🔗",
			Color: theme.TimelineSynchronized,
			Description: func(ahead, behind int) string {
				return StateDescriptions["timeline_in_sync"]
			},
		},
		git.Ahead: {
			Label: "Local ahead",
			Emoji: "🌎",
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

	return workingTreeInfo, timelineInfo
}
