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
				return "Your files match the remote."
			},
		},
		git.Modified: {
			Label: "Modified",
			Emoji: "📝",
			Color: theme.StatusModified,
			Description: func(ahead, behind int) string {
				return "You have uncommitted changes."
			},
		},
	}

	timelineInfo := map[git.Timeline]StateInfo{
		git.TimelineNoRemote: {
			Label: "No remote",
			Emoji: "🔌",
			Color: theme.FooterTextColor,
			Description: func(ahead, behind int) string {
				return "No remote repository configured."
			},
		},
		git.InSync: {
			Label: "Sync",
			Emoji: "🔗",
			Color: theme.TimelineSynchronized,
			Description: func(ahead, behind int) string {
				return "Local and remote are in sync."
			},
		},
		git.Ahead: {
			Label: "Local ahead",
			Emoji: "🌎",
			Color: theme.TimelineLocalAhead,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf("You have %d unsynced commit(s).", ahead)
			},
		},
		git.Behind: {
			Label: "Local behind",
			Emoji: "🪐",
			Color: theme.TimelineLocalBehind,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf("The remote has %d new commit(s).", behind)
			},
		},
		git.Diverged: {
			Label: "Diverged",
			Emoji: "💥",
			Color: theme.TimelineLocalBehind,
			Description: func(ahead, behind int) string {
				return fmt.Sprintf("Both have new commits. Ahead %d, Behind %d.", ahead, behind)
			},
		},
	}

	return workingTreeInfo, timelineInfo
}
