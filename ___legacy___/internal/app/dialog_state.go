package app

import "github.com/jrengmusic/tit/internal/ui"

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
