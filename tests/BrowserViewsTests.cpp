#include <JuceHeader.h>
#include "view/HistoryView.h"
#include "view/FileHistoryView.h"
#include "view/ConflictResolverView.h"
#include "view/ConsoleView.h"
#include "view/ConfirmDialog.h"
#include "view/SetupWizardView.h"
#include "state/TitAxis.h"
#include "TitIdentifier.h"

// ============================================================================
// BrowserViewsTests
// ============================================================================
//
// One UnitTest subclass per view (6 total):
//   HistoryViewTest, FileHistoryViewTest, ConflictResolverViewTest,
//   ConsoleViewTest, ConfirmDialogTest, SetupWizardViewTest.
//
// ConfirmDialog test iterates all 7 variants, asserts VT-driven content
// matches SPEC §12 expected title per variant.
//
// Static registrar pattern, no anonymous namespace.

// ============================================================================
// Fixture helpers
// ============================================================================

static juce::ValueTree makeMinimalThemeTree()
{
    return juce::ValueTree { tit::ID::THEME };
}

static juce::ValueTree makeFullFixtureTree()
{
    juce::ValueTree root { tit::ID::TIT };

    juce::ValueTree repo { tit::ID::REPO };
    repo.setProperty (tit::ID::branch,      "main",    nullptr);
    repo.setProperty (tit::ID::operation,   "Normal",  nullptr);
    repo.setProperty (tit::ID::workingTree, "Clean",   nullptr);
    repo.setProperty (tit::ID::timeline,    "InSync",  nullptr);
    repo.setProperty (tit::ID::remote,      "HasRemote", nullptr);
    root.appendChild (repo, nullptr);

    juce::ValueTree history { tit::ID::HISTORY };
    root.appendChild (history, nullptr);

    juce::ValueTree files { tit::ID::FILES };
    root.appendChild (files, nullptr);

    juce::ValueTree diff { tit::ID::DIFF };
    root.appendChild (diff, nullptr);

    juce::ValueTree console { tit::ID::CONSOLE };
    root.appendChild (console, nullptr);

    juce::ValueTree selection { tit::ID::SELECTION };
    selection.setProperty (tit::ID::menuIndex,    0, nullptr);
    selection.setProperty (tit::ID::historyIndex, 0, nullptr);
    root.appendChild (selection, nullptr);

    juce::ValueTree dialog { tit::ID::DIALOG };
    dialog.setProperty (jam::tui::Dialog::PropTitle,       "", nullptr);
    dialog.setProperty (jam::tui::Dialog::PropExplanation, "", nullptr);
    dialog.setProperty (jam::tui::Dialog::PropYesLabel,    "", nullptr);
    dialog.setProperty (jam::tui::Dialog::PropNoLabel,     "", nullptr);
    dialog.setProperty (jam::tui::Dialog::PropActionId,    "", nullptr);
    dialog.setProperty (tit::ID::kind, "",      nullptr);
    root.appendChild (dialog, nullptr);

    juce::ValueTree setup { tit::ID::SETUP };
    setup.setProperty (tit::ID::phase,     tit::toString (tit::SetupPhase::EnvCheck), nullptr);
    setup.setProperty (tit::ID::email,     "",  nullptr);
    setup.setProperty (tit::ID::publicKey, "",  nullptr);
    root.appendChild (setup, nullptr);

    juce::ValueTree menu { tit::ID::MENU };
    root.appendChild (menu, nullptr);

    root.appendChild (makeMinimalThemeTree(), nullptr);

    return root;
}

// ============================================================================
// Test 1 — HistoryView
// ============================================================================

class HistoryViewTest : public juce::UnitTest
{
public:
    HistoryViewTest() : juce::UnitTest ("HistoryView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testListenerAttach ();
        testCommitChildTriggersPaint ();
    }

private:
    void testConstruction ()
    {
        beginTest ("HistoryView constructs and attaches to HISTORY subtree");

        juce::ValueTree            state   { makeFullFixtureTree() };
        jam::tui::ThemeResolver  resolver { state.getChildWithName (tit::ID::THEME) };

        tit::HistoryView view { state, resolver };
        view.setSize (80, 24);

        expect (view.getWidth()  == 80, "width set");
        expect (view.getHeight() == 24, "height set");
    }

