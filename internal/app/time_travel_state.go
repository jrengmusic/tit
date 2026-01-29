package app

import "tit/internal/git"

// TimeTravelState manages time travel operation state.
type TimeTravelState struct {
	info             *git.TimeTravelInfo
	restoreInitiated bool
}

// NewTimeTravelState creates a new TimeTravelState.
func NewTimeTravelState() TimeTravelState {
	return TimeTravelState{}
}

// IsActive returns true if currently in time travel mode.
func (t *TimeTravelState) IsActive() bool {
	return t.info != nil
}

// GetInfo returns the time travel info (may be nil).
func (t *TimeTravelState) GetInfo() *git.TimeTravelInfo {
	return t.info
}

// SetInfo sets the time travel info.
func (t *TimeTravelState) SetInfo(info *git.TimeTravelInfo) {
	t.info = info
}

// Clear removes time travel state.
func (t *TimeTravelState) Clear() {
	t.info = nil
	t.restoreInitiated = false
}

// IsRestoreInitiated returns true if restore operation has started.
func (t *TimeTravelState) IsRestoreInitiated() bool {
	return t.restoreInitiated
}

// MarkRestoreInitiated marks the restore as initiated.
func (t *TimeTravelState) MarkRestoreInitiated() {
	t.restoreInitiated = true
}

// ClearRestore resets the restore initiated flag.
func (t *TimeTravelState) ClearRestore() {
	t.restoreInitiated = false
}
