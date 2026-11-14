# PLAN: TIT Part 8 — The Git Layer: State Detection

**RFC:** none — objective from ARCHITECT prompt. Canon: `PLAN-tit-0-tracker.md`.
**Date:** 2026-11-13
**BLESSED Compliance:** verified
**Language Constraints:** C++17 / JUCE 8 / JAM (LANGUAGE.md, C++ / JUCE)

## Overview

This part changes TIT only. It builds the state detection of TIT and runs it at the start of the window run.

1. **The rows.** `TITModel` gets a `git` child with the five axes of SPEC, the counts, the branch and hash texts, the marker flags, and the LFS fields (ruling 11; Part 2 Step 3).
2. **`TITGitProcessor`** (ruling 57). Each row of the `git` child is a `jam::Parameter`, and so is the `directory` row. The processor reacts to the `directory` row through a `jam::Model::ParameterAttachment`, as `EVETerminalProcessor` reacts to `winsize` (`eve/Source/EVETerminalProcessor.cpp:68-72`). A change starts one detection job on a one-thread pool. The job runs each git command as a `jam::Subprocess` and waits for it (`ling/Source/CompilerManifest.cpp:147-161`). It writes each row with `jam::Model::storeValue`, from the pool thread, as the handlers of EVE do (`EVETerminalProcessor.cpp:143-144,161-162`; `eve/SPEC.md:802-805`). The flush of the Model writes the changed rows to the tree on the message thread, and the views react there.
3. **The start.** The window run finds the repository folder, deletes a stale index lock, and sets the `directory` row (legacy `app_constructor.go:52-67`).

No view reads the rows yet. Part 10 shows them. Step 9 of this part measures them with a temporary log.

Tracker and canon: `PLAN-tit-0-tracker.md`. This part depends on Parts 1, 2, and 7.

## Language / Framework Constraints

- LANGUAGE.md, C++ / JUCE, and CODING.md: one object is one header and one `.cpp`. A non-template body is in the `.cpp`. The one template, `getFirstMatch`, is in the header.
- CODING.md: no bail-out guard; `not`, `and`, `or`; brace initialization, and parentheses for an initializer of the same container type (`juce::StringArray`, `juce::Array<juce::File>`, `jam::Strings`); no out-parameter, except the `apply` target of the Verb Contract (NAMES.md Rule 1); structured bindings for each pair; no `.first` or `.second`; no `==` on strings (`juce::String::compare (…) == 0`); no anonymous namespace.
- MANIFESTO L: no `if` chain longer than three branches. Two priority chains are first-match tables (Step 7).
- MVP (ruling 57; `end/ARCHITECTURE.md:898-919`): the processor writes state only through the atomic lane, `storeValue`. The view lane is `valueTreePropertyChanged` on the message thread. The processor is not a `juce::ValueTree::Listener`, and it writes no tree.
- EVE's rule for the rows of a state machine, ARCHITECT: *"everything should be Parameter, no strays ana manual state management allowed"* (`eve/SPEC.md:796-797`).
- MANIFESTO E, data in transit: "No temporary container — do not stage the data". The job stores each value as it gets it. It builds no tree.
- No tree of the Model reaches the pool thread. The job receives the folder path as a string.
- ARCHETYPE.md §1: no step edits a generated file by hand (`archetype/ARCHETYPE.md:50-51`). Each cast table keeps its `+---` separator lines (`:41-43`).
- CAROL.md, Code Hygiene: no comment and no doxygen block in this part.
- CAROL.md, Build and Git: agents build only on ARCHITECT's explicit request. The fixture repositories of Step 9 need git commands; ARCHITECT runs them, or gives the git instruction.

## Dependency & API Inventory

Paths are relative to `~/Documents/Poems/dev/` unless the row says otherwise. JUCE paths are relative to `~/Documents/Poems/JUCE/modules/`. `legacy/` stands for `tit/___legacy___/internal/`.

### The legacy detection that this part carries

| Piece | Legacy | Where |
|---|---|---|
| The order of detection: not a repository, then the dirty operation marker, then LFS, then the commits, the tree, the operation, the remote, the timeline, the hash, the branch, the upstream | `DetectState` | `legacy/git/state.go:31-159` |
| Not a repository: `.git` does not exist in the folder. The state holds only the operation. | `IsInitializedRepo`; `return &State{Operation: NotRepo}` | `legacy/git/state_helpers.go:12-25`; `state.go:35-38` |
| The repository folder: the folder or its nearest parent with `.git`, found once at start | `IsInitializedRepo`, `HasParentRepo`, `os.Chdir` | `legacy/app/app_constructor.go:52-67`; `state_helpers.go:27-46` |
| The stale lock: `.git/index.lock` is deleted once at start | `CleanStaleLocks` | `legacy/git/exec_utils.go:44-63`; `app_constructor.go:66` |
| The dirty operation marker: when `.git/TIT_DIRTY_OP` exists, the state holds only the operation | `IsDirtyOperationActive` | `legacy/git/state.go:40-45`; `dirtyop.go:85-91` |
| The working tree: `status --porcelain=v2`; a record that starts with `1`, `2`, or `?` counts as modified | `detectWorkingTree` | `legacy/git/state_detection.go:14-49` |
| Conflicts: a status record that starts with `u ` | `detectOperation` | `legacy/git/state_detection.go:176-186` |
| The time travel marker: valid when HEAD is detached; deleted when HEAD is on a branch | `detectOperation` | `legacy/git/state_detection.go:188-199` |
| Merging: `.git/MERGE_HEAD` exists. Rebasing: `.git/rebase-merge` or `.git/rebase-apply` exists | `detectOperation` | `legacy/git/state_detection.go:201-213` |
| The remote: `git remote` prints a line | `detectRemote` | `legacy/git/state_detection.go:220-231` |
| The commits: `rev-parse HEAD` succeeds with output | `hasCommits` | `legacy/git/state.go:53-55` |
| The timeline: only when the operation is normal, a remote exists, and a commit exists. The upstream range, else the `refs/remotes/origin/<branch>` range, else the count of HEAD as ahead | `detectTimeline` | `legacy/git/state.go:89-106`; `state_detection.go:52-138` |
| The timeline from the two counts: a table keyed by `ahead > 0` and `behind > 0` | `determineTimeline` | `legacy/git/state_detection.go:141-154`; test `state_detection_test.go:5-26` |
| The hash: `rev-parse --short HEAD`; empty on failure | `CurrentHash` | `legacy/git/state.go:108-112` |
| The branch: `symbolic-ref --short HEAD`. On failure, HEAD is detached: the branch of the time travel marker, or `DETACHED` | `CurrentBranch`, `Detached`, `IsTitTimeTravel` | `legacy/git/state.go:114-146` |
| The upstream: `config --get branch.<branch>.remote` succeeds | `CurrentBranchExistsOnRemote` | `legacy/git/state_parsing.go:18-26`; `state.go:148-156` |
| LFS: `.gitattributes` holds `filter=lfs`. Ready: `git-lfs` runs and `config --get filter.lfs.process` prints a value | `IsRepoLFS`, `IsLFSInstalled` | `legacy/git/lfs.go:10-27`; `state.go:47-51` |
| The environment, in this priority: the test variable, git missing, ssh missing, no SSH key, ready | `DetectGitEnvironment` | `legacy/git/environment.go:15-34` |
| The test variable: `TIT_TEST_SETUP=1` forces `NeedsSetup` | `DetectGitEnvironment` | `legacy/git/environment.go:14-19` |
| An SSH key: a file in `~/.ssh` whose name holds `_rsa`, `_ed25519`, `_ecdsa`, or `_dsa`, and that is not a `.pub` file | `sshKeyExists` | `legacy/git/environment.go:42-80` |
| The version commands | `git --version`; `ssh -V` | `legacy/git/environment.go:91-110` |
| The readers of each field. `RemoteHash` has none. | the header, push, remote, LFS, and menu code | `legacy/app/app_view_header.go:39-40,92-117`; `op_push.go:15,42`; `op_push_sync.go:82`; `handlers_remote.go:33`; `op_remote.go:64`; `op_pull.go:90`; `handlers_global_menu.go:282` |

### Three corrections to the legacy detection

Each correction follows SPEC. The plan must align with SPEC (plan-writer step 3).