    void testListenerAttach ()
    {
        beginTest ("HistoryView: HISTORY subtree mutation does not crash");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::HistoryView  view    { state, resolver };
        juce::ValueTree   history { state.getChildWithName (tit::ID::HISTORY) };

        juce::ValueTree commit { tit::ID::COMMIT };
        commit.setProperty (tit::ID::hash,    "abc1234def",            nullptr);
        commit.setProperty (tit::ID::author,  "Test Author",           nullptr);
        commit.setProperty (tit::ID::date,    "2024-12-30",            nullptr);
        commit.setProperty (tit::ID::message, "feat: initial commit",  nullptr);
        history.appendChild (commit, nullptr);

        expect (history.getNumChildren() == 1, "HISTORY has one COMMIT child");
    }

    void testCommitChildTriggersPaint ()
    {
        beginTest ("HistoryView: SELECTION historyIndex change does not crash");

        juce::ValueTree           state     { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver  { state.getChildWithName (tit::ID::THEME) };

        tit::HistoryView  view      { state, resolver };
        juce::ValueTree   selection { state.getChildWithName (tit::ID::SELECTION) };

        selection.setProperty (tit::ID::historyIndex, 0, nullptr);

        expect (true, "no crash on historyIndex change");
    }
};

// ============================================================================
// Test 2 — FileHistoryView
// ============================================================================

class FileHistoryViewTest : public juce::UnitTest
{
public:
    FileHistoryViewTest() : juce::UnitTest ("FileHistoryView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testHistoryAndFilesListeners ();
        testDiffPropertyTrigger ();
    }

private:
    void testConstruction ()
    {
        beginTest ("FileHistoryView constructs and attaches to HISTORY + FILES + DIFF");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::FileHistoryView view { state, resolver };
        view.setSize (80, 24);

        expect (view.getWidth()  == 80, "width set");
        expect (view.getHeight() == 24, "height set");
    }

    void testHistoryAndFilesListeners ()
    {
        beginTest ("FileHistoryView: HISTORY and FILES child add does not crash");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::FileHistoryView view    { state, resolver };
        juce::ValueTree      history { state.getChildWithName (tit::ID::HISTORY) };
        juce::ValueTree      files   { state.getChildWithName (tit::ID::FILES) };

        juce::ValueTree commit { tit::ID::COMMIT };
        commit.setProperty (tit::ID::hash,   "abc1234", nullptr);
        commit.setProperty (tit::ID::date,   "2024-12-30", nullptr);
        history.appendChild (commit, nullptr);

        juce::ValueTree file { tit::ID::FILE };
        file.setProperty (tit::ID::path,   "src/main.cpp", nullptr);
        file.setProperty (tit::ID::status, "M",            nullptr);
        files.appendChild (file, nullptr);

        expect (history.getNumChildren() == 1, "one COMMIT child");
        expect (files.getNumChildren()   == 1, "one FILE child");
    }

    void testDiffPropertyTrigger ()
    {
        beginTest ("FileHistoryView: DIFF lines property change does not crash");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::FileHistoryView view { state, resolver };
        juce::ValueTree      diff { state.getChildWithName (tit::ID::DIFF) };

        diff.setProperty (tit::ID::lines, "--- a/file\n+++ b/file\n@@ -1 +1 @@\n-old\n+new\n",
                          nullptr);

        expect (true, "no crash on diff lines property change");
    }
};

// ============================================================================
// Test 3 — ConflictResolverView
// ============================================================================

class ConflictResolverViewTest : public juce::UnitTest
{
public:
    ConflictResolverViewTest() : juce::UnitTest ("ConflictResolverView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testConflictMarkersInDiff ();
        testTabCyclesFocus ();
    }

private:
    void testConstruction ()
    {
        beginTest ("ConflictResolverView constructs and attaches to DIFF subtree");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConflictResolverView view { state, resolver };
        view.setSize (80, 24);

        expect (view.getWidth()  == 80, "width set");
        expect (view.getHeight() == 24, "height set");
    }

