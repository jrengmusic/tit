#pragma once

namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// TextBox
// ============================================================================

/** Single-line text input component.  API mirrors juce::TextEditor.
 *
 *  Renders as a rounded-rectangle box using box-drawing codepoints.
 *  Key dispatch is driven by a lookup table — no switch chains.
 *  Coordinate contract: setBounds() uses cell coordinates (col, row).
 */
class TextBox : public Component
{
public:
    TextBox();

    // -------------------------------------------------------------------------
    // Text
    // -------------------------------------------------------------------------
    juce::String getText()   const;
    bool         isEmpty()   const;

    void setText            (const juce::String& newText, bool sendChangeMessage = true);
    void insertTextAtCaret  (const juce::String& textToInsert);
    void clear();

    // -------------------------------------------------------------------------
    // Caret
    // -------------------------------------------------------------------------
    int  getCaretPosition() const;
    void setCaretPosition   (int newIndex);
    void moveCaretToEnd();

    // -------------------------------------------------------------------------
    // Appearance
    // -------------------------------------------------------------------------
    void setTextToShowWhenEmpty (const juce::String& text, juce::Colour colourToUse);
    void setCaretVisible        (bool shouldBeVisible);

    // -------------------------------------------------------------------------
    // Callbacks
    // -------------------------------------------------------------------------
    std::function<void()> onTextChange;
    std::function<void()> onReturnKey;
    std::function<void()> onEscapeKey;
    std::function<void()> onFocusLost;

    // -------------------------------------------------------------------------
    // Component overrides
    // -------------------------------------------------------------------------
    void paint       (Graphics& g)            override;
    void handleInput (const KeyEvent& event)  override;
    bool isFocusable ()                 const override { return true; }

private:
    // -------------------------------------------------------------------------
    // State
    // -------------------------------------------------------------------------
    juce::String content;
    int          caretPosition  { 0 };
    int          scrollOffset   { 0 };
    bool         caretVisible   { true };

    juce::String placeholderText;
    juce::Colour placeholderColour;

    // -------------------------------------------------------------------------
    // Box-drawing codepoints
    // -------------------------------------------------------------------------
    static constexpr uint32_t cornerTopLeft     { 0x256du };
    static constexpr uint32_t cornerTopRight    { 0x256eu };
    static constexpr uint32_t cornerBottomLeft  { 0x2570u };
    static constexpr uint32_t cornerBottomRight { 0x256fu };
    static constexpr uint32_t horizontalBar     { 0x2500u };
    static constexpr uint32_t verticalBar       { 0x2502u };
    static constexpr uint32_t promptChar        { 0x276fu };

    // -------------------------------------------------------------------------
    // Render helpers
    // -------------------------------------------------------------------------
    void renderBorder  (Graphics& g, int cols, int rows);
    void renderContent (Graphics& g, int cols);

    // -------------------------------------------------------------------------
    // Input dispatch
    // -------------------------------------------------------------------------
    using HandlerFn = void (TextBox::*)(const KeyEvent&);
    struct KeyBinding { KeyType key; HandlerFn handler; };

    void handleCharacter  (const KeyEvent& event);
    void handleBackspace  (const KeyEvent& event);
    void handleDeleteKey  (const KeyEvent& event);
    void handleArrowLeft  (const KeyEvent& event);
    void handleArrowRight (const KeyEvent& event);
    void handleHome       (const KeyEvent& event);
    void handleEnd        (const KeyEvent& event);
    void handleEnter      (const KeyEvent& event);
    void handleEscape     (const KeyEvent& event);
    void handlePaste      (const KeyEvent& event);

    static const KeyBinding keyBindings[];
    static constexpr int    keyBindingCount { 10 };

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TextBox)
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
