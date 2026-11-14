# PLAN: TIT Part 4 — Input Bytes to `juce::KeyPress`

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-13
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes JAM only, in `jam_terminal/keyboard/`. `TerminalKeyboard` turns a `juce::KeyPress` into bytes today (`jam_TerminalKeyboard.h:117`). This part adds the other direction: the bytes that a host terminal sends become `juce::KeyPress` values, and a bracketed paste becomes one text.

The decoder is the inverse of the encoder. Each inverse lookup comes from the encoder's own table, through `jam::Map::getKey` (`jam_core/utils/jam_Bimap.h:46-57`). So one table holds each key in both directions (MANIFESTO S).

The decoder holds no state. It returns the number of bytes that it used. The caller keeps the rest for the next read. Part 7 is the caller (ruling 30). Its terminal chain reads `TerminalDisplay` into an input lane on the device context. On the message thread it decodes the lane and calls `handleKeyPress`, or `insertTextAtCaret` through `findCurrentTextInputTarget()`, on the peer of `TerminalWindow` (Part 5).

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Part 3 (the display gives the bytes; `writePrivateMode` sets bracketed paste).

## Language / Framework Constraints

- CODING.md: templates stay in the header; every other body is in `jam_TerminalKeyboard.cpp`.
- CODING.md: no `==` on strings: a prefix test is `compare (...) == 0`. More than three cases is a lookup table (MANIFESTO L).
- CODING.md, "Owner first, narrowest scope": a value used by one function lives in that function. A value that two members use moves to class scope.
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build: agents build only on ARCHITECT's explicit request.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/jam/` unless the row says otherwise.

### The encoder that this part inverts

| Piece | Where |
|---|---|
| `map (key, applicationCursor, keyboardFlags)`: Ctrl without Alt, then Alt, then plain | `jam_terminal/keyboard/jam_TerminalKeyboard.h:117`; `.cpp:22-32` |
| `simpleKeys`: return `\r`, backspace `\x7f`, tab `\t`, escape `\x1b` | `.h:328`; `.cpp:194-200` |
| `ctrlKeys`: `[` `\\` `]` `^` `_` `@` space to `\x1b` `\x1c` `\x1d` `\x1e` `\x1f` NUL NUL | `.h:343`; `.cpp:202-211` |
| Ctrl+A to Z: `code - 'A' + ctrlLetterBase` | `.h:309`; `.cpp:40-43` |
| `cursorKeys`: up, down, right, left to `A` `B` `C` `D` | `.h:357`; `.cpp:213-219` |
| `editingCodes`: insert 2, delete 3, page up 5, page down 6 | `.h:373`; `.cpp:221-227` |
| `editingFinalChars`: home `"H"`, end `"F"`, as `const char*` | `.h:544`; `.cpp:381-385` |
| F1 to F4: a local `ss3Final { 'P', 'Q', 'R', 'S' }` | `.cpp:343` |
| F5 to F12: a local `csiCode { 15, 17, 18, 19, 20, 21, 23, 24 }` | `.cpp:361` |
| Shift+Tab: a local `backTabFinal { 'Z' }`, `CSI Z` | `.cpp:129,144-150` |
| The modifier parameter: `noModifier + shiftBit + altBit + ctrlBit`, and `metaBit` for Cmd on macOS only | `.h:285-297,398`; `.cpp:229-259` |
| The legacy Alt form: ESC, then the bytes of the key without Alt | `.cpp:79-125`, lines `92-120` |
| `bracketedPasteStart` `\x1b[200~`, `bracketedPasteEnd` `\x1b[201~` | `.h:124,131` |

### The pieces the decoder uses

