package app

import "tit/internal/ui"

// UIState manages display and layout state
type UIState struct {
	width      int
	height     int
	sizing     ui.DynamicSizing
	theme      ui.Theme
	footerHint string
}

// SetSize updates dimensions and recalculates sizing
func (u *UIState) SetSize(width, height int) {
	u.width = width
	u.height = height
	u.sizing = ui.CalculateDynamicSizing(width, height)
}

// ContentWidth returns the available content width
func (u *UIState) ContentWidth() int {
	return u.sizing.ContentInnerWidth
}

// ContentHeight returns the available content height
func (u *UIState) ContentHeight() int {
	return u.sizing.ContentHeight
}

// SetFooterHint updates the footer message
func (u *UIState) SetFooterHint(hint string) {
	u.footerHint = hint
}

// GetFooterHint returns current footer message
func (u *UIState) GetFooterHint() string {
	return u.footerHint
}
