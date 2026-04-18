// braille_ref — dumps braille grid to stdout for byte-identity fixture validation.
//
// Output format (one line per braille row):
//   <codepoint_hex>,<R>,<G>,<B>|<codepoint_hex>,<R>,<G>,<B>|...
// One line per grid row, rows separated by newline.
package main

import (
	"fmt"
	"os"

	"github.com/jrengmusic/tit/internal/banner"
)

func main() {
	svgPath := "internal/ui/assets/tit-logo.svg"
	if len(os.Args) > 1 {
		svgPath = os.Args[1]
	}

	// Terminal cell dimensions — must match C++ test driver
	const termCols = 40
	const termRows = 10

	data, err := os.ReadFile(svgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot read %s: %v\n", svgPath, err)
		os.Exit(1)
	}

	canvasWidth  := termCols * 2
	canvasHeight := termRows * 4

	grid := banner.SvgToBrailleArray(string(data), canvasWidth, canvasHeight)

	for ri, row := range grid {
		for ci, cell := range row {
			fmt.Printf("%04X,%d,%d,%d",
				cell.Char,
				cell.Color.R,
				cell.Color.G,
				cell.Color.B)
			if ci < len(row)-1 {
				fmt.Print("|")
			}
		}
		if ri < len(grid)-1 {
			fmt.Print("\n")
		}
	}
	fmt.Print("\n")
}
