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

// ResetHistory clears the history state.
func (p *PickerState) ResetHistory() {
	p.History = nil
}

// ResetFileHistory clears the file history state.
func (p *PickerState) ResetFileHistory() {
	p.FileHistory = nil
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
