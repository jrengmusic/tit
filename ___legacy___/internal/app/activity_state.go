package app

import "time"

// ActivityState tracks menu navigation activity and auto-update status.
type ActivityState struct {
	lastActivity         time.Time
	activityTimeout      time.Duration
	autoUpdateInProgress bool
	autoUpdateFrame      int
}

// NewActivityState creates a new ActivityState with defaults.
func NewActivityState() ActivityState {
	return ActivityState{
		lastActivity:    time.Now(),
		activityTimeout: 30 * time.Second,
	}
}

// MarkActivity updates the last activity timestamp to now.
func (a *ActivityState) MarkActivity() {
	a.lastActivity = time.Now()
}

// IsInactive returns true if no activity for longer than timeout.
func (a *ActivityState) IsInactive() bool {
	return time.Since(a.lastActivity) > a.activityTimeout
}

// StartAutoUpdate marks auto-update as in progress.
func (a *ActivityState) StartAutoUpdate() {
	a.autoUpdateInProgress = true
	a.autoUpdateFrame = 0
}

// StopAutoUpdate marks auto-update as complete.
func (a *ActivityState) StopAutoUpdate() {
	a.autoUpdateInProgress = false
}

// IsAutoUpdateInProgress returns true if auto-update is running.
func (a *ActivityState) IsAutoUpdateInProgress() bool {
	return a.autoUpdateInProgress
}

// IncrementFrame advances the auto-update animation frame.
func (a *ActivityState) IncrementFrame() {
	a.autoUpdateFrame++
}

