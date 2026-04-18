// jreng_ansi_markdown_renderer.cpp
// MarkdownRenderer — consumes jreng::Markdown::ParsedDocument, emits ANSI rows.
// Parser SSOT: jreng_markdown. This module only renders.

namespace jreng::tui
{ /*____________________________________________________________________________*/

static juce::String buildInlineStylePrefix (jreng::Markdown::InlineStyle style,
                                             juce::Colour defaultFg)
{
    juce::String prefix;

    if ((style & jreng::Markdown::Bold) != jreng::Markdown::None)
        prefix += ANSI::BOLD_ON;

    if ((style & jreng::Markdown::Italic) != jreng::Markdown::None)
        prefix += ANSI::ITALIC_ON;

    if ((style & jreng::Markdown::Code) != jreng::Markdown::None)
    {
        const juce::Colour codeFg { defaultFg.darker (0.3f) };
        prefix += juce::String (ANSI::CSI_PREFIX)
                  + "38;2;"
                  + juce::String (codeFg.getRed())   + ";"
                  + juce::String (codeFg.getGreen()) + ";"
                  + juce::String (codeFg.getBlue())  + "m";
    }

    return prefix;
}

static juce::String buildSpanText (const jreng::Markdown::ParsedDocument& doc,
                                    int spanIndex,
                                    juce::Colour defaultFg)
{
    jassert (spanIndex < doc.spanCount);
    const jreng::Markdown::InlineSpan& span { doc.spans[spanIndex] };
    const int length { span.endOffset - span.startOffset };
    juce::String text;

    if (length > 0)
    {
        const juce::String content {
            juce::String::fromUTF8 (doc.text + span.startOffset, length)
        };
        const juce::String prefix { buildInlineStylePrefix (span.style, defaultFg) };
        text = prefix + content + ANSI::RESET;
    }

    return text;
}

static juce::String buildInlineRow (const jreng::Markdown::Block& block,
                                     const jreng::Markdown::ParsedDocument& doc,
                                     juce::Colour defaultFg)
{
    juce::String row;

    for (int i { 0 }; i < block.spanCount; ++i)
        row += buildSpanText (doc, block.spanOffset + i, defaultFg);

    return row;
}

static int renderHeadingBlock (Graphics& g,
                                const jreng::Markdown::Block& block,
                                const jreng::Markdown::ParsedDocument& doc,
                                int startRow, int widthCols,
                                juce::Colour defaultFg)
{
    const juce::String content {
        juce::String::fromUTF8 (doc.text + block.contentOffset, block.contentLength)
    };

    const juce::String headingText {
        juce::String (ANSI::BOLD_ON) + content + ANSI::BOLD_OFF
    };

    const Rectangle bounds { 0, startRow, widthCols, 1 };
    g.drawText (headingText, bounds, juce::Justification::left, false);

    return 1;
}

static int renderParagraphBlock (Graphics& g,
                                  const jreng::Markdown::Block& block,
                                  const jreng::Markdown::ParsedDocument& doc,
                                  int startRow, int widthCols,
                                  juce::Colour defaultFg)
{
    const juce::String inlineRow { buildInlineRow (block, doc, defaultFg) };

    juce::AttributedString attributed;
    attributed.append (inlineRow, juce::Font { juce::FontOptions{}.withHeight (14.0f) }, defaultFg);

    const Rectangle bounds { 0, startRow, widthCols, 4 };
    g.drawAttributedString (attributed, bounds);

    return 1;
}

static int renderMarkdownBlock (Graphics& g,
                                 const jreng::Markdown::Block& block,
                                 const jreng::Markdown::ParsedDocument& doc,
                                 int startRow, int widthCols,
                                 juce::Colour defaultFg)
{
    int rowsConsumed { 0 };

    if (block.level > 0)
        rowsConsumed = renderHeadingBlock (g, block, doc, startRow, widthCols, defaultFg);
    else
        rowsConsumed = renderParagraphBlock (g, block, doc, startRow, widthCols, defaultFg);

    return rowsConsumed;
}

static int renderCodeFenceBlock (Graphics& g,
                                  const jreng::Markdown::Block& block,
                                  const jreng::Markdown::ParsedDocument& doc,
                                  int startRow, int widthCols,
                                  juce::Colour defaultFg)
{
    const juce::String content {
        juce::String::fromUTF8 (doc.text + block.contentOffset, block.contentLength)
    };

    const juce::StringArray codeLines { juce::StringArray::fromLines (content) };
    const juce::Colour codeFg { defaultFg.darker (0.2f) };
    int rowsConsumed { 0 };

    for (int i { 0 }; i < codeLines.size(); ++i)
    {
        const juce::String codeLine {
            juce::String (ANSI::CSI_PREFIX)
            + "38;2;"
            + juce::String (codeFg.getRed())   + ";"
            + juce::String (codeFg.getGreen()) + ";"
            + juce::String (codeFg.getBlue())  + "m"
            + codeLines[i] + ANSI::RESET
        };
        const Rectangle bounds { 0, startRow + rowsConsumed, widthCols, 1 };
        g.drawText (codeLine, bounds, juce::Justification::left, false);
        ++rowsConsumed;
    }

    return rowsConsumed;
}

static int renderTableBlock (Graphics& g,
                               const jreng::Markdown::Block& block,
                               const jreng::Markdown::ParsedDocument& doc,
                               int startRow, int widthCols,
                               juce::Colour defaultFg)
{
    // MVP: tables not supported — render as plain text.
    const juce::String content {
        juce::String::fromUTF8 (doc.text + block.contentOffset, block.contentLength)
    };

    const Rectangle bounds { 0, startRow, widthCols, 1 };
    g.drawText (content, bounds, juce::Justification::left, false);

    return 1;
}

// ============================================================================
// Static dispatch table — 3 entries matching BlockType enum
// ============================================================================

const std::unordered_map<jreng::Markdown::BlockType, MarkdownRenderer::BlockRenderer>
    MarkdownRenderer::blockRenderers
{
    { jreng::Markdown::BlockType::Markdown,  renderMarkdownBlock  },
    { jreng::Markdown::BlockType::CodeFence, renderCodeFenceBlock },
    { jreng::Markdown::BlockType::Table,     renderTableBlock     }
};

// ============================================================================
// MarkdownRenderer public API
// ============================================================================

int MarkdownRenderer::paint (Graphics& g,
                               const jreng::Markdown::ParsedDocument& doc,
                               int startRow,
                               int widthCols,
                               juce::Colour defaultFg)
{
    if (widthCols != cachedWidthCols)
    {
        invalidate();
        cachedWidthCols = widthCols;
    }

    int currentRow { startRow };

    for (int i { 0 }; i < doc.blockCount; ++i)
    {
        jassert (i < doc.blockCount);
        const jreng::Markdown::Block& block { doc.blocks[i] };
        const auto it { blockRenderers.find (block.type) };

        if (it != blockRenderers.end())
            currentRow += it->second (g, block, doc, currentRow, widthCols, defaultFg);
    }

    return currentRow - startRow;
}

void MarkdownRenderer::invalidate()
{
    cachedWidthCols = 0;
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
