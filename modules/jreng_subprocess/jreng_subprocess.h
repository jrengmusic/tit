/*******************************************************************************
 BEGIN_JUCE_MODULE_DECLARATION
   ID:                 jreng_subprocess
   vendor:             jreng
   version:            0.1.0
   name:               Subprocess
   description:        JUCE-based subprocess launcher with streaming output and byte-cap truncation.
   dependencies:       jreng_core, juce_core, juce_events
   minimumCppStandard: 17
 END_JUCE_MODULE_DECLARATION
 *******************************************************************************/

#pragma once

#include <juce_core/juce_core.h>
#include <juce_events/juce_events.h>
#include <jreng_core/jreng_core.h>

#include "subprocess/jreng_subprocess_subprocess.h"
