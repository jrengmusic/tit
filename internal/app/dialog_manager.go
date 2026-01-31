package app

import "tit/internal/ui"

// DialogManager manages all dialog and picker state
type DialogManager struct {
	dialogState DialogState
	pickerState PickerState
}

// Dialog State Helpers

// GetDialogState returns the dialog state
func (d *DialogManager) GetDialogState() *DialogState {
	return &d.dialogState
}

// ShowDialog activates a dialog
func (d *DialogManager) ShowDialog(dialog *ui.ConfirmationDialog, context map[string]string) {
	d.dialogState.Show(dialog, context)
}

// HideDialog deactivates the dialog
func (d *DialogManager) HideDialog() {
	d.dialogState.Hide()
}

// IsDialogActive returns true if a dialog is showing
func (d *DialogManager) IsDialogActive() bool {
	return d.dialogState.IsVisible()
}

// GetDialogContext returns the dialog context data
func (d *DialogManager) GetDialogContext() map[string]string {
	return d.dialogState.GetContext()
}

// ShowConfirmation displays a confirmation dialog for the given type
func (d *DialogManager) ShowConfirmation(confirmType string, context map[string]string, width int, theme *ui.Theme) {
	msg := ConfirmationMessages[confirmType]
	config := ui.ConfirmationConfig{
		Title:       msg.Title,
		Explanation: msg.Explanation,
		YesLabel:    msg.YesLabel,
		NoLabel:     msg.NoLabel,
		ActionID:    confirmType,
	}
	dialog := ui.NewConfirmationDialog(config, width, theme)
	d.dialogState.Show(dialog, context)
}

// Picker State Helpers

// GetPickerState returns the picker state
func (d *DialogManager) GetPickerState() *PickerState {
	return &d.pickerState
}
