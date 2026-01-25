package banner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"tit/internal"
)

// Point represents a 2D coordinate
type Point struct {
	X float64
	Y float64
}

// Color represents an RGB color
type Color struct {
	R, G, B int
}

// PixelColor represents a pixel with RGB color
type PixelColor struct {
	R, G, B int
}

// ScanlineRange represents a filled horizontal range
type ScanlineRange struct {
	ScanY  int
	StartX int
	EndX   int
	Color  Color
}

// Intersection represents where an edge crosses a scanline
type Intersection struct {
	X         int
	Direction int
}

// parseFloatOrPanic parses string to float64 with panic on failure
// FAIL FAST: Panics immediately if parsing fails, enabling early debugging
func parseFloatOrPanic(value string) float64 {
	result, err := strconv.ParseFloat(value, internal.FloatParseBitSize)
	if err != nil {
		panic(fmt.Sprintf("SVG parse error: failed to parse float '%s': %v", value, err))
	}
	return result
}

// parseIntOrPanic parses string to int64 with panic on failure
// FAIL FAST: Panics immediately if parsing fails, enabling early debugging
func parseIntOrPanic(value string, base, bitSize int) int {
	result, err := strconv.ParseInt(value, base, bitSize)
	if err != nil {
		panic(fmt.Sprintf("SVG parse error: failed to parse int '%s': %v", value, err))
	}
	return int(result)
}

// extractSvgDimensions extracts width/height from SVG string
func extractSvgDimensions(svgString string) (width, height float64) {
	// Try explicit width/height attributes
	widthRe := regexp.MustCompile(`width\s*=\s*["']?(\d+(?:\.\d+)?)`)
	heightRe := regexp.MustCompile(`height\s*=\s*["']?(\d+(?:\.\d+)?)`)

	widthMatch := widthRe.FindStringSubmatch(svgString)
	heightMatch := heightRe.FindStringSubmatch(svgString)

	if len(widthMatch) > 1 && len(heightMatch) > 1 {
		w := parseFloatOrPanic(widthMatch[1])
		h := parseFloatOrPanic(heightMatch[1])
		return w, h
	}

	// Fall back to viewBox
	viewBoxRe := regexp.MustCompile(`viewBox\s*=\s*["']([^"']*)["']`)
	viewBoxMatch := viewBoxRe.FindStringSubmatch(svgString)

	if len(viewBoxMatch) > 1 {
		parts := strings.FieldsFunc(viewBoxMatch[1], func(r rune) bool {
			return r == ' ' || r == ','
		})
		if len(parts) >= 4 {
			w := parseFloatOrPanic(parts[2])
			h := parseFloatOrPanic(parts[3])
			return w, h
		}
	}

	// Default fallback
	return 400, 320
}

// parseFillPaths extracts all fill paths with colors from SVG
func parseFillPaths(svgString string) []struct {
	Subpaths [][]Point
	Color    Color
} {
	var fillPaths []struct {
		Subpaths [][]Point
		Color    Color
	}

	// Extract all <path> elements (handles multi-line paths and both /> and > closings)
	pathRe := regexp.MustCompile(`(?s)<path\s+([^>]*)>`)
	matches := pathRe.FindAllStringSubmatch(svgString, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		pathAttrs := match[1]

		// Extract d attribute (path data)
		dRe := regexp.MustCompile(`d="([^"]*)`)
		dMatch := dRe.FindStringSubmatch(pathAttrs)
		if len(dMatch) < 2 {
			continue
		}
		pathData := dMatch[1]

		// Extract style attribute
		styleRe := regexp.MustCompile(`style="([^"]*)`)
		styleMatch := styleRe.FindStringSubmatch(pathAttrs)
		if len(styleMatch) < 2 {
			continue
		}
		style := styleMatch[1]

		// Check if it has fill (not fill:none)
		if strings.Contains(style, "fill:rgb") && !strings.Contains(style, "fill:none") {
			subpaths := parsePathDataAsSubpaths(pathData)
			color := parseColor(style)
			fillPaths = append(fillPaths, struct {
				Subpaths [][]Point
				Color    Color
			}{subpaths, color})
		}
	}

	return fillPaths
}

// approximateCubicBezier approximates a cubic Bezier curve as line segments
func approximateCubicBezier(p0, cp1, cp2, p1 Point, segments int) []Point {
	var points []Point

	for i := 0; i <= segments; i++ {
		t := float64(i) / float64(segments)
		mt := 1 - t

		// Cubic Bezier formula
		x := mt*mt*mt*p0.X +
			3*mt*mt*t*cp1.X +
			3*mt*t*t*cp2.X +
			t*t*t*p1.X

		y := mt*mt*mt*p0.Y +
			3*mt*mt*t*cp1.Y +
			3*mt*t*t*cp2.Y +
			t*t*t*p1.Y

		points = append(points, Point{x, y})
	}

	return points
}