1. **A detached HEAD hid every mid-operation state.** `DetectState` sets the operation to `TimeTraveling` whenever `symbolic-ref` fails (`legacy/git/state.go:114-143`). A rebase does this: "Check out `<upstream>` with the equivalent of `git checkout --detach <upstream>`" (git `Documentation/git-rebase.adoc:76-77`; `sequencer.c:5006-5024`). `symbolic-ref` on a detached HEAD dies with exit 128 (`builtin/symbolic-ref.c:17-32`; `usage.c:80-87`). So `Rebasing` and a rebase conflict never reached the menu. The legacy start code expects `Rebasing` (`legacy/app/app_constructor.go:190-200`), and SPEC gives the mid-operation states priority 2, before `TimeTraveling` at priority 3 (`tit/SPEC.md:120-133`). This part sets `timeTraveling` only when no mid-operation state exists.
2. **The order of the time travel marker and the merge.** Legacy tests the marker before `MERGE_HEAD` (`legacy/git/state_detection.go:188-205`). SPEC puts `Merging` at priority 2 (`tit/SPEC.md:120-125`). This part tests in the SPEC order: conflicted, merging, rebasing, time traveling, pinned, normal.
3. **The timeline of a manual detached HEAD.** Legacy computed `InSync` for it before the override (`legacy/git/state.go:89-106`). SPEC: "When `Operation = TimeTraveling` ... Timeline = empty" (`tit/SPEC.md:72`). This part computes the timeline only when the operation is `normal`.

### What this part adds to the legacy detection

| Addition | Source |
|---|---|
| The `Pinned` operation and the marker `.git/TIT_PIN`: lines branch, original HEAD, pinned hash | Part 2 Step 3, items 2 and 4; ruling 1 |
| The stale pin marker: HEAD is not on the marker's branch, or HEAD is not the pinned hash. Detection deletes a stale marker. | Part 2 Step 3, item 5 |
| One `status` run gives the working tree and the conflicts. Legacy runs `status` twice with the same output (`legacy/git/state_detection.go:15,177`). | MANIFESTO S (SSOT) |

### What this part drops from the legacy detection

| Dropped | Basis |
|---|---|
| `RemoteHash` (`legacy/git/types.go:81`) | No code reads it (inventory, last row of the first table). MANIFESTO L, YAGNI. |
| The fallback of the branch name to `rev-parse --abbrev-ref HEAD`, and the test for an empty name or `HEAD` (`legacy/git/state_detection.go:62-71`) | The timeline runs only when the operation is `normal`. `normal` comes after `timeTraveling` in the first-match table, so the branch is known. The fallback cannot run. |
| The exclusions of `known_hosts`, `known_hosts.old`, and `config` (`legacy/git/environment.go:69-71`) | None of the three names holds `_rsa`, `_ed25519`, `_ecdsa`, or `_dsa`. The result is the same without them. |

### The patterns that this part follows

| Piece | Where |
|---|---|
| The two lanes: "VTPC (`valueTreePropertyChanged`) — the MESSAGE-thread lane" and the parameter lane, "the any-thread lane"; "one sole AUTHOR writes the change; every listener reacts automatically on its own lane" | `end/ARCHITECTURE.md:910-917` |
| A processor reacts to a row through an attachment: `winsizeAttachment = std::make_unique<jam::Model::ParameterAttachment> (*state.getParameter<jam::Parameter<int64_t>> (…, Id::winsize), [this] (const juce::var&) { restart(); });` | `eve/Source/EVETerminalProcessor.cpp:68-72` |
| `ParameterAttachment`: listens to one parameter; "Delivers changes to a callback on the message thread"; on the message thread the callback runs at once | `jam_data_structures/model/jam_Model.h:237-282`; `jam_Model.cpp:220-240` |
| A tree write of a view reaches the parameter: `Model::valueTreePropertyChanged` calls the adapter, which calls `setValueFromVar`; that stores and notifies the listeners | `jam_Model.cpp:14-21,364-378`; `jam_data_structures/model/jam_ParameterBase.h:168-172` |
| A processor writes a scalar row from its own thread: `state.storeValue (terminalTag, key, value);` | `EVETerminalProcessor.cpp:143-144,193` |
| A processor writes a text row from its own thread: `state.storeValue (terminalTag, key, path, size);` | `EVETerminalProcessor.cpp:161-162` |
| `storeValue`: "Any thread"; it stores the atomic, sets `needsUpdate`, and notifies no listener. The text form stores into a `jam::ParameterText`. | `jam_Model.h:1002-1031`; `jam_ParameterBase.h:174-180` |
| `jam::ParameterText`: "Lock-free SPSC"; one producer thread; `maxTextLength { 1024 }` | `jam_data_structures/model/jam_ParameterText.h:5-8,35` |
| A boolean row is a `jam::Parameter<int>`: "Bool parameters are stored as int (0/1)" | `jam_ParameterBase.h:145` |
| The rows of a state machine are parameters, made at construction: `createAndAddParameter<jam::ParameterText> (overlayRow, Id::message, juce::String {}, maximumMessageLength);`, `model.createAndAddParameter<jam::Parameter<int64_t>> (sessionsTree, Id::focusedSession, …);` | `end/Source/end/ENDModel.cpp:8-9`; `end/Source/plugins/Nexus.cpp:35-39`; `eve/SPEC.md:783-797` |
| The pool: `static constexpr int threadCount { 1 };` and `juce::ThreadPool pool { juce::ThreadPool::Options{}.withNumberOfThreads (threadCount) };` | `ling/Source/Manifest.h:68,74` |
| The application creates the object in its run, then sets the property | `ling/Source/Main.cpp:398-402` |
| A blocking child process on a worker thread: `juce::WaitableEvent finished; int exitCode { EXIT_FAILURE }; ... subprocess.launch (arguments, directory, [&finished, &exitCode] (int processExitCode, const std::string&) { exitCode = processExitCode; finished.signal(); }, [&record] (std::string_view chunk, bool) { record.append (chunk); }); finished.wait();` | `ling/Source/CompilerManifest.cpp:147-161` |
| A command line from a table cell, split into arguments: `jam::ChildProcess::getArguments (commandText)` | `ling/Source/CompilerManifest.cpp:139` |
| The flush writes a row to the tree only when the value differs, on the message thread | `jam_Model.cpp:29-58,291-303` |
| The words of a child tool in a cast table, generated as `text::CLanguage`: `static constexpr const char* const name { "value" };` | `ling/cast/c_language.md:1-18`; `ling/cast/spell.md:25,177-183`; `ling/Source/generated/Text.h:52-63` |
| A name and key table, generated as a `map::` bimap with an enum | `jam/cast/gui.md:257-275`; `jam/cast/spell.md:72,840-851`; `jam/cast/code.cast:160-194` |
| A project bimap in `cast/identifiers.md`, generated into `Identifiers.h` | `eve/cast/identifiers.md:111-136`; `eve/cast/spell.md:27,126-136`; `eve/Source/generated/Identifiers.h:53-93` |
| The first row of a name and key table that means "no value": `none`, key `0` | `jam/cast/terminal.md:691,992,1016` |
| The module row: `| @user-module | jam_subprocess | JUCE-based subprocess launcher with streaming output and byte-cap truncation |` | `ling/project-info.md:135` |
| A first-match search with a result return | MANIFESTO E, `findFirst` example |

### The JAM and JUCE API facts

| Fact | Where |
|---|---|
| `jam::Subprocess::launch (const jam::Strings& command, const juce::File& workingDir, Callbacks::Completion onComplete, Callbacks::Chunk onChunk)`. It asserts `workingDir.isDirectory()`. | `jam_subprocess/subprocess/jam_Subprocess.h:95-98`; `jam_Subprocess.cpp:194-210` |
| The completion capture is capped at `byteCap` (100000 bytes) with a notice. The chunk callback has no cap. So the job collects the chunks. | `jam_Subprocess.h:21-24,108-111`; `jam_Subprocess.cpp:134-173` |
| Stdout and stderr of the child are one pipe | `jam_subprocess/subprocess/jam_ChildProcess.h:12-14` |
| A launch that fails reports exit code -1. On macOS, a missing executable exits with `EXIT_FAILURE` from the child. Both are not 0. | `jam_Subprocess.cpp:78-82,178`; `native/jam_ChildProcess_posix.cpp:83-86` |
| `jam::ChildProcess::getArguments (std::string_view commandLine)` splits at spaces | `jam_ChildProcess.h:61-74` |
| `jam::Strings::add (const juce::String&)` | doxygen `classjam_1_1Strings` |
| The module depends on `jam_core`, `juce_core`, `juce_events` | `jam_subprocess/jam_subprocess.h:10-12` |
| `juce::File::isRoot`, `findChildFiles`, `deleteFile` ("true if the file has been successfully deleted (or if it didn't exist to begin with)"), `hasFileExtension`, `loadFileAsString`, `existsAsFile`, `getCurrentWorkingDirectory`, `getParentDirectory` | doxygen `classjuce_1_1File` |
| `juce::StringArray::fromLines (StringRef)`, `fromTokens (StringRef, bool)` | doxygen `classjuce_1_1StringArray` |
| `git-lfs version` prints the version and exits with 0 | git-lfs `git-lfs.go:31-33`; `commands/run.go:145-159`; `commands/command_version.go:21` |
| `ssh -V` prints the version to stderr and exits with 0 | openssh-portable `ssh.c:906-909` |
| `jam::debug::Log::Scope (const juce::File&)` opens a log file; `jam::debug::Log::write (args...)` joins its arguments with spaces | `jam_core/debug/jam_Log.h:33-41,110-116` |

