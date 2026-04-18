//go:build !windows

package main

import "io"

// platformInput is a no-op on non-Windows platforms.
// Bubble Tea's default input handling works correctly on Mac/Linux.
func platformInput() (io.Reader, func()) {
	return nil, nil
}
