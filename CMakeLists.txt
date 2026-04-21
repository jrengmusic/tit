# CMakeLists.txt - TIT-cpp: Terminal Interface for git
# ============================================================================
# State-aware git TUI using JUCE (console application)
#
# Usage:
#   cmake -S . -B Builds/Ninja -G Ninja
#   cmake --build Builds/Ninja
#
# Directory Structure:
#   dev/
#   ├── tit/
#   │   ├── CMakeLists.txt      <- This file
#   │   ├── Source/             <- .cpp/.h files
#   │   └── modules/            <- Local JUCE modules (added Step 1.3+)
#   └── JUCE/                   <- JUCE framework (sibling to dev/)
# ============================================================================

cmake_minimum_required(VERSION 4.2.0)


# ============================================================================
# CONFIGURATION
# ============================================================================

set(PROJECT_NAME "tit-cpp")
set(PROJECT_VERSION "0.0.0")
set(PROJECT_COMPANY "jrengmusic")
set(PROJECT_WEBSITE "https://jrengmusic.com")

set(TARGET_TYPE "CONSOLE")

set(JUCE_MODULES
    juce_core
    juce_events
    juce_data_structures
    juce_graphics
    juce_gui_basics
)

set(SOURCE_MODE "AUTO_GLOB")
set(SOURCE_FILES)

set(MACOS_DEPLOYMENT_TARGET "10.14")
set(MACOS_ARCHITECTURES "x86_64;arm64")
set(BUNDLE_ID_PREFIX "com.${PROJECT_COMPANY}")

# macOS universal binary settings must be set before project() to take effect
if(APPLE)
    set(CMAKE_OSX_ARCHITECTURES "${MACOS_ARCHITECTURES}" CACHE STRING "" FORCE)
    set(CMAKE_OSX_DEPLOYMENT_TARGET "${MACOS_DEPLOYMENT_TARGET}" CACHE STRING "" FORCE)
endif()

# ============================================================================
# JUCE MODULE SOURCE GROUPS
# ============================================================================
set(JUCE_ENABLE_MODULE_SOURCE_GROUPS ON CACHE BOOL "Show JUCE module sources in IDE")

# ============================================================================
# JUCE FRAMEWORK DISCOVERY
# ============================================================================

set(JUCE_FOUND FALSE)

if(DEFINED JUCE_PATH AND EXISTS "${JUCE_PATH}/CMakeLists.txt")
    message(STATUS "JUCE found: ${JUCE_PATH} (CMake variable)")
    add_subdirectory("${JUCE_PATH}" "${CMAKE_BINARY_DIR}/JUCE")
    set(JUCE_FOUND TRUE)

