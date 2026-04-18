namespace jreng::Markdown
{ /*____________________________________________________________________________*/

// ============================================================================
// Block Types
// ============================================================================

enum class BlockType
{
    Markdown,   ///< A block of standard markdown text.
    CodeFence,  ///< A generic fenced code block.
    Table       ///< A GitHub-flavored markdown table.
};

// ============================================================================
// Line Classification
// ============================================================================

enum class LineType
{
    Header,
    ListItem,
    Paragraph,
    ThematicBreak,
    Blank
};

// ============================================================================
// Inline Styles
// ============================================================================

enum InlineStyle : uint16_t
{
    None   = 0,
    Bold   = 1u << 0,
    Italic = 1u << 1,
    Code   = 1u << 2,
    Link   = 1u << 3
};

inline InlineStyle operator| (InlineStyle a, InlineStyle b)
{ return static_cast<InlineStyle> (static_cast<uint16_t> (a) | static_cast<uint16_t> (b)); }

inline InlineStyle operator& (InlineStyle a, InlineStyle b)
{ return static_cast<InlineStyle> (static_cast<uint16_t> (a) & static_cast<uint16_t> (b)); }

inline InlineStyle operator^ (InlineStyle a, InlineStyle b)
{ return static_cast<InlineStyle> (static_cast<uint16_t> (a) ^ static_cast<uint16_t> (b)); }

inline InlineStyle operator~ (InlineStyle a)
{ return static_cast<InlineStyle> (~static_cast<uint16_t> (a)); }

inline InlineStyle& operator|= (InlineStyle& a, InlineStyle b) { a = a | b; return a; }
inline InlineStyle& operator&= (InlineStyle& a, InlineStyle b) { a = a & b; return a; }
inline InlineStyle& operator^= (InlineStyle& a, InlineStyle b) { a = a ^ b; return a; }

// ============================================================================
// Inline Span
// ============================================================================

struct InlineSpan
{
    int startOffset { 0 };  ///< Relative to block content offset in ParsedDocument::text
    int endOffset { 0 };
    InlineStyle style { InlineStyle::None };  ///< Bitmask: Bold | Italic | Code | Link
    int uriOffset { 0 };      ///< Into ParsedDocument::text (0 if not a link)
    int uriLength { 0 };

    // Resolved style (filled by Parser when StyleConfig is provided)
    float fontSize { 0.0f };      ///< Resolved point height
    juce::Colour colour {};       ///< Resolved text colour
    uint8_t fontFamily { 0 };     ///< 0 = body, 1 = code
};

// ============================================================================
// Block
// ============================================================================

struct Block
{
    BlockType type { BlockType::Markdown };  ///< Markdown, CodeFence, Table
    int contentOffset { 0 };    ///< Into ParsedDocument::text
    int contentLength { 0 };
    int languageOffset { 0 };   ///< Into ParsedDocument::text (code fence language tag)
    int languageLength { 0 };
    int spanOffset { 0 };       ///< Index into ParsedDocument::spans
    int spanCount { 0 };
    int level { 0 };            ///< Heading level 1-6, 0 for non-headings

    // Resolved style (filled by Parser when StyleConfig is provided)
    float fontSize { 0.0f };      ///< Resolved from heading level
    juce::Colour colour {};       ///< Resolved heading/body colour
};

static_assert (std::is_trivially_copyable_v<Block>,      "Block must be trivially copyable");
static_assert (std::is_trivially_copyable_v<InlineSpan>, "InlineSpan must be trivially copyable");

// ============================================================================
// Parsed Document
// ============================================================================

struct ParsedDocument
{
    juce::HeapBlock<char> text {};
    int textSize { 0 };
    int textCapacity { 0 };

    juce::HeapBlock<Block> blocks {};
    int blockCount { 0 };
    int blockCapacity { 0 };

    juce::HeapBlock<InlineSpan> spans {};
    int spanCount { 0 };
    int spanCapacity { 0 };
};

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::Markdown
