package app

// AsyncState manages async operation lifecycle.
// Thread-safe: All operations happen on UI thread (single-threaded Bubbletea).
//
// Lifecycle:
//  1. Start()  — marks operation active, clears aborted flag
//  2. Abort()  — marks operation aborted (ESC pressed)
//  3. End()    — marks operation complete
//
// Exit control:
//   - SetExitAllowed(true) allows quit during long operations
//   - SetExitAllowed(false) prevents quit during critical operations
type AsyncState struct {
	active      bool // True while async operation running
	aborted     bool // True if user pressed ESC to abort
	exitAllowed bool // True if exit allowed during operation
}

// Start marks an async operation as active.
func (s *AsyncState) Start() {
	s.active = true
	s.aborted = false
}

// End marks an async operation as complete.
func (s *AsyncState) End() {
	s.active = false
}

// Abort marks the current async operation as aborted.
func (s *AsyncState) Abort() {
	s.aborted = true
}

// ClearAborted resets the aborted flag.
func (s *AsyncState) ClearAborted() {
	s.aborted = false
}

// IsActive returns true if an async operation is running.
func (s *AsyncState) IsActive() bool {
	return s.active
}

// IsAborted returns true if the current async operation was aborted.
func (s *AsyncState) IsAborted() bool {
	return s.aborted
}

// CanExit returns true if exit is allowed during operation.
func (s *AsyncState) CanExit() bool {
	return s.exitAllowed
}

// SetExitAllowed sets whether exit is allowed during operation.
func (s *AsyncState) SetExitAllowed(allowed bool) {
	s.exitAllowed = allowed
}
