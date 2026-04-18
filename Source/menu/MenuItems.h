#pragma once
#include <JuceHeader.h>

// ============================================================================
// MenuItems — SSOT for all 27 menu item definitions
// ============================================================================
//
// IDs, labels, and hotkey chars trace verbatim to Go
// ___legacy___/internal/app/menu_items.go.
//
// Each field:
//   id          — matches Go MenuItem.ID field
//   label       — matches Go MenuItem.Label field
//   hotkey      — matches Go MenuItem.Shortcut field (single char only;
//                 multi-char bindings like "ctrl+r" or "]" stored as primary
//                 char; display logic in the view layer)
//   destructive — true for items whose hint begins with "DESTRUCTIVE"

namespace tit::menu
{

struct MenuItemDef
{
    const char*  id;
    const char*  label;
    juce::juce_wchar hotkey;     // 0 if none or multi-char binding
    bool         destructive;
};

// ---- NotRepo ---------------------------------------------------------------

inline constexpr MenuItemDef INIT  { "init",  "Initialize repository", 'i', false };
inline constexpr MenuItemDef CLONE { "clone", "Clone repository",      'c', false };

// ---- Working tree (Normal, Dirty) ------------------------------------------

inline constexpr MenuItemDef COMMIT       { "commit",      "Commit changes",    'c', false };
inline constexpr MenuItemDef COMMIT_PUSH  { "commit_push", "Commit and push",   'p', false };
inline constexpr MenuItemDef RESET_DISCARD_CHANGES { "reset_discard_changes",
                                                     "Discard all changes",     0,   true  };

// ---- Timeline: Ahead -------------------------------------------------------

inline constexpr MenuItemDef PUSH        { "push",       "Push to remote", ']', false };
inline constexpr MenuItemDef FORCE_PUSH  { "force_push", "Force push",      0,  true  };

// ---- Timeline: Behind ------------------------------------------------------

inline constexpr MenuItemDef DIRTY_PULL_MERGE { "dirty_pull_merge",
                                                "Pull (save changes)",     0,   false };
inline constexpr MenuItemDef PULL_MERGE       { "pull_merge",
                                                "Pull (fetch + merge)",   '[',  false };
inline constexpr MenuItemDef REPLACE_LOCAL    { "replace_local",
                                                "Replace local",          'x',  true  };

// ---- Timeline: Diverged ----------------------------------------------------

inline constexpr MenuItemDef PULL_MERGE_DIVERGED { "pull_merge_diverged",
                                                   "Pull (merge)",        '[',  false };
inline constexpr MenuItemDef PUSH_AUTO_SYNC      { "push_auto_sync",
                                                   "Push (auto sync)",    ']',  false };

// ---- History ---------------------------------------------------------------

inline constexpr MenuItemDef HISTORY      { "history",      "History",       'h', false };
inline constexpr MenuItemDef FILE_HISTORY { "file_history", "File(s) history", 'f', false };

// ---- Remote ----------------------------------------------------------------

inline constexpr MenuItemDef ADD_REMOTE { "add_remote", "Add remote", 'r', false };

// ---- TimeTraveling ---------------------------------------------------------

inline constexpr MenuItemDef TIME_TRAVEL_HISTORY       { "time_travel_history",
                                                         "History",       'h', false };
inline constexpr MenuItemDef TIME_TRAVEL_FILES_HISTORY { "time_travel_files_history",
                                                         "File(s) history", 'g', false };
inline constexpr MenuItemDef TIME_TRAVEL_MERGE         { "time_travel_merge",
                                                         "Merge back",    'm', false };
inline constexpr MenuItemDef TIME_TRAVEL_RETURN        { "time_travel_return",
                                                         "Return",        'r', false };

// ---- Init / Clone location (sub-menus, counted as part of NotRepo scope) ---

inline constexpr MenuItemDef INIT_HERE    { "init_here",   "Initialize directory", '1', false };
inline constexpr MenuItemDef INIT_SUBDIR  { "init_subdir", "Create subdirectory",  '2', false };
inline constexpr MenuItemDef CLONE_HERE   { "clone_here",  "Clone to directory",   '1', false };
inline constexpr MenuItemDef CLONE_SUBDIR { "clone_subdir","Create subdirectory",  '2', false };

// ---- Mid-operation recovery ------------------------------------------------

inline constexpr MenuItemDef FINALIZE_MERGE  { "finalize_merge",  "Finalize merge",  'f', false };
inline constexpr MenuItemDef ABORT_MERGE     { "abort_merge",     "Abort merge",     'a', false };
inline constexpr MenuItemDef REBASE_CONTINUE { "rebase_continue", "Continue rebase", 'c', false };
inline constexpr MenuItemDef REBASE_ABORT    { "rebase_abort",    "Abort rebase",    'a', false };

} // namespace tit::menu
