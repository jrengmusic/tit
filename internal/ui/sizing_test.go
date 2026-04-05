package ui

import (
	"testing"
)

func TestCalculateDynamicSizing(t *testing.T) {
	cases := []struct {
		name            string
		termWidth       int
		termHeight      int
		wantIsTooSmall  bool
		checkPositive   bool // verify content dimensions are positive
	}{
		{"minimum size 70x20", 70, 20, false, true},
		{"below minimum 68x18", 68, 18, true, false},
		{"standard 80x24", 80, 24, false, true},
		{"very large 200x60", 200, 60, false, true},
		{"very small 1x1", 1, 1, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := CalculateDynamicSizing(tc.termWidth, tc.termHeight)

			if s.IsTooSmall != tc.wantIsTooSmall {
				t.Errorf("CalculateDynamicSizing(%d, %d).IsTooSmall = %v, want %v",
					tc.termWidth, tc.termHeight, s.IsTooSmall, tc.wantIsTooSmall)
			}

			// Verify terminal dimensions are stored verbatim
			if s.TerminalWidth != tc.termWidth {
				t.Errorf("TerminalWidth = %d, want %d", s.TerminalWidth, tc.termWidth)
			}
			if s.TerminalHeight != tc.termHeight {
				t.Errorf("TerminalHeight = %d, want %d", s.TerminalHeight, tc.termHeight)
			}

			if tc.checkPositive {
				if s.ContentHeight <= 0 {
					t.Errorf("ContentHeight = %d, want > 0", s.ContentHeight)
				}
				if s.ContentInnerWidth <= 0 {
					t.Errorf("ContentInnerWidth = %d, want > 0", s.ContentInnerWidth)
				}
				if s.HeaderInnerWidth <= 0 {
					t.Errorf("HeaderInnerWidth = %d, want > 0", s.HeaderInnerWidth)
				}
				if s.FooterInnerWidth <= 0 {
					t.Errorf("FooterInnerWidth = %d, want > 0", s.FooterInnerWidth)
				}
			}
		})
	}
}

func TestCalculateDynamicSizingArithmetic(t *testing.T) {
	// Verify exact arithmetic matches sizing.go constants
	s := CalculateDynamicSizing(80, 24)

	wantContentHeight := 24 - HeaderHeight - FooterHeight // 16
	if wantContentHeight < MinContentHeight {
		wantContentHeight = MinContentHeight
	}
	if s.ContentHeight != wantContentHeight {
		t.Errorf("ContentHeight = %d, want %d", s.ContentHeight, wantContentHeight)
	}

	wantInnerWidth := 80 - (HorizontalMargin * 2) // 76
	if s.ContentInnerWidth != wantInnerWidth {
		t.Errorf("ContentInnerWidth = %d, want %d", s.ContentInnerWidth, wantInnerWidth)
	}
	if s.HeaderInnerWidth != wantInnerWidth {
		t.Errorf("HeaderInnerWidth = %d, want %d", s.HeaderInnerWidth, wantInnerWidth)
	}
	if s.FooterInnerWidth != wantInnerWidth {
		t.Errorf("FooterInnerWidth = %d, want %d", s.FooterInnerWidth, wantInnerWidth)
	}

	wantMenuColumnWidth := wantInnerWidth/2 - 2
	if s.MenuColumnWidth != wantMenuColumnWidth {
		t.Errorf("MenuColumnWidth = %d, want %d", s.MenuColumnWidth, wantMenuColumnWidth)
	}
}

func TestCalculateDynamicSizingContentHeightFloor(t *testing.T) {
	// When termHeight - HeaderHeight - FooterHeight < MinContentHeight, floor applies
	// HeaderHeight=7, FooterHeight=1 => need termHeight <= 7+1+MinContentHeight-1 = 11
	s := CalculateDynamicSizing(80, 10)
	if s.ContentHeight < MinContentHeight {
		t.Errorf("ContentHeight = %d, below MinContentHeight %d", s.ContentHeight, MinContentHeight)
	}
}
