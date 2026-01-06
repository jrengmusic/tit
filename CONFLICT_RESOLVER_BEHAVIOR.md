# Conflict Resolver - Expected Behavior

## Border Colors (Fixed)

### Unfocused Panes
- **Color:** #2C4144 (littleMermaid - dark subtle line)
- **Purpose:** Recedes into background, not distracting

### Focused Pane
- **Color:** #8CC9D9 (dolphin - bright teal)
- **Purpose:** Clearly indicates which pane is active

**Visual Contract:**
- Dark borders = inactive
- Bright border = "I'm accepting input"

## Navigation Behavior

### Top Row (File Lists) - SHARED NAVIGATION
**All top panes show the same file list.**

**↑↓ Behavior:**
- Moves selection **across all top panes simultaneously**
- Highlighted file is shared (same SelectedFileIndex)
- Each column shows checkbox for **that column's choice**

**Example:**
```
LOCAL (focused)         REMOTE
[✓] src/main.go    ←    [ ] src/main.go    (both show same file, different marks)
[ ] README.md           [✓] README.md
```

Pressing ↓ moves **both** to README.md:
```
LOCAL                   REMOTE
[ ] src/main.go         [ ] src/main.go
[✓] README.md      ←    [✓] README.md      (selection moved in both)
```

### Bottom Row (Content Panes) - INDEPENDENT SCROLLING
**Each pane shows content for different version.**

**↑↓ Behavior:**
- Scrolls **only the focused pane**
- Each pane has its own LineCursor[paneIndex]
- Other panes' scroll positions unchanged

**Example:**
```
LOCAL (line 5)          REMOTE (line 10)
package main            package main
                        
func main() {           import "log"
  fmt.Println() ←         
}                       func main() {
                          log.Println()  ←
                        }
```

Pressing ↓ in LOCAL moves its cursor to line 6.
REMOTE cursor stays at line 10 (independent).

## SPACE Key - Exclusive Marking

**Rule:** Exactly ONE column must be chosen per file.

**Behavior:**
1. SPACE on **unfocused column** → marks that column, unmarks others
2. SPACE on **already marked column** → no change (shows hint)

**Example Workflow:**
```
Step 1: Focus on LOCAL (pane 0)
[✓] src/main.go         [ ] src/main.go
(already marked)

Press SPACE → "Already marked - choose different column"

Step 2: TAB to REMOTE (pane 1)
[✓] src/main.go         [ ] src/main.go
                        (now focused)

Press SPACE → marks REMOTE:
[ ] src/main.go         [✓] src/main.go
```

**Footer Hints:**
- When SPACE pressed on new column: "Marked: src/main.go (column 1)"
- When SPACE pressed on current column: "Already marked - choose different column"

## TAB Navigation

**Order:** Top-left → Top-right → Bottom-left → Bottom-right → wrap

**For 2 columns:**
```
Pane 0 (LOCAL list)  →  Pane 1 (REMOTE list)
      ↑                          ↓
      └─ Pane 3 (REMOTE diff) ← Pane 2 (LOCAL diff)
```

**Visual Feedback:**
- Border changes from dark to bright
- Footer hint updates (if available)

## Summary

| Aspect | Behavior |
|--------|----------|
| **Unfocused Border** | Dark (#2C4144) |
| **Focused Border** | Bright (#8CC9D9) |
| **Top Row ↑↓** | Shared selection across all file lists |
| **Bottom Row ↑↓** | Independent scrolling per pane |
| **SPACE** | Exclusive marking (one column per file) |
| **TAB** | Cycle through all panes |