| Piece | Where |
|---|---|
| `jam::Map::getKey (const M&)`: the inverse of a map, as a `jam::HashMap`. It uses `emplace`, so the first row of a value wins. | `jam_core/utils/jam_Bimap.h:46-57` |
| `jam::HashMap` iterates in insertion order | `jam_core/utils/jam_Bimap.h:213-216` |
| `AnsiDocument::getToken (Cursor)`: a text run, a control byte, a CSI sequence, or `unterminated`. A text run drops a cut UTF-8 character at its end. | `jam_terminal/document/jam_AnsiDocument.h:285-298`; `.cpp:36-121`, lines `56-66` |
| `AnsiDocument::getCsiParameters (text)`: the fields between the introducer and the final byte | `jam_AnsiDocument.h:218-227` |
| `AnsiDocument::getParameterValue (parameter, defaultValue)` | `jam_AnsiDocument.h:264-271` |
| `jam::Document::Cursor { source, position }` | `jam_core/document/jam_Document.h:817` (doxygen `structjam_1_1Document_1_1Cursor`) |
| The applied-bytes return of a decoder: `applyTokens` returns the "Number of bytes that the call applied" | `jam_terminal/processor/jam_TerminalProcessor.h:293-311` |
| A callback that receives each item: `applyReleasedImages (Callback&&)` | `jam_terminal/buffers/jam_TerminalBuffer.h:223-236` |

### How `jam::CommandManager` matches a key

| Fact | Where |
|---|---|
| A sheet token of `sequenceLength` (2) characters or fewer matches the text characters typed. A longer token is a `juce::KeyPress::createFromDescription` and matches by `KeyPress ==`. | `jam_gui/commands/jam_CommandManager.cpp:21-24,96-109`; doxygen `classjam_1_1CommandManager`: `sequenceLength { 2 }` |
| The typed text is `juce::String::charToString (key.getTextCharacter())` | `jam_CommandManager.cpp:50-52` |
| `KeyPress ==` needs equal raw modifier flags. A text character of 0 on either side matches. A key code below 256 matches without case. | JUCE `juce_gui_basics/keyboard/juce_KeyPress.cpp:52-63` |

So a decoded text key needs its text character. A decoded named key (`escape`, `return`, `cursor up`, `ctrl + R`, `alt + V`) needs its key code and its exact modifiers.

### The keys that TIT binds today

| Key | Where |
|---|---|
| `ctrl+c`, `ctrl+v`, `alt+v` | `tit/___legacy___/internal/app/app_keys.go:10,14,17` |
| `ctrl+r` | `app_keys.go:87`; `tit/SPEC.md:651,988` |
| `ctrl+j` inserts a newline in an input | `tit/___legacy___/internal/app/app_update_msg.go:99-103` |
| `shift+]`, `Shift + / -` | `tit/SPEC.md:265,322,998` |

Every key in this table has a legacy encoding with no ambiguity. A terminal sends `shift+]` as the text `}`, and `Shift +` as the text `+`. Part 9 writes those characters in the key sheet.

### Protocol facts (research, session 2026-11-13)

| Fact | Source |
|---|---|
| Bracketed paste: `"ESC [ 200 ~, followed by the pasted text, followed by ESC [ 201 ~."`; set with `CSI ? 2004 h`, reset with `CSI ? 2004 l` | xterm ctlseqs (xfree86.org mirror; the invisible-island.net copy refused the fetch) |
| xterm Alt: `"Send ESC when Meta modifies a key (enables the metaSendsEscape resource)."` | same mirror |
| kitty legacy table: F1 to F4 are `SS3 P` to `SS3 S`; up is `CSI A` or `SS3 A` | sw.kovidgoyal.net/kitty/keyboard-protocol, "Legacy functional keys" |
| A lone ESC, crossterm: Esc is a key only when no more input is available; no timer | crossterm `src/event/sys/unix/parse.rs` |
| A lone ESC, termwiz (WezTerm): the `maybe_more` flag: "If you have attempted a read and no data is immediately available, you should follow up with a call to parse" | termwiz `src/input.rs` |
| A lone ESC, notcurses: "an escape must arrive as a single unit to be interpreted as an escape." | notcurses `src/lib/in.c` |
| A lone ESC, Vim and Neovim: a timer, `'ttimeoutlen'` | Vim `runtime/doc/options.txt`; Neovim `src/nvim/tui/input.c` |

