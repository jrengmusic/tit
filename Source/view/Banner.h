#pragma once
#include <JuceHeader.h>

namespace tit
{

// ============================================================================
// Banner
// ============================================================================
//
// Static startup banner — renders the TIT SVG logo as a braille character grid.
// Ported from ___legacy___/internal/ui/layout.go RenderBannerDynamic().
//
// Does NOT observe VT — banner is static content per SPEC §14.
// BLESSED S: no persistent state beyond the cached BrailleGrid.
// BLESSED E: imports no Source/git/ headers.

class Banner : public jam::tui::Component
{
public:
    Banner();
    ~Banner() override = default;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint   (jam::tui::Graphics& g) override;
    void resized ()                        override;

private:
    // Cached braille render; rebuilt in resized().
    jam::braille::BrailleGrid cachedGrid;
    int                         cachedCols { 0 };
    int                         cachedRows { 0 };

    // Builds cachedGrid for the given terminal dimensions.
    void rebuildGrid (int cols, int rows);

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (Banner)
};

} // namespace tit
