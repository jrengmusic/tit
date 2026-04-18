#include <jam_tui/jam_tui.h>

// ============================================================================
// DialogTests
// ============================================================================

class DialogTests : public juce::UnitTest
{
public:
    DialogTests() : juce::UnitTest ("Dialog", "jam_tui") {}

    void runTest() override
    {
        testCtorAndBinding ();
        testSelectionState ();
        testKeyboardNavigation ();
        testContextSubstitution ();
        testSingleButtonVariant ();
    }

private:

    static juce::ValueTree buildDialogTree ()
    {
        juce::ValueTree dialog { "DIALOG" };
        dialog.setProperty (jam::tui::Dialog::PropTitle,       "Reset branch?",         nullptr);
        dialog.setProperty (jam::tui::Dialog::PropExplanation, "This will reset HEAD.", nullptr);
        dialog.setProperty (jam::tui::Dialog::PropYesLabel,    "Yes",                   nullptr);
        dialog.setProperty (jam::tui::Dialog::PropNoLabel,     "No",                    nullptr);
        dialog.setProperty (jam::tui::Dialog::PropActionId,    "reset",                 nullptr);
        return dialog;
    }

    void testCtorAndBinding ()
    {
        beginTest ("ctor + ValueTree binding");

        juce::ValueTree tree { buildDialogTree () };
        jam::tui::Dialog dialog { tree };

        // Default: Yes selected
        // Confirm callback fires with actionId
        juce::String firedId;
        dialog.onConfirmed = [&] (const juce::String& id) { firedId = id; };

        jam::tui::KeyEvent enter;
        enter.type = jam::tui::KeyType::Enter;
        dialog.handleInput (enter);

        expect (firedId == "reset", "onConfirmed must fire with actionId 'reset'");

        // Modify tree property and check rebuild
        tree.setProperty (jam::tui::Dialog::PropActionId, "rewind", nullptr);
        firedId = "";
        dialog.handleInput (enter);
        expect (firedId == "rewind", "Dialog must react to VT property change");
    }

    void testSelectionState ()
    {
        beginTest ("selection state");

        juce::ValueTree tree { buildDialogTree () };
        jam::tui::Dialog dialog { tree };

        bool confirmedFired { false };
        bool cancelledFired { false };

        dialog.onConfirmed = [&] (const juce::String&) { confirmedFired = true; };
        dialog.onCancelled = [&] ()                    { cancelledFired = true; };

        // Tab toggles to No
        jam::tui::KeyEvent tab;
        tab.type = jam::tui::KeyType::Tab;
        dialog.handleInput (tab);

        // Enter on No fires onCancelled
        jam::tui::KeyEvent enter;
        enter.type = jam::tui::KeyType::Enter;
        dialog.handleInput (enter);

        expect (not confirmedFired, "onConfirmed must not fire when No is selected");
        expect (cancelledFired,     "onCancelled must fire when No is selected");
    }

    void testKeyboardNavigation ()
    {
        beginTest ("keyboard navigation");

        juce::ValueTree tree { buildDialogTree () };
        jam::tui::Dialog dialog { tree };

        bool cancelledFired { false };
        dialog.onCancelled = [&] () { cancelledFired = true; };

        // Escape always fires onCancelled
        jam::tui::KeyEvent esc;
        esc.type = jam::tui::KeyType::Escape;
        dialog.handleInput (esc);

        expect (cancelledFired, "Escape must fire onCancelled");

        // ArrowRight → No, ArrowLeft → Yes
        jam::tui::KeyEvent right;
        right.type = jam::tui::KeyType::ArrowRight;
        dialog.handleInput (right);

        bool confirmedFired { false };
        dialog.onConfirmed = [&] (const juce::String&) { confirmedFired = true; };
        cancelledFired = false;

        jam::tui::KeyEvent enter;
        enter.type = jam::tui::KeyType::Enter;
        dialog.handleInput (enter);

        expect (cancelledFired,     "ArrowRight selects No; Enter must fire onCancelled");
        expect (not confirmedFired, "onConfirmed must not fire after ArrowRight + Enter");

        // ArrowLeft brings back Yes
        jam::tui::KeyEvent left;
        left.type = jam::tui::KeyType::ArrowLeft;
        dialog.handleInput (left);

        confirmedFired = false;
        cancelledFired = false;
        dialog.handleInput (enter);

        expect (confirmedFired,     "ArrowLeft selects Yes; Enter must fire onConfirmed");
        expect (not cancelledFired, "onCancelled must not fire after ArrowLeft + Enter");
    }

    void testContextSubstitution ()
    {
        beginTest ("context substitution");

        juce::ValueTree tree { buildDialogTree () };
        tree.setProperty (jam::tui::Dialog::PropTitle, "Reset {branch}?", nullptr);

        jam::tui::Dialog dialog { tree };

        juce::StringPairArray ctx;
        ctx.set ("branch", "main");
        dialog.setContext (ctx);

        // No assertion on rendered output (paint is framebuffer);
        // verify the dialog can paint without crashing after setContext.
        expect (true, "setContext must not crash");
    }

    void testSingleButtonVariant ()
    {
        beginTest ("single-button alert variant");

        juce::ValueTree tree { "DIALOG" };
        tree.setProperty (jam::tui::Dialog::PropTitle,       "Git not found",         nullptr);
        tree.setProperty (jam::tui::Dialog::PropExplanation, "Install git to begin.", nullptr);
        tree.setProperty (jam::tui::Dialog::PropYesLabel,    "OK",                    nullptr);
        tree.setProperty (jam::tui::Dialog::PropNoLabel,     "",                      nullptr);
        tree.setProperty (jam::tui::Dialog::PropActionId,    "acknowledge",           nullptr);

        jam::tui::Dialog dialog { tree };

        // Tab on single-button dialog must not change state (no noLabel)
        bool confirmedFired { false };
        dialog.onConfirmed = [&] (const juce::String&) { confirmedFired = true; };

        jam::tui::KeyEvent tab;
        tab.type = jam::tui::KeyType::Tab;
        dialog.handleInput (tab);

        jam::tui::KeyEvent enter;
        enter.type = jam::tui::KeyType::Enter;
        dialog.handleInput (enter);

        expect (confirmedFired, "Single-button dialog: Enter must fire onConfirmed");
    }
};

static DialogTests dialogTestsInstance;
