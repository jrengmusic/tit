// ============================================================================
// jreng::braille — SVG-to-braille rasterization pipeline
// ============================================================================

namespace jreng::braille
{ /*____________________________________________________________________________*/

//==============================================================================
// Internal helpers
//==============================================================================

/** Extracts the axis-aligned bounding box of a set of scaled+offset subpaths. */
static juce::Rectangle<float> computeSubpathBounds (
    const std::vector<std::vector<juce::Point<float>>>& subpaths,
    float scaleX,
    float scaleY,
    float offsetX,
    float offsetY)
{
    float minY {  1.0e9f };
    float maxY { -1.0e9f };

    for (const auto& subpath : subpaths)
    {
        for (const auto& pt : subpath)
        {
            const float sy { pt.y * scaleY + offsetY };
            minY = juce::jmin (minY, sy);
            maxY = juce::jmax (maxY, sy);
        }
    }

    return { 0.0f, minY, 0.0f, maxY };
}

/** Insertion-sort `intersections` by x coordinate (ascending).

    Chosen for stability on small arrays (typically < 10 elements per scanline).
    Mirrors the bubble-sort in the Go reference implementation.
*/
static void sortIntersectionsByX (std::vector<Intersection>& intersections)
{
    const int n { static_cast<int> (intersections.size()) };

    for (int i { 1 }; i < n; ++i)
    {
        const Intersection key { intersections.at (static_cast<size_t> (i)) };
        int j { i - 1 };

        while (j >= 0 and intersections.at (static_cast<size_t> (j)).x > key.x)
        {
            intersections.at (static_cast<size_t> (j + 1)) =
                intersections.at (static_cast<size_t> (j));
            j -= 1;
        }

        intersections.at (static_cast<size_t> (j + 1)) = key;
    }
}

/** Tests whether edge [e, e+1] straddles scanline `scanYFloat` and, if so,
    appends the interpolated Intersection to `intersections`.

    Direction is +1 when the edge goes downward (y increasing), -1 when upward,
    consistent with the non-zero winding convention used in the Go reference.
*/
static void appendEdgeIntersection (
    const std::vector<juce::Point<float>>& subpath,
    int edgeIndex,
    float scaleX,
    float scaleY,
    float offsetX,
    float offsetY,
    float scanYFloat,
    std::vector<Intersection>& intersections)
{
    const float y1 { subpath.at (static_cast<size_t> (edgeIndex)).y     * scaleY + offsetY };
    const float y2 { subpath.at (static_cast<size_t> (edgeIndex + 1)).y * scaleY + offsetY };

    const bool crosses { (y1 < scanYFloat and y2 >= scanYFloat)
                      or (y2 < scanYFloat and y1 >= scanYFloat) };

    if (crosses)
    {
        const float x1 { subpath.at (static_cast<size_t> (edgeIndex)).x     * scaleX + offsetX };
        const float x2 { subpath.at (static_cast<size_t> (edgeIndex + 1)).x * scaleX + offsetX };

        const float t          { (scanYFloat - y1) / (y2 - y1) };
        const float intersectX { x1 + t * (x2 - x1) };
        const int   ix         { static_cast<int> (intersectX) };

        const int direction { y1 < y2 ? 1 : -1 };

        intersections.push_back ({ ix, direction });
    }
}

/** Collects edge-scanline intersections for all subpaths at pixel row `scanY`.

    An edge from point[i] to point[i+1] contributes an intersection when it
    straddles the horizontal line y = scanY.  The x-coordinate is interpolated
    linearly.  Direction is +1 when the edge goes downward (increasing y), -1
    when upward — consistent with the non-zero winding convention used in the Go
    reference implementation.
*/
static std::vector<Intersection> collectIntersections (
    const std::vector<std::vector<juce::Point<float>>>& subpaths,
    float scaleX,
    float scaleY,
    float offsetX,
    float offsetY,
    float scanYFloat)
{
    std::vector<Intersection> intersections;

    for (const auto& subpath : subpaths)
    {
        const int edgeCount { static_cast<int> (subpath.size()) - 1 };
        jassert (edgeCount >= 0);

        for (int e { 0 }; e < edgeCount; ++e)
            appendEdgeIntersection (subpath, e, scaleX, scaleY, offsetX, offsetY,
                                    scanYFloat, intersections);
    }

    return intersections;
}

/** Applies the non-zero winding rule to a sorted intersection list to produce
    filled ScanlineRange spans for a single pixel row.
*/
static void applyWindingRule (std::vector<Intersection>& intersections,
                              int scanY,
                              juce::Colour color,
                              std::vector<ScanlineRange>& ranges)
{
    int windingCount { 0 };
    int fillStartX   { -1 };

    for (const auto& inter : intersections)
    {
        const bool wasInside { windingCount != 0 };
        windingCount += inter.direction;
        const bool isInside  { windingCount != 0 };

        if (wasInside and not isInside)
        {
            // Exiting fill — emit span
            ranges.push_back ({ scanY, fillStartX, inter.x, color });
        }

        if (not wasInside and isInside)
        {
            // Entering fill
            fillStartX = inter.x;
        }
    }
}

//==============================================================================
// extractSubpaths — pulls polylines from a juce::Path via PathFlatteningIterator
//==============================================================================

/** Collects the start point of each sub-path by scanning Path::Iterator elements.

    Returns one entry per startNewSubPath element encountered, in order.
    Called as the first pass inside extractSubpaths before the flattening pass.
*/
static std::vector<juce::Point<float>> collectSubpathStartPoints (
    const juce::Path& path)
{
    std::vector<juce::Point<float>> startPoints;

    juce::Path::Iterator rawIt { path };

    while (rawIt.next())
    {
        if (rawIt.elementType == juce::Path::Iterator::startNewSubPath)
            startPoints.push_back ({ rawIt.x1, rawIt.y1 });
    }

    return startPoints;
}

/** Flattens `path` into a vector of polylines using PathFlatteningIterator.

    Each sub-path of the juce::Path becomes one entry in the returned vector.
    Points are in the path's own coordinate space (before any canvas scale).

    PathFlatteningIterator converts all bezier curves to straight line segments
    using JUCE's tolerance-based adaptive subdivision.  Each returned polyline
    starts at the sub-path origin and ends at the last point emitted before
    a closePath or the path end.  When closePath is reached the sub-path is
    explicitly closed by appending the start point again — matching Go's Z/z
    handling.
*/
static std::vector<std::vector<juce::Point<float>>> extractSubpaths (
    const juce::Path& path)
{
    const std::vector<juce::Point<float>> startPoints {
        collectSubpathStartPoints (path)
    };

    std::vector<std::vector<juce::Point<float>>> subpaths;
    std::vector<juce::Point<float>> currentSubpath;
    int subPathIndex { 0 };

    juce::PathFlatteningIterator flatIt { path };

    while (flatIt.next())
    {
        if (currentSubpath.empty() and subPathIndex < static_cast<int> (startPoints.size()))
            currentSubpath.push_back (startPoints.at (static_cast<size_t> (subPathIndex)));

        currentSubpath.push_back ({ flatIt.x2, flatIt.y2 });

        if (flatIt.closesSubPath)
        {
            if (not currentSubpath.empty())
                currentSubpath.push_back (currentSubpath.front());

            subpaths.push_back (std::move (currentSubpath));
            currentSubpath.clear();
            subPathIndex += 1;
        }
    }

    if (not currentSubpath.empty())
        subpaths.push_back (std::move (currentSubpath));

    return subpaths;
}

//==============================================================================
// renderSvgToPixelCanvas helpers
//==============================================================================

/** Paints a vector of ScanlineRange spans onto `canvas`, clipped to canvas bounds. */
static void paintScanlineRanges (const std::vector<ScanlineRange>& ranges,
                                 PixelCanvas& canvas)
{
    for (const auto& range : ranges)
    {
        const bool rowInBounds { range.scanY >= 0 and range.scanY < canvas.height };

        if (rowInBounds)
        {
            const int xBegin { juce::jmax (0, range.startX) };
            const int xEnd   { juce::jmin (canvas.width - 1, range.endX) };

            for (int x { xBegin }; x <= xEnd; ++x)
            {
                canvas.pixels.at (
                    static_cast<size_t> (range.scanY * canvas.width + x)) = range.color;
            }
        }
    }
}

/** Processes one DrawablePath child: extracts subpaths, rasterizes, paints onto `canvas`.

    Does nothing when the child is not a DrawablePath, its fill is not a colour,
    or its fill alpha is zero.
*/
static void processSvgDrawablePath (const juce::Component* child,
                                    float scale,
                                    float offsetX,
                                    float offsetY,
                                    PixelCanvas& canvas)
{
    const auto* drawablePath { dynamic_cast<const juce::DrawablePath*> (child) };

    const bool isDrawablePath { drawablePath != nullptr };

    if (isDrawablePath)
    {
        const juce::FillType& fill { drawablePath->getFill() };

        const bool isColourFill { fill.isColour() };

        if (isColourFill)
        {
            const juce::Colour pathColor { fill.colour };

            const bool isVisible { pathColor.getAlpha() != 0 };

            if (isVisible)
            {
                const juce::Path& juicePath { drawablePath->getPath() };

                const std::vector<std::vector<juce::Point<float>>> subpaths {
                    extractSubpaths (juicePath)
                };

                const std::vector<ScanlineRange> pathRanges {
                    computeScanlineRanges (subpaths, pathColor, scale, scale, offsetX, offsetY)
                };

                paintScanlineRanges (pathRanges, canvas);
            }
        }
    }
}

//==============================================================================
// encodeCanvasToBrailleGrid helpers
//==============================================================================

/** Per-pixel sample result for the braille encoder: brightness and color. */
struct BraillePixelSample
{
    int          brightness { 0 };
    juce::Colour color;
    bool         inBounds   { false };
};

/** Samples one pixel from `canvas` at (`px`, `py`).

    Returns inBounds=false when the coordinates fall outside the canvas.
    Brightness is the sum of the three 8-bit channels (range 0–765).
*/
static BraillePixelSample sampleBraillePixelAt (const PixelCanvas& canvas,
                                                int px,
                                                int py)
{
    const bool inBounds { px < canvas.width and py < canvas.height };

    BraillePixelSample sample;
    sample.inBounds = inBounds;

    if (inBounds)
    {
        sample.color = canvas.pixels.at (
            static_cast<size_t> (py * canvas.width + px));

        sample.brightness = static_cast<int> (sample.color.getRed())
                          + static_cast<int> (sample.color.getGreen())
                          + static_cast<int> (sample.color.getBlue());
    }

    return sample;
}

/** Encodes a single 2×4 pixel block at (`pixelX`, `pixelY`) into a tui::Cell.

    The braille codepoint is assembled from the 8 pixels of the block using the
    column-major dot ordering defined in jreng_braille_grid.h.
    The cell foreground is set to the brightest (highest R+G+B) pixel in the block.
*/
static jreng::tui::Cell encodeBrailleCellAt (const PixelCanvas& canvas,
                                             int pixelX,
                                             int pixelY)
{
    uint32_t     brailleBits   { 0 };
    juce::Colour dominantColor { juce::Colours::black };
    int          maxBrightness { -1 };

    for (int dy { 0 }; dy < BRAILLE_CELL_HEIGHT; ++dy)
    {
        for (int dx { 0 }; dx < BRAILLE_CELL_WIDTH; ++dx)
        {
            const BraillePixelSample sample {
                sampleBraillePixelAt (canvas, pixelX + dx, pixelY + dy)
            };

            if (sample.inBounds)
            {
                if (sample.brightness > BRAILLE_BRIGHTNESS_THRESHOLD)
                {
                    const int bitIndex { dx == 0 ? dy : dy + BRAILLE_CELL_HEIGHT };
                    brailleBits |= (1u << static_cast<unsigned> (bitIndex));
                }

                if (sample.brightness > maxBrightness)
                {
                    dominantColor = sample.color;
                    maxBrightness = sample.brightness;
                }
            }
        }
    }

    jreng::tui::Cell cell;
    cell.codepoint = BRAILLE_BASE_CODEPOINT | brailleBits;
    cell.width     = 1;
    cell.fg.setRGB (dominantColor.getRed(),
                    dominantColor.getGreen(),
                    dominantColor.getBlue());

    return cell;
}

//==============================================================================
// Public functions
//==============================================================================

std::vector<ScanlineRange> computeScanlineRanges (
    const std::vector<std::vector<juce::Point<float>>>& subpaths,
    juce::Colour color,
    float scaleX,
    float scaleY,
    float offsetX,
    float offsetY)
{
    jassert (scaleX > 0.0f);
    jassert (scaleY > 0.0f);

    std::vector<ScanlineRange> ranges;

    const bool hasSubpaths { not subpaths.empty() };

    if (hasSubpaths)
    {
        const juce::Rectangle<float> bounds {
            computeSubpathBounds (subpaths, scaleX, scaleY, offsetX, offsetY)
        };

        const int scanYMin { static_cast<int> (bounds.getY()) };
        const int scanYMax { static_cast<int> (bounds.getHeight()) + 1 };

        for (int scanY { scanYMin }; scanY < scanYMax; ++scanY)
        {
            const float scanYFloat { static_cast<float> (scanY) };

            std::vector<Intersection> intersections {
                collectIntersections (subpaths, scaleX, scaleY, offsetX, offsetY, scanYFloat)
            };

            sortIntersectionsByX (intersections);
            applyWindingRule (intersections, scanY, color, ranges);
        }
    }

    return ranges;
}

PixelCanvas renderSvgToPixelCanvas (const juce::XmlElement& svgDocument,
                                    int pixelWidth,
                                    int pixelHeight)
{
    jassert (pixelWidth  > 0);
    jassert (pixelHeight > 0);

    PixelCanvas canvas;
    canvas.width  = pixelWidth;
    canvas.height = pixelHeight;
    canvas.pixels.assign (static_cast<size_t> (pixelWidth * pixelHeight),
                          juce::Colours::black);

    std::unique_ptr<juce::Drawable> drawable {
        juce::Drawable::createFromSVG (svgDocument)
    };

    const bool drawableValid { drawable != nullptr };

    if (drawableValid)
    {
        const juce::Rectangle<float> svgBounds { drawable->getDrawableBounds() };

        const float svgW { svgBounds.getWidth() };
        const float svgH { svgBounds.getHeight() };

        const bool boundsValid { svgW > 0.0f and svgH > 0.0f };

        if (boundsValid)
        {
            const float scaleX { static_cast<float> (pixelWidth)  / svgW };
            const float scaleY { static_cast<float> (pixelHeight) / svgH };
            const float scale  { juce::jmin (scaleX, scaleY) };

            const float scaledW { svgW * scale };
            const float scaledH { svgH * scale };
            const float offsetX { (static_cast<float> (pixelWidth)  - scaledW) * 0.5f
                                  - svgBounds.getX() * scale };
            const float offsetY { (static_cast<float> (pixelHeight) - scaledH) * 0.5f
                                  - svgBounds.getY() * scale };

            const int childCount { drawable->getNumChildComponents() };

            for (int ci { 0 }; ci < childCount; ++ci)
            {
                processSvgDrawablePath (drawable->getChildComponent (ci),
                                        scale, offsetX, offsetY, canvas);
            }
        }
    }

    return canvas;
}

BrailleGrid encodeCanvasToBrailleGrid (const PixelCanvas& canvas)
{
    jassert (canvas.width  > 0);
    jassert (canvas.height > 0);

    BrailleGrid grid;
    grid.cols = canvas.width  / BRAILLE_CELL_WIDTH;
    grid.rows = canvas.height / BRAILLE_CELL_HEIGHT;
    grid.cells.resize (static_cast<size_t> (grid.cols * grid.rows));

    for (int cellRow { 0 }; cellRow < grid.rows; ++cellRow)
    {
        const int pixelY { cellRow * BRAILLE_CELL_HEIGHT };

        for (int cellCol { 0 }; cellCol < grid.cols; ++cellCol)
        {
            const int pixelX { cellCol * BRAILLE_CELL_WIDTH };

            grid.cells.at (static_cast<size_t> (cellRow * grid.cols + cellCol)) =
                encodeBrailleCellAt (canvas, pixelX, pixelY);
        }
    }

    return grid;
}

BrailleGrid renderSvgToBrailleGrid (const juce::XmlElement& svgDocument,
                                    int pixelWidth,
                                    int pixelHeight)
{
    jassert (pixelWidth  > 0);
    jassert (pixelHeight > 0);

    const PixelCanvas canvas {
        renderSvgToPixelCanvas (svgDocument, pixelWidth, pixelHeight)
    };

    return encodeCanvasToBrailleGrid (canvas);
}

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::braille
