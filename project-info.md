# Project Info

## index

+--------------+----------------------------------------------+
| alias        | symbol                                       |
+==============+==============================================+
| @char        | const char* const                            |
| @juce-path   | ${CMAKE_CURRENT_SOURCE_DIR}/../../JUCE       |
| @user-module | ${CMAKE_CURRENT_SOURCE_DIR}/../jam           |
| @source      | ${CMAKE_CURRENT_SOURCE_DIR}/Source           |
| @generated   | ${CMAKE_CURRENT_SOURCE_DIR}/Source/generated |
+--------------+----------------------------------------------+

## project info

```
@brief Project metadata — the ProjectInfo namespace, generated.

Every field is a complete literal; nothing downstream derives, concatenates, or restates a value.
```

+------------------+-------+--------------------------------------------------+-----------+--------------------------------------------------+
| name             | type  | value                                            | format    | description                                      |
+==================+=======+==================================================+===========+==================================================+
| projectName      | @char | titc                                             | toLiteral | Terminal Interface for git                       |
| companyName      | @char | JRENG                                            | toLiteral | Company name.                                    |
| legalCompanyName | @char | Jubilant Research of Eclectic Novelty Generation | toLiteral | Full legal company name.                         |
| versionString    | @char | 0.0.0                                            | toLiteral | Product version string.                          |
| versionNumber    | int   | 0x0                                              |           | Product version, JUCE hex encoding.              |
| productWebsite   | @char | `https://jrengmusic.com`                         |           | Product website URL.                             |
| companyEmail     | @char | info@jrengmusic.com                              | toLiteral | Company contact email.                           |
| presetExtension  | @char | titc                                             | toLiteral | Preset file extension, without the leading dot.  |
| presetDefault    | @char | INIT                                             | toLiteral | Default init preset name, without the extension. |
+------------------+-------+--------------------------------------------------+-----------+--------------------------------------------------+

## cmake

+-----------------------------+-----------------------------------------------+------------------------------------------+
| key                         | value                                         | description                              |
+=============================+===============================================+==========================================+
| projectDescription          | Terminal Interface for git                    | Project description, single line         |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| banner                      | titc                                          | Banner artwork, printed at configuretime |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| minimumVersion              | 3.25                                          | CMake minimum version                    |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| juceTargetFunction          | juce_add_console_app                          | JUCE target-creation function            |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| juceVersion                 | 8.0.17                                        | Required JUCE version, exact             |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| cxxStandard                 | 17                                            | C++ language standard                    |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| deploymentTarget            | `11.0`                                        | Minimum macOS deployment target          |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| architecture                | x86_64;arm64                                  | macOS architectures                      |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| msvcRuntime                 | `MultiThreaded$<$<CONFIG:Debug>:Debug>`       | MSVC runtime library selection           |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| jucePath                    | @juce-path                                    | JUCE root                                |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| userModulePath              | @user-module                                  | User module root                         |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| installDirectory            | $ENV{HOME}/.local/bin                         | Installed binary directory               |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| bundleIdentifier            | com.jrengmusic.titc                           | JUCE BUNDLE_ID                           |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| targetName                  | titc                                          | CMake target name                        |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| companyCopyright            | © 2026 JRENG                                  | JUCE COMPANY_COPYRIGHT                   |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| productName                 | titc                                          | JUCE PRODUCT_NAME                        |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| buildVersion                | 0.0.0                                         | JUCE BUILD_VERSION                       |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| documentExtensions          | titc                                          | JUCE DOCUMENT_EXTENSIONS                 |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| binaryDataNamespace         | BinaryData                                    | juce_add_binary_data NAMESPACE           |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| needsCurl                   | OFF                                           | JUCE NEEDS_CURL                          |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| needsWebBrowser             | OFF                                           | JUCE NEEDS_WEB_BROWSER                   |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| needsWebview2               | OFF                                           | JUCE NEEDS_WEBVIEW2                      |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| needsStoreKit               | OFF                                           | JUCE NEEDS_STORE_KIT                     |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| interproceduralOptimization | ON                                            | CMAKE_INTERPROCEDURAL_OPTIMIZATION       |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| compileJobMemory            | 1280                                          | Compile pool RAM per job, MiB            |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| precompiledHeader           | JuceHeader.h                                  | Project-source precompiled header        |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| qaDirectory                 | `$ENV{HOME}/Documents/Poems/dev/___builds___` | QA build archive root                    |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| repository                  | jrengmusic/tit                                | GitHub repository of the release         |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| installerResourceDirectory  | ${CMAKE_SOURCE_DIR}/cast/installer/resources  | Installer artwork directory              |
+-----------------------------+-----------------------------------------------+------------------------------------------+
| installDirectoryWindows     | $PROFILE\\.local\\bin                         | Windows installer default directory      |
+-----------------------------+-----------------------------------------------+------------------------------------------+

