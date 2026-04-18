/*******************************************************************************
 BEGIN_JUCE_MODULE_DECLARATION
   ID:                 jreng_tui
   vendor:             jreng
   version:            0.1.0
   name:               TUI framework
   description:        JUCE-based TUI framework — ANSI rendering, terminal metrics, raw input, ANSI markdown renderer.
   dependencies:       jreng_core, jreng_markdown, juce_core, juce_events, juce_graphics, juce_gui_basics
   minimumCppStandard: 17
 END_JUCE_MODULE_DECLARATION
 *******************************************************************************/
#pragma once

#include <juce_core/juce_core.h>
#include <juce_events/juce_events.h>
#include <juce_graphics/juce_graphics.h>
#include <juce_gui_basics/juce_gui_basics.h>
#include <jreng_core/jreng_core.h>
#include <jreng_markdown/jreng_markdown.h>

#include "ansi/jreng_ansi_escapes.h"
#include "ansi/jreng_color.h"
#include "ansi/jreng_cell.h"
#include "graphics/jreng_tui_rectangle.h"
#include "metrics/jreng_tui_metrics.h"
#include "ansi/jreng_ansi_writer.h"
#include "ansi/jreng_ansi_graphics.h"
#include "ansi/jreng_ansi_component.h"
#include "input/jreng_key_event.h"
#include "ansi/jreng_textbox.h"
#include "ansi/jreng_ansi_screen.h"
#include "input/jreng_tui_input.h"
#include "markdown/jreng_ansi_markdown_renderer.h"
#include "braille/jreng_braille_grid.h"
