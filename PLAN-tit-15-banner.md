# PLAN: TIT Part 15 — The Banner

**RFC:** none — objective from ARCHITECT prompt (`/goplan`) and the tracker row 15. Canon: `PLAN-tit-0-tracker.md`, rulings 23, 24, 25, 28, 29.
**Date:** 2026-11-14
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

TIT shows its logo, `Assets/tit_logo.svg`, as a `juce::Image`. The part does these things:

1. JAM gets `jam::BrailleComponent`. It shows a `juce::Image` as braille cells, two dots across and four dots down in each cell (ruling 25). It reads the braille dot table of JAM, which Step 1 moves to two cast lookup tables (ruling 50).
2. `titc --help` prints the banner above the guide, always in braille. The guide can go to a pipe or a file (Part 7, item 3), and braille is text.
3. `TITView` shows the banner in the right half of the content, beside the main menus, Config, and Preferences, as legacy does.
4. Ruling 23: *"BOTH, decide at execution. but implement both component"*. The pixel banner is `juce::ImageComponent` with the same image, through the inline image lane of Part 3, which the peer of Part 5 turns on. Step 8 shows both in the window, and ARCHITECT keeps one.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 1, 3, 5, 10, and 13.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE; CODING.md, "One object, one translation unit": `BrailleComponent` is one header and one `.cpp`. `TerminalGlyph` is a header-only struct today (`jam_graphics/fonts/jam_TerminalGlyph.h`); this part adds `constexpr` members only, which stay in the header (CODING.md, "What stays in the header").
- CODING.md: no include in a submodule header or source file. `jam_gui.h` and `jam_gui.cpp` hold the includes.
- CODING.md: no bail-out guard, no `==` on strings, no out-parameter, no fake-carrier struct, `.at()` for container access, `not`/`and`/`or`.
- Ruling 50: *"braille dot table should not be hand-authored use LUT with cast table manifest"*. The table is two `## ` tables of `jam/cast/graphics.md`, generated as `jam::LookupTable` in `namespace map`, in the form of the `os2UnicodeRange` tables (`jam/cast/opentype.md:65`; `jam/cast/spell.md:93,1625-1651`). `TerminalGlyph::drawBraille` and `BrailleComponent` read it.
- MANIFESTO E, Materialisation: the logo image and the dot image are concrete objects. Each owner builds its object once and replaces it when its input changes.
- CAROL.md, Code Hygiene: no comment and no doxygen block. The doxygen pass is separate. The existing doxygen line of `brailleFirst` moves with it.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`.

### The legacy banner

| Fact | Where |
|---|---|
| The banner is the logo, rasterized to `width * 2` by `height * 4` dots, then one braille character for each 2 by 4 dots | `legacy/ui/layout.go:15-46`; `legacy/banner/braille.go:48-122` |
| A dot is filled when R + G + B of its pixel is above 50, on a black canvas | `legacy/banner/braille.go:76-79`; `legacy/banner/svg_render.go:105-112` |
| The colour of a cell is the colour of its brightest dot, by R + G + B | `legacy/banner/braille.go:96-106` |
| The logo is scaled uniformly, to fit, and centred in both directions | `legacy/banner/svg_render.go:116-133` |
| The banner fills the right half of the content beside the main menu, the Config menu, and the Preferences menu. The left half is `width / 2`. The location menus and the other modes show no banner. | `legacy/ui/menu.go:189-217`; `legacy/ui/preferences.go:174-197`; `legacy/app/app_view_main.go:15-16,86-89,141-142,159-165` |
| The reference grid of the banner: 40 columns by 10 rows | `tit/___legacy___/braille_ref/main.go:21-23` |
| The logo: 128 by 128 px, ten `<path>` elements, `fill-rule:evenodd` on the root | `tit/Assets/tit_logo.svg:3`; byte-identical to `legacy/ui/assets/tit-logo.svg` (`cmp`) |

### Legacy faults that this part corrects

1. **The dot bits.** Legacy sets bit `dy` for the left column and bit `dy + 4` for the right column (`legacy/banner/braille.go:80-86`). Unicode sets bit n − 1 for dot n. The left column holds dots 1, 2, 3, 7 from the top, and the right column holds dots 4, 5, 6, 8 (Unicode 18.0.0 Core Specification §21.1, the dot diagram; the character names `BRAILLE PATTERN DOTS-1` at U+2801 to `DOTS-8` at U+2880, Unicode database 17.0.0, read in this session). So legacy puts six of the eight dots in the wrong place. Legacy's own comment gives the right layout (`:60-64`). JAM's table is the Unicode table (`jam_TerminalGlyph.h:840-845`).
2. **The fill rule.** Legacy fills each path with nonzero winding (`legacy/banner/svg_render.go:3,77-97`). The logo says `fill-rule:evenodd` (`tit/Assets/tit_logo.svg:3`). JUCE reads the rule: `if (getStyleAttribute (xml, "fill-rule").trim().equalsIgnoreCase ("evenodd")) path.setUsingNonZeroWinding (false);` (`juce_gui_basics/drawables/juce_SVGParser.cpp:566-567`).
3. **A hand-written rasterizer.** Legacy parses the path data and fills scan lines by hand (`legacy/banner/svg_paths.go`, `svg_render.go`). This part uses `jam::Svg::getDrawable` and `juce::Graphics` (CODING.md, "Use the framework API fully").

### JAM

| Fact | Where |
|---|---|
| The braille codepoints: `brailleFirst { 0x2800 }`, `brailleLast { 0x28FF }`, private | `jam_graphics/fonts/jam_TerminalGlyph.h:240-244` |
| `drawBraille`: `cellW { w / 2 }`, `cellH { h / 4 }`; the hand-written local tables `dotCol { 0, 0, 0, 1, 1, 1, 0, 1 }` and `dotRow { 0, 1, 2, 0, 1, 2, 3, 3 }`, indexed by bit; `for (int i { 0 }; i < 8; ++i)` | `jam_TerminalGlyph.h:832-858` |
| `getCoverage` is public; the private section starts at `:227` | `jam_TerminalGlyph.h:202-227` |
| `static std::unique_ptr<juce::Drawable> getDrawable (const char* svgString);` with the body `juce::DrawableImage::createFromSVG (*juce::parseXML (svgString))` | `jam_graphics/svg/jam_Svg.h:661`; `jam_SvgDrawing.cpp:69-72` |
| `jam::ImageComponent`: `juce::Component`, `setInterceptsMouseClicks (false, false)` in its constructor, `setImage`, a `juce::Image image` member; a filmstrip display | `jam_gui/widgets/jam_ImageComponent.h:19-30,40,140`; tracker, banner table |
| `jam::PreviewComponent` builds its images with `juce::SoftwareImageType` | `jam_gui/widgets/jam_PreviewComponent.cpp:9,66,170` |
| The widget includes: `jam_gui.h:135` `jam_ImageComponent.h`, `:136` `jam_PreviewComponent.h`; `jam_gui.cpp:32` `jam_PreviewComponent.cpp` | `jam_gui/jam_gui.h:134-158`; `jam_gui/jam_gui.cpp:14-32` |
| The terminal context paints `juce::Graphics::drawText` into cells through the JAM patch, and colours the cells through `trueColours.getOrCreateColourId` | tracker, "The terminal side: what exists"; `jam_terminal/graphics/jam_TerminalGraphicsContext.cpp:302-325,734-750` |
| `getOrCreateColourId`: "When every slot holds a colour, the function returns the id of the nearest held colour" | doxygen `classjam_1_1Colours`; `jam_graphics/colours/jam_Colours.h` |
| `setOrigin (juce::Point<int>)`: "Moves the unit origin by o, through addTransform()"; a translation moves the origin to a whole cell | `jam_terminal/graphics/jam_TerminalGraphicsContext.h:95,107` |
| `drawImage` is the inline image lane only (Part 3 Step 4; ruling 28). The peer of Part 5 builds `jam::TerminalBuffer terminalBuffer { true }`. | `PLAN-tit-3-terminal-display.md` Step 4; `PLAN-tit-5-terminal-window.md` Step 1 |

### JUCE

| Fact | Where |
|---|---|
| `Image (PixelFormat, int imageWidth, int imageHeight, bool clearImage, const ImageType& type)`; `Colour getPixelAt (int x, int y) const` | `juce_graphics/images/juce_Image.h:117,261` |
| `Graphics (const Image& imageToDrawOnto)`; `drawImage (const Image&, Rectangle<float>, RectanglePlacement, bool)` | `juce_graphics/contexts/juce_GraphicsContext.h:66,563` |
| `ImageComponent::setImage (const Image&)`; the constructor sets `placement (RectanglePlacement::centred)` | `juce_gui_basics/widgets/juce_ImageComponent.h:59`; `.cpp:38-41` |

### TIT

| Fact | Where |
|---|---|
| `printHelp`: the shared instances, the style manager, `styleMarkdown.getColumnWidth()`, the editor, one frame, `writeToStream`, the newline of Part 3 | `ling/Source/Help.cpp:3-35`; `PLAN-tit-1-project.md` Step 9; `PLAN-tit-3-terminal-display.md` Step 3, item 8 |
| The help file form: `#pragma once`, `<JuceHeader.h>`, `"generated/Generated.h"`, one declaration | `ling/Source/Help.h:1-5` |
| A carried file: a `## binary` row in `project-info.md` and a `@string` row in `cast/files.md` | `PLAN-tit-13-setup.md` Step 2, items 4 and 5 |
| `TITView::resized` sets `menu.setBounds (content);` and ends with `setView()`; the mode event calls `setView()` | `PLAN-tit-10-views.md` Step 16, items 1 and 3; `PLAN-tit-12-protocols.md` Step 11, item 6 |
| `setView` sets `menu.setVisible (isLarge and (isMenuMode() or isSubmenuMode() or mode == map::EditorMode::startup));` | `PLAN-tit-13-setup.md` Step 9, item 6 |
| `isMenuMode()`: the operation menus, `menuNotRepo` to `menuPinned` | `PLAN-tit-11-history.md` Step 11, item 2 |
| The Config and Preferences modes: `map::EditorMode::menuConfig`, `menuPreferences` | `PLAN-tit-13-setup.md` Step 1 |
| `listHeightDivisor`: the family of a named layout divisor | `PLAN-tit-11-history.md` Step 11 |

