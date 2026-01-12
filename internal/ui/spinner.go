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

// spinnerFrameSet is a lookup set for fast spinner detection
var spinnerFrameSet = map[string]bool{
	"⠋": true,
	"⠙": true,
	"⠹": true,
	"⠸": true,
	"⠼": true,
	"⠴": true,
	"⠦": true,
	"⠧": true,
	"⠇": true,
	"⠏": true,
}

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