### Words the framework already holds

| Word | Where |
|---|---|
| `Id::directory` | `jam/cast/core.md` |
| `Id::tit` | Part 7 Step 2 |

`Id::branch` is in `jam/cast/mermaid.md`, a word of `jam_mermaid_diagram`, which TIT does not link. This part uses `currentBranch`, the legacy field name.

### Closed by a read

ARCHITECT can reverse each one.

1. **The name `TITGitProcessor`.** MVP: "The Processor holds the business logic" (MANIFESTO E). Family: `TITTerminalProcessor` (Part 7), `EVETerminalProcessor` (`eve/Source/EVETerminalProcessor.h`). The domain word is `git`, the legacy package name (`legacy/git/`).
2. **Each row of the `git` child is a `jam::Parameter` (ruling 57).** A number, an axis value, or a flag is a `jam::Parameter<int>`. A text is a `jam::ParameterText` of `jam::ParameterText::maxTextLength` bytes: a branch name, a hash, or a path, the size of the text rows of EVE (`eve/Source/layout/parameters.md:39-59`). The job stores each row one time in each detection. The pool has one thread, so each text row has one producer thread.
3. **The environment is a row of the same `git` child, detected first in each job.** SPEC lists it in the state tuple (`tit/SPEC.md:105`) and checks it before the repository (`:111-118`). Legacy detects it at start and in the setup wizard (`legacy/app/app_constructor.go:46`; `setup_wizard_handlers.go:32`). One job serves both: when the environment is not ready, the job detects no repository, as legacy shows the wizard and detects no state (`app_constructor.go:46-50`).
4. **When the operation is `notRepo` or `dirtyOperation`, every other row holds its default.** Legacy returns a state with only the operation in both cases (`legacy/git/state.go:35-45`). The job stores every row in each detection, so the other rows take the defaults of `TITModel::getDefaultGitValues` and the empty text.
5. **The detection runs again only on a trigger.** This part triggers it once, at the start of the window run. A later part runs it again after a command: the processor calls its own `addJob()`. Legacy runs it after each operation and on the auto-update timer (`legacy/app/app.go:105-112`; `auto_update.go:78`).
6. **The streaming run of a user command is not in this part.** Console lines, `--progress`, and `GIT_TERMINAL_PROMPT=0` (`legacy/git/execute.go:96-287`) have no caller here (MANIFESTO L, YAGNI). The tracker lists the child-process facts for the part that runs the first user command.
7. **`rewinding` is a row of the `Operation` table that detection never writes.** SPEC lists it (`tit/SPEC.md:84`), and Part 2 keeps nine rows. Legacy never assigns it; only `legacy/app/menu.go:30` names it.

## Validation Gate

Each step is validated by COUNSELOR before proceeding to the next — against
MANIFESTO.md (BLESSED), NAMES.md, ~/.carol/CODING.md, and the locked PLAN decisions
(no deviation, no scope drift). @Auditor runs ONCE, after the final step, covering
the whole sprint — never per step.

## Steps

### Step 1: The module row
**Scope:** `tit/project-info.md`, table `## user module`.
**Action:** After the `jam_audio_devices` row of Part 7 Step 1, add `| @user-module | jam_subprocess | JUCE-based subprocess launcher with streaming output and byte-cap truncation |`, the row of `ling/project-info.md:135`. Widen the column separators of the table to the new cell widths.
**Validation:**
- The new row has the cell values of the cited row.
- No other row changes. Part 1 named this addition: "Part 8 adds `jam_subprocess`" (`PLAN-tit-1-project.md:179`).

### Step 2: The identifiers and the axis tables
**Scope:** `tit/cast/identifiers.md`.
**Action:**
1. Table `## identifiers`: after the `numKeyPresses` row of Part 7 Step 2, add 16 rows of type `@id` with an empty description. The value of each row is its name in backticks.

   | Name | Family |
   |---|---|
   | `git` | the type of the state child; the legacy package `legacy/git/` |
   | `gitEnvironment` | SPEC axis `GitEnvironment` (`tit/SPEC.md:41`) |
   | `operation` | SPEC axis `Operation` (`tit/SPEC.md:74`) |
   | `workingTree` | SPEC axis `WorkingTree` (`tit/SPEC.md:51`) |
   | `modifiedCount` | legacy field `ModifiedCount` (`legacy/git/types.go:75`) |
   | `timeline` | SPEC axis `Timeline` (`tit/SPEC.md:59`) |
   | `commitsAhead` | legacy field `CommitsAhead` (`types.go:82`) |
   | `commitsBehind` | legacy field `CommitsBehind` (`types.go:83`) |
   | `remote` | SPEC axis `Remote` (`tit/SPEC.md:86`) |
   | `currentBranch` | legacy field `CurrentBranch` (`types.go:79`) |
   | `currentHash` | legacy field `CurrentHash` (`types.go:80`) |
   | `hasUpstream` | the reader's name: `hasUpstream := a.gitState.LocalBranchOnRemote` (`legacy/app/op_push.go:15`) |
   | `isDetached` | legacy field `Detached` (`types.go:85`), with the `is` prefix of a boolean (NAMES.md Rule 1; `Id::isReadOnly`, tracker) |
   | `isTitTimeTravel` | SPEC `IsTitTimeTravel` (`tit/SPEC.md:92`) |
   | `isUsingLfs` | legacy field `LFS`, "Repo has .gitattributes with filter=lfs" (`types.go:87`), with the prefix of `isUsingWhitespace` (NAMES.md Rule 1) |
   | `isLfsReady` | legacy field `LFSReady` (`types.go:88`), with the `is` prefix |

2. After the table `## identifiers`, add five tables in the form of `## EditorMode` (`jam/cast/gui.md:257-275`): a heading, a fenced `@brief` block, and a `name` and `key` table. Write the rows in this order.

   | Table | `@brief` | Rows (name, key) | Source of the rows |
   |---|---|---|---|
   | `## GitEnvironment` | `The readiness of git and SSH on this machine. The first row is the default.` | `ready` 0, `needsSetup` 1, `missingGit` 2, `missingSsh` 3 | `tit/SPEC.md:41-47`; the legacy order `legacy/git/types.go:50-55` |
   | `## WorkingTree` | `The local file changes. The first row is the default.` | `clean` 0, `dirty` 1 | `tit/SPEC.md:51-55` |
   | `## Timeline` | `The local branch against its remote. The first row means no comparison: no remote, no commit, or an operation that is not normal.` | `none` 0, `inSync` 1, `ahead` 2, `behind` 3, `diverged` 4 | `tit/SPEC.md:59-72`; `none` from `jam/cast/terminal.md:691` |
   | `## Operation` | `The git operation in progress. The first row is the default.` | `notRepo` 0, `normal` 1, `merging` 2, `conflicted` 3, `rebasing` 4, `timeTraveling` 5, `dirtyOperation` 6, `rewinding` 7, `pinned` 8 | `tit/SPEC.md:74-84`; `pinned` from Part 2 Step 3, item 2 |
   | `## Remote` | `The presence of a remote. The first row is the default.` | `noRemote` 0, `hasRemote` 1 | `tit/SPEC.md:86-90` |

**Validation:**
- Each row name is the camel-case form of its SPEC or legacy word, or a cited family.
- No word of JAM is declared again (`Id::directory`, `Id::tit`).
- The Operation table has nine rows, as Part 2 Step 3 requires.

