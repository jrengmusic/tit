#pragma once

namespace jreng::tui
{ /*____________________________________________________________________________*/

struct KeyEvent; // forward declaration — jreng_key_event.h NOT included here

// ============================================================================
// Component
// ============================================================================

/** Base class for all CAROLINE TUI components.
 *
 *  Paint contract:
 *    paint() is message-thread only. g is pre-clipped to getBounds() by the
 *    caller (Screen or parent component). The Graphics& reference must
 *    not be cached — it is valid only for the duration of the paint() call.
 *    Focusable components should call g.emitCursorMarker() to position the
 *    hardware cursor.
 */
class Component : public juce::Component
{
public:
    void         paint       (juce::Graphics&) override final {}
    virtual void paint       (Graphics& g)            = 0;
    virtual void handleInput (const KeyEvent& event)  {}
    virtual void invalidate  ()                       {}
    virtual bool isFocusable () const                 { return false; }
};

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
