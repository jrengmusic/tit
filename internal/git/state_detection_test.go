package git

import "testing"

func TestDetermineTimeline(t *testing.T) {
	tests := []struct {
		ahead  int
		behind int
		want   Timeline
	}{
		{0, 0, InSync},
		{1, 0, Ahead},
		{0, 1, Behind},
		{1, 1, Diverged},
		{5, 3, Diverged},
		{100, 0, Ahead},
		{0, 100, Behind},
	}

	for _, tc := range tests {
		got := determineTimeline(tc.ahead, tc.behind)
		if got != tc.want {
			t.Errorf("determineTimeline(%d, %d) = %q, want %q", tc.ahead, tc.behind, got, tc.want)
		}
	}
}
