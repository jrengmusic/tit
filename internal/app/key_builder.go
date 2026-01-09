package app

// ModeHandlerBuilder provides a fluent API for building key handler maps.
type ModeHandlerBuilder struct {
	handlers map[string]KeyHandler
}

// NewModeHandlers creates a new ModeHandlerBuilder.
func NewModeHandlers() *ModeHandlerBuilder {
	return &ModeHandlerBuilder{
		handlers: make(map[string]KeyHandler),
	}
}

// WithMenuNav adds standard menu navigation keys (up, down, j, k).
func (b *ModeHandlerBuilder) WithMenuNav(a *Application) *ModeHandlerBuilder {
	b.handlers["up"] = a.handleMenuUp
	b.handlers["k"] = a.handleMenuUp
	b.handlers["down"] = a.handleMenuDown
	b.handlers["j"] = a.handleMenuDown
	return b
}

// WithCursorNav adds standard cursor navigation keys from a pre-built map.
func (b *ModeHandlerBuilder) WithCursorNav(navHandlers map[string]KeyHandler) *ModeHandlerBuilder {
	b.handlers["left"] = navHandlers["left"]
	b.handlers["right"] = navHandlers["right"]
	b.handlers["home"] = navHandlers["home"]
	b.handlers["end"] = navHandlers["end"]
	return b
}

// On adds a key handler for a specific key.
func (b *ModeHandlerBuilder) On(key string, handler KeyHandler) *ModeHandlerBuilder {
	b.handlers[key] = handler
	return b
}

// Build returns the final map of key handlers.
func (b *ModeHandlerBuilder) Build() map[string]KeyHandler {
	return b.handlers
}