### Step 3: The file names and the git words
**Scope:** `tit/cast/files.md`, `tit/cast/git.md` (new).
**Action:**
1. `cast/files.md`, table `## files`: after the `titcKeys` row of Part 7 Step 2, add these rows of type `@string`. Change the brief to `@brief File names titc reads: help, keys, and the files of git and SSH.`

   | name | value | description |
   |---|---|---|
   | `gitDirectory` | `` `.git` `` | The folder of the repository data. |
   | `indexLock` | `` `index.lock` `` | The lock file of the git index. titc deletes it at start. |
   | `dirtyOperationMarker` | `` `TIT_DIRTY_OP` `` | The marker of a dirty operation in progress. |
   | `timeTravelMarker` | `` `TIT_TIME_TRAVEL` `` | The marker of a time travel. |
   | `pinMarker` | `` `TIT_PIN` `` | The marker of a pin. |
   | `mergeHead` | `` `MERGE_HEAD` `` | The file of a merge in progress. |
   | `rebaseMerge` | `` `rebase-merge` `` | The folder of a rebase in progress. |
   | `rebaseApply` | `` `rebase-apply` `` | The folder of a rebase in progress. |
   | `gitAttributes` | `` `.gitattributes` `` | The attribute file of the repository. |
   | `sshDirectory` | `` `.ssh` `` | The folder of the SSH keys in the home folder. |

   The values are the legacy literals (`legacy/git/exec_utils.go:54`; `dirtyop.go:21`; `state_detection.go:191,203,208,211`; `lfs.go:13`; `environment.go:50`) and the pin marker of Part 2 Step 3.
2. `cast/git.md`: a new file with one table `## git`, in the form of `ling/cast/c_language.md:1-18`. The brief is `@brief The words of the git, git-lfs, and ssh command lines and outputs that titc reads.` Columns `name`, `value`, `description`. The rows, in order:

   | name | value | description |
   |---|---|---|
   | `gitVersion` | `` `git --version` `` | Exits with 0 when git runs. |
   | `sshVersion` | `` `ssh -V` `` | Exits with 0 when ssh runs. |
   | `lfsVersion` | `` `git-lfs version` `` | Exits with 0 when git-lfs runs. |
   | `lfsFilterProcess` | `` `git config --get filter.lfs.process` `` | Prints the LFS filter command when the LFS filters are registered. |
   | `head` | `` `git rev-parse HEAD` `` | Prints the full hash of HEAD. Fails when the repository has no commit. |
   | `headShort` | `` `git rev-parse --short HEAD` `` | Prints the short hash of HEAD. |
   | `status` | `` `git status --porcelain=v2` `` | Prints one record for each changed, untracked, or unmerged file. |
   | `branch` | `` `git symbolic-ref --short HEAD` `` | Prints the branch of HEAD. Fails when HEAD is detached. |
   | `remote` | `` `git remote` `` | Prints one line for each remote. |
   | `upstream` | `` `git rev-parse @{u}` `` | Fails when the branch has no upstream. |
   | `revParse` | `` `git rev-parse` `` | Fails when the next argument names no object. |
   | `countLeftRight` | `` `git rev-list --left-right --count` `` | Prints the counts of the commits only on the left and only on the right of the next range. |
   | `countHead` | `` `git rev-list --count HEAD` `` | Prints the count of the commits of HEAD. |
   | `configGet` | `` `git config --get` `` | Fails when the next key has no value. |
   | `upstreamRange` | `` `HEAD...@{u}` `` | The range of HEAD and its upstream. |
   | `headRange` | `` `HEAD...` `` | The start of the range of HEAD and a remote branch. |
   | `remoteBranchPrefix` | `` `refs/remotes/origin/` `` | The start of the remote branch of a local branch. |
   | `branchConfigPrefix` | `` `branch.` `` | The start of the config key of a branch. |
   | `remoteConfigSuffix` | `` `.remote` `` | The end of the config key of the remote of a branch. |
   | `changedRecordTypes` | `` `12?` `` | The first characters of the status records of a changed or untracked file. |
   | `unmergedRecordTypes` | `` `u` `` | The first character of the status record of an unmerged file. |
   | `lfsFilter` | `` `filter=lfs` `` | The attribute of a file that LFS stores. |
   | `detachedBranch` | `` `DETACHED` `` | The branch text when HEAD is detached and no time travel marker names a branch. |
   | `sshKeyTypes` | `` `_rsa _ed25519 _ecdsa _dsa` `` | The name parts of an SSH private key, separated by spaces. |
   | `publicKeyExtension` | `` `.pub` `` | The extension of an SSH public key. |
   | `testSetupVariable` | `` `TIT_TEST_SETUP` `` | The environment variable that forces needsSetup for a test of the setup wizard. |
   | `testSetupValue` | `` `1` `` | The value of testSetupVariable that forces needsSetup. |

   Each command line and each text is a legacy literal (inventory, first table). `lfsVersion` runs the `git-lfs` binary, as legacy looks it up in `PATH` (`legacy/git/lfs.go:20`). `unmergedRecordTypes` is one character: each porcelain v2 record type is one character and a space, so the first character decides, as `u ` decides in legacy (`state_detection.go:183`).

**Validation:**
- Each value equals its legacy literal, or a cited Part 2 word.
- `git.md` has the column form of `ling/cast/c_language.md`.

### Step 4: The manifest wiring
**Scope:** `tit/cast/spell.md`.
**Action:**
1. `## index`: after the `@text` row, add `| @git | git.md |` (the `@cLanguage` row of `ling/cast/spell.md:25`). After the `@string` row, add `| @bimap | jam::Bimap<int> |` (the `@bimap` row of `jam/cast/spell.md:72`). Widen the column separators to the new cell widths.
2. `## output`: after the `@text:diagnostics` group, add the group of `ling/cast/spell.md:177-183` with `@git:git` in place of `@cLanguage:cLanguage`, and `- name: Git` in place of `- name: CLanguage`. The file cell is `@Text`.
3. `## output`: after the `@identifiers:identifiers` group whose file cell is `@Identifiers`, add five groups, one for each table of Step 2, in the order of Step 2. Each group is the group of `eve/cast/spell.md:126-136` with these changes, which are the cells of `jam/cast/spell.md:840-851`:
   - `@identifiers:<Table>` in place of `@identifiers:CommandId`, in both list cells;
   - `- name: <Table>` in place of `- name: CommandId`;
   - `- [description]: <the first sentence of the table's @brief>`;
   - `- base: @bimap` in place of `- base: @bimapIdentifier`;
   - `- valueType: juce::String` in place of `- valueType: juce::Identifier`;
   - `> > > - [list]: @code:name-entry` in place of `> > > - [list]: @code:identifier-entry`.
   `- keyType: int` and `> > - [list]: @code:enum-entry` stay.

**Validation:**
- After `cast cast/spell.md --debug` (Step 9), `Source/generated/Text.h` holds `struct Git` with 27 constants, and `Identifiers.h` holds `map::GitEnvironment`, `map::WorkingTree`, `map::Timeline`, `map::Operation`, and `map::Remote`, each a `jam::Bimap<int>` with an `enum value : int` (`jam/cast/code.cast:160-194`).
- No group of the LING or EVE spell is copied beyond the cited lines.

### Step 5: The `git` rows of `TITModel`
**Scope:** `tit/Source/TITModel.h`, `tit/Source/TITModel.cpp`.
**Action:**
1. Header: in the public section, after the destructor, declare `static const std::vector<std::pair<juce::Identifier, int>>& getDefaultGitValues();` and `static const juce::Array<juce::Identifier>& getGitTexts();`.
2. Source: at the end of the constructor body, add:

```cpp
    auto gitRow { getOrCreateChildWithName (jam::Format::toType (Id::git)) };

    createAndAddParameter<jam::ParameterText> (state, Id::directory, juce::String {}, jam::ParameterText::maxTextLength);
    createAndAddParameter<jam::Parameter<int>> (gitRow, Id::gitEnvironment, static_cast<int> (map::GitEnvironment::ready));
    createAndAddParameter<jam::Parameter<int>> (gitRow, Id::operation, static_cast<int> (map::Operation::notRepo));

    for (const auto& [id, value] : getDefaultGitValues())
        createAndAddParameter<jam::Parameter<int>> (gitRow, id, value);

    for (const auto& id : getGitTexts())
        createAndAddParameter<jam::ParameterText> (gitRow, id, juce::String {}, jam::ParameterText::maxTextLength);
```

   Then define:

