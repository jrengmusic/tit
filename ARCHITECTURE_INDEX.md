# TIT Architecture Documentation Index

## Overview

Complete architectural analysis of the TIT codebase (Terminal Interactive Tool) - a git-focused terminal UI in Go.

**Total Documentation**: 1,828 lines across 3 files  
**Codebase**: ~15,200 LOC across 56 files  
**Generated**: January 22, 2026

---

## Documentation Files

### 1. ARCHITECTURE_ANALYSIS.md (811 lines, 24KB)
**Comprehensive Deep Dive**

Complete architectural documentation covering:

1. **Package Structure & Responsibilities** (1,200 lines)
   - internal/app (26 files, ~6000 LOC)
   - internal/ui (18 files, ~5000 LOC)
   - internal/git (8 files, ~2500 LOC)
   - internal/config (1 file)
   - internal/banner (2 files)

2. **Application State Management**
   - Main Application struct (50 fields)
   - State initialization flow
   - Mode transitions

3. **Application Modes (14 Total)**
   - Mode enumeration with metadata
   - Setup wizard steps (5 steps)
   - Mode-specific behavior

4. **UI Components & Sizing**
   - Major UI components (15+)
   - Dynamic sizing system
   - State header layout

5. **Git State Detection (5-Axis System)**
   - GitEnvironment (Axis 0)
   - Operation (Axis 1)
   - WorkingTree (Axis 2)
   - Remote (Axis 3)
   - Timeline (Axis 4)

6. **Menu System**
   - Menu generation pipeline
   - MenuItem structure
   - MenuItems map (30+ items)
   - Keyboard shortcuts
   - Cache integration

7. **Event Flow & Message Types**
   - Update loop flow
   - Message types (20+)
   - Handler chain

8. **Key Files & Purposes**
   - Detailed file listing
   - Purpose descriptions
   - Line counts

9. **Limitations & Architectural Notes**
   - Known limitations (6)
   - Architectural decisions (6)
   - Design patterns used (6)
   - Thread safety

10. **Complete Mode List**
    - Full mode table
    - Purpose and transitions

**Best For**: Understanding the complete architecture, detailed implementation patterns, state detection logic, and design decisions.

---

### 2. QUICK_REFERENCE.md (486 lines, 14KB)
**Fast Lookup Guide**

Quick reference for developers, organized by topic:

- **AppMode Values** (14 modes with descriptions)
- **Git State Detection** (5-axis system, commands)
- **File Structure** (56 files organized by package)
- **Key Data Structures** (Application, State, MenuItem)
- **Menu Item Categories** (30+ items)
- **Render Functions** (15+ rendering functions)
- **Sizing Constants** (7 constants)
- **Message Types** (20+ message types)
- **Key Handler Registry** (Keyboard shortcuts)
- **Setup Wizard Steps** (5 steps)
- **State Info Maps** (Display info for each state)
- **Cache System** (Caching architecture)
- **Time Travel State File** (Format and usage)
- **Error Handling Pattern**
- **Common Workflows** (Init, Clone, Conflict, Time Travel)

**Best For**: Quick lookups, keyboard shortcuts, state values, function names, and common workflows.

---

### 3. ARCHITECTURE_ANALYSIS_SUMMARY.txt (531 lines, 20KB)
**Executive Summary**

Structured summary of key architectural concepts:

1. **AppModes** (14 values with brief descriptions)
2. **Git State Detection** (5-axis priority order)
3. **Package Structure** (Overview of 5 packages)
4. **Application Struct** (~50 fields, categorized)
5. **Major UI Components** (Layout, content, header components)
6. **Menu System** (Generation pipeline, items, shortcuts)
7. **Git State Detection** (Detailed detection sequence)
8. **Event Flow** (Bubble Tea message handling)
9. **Key Files** (Organized by package, with purposes)
10. **Architectural Decisions** (6 key decisions + patterns)
11. **Current Limitations** (6 known limitations)
12. **Generated Documentation** (This index)

**Best For**: High-level understanding, architectural overview, management presentations.

---

## Quick Navigation

### By Topic

**Application State**
- ARCHITECTURE_ANALYSIS.md → Section 2 (Application State Management)
- QUICK_REFERENCE.md → "Key Data Structures"
- SUMMARY.txt → Section 4 (Application Struct)

**Git State Detection**
- ARCHITECTURE_ANALYSIS.md → Section 5 (Git State Detection)
- QUICK_REFERENCE.md → "Git State Detection (5-Axis Tuple)"
- SUMMARY.txt → Section 2 (Git State Detection)

**AppModes**
- ARCHITECTURE_ANALYSIS.md → Section 3 (Application Modes)
- QUICK_REFERENCE.md → "AppMode Values"
- SUMMARY.txt → Section 1 (Complete AppModes)

**UI Components**
- ARCHITECTURE_ANALYSIS.md → Section 4 (UI Components & Sizing)
- QUICK_REFERENCE.md → "Render Functions"
- SUMMARY.txt → Section 5 (Major UI Components)

**Menu System**
- ARCHITECTURE_ANALYSIS.md → Section 6 (Menu System)
- QUICK_REFERENCE.md → "Menu Item Categories"
- SUMMARY.txt → Section 6 (Menu Generation System)

**Event Flow**
- ARCHITECTURE_ANALYSIS.md → Section 7 (Event Flow)
- QUICK_REFERENCE.md → "Key Handler Registry"
- SUMMARY.txt → Section 8 (Event Flow)

**Files & Purposes**
- ARCHITECTURE_ANALYSIS.md → Section 8 (Key Files)
- QUICK_REFERENCE.md → "File Structure"
- SUMMARY.txt → Section 9 (Key Files)

