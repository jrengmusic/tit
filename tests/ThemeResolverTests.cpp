#include <jam_tui/jam_tui.h>

// ============================================================================
// ThemeResolverTests
// ============================================================================

class ThemeResolverTests : public juce::UnitTest
{
public:
    ThemeResolverTests() : juce::UnitTest ("ThemeResolver", "jam_tui") {}

    void runTest() override
    {
        testIsValid ();
        testGetColourFromHexString ();
        testGetColourFallbackOnMissingKey ();
        testGetString ();
        testGetStringFallbackOnMissingKey ();
        testGetInt ();
        testGetIntFallbackOnMissingKey ();
    }

private:

    static juce::ValueTree buildThemeTree ()
    {
        juce::ValueTree theme { "THEME" };
        theme.setProperty ("menuSelectedBg",    "#ff005f5f", nullptr);
        theme.setProperty ("contentTextColor",  "#ffcccccc", nullptr);
        theme.setProperty ("spinnerColor",      "#ff44ffcc", nullptr);
        theme.setProperty ("labelText",         "Hello",     nullptr);
        theme.setProperty ("timeoutMs",         250,         nullptr);
        return theme;
    }

    void testIsValid ()
    {
        beginTest ("isValid reflects ValueTree validity");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        expect (resolver.isValid (), "ThemeResolver must be valid when constructed with valid VT");
    }

    void testGetColourFromHexString ()
    {
        beginTest ("getColour parses #aarrggbb hex string");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const juce::Colour fallback { juce::Colours::black };
        const juce::Colour result   { resolver.getColour (juce::Identifier { "menuSelectedBg" },
                                                          fallback) };

        // #ff005f5f = alpha=ff r=00 g=5f b=5f
        expect (result.getRed ()   == 0x00, "Red channel must be 0x00");
        expect (result.getGreen () == 0x5f, "Green channel must be 0x5f");
        expect (result.getBlue ()  == 0x5f, "Blue channel must be 0x5f");
        expect (result.getAlpha () == 0xff, "Alpha channel must be 0xff");
    }

    void testGetColourFallbackOnMissingKey ()
    {
        beginTest ("getColour returns fallback for missing key");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const juce::Colour fallback { juce::Colours::red };
        const juce::Colour result   { resolver.getColour (juce::Identifier { "nonExistentKey" },
                                                          fallback) };

        expect (result == fallback,
                "getColour must return fallback for missing key");
    }

    void testGetString ()
    {
        beginTest ("getString returns stored string value");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const juce::String result { resolver.getString (juce::Identifier { "labelText" }, "") };
        expect (result == "Hello", "getString must return the stored string value");
    }

    void testGetStringFallbackOnMissingKey ()
    {
        beginTest ("getString returns fallback for missing key");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const juce::String fallback { "default" };
        const juce::String result   { resolver.getString (juce::Identifier { "noSuchKey" },
                                                          fallback) };

        expect (result == fallback,
                "getString must return fallback for missing key");
    }

    void testGetInt ()
    {
        beginTest ("getInt returns stored int value");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const int result { resolver.getInt (juce::Identifier { "timeoutMs" }, 0) };
        expect (result == 250, "getInt must return the stored integer value");
    }

    void testGetIntFallbackOnMissingKey ()
    {
        beginTest ("getInt returns fallback for missing key");

        juce::ValueTree theme { buildThemeTree () };
        jam::tui::ThemeResolver resolver { theme };

        const int result { resolver.getInt (juce::Identifier { "noSuchKey" }, 42) };
        expect (result == 42, "getInt must return fallback for missing key");
    }
};

static ThemeResolverTests themeResolverTestsInstance;
