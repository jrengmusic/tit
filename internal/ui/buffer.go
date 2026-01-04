package ui

import (
	"sync"
	"time"
)

// OutputLineType represents the type/source of an output line
type OutputLineType string

const (
	TypeStdout  OutputLineType = "stdout"  // Regular command output
	TypeStderr  OutputLineType = "stderr"  // Error output
	TypeCommand OutputLineType = "command" // Command being executed
	TypeStatus  OutputLineType = "status"  // Success/status messages
	TypeWarning OutputLineType = "warning" // Warning messages
	TypeDebug   OutputLineType = "debug"   // Debug/info messages
	TypeInfo    OutputLineType = "info"    // TIT-generated info
)

// OutputLine represents a single line in the output buffer
type OutputLine struct {
	Time string         // Timestamp in HH:MM:SS format
	Type OutputLineType // Line type for color coding
	Text string         // Line content
}

// OutputBuffer is a circular buffer for storing console output
// Thread-safe singleton pattern (accessed from multiple goroutines)
type OutputBuffer struct {
	mu       sync.RWMutex
	maxLines int
	lines    []OutputLine
}

// Global singleton instance (1000 line buffer)
var globalBuffer = &OutputBuffer{
	maxLines: 1000,
	lines:    make([]OutputLine, 0, 1000),
}

// GetBuffer returns the global output buffer instance
func GetBuffer() *OutputBuffer {
	return globalBuffer
}

// Append adds a new line to the buffer with automatic timestamp
// Thread-safe for concurrent writes
func (b *OutputBuffer) Append(text string, lineType OutputLineType) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	timestamp := now.Format("15:04:05") // HH:MM:SS

	line := OutputLine{
		Time: timestamp,
		Type: lineType,
		Text: text,
	}

	b.lines = append(b.lines, line)

	// Maintain circular buffer (remove oldest if exceeds max)
	if len(b.lines) > b.maxLines {
		b.lines = b.lines[1:] // Remove first element
	}
}

// GetLines returns a slice of lines from startIdx to startIdx+count
// Thread-safe for concurrent reads
func (b *OutputBuffer) GetLines(startIdx, count int) []OutputLine {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if startIdx < 0 {
		startIdx = 0
	}

	if startIdx >= len(b.lines) {
		return []OutputLine{}
	}

	endIdx := startIdx + count
	if endIdx > len(b.lines) {
		endIdx = len(b.lines)
	}

	// Return a copy to prevent race conditions
	result := make([]OutputLine, endIdx-startIdx)
	copy(result, b.lines[startIdx:endIdx])
	return result
}

// GetAllLines returns all lines in the buffer
// Thread-safe for concurrent reads
func (b *OutputBuffer) GetAllLines() []OutputLine {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]OutputLine, len(b.lines))
	copy(result, b.lines)
	return result
}

// GetLineCount returns the total number of lines in the buffer
// Thread-safe for concurrent reads
func (b *OutputBuffer) GetLineCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.lines)
}

// Clear removes all lines from the buffer
// Thread-safe for concurrent writes
func (b *OutputBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.lines = make([]OutputLine, 0, b.maxLines)
}
