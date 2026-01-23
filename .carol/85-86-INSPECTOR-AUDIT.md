# Session 85-86 Task Summary

**Role:** INSPECTOR
**Agent:** GPT-5.1-Codex-Max (Droid)
**Date:** 2026-01-24
**Task:** Audit timeline sync + config/preferences implementation and provide fixes for CARETAKER

## Objective
Audited the Timeline Sync (Session 85) and Config/Preferences (Session 86) implementations for SSOT, fail-fast, naming, and completeness; produced remediation plan for CARETAKER.

## Findings
- Config SSOT drift: `config.Load()` defaults mismatch (`default` vs `gfx`), silent fallback when UserHome/Save fails; interval clamp mismatch (setter 1–60, handlers allow 120).
- Timeline sync ignores config: Init always starts sync even when auto-update disabled; interval hardcoded to 60s; spinner frame stalls outside menu; no-remote path reports success and updates last-sync, fetch errors drop stderr.
- Preferences UI stub: duplicate placeholder config structs, hardcoded colors, ignores `DynamicSizing`/theme; toggle handler doesn’t update `app.appConfig` or reschedule sync; unused locals.
- Config menu SSOT: ID drift (`config_toggle_autoupdate` dead helper) and unused `configSelectedIdx`.
- Branch picker incomplete: render-only, no branch switch action or dirty-tree handling; uses raw widths/colors; scroll offset unused.
- Fail-fast gaps: config load/create swallow errors; timeline sync masks missing-remote and fetch stderr.

## Recommended Fixes (for CARETAKER)
1) Config SSOT & fail-fast: unify defaults to theme `gfx`, interval `5`; return errors on UserHome/save failures; clamp consistently (choose 1–60) and match UI. 
2) Timeline sync compliance: only start when `AutoUpdate.Enabled` and remote present; add SSOT `TimelineSyncInterval`; honor `AutoUpdate.IntervalMinutes`; keep spinner frames consistent or suppress when hidden; no-remote should not mark last-sync and should report condition; include stderr on fetch errors.
3) Preferences UI/state: use real `config.Config`, theme, and `DynamicSizing`; remove placeholder structs/colors; on toggle/interval/theme, mutate `app.appConfig`, persist, and reschedule/cancel sync; align clamps.
4) Config menu SSOT: use single ID `config_toggle_auto_update`; delete dead `getConfigMenuItems()` or reconcile; drop unused `configSelectedIdx` if unneeded.
5) Branch picker: wire ENTER to branch switch (with dirty-tree flow per kickoff: stash/commit prompt), use theme+sizing, honor scroll offset, show spinner while loading.
6) Error messaging: surface timeline fetch stderr and no-remote; avoid marking last-sync on no-remote.

## Files to Touch (guidance)
- `internal/config/config.go`
- `internal/app/app.go`, `internal/app/timeline_sync.go`, `internal/app/history_cache.go`
- `internal/ui/preferences.go`
- `internal/app/menu_items.go`, `internal/app/menu.go`, `internal/app/dispatchers.go`
- `internal/git/branch.go` (switch flow/dirty handling as needed)

## Notes
- Maintain fail-fast: no silent defaults on critical paths; surface footer/console errors. 
- Keep menu/timeline constants in SSOT; remove duplicate helpers.
