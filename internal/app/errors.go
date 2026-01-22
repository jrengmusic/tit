package app

import (
	"fmt"
	"tit/internal/ui"
)

// ErrorLevel categorizes the severity and visibility of errors in TIT.
// This hierarchy determines how errors are handled and displayed to users.
//
// Error Visibility Levels:
// - ErrorInfo: Internal logging only (debug) - invisible to users
// - ErrorWarn: User-visible warnings - logged and shown in UI
// - ErrorFatal: Critical failures - panic with user-visible message
//
// Usage Pattern:
// - Use ErrorInfo for internal state tracking and debugging
// - Use ErrorWarn for recoverable issues that users should know about
// - Use ErrorFatal for unrecoverable errors that require immediate termination
//
// CONTRACT: Never silently ignore errors. Always use appropriate ErrorLevel.

type ErrorLevel int

const (
	// ErrorInfo: Internal logging only (debug)
	// Does not display to user
	ErrorInfo ErrorLevel = iota

	// ErrorWarn: Logged to output buffer and footer hint
	// User-visible but non-fatal
	ErrorWarn

	// ErrorFatal: Panic with message
	// Terminates operation, user sees message before crash (if any)
	ErrorFatal
)

// ErrorConfig standardizes error reporting across the application
// Provides consistent logging, display, and recovery patterns
type ErrorConfig struct {
	// Level determines visibility: Info (debug), Warn (show to user), Fatal (panic)
	Level ErrorLevel

	// Message is the primary error description
	Message string

	// InnerError is the underlying Go error (can be nil)
	InnerError error

	// BufferLine is what to display in the console output buffer
	// Typically: ErrorMessages[key] or formatted error string
	// Only shown if Level >= ErrorWarn
	BufferLine string

	// FooterLine is what to display in the footer hint
	// Short message (1 line), typically user action-oriented
	// Only shown if Level >= ErrorWarn
	FooterLine string
}

// LogError handles error logging with consistent level-based behavior
// UI THREAD - Called from event handlers
func (a *Application) LogError(config ErrorConfig) {
	fullMsg := fmt.Sprintf("%s: %v", config.Message, config.InnerError)

	switch config.Level {
	case ErrorInfo:
		// Internal logging only (debug)
		// Could add to debug log here if needed

	case ErrorWarn:
		// Show to user in buffer and footer
		buffer := ui.GetBuffer()
		if config.BufferLine != "" {
			buffer.Append(config.BufferLine, ui.TypeStderr)
		}
		if config.FooterLine != "" {
			a.footerHint = config.FooterLine
		}

	case ErrorFatal:
		// Panic with full message
		panic(fullMsg)
	}
}

// LogErrorSimple is a convenience wrapper for common warn-level errors
// Does not require custom BufferLine/FooterLine
// UI THREAD - Called from event handlers
func (a *Application) LogErrorSimple(message string, err error, footerMsg string) {
	a.LogError(ErrorConfig{
		Level:      ErrorWarn,
		Message:    message,
		InnerError: err,
		BufferLine: fmt.Sprintf("[ERROR] %s: %v", message, err),
		FooterLine: footerMsg,
	})
}

// LogErrorFatal is a convenience wrapper for fatal errors
// THREAD CONTEXT: Can be called from any context, will panic
func LogErrorFatal(message string, err error) {
	panic(fmt.Sprintf("[FATAL] %s: %v", message, err))
}