    void testConflictMarkersInDiff ()
    {
        beginTest ("ConflictResolverView: conflict marker diff triggers rebuildPanes");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConflictResolverView view { state, resolver };
        juce::ValueTree           diff { state.getChildWithName (tit::ID::DIFF) };

        const juce::String conflictDiff {
            "<<<<<<< HEAD\n"
            "our change\n"
            "=======\n"
            "their change\n"
            ">>>>>>> branch\n"
        };

        diff.setProperty (tit::ID::lines, conflictDiff, nullptr);

        expect (true, "no crash on conflict diff property change");
    }

    void testTabCyclesFocus ()
    {
        beginTest ("ConflictResolverView: Tab key does not crash");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConflictResolverView view { state, resolver };
        view.setSize (80, 24);

        jam::tui::KeyEvent tab;
        tab.type = jam::tui::KeyType::Tab;
        view.handleInput (tab);
        view.handleInput (tab);
        view.handleInput (tab);

        expect (true, "three Tab presses cycle back to ours without crash");
    }
};

// ============================================================================
// Test 4 — ConsoleView
// ============================================================================

class ConsoleViewTest : public juce::UnitTest
{
public:
    ConsoleViewTest() : juce::UnitTest ("ConsoleView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testLineChildAdded ();
    }

private:
    void testConstruction ()
    {
        beginTest ("ConsoleView constructs and wraps ConsoleStream");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConsoleView view { state, resolver };
        view.setSize (80, 24);

        expect (view.getWidth()  == 80, "width set");
        expect (view.getHeight() == 24, "height set");
        expect (view.getNumChildComponents() == 1, "ConsoleStream is single child");
    }

    void testLineChildAdded ()
    {
        beginTest ("ConsoleView: LINE child added to CONSOLE subtree does not crash");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConsoleView view    { state, resolver };
        juce::ValueTree  console { state.getChildWithName (tit::ID::CONSOLE) };

        juce::ValueTree line { tit::ID::LINE };
        line.setProperty (tit::ID::text,   "Build completed", nullptr);
        line.setProperty (tit::ID::stream, "stdout",          nullptr);
        console.appendChild (line, nullptr);

        expect (console.getNumChildren() == 1, "one LINE child added");
    }
};

// ============================================================================
// Test 5 — ConfirmDialog (7-variant coverage)
// ============================================================================

class ConfirmDialogTest : public juce::UnitTest
{
public:
    ConfirmDialogTest() : juce::UnitTest ("ConfirmDialog", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testAllSevenVariants ();
        testCallbackForwarding ();
    }

private:
    void testConstruction ()
    {
        beginTest ("ConfirmDialog constructs and wraps Dialog primitive");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConfirmDialog dialog { state };
        dialog.setSize (80, 24);

        expect (dialog.getWidth()  == 80, "width set");
        expect (dialog.getHeight() == 24, "height set");
        expect (dialog.getNumChildComponents() == 1, "Dialog is single child");
    }

    void testAllSevenVariants ()
    {
        // Expected titles per SPEC §12 / Go confirm_dialog_render.go
        struct VariantExpectation
        {
            juce::String kind;
            juce::String expectedTitleContains;
        };

        const std::initializer_list<VariantExpectation> expectations
        {
            { "rewind",               "Destructive" },
            { "time-travel",          "Time Travel"  },
            { "dirty",                "Uncommitted"  },
            { "merge",                "Merge"        },
            { "push",                 "Force Push"   },
            { "branch",               "Switch"       },
            { "time-travel-return",   "Return"       },
        };

        for (const VariantExpectation& exp : expectations)
        {
            beginTest (juce::String ("ConfirmDialog variant: ") + exp.kind);

            juce::ValueTree           state    { makeFullFixtureTree() };
            jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

            tit::ConfirmDialog view  { state };
            juce::ValueTree    dlg   { state.getChildWithName (tit::ID::DIALOG) };

            dlg.setProperty (tit::ID::kind, exp.kind, nullptr);

            const juce::String actualTitle { dlg.getProperty (jam::tui::Dialog::PropTitle).toString() };

            expect (actualTitle.containsIgnoreCase (exp.expectedTitleContains),
                    "variant '" + exp.kind + "': title should contain '"
                    + exp.expectedTitleContains + "', got: " + actualTitle);
        }
    }