## Closed by a read

1. **The help banner is braille only.** `titc` off a terminal prints the guide (`PLAN-tit-7-model.md:147`), so the output of `printHelp` can be a file. An inline image is an escape sequence, and braille is text. The pixel choice of Step 8 is for the window.
2. **The help banner has 10 rows.** The legacy reference grid is 40 by 10 cells (`braille_ref/main.go:21-23`). The logo is square, so it takes 20 columns by 10 rows, centred in the column width of the guide.
3. **The colour store cannot overflow.** Each cell colour gets an id. A full store returns the nearest held colour (JAM inventory).
4. **One logo image.** `getLogo()` builds the image at the drawable's size. `BrailleComponent` scales it into its dot image. `juce::ImageComponent` gives it to the inline lane, and the host scales it. So one image serves both components (ruling 23).
5. **No new colour id, no style row.** `jam::ImageComponent` and `jam::PreviewComponent` read no colour from the LookAndFeel (JAM inventory). The banner colours are the logo's own colours.
6. **The banner is a child of `TITView`.** Legacy draws it in the menu views only (legacy table). `TITView` sets the bounds and the visibility of every view, by mode.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: JAM, the braille dot table as cast lookup tables (ruling 50)
**Scope:** `jam/cast/graphics.md`, `jam/cast/spell.md`, `jam_graphics/fonts/jam_TerminalGlyph.h`. Generated: `jam/generated/jam_Graphics.h`.
**Action:** Delete first: remove the two local tables of `drawBraille` (`jam_TerminalGlyph.h:840-845`), and remove `brailleFirst` and its doxygen line from the private section (`:240-241`).
1. `jam/cast/graphics.md`: after the `## MouseButton` table and before `## files`, add two tables in the form of `## os2UnicodeRangeStarts` (`jam/cast/opentype.md:65-83`):