// parsePathDataAsSubpaths parses SVG path "d" attribute into subpaths
func parsePathDataAsSubpaths(pathData string) [][]Point {
	var subpaths [][]Point

	// Normalize: ensure spaces around commands (both uppercase and lowercase)
	normalized := pathData
	normalized = regexp.MustCompile(`([MLCSZmlcsz])`).ReplaceAllString(normalized, " $1 ")
	normalized = strings.ReplaceAll(normalized, ",", " ")
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	tokens := strings.Fields(normalized)

	var currentPath []Point
	currentPos := Point{0, 0}
	var lastControlPoint *Point
	i := 0

	for i < len(tokens) {
		cmd := tokens[i]

		switch cmd {
		case "M", "m":
			if len(currentPath) > 0 {
				subpaths = append(subpaths, currentPath)
				currentPath = nil
			}

			if i+2 < len(tokens) {
				x := parseFloatOrPanic(tokens[i+1])
				y := parseFloatOrPanic(tokens[i+2])

				if cmd == "M" {
					currentPos = Point{x, y}
				} else {
					currentPos = Point{currentPos.X + x, currentPos.Y + y}
				}
				currentPath = append(currentPath, currentPos)
				lastControlPoint = nil
				i += 3
			} else {
				i++
			}

		case "L", "l":
			if i+2 < len(tokens) {
				x := parseFloatOrPanic(tokens[i+1])
				y := parseFloatOrPanic(tokens[i+2])

				if cmd == "L" {
					currentPos = Point{x, y}
				} else {
					currentPos = Point{currentPos.X + x, currentPos.Y + y}
				}
				currentPath = append(currentPath, currentPos)
				lastControlPoint = nil
				i += 3
			} else {
				i++
			}

		case "C", "c":
			if i+6 < len(tokens) {
				cp1x := parseFloatOrPanic(tokens[i+1])
				cp1y := parseFloatOrPanic(tokens[i+2])
				cp2x := parseFloatOrPanic(tokens[i+3])
				cp2y := parseFloatOrPanic(tokens[i+4])
				x := parseFloatOrPanic(tokens[i+5])
				y := parseFloatOrPanic(tokens[i+6])

				var cp1, cp2, endPoint Point

				if cmd == "C" {
					cp1 = Point{cp1x, cp1y}
					cp2 = Point{cp2x, cp2y}
					endPoint = Point{x, y}
				} else {
					cp1 = Point{currentPos.X + cp1x, currentPos.Y + cp1y}
					cp2 = Point{currentPos.X + cp2x, currentPos.Y + cp2y}
					endPoint = Point{currentPos.X + x, currentPos.Y + y}
				}

				curvePoints := approximateCubicBezier(currentPos, cp1, cp2, endPoint, internal.BezierCurveResolution)
				if len(curvePoints) > 1 {
					currentPath = append(currentPath, curvePoints[1:]...)
				}

				currentPos = endPoint
				lastControlPoint = &cp2
				i += 7
			} else {
				i++
			}

		case "S", "s":
			if i+4 < len(tokens) {
				cp2x := parseFloatOrPanic(tokens[i+1])
				cp2y := parseFloatOrPanic(tokens[i+2])
				x := parseFloatOrPanic(tokens[i+3])
				y := parseFloatOrPanic(tokens[i+4])

				var cp1 Point
				if lastControlPoint != nil {
					cp1 = Point{
						2*currentPos.X - lastControlPoint.X,
						2*currentPos.Y - lastControlPoint.Y,
					}
				} else {
					cp1 = currentPos
				}

				var cp2, endPoint Point

				if cmd == "S" {
					cp2 = Point{cp2x, cp2y}
					endPoint = Point{x, y}
				} else {
					cp2 = Point{currentPos.X + cp2x, currentPos.Y + cp2y}
					endPoint = Point{currentPos.X + x, currentPos.Y + y}
				}

				curvePoints := approximateCubicBezier(currentPos, cp1, cp2, endPoint, internal.BezierCurveResolution)
				if len(curvePoints) > 1 {
					currentPath = append(currentPath, curvePoints[1:]...)
				}

				currentPos = endPoint
				lastControlPoint = &cp2
				i += 5
			} else {
				i++
			}

		case "Z", "z":
			if len(currentPath) > 0 {
				first := currentPath[0]
				last := currentPath[len(currentPath)-1]
				if first.X != last.X || first.Y != last.Y {
					currentPath = append(currentPath, first)
				}
			}
			lastControlPoint = nil
			i++

		default:
			i++
		}
	}

	if len(currentPath) > 0 {
		subpaths = append(subpaths, currentPath)
	}

	return subpaths
}

// parseColor extracts RGB color from SVG style attribute
func parseColor(styleAttr string) Color {
	// Try RGB format: rgb(255, 0, 0)
	rgbRe := regexp.MustCompile(`rgb\((\d+),\s*(\d+),\s*(\d+)\)`)
	rgbMatch := rgbRe.FindStringSubmatch(styleAttr)
	if len(rgbMatch) == 4 {
		r := parseIntOrPanic(rgbMatch[1], 10, 16)
		g := parseIntOrPanic(rgbMatch[2], 10, 16)
		b := parseIntOrPanic(rgbMatch[3], 10, 16)
		return Color{r, g, b}
	}

	// Try HEX format: #RRGGBB
	hexRe := regexp.MustCompile(`#([0-9A-Fa-f]{6})`)
	hexMatch := hexRe.FindStringSubmatch(styleAttr)
	if len(hexMatch) == 2 {
		hex := hexMatch[1]
		r := parseIntOrPanic(hex[0:2], 16, internal.IntParseBitSize)
		g := parseIntOrPanic(hex[2:4], 16, internal.IntParseBitSize)
		b := parseIntOrPanic(hex[4:6], 16, internal.IntParseBitSize)
		return Color{int(r), int(g), int(b)}
	}

	// Default white
	return Color{255, 255, 255}
}

