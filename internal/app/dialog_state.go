package app

import "tit/internal/ui"

// DialogState manages confirmation dialog display and context.
type DialogState struct {
	dialog  *ui.ConfirmationDialog
	context map[string]string
}

// NewDialogState creates a new DialogState.
func NewDialogState() DialogState {
	return DialogState{
		context: make(map[string]string),
	}
}

// Show sets the dialog to display with context.
func (d *DialogState) Show(dialog *ui.ConfirmationDialog, ctx map[string]string) {
	d.dialog = dialog
	if ctx != nil {
		d.context = ctx
	} else {
		d.context = make(map[string]string)
	}
}

// Hide clears the current dialog.
func (d *DialogState) Hide() {
	d.dialog = nil
	d.context = make(map[string]string)
}

// GetDialog returns the current dialog (may be nil).
func (d *DialogState) GetDialog() *ui.ConfirmationDialog {
	return d.dialog
}

// IsVisible returns true if a dialog is currently shown.
func (d *DialogState) IsVisible() bool {
	return d.dialog != nil
}

// GetContext returns the dialog context map.
func (d *DialogState) GetContext() map[string]string {
	return d.context
}

// SetContextValue sets a single context value.
func (d *DialogState) SetContextValue(key, value string) {
	d.context[key] = value
}

// GetContextValue returns a single context value (empty if not found).
func (d *DialogState) GetContextValue(key string) string {
	return d.context[key]
}