```cpp
const std::vector<std::pair<juce::Identifier, int>>& TITModel::getDefaultGitValues()
{
    static const std::vector<std::pair<juce::Identifier, int>> defaultGitValues {
        { Id::workingTree, map::WorkingTree::clean },
        { Id::modifiedCount, 0 },
        { Id::timeline, map::Timeline::none },
        { Id::commitsAhead, 0 },
        { Id::commitsBehind, 0 },
        { Id::remote, map::Remote::noRemote },
        { Id::hasUpstream, 0 },
        { Id::isDetached, 0 },
        { Id::isTitTimeTravel, 0 },
        { Id::isUsingLfs, 0 },
        { Id::isLfsReady, 0 },
    };

    return defaultGitValues;
}

const juce::Array<juce::Identifier>& TITModel::getGitTexts()
{
    static const juce::Array<juce::Identifier> gitTexts { Id::currentBranch, Id::currentHash };

    return gitTexts;
}
```

   - Every row is a parameter, made at construction on the message thread, in the form of `end/Source/end/ENDModel.cpp:8-9` and `end/Source/plugins/Nexus.cpp:35-39` (Closed by a read, item 2). `createAndAddParameter` seeds the row of the tree with the default (`jam_Model.h:1052-1101`).
   - `directory` is a row of the root. The group of a row is the type of its tree (`jam_Model.cpp:178-185`): `jam::Format::toType (Id::tit)` for the root, and `jam::Format::toType (Id::git)` for the `git` child.
   - `gitEnvironment` and `operation` are not in the table. The job always computes them (Step 7, item 3). The table holds the rows that take a default with no repository.
   - Each default is the first row of its table, 0, or the empty text: the zero values of the legacy state (`legacy/git/types.go:73-89`). A flag is 0 or 1 (`jam_ParameterBase.h:145`).
   - The two functions give the one list of the rows. The constructor and the job of Step 7 read them (MANIFESTO S).

**Validation:**
- The two functions read no member. Each one returns a table that it builds one time.
- The child has 15 rows, one for each identifier of Step 2 except `git`. The root has the row `directory`.
- New names and their families: `getDefaultGitValues` (`getDefaultKeys`, `jam_gui/widgets/jam_TextEditor.h`; `getDefaultConfig`, `end/Source/config/ConfigModel.h:15`); `getGitTexts` (the same family, with the noun `text` of `jam::ParameterText`); `gitRow` (`terminalRow`, Part 7 Step 4; `overlayRow`, `ENDModel.cpp:8`).

### Step 6: `TITGitProcessor`, the header
**Scope:** `tit/Source/TITGitProcessor.h` (new).
**Action:** `#pragma once`, `#include <JuceHeader.h>`, `#include "generated/Generated.h"`, `#include "TITModel.h"` (`ling/Source/Manifest.h:1-4`). Declare:

```cpp
class TITGitProcessor
{
public:
    explicit TITGitProcessor (TITModel& modelToUse);

    static juce::File getRepositoryDirectory (const juce::File& directory);

private:
    void addJob();
    void setGit (const juce::File& directory) const;
    void setRepositoryGit (const juce::File& directory) const;
    void setTimeline (int commitsAhead, int commitsBehind, bool hasTimeline) const;
    void storeValue (const juce::Identifier& id, int value) const;
    void storeValue (const juce::Identifier& id, const juce::String& text) const;

    static int getGitEnvironment (const juce::File& directory);
    static int getOperation (const juce::File& gitDirectory, const juce::StringArray& records, bool isOnBranch);
    static juce::String getCurrentBranch (const juce::File& gitDirectory, const std::optional<juce::String>& branch);
    static void removeStaleMarkers (const juce::File& gitDirectory, const std::optional<juce::String>& branch, const std::optional<juce::String>& head);
    static std::pair<int, int> getCommitCounts (const juce::File& directory, const juce::String& branch);
    static std::pair<int, int> getLeftRightCounts (const juce::File& directory, const juce::String& range);
    static bool hasSshKey();
    static bool hasUpstream (const juce::File& directory, const juce::String& branch);
    static bool isLfsReady (const juce::File& directory);
    static std::optional<juce::String> getOutput (const juce::File& directory, const jam::Strings& arguments);
    static jam::Strings getArguments (std::string_view commandLine, const juce::String& argument);

    template <size_t numEntries>
    static int getFirstMatch (const std::array<std::pair<bool, int>, numEntries>& entries, int defaultValue)
    {
        for (const auto& [isMatch, value] : entries)
            if (isMatch) return value;

        return defaultValue;
    }

    static constexpr int threadCount { 1 };
    static constexpr int branchLine { 0 };
    static constexpr int pinnedHashLine { 2 };

    const juce::Identifier gitTag { jam::Format::toType (Id::git) };
    TITModel& model;
    juce::ThreadPool pool { juce::ThreadPool::Options{}.withNumberOfThreads (threadCount) };
    jam::Model::ParameterAttachment directoryAttachment;

    JUCE_DECLARE_NON_COPYABLE_WITH_LEAK_DETECTOR (TITGitProcessor)
};
```

- The class is not a `juce::ValueTree::Listener`, as `EVETerminalProcessor` is not (`eve/Source/EVETerminalProcessor.h:30-32`). Its one reaction is `directoryAttachment`.
- `directoryAttachment` is the last member. It is destroyed first, so no callback reaches a pool that is gone. The class needs no destructor (CODING.md, RAII).
- `gitTag` is the form of `terminalTag` (Part 7 Step 4; `EVETerminalProcessor.h:344-359`).
- The two `storeValue` functions are the two forms of `jam::Model::storeValue` (`jam_Model.h:1013-1031`) for the `git` group. The `set…` functions replace rows of the Model (NAMES.md Rule 1, `set`).
- `getFirstMatch` is the `findFirst` form of MANIFESTO E: a loop with a result return. It serves the two priority tables of Step 7.
- `branchLine` and `pinnedHashLine` are the line numbers of the markers: the branch is line 1 of the time travel marker (`legacy/git/state_helpers.go:65-70`) and of the pin marker; the pinned hash is line 3 of the pin marker (Part 2 Step 3, item 4). Two functions read `branchLine`, so it has class scope (CODING.md, narrowest scope).

**Validation:**
- New names and their families: `TITGitProcessor` (Closed by a read, item 1); `modelToUse` (Part 7 Step 4); `addJob`, `pool`, `threadCount`, `model` (`ling/Source/Manifest.h:49,68,71,74`); `getRepositoryDirectory` (`getCacheDirectory`, `ling/Source/LINGModel.h:72`; `getBuildDirectory`, `ling/Source/CompilerManifest.cpp:250`); `setGit`, `setRepositoryGit`, `setTimeline` (NAMES.md Rule 1, `set`, with the noun of the child and of the axis); `storeValue` (`jam::Model::storeValue`, `jam_Model.h:1014`); `gitTag` (`terminalTag`, Part 7 Step 4); `directoryAttachment` (`winsizeAttachment`, `eve/Source/EVETerminalProcessor.cpp:68`); `getGitEnvironment`, `getOperation`, `getCurrentBranch` (the SPEC and legacy nouns of Step 2); `removeStaleMarkers` (Part 2 Step 3, item 5: "TIT removes a stale marker"); `getCommitCounts`, `getLeftRightCounts` (the legacy `CommitsAhead`, `CommitsBehind`, and `--left-right --count`); `hasSshKey` (`sshKeyExists`, `legacy/git/environment.go:42`); `hasUpstream`, `isLfsReady` (Step 2 words); `getOutput` (the legacy `Output()`, `state_detection.go:16`); `getArguments` (`jam::ChildProcess::getArguments`); `getFirstMatch` (`findFirst`, MANIFESTO E); `branchLine`, `pinnedHashLine` (the marker lines of Part 2 Step 3).
- No member stores repository state. The class holds the model reference, the pool, the group tag, and one attachment.

### Step 7: `TITGitProcessor`, the source
**Scope:** `tit/Source/TITGitProcessor.cpp` (new).
**Action:** `#include "TITGitProcessor.h"`. Then define each function.

1. The constructor, the job, and the two stores. The attachment is the form of `eve/Source/EVETerminalProcessor.cpp:68-72`.

