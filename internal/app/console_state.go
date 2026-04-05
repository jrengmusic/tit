package app

import "tit/internal/ui"

// ConsoleState manages console scroll position and auto-scroll behavior.
// Buffer access is via ui.GetBuffer() — ConsoleState does not own the buffer.
type ConsoleState struct {
	state      ui.ConsoleOutState // Scroll position, etc.
	autoScroll bool               // Auto-scroll to bottom
}

// NewConsoleState creates a new ConsoleState.
func NewConsoleState() ConsoleState {
	return ConsoleState{
		autoScroll: true,
	}
}

// Reset clears buffer, resets scroll, and enables auto-scroll
func (c *ConsoleState) Reset() {
	ui.GetBuffer().Clear()
	c.state.ScrollOffset = 0
	c.autoScroll = true
}

// ScrollUp scrolls the console view up.
func (c *ConsoleState) ScrollUp() {
	if c.state.ScrollOffset > 0 {
		c.state.ScrollOffset--
	}
}

// ScrollDown scrolls the console view down.
func (c *ConsoleState) ScrollDown() {
	c.state.ScrollOffset++
}

// PageUp scrolls up by page.
func (c *ConsoleState) PageUp() {
	if c.state.ScrollOffset > PageScrollLines {
		c.state.ScrollOffset -= PageScrollLines
	} else {
		c.state.ScrollOffset = 0
	}
}

// PageDown scrolls down by page.
func (c *ConsoleState) PageDown() {
	c.state.ScrollOffset += PageScrollLines
}

// ToggleAutoScroll toggles auto-scroll behavior.
func (c *ConsoleState) ToggleAutoScroll() {
	c.autoScroll = !c.autoScroll
}

// IsAutoScroll returns true if auto-scroll is enabled.
func (c *ConsoleState) IsAutoScroll() bool {
	return c.autoScroll
}

// ViewState returns a reference to the console state for UI rendering
func (c *ConsoleState) ViewState() *ui.ConsoleOutState {
	return &c.state
}

