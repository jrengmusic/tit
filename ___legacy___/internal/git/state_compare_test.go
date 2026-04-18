package git

import "testing"

func TestCompareStates(t *testing.T) {
	// base returns a State with all fields set to a known baseline.
	base := func() *State {
		return &State{
			Operation:   Normal,
			WorkingTree: Clean,
			Timeline:    InSync,
			Remote:      NoRemote,
			CurrentBranch: "main",
			CurrentHash: "abc1234",
			CommitsAhead:  0,
			CommitsBehind: 0,
		}
	}

	tests := []struct {
		name string
		old  *State
		new  *State
		want bool
	}{
		{
			name: "same state",
			old:  base(),
			new:  base(),
			want: false,
		},
		{
			name: "operation changes Normal->Merging",
			old:  base(),
			new:  func() *State { s := base(); s.Operation = Merging; return s }(),
			want: true,
		},
		{
			name: "working tree changes Clean->Dirty",
			old:  base(),
			new:  func() *State { s := base(); s.WorkingTree = Dirty; return s }(),
			want: true,
		},
		{
			name: "timeline changes InSync->Ahead",
			old:  base(),
			new:  func() *State { s := base(); s.Timeline = Ahead; return s }(),
			want: true,
		},
		{
			name: "remote changes NoRemote->HasRemote",
			old:  base(),
			new:  func() *State { s := base(); s.Remote = HasRemote; return s }(),
			want: true,
		},
		{
			name: "branch name changes",
			old:  base(),
			new:  func() *State { s := base(); s.CurrentBranch = "feature"; return s }(),
			want: true,
		},
		{
			// Both are Ahead — menu does not change even though commit count differs.
			name: "commits ahead 1->2, same Timeline=Ahead",
			old:  func() *State { s := base(); s.Timeline = Ahead; s.CommitsAhead = 1; return s }(),
			new:  func() *State { s := base(); s.Timeline = Ahead; s.CommitsAhead = 2; return s }(),
			want: false,
		},
		{
			// Both are Behind — menu does not change even though commit count differs.
			name: "commits behind 1->2, same Timeline=Behind",
			old:  func() *State { s := base(); s.Timeline = Behind; s.CommitsBehind = 1; return s }(),
			new:  func() *State { s := base(); s.Timeline = Behind; s.CommitsBehind = 2; return s }(),
			want: false,
		},
		{
			// CurrentHash is not compared by CompareStates — menu is unaffected.
			name: "only CurrentHash changes",
			old:  base(),
			new:  func() *State { s := base(); s.CurrentHash = "zzz9999"; return s }(),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CompareStates(tc.old, tc.new)
			if got != tc.want {
				t.Errorf("CompareStates() = %v, want %v", got, tc.want)
			}
		})
	}
}
