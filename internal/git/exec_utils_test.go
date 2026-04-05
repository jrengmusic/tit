package git

import "testing"

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"/local/path/to/repo.git", "repo"},
		{"~/my-repo.git", "my-repo"},
		// Empty string: filepath.Base("") returns "." — documented actual behavior.
		{"", "."},
	}

	for _, tc := range tests {
		got := ExtractRepoName(tc.input)
		if got != tc.want {
			t.Errorf("ExtractRepoName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
