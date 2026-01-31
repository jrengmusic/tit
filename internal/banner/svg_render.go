package banner

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