````markdown
## brailleDotColumns

```
@brief Braille patterns — the column of each dot in its cell.

Key N is bit N of the offset from U+2800, which is dot N + 1. Row N of this
table and of `brailleDotRows` together give the place of that dot in the cell.
Column 0 is the left column.
```

+-----+-------+
| key | value |
+=====+=======+
| 0   | 0     |
| 1   | 0     |
| 2   | 0     |
| 3   | 1     |
| 4   | 1     |
| 5   | 1     |
| 6   | 0     |
| 7   | 1     |
+-----+-------+

## brailleDotRows

```
@brief Braille patterns — the row of each dot in its cell.

Row N pairs with the same row of `brailleDotColumns`. Row 0 is the top row.
```

+-----+-------+
| key | value |
+=====+=======+
| 0   | 0     |
| 1   | 1     |
| 2   | 2     |
| 3   | 0     |
| 4   | 1     |
| 5   | 2     |
| 6   | 3     |
| 7   | 3     |
+-----+-------+
````

2. `jam/cast/spell.md`, `## index`: after the `@os2UnicodeRange` row (`:93`), add `| @brailleDot | jam::LookupTable<int, int, 8> |`, padded to the column widths. The capacity is a literal, as `@os2UnicodeRange` writes `169` (`:93`).
3. `jam/cast/spell.md`, `## output`: after the `@graphics:MouseButton` group (`:565-576`), add two groups in the form of the `@opentype:os2UnicodeRangeStarts` group (`:1625-1633`), with the target `@jam_Graphics`:
   - source rows `@graphics:brailleDotColumns`, `@graphics:brailleDotColumns:key`, `@graphics:brailleDotColumns:value`; `- type: @brailleDot`; `- name: brailleDotColumns`;
   - the same for `brailleDotRows`.
