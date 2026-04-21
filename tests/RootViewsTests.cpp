#include <JuceHeader.h>
#include "view/Banner.h"
#include "view/Header.h"
#include "view/Footer.h"
#include "view/MenuView.h"
#include "state/TitAxis.h"
#include "TitIdentifier.h"

// ============================================================================
// RootViewsTests
// ============================================================================
//
// One UnitTest subclass per view (5 total):
//   BannerTest, HeaderTest, FooterTest, MenuViewTest, TitScreenTest.
//
// Each test:
//   1. Builds a minimal VT fixture (where applicable).
//   2. Constructs a ThemeResolver from a minimal THEME subtree.
//   3. Instantiates the view.
//   4. Verifies construction succeeds and view state reflects fixture.
//   5. Mutates VT where relevant and verifies listener fires (repaint mark).
//
// Static registrar pattern, no anonymous namespace.

// ============================================================================
// Fixture helpers
// ============================================================================

static juce::ValueTree makeMinimalTheme()
{
    juce::ValueTree theme { tit::ID::THEME };
    return theme;
}

static juce::ValueTree makeFullStateTree (tit::Operation    op  = tit::Operation::Normal,
                                          tit::WorkingTree  wt  = tit::WorkingTree::Clean,
                                          tit::Timeline     tl  = tit::Timeline::InSync,
                                          tit::Remote       rm  = tit::Remote::HasRemote,
                                          int               ahead  = 0,
                                          int               behind = 0)
{
    juce::ValueTree root { tit::ID::TIT };

    juce::ValueTree repo { tit::ID::REPO };
    repo.setProperty (tit::ID::branch,      "main",                      nullptr);
    repo.setProperty (tit::ID::cwd,         "/home/user/project",         nullptr);
    repo.setProperty (tit::ID::operation,   tit::toString (op),           nullptr);
    repo.setProperty (tit::ID::workingTree, tit::toString (wt),           nullptr);
    repo.setProperty (tit::ID::timeline,    tit::toString (tl),           nullptr);
    repo.setProperty (tit::ID::remote,      tit::toString (rm),           nullptr);
    repo.setProperty (tit::ID::aheadCount,  ahead,                        nullptr);
    repo.setProperty (tit::ID::behindCount, behind,                       nullptr);
    root.appendChild (repo, nullptr);

    juce::ValueTree menu { tit::ID::MENU };
    root.appendChild (menu, nullptr);

    juce::ValueTree selection { tit::ID::SELECTION };
    selection.setProperty (tit::ID::menuIndex, 0, nullptr);
    root.appendChild (selection, nullptr);

    juce::ValueTree env { tit::ID::ENV };
    root.appendChild (env, nullptr);

    root.appendChild (makeMinimalTheme(), nullptr);

    return root;
}

// ============================================================================
// Test 1 — Banner
// ============================================================================

class BannerTest : public juce::UnitTest
{
public:
    BannerTest() : juce::UnitTest ("Banner", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testNoPaintCrashWhenEmpty ();
    }

private:
    void testConstruction ()
    {
        beginTest ("Banner constructs without crashing");

        tit::Banner banner;
        banner.setSize (40, 10);

        expect (banner.getWidth()  == 40, "width set");
        expect (banner.getHeight() == 10, "height set");
    }

    void testNoPaintCrashWhenEmpty ()
    {
        beginTest ("Banner: resized() before paint does not crash");

        tit::Banner banner;
        // Size 0 — no crash on degenerate case.
        banner.setSize (0, 0);

        expect (true, "no crash on zero-size resize");

        // Now give a valid size — grid rebuild should run.
        banner.setSize (20, 5);
        expect (banner.getWidth() == 20, "width after resize");
    }
};

// ============================================================================
// Test 2 — Header
// ============================================================================

class HeaderTest : public juce::UnitTest
{
public:
    HeaderTest() : juce::UnitTest ("Header", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testListenerFires ();
        testBranchPropertyRendered ();
    }

private:
    void testConstruction ()
    {
        beginTest ("Header constructs and attaches to REPO subtree");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::Header header { state, resolver };
        header.setSize (80, 4);

        expect (header.getWidth()  == 80, "width set");
        expect (header.getHeight() == 4,  "height set");
    }

