/**
 * @file jreng_braille_grid.h
 * @brief SVG-to-braille rasterization pipeline.
 *
 * Converts a filled SVG document into a grid of jreng::tui::Cell values whose
 * codepoints are Unicode braille patterns (U+2800–U+28FF).  Each braille cell
 * covers a 2×4 pixel block of the intermediate PixelCanvas.  Fill color from
 * each SVG path is preserved and written into the Cell foreground.
 *
 * Entry points:
 *   - renderSvgToBrailleGrid()   — full pipeline: SVG → BrailleGrid
 *   - renderSvgToPixelCanvas()   — SVG → intermediate pixel canvas
 *   - computeScanlineRanges()    — polyline subpaths → filled scanline spans
 *   - encodeCanvasToBrailleGrid() — pixel canvas → BrailleGrid
 *
 * Namespace: jreng::braille
 * Physical location: modules/jreng_tui/braille/ (inside jreng_tui module)
 */

#pragma once

#include <vector>
#include <juce_graphics/juce_graphics.h>
#include <juce_gui_basics/juce_gui_basics.h>

#include "jreng_tui/ansi/jreng_color.h"
#include "jreng_tui/ansi/jreng_cell.h"

namespace jreng::braille
{ /*____________________________________________________________________________*/

//==============================================================================
// Constants
//==============================================================================

/** Pixel columns covered by one braille character. */
static constexpr int BRAILLE_CELL_WIDTH  { 2 };

/** Pixel rows covered by one braille character. */
static constexpr int BRAILLE_CELL_HEIGHT { 4 };

/** Minimum sum of R+G+B for a pixel to be treated as "lit" in braille encoding. */
static constexpr int BRAILLE_BRIGHTNESS_THRESHOLD { 50 };

/** First Unicode codepoint of the braille patterns block (U+2800). */
static constexpr uint32_t BRAILLE_BASE_CODEPOINT { 0x2800u };

//==============================================================================
// Data structures
//==============================================================================

/*____________________________________________________________________________*/
/** A single RGB pixel on the intermediate rasterization canvas.

    Coordinates are pixel-space: x in [0, width), y in [0, height).
    All channels are 8-bit unsigned.
*/
struct PixelCanvas
{
    /** Width of the canvas in pixels. */
    int width  { 0 };

    /** Height of the canvas in pixels. */
    int height { 0 };

    /** Flat row-major pixel storage: index = y * width + x. */
    std::vector<juce::Colour> pixels;
};

/*____________________________________________________________________________*/
/** Output grid of jreng::tui::Cell values, one cell per 2×4 pixel block.

    Grid dimensions in cells:
      - cols = canvas.width  / BRAILLE_CELL_WIDTH
      - rows = canvas.height / BRAILLE_CELL_HEIGHT

    Storage is row-major: index = row * cols + col.
*/
struct BrailleGrid
{
    /** Number of cell columns. */
    int cols { 0 };

    /** Number of cell rows. */
    int rows { 0 };

    /** Flat row-major cell storage. */
    std::vector<jreng::tui::Cell> cells;
};

/*____________________________________________________________________________*/
/** A filled horizontal pixel span produced by the scanline rasterizer.

    StartX and endX are inclusive pixel column indices.
    ScanY is the pixel row index.
    Color is the fill color of the SVG path that produced this range.
*/
struct ScanlineRange
{
    int          scanY  { 0 };
    int          startX { 0 };
    int          endX   { 0 };
    juce::Colour color;
};

/*____________________________________________________________________________*/
/** Intersection of a polyline edge with a scanline.

    X is the column of the intersection.
    direction is +1 when the edge goes upward (y decreasing) and -1 when it
    goes downward (y increasing), per the non-zero winding rule.
*/
struct Intersection
{
    int x         { 0 };
    int direction { 0 };
};

//==============================================================================
// Pipeline functions
//==============================================================================

/*____________________________________________________________________________*/
/** Full pipeline: parse SVG, rasterize paths, encode to braille grid.

    @param svgDocument  Parsed SVG XML element. Must not be null.
    @param pixelWidth   Width of the intermediate pixel canvas in pixels.
                        Typically terminalCols * BRAILLE_CELL_WIDTH.
    @param pixelHeight  Height of the intermediate pixel canvas in pixels.
                        Typically terminalRows * BRAILLE_CELL_HEIGHT.
    @returns            BrailleGrid sized pixelWidth/2 × pixelHeight/4.
*/
BrailleGrid renderSvgToBrailleGrid (const juce::XmlElement& svgDocument,
                                    int pixelWidth,
                                    int pixelHeight);

/*____________________________________________________________________________*/
/** Rasterizes all filled SVG paths from `svgDocument` into a pixel canvas.

    Uses juce::Drawable::createFromSVG to parse the tree, then iterates
    DrawablePath children, extracts fill colors, and runs the scanline
    rasterizer for each path.

    @param svgDocument  Parsed SVG XML element. Must not be null.
    @param pixelWidth   Canvas width in pixels.
    @param pixelHeight  Canvas height in pixels.
    @returns            Pixel canvas of dimensions pixelWidth × pixelHeight.
*/
PixelCanvas renderSvgToPixelCanvas (const juce::XmlElement& svgDocument,
                                    int pixelWidth,
                                    int pixelHeight);

/*____________________________________________________________________________*/
/** Computes filled horizontal scanline spans for a set of polyline subpaths.

    Implements the non-zero winding rule.  Subpaths are polylines (line
    segments already tessellated from bezier curves by PathFlatteningIterator).

    The subpaths are scaled and offset before rasterization:
      pixel_x = point.x * scaleX + offsetX
      pixel_y = point.y * scaleY + offsetY

    @param subpaths  Vector of polylines; each polyline is a vector of 2D points.
    @param color     Fill color applied to all produced spans.
    @param scaleX    Horizontal scale factor.
    @param scaleY    Vertical scale factor.
    @param offsetX   Horizontal translation after scaling.
    @param offsetY   Vertical translation after scaling.
    @returns         Vector of ScanlineRange spans covering the filled region.
*/
std::vector<ScanlineRange> computeScanlineRanges (
    const std::vector<std::vector<juce::Point<float>>>& subpaths,
    juce::Colour color,
    float scaleX,
    float scaleY,
    float offsetX,
    float offsetY);

/*____________________________________________________________________________*/
/** Encodes a pixel canvas into a braille cell grid.

    Each 2×4 block of pixels maps to one braille codepoint.  The 8 dots of the
    braille character correspond to the 8 pixels in column-major order:
      dot 0 → (dx=0, dy=0)    dot 4 → (dx=1, dy=0)
      dot 1 → (dx=0, dy=1)    dot 5 → (dx=1, dy=1)
      dot 2 → (dx=0, dy=2)    dot 6 → (dx=1, dy=2)
      dot 3 → (dx=0, dy=3)    dot 7 → (dx=1, dy=3)

    A pixel is "lit" when R+G+B > BRAILLE_BRIGHTNESS_THRESHOLD.
    The codepoint is BRAILLE_BASE_CODEPOINT | (bit mask of lit dots).
    The cell foreground is set to the brightest pixel in the 2×4 block.

    @param canvas  Source pixel canvas.
    @returns       BrailleGrid of dimensions (width/2) × (height/4).
*/
BrailleGrid encodeCanvasToBrailleGrid (const PixelCanvas& canvas);

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::braille
