# CMakeLists.txt - TIT-cpp: Terminal Interface for git
# ============================================================================
# State-aware git TUI using JUCE (console application)
# ============================================================================

cmake_minimum_required(VERSION 4.2.0)

# ============================================================================
# JAM BUILD SYSTEM
# ============================================================================

set(JAM_ROOT "$ENV{HOME}/Documents/Poems/dev/jam")
list(APPEND CMAKE_MODULE_PATH "${JAM_ROOT}/cmake")
include(BuildSetup)

# ============================================================================
# PROJECT
# ============================================================================

if(APPLE)
    set(_EXTRA_LANGS OBJC OBJCXX)
else()
    set(_EXTRA_LANGS "")
endif()

project(tit-cpp
    VERSION 0.0.0
    LANGUAGES C CXX ${_EXTRA_LANGS}
    DESCRIPTION "TIT-cpp: Terminal Interface for git"
)

# ============================================================================
# BUILD
# ============================================================================

configure_app(
    TARGET_NAME      titc
    PRODUCT_NAME     titc
    VERSION          "${PROJECT_VERSION}"
    COMPANY          jrengmusic
    COMPANY_WEBSITE  "https://jrengmusic.com"
    BUNDLE_ID        "com.jrengmusic.titc"
    MODULES
        juce_core
        juce_events
        juce_data_structures
        juce_graphics
        juce_gui_basics
    JAM_MODULES
        jam_core
        jam_data_structures
        jam_markdown
        jam_mermaid
        jam_tui
        jam_subprocess
    BINARY_FILES     "${CMAKE_CURRENT_SOURCE_DIR}/Assets/tit_logo.svg"
    BINARY_NAMESPACE "BinaryData"
    EXTRA_DEFINES
        JUCE_WEB_BROWSER=0
        JUCE_USE_CURL=0
)

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
message(STATUS "Configuration complete!")
message(STATUS "Generator: ${CMAKE_GENERATOR}")
message(STATUS "Build directory: ${CMAKE_BINARY_DIR}")
message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
