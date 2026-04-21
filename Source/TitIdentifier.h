#pragma once
#include <juce_core/juce_core.h>

namespace ID
{
    // ---- A. VT root + top-level nodes (RFC §4.6) ----
    const juce::Identifier TIT          { "TIT" };
    const juce::Identifier ENV          { "ENV" };
    const juce::Identifier REPO         { "REPO" };
    const juce::Identifier HISTORY      { "HISTORY" };
    const juce::Identifier FILES        { "FILES" };
    const juce::Identifier DIFF         { "DIFF" };
    const juce::Identifier MENU         { "MENU" };
    const juce::Identifier CONSOLE      { "CONSOLE" };
    const juce::Identifier SELECTION    { "SELECTION" };
    const juce::Identifier THEME        { "THEME" };
    const juce::Identifier SETUP        { "SETUP" };

    // ---- B. Repeating child record nodes (RFC §4.6) ----
    const juce::Identifier COMMIT       { "COMMIT" };
    const juce::Identifier FILE         { "FILE" };
    const juce::Identifier HUNK         { "HUNK" };
    const juce::Identifier ITEM         { "ITEM" };
    const juce::Identifier LINE         { "LINE" };

    // ---- C. ENV properties (RFC §4.6) ----
    const juce::Identifier gitAvailable    { "gitAvailable" };
    const juce::Identifier sshAvailable    { "sshAvailable" };
    const juce::Identifier sshKeysPresent  { "sshKeysPresent" };
    const juce::Identifier setupState      { "setupState" };

    // ---- D. REPO properties (RFC §4.6) ----
    const juce::Identifier workingTree     { "workingTree" };
    const juce::Identifier timeline        { "timeline" };
    const juce::Identifier operation       { "operation" };
    const juce::Identifier remote          { "remote" };
    const juce::Identifier isTitTimeTravel { "isTitTimeTravel" };
    const juce::Identifier branch          { "branch" };
    const juce::Identifier aheadCount      { "aheadCount" };
    const juce::Identifier behindCount     { "behindCount" };
    const juce::Identifier cwd             { "cwd" };

    // ---- E. HISTORY / COMMIT properties ----
    const juce::Identifier hash            { "hash" };
    const juce::Identifier author          { "author" };
    const juce::Identifier date            { "date" };
    const juce::Identifier message         { "message" };

    // ---- F. FILES / FILE properties ----
    const juce::Identifier path            { "path" };
    const juce::Identifier status          { "status" };

    // ---- G. DIFF / HUNK properties ----
    const juce::Identifier oldStart        { "oldStart" };
    const juce::Identifier newStart        { "newStart" };
    const juce::Identifier lines           { "lines" };

    // ---- H. MENU / ITEM properties ----
    const juce::Identifier id              { "id" };
    const juce::Identifier label           { "label" };
    const juce::Identifier hotkey          { "hotkey" };
    const juce::Identifier enabled         { "enabled" };
    const juce::Identifier destructive     { "destructive" };

    // ---- I. CONSOLE / LINE properties ----
    const juce::Identifier text            { "text" };
    const juce::Identifier stream          { "stream" };

    // ---- J. SELECTION properties ----
    const juce::Identifier menuIndex       { "menuIndex" };
    const juce::Identifier historyIndex    { "historyIndex" };
    const juce::Identifier fileIndex       { "fileIndex" };
    const juce::Identifier activePane      { "activePane" };

    // ---- K. SETUP properties (RFC §4.6) ----
    const juce::Identifier phase           { "phase" };
    const juce::Identifier email           { "email" };
    const juce::Identifier publicKey       { "publicKey" };

    // ---- K2. DIALOG variant property (Step 3.6 ConfirmDialog dispatch) ----
    const juce::Identifier kind            { "kind" };

    // ---- L. THEME hierarchy — component-family nodes (themes/default.xml) ----
    // THEME already declared at A; MENU already declared at A.
    const juce::Identifier LOOK_AND_FEEL       { "LOOK_AND_FEEL" };
    const juce::Identifier SCREEN              { "SCREEN" };
    const juce::Identifier TEXT                { "TEXT" };
    const juce::Identifier BORDER              { "BORDER" };
    const juce::Identifier DIALOG              { "DIALOG" };
    const juce::Identifier CONFLICT_RESOLVER   { "CONFLICT_RESOLVER" };
    const juce::Identifier STATUS              { "STATUS" };
    const juce::Identifier TIMELINE            { "TIMELINE" };
    const juce::Identifier OPERATION           { "OPERATION" };
    const juce::Identifier SPINNER             { "SPINNER" };
    // DIFF already declared at A.
    const juce::Identifier COPY_HASH_LABEL     { "COPY_HASH_LABEL" };
    const juce::Identifier CONSOLE_STREAM      { "CONSOLE_STREAM" };