4. `jam_TerminalGlyph.h`: in the public section, after `getCoverage` (`:225`) and before `private:` (`:227`), add:

```cpp
    /** @brief Codepoint of the first braille pattern. */
    static constexpr uint32_t brailleFirst { 0x2800 };

    static constexpr int numBrailleColumns { 2 };
    static constexpr int numBrailleRows { 4 };
```

5. `drawBraille`: `const int cellW { w / numBrailleColumns };`, `const int cellH { h / numBrailleRows };`. The loop becomes `for (int i { 0 }; i < map::brailleDotColumns.capacity; ++i)`, and its body reads `map::brailleDotColumns[i]` and `map::brailleDotRows[i]`. Nothing else changes.

- The generated header `jam_Graphics.h` is included at `jam_graphics.h:41`, before `fonts/jam_TerminalGlyph.h` (`:78`). So `TerminalGlyph` and `jam_gui` read `map::brailleDotColumns` and `map::brailleDotRows`.
- The generator writes `inline constexpr jam::LookupTable<int, int, 8> brailleDotColumns { … }` in `namespace map`, as it writes `os2UnicodeRangeStarts` (`jam/generated/jam_OpenType.h:428`).
- `LookupTable::capacity` is `static constexpr`. `operator[]` is a direct lookup (`jam_core/utils/jam_LookupTable.h:84,157-160`).

**Validation:**
- The values are the values of the deleted local tables, in the same order, so `drawBraille` gives the same coverage bytes.
- Key `i` of the two tables is the column and the row of dot `i + 1`, and bit `i` of the pattern (Legacy faults, item 1).
- No C++ file holds the dot table. `numBrailleColumns` and `numBrailleRows` are the only names of the dot grid.
- A second `cast` run of JAM gives no change.
- New names and their families: `brailleDotColumns`, `brailleDotRows` (`os2UnicodeRangeStarts`, `os2UnicodeRangeEnds`; the local `dotCol`, `dotRow`, written in full, NAMES.md Rule 4, "Clarity over brevity"); `@brailleDot` (`@os2UnicodeRange`); `numBrailleColumns`, `numBrailleRows` (`numBoxDrawings`, `jam_TerminalGlyph.h:247`).

