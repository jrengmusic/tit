package app

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ActionHandler is a function type for action dispatchers
type ActionHandler func(*Application) tea.Cmd

// isCwdEmpty checks if current working directory is empty
// Ignores macOS metadata files (.DS_Store)
// Used for smart dispatch in init/clone workflows
func isCwdEmpty() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}

	entries, err := os.ReadDir(cwd)
	if err != nil {
		return false
	}

	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if name != ".DS_Store" && name != ".AppleDouble" {
			count++
		}
	}

	return count == 0
}

// parseCommitDate parses a Git date string into a time.Time
func parseCommitDate(dateStr string) (time.Time, error) {
	return time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", dateStr)
}