    // ---- M. THEME — color attribute names (44 total, per themes/default.xml) ----
    // SCREEN attributes
    const juce::Identifier mainBackgroundColor         { "mainBackgroundColor" };
    const juce::Identifier inlineBackgroundColor       { "inlineBackgroundColor" };
    const juce::Identifier selectionBackgroundColor    { "selectionBackgroundColor" };

    // TEXT attributes
    const juce::Identifier contentTextColor            { "contentTextColor" };
    const juce::Identifier labelTextColor              { "labelTextColor" };
    const juce::Identifier dimmedTextColor             { "dimmedTextColor" };
    const juce::Identifier accentTextColor             { "accentTextColor" };
    const juce::Identifier highlightTextColor          { "highlightTextColor" };
    const juce::Identifier terminalTextColor           { "terminalTextColor" };
    const juce::Identifier cwdTextColor                { "cwdTextColor" };
    const juce::Identifier footerTextColor             { "footerTextColor" };

    // BORDER attributes
    const juce::Identifier boxBorderColor              { "boxBorderColor" };
    const juce::Identifier separatorColor              { "separatorColor" };

    // DIALOG attributes
    const juce::Identifier confirmationDialogBackground { "confirmationDialogBackground" };
    const juce::Identifier buttonSelectedTextColor      { "buttonSelectedTextColor" };

    // CONFLICT_RESOLVER attributes
    const juce::Identifier conflictPaneUnfocusedBorder  { "conflictPaneUnfocusedBorder" };
    const juce::Identifier conflictPaneFocusedBorder    { "conflictPaneFocusedBorder" };
    const juce::Identifier conflictSelectionForeground  { "conflictSelectionForeground" };
    const juce::Identifier conflictSelectionBackground  { "conflictSelectionBackground" };
    const juce::Identifier conflictPaneTitleColor       { "conflictPaneTitleColor" };

    // STATUS attributes
    const juce::Identifier statusClean                 { "statusClean" };
    const juce::Identifier statusDirty                 { "statusDirty" };

    // TIMELINE attributes
    const juce::Identifier timelineSynchronized        { "timelineSynchronized" };
    const juce::Identifier timelineLocalAhead          { "timelineLocalAhead" };
    const juce::Identifier timelineLocalBehind         { "timelineLocalBehind" };

    // OPERATION attributes
    const juce::Identifier operationReady              { "operationReady" };
    const juce::Identifier operationNotRepo            { "operationNotRepo" };
    const juce::Identifier operationTimeTravel         { "operationTimeTravel" };
    const juce::Identifier operationConflicted         { "operationConflicted" };
    const juce::Identifier operationMerging            { "operationMerging" };
    const juce::Identifier operationRebasing           { "operationRebasing" };
    const juce::Identifier operationDirtyOp            { "operationDirtyOp" };

    // MENU attributes
    const juce::Identifier menuSelectionBackground     { "menuSelectionBackground" };

    // SPINNER attributes
    const juce::Identifier spinnerColor                { "spinnerColor" };

    // DIFF attributes
    const juce::Identifier diffAddedLineColor          { "diffAddedLineColor" };
    const juce::Identifier diffRemovedLineColor        { "diffRemovedLineColor" };

    // COPY_HASH_LABEL attributes
    const juce::Identifier copyHashLabelForeground     { "copyHashLabelForeground" };
    const juce::Identifier copyHashLabelBackground     { "copyHashLabelBackground" };

    // CONSOLE_STREAM attributes
    const juce::Identifier outputStdoutColor           { "outputStdoutColor" };
    const juce::Identifier outputStderrColor           { "outputStderrColor" };
    const juce::Identifier outputStatusColor           { "outputStatusColor" };
    const juce::Identifier outputWarningColor          { "outputWarningColor" };
    const juce::Identifier outputDebugColor            { "outputDebugColor" };
    const juce::Identifier outputInfoColor             { "outputInfoColor" };

    // ---- N. THEME metadata attributes (on root <THEME>) ----
    // name / description declared here; version shares string "version"
    const juce::Identifier name                        { "name" };
    const juce::Identifier description                 { "description" };
    const juce::Identifier version                     { "version" };

} // namespace ID