```cpp
TITGitProcessor::TITGitProcessor (TITModel& modelToUse)
    : model (modelToUse)
    , directoryAttachment (*model.getParameter<jam::ParameterText> (jam::Format::toType (Id::tit), Id::directory), [this] (const juce::var&) { addJob(); })
{
}

void TITGitProcessor::addJob()
{
    pool.addJob ([this, directory = juce::File { model.state[Id::directory].toString() }] { setGit (directory); });
}

void TITGitProcessor::storeValue (const juce::Identifier& id, int value) const
{
    model.storeValue (gitTag, id, value);
}

void TITGitProcessor::storeValue (const juce::Identifier& id, const juce::String& text) const
{
    model.storeValue (gitTag, id, text.toRawUTF8(), static_cast<int> (text.getNumBytesAsUTF8()));
}
```

   - The application writes the `directory` row of the tree (Step 8). The Model gives that value to the parameter and notifies its listeners (`jam_Model.cpp:14-21,364-378`). The attachment runs its callback on the message thread (`jam_Model.cpp:220-240`), so `addJob` reads the row of the tree there.
   - The job gets the folder as a `juce::File` built from a string on the message thread. No tree of the Model reaches the pool thread.
   - The job stores each row with `storeValue`, on the pool thread: an atomic store and a dirty mark, with no lock, no message, and no tree (`eve/SPEC.md:802-805`; `jam_ParameterBase.h:174-180`). The flush of the Model writes the changed rows to the tree on the message thread (`jam_Model.cpp:291-303`), and the views react in `valueTreePropertyChanged`.
   - The pool has one thread, so each `jam::ParameterText` row has one producer thread (`jam_ParameterText.h:5-8`).

2. The repository folder (`legacy/git/state_helpers.go:12-46`):

```cpp
juce::File TITGitProcessor::getRepositoryDirectory (const juce::File& directory)
{
    for (auto folder { directory };; folder = folder.getParentDirectory())
    {
        if (folder.getChildFile (files::gitDirectory).exists()) return folder;
        if (folder.isRoot()) return directory;
    }
}
```

   The loop tests the folder, then each parent up to the root, as legacy tests the folder and then each parent up to the root.

3. The rows of one detection:

```cpp
void TITGitProcessor::setGit (const juce::File& directory) const
{
    const auto gitEnvironment { getGitEnvironment (directory) };
    const auto gitDirectory { directory.getChildFile (files::gitDirectory) };
    const auto isRepository { gitEnvironment == map::GitEnvironment::ready and gitDirectory.exists() };
    const auto isDirtyOperation { isRepository and gitDirectory.getChildFile (files::dirtyOperationMarker).exists() };

    storeValue (Id::gitEnvironment, gitEnvironment);

    if (isRepository and not isDirtyOperation)
    {
        setRepositoryGit (directory);
    }
    else
    {
        storeValue (Id::operation, isDirtyOperation ? map::Operation::dirtyOperation : map::Operation::notRepo);

        for (const auto& [id, value] : TITModel::getDefaultGitValues())
            storeValue (id, value);

        for (const auto& id : TITModel::getGitTexts())
            storeValue (id, juce::String {});
    }
}
```

   - The order is the legacy order: the environment (`legacy/app/app_constructor.go:46-50`), the repository (`state.go:35-38`), the dirty operation marker (`state.go:40-45`).
   - The function runs on the pool thread. It stores each row one time: the computed value, or the default of the Model's table (Closed by a read, item 4).

4. The environment (`legacy/git/environment.go:15-34`):

```cpp
int TITGitProcessor::getGitEnvironment (const juce::File& directory)
{
    const auto isTestSetup { juce::SystemStats::getEnvironmentVariable (text::Git::testSetupVariable, {}).compare (text::Git::testSetupValue) == 0 };
    const std::array<std::pair<bool, int>, 4> gitEnvironments { {
        { isTestSetup, map::GitEnvironment::needsSetup },
        { not getOutput (directory, jam::ChildProcess::getArguments (text::Git::gitVersion)).has_value(), map::GitEnvironment::missingGit },
        { not getOutput (directory, jam::ChildProcess::getArguments (text::Git::sshVersion)).has_value(), map::GitEnvironment::missingSsh },
        { not hasSshKey(), map::GitEnvironment::needsSetup },
    } };

    return getFirstMatch (gitEnvironments, map::GitEnvironment::ready);
}

bool TITGitProcessor::hasSshKey()
{
    const auto keyTypes (juce::StringArray::fromTokens (text::Git::sshKeyTypes, false));
    const auto keys (juce::File::getSpecialLocation (juce::File::userHomeDirectory).getChildFile (files::sshDirectory).findChildFiles (juce::File::findFiles, false));

    return std::any_of (keys.begin(), keys.end(), [&keyTypes] (const juce::File& key)
    {
        return not key.hasFileExtension (text::Git::publicKeyExtension)
           and std::any_of (keyTypes.begin(), keyTypes.end(), [&key] (const juce::String& keyType) { return key.getFileName().contains (keyType); });
    });
}
```

   The table order is the legacy priority: the test variable, then git, then ssh, then the key. A folder that does not exist gives no file, so no key (`environment.go:52-55`).

5. The rows of a repository (`legacy/git/state.go:47-156`, with the corrections of the inventory):

```cpp
void TITGitProcessor::setRepositoryGit (const juce::File& directory) const
{
    const juce::String changedRecordTypes { text::Git::changedRecordTypes };
    const auto gitDirectory { directory.getChildFile (files::gitDirectory) };
    const auto head { getOutput (directory, jam::ChildProcess::getArguments (text::Git::head)) };
    const auto branch { getOutput (directory, jam::ChildProcess::getArguments (text::Git::branch)) };
    const auto records (juce::StringArray::fromLines (getOutput (directory, jam::ChildProcess::getArguments (text::Git::status)).value_or (juce::String {})));
    const auto hasRemote { getOutput (directory, jam::ChildProcess::getArguments (text::Git::remote)).value_or (juce::String {}).isNotEmpty() };
    const auto isUsingLfs { directory.getChildFile (files::gitAttributes).loadFileAsString().contains (text::Git::lfsFilter) };
    const auto modifiedCount { static_cast<int> (std::count_if (records.begin(), records.end(), [&changedRecordTypes] (const juce::String& record) { return changedRecordTypes.containsChar (record[0]); })) };

    removeStaleMarkers (gitDirectory, branch, head);

    const auto operation { getOperation (gitDirectory, records, branch.has_value()) };
    const auto hasTimeline { operation == map::Operation::normal and hasRemote and head.has_value() };
    const auto [commitsAhead, commitsBehind] { hasTimeline ? getCommitCounts (directory, *branch) : std::pair<int, int> {} };

    storeValue (Id::operation, operation);
    storeValue (Id::workingTree, modifiedCount > 0 ? map::WorkingTree::dirty : map::WorkingTree::clean);
    storeValue (Id::modifiedCount, modifiedCount);
    storeValue (Id::remote, hasRemote ? map::Remote::hasRemote : map::Remote::noRemote);
    storeValue (Id::currentBranch, getCurrentBranch (gitDirectory, branch));
    storeValue (Id::currentHash, getOutput (directory, jam::ChildProcess::getArguments (text::Git::headShort)).value_or (juce::String {}));
    storeValue (Id::hasUpstream, hasRemote and branch.has_value() and hasUpstream (directory, *branch));
    storeValue (Id::isDetached, not branch.has_value());
    storeValue (Id::isTitTimeTravel, not branch.has_value() and gitDirectory.getChildFile (files::timeTravelMarker).exists());
    storeValue (Id::isUsingLfs, isUsingLfs);
    storeValue (Id::isLfsReady, isUsingLfs and isLfsReady (directory));
    setTimeline (commitsAhead, commitsBehind, hasTimeline);
}
```

   - Each row is stored one time, in the order of the rows of Step 5. A flag goes to the `int` form as 0 or 1.
   - With no timeline, the three timeline rows take `none`, 0, and 0: the defaults of Step 5.

   - One `status` run gives the count and, in `getOperation`, the conflicts.
   - A failed `status` gives no record, so the tree is clean, as legacy assumes Clean (`state_detection.go:17-19`).
   - `hasUpstream` runs only when the branch is known; legacy asks the config of `HEAD` when detached, and that key has no value (`state_parsing.go:18-26`).
   - `normal` implies a branch: the first-match table of `getOperation` gives `timeTraveling` first when the branch is unknown. So `*branch` is valid in the timeline call.
   - `isLfsReady` runs only when the repository uses LFS (`state.go:48-51`).

6. The operation, in the SPEC order (inventory, corrections 1 and 2):