The kitty keyboard protocol is not pushed. The legacy encoding covers every key that TIT binds (table above), and a lone ESC follows the rule of crossterm, termwiz, and notcurses. A pushed protocol would also need a pop at every exit path ("Emit the escape sequence `CSI < u` at application exit", kitty protocol, "Quickstart") (MANIFESTO L, YAGNI).

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: Put the function key and back tab values in tables
**Scope:** `jam_terminal/keyboard/jam_TerminalKeyboard.h`, `jam_terminal/keyboard/jam_TerminalKeyboard.cpp`.
**Action:** The decoder needs these values too, so each moves from a function body to class scope (CODING.md, "Owner first, narrowest scope").
1. Delete the local `ss3Final` (`.cpp:343`) and the local `csiCode` (`.cpp:361`). Add two private tables after `editingCodes` (`.h:373`):
   - `static const jam::HashMap<int, char> functionFinalChars;` with `F1Key` `'P'`, `F2Key` `'Q'`, `F3Key` `'R'`, `F4Key` `'S'`.
   - `static const jam::HashMap<int, int> functionCodes;` with `F5Key` 15, `F6Key` 17, `F7Key` 18, `F8Key` 19, `F9Key` 20, `F10Key` 21, `F11Key` 23, `F12Key` 24.
   `getFunctionKey` looks up `juceKey` in these tables, with `find`, as `getModifiedKey` looks up `cursorKeys` (`.cpp:64-69`). It gives the same bytes as before.
2. Change `editingFinalChars` (`.h:544`; `.cpp:381-385`) from `jam::HashMap<int, const char*>` to `jam::HashMap<int, char>`, with `'H'` and `'F'`. A `const char*` key of an inverse map compares pointers, not text. `getEditingKey` appends the `char` (`.cpp:399,405`), and gives the same bytes as before.
3. Delete the local `backTabFinal` (`.cpp:129`). Add `static constexpr char backTabFinal { 'Z' };` at class scope, beside `ctrlLetterBase` (`.h:309`).

**Validation:**
- `map()` gives the same bytes as before for every key of the tables, with each modifier set.
- New names and their families: `functionFinalChars` (`editingFinalChars`), `functionCodes` (`editingCodes`); `backTabFinal` keeps its name.

### Step 2: The decoder
**Scope:** `jam_terminal/keyboard/jam_TerminalKeyboard.h`, `jam_terminal/keyboard/jam_TerminalKeyboard.cpp`.
**Action:**
1. Add the public template after `map()` (`.h:117`):

```cpp
template <typename KeyCallback, typename PasteCallback>
static int applyKeyPresses (std::string_view bytes, bool hasMoreInput, KeyCallback&& keyCallback, PasteCallback&& pasteCallback)
{
    size_t offset { 0 };
    size_t length { getKeyLength (bytes, hasMoreInput) };

    while (length > 0)
    {
        const auto key { bytes.substr (offset, length) };

        if (key.compare (0, bracketedPasteStart.size(), bracketedPasteStart) == 0)
            pasteCallback (key.substr (bracketedPasteStart.size(), key.size() - bracketedPasteStart.size() - bracketedPasteEnd.size()));
        else
            applyKeyPress (key, keyCallback);

        offset += length;
        length = getKeyLength (bytes.substr (offset), hasMoreInput);
    }

    return static_cast<int> (offset);
}
```

   The return is the number of bytes used. The caller keeps the rest, which starts with a cut key (MANIFESTO, permitted copy 3, "Look-ahead in a parse").
2. Add the private template that gives the key presses of one key:

```cpp
template <typename KeyCallback>
static void applyKeyPress (std::string_view key, KeyCallback&& keyCallback)
{
    if (static_cast<uint8_t> (key.front()) >= static_cast<uint8_t> (Chars::space) and key.front() != static_cast<char> (Chars::deleteCharacter))
    {
        for (juce::CharPointer_UTF8 character { key.data() }; character.getAddress() < key.data() + key.size();)
            keyCallback (getTextKeyPress (character.getAndAdvance()));
    }
    else if (const auto keyPress { getKeyPress (key) }; keyPress.isValid())
    {
        keyCallback (keyPress);
    }
}
```

   A text run gives one key press for each character. Any other key gives one key press, or none when it has no mapping, as `map()` gives an empty string for a key with no mapping (`.h:113-115`).
