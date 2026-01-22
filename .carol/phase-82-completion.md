# Phase 82 Completion Report - Code Audit
## INSPECTOR: OpenCode (CLI Agent)
**Date:** 2026-01-22  
**Audit Scope:** Dead code, SSOT violations, silent failures, architectural issues

---

## EXECUTIVE SUMMARY

**Status:** ⚠️ **CRITICAL ISSUES FOUND**

- **Dead Code:** 2 orphaned files (100+ lines)
- **Unused Functions:** 5 functions (2 UI, 3 dispatchers with TODO)
- **SSOT Violations:** 19 instances (15 dimension access, 4 confirmation dialogs)
- **Legacy Code:** Deprecated constants still in use (confusion risk)
- **Silent Failures:** Clean (only 1 documented suppression)
- **Architecture:** Clean (no major violations)

**Recommendation:** CLEANUP REQUIRED before next feature work

---

## CRITICAL FINDINGS

### 1. ORPHANED FILES (DEAD CODE - HIGH PRIORITY)

**Files:**
- `models.go` (7 lines) - ValidationError struct for API project
- `services.go` (41 lines) - UserService/ApiKeyService interfaces

**Evidence:**
```go
// services.go - completely wrong project!
type UserService interface {
    GetUser(ctx context.Context, id int64) (*User, error)
    CreateUser(ctx context.Context, email, firstName, lastName string) (*User, error)
    // ... API service code
}
```

**Impact:**
- LSP errors shown in editor (confusing developers)
- Files from different project accidentally committed
- NO connection to TIT codebase

**Action Required:**
```bash
rm models.go services.go
```

---

### 2. UNUSED FUNCTIONS (DEAD CODE - MEDIUM PRIORITY)

#### A. UI Wrapper Functions (Never Called)

**File:** `internal/ui/menu.go`

**Functions:**
```go
func RenderMenu(items interface{}) string {
    return RenderMenuWithHeight(items, 0, Theme{}, 24, ContentInnerWidth)
}

func RenderMenuWithSelection(items interface{}, selectedIndex int, theme Theme) string {
    return RenderMenuWithHeight(items, selectedIndex, theme, 24, ContentInnerWidth)
}
```

**Evidence:** Grep shows ZERO calls to these functions in codebase

