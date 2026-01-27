# TIT — Terminal Interface for Git

**Philosophy: If it's in the menu, it works. Period.**

Stop wrestling with git commands that might fail. TIT shows you exactly what's possible right now, based on your actual git state. No surprises. No "command failed" messages. No confusion.

## Why TIT Kicks Ass

**✅ Zero-Surprise Guarantee**  
TIT analyzes your git state first, then builds the menu. If an action appears, it will succeed. No more `error: cannot push` after you spend time crafting the perfect commit message.

**🚀 5-Axis State Engine**  
While others show static menus, TIT tracks WorkingTree + Timeline + Operation + Remote + Environment. Dynamic menus that match reality.

**⏰ Time Travel** *(Not just history viewing)*  
Browse any commit, make changes, test them, then merge back to your branch. True exploration with zero consequences until you decide to keep changes.

**🔧 SSH Setup Wizard**  
First run? TIT detects missing SSH keys and walks you through setup. Generate keys, display public key for GitHub/GitLab—done in 30 seconds.

**🎨 Seasonal Themes**  
5 built-in themes including Spring, Summer, Autumn, Winter. Auto-generated color variations that actually look good.

**🔍 3-Pane File History**  
Not just "what changed"—see Commits + Files + Actual Diffs. Navigate years of changes instantly with cached history.

**⚡ Auto-Update State**  
Background git state detection keeps TIT current. Menu updates when you switch branches, no refresh needed.

**💪 Conflict Resolution**  
Built-in 3-way merge resolver. No external tools, no confusion—just mark sections and apply.

## The Difference

| Feature | lazygit | tig | **TIT** |
|---------|---------|-----|---------|
| Menu reflects git state | ❌ | ❌ | **✅ Always** |
| Time travel + merge back | ❌ | ❌ | **✅ Full workflow** |
| SSH wizard | ❌ | ❌ | **✅ Guided setup** |
| Seasonal themes | ❌ | ❌ | **✅ 5 themes** |
| State engine | Basic | None | **✅ 5-axis detection** |
| Zero surprises | ❌ | ❌ | **✅ Guaranteed** |

## Get Started

```bash
./build.sh
./tit_x64
```

**Requirements:** Go 1.21+, Git, Terminal (80×30 minimum)

That's it. TIT detects your setup and guides you through anything missing.

## The Rock 'n Roll Workflow

**Start anywhere:** TIT works in any directory. Not a repo? Get init/clone options.

**See what's possible:** Menu shows only actions that will succeed right now.

**Explore fearlessly:** Time travel lets you test ideas without commitment.

**Stay current:** Auto-update keeps state fresh as you work.

**Resolve conflicts:** Built-in merger handles 3-way conflicts visually.

**Never get stuck:** Every operation has clear escape routes.

## Navigation

| Key | Action |
|-----|--------|
| `↑/k` `↓/j` | Navigate |
| `Enter` | Execute (always works) |
| `Tab` | Switch panes |
| `Esc` | Back/Cancel |
| `Ctrl+C` | Exit (press twice) |

## The "Fuck Rebase" Philosophy

**TIT doesn't do interactive rebase.** Why? Because rebase is history vandalism.

Instead:
- **Time travel** to explore old commits safely
- **Merge back** to bring old ideas forward  
- **Clean conflicts** with visual resolution
- **Preserve truth** in your git history

Your timeline should tell the story of what actually happened, not some sanitized fiction.

## For Developers

**Built with:** Go + Bubble Tea + Lip Gloss  
**Architecture:** State-driven Model-View-Update  
**No dependencies:** Single static binary  
**No config files:** State reflects git reality  

**Documentation:**
- [SPEC.md](SPEC.md) — Complete technical specification
- [ARCHITECTURE.md](ARCHITECTURE.md) — System design (2000+ lines)
- [CODEBASE-MAP.md](CODEBASE-MAP.md) — Navigation guide

**Philosophy docs:**
- Truth-preserving git workflows
- 5-axis state detection engine
- Zero-surprise menu contracts
- Time travel implementation

## License

MIT — Use it, break it, fix it, ship it.

---

**TIT: Because git deserves a UI that doesn't suck.**