```cpp
int TITGitProcessor::getOperation (const juce::File& gitDirectory, const juce::StringArray& records, bool isOnBranch)
{
    const juce::String unmergedRecordTypes { text::Git::unmergedRecordTypes };
    const auto isConflicted { std::any_of (records.begin(), records.end(), [&unmergedRecordTypes] (const juce::String& record) { return unmergedRecordTypes.containsChar (record[0]); }) };
    const std::array<std::pair<bool, int>, 5> operations { {
        { isConflicted, map::Operation::conflicted },
        { gitDirectory.getChildFile (files::mergeHead).exists(), map::Operation::merging },
        { gitDirectory.getChildFile (files::rebaseMerge).exists() or gitDirectory.getChildFile (files::rebaseApply).exists(), map::Operation::rebasing },
        { not isOnBranch, map::Operation::timeTraveling },
        { gitDirectory.getChildFile (files::pinMarker).exists(), map::Operation::pinned },
    } };

    return getFirstMatch (operations, map::Operation::normal);
}
```

   `removeStaleMarkers` runs before this function, so a pin marker that exists is valid. `exists()` is the legacy test, `os.Stat` (`state_detection.go:191,203,208,211`).

7. The stale markers and the branch text (`legacy/git/state_detection.go:188-199`; `state.go:114-146`; Part 2 Step 3, item 5):

```cpp
void TITGitProcessor::removeStaleMarkers (const juce::File& gitDirectory, const std::optional<juce::String>& branch, const std::optional<juce::String>& head)
{
    const auto pinMarker { gitDirectory.getChildFile (files::pinMarker) };
    const auto pin (juce::StringArray::fromLines (pinMarker.loadFileAsString().trim()));
    const auto isPinStale { not (branch.has_value() and head.has_value()
                                 and pin[branchLine].trim().compare (*branch) == 0
                                 and pin[pinnedHashLine].trim().compare (*head) == 0) };

    if (branch.has_value())
        gitDirectory.getChildFile (files::timeTravelMarker).deleteFile();

    if (isPinStale)
        pinMarker.deleteFile();
}

juce::String TITGitProcessor::getCurrentBranch (const juce::File& gitDirectory, const std::optional<juce::String>& branch)
{
    if (branch.has_value()) return *branch;

    const auto markerBranch { juce::StringArray::fromLines (gitDirectory.getChildFile (files::timeTravelMarker).loadFileAsString().trim())[branchLine].trim() };

    if (markerBranch.isNotEmpty()) return markerBranch;
    return text::Git::detachedBranch;
}
```

   - `deleteFile` returns true for a file that does not exist (doxygen `classjuce_1_1File`). So each call needs no test of existence.
   - `juce::StringArray::operator[]` gives the empty text for an index out of range. A marker with fewer lines is stale.
   - The pinned hash compares with the full hash of `rev-parse HEAD`. Part 12 writes the full hash, the hash of the History list (`legacy/git/types.go:93`).

8. The timeline (`legacy/git/state_detection.go:52-154`, with the fallback that cannot run removed):

```cpp
void TITGitProcessor::setTimeline (int commitsAhead, int commitsBehind, bool hasTimeline) const
{
    static constexpr int aheadBit { 1 };
    static constexpr int behindBit { 2 };
    static constexpr std::array timelines { map::Timeline::inSync, map::Timeline::ahead, map::Timeline::behind, map::Timeline::diverged };

    storeValue (Id::timeline, hasTimeline ? timelines.at (static_cast<size_t> ((commitsAhead > 0 ? aheadBit : 0) + (commitsBehind > 0 ? behindBit : 0))) : map::Timeline::none);
    storeValue (Id::commitsAhead, commitsAhead);
    storeValue (Id::commitsBehind, commitsBehind);
}

std::pair<int, int> TITGitProcessor::getCommitCounts (const juce::File& directory, const juce::String& branch)
{
    const juce::String remoteBranch { text::Git::remoteBranchPrefix + branch };

    if (getOutput (directory, jam::ChildProcess::getArguments (text::Git::upstream)).has_value())
        return getLeftRightCounts (directory, text::Git::upstreamRange);

    if (getOutput (directory, getArguments (text::Git::revParse, remoteBranch)).has_value())
        return getLeftRightCounts (directory, text::Git::headRange + remoteBranch);

    return { getOutput (directory, jam::ChildProcess::getArguments (text::Git::countHead)).value_or (juce::String {}).getIntValue(), 0 };
}

std::pair<int, int> TITGitProcessor::getLeftRightCounts (const juce::File& directory, const juce::String& range)
{
    static constexpr int numCounts { 2 };
    static constexpr int aheadIndex { 0 };
    static constexpr int behindIndex { 1 };

    const auto counts (juce::StringArray::fromTokens (getOutput (directory, getArguments (text::Git::countLeftRight, range)).value_or (juce::String {}), false));

    if (counts.size() == numCounts)
        return { counts[aheadIndex].getIntValue(), counts[behindIndex].getIntValue() };

    return {};
}
```

   - The table of four timelines replaces the legacy map keyed by two booleans (`state_detection.go:142-152`), with the same four results (`state_detection_test.go:10-18`).
   - A failed count gives 0 and 0, as legacy gives `InSync, 0, 0` (`state_detection.go:97-112,119-135`).
   - A branch with no remote branch counts every commit of HEAD as ahead (`state_detection.go:77-91`).

9. The upstream, LFS, and the child process (`legacy/git/state_parsing.go:18-26`; `lfs.go:18-27`; `ling/Source/CompilerManifest.cpp:147-161`):

```cpp
bool TITGitProcessor::hasUpstream (const juce::File& directory, const juce::String& branch)
{
    return getOutput (directory, getArguments (text::Git::configGet, text::Git::branchConfigPrefix + branch + text::Git::remoteConfigSuffix)).has_value();
}

bool TITGitProcessor::isLfsReady (const juce::File& directory)
{
    return getOutput (directory, jam::ChildProcess::getArguments (text::Git::lfsVersion)).has_value()
       and getOutput (directory, jam::ChildProcess::getArguments (text::Git::lfsFilterProcess)).value_or (juce::String {}).isNotEmpty();
}

std::optional<juce::String> TITGitProcessor::getOutput (const juce::File& directory, const jam::Strings& arguments)
{
    juce::WaitableEvent finished;
    int exitCode { EXIT_FAILURE };
    std::string output;
    jam::Subprocess subprocess;

    subprocess.launch (arguments,
                       directory,
                       [&finished, &exitCode] (int processExitCode, const std::string&)
                       {
                           exitCode = processExitCode;
                           finished.signal();
                       },
                       [&output] (std::string_view chunk, bool) { output.append (chunk); });

    finished.wait();

    if (exitCode == 0) return juce::String { output }.trim();
    return std::nullopt;
}

jam::Strings TITGitProcessor::getArguments (std::string_view commandLine, const juce::String& argument)
{
    auto arguments (jam::ChildProcess::getArguments (commandLine));

    arguments.add (argument);
    return arguments;
}
```

   - `getOutput` gives the output only for exit code 0, trimmed, as legacy trims (`state_parsing.go:9-16`). Each legacy reader tests the error first, so a failure is `std::nullopt`.
   - The chunks collect with no cap. The completion capture is capped (inventory, API facts), and a large `status` would lose records.
   - The read of each chunk is the OS boundary copy (MANIFESTO, Permitted copy 1). The pool thread is not a deadline thread.

**Validation:**
- Each function has no bail-out guard. Each early return is a result return.
- No `if` chain has more than three branches. The two priority chains are tables.
- No string compares with `==`.
- No function body is longer than 30 lines.
- No tree of the Model is read or written on the pool thread. The pool thread writes the Model only with `storeValue`. `model.state[Id::directory]` is read in the capture list, on the message thread.
- The processor has no `juce::ValueTree::Listener` base, no `setProperty`, no `setValuesFrom`, and no `juce::MessageManager::callAsync`.
- Each row of the `git` child is stored one time in each detection.
- New local names come from the legacy names of the same values: `head`, `branch`, `records`, `hasRemote`, `isUsingLfs`, `modifiedCount`, `operation`, `isConflicted`, `isPinStale`, `markerBranch`, `pin`, `keys`, `keyTypes`, `counts`, `remoteBranch`, `commitsAhead`, `commitsBehind`, `gitEnvironments`, `operations`, `timelines`, `isTestSetup`, `isRepository`, `isDirtyOperation`, `gitDirectory`, `folder`, `finished`, `exitCode`, `output`, `subprocess`, `arguments`.

### Step 8: The application starts the detection
**Scope:** `tit/Source/Main.cpp`.
**Action:**
1. After `#include "TITTerminalProcessor.h"`, add `#include "TITGitProcessor.h"`.
2. After the member `std::unique_ptr<TITTerminalProcessor> processor;` of Part 7 Step 6, add `std::unique_ptr<TITGitProcessor> gitProcessor;`. It comes after `model`, so it is destroyed first, and its attachment leaves a live parameter.
3. At the end of `runWindow` of Part 7 Step 7, after `model.addListener (this);`, add:

```cpp
        const auto directory { TITGitProcessor::getRepositoryDirectory (juce::File::getCurrentWorkingDirectory()) };

        directory.getChildFile (files::gitDirectory).getChildFile (files::indexLock).deleteFile();
        gitProcessor = std::make_unique<TITGitProcessor> (model);
        model.state.setProperty (Id::directory, directory.getFullPathName(), nullptr);
```

   - The order is the legacy start: find the repository, delete the stale lock, detect (`legacy/app/app_constructor.go:52-67`).
   - The object exists before the property changes, as LING creates `manifest` before it sets `Id::manifest` (`ling/Source/Main.cpp:399-402`). The application writes the row of the tree, the Model gives it to the parameter, and the attachment of the processor starts the job (Step 7, item 1).
   - The detection runs only in the window run. The command-line runs of Part 1 do not change.

**Validation:**
- `runWindow` stays within 30 lines and has no bail-out guard.
- New names and their families: `gitProcessor` (the member name of its class, as `manifest` is of `Manifest`, `ling/Source/Main.cpp:473`); `directory` (`Id::directory`).
- `titc --version`, `titc --help`, `titc --wrong`, and `titc < /dev/null` start no detection.

### Step 9: Build, run, and measure on macOS
**Scope:** the `tit/` project; fixture repositories outside every repository, for example in the session scratchpad.
**Owner:** ARCHITECT builds, or COUNSELOR on ARCHITECT's explicit request (CAROL.md Build). The fixture commands are git commands: ARCHITECT runs them, or gives the git instruction (CAROL.md Git).
**Action:**
1. Add the temporary log. In `Main.cpp`, add the first member `jam::debug::Log::Scope logScope { juce::File::getSpecialLocation (juce::File::userHomeDirectory).getChildFile ("titc.log") };`. In `valueTreePropertyChanged`, before the `terminal` dispatch, add:

```cpp
        if (tree.hasType (jam::Format::toType (Id::git)))
            jam::debug::Log::write (property.toString(), tree[property].toString());
```

2. Run `cast cast/spell.md --debug` from `tit/`.
3. For each fixture below, start `titc` in END in the named folder, wait one second, press Ctrl+C, and read `~/titc.log`. Delete the log before the next fixture. A row that the log does not name holds its default (Step 5). The numbers are the keys of Step 2.

   | # | Folder and commands | Expected rows in the log |
   |---|---|---|
   | 1 | an empty folder | no line |
   | 2 | the same folder, started as `TIT_TEST_SETUP=1 titc` | `gitEnvironment 1` |
   | 3 | `git init` | `operation 1`; `currentBranch` = the output of `git symbolic-ref --short HEAD` |
   | 4 | fixture 3, then one commit, then one new file that is not added | `operation 1`, `workingTree 1`, `modifiedCount 1`, `currentBranch`, `currentHash` = `git rev-parse --short HEAD` |
   | 5 | `git init --bare origin.git`; `git clone origin.git clone`; in `clone`, one commit and `git push -u origin HEAD`, then one more commit | `remote 1`, `hasUpstream 1`, `timeline 2`, `commitsAhead 1` |
   | 6 | fixture 5; `git clone origin.git other`; in `other`, one commit and `git push`; in `clone`, `git fetch` | `timeline 4`, `commitsAhead 1`, `commitsBehind 1` |
   | 7 | fixture 6; in `clone`, `git reset --hard @{u}` | `timeline 1`; no `commitsAhead` or `commitsBehind` line |
   | 8 | fixture 7; in `clone`, `git reset --hard HEAD~1` | `timeline 3`, `commitsBehind 1` |
   | 9 | fixture 7; in `clone`, `git switch -c topic` | `timeline 2`, `commitsAhead` = `git rev-list --count HEAD`; no `hasUpstream` line |
   | 10 | fixture 4 with the file committed; a branch `side` with a change to another file; on the first branch, `git merge --no-commit --no-ff side` | `operation 2` |
   | 11 | fixture 4 with the file committed; two branches with different changes to one line; `git merge` of the second | `operation 3`; no `workingTree` line, because an unmerged record starts with `u` and is not in `changedRecordTypes` |
   | 12 | fixture 4 with the file committed and a second commit; `GIT_SEQUENCE_EDITOR=true git rebase -i --exec false HEAD~1` | `operation 4`, `isDetached 1`, `currentBranch DETACHED`. Legacy gives `TimeTraveling` here (inventory, correction 1). |
   | 13 | fixture 4 with the file committed; `git checkout --detach` | `operation 5`, `isDetached 1`, `currentBranch DETACHED` |
   | 14 | fixture 13; write the branch name of fixture 3 and a newline to `.git/TIT_TIME_TRAVEL` | `operation 5`, `isDetached 1`, `isTitTimeTravel 1`, `currentBranch` = that branch |
   | 15 | fixture 14; `git switch -` | `operation 1`; `.git/TIT_TIME_TRAVEL` no longer exists |
   | 16 | fixture 5 after `git reset --hard @{u}`; write three lines to `.git/TIT_PIN`: the branch, any text, and the output of `git rev-parse HEAD` | `operation 8`; no `timeline` line |
   | 17 | fixture 16 with a different third line | `operation 1`; `.git/TIT_PIN` no longer exists |
   | 18 | fixture 4; create the empty file `.git/TIT_DIRTY_OP` | `operation 6` only; no `workingTree` line |
   | 19 | fixture 4; write `*.bin filter=lfs diff=lfs merge=lfs -text` to `.gitattributes` | `isUsingLfs 1`; `isLfsReady 1` only when `git-lfs version` runs and `git config --get filter.lfs.process` prints a value |
   | 20 | fixture 4; create the empty file `.git/index.lock` | `.git/index.lock` no longer exists |
   | 21 | a subfolder of fixture 4 | the rows of fixture 4 |

4. Remove the member `logScope` and the two lines of item 1. Delete `~/titc.log`.
5. Run `cast cast/spell.md --debug` a second time.

**Validation:**
- The build ends with exit code 0. `Identifiers.h` holds the 16 words of Step 2 and the five tables. `Text.h` holds `struct Git`. `Files.h` holds the ten names of Step 3.
- Each fixture gives the expected rows.
- The window shows the guide during each run. No fixture blocks the keys: the job runs on the pool thread.
- After item 4, `Main.cpp` holds no `jam::debug::Log` line and no `logScope` member (CODING.md, DIAGNOSTICS).
- The second `cast` run changes no generated file (`archetype/ARCHETYPE.md:478-479`).

### Step 10: Run on Windows
**Scope:** the `tit/` project on a Windows machine.
**Owner:** ARCHITECT (CAROL.md Build; `~/.config/WINDOWS-SETUP.md:11,13`).
**Action:** Do Step 9, items 1 to 5, in END on a Windows machine with fixtures 1, 4, 5, 13, and 21.
**Validation:**
- The rows of each fixture match Step 9.
- `gitEnvironment` stays at 0: `git --version` and `ssh -V` start through `CreateProcessW`, and the SSH key is in the `.ssh` folder of the profile folder.

## BLESSED Alignment

- **B — Bound.** The application owns the processor. The processor owns its pool and its attachment. Each `jam::Subprocess` lives in one call of `getOutput`. The processor is destroyed before the model, and the attachment removes itself from its parameter.
- **L — Lean.** Each function is one step of the legacy detection. The two priority chains are tables. Nothing streams, and no field without a reader is carried.
- **E — Explicit.** Every command line, file name, and text is a named cast constant. Every axis value is a named enum value.
- **S — SSOT.** One table of the Model gives the rows and their defaults. One `status` run gives the tree state and the conflicts. The Model holds the detected state; the processor holds none.
- **S — Stateless.** The job is a function of the folder and the repository. It keeps nothing between runs.
- **E — Encapsulation.** MVP: the processor writes the Model through the atomic lane, and the views read the tree on the message thread. No tree of the Model reaches the pool thread. Data flows one way: repository, job, parameter, flush, tree, listeners.
- **D — Deterministic.** The same repository gives the same rows. An equal row gives no event, because the flush writes only a value that differs.
- LANGUAGE.md: no override applies beyond the C++ / JUCE Lean adaptation.

## Risks / Open Questions

None open. The three corrections follow SPEC. The runtime facts of this part are in Steps 9 and 10.
