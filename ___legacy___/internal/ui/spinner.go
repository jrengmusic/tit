package ui

// Braille dot spinner animation frames
// Uses Unicode braille patterns to create a rotating dot effect
var SpinnerFrames = []string{
	"⠋",
	"⠙",
	"⠹",
	"⠸",
	"⠼",
	"⠴",
	"⠦",
	"⠧",
	"⠇",
	"⠏",
}

// spinnerFrameSet is derived from SpinnerFrames for fast O(1) lookup
var spinnerFrameSet = func() map[string]bool {
	set := make(map[string]bool, len(SpinnerFrames))
	for _, frame := range SpinnerFrames {
		set[frame] = true
	}
	return set
}()

// SpinnerFrameCount returns the total number of animation frames
func SpinnerFrameCount() int {
	return len(SpinnerFrames)
}

// GetSpinnerFrame returns the spinner character for the given frame index
// Frame index wraps around automatically
func GetSpinnerFrame(frame int) string {
	return SpinnerFrames[frame%len(SpinnerFrames)]
}

// IsSpinnerFrame returns true if the given string is a spinner frame character
func IsSpinnerFrame(s string) bool {
	return spinnerFrameSet[s]
}
