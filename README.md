<div align="center">
  <img src="screenshot/tit-menu.png" alt="TIT Menu Screenshot">
</div>

# TIT — Terminal Interface for git

I have severe git skills issues.

I'm sick and tired wrestling with git commands that might failed. So i made TUI that always shows exactly what's possible operation based on actual git state. No surprises. No "command failed" messages. No confusion.

**If it's in the menu, it works. Period.**


## TIT Kicks git's Ass 

**✅ Zero-Surprise Guarantee**  
TIT analyzes your git state first, then builds the menu. If an action appears, it will succeed. No more `error: cannot push` after you spend time crafting the perfect commit message.

**🚀 5-Axis State Engine**  
While others show static menus, TIT tracks WorkingTree + Timeline + Operation + Remote + Environment. Dynamic menus that match reality.

**⏰ Time Travel** *(Not just history viewing)*  
Browse any commit, make changes, test them, then merge back to your branch. True exploration with zero consequences until you decide to keep changes.

**🔧 SSH Setup Wizard**  
First run? TIT detects missing SSH keys and walks you through setup. Generate keys, display public key for any remote service of your choice in 30 seconds.

**🎨 Seasonal Themes**  
5 built-in themes including Spring, Summer, Autumn, Winter. Meticulously hand picked color palette that would be a sight for sore eyes.

**🔍 3-Pane File History**  
Not just "what changed"—see Commits + Files + Actual Diffs. Navigate changes instantly with cached history.

**⚡ Auto-Update State**  
Background git state detection keeps TIT current. Menu updates when you switch branches, no refresh needed.

**💪 Conflict Resolution**  
Built-in 3-way merge resolver. TIT will explicitly asked you to resolve immediately for any operations where conflicts might occur before even running.

**🧼 Dirty Operations**
There's no manual stash management. If you choose to pull or time travel with dirty working tree TIT will stash uncommitted changes before continue running operation, and apply that that changes back on top whatever state you currently have after operation. When conflicts occur you must resolve before continue, otherwise it will bring back your dirty tree.

**✍🏻 No rebase**
TIT doesn't write false history. TIT doesn't lie.

Instead:
- **Time travel** to explore old commits safely
- **Merge back** to bring old ideas forward  
- **Clean conflicts** with visual resolution
- **Preserve truth** in your git history

Your timeline should tell the story of what actually happened, not some sanitized fiction.

## Get Started

```bash
./build.sh
./tit_x64
```

**Requirements:** Go 1.25+, Git, Terminal (70×30 minimum)


## Rock 'n Roll Workflow

**Start anywhere:** TIT works in any directory. Not a repo? Get init/clone options.

**See what's possible:** Menu shows only actions that guaranteed to be successful.

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
| `/` | Configs |

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

<div align="center">
  <img src="internal/ui/assets/tit-logo.svg" alt="TIT Logo" width="128" height="128">
</div>

**TIT: lightning in a bottle. Because git is thunder.**

---
Rock 'n Roll!

**JRENG!** 🎸
---
conceived with [CAROL](https://github.com/jrengmusic/carol)
