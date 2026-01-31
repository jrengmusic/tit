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
	viewBoxRe := regexp.MustCompile(`viewBox\s*=\s*["']([^"]*)["']`)
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
