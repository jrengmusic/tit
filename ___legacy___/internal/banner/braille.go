package banner

import "fmt"

// BrailleChar represents a braille character with its color
type BrailleChar struct {
	Char  rune
	Color Color
}

// Braille patterns: each character represents a 2×4 grid of pixels
// Unicode braille range: U+2800 to U+28FF
var braillePatterns = []rune{
	'⠀', '⠁', '⠂', '⠃', '⠄', '⠅', '⠆', '⠇',
	'⠈', '⠉', '⠊', '⠋', '⠌', '⠍', '⠎', '⠏',
	'⠐', '⠑', '⠒', '⠓', '⠔', '⠕', '⠖', '⠗',
	'⠘', '⠙', '⠚', '⠛', '⠜', '⠝', '⠞', '⠟',
	'⠠', '⠡', '⠢', '⠣', '⠤', '⠥', '⠦', '⠧',
	'⠨', '⠩', '⠪', '⠫', '⠬', '⠭', '⠮', '⠯',
	'⠰', '⠱', '⠲', '⠳', '⠴', '⠵', '⠶', '⠷',
	'⠸', '⠹', '⠺', '⠻', '⠼', '⠽', '⠾', '⠿',
	'⡀', '⡁', '⡂', '⡃', '⡄', '⡅', '⡆', '⡇',
	'⡈', '⡉', '⡊', '⡋', '⡌', '⡍', '⡎', '⡏',
	'⡐', '⡑', '⡒', '⡓', '⡔', '⡕', '⡖', '⡗',
	'⡘', '⡙', '⡚', '⡛', '⡜', '⡝', '⡞', '⡟',
	'⡠', '⡡', '⡢', '⡣', '⡤', '⡥', '⡦', '⡧',
	'⡨', '⡩', '⡪', '⡫', '⡬', '⡭', '⡮', '⡯',
	'⡰', '⡱', '⡲', '⡳', '⡴', '⡵', '⡶', '⡷',
	'⡸', '⡹', '⡺', '⡻', '⡼', '⡽', '⡾', '⡿',
	'⢀', '⢁', '⢂', '⢃', '⢄', '⢅', '⢆', '⢇',
	'⢈', '⢉', '⢊', '⢋', '⢌', '⢍', '⢎', '⢏',
	'⢐', '⢑', '⢒', '⢓', '⢔', '⢕', '⢖', '⢗',
	'⢘', '⢙', '⢚', '⢛', '⢜', '⢝', '⢞', '⢟',
	'⢠', '⢡', '⢢', '⢣', '⢤', '⢥', '⢦', '⢧',
	'⢨', '⢩', '⢪', '⢫', '⢬', '⢭', '⢮', '⢯',
	'⢰', '⢱', '⢲', '⢳', '⢴', '⢵', '⢶', '⢷',
	'⢸', '⢹', '⢺', '⢻', '⢼', '⢽', '⢾', '⢿',
	'⣀', '⣁', '⣂', '⣃', '⣄', '⣅', '⣆', '⣇',
	'⣈', '⣉', '⣊', '⣋', '⣌', '⣍', '⣎', '⣏',
	'⣐', '⣑', '⣒', '⣓', '⣔', '⣕', '⣖', '⣗',
	'⣘', '⣙', '⣚', '⣛', '⣜', '⣝', '⣞', '⣟',
	'⣠', '⣡', '⣢', '⣣', '⣤', '⣥', '⣦', '⣧',
	'⣨', '⣩', '⣪', '⣫', '⣬', '⣭', '⣮', '⣯',
	'⣰', '⣱', '⣲', '⣳', '⣴', '⣵', '⣶', '⣷',
	'⣸', '⣹', '⣺', '⣻', '⣼', '⣽', '⣾', '⣿',
}

// CanvasToBrailleArray converts pixel canvas to braille character array
// Returns array of rows, each row is an array of BrailleChar
// Braille: 2×4 dots per character
func CanvasToBrailleArray(canvas [][]PixelColor, width, height int) [][]BrailleChar {
	var output [][]BrailleChar

	// Process 2 columns at a time (braille width), 4 rows at a time (braille height)
	for y := 0; y < height; y += 4 {
		var row []BrailleChar

		for x := 0; x < width; x += 2 {
			// Get colors for 2×4 dot pattern
			// Braille layout (reading top-to-bottom, left-to-right):
			// 1 4     (col 0 top, col 1 top)
			// 2 5     (col 0 mid-top, col 1 mid-top)
			// 3 6     (col 0 mid-bot, col 1 mid-bot)
			// 7 8     (col 0 bot, col 1 bot)

			brailleIndex := 0
			var colors []PixelColor

			// Collect pixel colors in 2×4 grid (column-major order for braille)
			// Braille dots ordered as: 0,2,4,6 (left col) then 1,3,5,7 (right col)
			for dy := 0; dy < 4 && y+dy < height; dy++ {
				for dx := 0; dx < 2 && x+dx < width; dx++ {
					pixel := canvas[y+dy][x+dx]
					colors = append(colors, pixel)

					// If pixel has color (not black/empty), set bit
					// Threshold: any pixel with combined RGB > 50 counts as "filled"
					brightness := pixel.R + pixel.G + pixel.B
					if brightness > 50 {
						var bitIndex int
						if dx == 0 {
							bitIndex = dy
						} else {
							bitIndex = dy + 4
						}
						brailleIndex |= (1 << bitIndex)
					}
				}
			}

			if brailleIndex >= len(braillePatterns) {
				brailleIndex = 0
			}
			char := braillePatterns[brailleIndex]

			// Use dominant color from the 2×4 grid
			dominantColor := Color{0, 0, 0}
			maxBrightness := 0

			for _, color := range colors {
				brightness := color.R + color.G + color.B
				if brightness > maxBrightness {
					dominantColor = Color{color.R, color.G, color.B}
					maxBrightness = brightness
				}
			}

			row = append(row, BrailleChar{char, dominantColor})
		}

		output = append(output, row)
	}

	return output
}

// SvgToBrailleArray converts SVG to braille character array
// Each character has RGB color that can be converted to ANSI color codes
func SvgToBrailleArray(svgString string, width, height int) [][]BrailleChar {
	canvas := RenderSvgToBraille(svgString, width, height)
	return CanvasToBrailleArray(canvas, width, height)
}

// RGBToHex converts RGB to hex color string
func RGBToHex(r, g, b int) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// RGBToANSI256 maps RGB to nearest terminal 256-color palette index
func RGBToANSI256(r, g, b int) int {
	// Map to 6×6×6 cube (values 0-5)
	r6 := (r * 5) / 255
	g6 := (g * 5) / 255
	b6 := (b * 5) / 255

	// Calculate index in 6×6×6 cube (starts at 16)
	return 16 + (r6 * 36) + (g6 * 6) + b6
}
