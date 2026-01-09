# TIT — Terminal Interface for Git

A state-driven terminal UI for git repository management. Built with Go, Bubble Tea, and Lip Gloss.

**Philosophy:** Git state determines UI. If an action appears in the menu, it will succeed. Zero surprises.

## Features

- **State-Driven Menu** — Available actions derived from actual git state (clean/dirty, ahead/behind/diverged)
- **Time Travel** — Browse commit history, checkout old commits (read-only), merge changes back
- **History Browser** — 2-pane commit viewer with full metadata
- **File History** — 3-pane browser with per-file diffs across commits
- **Dirty Operation Protocol** — Automatic stash/restore for pulls, merges, and time travel
- **Conflict Resolution** — Built-in 3-way merge conflict resolver
- **Themed UI** — Customizable colors via `~/.config/tit/themes/default.toml`

## Requirements

- Go 1.21+
- Git
- Terminal: 80×30 minimum

## Build

```bash
./build.sh
```

Creates `tit_x64` (or `tit_arm64` on ARM).

## Usage

```bash
./tit_x64
```

Run from any directory. TIT detects git state and shows appropriate options.

### Keyboard

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Navigate menu |
| `Enter` | Execute action |
| `Tab` | Cycle panes (in browsers) |
| `Esc` | Back / Cancel / Abort |
| `Ctrl+C` | Exit (press twice) |

## State Model

TIT tracks four axes:

| Axis | Values |
|------|--------|
| **WorkingTree** | Clean, Dirty |
| **Timeline** | InSync, Ahead, Behind, Diverged |
| **Operation** | Normal, Merging, Conflicted, TimeTraveling, NotRepo |
| **Remote** | HasRemote, NoRemote |

Menu items appear only when git state permits them.

## Architecture

```
cmd/tit/          Entry point
internal/
├── app/          Application logic (modes, menus, handlers, keyboard)
├── git/          Git operations (state detection, commands)
├── ui/           Rendering (layout, theme, history, conflict resolver)
├── banner/       ASCII art banner
└── config/       Configuration loading
```

**Key patterns:**
- Bubble Tea Model-View-Update
- Four-axis state detection (`git.DetectState()`)
- Async operations via `tea.Cmd` (never block UI)
- Cache precomputation (history always instant)

## Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) — Full system design (2000+ lines)
- [SPEC.md](SPEC.md) — Original specification
- [CODEBASE-MAP.md](CODEBASE-MAP.md) — File navigation guide

## License

MIT
