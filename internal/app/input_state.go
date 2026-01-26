package app

import "tit/internal/ui"

// InputState manages user input in text entry modes.
// Thread-safe: All operations happen on UI thread (single-threaded).
type InputState struct {
	Prompt          string
	Value           string
	CursorPosition  int
	Height          int
	Action          string
	ValidationMsg   string
	ClearConfirming bool
}

// Reset clears Value, CursorPosition, ValidationMsg (preserves Height, Action).
func (s *InputState) Reset() {
	s.Value = ""
	s.CursorPosition = 0
	s.ValidationMsg = ""
}

// SetValue updates Value and CursorPosition to end of string.
func (s *InputState) SetValue(value string) {
	s.Value = value
	s.CursorPosition = len(value)
}

// GetValue returns current input value.
func (s *InputState) GetValue() string {
	return s.Value
}

// SetCursorPos sets cursor position with bounds checking.
func (s *InputState) SetCursorPos(pos int) {
	if pos < 0 {
		pos = 0
	} else if pos > len(s.Value) {
		pos = len(s.Value)
	}
	s.CursorPosition = pos
}

// GetCursorPos returns current cursor position.
func (s *InputState) GetCursorPos() int {
	return s.CursorPosition
}

// MoveCursorBy moves cursor by delta positions.
func (s *InputState) MoveCursorBy(delta int) {
	s.SetCursorPos(s.CursorPosition + delta)
}

// InsertAtCursor inserts text at cursor position.
func (s *InputState) InsertAtCursor(text string) {
	if text == "" {
		return
	}
	before := s.Value[:s.CursorPosition]
	after := s.Value[s.CursorPosition:]
	s.Value = before + text + after
	s.CursorPosition += len(text)
}

// DeleteBeforeCursor deletes character before cursor.
func (s *InputState) DeleteBeforeCursor() {
	if s.CursorPosition == 0 || s.Value == "" {
		return
	}
	before := s.Value[:s.CursorPosition-1]
	after := s.Value[s.CursorPosition:]
	s.Value = before + after
	s.CursorPosition--
}

// DeleteAfterCursor deletes character after cursor.
func (s *InputState) DeleteAfterCursor() {
	if s.CursorPosition >= len(s.Value) || s.Value == "" {
		return
	}
	before := s.Value[:s.CursorPosition]
	after := s.Value[s.CursorPosition+1:]
	s.Value = before + after
}

// SetPrompt sets the input prompt and action.
func (s *InputState) SetPrompt(prompt, action string, height int) {
	s.Prompt = prompt
	s.Action = action
	s.Height = height
}

// GetPrompt returns current input prompt.
func (s *InputState) GetPrompt() string {
	return s.Prompt
}

// GetAction returns current input action for routing.
func (s *InputState) GetAction() string {
	return s.Action
}

// SetValidationMessage sets validation error message.
func (s *InputState) SetValidationMessage(msg string) {
	s.ValidationMsg = msg
}

// ClearValidationMessage clears validation message.
func (s *InputState) ClearValidationMessage() {
	s.ValidationMsg = ""
}

// GetValidationMessage returns current validation message.
func (s *InputState) GetValidationMessage() string {
	return s.ValidationMsg
}

// HasValidationError returns true if validation message exists.
func (s *InputState) HasValidationError() bool {
	return s.ValidationMsg != ""
}

// SetClearConfirming toggles clear confirmation state.
func (s *InputState) SetClearConfirming(active bool) {
	s.ClearConfirming = active
}

// IsClearConfirming returns clear confirmation state.
func (s *InputState) IsClearConfirming() bool {
	return s.ClearConfirming
}

// ToggleClearConfirming toggles and returns new state.
func (s *InputState) ToggleClearConfirming() bool {
	s.ClearConfirming = !s.ClearConfirming
	return s.ClearConfirming
}

// TextInputState returns a ui state for text input rendering.
func (s *InputState) TextInputState() ui.TextInputState {
	return ui.TextInputState{
		Value:                 s.Value,
		CursorPos:             s.CursorPosition,
		ShowClearConfirmation: s.ClearConfirming,
		Height:                s.Height,
	}
}