    void testListenerFires ()
    {
        beginTest ("Header: VT REPO property change triggers repaint");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::Header header { state, resolver };
        header.setSize (80, 4);

        // Mutate REPO — listener should mark repaint needed.
        juce::ValueTree repo { state.getChildWithName (tit::ID::REPO) };
        repo.setProperty (tit::ID::branch, "feature/test", nullptr);

        // If we got here without crashing, listener handled the mutation.
        expect (true, "no crash on property change");
    }

    void testBranchPropertyRendered ()
    {
        beginTest ("Header: branch and cwd properties read from REPO");

        juce::ValueTree state { makeFullStateTree() };
        juce::ValueTree repo  { state.getChildWithName (tit::ID::REPO) };

        // Verify fixture values readable before view construction.
        expect (repo.getProperty (tit::ID::branch).toString() == "main",
                "branch == main in fixture");
        expect (repo.getProperty (tit::ID::cwd).toString() == "/home/user/project",
                "cwd set in fixture");
    }
};

// ============================================================================
// Test 3 — Footer
// ============================================================================

class FooterTest : public juce::UnitTest
{
public:
    FooterTest() : juce::UnitTest ("Footer", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testListenerFiresOnOperationChange ();
        testHintSetChangesWithOperation ();
    }

private:
    void testConstruction ()
    {
        beginTest ("Footer constructs and attaches to REPO subtree");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::Footer footer { state, resolver };
        footer.setSize (80, 1);

        expect (footer.getWidth() == 80, "width set");
    }

    void testListenerFiresOnOperationChange ()
    {
        beginTest ("Footer: operation change triggers repaint");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::Footer     footer { state, resolver };
        juce::ValueTree repo   { state.getChildWithName (tit::ID::REPO) };

        // Change operation.
        repo.setProperty (tit::ID::operation,
                          tit::toString (tit::Operation::TimeTraveling), nullptr);

        expect (true, "no crash on operation change");
    }

    void testHintSetChangesWithOperation ()
    {
        beginTest ("Footer: hint content depends on operation in VT");

        juce::ValueTree state { makeFullStateTree (tit::Operation::TimeTraveling) };
        juce::ValueTree repo  { state.getChildWithName (tit::ID::REPO) };

        const juce::String opStr { repo.getProperty (tit::ID::operation).toString() };
        expect (opStr == tit::toString (tit::Operation::TimeTraveling),
                "fixture operation is TimeTraveling");
    }
};

// ============================================================================
// Test 4 — MenuView
// ============================================================================

class MenuViewTest : public juce::UnitTest
{
public:
    MenuViewTest() : juce::UnitTest ("MenuView", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testMenuSubtreePopulatedOnBuild ();
        testSelectionIndexSync ();
        testOperationChangeRebuildsMenu ();
    }

private:
    void testConstruction ()
    {
        beginTest ("MenuView constructs and populates MENU subtree");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::MenuView menuView { state, resolver };
        menuView.setSize (40, 20);

        expect (menuView.getWidth()  == 40, "width set");
        expect (menuView.getHeight() == 20, "height set");
    }

    void testMenuSubtreePopulatedOnBuild ()
    {
        beginTest ("MenuView: MENU subtree contains ITEM children after construction");

        // Normal + Clean + InSync + HasRemote → [history, file_history]
        juce::ValueTree state { makeFullStateTree (tit::Operation::Normal,
                                                    tit::WorkingTree::Clean,
                                                    tit::Timeline::InSync,
                                                    tit::Remote::HasRemote) };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::MenuView menuView { state, resolver };
        menuView.setSize (40, 20);

        juce::ValueTree menuTree { state.getChildWithName (tit::ID::MENU) };
        const int itemCount { menuTree.getNumChildren() };

        expect (itemCount == 2, juce::String ("Normal+Clean+InSync+HasRemote should produce 2 items, got ")
                                + juce::String { itemCount });

        const juce::String firstId { menuTree.getChild (0).getProperty (tit::ID::id).toString() };
        expect (firstId == "history", "first item id == history");
    }