### Step 2: JAM, `BrailleComponent`
**Scope:** `jam_gui/widgets/jam_BrailleComponent.h`, `jam_gui/widgets/jam_BrailleComponent.cpp` (new), `jam_gui/jam_gui.h`, `jam_gui/jam_gui.cpp`.
**Action:**
1. The header:

```cpp
#pragma once

namespace jam
{

class BrailleComponent : public juce::Component
{
public:
    BrailleComponent();

    void setImage (const juce::Image& newImage);

    void paint (juce::Graphics& g) override;
    void resized() override;

private:
    void setDots();
    uint32_t getPattern (int column, int row) const;
    juce::Colour getColour (int column, int row) const;
    juce::Colour getDot (int column, int row, int dot) const;

    static int getChannelSum (juce::Colour colour);

    static constexpr int filledChannelSum { 50 };

    juce::Image image;
    juce::Image dots;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (BrailleComponent)
};

} // namespace jam
```

2. The source:

```cpp
namespace jam
{

BrailleComponent::BrailleComponent()
{
    setInterceptsMouseClicks (false, false);
}

void BrailleComponent::setImage (const juce::Image& newImage)
{
    image = newImage;
    setDots();
    repaint();
}

void BrailleComponent::resized()
{
    setDots();
}

void BrailleComponent::setDots()
{
    dots = {};

    if (image.isValid() and not getLocalBounds().isEmpty())
    {
        dots = juce::Image { juce::Image::RGB,
                             getWidth() * TerminalGlyph::numBrailleColumns,
                             getHeight() * TerminalGlyph::numBrailleRows,
                             true,
                             juce::SoftwareImageType() };

        juce::Graphics g { dots };
        g.drawImage (image, dots.getBounds().toFloat(), juce::RectanglePlacement::centred);
    }
}

void BrailleComponent::paint (juce::Graphics& g)
{
    if (dots.isValid())
    {
        for (int row { 0 }; row < getHeight(); ++row)
        {
            for (int column { 0 }; column < getWidth(); ++column)
            {
                if (const auto pattern { getPattern (column, row) }; pattern != 0)
                {
                    g.setColour (getColour (column, row));
                    g.drawText (juce::String::charToString (static_cast<juce::juce_wchar> (TerminalGlyph::brailleFirst + pattern)),
                                juce::Rectangle<int> { column, row, 1, 1 },
                                juce::Justification::centred,
                                false);
                }
            }
        }
    }
}

uint32_t BrailleComponent::getPattern (int column, int row) const
{
    uint32_t pattern { 0 };

    for (int dot { 0 }; dot < map::brailleDotColumns.capacity; ++dot)
        if (getChannelSum (getDot (column, row, dot)) > filledChannelSum)
            pattern |= 1u << dot;

    return pattern;
}

juce::Colour BrailleComponent::getColour (int column, int row) const
{
    juce::Colour brightest;

    for (int dot { 0 }; dot < map::brailleDotColumns.capacity; ++dot)
    {
        const auto colour { getDot (column, row, dot) };

        if (getChannelSum (colour) > getChannelSum (brightest))
            brightest = colour;
    }

    return brightest;
}

juce::Colour BrailleComponent::getDot (int column, int row, int dot) const
{
    return dots.getPixelAt (column * TerminalGlyph::numBrailleColumns + map::brailleDotColumns[dot],
                            row * TerminalGlyph::numBrailleRows + map::brailleDotRows[dot]);
}

int BrailleComponent::getChannelSum (juce::Colour colour)
{
    return colour.getRed() + colour.getGreen() + colour.getBlue();
}

} // namespace jam
```

3. `jam_gui.h`: after `#include "widgets/jam_PreviewComponent.h"` (`:136`), add `#include "widgets/jam_BrailleComponent.h"`.
4. `jam_gui.cpp`: after `#include "widgets/jam_PreviewComponent.cpp"` (`:32`), add `#include "widgets/jam_BrailleComponent.cpp"`.

- One unit is one cell (`ling/Source/style.css:6-9`), so the dot image has two pixels across and four down for each unit.
- The dot image is the materialisation of the image at the bounds. `setImage` and `resized` replace it whole. `paint` reads it and builds nothing.
- The canvas is black and the dot test is the legacy test (legacy table). The colour of a cell is its brightest dot. A cell with no filled dot paints nothing, so the background of the parent shows.
- `dots = {}` makes the image invalid when no image or no bounds exist. `paint` then paints nothing. This is a result of the state, not a guard: the component has nothing to show.