3. Add the private `static size_t getKeyLength (std::string_view bytes, bool hasMoreInput);`. It gives the length of the first key at the start of `bytes`, or 0 when the bytes hold no complete key:
   - Empty bytes give 0.
   - A first byte that is not ESC gives the length of `AnsiDocument::getToken (jam::Document::Cursor { bytes, 0 })`: one control byte, or a text run. The control bytes are C0 and DEL (`jam/cast/terminal.md:437-440,480`), so Backspace (`\x7f`) is never in a text run. A token of type `map::AnsiTokenType::unterminated` gives 0. The text run drops a cut UTF-8 character at its end (`jam_AnsiDocument.cpp:56-66`).
   - ESC as the last byte gives 1 when `hasMoreInput` is `false`: the Escape key (crossterm, termwiz, notcurses). It gives 0 when `hasMoreInput` is `true`.
   - ESC then `[`: the length of the CSI token from `AnsiDocument::getToken`, or 0 when it is `unterminated`. When the CSI token is `bracketedPasteStart`, the length runs to the end of the first `bracketedPasteEnd` after it (`std::string_view::find`). No end in the bytes gives 0.
   - ESC then `O`: 3 when a third byte exists. Otherwise 0 when `hasMoreInput` is `true`, and 2 (Alt+O) when it is `false`.
   - ESC then any other byte: 1 plus the length of the key that follows, by the two rules for a first byte that is not ESC, but one character only for a text run.
   The ESC cases are a lookup on the byte after ESC, with the other bytes as the default (MANIFESTO L, three branches).
4. Add the private `static juce::KeyPress getKeyPress (std::string_view key);`. A first byte of ESC gives `getEscapeKeyPress (key)`. Any other first byte is a control byte and gives `getControlKeyPress (key.front())`.
5. Add the private helpers. Each inverse map is a function-local `static const auto`, built once from the encoder's table with `jam::Map::getKey`:
   - `static juce::KeyPress getTextKeyPress (juce::juce_wchar character) noexcept;`: `a` to `z` give the upper-case key code, no modifier, and the character. `A` to `Z` give the key code, `shiftModifier`, and the character. Any other character gives itself as key code and character. This is the inverse of `getTextKey` and of the letter rule of `mapAlt` (`.cpp:155-170,107-108`).
   - `static juce::KeyPress getControlKeyPress (char byte);`: the inverse of `simpleKeys` first, with no modifier: return, backspace, tab, escape. Then bytes 1 to 26 give `'A' + byte - ctrlLetterBase` with `ctrlModifier` and character 0 (the inverse of `.cpp:43`). Then the inverse of `ctrlKeys`, with `ctrlModifier`. NUL gives `@`, the first row of NUL in `ctrlKeys` (`.cpp:209`). Any other byte gives an invalid `KeyPress`.
   - `static juce::ModifierKeys getModifierKeys (int modifierCode) noexcept;`: the inverse of `getModifierCode` (`.cpp:229-259`). Bits of `modifierCode - noModifier`: `shiftBit` gives `shiftModifier`, `altBit` gives `altModifier`, `ctrlBit` gives `ctrlModifier`. Under `#if JUCE_MAC` only, `metaBit` gives `commandModifier`, as `getModifierCode` adds it on macOS only.
   - `static juce::KeyPress getEscapeKeyPress (std::string_view key);`: ESC alone gives `escapeKey`. `ESC [` gives `getCsiKeyPress (key)`. `ESC O` with a third byte gives the inverse of `functionFinalChars`, then of `cursorKeys`, with no modifier. ESC with any other key gives the key press of that key through `getControlKeyPress` or `getTextKeyPress`, with `altModifier` added: the inverse of the legacy Alt form (`.cpp:92-120`).
   - `static juce::KeyPress getCsiKeyPress (std::string_view sequence);`: the fields come from `AnsiDocument::getCsiParameters (sequence)`. The key number is `getParameterValue` of the first field, default 1. The modifiers are `getModifierKeys` of the second field, default `noModifier`. The final byte selects the function in the private table `csiKeyConstruction`, a `jam::HashMap<char, juce::KeyPress (*) (int keyNumber, juce::ModifierKeys modifiers)>`:
     - `A` `B` `C` `D`: the inverse of `cursorKeys`.
     - `H` `F`: the inverse of `editingFinalChars`.
     - `P` `Q` `R` `S`: the inverse of `functionFinalChars`.
     - `~`: the inverse of `editingCodes` by key number, then the inverse of `functionCodes`.
     - `backTabFinal`: `tabKey` with `shiftModifier`.
     A final byte with no row gives an invalid `KeyPress`. A named key has character 0.
