package app

import "context"

// OperationState manages async operations and workflow state
type OperationState struct {
	asyncState           AsyncState
	workflowState        WorkflowState
	consoleState         *ConsoleState
	inputState           InputState
	cancelContext        context.CancelFunc
	conflictResolveState *ConflictResolveState
	dirtyOperationState  *DirtyOperationState
}

// Async Operation Helpers

// StartAsyncOp marks an async operation as active
func (o *OperationState) StartAsyncOp() {
	o.asyncState.Start()
}

// EndAsyncOp marks async operation as complete
func (o *OperationState) EndAsyncOp() {
	o.asyncState.End()
}

// AbortAsyncOp marks async operation as aborted
func (o *OperationState) AbortAsyncOp() {
	o.asyncState.Abort()
	o.asyncState.End()
}

// IsAsyncActive returns true if async operation is running
func (o *OperationState) IsAsyncActive() bool {
	return o.asyncState.IsActive()
}

// IsAsyncAborted returns true if async operation was aborted
func (o *OperationState) IsAsyncAborted() bool {
	return o.asyncState.IsAborted()
}

// ClearAsyncAborted resets aborted flag
func (o *OperationState) ClearAsyncAborted() {
	o.asyncState.ClearAborted()
}

// PermitExit sets whether app can exit during async op
func (o *OperationState) PermitExit(allowed bool) {
	o.asyncState.exitAllowed = allowed
}

// CanExit returns true if app can exit
func (o *OperationState) CanExit() bool {
	if !o.asyncState.IsActive() {
		return true
	}
	return o.asyncState.CanExit()
}

// CancelContext Helpers

// ClearCancelContext removes stored cancel function
func (o *OperationState) ClearCancelContext() {
	o.cancelContext = nil
}

// Workflow State Helpers

// WorkflowState returns the workflow state reference
func (o *OperationState) WorkflowState() *WorkflowState {
	return &o.workflowState
}

// ResetWorkflow clears workflow state
func (o *OperationState) ResetWorkflow() {
	o.workflowState = WorkflowState{}
}

// Console State Helpers

// EnsureConsoleState returns the console state, initializing if needed
func (o *OperationState) EnsureConsoleState() *ConsoleState {
	if o.consoleState == nil {
		newState := NewConsoleState()
		o.consoleState = &newState
	}
	return o.consoleState
}

// EnterConsoleMode prepares console and workflow state for async operation display.
// Callers must also set NavigationState mode to ModeConsole and UIState footerHint.
func (o *OperationState) EnterConsoleMode() {
	o.consoleState.Reset()
	o.workflowState.PreviousMode = ModeMenu
	o.workflowState.PreviousMenuIndex = 0
}

// Input State Helpers

// InputState returns the input state reference
func (o *OperationState) InputState() *InputState {
	return &o.inputState
}

// Conflict Resolution Helpers

// ClearConflictResolveState clears conflict resolution state
func (o *OperationState) ClearConflictResolveState() {
	o.conflictResolveState = nil
}

// Dirty Operation Helpers

// ClearDirtyOperationState clears dirty operation state
func (o *OperationState) ClearDirtyOperationState() {
	o.dirtyOperationState = nil
}