---

## Key Statistics

### Codebase Size
- **Total LOC**: ~15,200
- **Total Files**: 56
- **Packages**: 5 (app, ui, git, config, banner)

### Package Breakdown
| Package | Files | LOC | Focus |
|---------|-------|-----|-------|
| internal/app | 26 | ~6000 | State, modes, handlers |
| internal/ui | 18 | ~5000 | Rendering, components |
| internal/git | 8 | ~2500 | State detection, commands |
| internal/config | 1 | ~100 | Configuration |
| internal/banner | 2 | ~500 | ASCII art |
| **TOTAL** | **56** | **~15,200** | |

### Application Struct
- **Total Fields**: ~50
- **Organized By**: 14 categories
- **Key Fields**: mode, gitState, menuItems, caches

### AppModes
- **Total Modes**: 14
- **Input Modes**: 12
- **Async Modes**: 2
- **Setup Steps**: 5 (in setup wizard)

### Git State Detection
- **Axes**: 5 (priority-ordered)
- **Operations**: 8 types
- **WorkingTree States**: 2 types
- **Remote States**: 2 types
- **Timeline States**: 4 types (+ N/A)

### Menu Items
- **Total Items**: 30+
- **Categories**: 7 (NotRepo, WorkingTree, Timeline, Remote, History, TimeTravel, Conflict)
- **Shortcuts**: Dynamically registered

### UI Components
- **Render Functions**: 15+
- **Layout Components**: 3
- **Content Modes**: 5
- **Header Components**: 2
- **Utility Components**: 6

---

## Common Tasks

### Finding Information About...

**A Specific Mode**
1. Check QUICK_REFERENCE.md → "AppMode Values"
2. Find brief description
3. Go to ARCHITECTURE_ANALYSIS.md → Section 3 for full details

**State Detection Logic**
1. Check QUICK_REFERENCE.md → "Git State Detection (5-Axis Tuple)"
2. See state flow diagram
3. Go to SUMMARY.txt → Section 7 for detection sequence

**Menu Items**
1. Check QUICK_REFERENCE.md → "Menu Item Categories"
2. Find item ID and purpose
3. Go to ARCHITECTURE_ANALYSIS.md → Section 6 for generation logic

**Keyboard Shortcuts**
1. Check QUICK_REFERENCE.md → "Key Handler Registry"
2. Find mode and key binding
3. Go to ARCHITECTURE_ANALYSIS.md → Section 6 for details

**File Purposes**
1. Check QUICK_REFERENCE.md → "File Structure"
2. Find file and brief purpose
3. Go to SUMMARY.txt → Section 9 for full descriptions

**Data Structures**
1. Check QUICK_REFERENCE.md → "Key Data Structures"
2. See struct definition
3. Go to SUMMARY.txt → Section 4 for field descriptions

**Render Functions**
1. Check QUICK_REFERENCE.md → "Render Functions"
2. Find function name
3. Go to ARCHITECTURE_ANALYSIS.md → Section 4 for details

---

## Architecture Overview (ASCII Art)

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                    MAIN APPLICATION                         │
│                 (internal/app/app.go)                       │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │ Mode State   │  │ Git State    │  │ Menu Items   │     │
│  │ (AppMode)    │  │ (5-Axis)     │  │ (30+ items)  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Event Loop (Bubble Tea)                      │  │
│  │  WindowSize → KeyMsg → Update → View (Render)       │  │
│  └──────────────────────────────────────────────────────┘  │
│                      ↓                                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Handler Registry (Dynamic)                   │  │
│  │  keyHandlers[mode][key] → Handler Function          │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
    ┌────────────┐      ┌──────────┐       ┌──────────────┐
    │ GIT LAYER  │      │ UI LAYER │       │ CONFIG LAYER │
    │(internal/) │      │(internal/│       │(internal/)   │
    │git)        │      │ui)       │       │config)       │
    └────────────┘      └──────────┘       └──────────────┘
        ↓                    ↓                    ↓
   Detect State         Render Views         Persist State
   Execute Cmd          Handle Layout         Load Config
   Track Op             Dynamic Size          Track Stash
   Manage SSH           Color Theme
```

---

## Document Cross-References

### ARCHITECTURE_ANALYSIS.md
- Full reference for: State detection, menu generation, event flow
- Contains: Detailed code structures, type definitions, algorithm explanations
- Use when: Understanding implementation details, design patterns, architectural decisions

### QUICK_REFERENCE.md
- Quick lookup for: Mode values, state constants, function names, shortcuts
- Contains: Lists, tables, code snippets, value enumerations
- Use when: Need specific values, function signatures, or keyboard shortcuts

### ARCHITECTURE_ANALYSIS_SUMMARY.txt
- Executive overview of: Complete system, key decisions, limitations
- Contains: Summaries, organized sections, workflow descriptions
- Use when: Need high-level understanding, presenting to others, or quick review

---

## How to Update This Documentation

When making architectural changes:

1. **Update ARCHITECTURE_ANALYSIS.md** for detailed explanations
2. **Update QUICK_REFERENCE.md** for quick lookup values
3. **Update ARCHITECTURE_ANALYSIS_SUMMARY.txt** for overview impact
4. **Update this INDEX** if file structure changes

---

## Notes

- All documentation current as of January 22, 2026
- Based on complete codebase analysis (56 files, 15,200 LOC)
- Includes all 14 AppModes and 30+ menu items
- Covers 5-axis git state detection system
- Documents current limitations and architectural decisions

---

**Generated**: January 22, 2026  
**For**: Architecture Documentation  
**Updated**: As needed when codebase changes