// computeScanlineRanges computes filled pixel ranges using non-zero winding rule
func computeScanlineRanges(allSubpaths [][]Point, color Color, scaleX, scaleY, offsetX, offsetY float64) []ScanlineRange {
	var ranges []ScanlineRange

	if len(allSubpaths) == 0 {
		return ranges
	}

	// Scale and offset all points
	var scaledSubpaths [][]Point
	for _, subpath := range allSubpaths {
		var scaledPath []Point
		for _, pt := range subpath {
			scaledPath = append(scaledPath, Point{
				pt.X*scaleX + offsetX,
				pt.Y*scaleY + offsetY,
			})
		}
		scaledSubpaths = append(scaledSubpaths, scaledPath)
	}

	// Find bounds
	minY := 10000.0
	maxY := -10000.0
	for _, subpath := range scaledSubpaths {
		for _, pt := range subpath {
			if pt.Y < minY {
				minY = pt.Y
			}
			if pt.Y > maxY {
				maxY = pt.Y
			}
		}
	}

	// Scanline fill using non-zero winding rule
	for scanY := int(minY); scanY < int(maxY)+1; scanY++ {
		scanYFloat := float64(scanY)
		var intersections []Intersection

		// Collect intersections from all subpaths
		for _, scaledPoints := range scaledSubpaths {
			for i := 0; i < len(scaledPoints)-1; i++ {
				y1 := scaledPoints[i].Y
				y2 := scaledPoints[i+1].Y

				// Check if edge crosses scanline
				if (y1 < scanYFloat && y2 >= scanYFloat) || (y2 < scanYFloat && y1 >= scanYFloat) {
					x1 := scaledPoints[i].X
					x2 := scaledPoints[i+1].X

					// Calculate intersection x
					intersectX := x1 + (scanYFloat-y1)*(x2-x1)/(y2-y1)
					ix := int(intersectX)

					// Direction: +1 if going up, -1 if going down
					direction := 1
					if y1 >= y2 {
						direction = -1
					}
					intersections = append(intersections, Intersection{ix, direction})
				}
			}
		}

		// Sort by x coordinate
		for i := 0; i < len(intersections); i++ {
			for j := i + 1; j < len(intersections); j++ {
				if intersections[j].X < intersections[i].X {
					intersections[i], intersections[j] = intersections[j], intersections[i]
				}
			}
		}

		// Apply non-zero winding rule
		windingCount := 0
		fillStartX := -1

		for _, inter := range intersections {
			// Check if we were filling and winding becomes 0
			if windingCount != 0 && windingCount+inter.Direction == 0 {
				ranges = append(ranges, ScanlineRange{
					ScanY:  scanY,
					StartX: fillStartX,
					EndX:   inter.X,
					Color:  color,
				})
			}
			// Check if we weren't filling and winding becomes non-zero
			if windingCount == 0 && windingCount+inter.Direction != 0 {
				fillStartX = inter.X
			}

			windingCount += inter.Direction
		}
	}

	return ranges
}

// RenderSvgToBraille renders SVG to a pixel canvas
func RenderSvgToBraille(svgString string, width, height int) [][]PixelColor {
	// Create canvas
	canvas := make([][]PixelColor, height)
	for i := range canvas {
		canvas[i] = make([]PixelColor, width)
		for j := range canvas[i] {
			canvas[i][j] = PixelColor{0, 0, 0}
		}
	}

	fillPaths := parseFillPaths(svgString)

	// Extract SVG dimensions and calculate uniform scale + centering
	svgWidth, svgHeight := extractSvgDimensions(svgString)
	offsetX := 0.0
	offsetY := 0.0

	// Use uniform scale (smaller of the two to fit within canvas)
	scaleX := float64(width) / svgWidth
	scaleY := float64(height) / svgHeight
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	// Center both horizontally and vertically
	scaledWidth := svgWidth * scale
	scaledHeight := svgHeight * scale
	offsetX = (float64(width) - scaledWidth) / 2
	offsetY = (float64(height) - scaledHeight) / 2

	// Draw all fill paths
	for _, fp := range fillPaths {
		ranges := computeScanlineRanges(fp.Subpaths, fp.Color, scale, scale, offsetX, offsetY)

		// Draw filled ranges
		for _, r := range ranges {
			for x := r.StartX; x <= r.EndX; x++ {
				if r.ScanY >= 0 && r.ScanY < height && x >= 0 && x < width {
					canvas[r.ScanY][x] = PixelColor{r.Color.R, r.Color.G, r.Color.B}
				}
			}
		}
	}

	return canvas
}
