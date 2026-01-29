package app

import "tit/internal/ui"

// ConsoleState manages console output display and scrolling.
// Thread-safe: Uses thread-safe OutputBuffer.
type ConsoleState struct {
	state      ui.ConsoleOutState // Scroll position, etc.
	buffer     *ui.OutputBuffer   // Thread-safe output buffer
	autoScroll bool               // Auto-scroll to bottom
}

// NewConsoleState creates a new ConsoleState.
func NewConsoleState() ConsoleState {
	return ConsoleState{
		buffer:     ui.GetBuffer(),
		autoScroll: true,
	}
}

// GetBuffer returns the output buffer.
func (c *ConsoleState) GetBuffer() *ui.OutputBuffer {
	return c.buffer
}

// Clear clears the console buffer.
func (c *ConsoleState) Clear() {
	c.buffer.Clear()
}

// Reset clears the buffer and resets scroll state.
func (c *ConsoleState) Reset() {
	c.buffer.Clear()
	c.state.ScrollOffset = 0
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
	if c.state.ScrollOffset > 10 {
		c.state.ScrollOffset -= 10
	} else {
		c.state.ScrollOffset = 0
	}
}

// PageDown scrolls down by page.
func (c *ConsoleState) PageDown() {
	c.state.ScrollOffset += 10
}

// ToggleAutoScroll toggles auto-scroll behavior.
func (c *ConsoleState) ToggleAutoScroll() {
	c.autoScroll = !c.autoScroll
}

// SetAutoScroll sets the auto-scroll state directly.
func (c *ConsoleState) SetAutoScroll(enabled bool) {
	c.autoScroll = enabled
}

// IsAutoScroll returns true if auto-scroll is enabled.
func (c *ConsoleState) IsAutoScroll() bool {
	return c.autoScroll
}

// GetState returns the console state.
func (c *ConsoleState) GetState() ui.ConsoleOutState {
	return c.state
}

// GetStateRef returns a reference to the console state (for passing to UI functions).
func (c *ConsoleState) GetStateRef() *ui.ConsoleOutState {
	return &c.state
}

// SetScrollOffset sets the scroll offset directly.
func (c *ConsoleState) SetScrollOffset(offset int) {
	c.state.ScrollOffset = offset
}
