package ui

import (
	"strings"
	"testing"
)

func TestPadTextToHeight(t *testing.T) {
	cases := []struct {
		name      string
		text      string
		height    int
		wantLines int
	}{
		{"pad 2 lines to 4", "line1\nline2", 4, 4},
		{"truncate 4 lines to 2", "a\nb\nc\nd", 2, 2},
		{"empty string to 3", "", 3, 3},
		{"single line height 1", "single", 1, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PadTextToHeight(tc.text, tc.height)
			lines := strings.Split(got, "\n")
			if len(lines) != tc.wantLines {
				t.Errorf("PadTextToHeight(%q, %d): got %d lines, want %d\nresult: %q",
					tc.text, tc.height, len(lines), tc.wantLines, got)
			}
		})
	}
}

func TestPadTextToHeightContent(t *testing.T) {
	// Verify content lines are preserved when padding
	got := PadTextToHeight("line1\nline2", 4)
	lines := strings.Split(got, "\n")
	if lines[0] != "line1" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "line1")
	}
	if lines[1] != "line2" {
		t.Errorf("lines[1] = %q, want %q", lines[1], "line2")
	}
	if lines[2] != "" {
		t.Errorf("lines[2] = %q, want empty", lines[2])
	}
	if lines[3] != "" {
		t.Errorf("lines[3] = %q, want empty", lines[3])
	}
}

func TestPadTextToHeightTruncation(t *testing.T) {
	// Verify truncation keeps first N lines
	got := PadTextToHeight("a\nb\nc\nd", 2)
	lines := strings.Split(got, "\n")
	if lines[0] != "a" {
		t.Errorf("lines[0] = %q, want %q", lines[0], "a")
	}
	if lines[1] != "b" {
		t.Errorf("lines[1] = %q, want %q", lines[1], "b")
	}
}