**Validation:**
- The class is in the `<What>Component` family of `ImageComponent` and `PreviewComponent` (ruling 25). `setImage (const juce::Image&)` has the signature of `juce::ImageComponent::setImage` (`juce_ImageComponent.h:59`) and `jam::ImageComponent::setImage` (`jam_ImageComponent.h:40`).
- Each function is 30 lines or fewer. No bail-out guard. No magic number: `filledChannelSum` is the legacy threshold, the dot table is `map::brailleDotColumns` and `map::brailleDotRows` (Step 1), and the grid size is `TerminalGlyph::numBrailleColumns` and `numBrailleRows`.
- No member is a copy of a Model value. The component holds its image as `juce::ImageComponent` does (`juce_ImageComponent.h:83-84`).
- New names and their families: `BrailleComponent` (ruling 25); `setDots`, `dots` (`setImage`, `image`); `getPattern` (`pattern`, `jam_TerminalGlyph.h:834`); `getColour`, `getDot`; `getChannelSum`, `filledChannelSum` (the legacy sum of R, G, and B, `braille.go:78`; JUCE `Colour::getBrightness` is the HSB brightness, another measure). Local names: `g`, `row`, `column`, `pattern`, `dot`, `brightest`, `colour`.

### Step 3: TIT, the logo file
**Scope:** `tit/cast/files.md`, `tit/project-info.md`.
**Action:**
1. `cast/files.md`, table `## files`: after the last row of Part 13 Step 2, add the `@string` row `titcLogo`, value `` `tit_logo.svg` ``, description `The TIT logo.`
2. `project-info.md`, table `## binary`: after the `gitIgnore` row of Part 13 Step 2, add `| logo | Assets/tit_logo.svg | TIT logo, embedded as BinaryData |`, in the form of that row. Keep the `+---` separator.

**Validation:**
- The rows follow the rows of `titcUser` and `user` (Part 13 Step 2, items 4 and 5).
- `cast cast/spell.md --debug` puts `tit_logo.svg` in BinaryData and `files::titcLogo` in `Files.h`.

### Step 4: TIT, `getLogo`
**Scope:** `tit/Source/Logo.h`, `tit/Source/Logo.cpp` (new).
**Action:**
1. `Logo.h`, in the form of `ling/Source/Help.h:1-5`:

```cpp
#pragma once
#include <JuceHeader.h>
#include "generated/Generated.h"

juce::Image getLogo();
```

2. `Logo.cpp`:

```cpp
#include "Logo.h"

juce::Image getLogo()
{
    const auto drawable { jam::Svg::getDrawable (BinaryData::getString (files::titcLogo).toRawUTF8()) };
    const auto bounds { drawable->getDrawableBounds().getSmallestIntegerContainer() };
    juce::Image logo { juce::Image::ARGB, bounds.getWidth(), bounds.getHeight(), true, juce::SoftwareImageType() };

    juce::Graphics g { logo };
    drawable->drawWithin (g, logo.getBounds().toFloat(), juce::RectanglePlacement::centred, 1.0f);

    return logo;
}
```

- The image is transparent outside the paths. `BrailleComponent` draws it on black. The inline lane sends its PNG with the alpha.
- `getDrawableBounds` gives the bounds of the paths, so the image holds the logo with no margin.

**Validation:**
- The function returns a complete image. It stores nothing (NAMES.md Rule 1, `get`).
- New names and their families: `getLogo` (`getDefaultGit`, Part 8; the file word `titcLogo`); `Logo.h`, `Logo.cpp` (`Help.h`, `Help.cpp`). Local names: `drawable`, `bounds`, `logo`, `g`.

### Step 5: TIT, the banner of `titc --help`
**Scope:** `tit/Source/Help.cpp`.
**Action:** After `#include "Help.h"`, add `#include "Logo.h"`. Add `static constexpr int bannerRows { 10 };` before `printHelp`. Make the body of `printHelp` this:

```cpp
void printHelp (std::string_view helpText)
{
    jam::SharedInstance<jam::SharedDocuments> documents { std::in_place };
    jam::Typeface typeface;

    juce::MemoryOutputStream frame;
    jam::TerminalBuffer buffer;
    const jam::GlyphArrangement glyphArrangement;
    jam::StyleManager styleManager { { BinaryData::fetcher }, jam::Strings { files::styleSheet } };
    jam::StyleMarkdown styleMarkdown;

    const auto columnWidth { styleMarkdown.getColumnWidth() };

    jam::BrailleComponent banner;
    banner.setBounds (0, 0, columnWidth, bannerRows);
    banner.setImage (getLogo());

    jam::TextEditor component;
    component.setLookAndFeel (&styleMarkdown);
    component.setBounds (0, 0, columnWidth, 0);
    component.setDocument (std::move (*jam::MarkdownDocument::parse (helpText, files::titcHelp)));

    const auto rows { bannerRows + component.getViewedComponent()->getHeight() };

    {
        buffer.prepare (columnWidth, rows, rows);

        jam::TerminalGraphicsContext context { buffer, *jam::Colours::getInstance() };
        juce::Graphics g { context };
        banner.paintEntireComponent (g, false);
        g.setOrigin (0, bannerRows);
        component.getViewedComponent()->paintEntireComponent (g, false);
    }

    buffer.process();
    buffer.writeToStream (frame, styleMarkdown);
    frame << static_cast<char> (Chars::newline);

    fwrite (frame.getData(), sizeof (char), frame.getDataSize(), stdout);
}
```

- The buffer keeps the default constructor, so the help frame has no inline lane (Closed by a read, item 1).
- The banner is painted first at the origin. `setOrigin` moves the guide down by `bannerRows` cells.

**Validation:**
- The body differs from Part 1 Step 9 and Part 3 Step 3, item 8, only in the banner lines and `rows`. It is 30 lines or fewer.
- `bannerRows` is the legacy reference grid (Closed by a read, item 2).
- New name and its family: `bannerRows` (`searchRows`, `titleRows`, `buttonRows`, Parts 10 and 11); `banner` (the view word of tracker row 15).

### Step 6: TIT, the banner beside the menu
**Scope:** `tit/Source/TITView.h`, `tit/Source/TITView.cpp`.
**Action:**
1. `TITView.h`: add `#include "Logo.h"` after the last include. After the member `MenuView menu;`, add `jam::BrailleComponent banner;`. After `isSubmenuMode` (Part 13), add `bool isBannerMode() const;`. After `shortHashLength` (Part 11), add `static constexpr int bannerWidthDivisor { 2 };`.
2. The constructor: after `addAndMakeVisible (menu);`, add `addChildComponent (banner);` and `banner.setImage (getLogo());`.
3. `resized()` (Parts 10 to 12): replace `menu.setBounds (content);` with:

```cpp
    menu.setBounds (isBannerMode() ? content.withWidth (content.getWidth() / bannerWidthDivisor) : content);
    banner.setBounds (content.withTrimmedLeft (content.getWidth() / bannerWidthDivisor));
```

4. `setView` (Parts 10 to 13): after the `menu.setVisible` line of Part 13, add `banner.setVisible (isLarge and isBannerMode());`.
5. `registerEvents`: in the `Id::mode` entry of Part 12 Step 11, item 6, replace the last statement `setView();` with `resized();`. `resized()` sets the bounds of this step and ends with `setView()` (Part 10 Step 16, item 1).
6. Add:

```cpp
bool TITView::isBannerMode() const
{
    const auto mode { static_cast<int> (state.getProperty (Id::mode)) };

    return isMenuMode() or mode == map::EditorMode::menuConfig or mode == map::EditorMode::menuPreferences;
}
```

- The mode event calls `resized()` (item 5), so a mode change moves the menu and shows or hides the banner.
- The left half is `width / 2` and the banner takes the rest, as legacy (`legacy/ui/menu.go:191-193`).

