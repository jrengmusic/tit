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

// Resize updates dimensions and recalculates sizing
func (u *UIState) Resize(width, height int) {
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

