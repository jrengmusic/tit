package app

import "tit/internal/ui"

// PickerState manages all picker mode states (history, file history, branch picker).
// These share a common pattern: list pane + details pane with coordinated scrolling.
type PickerState struct {
	History      *ui.HistoryState
	FileHistory  *ui.FileHistoryState
	BranchPicker *ui.BranchPickerState
}

// NewPickerState creates a new PickerState with nil states.
func NewPickerState() PickerState {
	return PickerState{}
}

// GetHistory returns the history state (may be nil).
func (p *PickerState) GetHistory() *ui.HistoryState {
	return p.History
}

// SetHistory sets the history state.
func (p *PickerState) SetHistory(state *ui.HistoryState) {
	p.History = state
}

// ResetHistory clears the history state.
func (p *PickerState) ResetHistory() {
	p.History = nil
}

// GetFileHistory returns the file history state (may be nil).
func (p *PickerState) GetFileHistory() *ui.FileHistoryState {
	return p.FileHistory
}

// SetFileHistory sets the file history state.
func (p *PickerState) SetFileHistory(state *ui.FileHistoryState) {
	p.FileHistory = state
}

// ResetFileHistory clears the file history state.
func (p *PickerState) ResetFileHistory() {
	p.FileHistory = nil
}

// GetBranchPicker returns the branch picker state (may be nil).
func (p *PickerState) GetBranchPicker() *ui.BranchPickerState {
	return p.BranchPicker
}

// SetBranchPicker sets the branch picker state.
func (p *PickerState) SetBranchPicker(state *ui.BranchPickerState) {
	p.BranchPicker = state
}

// ResetBranchPicker clears the branch picker state.
func (p *PickerState) ResetBranchPicker() {
	p.BranchPicker = nil
}

// ResetAll clears all picker states.
func (p *PickerState) ResetAll() {
	p.History = nil
	p.FileHistory = nil
	p.BranchPicker = nil
}
