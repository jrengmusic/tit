#pragma once

#include <juce_core/juce_core.h>
#include <juce_data_structures/juce_data_structures.h>

namespace jreng::json
{ /*____________________________________________________________________________*/

class ValueTree
{
public:
    static juce::ValueTree fromJson (const juce::var& source,
                                     const juce::Identifier& rootId);

    static juce::var toJson (const juce::ValueTree& tree);

private:
    ValueTree() = delete;
};

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::json
