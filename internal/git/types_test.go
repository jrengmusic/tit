package git

import "testing"

func TestShortenHash(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"abc1234", "abc1234"},                            // exactly 7 chars: unchanged
		{"abc12345", "abc1234"},                           // 8 chars: truncated to 7
		{"a94a8fe5ccb19ba61c4c0873d391e987982fbbd3", "a94a8fe"}, // 40-char SHA: first 7
		{"abc", "abc"},                                    // 3 chars: unchanged
	}

	for _, tc := range tests {
		got := ShortenHash(tc.input)
		if got != tc.want {
			t.Errorf("ShortenHash(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
