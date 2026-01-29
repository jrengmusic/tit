Refactor codebase architecture and implement LIFESTAR compliance improvements

This comprehensive update spans three completed sprints addressing critical architectural issues and LIFESTAR compliance violations identified through systematic auditing.

## Sprint 3: LIFESTAR Compliance and Auditing (2026-01-30)
**Agents:** Amp (Claude Sonnet 4) - AUDITOR, glm-4.6 (zai-coding-plan/glm-4.6) - COUNSELOR
- AUDITOR conducted comprehensive LIFESTAR compliance audit identifying 4 critical refactoring opportunities and 2 anti-patterns
- COUNSELOR created detailed kickoff plan addressing all high-priority findings with specific implementation phases

**Key Changes:**
- Added DetectionWarnings field to State struct for transparent error reporting
- Replaced panic in FindStashRefByHash() with graceful error handling
- Added TimelineConfidence enum to track detection reliability
- Injected OutputBuffer via Application constructor for improved testability

## Sprint 2: Code Refactoring and Optimization (2026-01-30)
**Agents:** glm-4.7 (zai-coding-plan/glm-4.7) - ENGINEER, glm-4.6 (zai-coding-plan/glm-4.6) - COUNSELOR
- Major refactoring to reduce file sizes and improve maintainability
- ENGINEER executed multiple phases extracting core Bubble Tea methods and splitting large handler files

**Key Changes:**
- app.go: Reduced from 1,771 to 392 lines (78% reduction)
- Created focused modules: app_init.go, app_update.go, app_view.go, app_constructor.go, app_keys.go
- Split confirmation_handlers.go into confirm_dialog.go and confirm_handlers.go
- Removed 41 unused delegation methods (171 lines)

## Sprint 1: Full Codebase Audit (2026-01-29)
**Agents:** Amp (Claude Sonnet 4) - AUDITOR
- Comprehensive audit of entire codebase identifying structural issues and LIFESTAR violations

**Key Changes:**
- Removed duplicate type definitions (CommitInfo, FileInfo)
- Identified Application struct as God Object (93 fields)
- Found hardcoded strings and magic numbers
- Cleaned up backup and temporary files

**Architecture Updates:**
- Updated ARCHITECTURE.md to reflect new file organization
- Documented reduced Application struct (31 fields from original 47)
- Removed references to deleted delegation methods
- Added comprehensive file organization section
- Updated key handler registration documentation

**Files Modified:**
- SPRINT-LOG.md - Added sprint entries and updated role registrations
- ARCHITECTURE.md - Updated to reflect current architecture post-refactoring
- .carol/ - Cleaned up compiled sprint summary files
- internal/app/ - Major refactoring with new file structure
- internal/git/ - LIFESTAR compliance improvements
- internal/ui/ - Buffer injection for testability

This work establishes a solid foundation following LIFESTAR principles with improved error handling, reduced complexity, and enhanced maintainability while preserving all existing functionality.