**Validation:**
- The banner shows in the operation menus, Config, and Preferences, and nowhere else (legacy table).
- No member holds state. `banner` is a view.
- New names and their families: `banner` (Step 5); `isBannerMode` (`isMenuMode`, `isHistoryMode`, `isSubmenuMode`, Parts 11 and 13); `bannerWidthDivisor` (`listHeightDivisor` of `HistoryView`, Part 11 Step 10).

### Step 7: Build and check
**Scope:** the JAM change and the `tit/` project on macOS.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:**
1. From `tit/`, run `cast cast/spell.md --debug`.
2. In END, run `titc --help`. Run `titc --help > ~/titc-help.txt`, then read the file.
3. In END, in a repository fixture of Part 8 Step 9, run `titc`. Open Config and Preferences with the keys of Part 13 Step 14, fixture 6. Open History. Open the location menu of `init`. Resize the pane below and above 70 by 20 cells. Press Ctrl+C.
4. Do item 3 in kitty.
5. Build EVE, LING, and END, the other users of `jam_graphics` and `jam_gui`.

**Validation:**
- Item 2: the logo shows in the first 10 rows, 20 columns wide, centred, in its colours, then the guide. The file holds braille characters and SGR colour sequences, and no OSC 1337 sequence.
- Item 3: the banner fills the right half beside the main menu, Config, and Preferences. History and the location menu show no banner. A resize rebuilds the banner at the new size.
- The dots of the logo match its shape: a braille character with only dot 7 is at the lower left of its cell (Legacy faults, item 1). The holes of the evenodd paths are open (item 2).
- Item 4 gives the same banner as item 3.
- Item 5: each build ends with exit code 0. A braille glyph in END and EVE is drawn as before (Step 1, Validation).

### Step 8: ARCHITECT's choice of the window banner (ruling 23)
**Scope:** `tit/Source/TITView.h`.
**Owner:** ARCHITECT decides. COUNSELOR makes the change.
**Action:**
1. In `TITView.h`, change the member `jam::BrailleComponent banner;` to `juce::ImageComponent banner;`. Nothing else changes: `setImage (const juce::Image&)` is the same call (`juce_ImageComponent.h:59`), and the placement is centred (`juce_ImageComponent.cpp:38-41`).
2. Build, and do Step 7, item 3, in END and in kitty. The image goes through the inline lane of the peer (Part 5 Step 1).
3. ARCHITECT compares Step 7 and this step, and names the banner.
4. If ARCHITECT names braille, change the member back to `jam::BrailleComponent banner;`. If ARCHITECT names pixels, keep `juce::ImageComponent banner;`.
5. Record the choice as a ruling in the tracker canon, in ARCHITECT's words.

**Validation:**
- The member type is the type that ARCHITECT named. No other line of `TITView` differs between the two types.
- `jam::BrailleComponent` stays in JAM in both cases, and `titc --help` keeps the braille banner (Closed by a read, item 1).

## BLESSED Alignment

- **B — Bound.** `TITView` owns the banner. Each component owns its image. `printHelp` owns its banner for one frame.
- **L — Lean.** One component in JAM, one function and one row in TIT. The braille table exists once. No rasterizer is written: JUCE draws the logo.
- **E — Explicit.** Each constant has a name. The dot test and the colour choice are the legacy rules, named.
- **S — SSOT.** One dot table, in the cast manifest of JAM (ruling 50). One logo file, carried once. One function builds the logo image.
- **S — Stateless.** No Model row is added. The banner reads no state; `TITView` sets its bounds and its visibility from the `mode` row.
- **E — Encapsulation.** `BrailleComponent` knows no terminal type. It paints text and colours through `juce::Graphics`, so it paints in any JUCE context. `jam_gui` reads `jam_graphics`, never the reverse.
- **D — Deterministic.** The same image and the same bounds give the same cells.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. Ruling 23 puts the choice of the window banner at execution; Step 8 holds it.

Facts that the run measures first:
- `juce::Image` with `juce::SoftwareImageType` in the console process. `jam::PreviewComponent` makes the same images (inventory). Step 7 runs it.
- kitty and the OSC 1337 sequence. Step 8, item 2, shows what kitty does with it.