elseif(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/../JUCE/CMakeLists.txt")
    set(JUCE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/../JUCE")
    message(STATUS "JUCE found: ${JUCE_PATH} (sibling directory)")
    add_subdirectory("${JUCE_PATH}" "${CMAKE_BINARY_DIR}/JUCE")
    set(JUCE_FOUND TRUE)

elseif(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/../../JUCE/CMakeLists.txt")
    set(JUCE_PATH "${CMAKE_CURRENT_SOURCE_DIR}/../../JUCE")
    message(STATUS "JUCE found: ${JUCE_PATH} (cousin directory)")
    add_subdirectory("${JUCE_PATH}" "${CMAKE_BINARY_DIR}/JUCE")
    set(JUCE_FOUND TRUE)

elseif(DEFINED ENV{JUCE_PATH} AND EXISTS "$ENV{JUCE_PATH}/CMakeLists.txt")
    set(JUCE_PATH "$ENV{JUCE_PATH}")
    message(STATUS "JUCE found: ${JUCE_PATH} (environment variable)")
    add_subdirectory("${JUCE_PATH}" "${CMAKE_BINARY_DIR}/JUCE")
    set(JUCE_FOUND TRUE)
endif()

if(NOT JUCE_FOUND)
    message(FATAL_ERROR
        "JUCE framework not found!\n"
        "\n"
        "Searched:\n"
        "  1. JUCE_PATH CMake variable: ${JUCE_PATH}\n"
        "  2. ${CMAKE_CURRENT_SOURCE_DIR}/../JUCE/\n"
        "  3. ${CMAKE_CURRENT_SOURCE_DIR}/../../JUCE/\n"
        "  4. JUCE_PATH environment variable\n"
        "\n"
        "Solutions:\n"
        "  - Clone JUCE to ${CMAKE_CURRENT_SOURCE_DIR}/../../JUCE/\n"
        "  - Pass -DJUCE_PATH=/path/to/JUCE\n"
    )
endif()

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# PROJECT DECLARATION
# ============================================================================

if(APPLE)
    set(_EXTRA_LANGS OBJC OBJCXX)
else()
    set(_EXTRA_LANGS "")
endif()

project("${PROJECT_NAME}"
    VERSION ${PROJECT_VERSION}
    LANGUAGES C CXX ${_EXTRA_LANGS}
    DESCRIPTION "TIT-cpp: Terminal Interface for git"
)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

message(STATUS "Project: ${PROJECT_NAME} v${PROJECT_VERSION}")
message(STATUS "Target type: ${TARGET_TYPE}")
message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# JAM MODULE REGISTRATION
# ============================================================================

set(JAM_ROOT "$ENV{HOME}/Documents/Poems/dev/jam" CACHE PATH "Path to jam modules")
add_subdirectory("${JAM_ROOT}" "${CMAKE_BINARY_DIR}/jam")
message(STATUS "JAM modules registered via ${JAM_ROOT}/CMakeLists.txt")
message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# SOURCE FILE ORGANIZATION
# ============================================================================

if(SOURCE_MODE STREQUAL "AUTO_GLOB")
    file(GLOB_RECURSE SOURCES CONFIGURE_DEPENDS
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.cpp"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.c"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.cc"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.m"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.mm"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.h"
        "${CMAKE_CURRENT_SOURCE_DIR}/Source/*.hpp"
    )

    message(STATUS "Source files (auto-discovered):")
    foreach(SRC ${SOURCES})
        file(RELATIVE_PATH REL_PATH "${CMAKE_CURRENT_SOURCE_DIR}" "${SRC}")
        message(STATUS "  ${REL_PATH}")
    endforeach()

else()
    set(SOURCES)
    foreach(SRC ${SOURCE_FILES})
        list(APPEND SOURCES "${CMAKE_CURRENT_SOURCE_DIR}/${SRC}")
    endforeach()

    message(STATUS "Source files (explicit):")
    foreach(SRC ${SOURCE_FILES})
        message(STATUS "  ${SRC}")
    endforeach()
endif()

if(NOT SOURCES)
    message(FATAL_ERROR "No source files found!")
endif()

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# TARGET CREATION
# ============================================================================

set(TARGET_NAME "titc")

juce_add_console_app(${TARGET_NAME}
    PRODUCT_NAME "titc"
    VERSION "${PROJECT_VERSION}"
    COMPANY_NAME "${PROJECT_COMPANY}"
    COMPANY_WEBSITE "${PROJECT_WEBSITE}"
    BUNDLE_ID "${BUNDLE_ID_PREFIX}.titc"
)

target_sources(${TARGET_NAME} PRIVATE ${SOURCES})

juce_add_binary_data(titc_assets
    SOURCES "${CMAKE_CURRENT_SOURCE_DIR}/Assets/tit_logo.svg"
)

target_link_libraries(${TARGET_NAME} PRIVATE titc_assets)

target_include_directories(${TARGET_NAME} PRIVATE
    "${CMAKE_CURRENT_SOURCE_DIR}/Source"
)

message(STATUS "Target created: ${TARGET_NAME}")
message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# SOURCE GROUPS FOR IDE
# ============================================================================

foreach(ABS_SRC ${SOURCES})
    file(RELATIVE_PATH REL_PATH "${CMAKE_CURRENT_SOURCE_DIR}" "${ABS_SRC}")
    get_filename_component(REL_DIR "${REL_PATH}" DIRECTORY)

    if(NOT REL_DIR STREQUAL "")
        source_group("${REL_DIR}" FILES "${ABS_SRC}")
    else()
        source_group("Source Files" FILES "${ABS_SRC}")
    endif()
endforeach()

# ============================================================================
# JUCEHEADER PRE-GENERATION
# ============================================================================

function(_tit_pre_generate_juce_header target)
    get_target_property(juce_library_code ${target} JUCE_GENERATED_SOURCES_DIRECTORY)
    if(juce_library_code)
        file(MAKE_DIRECTORY "${juce_library_code}")

        if(TARGET juce::juceaide)
            get_target_property(juceaide_exe juce::juceaide IMPORTED_LOCATION)

            if(NOT juceaide_exe)
                foreach(config Release Debug RelWithDebInfo MinSizeRel)
                    get_target_property(juceaide_exe juce::juceaide IMPORTED_LOCATION_${config})
                    if(juceaide_exe)
                        break()
                    endif()
                endforeach()
            endif()

            if(NOT juceaide_exe)
                set(juceaide_exe $<TARGET_FILE:juce::juceaide>)
            endif()

            get_target_property(compile_defs ${target} COMPILE_DEFINITIONS)
            set(module_defs "")
            if(compile_defs)
                foreach(def ${compile_defs})
                    if(def MATCHES "^JUCE_MODULE_AVAILABLE_" OR def MATCHES "^DONT_SET_USING_JUCE_NAMESPACE")
                        list(APPEND module_defs "${def}")
                    endif()
                endforeach()
            endif()

            string(ASCII 30 RS)
            string(ASCII 31 US)

            set(defs_file "${juce_library_code}/Defs.txt")
            set(defs_content "MODULE_DEFINITIONS${US}")
            foreach(def ${module_defs})
                string(APPEND defs_content "${def};")
            endforeach()
            string(APPEND defs_content "${RS}")
            file(WRITE "${defs_file}" "${defs_content}")

            set(juce_header "${juce_library_code}/JuceHeader.h")
            if(NOT EXISTS "${juce_header}")
                execute_process(
                    COMMAND "${juceaide_exe}" header "${defs_file}" "${juce_header}"
                    RESULT_VARIABLE result
                    ERROR_VARIABLE error
                    OUTPUT_QUIET
                )

                if(result EQUAL 0)
                    message(STATUS "Pre-generated JuceHeader.h")
                else()
                    message(WARNING "Failed to pre-generate JuceHeader.h: ${error}")
                endif()
            endif()
        else()
            message(WARNING "juce::juceaide not found")
        endif()
    endif()
endfunction()

_tit_pre_generate_juce_header(${TARGET_NAME})
juce_generate_juce_header(${TARGET_NAME})

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# MODULE LINKING
# ============================================================================

message(STATUS "Linking JUCE modules:")

foreach(MODULE ${JUCE_MODULES})
    if(TARGET juce::${MODULE})
        target_link_libraries(${TARGET_NAME} PRIVATE juce::${MODULE})
        message(STATUS "  ${MODULE}")
    else()
        message(FATAL_ERROR "JUCE module not found: ${MODULE}")
    endif()
endforeach()

target_link_libraries(${TARGET_NAME} PRIVATE jam_core)
target_link_libraries(${TARGET_NAME} PRIVATE jam_data_structures)
target_link_libraries(${TARGET_NAME} PRIVATE jam_markdown)
target_link_libraries(${TARGET_NAME} PRIVATE jam_tui)
target_link_libraries(${TARGET_NAME} PRIVATE jam_subprocess)
message(STATUS "  jam_core")
message(STATUS "  jam_data_structures")
message(STATUS "  jam_markdown")
message(STATUS "  jam_tui")
message(STATUS "  jam_subprocess")

target_compile_definitions(${TARGET_NAME} PRIVATE
    JUCE_WEB_BROWSER=0
    JUCE_USE_CURL=0
    DONT_SET_USING_JUCE_NAMESPACE=1
    JUCE_APPLICATION_NAME_STRING="$<TARGET_PROPERTY:${TARGET_NAME},JUCE_PRODUCT_NAME>"
    JUCE_APPLICATION_VERSION_STRING="$<TARGET_PROPERTY:${TARGET_NAME},JUCE_VERSION>"
)

if(COMMAND _juce_fixup_module_source_groups)
    _juce_fixup_module_source_groups()
endif()

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# PLATFORM CONFIGURATION
# ============================================================================

if(APPLE)
    message(STATUS "Platform: macOS")
    message(STATUS "  Deployment: ${CMAKE_OSX_DEPLOYMENT_TARGET}")
    message(STATUS "  Architectures: ${CMAKE_OSX_ARCHITECTURES}")

elseif(WIN32)
    set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
    message(STATUS "Platform: Windows")

elseif(UNIX)
    message(STATUS "Platform: Linux")
endif()

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

# ============================================================================
# COMPILER FLAGS
# ============================================================================

if(APPLE OR UNIX)
    target_compile_options(${TARGET_NAME} PRIVATE
        -Wno-unused-parameter
        -Wno-sign-conversion
        -Wno-shadow
    )
endif()

# ============================================================================
# CLANGD CONFIGURATION (Auto-generated for LSP)
# ============================================================================

set(CLANGD_CONFIG "${CMAKE_CURRENT_SOURCE_DIR}/.clangd")
set(CLANGD_CONTENT "CompileFlags:\n  CompilationDatabase: ${CMAKE_CURRENT_SOURCE_DIR}/Builds/Ninja/\n  Add:\n")

if(APPLE)
    string(APPEND CLANGD_CONTENT "    - -D__APPLE__=1\n")
    string(APPEND CLANGD_CONTENT "    - -DTARGET_OS_IPHONE=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_MAC=1\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_IOS=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_WINDOWS=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_LINUX=0\n")
elseif(WIN32)
    string(APPEND CLANGD_CONTENT "    - -D_WIN32=1\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_WINDOWS=1\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_MAC=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_IOS=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_LINUX=0\n")
elseif(UNIX)
    string(APPEND CLANGD_CONTENT "    - -D__linux__=1\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_LINUX=1\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_MAC=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_IOS=0\n")
    string(APPEND CLANGD_CONTENT "    - -DJUCE_WINDOWS=0\n")
endif()

string(APPEND CLANGD_CONTENT "Diagnostics:\n  MissingIncludes: None\n  UnusedIncludes: None\n")

file(WRITE "${CLANGD_CONFIG}" "${CLANGD_CONTENT}")
message(STATUS "Generated .clangd for LSP")

message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
message(STATUS "Configuration complete!")
message(STATUS "Generator: ${CMAKE_GENERATOR}")
message(STATUS "Build directory: ${CMAKE_BINARY_DIR}")
message(STATUS "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
