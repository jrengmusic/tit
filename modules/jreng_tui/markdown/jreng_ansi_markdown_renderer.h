#pragma once

#include <unordered_map>

namespace jreng::tui
{ /*____________________________________________________________________________*/

class MarkdownRenderer
{
public:
    using BlockRenderer = int(*)(Graphics&,
                                  const jreng::Markdown::Block&,
                                  const jreng::Markdown::ParsedDocument&,
                                  int startRow, int widthCols,
                                  juce::Colour defaultFg);

    int paint (Graphics& g,
                const jreng::Markdown::ParsedDocument& doc,
                int startRow,
                int widthCols,
                juce::Colour defaultFg);

    void invalidate();

private:
    static const std::unordered_map<Markdown::BlockType, BlockRenderer> blockRenderers;

    int cachedWidthCols { 0 };
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
