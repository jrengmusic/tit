package ui

import (
	"testing"
)

func TestValidateRemoteURL(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", false},
		{"whitespace only", "   ", false},
		{"SSH with colon", "git@github.com:user/repo.git", true},
		{"SSH without colon", "git@github.com/user/repo", false},
		{"HTTPS", "https://github.com/user/repo.git", true},
		{"HTTP", "http://example.com/repo", true},
		{"HTTPS too short len=8", "https://", false},
		{"HTTP too short len=7", "http://", false},
		{"HTTPS len=9", "https://x", true},
		{"absolute path", "/absolute/path", true},
		{"tilde path", "~/relative/path", true},
		{"ftp scheme", "ftp://invalid", false},
		{"random string", "random-string", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateRemoteURL(tc.input)
			if got != tc.want {
				t.Errorf("ValidateRemoteURL(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeCommitMessage(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"printable ASCII unchanged", "hello world!", "hello world!"},
		{"NUL stripped", "hello\x00world", "helloworld"},
		{"CR stripped", "line1\r\nline2", "line1\nline2"},
		{"triple blank collapsed to single", "a\n\n\nb", "a\n\nb"},
		{"trimmed whitespace", "  hello  ", "hello"},
		{"empty string", "", ""},
		{"all stripped and trimmed", "\t\r\n", ""},
		{"unicode emoji stripped", "hi \U0001F600 there", "hi  there"},
		{"standard ASCII symbols preserved", "!\"#$%&", "!\"#$%&"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeCommitMessage(tc.input)
			if got != tc.want {
				t.Errorf("SanitizeCommitMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestValidatorsURL(t *testing.T) {
	v := Validators["url"]

	cases := []struct {
		name      string
		input     string
		wantOK    bool
		wantEmpty bool // true if we expect msg == ""
	}{
		{"empty string", "", false, false},
		{"valid SSH", "git@github.com:user/repo.git", true, true},
		{"invalid random string", "invalid", false, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, msg := v(tc.input)
			if ok != tc.wantOK {
				t.Errorf("Validators[\"url\"](%q) ok = %v, want %v", tc.input, ok, tc.wantOK)
			}
			if tc.wantEmpty && msg != "" {
				t.Errorf("Validators[\"url\"](%q) msg = %q, want empty", tc.input, msg)
			}
			if !tc.wantEmpty && msg == "" {
				t.Errorf("Validators[\"url\"](%q) msg is empty, want non-empty error", tc.input)
			}
		})
	}
}

func TestValidatorsBranchName(t *testing.T) {
	v := Validators["branch_name"]

	cases := []struct {
		name      string
		input     string
		wantOK    bool
		wantEmpty bool
	}{
		{"empty string", "", false, false},
		{"contains space", "my branch", false, false},
		{"valid branch with slash", "feature/test", true, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, msg := v(tc.input)
			if ok != tc.wantOK {
				t.Errorf("Validators[\"branch_name\"](%q) ok = %v, want %v", tc.input, ok, tc.wantOK)
			}
			if tc.wantEmpty && msg != "" {
				t.Errorf("Validators[\"branch_name\"](%q) msg = %q, want empty", tc.input, msg)
			}
			if !tc.wantEmpty && msg == "" {
				t.Errorf("Validators[\"branch_name\"](%q) msg is empty, want non-empty error", tc.input)
			}
		})
	}
}
