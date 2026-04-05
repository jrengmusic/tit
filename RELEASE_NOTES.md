First public release.

## What is TIT?

Terminal Interface for git. State-aware git TUI with zero-surprise guarantee.
Menu shows only actions that will succeed based on actual git state.

## Features

- Five-axis git state detection (Environment, WorkingTree, Timeline, Operation, Remote)
- Time travel mode (explore history, merge changes back)
- Dirty operation protocol (pull/merge/switch with uncommitted changes)
- Conflict resolver with visual file picker
- Rebase conflict handling (auto-loop through commits)
- Branch operations (create, switch, merge)
- Auto-update with activity-aware polling

## Platforms

| OS | Arch | Signed |
|----|------|--------|
| macOS | Intel (x86_64) | Yes (notarized) |
| macOS | Apple Silicon (arm64) | Yes (notarized) |
| Linux | x86_64 | - |
| Linux | arm64 | - |
| Windows | x86_64 | - |
| Windows | arm64 | - |

## Install

Download binary from this release, or:

```bash
go install github.com/jrengmusic/tit/cmd/tit@latest
```

## Requirements

- Git in PATH
- SSH configured (TIT guides setup if missing)
- Terminal 70x20 minimum
