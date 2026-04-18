/**
 * @file jreng_tui_rectangle.h
 * @brief Cell-coordinate rectangle — mirrors juce::Rectangle<int> API.
 *
 * jreng::tui::Rectangle represents an axis-aligned region of the terminal
 * cell grid.  Coordinates are in cell units (column, row), not pixels.
 *
 * The API mirrors juce::Rectangle<int> so callers already familiar with JUCE
 * layout code face zero learning curve.  All methods are constexpr and noexcept.
 */

#pragma once

namespace jreng::tui
{ /*____________________________________________________________________________*/

/**
 * @struct Rectangle
 * @brief An axis-aligned cell-grid rectangle.
 *
 * All coordinates are cell units.  x is the column, y is the row.
 * The struct is trivially copyable and passed by value (small type rule).
 */
struct Rectangle
{
    int x      { 0 };
    int y      { 0 };
    int width  { 0 };
    int height { 0 };

    //==========================================================================
    // Construction
    //==========================================================================

    /** Creates a zero-size rectangle at the origin. */
    Rectangle() = default;

    /** Creates a rectangle at (x_, y_) with the given dimensions. */
    constexpr Rectangle (int x_, int y_, int w, int h) noexcept
        : x { x_ }, y { y_ }, width { w }, height { h } {}

    /** Creates a rectangle at (0, 0) with the given dimensions. */
    constexpr Rectangle (int w, int h) noexcept
        : width { w }, height { h } {}

    //==========================================================================
    // Accessors
    //==========================================================================

    constexpr int getX()      const noexcept { return x; }
    constexpr int getY()      const noexcept { return y; }
    constexpr int getWidth()  const noexcept { return width; }
    constexpr int getHeight() const noexcept { return height; }

    /** Returns the column one past the right edge: x + width. */
    constexpr int getRight()  const noexcept { return x + width; }

    /** Returns the row one past the bottom edge: y + height. */
    constexpr int getBottom() const noexcept { return y + height; }

    //==========================================================================
    // JUCE interop
    //==========================================================================

    /** Constructs a Rectangle from a juce::Rectangle<int>.
     *  Use only at juce::Component API boundaries (getBounds, setBounds).
     *  All fields map 1:1 — juce::Component bounds are set in cell units.
     */
    static Rectangle fromJuce (juce::Rectangle<int> r) noexcept
    {
        return { r.getX(), r.getY(), r.getWidth(), r.getHeight() };
    }

    /** Converts to juce::Rectangle<int>.
     *  Use only when passing back to juce::Component API (setBounds, etc.).
     */
    juce::Rectangle<int> toJuce() const noexcept
    {
        return { x, y, width, height };
    }

    //==========================================================================
    // State
    //==========================================================================

    /** Returns true when either dimension is zero or negative. */
    constexpr bool isEmpty() const noexcept { return width <= 0 or height <= 0; }

    constexpr bool operator== (const Rectangle& other) const noexcept
    {
        return x == other.x and y == other.y
           and width == other.width and height == other.height;
    }

    constexpr bool operator!= (const Rectangle& other) const noexcept
    {
        return not (*this == other);
    }

    //==========================================================================
    // Containment
    //==========================================================================

    /** Returns true when the cell at (px, py) is inside this rectangle. */
    constexpr bool contains (int px, int py) const noexcept
    {
        return px >= x and py >= y and px < x + width and py < y + height;
    }

    /** Returns true when other is fully contained within this rectangle. */
    constexpr bool contains (const Rectangle& other) const noexcept
    {
        return other.x >= x and other.y >= y
           and other.getRight()  <= getRight()
           and other.getBottom() <= getBottom();
    }

    //==========================================================================
    // Fluent builders — each returns a new Rectangle, does not mutate self
    //==========================================================================

    constexpr Rectangle withX      (int newX) const noexcept { return { newX,  y,    width,    height }; }
    constexpr Rectangle withY      (int newY) const noexcept { return { x,     newY, width,    height }; }
    constexpr Rectangle withWidth  (int newW) const noexcept { return { x,     y,    newW,     height }; }
    constexpr Rectangle withHeight (int newH) const noexcept { return { x,     y,    width,    newH   }; }

    constexpr Rectangle withPosition (int newX, int newY) const noexcept
    {
        return { newX, newY, width, height };
    }

    constexpr Rectangle withSize (int newW, int newH) const noexcept
    {
        return { x, y, newW, newH };
    }

    /** Returns a copy shifted by (dx, dy). */
    constexpr Rectangle translated (int dx, int dy) const noexcept
    {
        return { x + dx, y + dy, width, height };
    }

    /**
     * Returns a copy shrunk by dx cells on each horizontal side and dy cells
     * on each vertical side.  Result may have negative dimensions if the
     * reduction exceeds the rectangle size — callers that need a safe clamp
     * should check isEmpty() on the result.
     */
    constexpr Rectangle reduced (int dx, int dy) const noexcept
    {
        return { x + dx, y + dy, width - dx * 2, height - dy * 2 };
    }

    /** Returns a copy with the left edge moved right by amount. */
    constexpr Rectangle withTrimmedLeft (int amount) const noexcept
    {
        return { x + amount, y, width - amount, height };
    }

    /** Returns a copy with the right edge moved left by amount. */
    constexpr Rectangle withTrimmedRight (int amount) const noexcept
    {
        return { x, y, width - amount, height };
    }

    /** Returns a copy with the top edge moved down by amount. */
    constexpr Rectangle withTrimmedTop (int amount) const noexcept
    {
        return { x, y + amount, width, height - amount };
    }

    /** Returns a copy with the bottom edge moved up by amount. */
    constexpr Rectangle withTrimmedBottom (int amount) const noexcept
    {
        return { x, y, width, height - amount };
    }

    //==========================================================================
    // Slicing — mutates self, returns the removed strip
    //==========================================================================

    /**
     * Removes a strip of the given height from the top of this rectangle.
     * Self shrinks by amount rows.  Returns the removed strip.
     */
    Rectangle removeFromTop (int amount) noexcept
    {
        const Rectangle strip { x, y, width, amount };
        y      += amount;
        height -= amount;
        return strip;
    }

    /**
     * Removes a strip of the given height from the bottom of this rectangle.
     * Self shrinks by amount rows.  Returns the removed strip.
     */
    Rectangle removeFromBottom (int amount) noexcept
    {
        height -= amount;
        return { x, y + height, width, amount };
    }

    /**
     * Removes a strip of the given width from the left of this rectangle.
     * Self shrinks by amount columns.  Returns the removed strip.
     */
    Rectangle removeFromLeft (int amount) noexcept
    {
        const Rectangle strip { x, y, amount, height };
        x     += amount;
        width -= amount;
        return strip;
    }

    /**
     * Removes a strip of the given width from the right of this rectangle.
     * Self shrinks by amount columns.  Returns the removed strip.
     */
    Rectangle removeFromRight (int amount) noexcept
    {
        width -= amount;
        return { x + width, y, amount, height };
    }
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
