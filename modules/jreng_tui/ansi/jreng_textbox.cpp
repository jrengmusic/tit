// ============================================================================
// TextBox implementation
// ============================================================================

namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// Key binding table
// ============================================================================

const TextBox::KeyBinding TextBox::keyBindings[]
{
    { KeyType::Character,  &TextBox::handleCharacter  },
    { KeyType::Backspace,  &TextBox::handleBackspace  },
    { KeyType::Delete,     &TextBox::handleDeleteKey  },
    { KeyType::ArrowLeft,  &TextBox::handleArrowLeft  },
    { KeyType::ArrowRight, &TextBox::handleArrowRight },
    { KeyType::Home,       &TextBox::handleHome       },
    { KeyType::End,        &TextBox::handleEnd        },
    { KeyType::Enter,      &TextBox::handleEnter      },
    { KeyType::Escape,     &TextBox::handleEscape     },
    { KeyType::Paste,      &TextBox::handlePaste      },
};

// ============================================================================
// Construction
// ============================================================================

TextBox::TextBox()
    : placeholderColour { 0xFF888888u }
{
}

// ============================================================================
// Text API
// ============================================================================

juce::String TextBox::getText() const
{
    return content;
}

bool TextBox::isEmpty() const
{
    return content.isEmpty();
}

void TextBox::setText (const juce::String& newText, bool sendChangeMessage)
{
    content       = newText;
    caretPosition = content.length();
    scrollOffset  = 0;

    if (sendChangeMessage and onTextChange != nullptr)
        onTextChange();
}

void TextBox::insertTextAtCaret (const juce::String& textToInsert)
{
    content       = content.substring (0, caretPosition)
                  + textToInsert
                  + content.substring (caretPosition);
    caretPosition += textToInsert.length();

    if (onTextChange != nullptr)
        onTextChange();
}

void TextBox::clear()
{
    content       = {};
    caretPosition = 0;
    scrollOffset  = 0;

    if (onTextChange != nullptr)
        onTextChange();
}

// ============================================================================
// Caret API
// ============================================================================

int TextBox::getCaretPosition() const
{
    return caretPosition;
}

void TextBox::setCaretPosition (int newIndex)
{
    caretPosition = juce::jlimit (0, content.length(), newIndex);
}

void TextBox::moveCaretToEnd()
{
    caretPosition = content.length();
}

// ============================================================================
// Appearance API
// ============================================================================

void TextBox::setTextToShowWhenEmpty (const juce::String& text, juce::Colour colourToUse)
{
    placeholderText   = text;
    placeholderColour = colourToUse;
}

void TextBox::setCaretVisible (bool shouldBeVisible)
{
    caretVisible = shouldBeVisible;
}

// ============================================================================
// Render
// ============================================================================

void TextBox::paint (Graphics& g)
{
    const auto bounds { getBounds() };
    const int  cols   { bounds.getWidth() };
    const int  rows   { bounds.getHeight() };

    renderBorder  (g, cols, rows);
    renderContent (g, cols);
}

void TextBox::renderBorder (Graphics& g, int cols, int rows)
{
    // Top border: ╭─ ... ─╮
    juce::String topRow;
    topRow += juce::String::charToString (static_cast<juce::juce_wchar> (cornerTopLeft));

    for (int i { 1 }; i < cols - 1; ++i)
        topRow += juce::String::charToString (static_cast<juce::juce_wchar> (horizontalBar));

    topRow += juce::String::charToString (static_cast<juce::juce_wchar> (cornerTopRight));

    g.drawCellText (topRow, 0, 0, cols);

    // Side bars for middle rows
    const juce::String sideBar { juce::String::charToString (
        static_cast<juce::juce_wchar> (verticalBar)) };

    for (int row { 1 }; row < rows - 1; ++row)
    {
        g.drawCellText (sideBar, 0,        row, 1);
        g.drawCellText (sideBar, cols - 1, row, 1);
    }

    // Bottom border: ╰─ ... ─╯
    juce::String bottomRow;
    bottomRow += juce::String::charToString (static_cast<juce::juce_wchar> (cornerBottomLeft));

    for (int i { 1 }; i < cols - 1; ++i)
        bottomRow += juce::String::charToString (static_cast<juce::juce_wchar> (horizontalBar));

    bottomRow += juce::String::charToString (static_cast<juce::juce_wchar> (cornerBottomRight));

    g.drawCellText (bottomRow, 0, rows - 1, cols);
}