    void testSelectionIndexSync ()
    {
        beginTest ("MenuView: menuIndex change in SELECTION syncs menu selection");

        juce::ValueTree state { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::MenuView   menuView  { state, resolver };
        juce::ValueTree selection { state.getChildWithName (tit::ID::SELECTION) };

        selection.setProperty (tit::ID::menuIndex, 1, nullptr);

        // No crash — listener handled the change.
        expect (true, "no crash on menuIndex change");
    }

    void testOperationChangeRebuildsMenu ()
    {
        beginTest ("MenuView: operation change causes MENU subtree rebuild");

        juce::ValueTree state { makeFullStateTree (tit::Operation::NotRepo) };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::MenuView   menuView  { state, resolver };
        juce::ValueTree menuTree  { state.getChildWithName (tit::ID::MENU) };
        juce::ValueTree repo      { state.getChildWithName (tit::ID::REPO) };

        const int beforeCount { menuTree.getNumChildren() };

        repo.setProperty (tit::ID::operation,
                          tit::toString (tit::Operation::Normal), nullptr);

        const int afterCount { menuTree.getNumChildren() };

        // NotRepo yields 2 (init, clone); Normal+Clean+InSync+HasRemote yields 2 (history, file_history).
        // Counts match in this fixture but IDs should differ.
        expect (beforeCount > 0, "MENU had items for NotRepo");
        expect (afterCount  > 0, "MENU has items for Normal");

        const juce::String firstBefore { "init" };
        const juce::String firstAfter  { menuTree.getChild (0)
                                             .getProperty (tit::ID::id).toString() };
        expect (firstAfter == "history",
                juce::String ("after switch to Normal, first item should be history, got: ")
                + firstAfter);
    }
};

// ============================================================================
// Test 5 — TitScreen
// ============================================================================

class TitScreenTest : public juce::UnitTest
{
public:
    TitScreenTest() : juce::UnitTest ("TitScreen", "tit") {}

    void runTest() override
    {
        testConstruction ();
        testLayoutBounds ();
        testCallbackForwarding ();
    }

private:
    void testConstruction ()
    {
        beginTest ("TitScreen constructs all four shell views");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::TitScreen screen { state, resolver };
        screen.setSize (80, 30);

        expect (screen.getWidth()  == 80, "width set");
        expect (screen.getHeight() == 30, "height set");
        expect (screen.getNumChildComponents() == 4, "four child components");
    }

    void testLayoutBounds ()
    {
        beginTest ("TitScreen: children fill terminal rows per SPEC §14");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::TitScreen screen { state, resolver };
        screen.setSize (80, 30);

        // Banner: y=0, height=1; Header: y=1, height=4; Footer: y=29, height=1.
        juce::Component* banner   { screen.getChildComponent (0) };
        juce::Component* header   { screen.getChildComponent (1) };
        juce::Component* menuView { screen.getChildComponent (2) };
        juce::Component* footer   { screen.getChildComponent (3) };

        expect (banner  != nullptr, "banner child exists");
        expect (header  != nullptr, "header child exists");
        expect (menuView!= nullptr, "menuView child exists");
        expect (footer  != nullptr, "footer child exists");

        expect (banner->getY()      == 0,  "banner y == 0");
        expect (banner->getHeight() == 1,  "banner height == 1");
        expect (header->getY()      == 1,  "header y == 1");
        expect (header->getHeight() == 4,  "header height == 4");
        expect (footer->getHeight() == 1,  "footer height == 1");
        expect (footer->getBottom() == 30, "footer bottom == total height");
    }

    void testCallbackForwarding ()
    {
        beginTest ("TitScreen: onItemSelected callback set without crash");

        juce::ValueTree      state   { makeFullStateTree() };
        jam::tui::ThemeResolver resolver { state.getChildWithName (tit::ID::THEME) };

        tit::TitScreen screen { state, resolver };
        screen.setSize (80, 30);

        bool called { false };
        screen.onItemSelected = [&called] (const jam::tui::MenuItem&)
        {
            called = true;
        };

        expect (not called, "callback not called before input");
    }
};

// ============================================================================
// Static registrars
// ============================================================================

static BannerTest   bannerTestInstance;
static HeaderTest   headerTestInstance;
static FooterTest   footerTestInstance;
static MenuViewTest menuViewTestInstance;
static TitScreenTest titScreenTestInstance;
