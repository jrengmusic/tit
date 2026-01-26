package app

import (
	"tit/internal/git"
	"tit/internal/ui"
)

// GitLogger implements git.Logger using UI buffer.
type GitLogger struct{}

func (l *GitLogger) Log(message string) {
	ui.GetBuffer().Append(message, ui.TypeInfo)
}

func (l *GitLogger) Warn(message string) {
	ui.GetBuffer().Append(message, ui.TypeStderr)
}

func (l *GitLogger) Error(message string) {
	ui.GetBuffer().Append(message, ui.TypeStderr)
}

// InitGitLogger sets up git package logger.
func InitGitLogger() {
	git.SetLogger(&GitLogger{})
}