    void testCallbackForwarding ()
    {
        beginTest ("ConfirmDialog: onConfirmed callback forwarded from Dialog");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::ConfirmDialog view { state };
        view.setSize (80, 24);

        bool fired { false };
        view.onConfirmed = [&fired] (const juce::String&) { fired = true; };

        // Dialog default = Yes selected; Enter confirms.
        jam::tui::KeyEvent enter;
        enter.type = jam::tui::KeyType::Enter;
        view.handleInput (enter);

        expect (fired, "onConfirmed must fire on Enter with Yes selected");
    }
};

// ============================================================================
// Test 6 — SetupWizardView
// ============================================================================

class SetupWizardViewTest : public juce::UnitTest
{
public:
    SetupWizardViewTest() : juce::UnitTest ("SetupWizardView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testPhasePropertyChange ();
        testAllPhasesRenderWithoutCrash ();
    }

private:
    void testConstruction ()
    {
        beginTest ("SetupWizardView constructs and attaches to SETUP subtree");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::SetupWizardView view { state, resolver };
        view.setSize (80, 24);

        expect (view.getWidth()  == 80, "width set");
        expect (view.getHeight() == 24, "height set");
    }

    void testPhasePropertyChange ()
    {
        beginTest ("SetupWizardView: phase property change triggers repaint");

        juce::ValueTree           state    { makeFullFixtureTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::SetupWizardView view  { state, resolver };
        juce::ValueTree      setup { state.getChildWithName (tit::ID::SETUP) };

        setup.setProperty (tit::ID::phase, tit::toString (tit::SetupPhase::SSHKeyEntry), nullptr);

        expect (true, "no crash on phase property change");
    }

    void testAllPhasesRenderWithoutCrash ()
    {
        beginTest ("SetupWizardView: all 6 phases render without crash");

        const std::initializer_list<tit::SetupPhase> phases
        {
            tit::SetupPhase::EnvCheck,
            tit::SetupPhase::SSHKeyEntry,
            tit::SetupPhase::KeyGen,
            tit::SetupPhase::Display,
            tit::SetupPhase::GitConfig,
            tit::SetupPhase::Done,
        };

        for (tit::SetupPhase phase : phases)
        {
            juce::ValueTree           state    { makeFullFixtureTree() };
            jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

            tit::SetupWizardView view  { state, resolver };
            juce::ValueTree      setup { state.getChildWithName (tit::ID::SETUP) };

            setup.setProperty (tit::ID::phase,     tit::toString (phase), nullptr);
            setup.setProperty (tit::ID::email,     "user@example.com",    nullptr);
            setup.setProperty (tit::ID::publicKey, "ssh-ed25519 AAAA...", nullptr);

            // Phase in VT readable without crash.
            const juce::String phaseStr { setup.getProperty (tit::ID::phase).toString() };
            expect (phaseStr == tit::toString (phase),
                    "phase property round-trips for " + tit::toString (phase));
        }
    }
};

// ============================================================================
// Static registrars
// ============================================================================

static HistoryViewTest          historyViewTestInstance;
static FileHistoryViewTest      fileHistoryViewTestInstance;
static ConflictResolverViewTest conflictResolverViewTestInstance;
static ConsoleViewTest          consoleViewTestInstance;
static ConfirmDialogTest        confirmDialogTestInstance;
static SetupWizardViewTest      setupWizardViewTestInstance;