6. Delete nothing more. `encodeWin32Input` (`.h:133-158`) does not change.

**Validation:**
- Round trip, by desk check: for each key of `simpleKeys`, `ctrlKeys`, `cursorKeys`, `editingCodes`, `editingFinalChars`, `functionFinalChars`, `functionCodes`, and Ctrl+A to Z, with no modifier and with each of Shift, Alt, Ctrl, `applyKeyPresses (map (key, false, 0), false, ...)` gives one key press equal to `key` by `KeyPress ==`. The two NUL keys both give Ctrl+@. Ctrl+M, Ctrl+I, and Ctrl+[ give Return, Tab, and Escape: their bytes are the bytes of `simpleKeys`, and `getControlKeyPress` tries `simpleKeys` first.
- A text run gives the text characters typed, so a sheet token of 2 characters or fewer matches (`jam_CommandManager.cpp:103`).
- Each key of the table "The keys that TIT binds today" decodes to the key of the sheet token: `ctrl+c`, `ctrl+v`, `alt+v`, `ctrl+r`, `ctrl+j`, `}`, `+`, `-`. A sheet token has no space, because `setKeyPresses` splits a value at spaces (`jam_CommandManager.cpp:16`).
- Bytes that end inside a CSI sequence, an SS3 sequence, a UTF-8 character, or a bracketed paste give a return below `bytes.size()`. The rest is the cut key.
- New names and their families:
  - `applyKeyPresses` (`applyReleasedImages`, `jam_TerminalBuffer.h:230`; the applied-bytes return of `TerminalProcessor::applyTokens`, `jam_TerminalProcessor.h:311`)
  - `applyKeyPress` (`applyKeyPresses`)
  - `getKeyLength` (the `get…` length family of `Document::Token::length`)
  - `getKeyPress`, `getTextKeyPress`, `getControlKeyPress`, `getEscapeKeyPress`, `getCsiKeyPress` (the encoder's `getTextKey`, `getSimpleKey`, `getCursorKey`, with the output domain `KeyPress`, NAMES.md Rule 7)
  - `getModifierKeys` (`getModifierCode`)
  - `csiKeyConstruction` (`csiConstruction`, `keyboardConstruction`, `jam_TerminalProcessor.h:238,243`; NAMES.md Rule 8)
  - `hasMoreInput` (NAMES.md Rule 1, a boolean; termwiz `maybe_more`)

### Step 3: Build
**Scope:** JAM through `ling` on macOS and on Windows.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build).
**Action:** Build `ling` with `cast cast/spell.md --debug` on both machines. Run `ling --help`.
**Validation:** Each build ends with exit code 0. `ling --help` is byte-identical to its output before this part: LING does not call the changed code paths. END and EVE use `map()`; its bytes do not change (Step 1). Part 7 runs the decoder with live keys.

## BLESSED Alignment

- **B — Bound.** The decoder owns nothing. The caller owns the bytes and the cut key.
- **L — Lean.** No kitty protocol push, no timer. The decoder covers the legacy encoding that the encoder writes.
- **E — Explicit.** `hasMoreInput` states the one fact that the lone ESC rule needs. The CSI final bytes are a table.
- **S — SSOT.** Each inverse comes from the encoder's own table. The function key values leave the function bodies, so the encoder and the decoder read one table.
- **S — Stateless.** `TerminalKeyboard` stays a namespace-struct of static members.
- **E — Encapsulation.** The decoder gives `juce::KeyPress` and text. It knows no component. `jam_terminal` includes no `jam_gui` header.
- **D — Deterministic.** The same bytes and the same `hasMoreInput` give the same key presses and the same return.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. The lone ESC rule has a known limit, shared with crossterm and termwiz: a host that splits ESC and the rest of a sequence over two reads gives an Escape key and then text.
