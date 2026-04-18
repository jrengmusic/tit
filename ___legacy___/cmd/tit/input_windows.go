//go:build windows

package main

import (
	"io"
	"os"

	"golang.org/x/sys/windows"
)

// stdinReader wraps os.Stdin as a plain io.Reader, stripping the *os.File
// concrete type. This bypasses Bubble Tea's Win32 console reader
// (readConInputs), forcing the ANSI reader path (readAnsiInputs) which
// supports bracketed paste detection.
type stdinReader struct {
	io.Reader
}

// platformInput configures the console for raw VT input and returns a
// wrapped stdin reader. On non-Windows platforms this is a no-op.
func platformInput() (io.Reader, func()) {
	handle := windows.Handle(os.Stdin.Fd())

	var originalMode uint32
	if err := windows.GetConsoleMode(handle, &originalMode); err != nil {
		return nil, nil
	}

	// Raw mode: strip echo, line buffering, processed input (Ctrl+C handled by Bubble Tea)
	// Enable VT input: terminal sends escape sequences parseable by ANSI reader
	const disableFlags = windows.ENABLE_ECHO_INPUT |
		windows.ENABLE_LINE_INPUT |
		windows.ENABLE_PROCESSED_INPUT

	rawMode := (originalMode &^ disableFlags) | windows.ENABLE_VIRTUAL_TERMINAL_INPUT

	if err := windows.SetConsoleMode(handle, rawMode); err != nil {
		return nil, nil
	}

	restore := func() {
		windows.SetConsoleMode(handle, originalMode)
	}

	return stdinReader{os.Stdin}, restore
}
