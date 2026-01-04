package app

import (
	"time"
)

// TickMsg is a custom message for quit confirmation timeout
type TickMsg time.Time

// ClearTickMsg is a custom message for input clear confirmation timeout
type ClearTickMsg time.Time

// GitOperationMsg represents the result of a git operation
type GitOperationMsg struct {
	Step    string // "init", "clone", "push", "pull", etc.
	Success bool
	Output  string
	Error   string
	Path    string // Working directory to change to after operation
}

// GitOperationCompleteMsg signals that a git operation completed
type GitOperationCompleteMsg struct {
	Success bool
	Output  string
	Error   string
}

// InputSubmittedMsg signals that user submitted input
type InputSubmittedMsg struct {
	Action string
	Value  string
}

// FooterMessageType enum for different footer message states
type FooterMessageType int

const (
	MessageNone FooterMessageType = iota
	MessageCtrlCConfirm
	MessageEscClearConfirm
	MessageInit
	MessageClone
	MessageCommit
	MessagePush
	MessagePull
	MessageResolveConflicts
)

// GetFooterMessageText returns display text for a message type
func GetFooterMessageText(msgType FooterMessageType) string {
	messages := map[FooterMessageType]string{
		MessageNone:              "",
		MessageCtrlCConfirm:      "Press Ctrl+C again to quit (3s timeout)",
		MessageEscClearConfirm:   "Press ESC again to clear input (3s timeout)",
		MessageInit:              "Initializing repository...",
		MessageClone:             "Cloning repository...",
		MessageCommit:            "Creating commit...",
		MessagePush:              "Pushing to remote...",
		MessagePull:              "Pulling from remote...",
		MessageResolveConflicts:  "Resolving conflicts...",
	}

	if msg, exists := messages[msgType]; exists {
		return msg
	}
	return ""
}