void TextBox::renderContent (Graphics& g, int cols)
{
    // Content lives in row 1, columns 1 to cols-2 (inside the border).
    // Layout: │ ❯ [text...]  │
    // Prefix: "❯ " = 2 cells after the left bar and space.
    // Available text columns: cols - 1 (left bar) - 1 (space) - 2 (prompt+space) - 1 (right bar)

    static constexpr int prefixWidth { 4 }; // "│ ❯ "
    const int            textAreaCols { cols - prefixWidth - 1 };

    const juce::String arrowStr { juce::String::charToString (
        static_cast<juce::juce_wchar> (promptChar)) };

    g.drawCellText (arrowStr, 2, 1, 1);

    if (content.isEmpty() and not placeholderText.isEmpty())
    {
        g.setColour (placeholderColour);
        g.drawCellText (placeholderText, prefixWidth, 1, textAreaCols);
    }
    else
    {
        // Compute visible window: scroll so caret is visible
        if (caretPosition - scrollOffset >= textAreaCols)
            scrollOffset = caretPosition - textAreaCols + 1;

        if (caretPosition < scrollOffset)
            scrollOffset = caretPosition;

        const juce::String visible { content.substring (scrollOffset,
                                     scrollOffset + textAreaCols) };
        g.drawCellText (visible, prefixWidth, 1, textAreaCols);

        if (caretVisible)
        {
            const int caretCol { prefixWidth + (caretPosition - scrollOffset) };
            g.emitCursorMarker (caretCol, 1);
        }
    }
}

// ============================================================================
// Input dispatch
// ============================================================================

void TextBox::handleInput (const KeyEvent& event)
{
    for (int i { 0 }; i < keyBindingCount; ++i)
    {
        if (keyBindings[i].key == event.type)
        {
            (this->*keyBindings[i].handler) (event);
            break;
        }
    }
}

void TextBox::handleCharacter (const KeyEvent& event)
{
    const juce::String ch { juce::String::charToString (
        static_cast<juce::juce_wchar> (event.character)) };

    content       = content.substring (0, caretPosition)
                  + ch
                  + content.substring (caretPosition);
    ++caretPosition;

    if (onTextChange != nullptr)
        onTextChange();
}

void TextBox::handleBackspace (const KeyEvent&)
{
    if (caretPosition > 0)
    {
        content = content.substring (0, caretPosition - 1)
                + content.substring (caretPosition);
        --caretPosition;

        if (onTextChange != nullptr)
            onTextChange();
    }
}

void TextBox::handleDeleteKey (const KeyEvent&)
{
    if (caretPosition < content.length())
    {
        content = content.substring (0, caretPosition)
                + content.substring (caretPosition + 1);

        if (onTextChange != nullptr)
            onTextChange();
    }
}

void TextBox::handleArrowLeft (const KeyEvent&)
{
    if (caretPosition > 0)
        --caretPosition;
}

void TextBox::handleArrowRight (const KeyEvent&)
{
    if (caretPosition < content.length())
        ++caretPosition;
}

void TextBox::handleHome (const KeyEvent&)
{
    caretPosition = 0;
    scrollOffset  = 0;
}

void TextBox::handleEnd (const KeyEvent&)
{
    caretPosition = content.length();
}

void TextBox::handleEnter (const KeyEvent&)
{
    if (onReturnKey != nullptr)
        onReturnKey();
}

void TextBox::handleEscape (const KeyEvent&)
{
    if (onEscapeKey != nullptr)
        onEscapeKey();
}

void TextBox::handlePaste (const KeyEvent& event)
{
    insertTextAtCaret (event.pasteContent);
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
