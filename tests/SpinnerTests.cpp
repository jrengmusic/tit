#include <jam_tui/jam_tui.h>

// ============================================================================
// SpinnerTests
// ============================================================================

class SpinnerTests : public juce::UnitTest
{
public:
    SpinnerTests() : juce::UnitTest ("Spinner", "jam_tui") {}

    void runTest() override
    {
        testFrameSet ();
        testStartStop ();
        testTimerTickAdvancesFrame ();
        testFrameWrap ();
    }

private:

    void testFrameSet ()
    {
        beginTest ("frame set — 10 frames verbatim from Go SpinnerFrames");

        expect (jam::tui::Spinner::frameCount () == 10,
                "SpinnerFrames must have exactly 10 entries (Go source)");

        // Spot-check first and last frame are non-empty braille strings
        expect (jam::tui::Spinner::frameCharAt (0).isNotEmpty (),
                "Frame 0 must be non-empty");
        expect (jam::tui::Spinner::frameCharAt (9).isNotEmpty (),
                "Frame 9 must be non-empty");

        // All 10 frames must be distinct
        juce::StringArray seen;
        bool allDistinct { true };

        for (int i { 0 }; i < jam::tui::Spinner::frameCount (); ++i)
        {
            const juce::String frame { jam::tui::Spinner::frameCharAt (i) };

            if (seen.contains (frame))
                allDistinct = false;
            else
                seen.add (frame);
        }

        expect (allDistinct, "All 10 spinner frames must be distinct characters");
    }

    void testStartStop ()
    {
        beginTest ("start/stop toggle");

        jam::tui::Spinner spinner;

        expect (not spinner.isRunning (), "Spinner must not be running on construction");

        spinner.start ();
        expect (spinner.isRunning (), "Spinner must be running after start()");

        spinner.stop ();
        expect (not spinner.isRunning (), "Spinner must not be running after stop()");
    }

    void testTimerTickAdvancesFrame ()
    {
        beginTest ("timer tick advances frame index");

        jam::tui::Spinner spinner;

        expect (spinner.getFrameIndex () == 0, "Initial frame index must be 0");

        // Simulate timer ticks by calling timerCallback directly
        spinner.timerCallback ();
        expect (spinner.getFrameIndex () == 1, "Frame index must be 1 after one tick");

        spinner.timerCallback ();
        expect (spinner.getFrameIndex () == 2, "Frame index must be 2 after two ticks");

        // getCurrentFrameChar must reflect updated index
        const juce::String expected { jam::tui::Spinner::frameCharAt (2) };
        expect (spinner.getCurrentFrameChar () == expected,
                "getCurrentFrameChar must return frame at current index");
    }

    void testFrameWrap ()
    {
        beginTest ("frame index wraps at end of cycle");

        jam::tui::Spinner spinner;

        const int total { jam::tui::Spinner::frameCount () };

        for (int i { 0 }; i < total; ++i)
            spinner.timerCallback ();

        expect (spinner.getFrameIndex () == 0,
                "Frame index must wrap to 0 after full cycle");

        // frameCharAt must also wrap
        const juce::String charAt10 { jam::tui::Spinner::frameCharAt (total) };
        const juce::String charAt0  { jam::tui::Spinner::frameCharAt (0) };
        expect (charAt10 == charAt0, "frameCharAt must wrap for out-of-range index");
    }
};

static SpinnerTests spinnerTestsInstance;
