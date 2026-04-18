namespace jreng::tui
{ /*____________________________________________________________________________*/

// ============================================================================
// OverlayHandle
// ============================================================================

OverlayHandle::OverlayHandle (int overlayId, Screen* ownerScreen)
    : id     (overlayId)
    , screen (ownerScreen)
{
}

OverlayHandle::~OverlayHandle()
{
    if (screen != nullptr)
        screen->hideOverlay (id);
}

OverlayHandle::OverlayHandle (OverlayHandle&& other) noexcept
    : id     (other.id)
    , screen (other.screen)
{
    other.screen = nullptr;
}

OverlayHandle& OverlayHandle::operator= (OverlayHandle&& other) noexcept
{
    if (screen != nullptr)
        screen->hideOverlay (id);

    id           = other.id;
    screen       = other.screen;
    other.screen = nullptr;

    return *this;
}

// ============================================================================
// Screen
// ============================================================================

Screen::Screen (Writer& writerRef)
    : writer (writerRef)
{
}

void Screen::start()
{
    requestRender();
}

void Screen::requestRender()
{
    if (not renderPending.exchange (true))
        juce::MessageManager::callAsync ([this] { doRender(); });
}

void Screen::requestFullRedraw()
{
    fullRedrawPending = true;
    requestRender();
}

void Screen::onTerminalResized()
{
    requestFullRedraw();
}

void Screen::paint (Graphics& g)
{
    renderChildren (g);
}

// ----------------------------------------------------------------------------

void Screen::doRender()
{
    renderPending.store (false);

    const auto termSize { getBounds() };
    Graphics g (termSize.getWidth(), termSize.getHeight());
    paint (g);
    renderOverlays (g);

    auto newFrame    = g.getLines();
    auto cursorPos   = extractCursorMarker (newFrame);

    if (firstRender)
    {
        renderFirstFrame (newFrame);
        firstRender = false;
    }
    else if (fullRedrawPending)
    {
        renderFull (newFrame);
        fullRedrawPending = false;
    }
    else
    {
        renderDifferential (newFrame);
    }

    positionHardwareCursor (cursorPos);
    previousFrame = newFrame;
}

// ----------------------------------------------------------------------------

void Screen::renderFirstFrame (const juce::StringArray& newFrame)
{
    writer.beginFrame();

    for (int i { 0 }; i < newFrame.size(); ++i)
    {
        writer.moveTo (i, 0);
        writer.writeLine (newFrame[i]);
    }

    writer.endFrame();
}

void Screen::renderFull (const juce::StringArray& newFrame)
{
    writer.beginFrame();
    writer.clearScreen();
    writer.moveTo (0, 0);

    for (int i { 0 }; i < newFrame.size(); ++i)
        writer.writeLine (newFrame[i]);

    writer.endFrame();
}

void Screen::renderDifferential (const juce::StringArray& newFrame)
{
    int firstChanged { -1 };
    int lastChanged  { -1 };
    const int compareSize { juce::jmin (newFrame.size(), previousFrame.size()) };

    for (int i { 0 }; i < compareSize; ++i)
    {
        if (newFrame[i] != previousFrame.getReference (i))
        {
            if (firstChanged < 0)
                firstChanged = i;

            lastChanged = i;
        }
    }

    for (int i { compareSize }; i < newFrame.size(); ++i)
    {
        if (firstChanged < 0)
            firstChanged = i;

        lastChanged = i;
    }

    if (firstChanged >= 0)
        writeDifferentialRange (newFrame, firstChanged, lastChanged);
}

void Screen::writeDifferentialRange (const juce::StringArray& newFrame,
                                     int firstChanged,
                                     int lastChanged)
{
    writer.beginFrame();

    for (int i { firstChanged }; i <= lastChanged; ++i)
    {
        writer.moveTo (i, 0);
        writer.clearLine();
        writer.writeLine (newFrame[i]);
    }

    if (previousFrame.size() > newFrame.size())
    {
        for (int i { newFrame.size() }; i < previousFrame.size(); ++i)
        {
            writer.moveTo (i, 0);
            writer.clearLine();
        }
    }

    writer.endFrame();
}

// ----------------------------------------------------------------------------

void Screen::renderChildren (Graphics& g)
{
    for (int i { 0 }; i < getNumChildComponents(); ++i)
    {
        auto* child = getChildComponent (i);

        if (auto* ansiChild = dynamic_cast<Component*> (child))
        {
            Graphics clipped { g.clip (Rectangle::fromJuce (child->getBounds())) };
            ansiChild->paint (clipped);
        }
    }
}

void Screen::renderOverlays (Graphics& g)
{
    for (auto& overlay : overlays)
    {
        if (not overlay.hidden)
        {
            Graphics clipped { g.clip (overlay.bounds) };
            overlay.component->paint (clipped);
        }
    }
}

// ----------------------------------------------------------------------------

OverlayHandle Screen::showOverlay (Component* component,
                                   Rectangle bounds)
{
    Overlay entry;
    entry.component  = component;
    entry.bounds     = bounds;
    entry.hidden     = false;
    entry.focusOrder = nextOverlayId;

    overlays.add (entry);
    return OverlayHandle { nextOverlayId++, this };
}

void Screen::hideOverlay (int id)
{
    int indexToRemove { -1 };

    for (int i { 0 }; i < overlays.size(); ++i)
    {
        if (overlays.getReference (i).focusOrder == id)
            indexToRemove = i;
    }

    if (indexToRemove >= 0)
    {
        overlays.remove (indexToRemove);
        requestRender();
    }
}

void Screen::hideTopOverlay()
{
    if (overlays.size() > 0)
    {
        overlays.remove (overlays.size() - 1);
        requestRender();
    }
}

bool Screen::hasOverlay() const noexcept
{
    return overlays.size() > 0;
}

// ----------------------------------------------------------------------------

juce::Point<int> Screen::extractCursorMarker (juce::StringArray& frame)
{
    juce::Point<int> result { -1, -1 };
    const juce::String marker { ANSI::CURSOR_MARKER };

    for (int row { 0 }; row < frame.size(); ++row)
    {
        const int col { frame.getReference (row).indexOf (marker) };

        if (col >= 0)
        {
            frame.getReference (row) = frame.getReference (row).replace (marker, {});
            result = { col, row };
        }
    }

    return result;
}

void Screen::positionHardwareCursor (juce::Point<int> cursorPos)
{
    if (cursorPos.x >= 0 and cursorPos.y >= 0)
        writer.moveTo (cursorPos.y, cursorPos.x);
    else
        writer.hideCursor();
}

/**______________________________END OF NAMESPACE______________________________*/
}// namespace jreng::tui