**Impact:**
- Hardcode magic numbers (24, ContentInnerWidth)
- Misleading API surface (looks like it should be used, but isn't)
- 10 lines of dead code

**Action Required:** Delete both functions

---

#### B. TODO Dispatcher Functions (Stubbed, Never Used)

**File:** `internal/app/dispatchers.go:198-214`

**Functions:**
```go
func (a *Application) dispatchResolveConflicts(app *Application) tea.Cmd {
    // TODO: Implement
    return nil
}

func (a *Application) dispatchAbortOperation(app *Application) tea.Cmd {
    // TODO: Implement
    return nil
}

func (a *Application) dispatchContinueOperation(app *Application) tea.Cmd {
    // TODO: Implement
    return nil
}
```

**Registration:** These are registered in dispatcher map but never triggered:
```go
// dispatchers.go:57-59
"resolve_conflicts":    a.dispatchResolveConflicts,
"abort_operation":      a.dispatchAbortOperation,
"continue_operation":   a.dispatchContinueOperation,
```

**Evidence:** No menu items or handlers call these action names

**Impact:**
- Dispatcher map pollution
- False impression these features exist
- TODO comments left for months

**Action Required:**
1. Delete functions if not planned
2. OR implement if planned and add to SPEC.md

---

### 3. SSOT VIOLATIONS (ARCHITECTURAL - HIGH PRIORITY)

#### A. Direct Constant Access Instead of Dynamic Sizing

**Violation Count:** 15 instances

**Pattern:**
```go
// WRONG - uses static constant
app.confirmationDialog = ui.NewConfirmationDialog(config, ui.ContentInnerWidth, &app.theme)

// CORRECT - uses dynamic sizing
app.confirmationDialog = ui.NewConfirmationDialog(config, a.sizing.ContentInnerWidth, &app.theme)
```

**Locations:**
1. `internal/app/dispatchers.go:191` - dispatchInit (spinner)
2. `internal/app/dispatchers.go:232` - dispatchForcePush (confirmation)
3. `internal/app/dispatchers.go:253` - dispatchRewind (confirmation)
4. `internal/app/dispatchers.go:411` - dispatchShowChangelog (confirmation)
5. `internal/app/dispatchers.go:488` - dispatchShowOperationMenu (spinner)
6. `internal/app/dispatchers.go:512` - dispatchShowBranchSyncMenu (spinner)
7. `internal/app/dispatchers.go:529` - dispatchShowBranchSyncMenu (spinner)
8. `internal/app/handlers.go:1008` - showRewindConfirmation (confirmation)
9. `internal/app/setup_wizard.go:160-161` - renderStepDisplayKey (2 instances)
10. `internal/app/setup_wizard.go:294-295` - renderStepCopyKey (2 instances)
11. `internal/app/setup_wizard.go:340` - renderStepCopyKey
12. `internal/app/setup_wizard.go:346` - renderStepCopyKey
13. `internal/app/confirmation_handlers.go:222` - executeConfirmInit

**Impact:**
- Reactive layout implementation (Session 80) broken for these components
- Confirmation dialogs won't resize with terminal
- Violates DynamicSizing contract established in REACTIVE-LAYOUT-PLAN.md

**Root Cause:**
Code written before reactive layout, not updated during refactor

**Action Required:**
Replace all `ui.ContentInnerWidth` and `ui.ContentHeight` with `a.sizing.ContentInnerWidth` and `a.sizing.ContentHeight`

---

#### B. Legacy Constants Confusion

**File:** `internal/ui/sizing.go:59-63`

```go
// Legacy constants for backward compatibility (deprecated, use DynamicSizing)
const (
    ContentInnerWidth = 76
    ContentHeight     = 24
)
```

**Problem:**
- DynamicSizing struct has SAME field names
- Marked "deprecated" but HEAVILY used (76 references)
- Creates confusion: is `ContentInnerWidth` the constant or the field?

**Example Confusion:**
```go
// Which ContentInnerWidth?
sizing.ContentInnerWidth  // struct field (dynamic) ✓
ui.ContentInnerWidth      // const (static 76)     ✗
```

**Impact:**
- Developers don't know which to use
- Easy to grab wrong one (compiler allows both)
- "Deprecated" label is misleading (can't deprecate what's still used)

**Action Required:**
1. Rename constants to `LegacyContentInnerWidth` / `LegacyContentHeight`
2. OR delete constants and fix all 76 references
3. Update SESSION-LOG rule to forbid `ui.ContentInnerWidth` access

---

### 4. SILENT FAILURE AUDIT RESULTS

**Status:** ✅ **CLEAN**

**Findings:**
- Only 1 error suppression found: `os.Remove(lockPath)` (documented as "best effort")
- Zero `return nil, nil` violations (1 found is valid: no time travel state)
- Zero empty string returns on error
- All error paths properly propagate or panic

**Conclusion:** FAIL-FAST rule is properly followed

---

### 5. MAGIC NUMBER AUDIT

**Findings:**
- `internal/ui/history.go:53` - `listPaneWidth := 24` (acceptable - UI constant)
- `internal/ui/menu.go:16,21` - Hardcoded `24` in dead functions (will be removed)
- `internal/ui/sizing.go:61-62` - Legacy constants `76` and `24` (discussed above)

**Status:** Clean except legacy constants

---

### 6. ARCHITECTURAL PATTERNS AUDIT

**Checked:**
- ✅ Panic usage: Appropriate (fatal errors only)
- ✅ Error handling: Proper propagation
- ✅ SSOT messages: All use `messages.go` (no hardcoded strings)
- ✅ Theme colors: All use `theme.go` (Session 82 fixed last violations)
- ✅ Import cycles: None detected
- ✅ Naming conventions: Follow SESSION-LOG rules

**Status:** Architecture is sound

---

## SUMMARY BY SEVERITY

### 🔴 HIGH PRIORITY (Must fix before next feature)

1. **Delete orphaned files** (models.go, services.go) - 2 minutes
2. **Fix 15 SSOT violations** (ui.ContentInnerWidth → a.sizing.ContentInnerWidth) - 30 minutes
3. **Resolve legacy constants** (rename or migrate) - 60 minutes

**Total effort:** ~90 minutes

### 🟡 MEDIUM PRIORITY (Technical debt)

4. **Delete unused UI functions** (RenderMenu, RenderMenuWithSelection) - 5 minutes
5. **Resolve TODO dispatchers** (delete or implement) - Decision needed

**Total effort:** ~30 minutes + planning

### 🟢 LOW PRIORITY (Quality improvements)

6. **Document component contracts** (when to use which dimension source)
7. **Add linter rule** to prevent `ui.ContentInnerWidth` access from app package

---

## RECOMMENDATIONS

### Immediate Actions (This Session)

```bash
# 1. Delete orphaned files
rm models.go services.go

# 2. Delete unused UI functions
# Edit internal/ui/menu.go - remove RenderMenu and RenderMenuWithSelection

# 3. Fix SSOT violations (15 files)
# Replace ui.ContentInnerWidth → a.sizing.ContentInnerWidth
# Replace ui.ContentHeight → a.sizing.ContentHeight
```

### Architecture Improvements

**Add to SESSION-LOG.md rules:**
```markdown
**DYNAMIC SIZING RULE (CRITICAL)**
- ❌ NEVER access `ui.ContentInnerWidth` or `ui.ContentHeight` from app package
- ✅ ALWAYS use `a.sizing.ContentInnerWidth` and `a.sizing.ContentHeight`
- ✅ These values are reactive and update on terminal resize
- Violation breaks responsive layout contract
```

### Future Work

1. Rename or remove legacy constants (breaking change, needs planning)
2. Implement or delete TODO dispatchers (needs SPEC decision)
3. Add pre-commit hook to catch `ui.ContentInnerWidth` usage

---

## COMPLIANCE CHECKLIST

Per CAROL.md Inspector responsibilities:

- ✅ Read ARCHITECTURE.md and implemented code
- ✅ Verified code against design patterns
- ✅ Checked for SSOT violations (19 found)
- ✅ Checked for silent failures (clean)
- ✅ Checked for dead code (2 files, 5 functions found)
- ✅ Checked for architectural violations (dimension access violations found)
- ✅ Written phase-82-completion.md (this document)
- ⏳ Update ARCHITECTURE.md (after fixes applied)

---

## CONCLUSION

**Audit Result:** ⚠️ **ISSUES FOUND - CLEANUP REQUIRED**

The codebase is architecturally sound but has accumulated technical debt from:
1. Orphaned files from wrong project
2. Incomplete reactive layout migration (Session 80)
3. Dead code from early development

**No critical bugs found** - issues are all maintainability and architecture compliance.

**Estimated cleanup time:** 2-3 hours total

**Recommended approach:**
1. Quick wins first (delete orphaned files, unused functions) - 10 minutes
2. SSOT violations fix (15 instances) - 1 hour
3. Legacy constants resolution - 1-2 hours

---

**INSPECTOR:** OpenCode (CLI Agent)  
**Date:** 2026-01-22  
**Status:** Audit Complete - Awaiting User Decision on Cleanup
