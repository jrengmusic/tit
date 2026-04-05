# RFC: Release Infrastructure for TIT

**Author:** COUNSELOR (CAKE sprint handoff)
**Date:** 2026-04-05
**Status:** Ready for execution

---

## Context

TIT (Terminal Interface for git) needs the same release infrastructure that was built for CAKE. This RFC documents the exact steps — COUNSELOR can execute directly without discovery.

**Reference implementation:** `~/Documents/Poems/dev/cake/` (completed sprint, all verified)

---

## Current State

| Item | Status |
|------|--------|
| Module path | Bare `tit` — needs `github.com/jrengmusic/tit` |
| GitHub remote | `git@github.com:jrengmusic/tit` |
| Version | Hardcoded `const AppVersion = "v1.3.1"` in `internal/constants.go` |
| Build script | `build.sh` — local only, no version injection |
| goreleaser | None |
| CI | None |
| Signing | None |
| Tests | Zero test files (171 .go files) |
| Entry point | `cmd/tit/main.go` (+ platform-specific input files) |
| Go version | 1.24.0 |

---

## Scope

### Phase 1: Module path migration

Change `tit` to `github.com/jrengmusic/tit` in:
- `go.mod` module line
- All internal imports across 171 .go files (grep `"tit/`)

Verify: `go build ./...` clean after migration.

### Phase 2: Version injection

In `internal/constants.go`:
- Change `AppVersion` from `const` to `var`
- Default value: `"dev"`
- Injected at build time via `-ldflags "-X github.com/jrengmusic/tit/internal.AppVersion=vX.Y.Z"`

Update `build.sh`:
- Add `VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")`
- Add `-X github.com/jrengmusic/tit/internal.AppVersion=$VERSION` to ldflags

### Phase 3: goreleaser setup

Copy pattern from CAKE's `.goreleaser.yaml` with these adjustments:
- `project_name: tit`
- `main: ./cmd/tit`
- `binary: tit_{{ .Os }}_{{ .Arch }}`
- ldflags: `-X github.com/jrengmusic/tit/internal.AppVersion={{.Version}}`
- Same 6 targets: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`
- macOS signing via `scripts/post-build.sh` wrapper (not inline hooks — goreleaser OSS does not support `if` filters on build hooks)
- Release to `github.com/jrengmusic/tit`
- Release notes via CLI flag `--release-notes=RELEASE_NOTES.md` (not YAML field — that's PRO-only)

Copy from CAKE (identical content):
- `entitlements.plist`
- `scripts/post-build.sh` — signs + notarizes darwin builds, skips others. Uses `mktemp -d` for unique zip per concurrent hook invocation, `ditto -c -k` to zip binary for notarytool submission.

Create `RELEASE_NOTES.md` at project root with release description for GitHub.

Add `/dist/` to `.gitignore` (goreleaser output directory — `git add -A` will commit it otherwise).

### Phase 4: Test coverage

Discover pure/testable functions across packages (same approach as CAKE):
- Identify functions with no filesystem/subprocess dependencies
- Write white-box tests using standard `testing` package
- Table-driven tests, no external deps
- Skip bubbletea model packages (no viable unit test path)

### Phase 5: release.sh

Copy pattern from CAKE's `release.sh` — single command to commit, tag, push, and goreleaser:
```bash
bash release.sh v0.0.1
```

Tag is required, commit message defaults to tag name. Release notes come from `RELEASE_NOTES.md`.

Script handles: delete existing tag if present, `git add -A`, commit, tag, push, `GITHUB_TOKEN=$(gh auth token) goreleaser release --clean --release-notes=RELEASE_NOTES.md`.

Uses `gh auth token` to bridge GitHub CLI auth — no separate token file needed.

---

## Signing

macOS only. Same setup as CAKE:
- Identity: `Developer ID Application: Bayu Ardianto (9BDSN9TDX3)`
- Keychain profile: `notary`
- Entitlements: `entitlements.plist` (allow-unsigned-executable-memory, disable-library-validation)
- No Windows signing

---

## What NOT to do

- No GitHub Actions (ARCHITECT preference)
- No goreleaser PRO (post-build hooks work fine)
- No Homebrew formula (not in scope)
- Do not change any application behavior — infrastructure only

---

## Acceptance Criteria

- [ ] Module path: `github.com/jrengmusic/tit`
- [ ] `go build ./...` clean
- [ ] `go test ./...` all pass (with new test files)
- [ ] `go vet ./...` clean
- [ ] `AppVersion` injected via ldflags (var, not const)
- [ ] `build.sh` injects version from git tag
- [ ] `.goreleaser.yaml` produces 6 binaries
- [ ] `goreleaser release --snapshot --clean` succeeds locally
- [ ] macOS binaries signed and notarized
- [ ] Checksums generated

---

## Pitfalls (learned from CAKE release)

- goreleaser OSS does not support `if` on build hooks — use wrapper script with `$HOOK_TARGET` env var
- goreleaser OSS does not support `release_notes` YAML field — use `--release-notes` CLI flag
- `notarytool` requires zip, not bare binary — use `ditto -c -k --keepParent` to zip, submit zip, delete after
- Hooks run concurrently per target — use `mktemp -d` (unique dir per invocation), not `mktemp` with fixed suffix
- `git add -A` before goreleaser will commit `dist/` — add `/dist/` to `.gitignore` first
- `archives.format: binary` is deprecated in goreleaser v2 — check current docs for correct syntax
- `GITHUB_TOKEN` required — bridge from `gh auth token` in release.sh, no separate token file

## Reference Files (CAKE)

These files in `~/Documents/Poems/dev/cake/` are the reference implementation:
- `.goreleaser.yaml` — goreleaser config with post-build hook (no inline signing)
- `scripts/post-build.sh` — macOS sign+notarize wrapper (concurrent-safe)
- `RELEASE_NOTES.md` — release description for GitHub (update before each release)
- `entitlements.plist` — macOS codesign entitlements
- `.gitignore` — includes `/dist/`
- `build.sh` — local build with version injection
- `release.sh` — one-command release: tag required, message optional, gh auth token bridge
- `internal/constants.go` — var AppVersion pattern