## juce module

+-----------------+-----------------------+
| name            | value                 |
+=================+=======================+
| core            | juce_core             |
| dataStructures  | juce_data_structures  |
| events          | juce_events           |
| graphics        | juce_graphics         |
| guiBasics       | juce_gui_basics       |
| audioProcessors | juce_audio_processors |
+-----------------+-----------------------+

## user module

+--------------+-----------------------+--------------------------------------------------------------------------------------------------+
| root         | name                  | description                                                                                      |
+==============+=======================+==================================================================================================+
| @user-module | jam_core              | JAM Core                                                                                         |
| @user-module | jam_document          | Universal line break and reflow (UAX #14)                                                        |
| @user-module | jam_markdown          | Clean-room native CommonMark + GFM markdown parsing and rendering                                |
| @user-module | jam_freetype          | Vendored FreeType font rasterization library                                                     |
| @user-module | jam_graphics          | Graphics utilities, blur, shadows, colours, fonts, mesh                                          |
| @user-module | jam_data_structures   | ValueTree management and data model utilities — model, parameters                                |
| @user-module | jam_web               | HTML authored-subset and CSS Syntax Level 3 subset tokenizers and parsers                        |
| @user-module | jam_style             | JAM Style — LookAndFeel base + ColourScheme-backed colour registry                               |
| @user-module | jam_gui               | GUI foundation — Window, Modal, Glass                                                            |
| @user-module | jam_mermaid_diagram   | Clean-room native mermaid diagram parsing over jam::Document producing the semantic Element tree |
| @user-module | jam_markdown_graphics | Markdown and mermaid materialisation, layout and view component                                  |
| @user-module | jam_terminal          | VT read-side engine, TTY, protocol vocabulary, keyboard encoding                                 |
| @user-module | jam_tui               | Terminal user interface components, painted through the terminal graphics context                |
+--------------+-----------------------+--------------------------------------------------------------------------------------------------+

## program module

+-------------+--------------+
| name        | value        |
+=============+==============+
| core        | juce_core    |
| events      | juce_events  |
| jamCore     | jam_core     |
| jamTerminal | jam_terminal |
+-------------+--------------+

## program user module

+------+------+-------------+
| root | name | description |
+======+======+=============+
+------+------+-------------+

## source glob

+---------+-----------+-------------------------------------+
| path    | extension | description                         |
+=========+===========+=====================================+
| @source | cpp       | Source .cpp files                   |
| @source | h         | Source headers, including generated |
+---------+-----------+-------------------------------------+

## macro

+------------------------------------+----------------------------------------------------------------------+---------------------------------------------------------------+
| name                               | value                                                                | description                                                   |
+====================================+======================================================================+===============================================================+
| useJuceNamespace                   | DONT_SET_USING_JUCE_NAMESPACE=1                                      | No using namespace juce in JuceHeader.h                       |
| declareProjectInfo                 | JUCE_DONT_DECLARE_PROJECTINFO=1                                      | No auto-generated ProjectInfo namespace                       |
| juceStrictRefCountedPointer        | JUCE_STRICT_REFCOUNTEDPOINTER=1                                      | Strict ReferenceCountedObjectPtr casting                      |
| juceWebBrowser                     | JUCE_WEB_BROWSER=0                                                   | WebBrowserComponent support                                   |
| juceUseCurl                        | JUCE_USE_CURL=0                                                      | libcurl for http on Linux                                     |
| juceUseFoundation                  | JUCE_USE_FOUNDATION=0                                                | POSIX file and system functions on macOS, no Foundation       |
| juceUseResilientMapping            | JUCE_USE_RESILIENT_MAPPING=1                                         | A read-only map of a file that becomes shorter gives no fault |
| win32LeanAndMean                   | `$<$<PLATFORM_ID:Windows>:WIN32_LEAN_AND_MEAN=1>`                    | Exclude rarely used Windows headers                           |
| nominmax                           | `$<$<PLATFORM_ID:Windows>:NOMINMAX=1>`                               | No Windows min/max macros                                     |
| win32Ie                            | `$<$<PLATFORM_ID:Windows>:_WIN32_IE=0x0A00>`                         | Windows IE platform version                                   |
| juceCoreIncludeNativeHeaders       | `$<$<PLATFORM_ID:Windows>:JUCE_CORE_INCLUDE_NATIVE_HEADERS=1>`       | Native platform headers in juce_core                          |
| juceGraphicsIncludeDirect2dHelpers | `$<$<PLATFORM_ID:Windows>:JUCE_GRAPHICS_INCLUDE_DIRECT2D_HELPERS=1>` | Direct2D helper visibility, from the framework patch          |
| juceCoreIncludeComSmartPtr         | `$<$<PLATFORM_ID:Windows>:JUCE_CORE_INCLUDE_COM_SMART_PTR=1>`        | Windows COM smart pointer helpers                             |
| haveFreetype                       | HAVE_FREETYPE                                                        | FreeType font backend, consumed by juce_graphics_Harfbuzz.cpp |
| haveUnistdH                        | `$<$<NOT:$<PLATFORM_ID:Windows>>:HAVE_UNISTD_H>`                     | POSIX unistd.h present, non-Windows                           |
+------------------------------------+----------------------------------------------------------------------+---------------------------------------------------------------+

## include

+------------+----------------------------------------------------------+-------------------------------------------+
| name       | value                                                    | description                               |
+============+==========================================================+===========================================+
| userModule | @user-module                                             | User module root, for #include resolution |
| generated  | @generated                                               | Generated headers                         |
| freetype   | `${CAST_USER_MODULE_PATH}/jam_freetype/freetype/include` |                                           |
| harfbuzz   | `${CAST_JUCE_PATH}/modules/juce_graphics/fonts/harfbuzz` |                                           |
+------------+----------------------------------------------------------+-------------------------------------------+

## help

+------------+-------------------------------+------------------------------------+
| name       | value                         | description                        |
+============+===============================+====================================+
| help       | Source/HELP.md                | Help document source               |
| styleSheet | ../jam/jam_tui/page/style.css | Stylesheet, embedded as BinaryData |
+------------+-------------------------------+------------------------------------+

## binary

+------+-------+-------------+
| name | value | description |
+======+=======+=============+
+------+-------+-------------+

## library

+------+-------+-------------+
| name | value | description |
+======+=======+=============+
+------+-------+-------------+

## toolchain

+----------+---------+----------------------------------------------------------------------------+
| argument | command | flag                                                                       |
+==========+=========+============================================================================+
|          | cast    | ../jam/cast/spell.md                                                       |
+----------+---------+----------------------------------------------------------------------------+
|          | cmake   | -S . -B Builds/Release -G Ninja -DCMAKE_BUILD_TYPE=Release -DCAST_SIGN=ON  |
+----------+---------+----------------------------------------------------------------------------+
|          | ninja   | -C Builds/Release                                                          |
+----------+---------+----------------------------------------------------------------------------+
|          | cmake   | -P gh.cmake                                                                |
+----------+---------+----------------------------------------------------------------------------+
| debug    | cast    | ../jam/cast/spell.md                                                       |
+----------+---------+----------------------------------------------------------------------------+
| debug    | cmake   | -S . -B Builds/Debug -G Ninja -DCMAKE_BUILD_TYPE=Debug                     |
+----------+---------+----------------------------------------------------------------------------+
| debug    | ninja   | -C Builds/Debug                                                            |
+----------+---------+----------------------------------------------------------------------------+
| no-sign  | cast    | ../jam/cast/spell.md                                                       |
+----------+---------+----------------------------------------------------------------------------+
| no-sign  | cmake   | -S . -B Builds/Release -G Ninja -DCMAKE_BUILD_TYPE=Release -DCAST_SIGN=OFF |
+----------+---------+----------------------------------------------------------------------------+
| no-sign  | ninja   | -C Builds/Release                                                          |
+----------+---------+----------------------------------------------------------------------------+

## pack

+----------+------------+------------+-----------------------------------+
| name     | mac        | win        | description                       |
+==========+============+============+===================================+
| folder   | ../Release | ../Release | Archive folder                    |
| format   | pkg        | exe        | Archive format                    |
| platform | macOS      | Windows    | Platform word in the archive name |
+----------+------------+------------+-----------------------------------+

## release notes

+------+
| note |
+======+
+------+

## release

+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| name                    | mac                            | win            | stage  | description                              |
+=========================+================================+================+========+==========================================+
| shadow                  | -Wno-shadow                    | /wd4456        |        | Lambda captures / declarations mayshadow |
|                         |                                |                |        |  intentionally                           |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| unusedParameter         | -Wno-unused-parameter          | /wd4100        |        | Debug/template code may not use          |
|                         |                                |                |        |  allparameters                           |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| floatEqual              | -Wno-float-equal               |                |        | DSP exact float comparisons for          |
|                         |                                |                |        |  bypassdetection                         |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| signConversion          | -Wno-sign-conversion           |                |        | Array indexing, safe in this context     |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| switchEnum              | -Wno-switch-enum               |                |        | Not every filter type needs every        |
|                         |                                |                |        |  casehandled                             |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| floatToDoubleConversion | -Wno-implicit-float-conversion | /wd4244        |        | DSP double/float conversions             |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| permissiveMinus         |                                | /permissive-   |        | Standards conformance matching clang —   |
|                         |                                |                |        | Function::Map identity-cast              |
|                         |                                |                |        |  deductionneeds it                       |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| rvalueCast              |                                | /Zc:rvalueCast |        | Off by default, not implied by /         |
|                         |                                |                |        | permissive- — needed for correct         |
|                         |                                |                |        |  Tdeduction                              |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| warningLevel4           |                                | /W4            |        | Warning level 4                          |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| fullPathInPdb           |                                | /FC            |        | Full path in PDB, required for           |
|                         |                                |                |        |  debuggersource-line resolution          |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| optimization            | -O3                            | /O2            |        | Full optimization                        |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| linkTimeOptimization    | -flto=thin                     | /GL            |        | Link-time optimization codegen           |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| linkTimeCodegen         |                                | /LTCG          | linker | MSVC whole-program link-time codegen     |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| deadCodeStripping       | -dead_strip                    | /OPT:REF       | linker | Strip unreferenced functions and data    |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| deadDylibStripping      | -Wl,-dead_strip_dylibs         |                | linker | Drop the libraries that no live code     |
|                         |                                |                |        |  references.                             |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| exportedSymbols         | -Wl,-exported_symbol,_main     |                | linker | Export the entry point only.             |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+
| identicalCodeFolding    |                                | /OPT:ICF       | linker | Fold identical COMDATs                   |
+-------------------------+--------------------------------+----------------+--------+------------------------------------------+

## debug

+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| name                    | mac                            | win            | stage | description                              |
+=========================+================================+================+=======+==========================================+
| shadow                  | -Wno-shadow                    | /wd4456        |       | Lambda captures / declarations mayshadow |
|                         |                                |                |       |  intentionally                           |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| unusedParameter         | -Wno-unused-parameter          | /wd4100        |       | Debug/template code may not use          |
|                         |                                |                |       |  allparameters                           |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| floatEqual              | -Wno-float-equal               |                |       | DSP exact float comparisons for          |
|                         |                                |                |       |  bypassdetection                         |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| signConversion          | -Wno-sign-conversion           |                |       | Array indexing, safe in this context     |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| switchEnum              | -Wno-switch-enum               |                |       | Not every filter type needs every        |
|                         |                                |                |       |  casehandled                             |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| floatToDoubleConversion | -Wno-implicit-float-conversion | /wd4244        |       | DSP double/float conversions             |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| permissiveMinus         |                                | /permissive-   |       | Standards conformance matching clang —   |
|                         |                                |                |       | Function::Map identity-cast              |
|                         |                                |                |       |  deductionneeds it                       |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| rvalueCast              |                                | /Zc:rvalueCast |       | Off by default, not implied by /         |
|                         |                                |                |       | permissive- — needed for correct         |
|                         |                                |                |       |  Tdeduction                              |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| warningLevel4           |                                | /W4            |       | Warning level 4                          |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| fullPathInPdb           |                                | /FC            |       | Full path in PDB, required for           |
|                         |                                |                |       |  debuggersource-line resolution          |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| optimization            | -O0                            | /Od            |       | No optimization                          |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
| debugSymbols            | -g                             | /Zi            |       | Debug symbols                            |
+-------------------------+--------------------------------+----------------+-------+------------------------------------------+
