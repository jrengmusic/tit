# GitEnvironment Setup Wizard — Implementation Plan

**Objective:** First-time git environment setup wizard for machines with only git/ssh installed.

**Scope:**
- Detect git + SSH prerequisites
- Guide user through SSH key generation
- Auto-configure ssh-agent and ~/.ssh/config
- Display public key for provider registration
- Service-agnostic (GitHub, GitLab, Bitbucket, etc.)

**Out of Scope:**
- Automated git/ssh installation (show instructions only)
- Provider API integration
- Git author config (name/email for commits)

---

## Architecture

### New 5th State Axis: GitEnvironment

```go
type GitEnvironment int

const (
    Ready        GitEnvironment = iota  // git + ssh + key exists
    NeedsSetup                          // git + ssh exist, no key
    MissingGit                          // git not installed
    MissingSSH                          // ssh not installed
)
```

**State priority (updated):**
```
Priority 0: GitEnvironment     ← NEW (before everything)
Priority 1: Pre-flight checks
Priority 2: Operation
Priority 3: Remote
Priority 4: Timeline + WorkingTree
```

### Wizard Flow (6 Steps)

| Step | Component | Purpose |
|------|-----------|---------|
| 1 | Confirmation | Welcome message |
| 2 | Confirmation | Prerequisites status + install instructions |
| 3 | Input | Email for key comment |
| 4 | Console | ssh-keygen + ssh-add + config write |
| 5 | Confirmation | Display key (copied) + provider URLs |
| 6 | Confirmation | Setup complete |

**Component reuse:** All steps use existing ModeConfirmation, ModeInput, ModeConsole.

---

## Phase 1: GitEnvironment Type + Detection

**Files:** `internal/git/types.go`, `internal/git/environment.go`

**Work:**

1. Add `GitEnvironment` type to `types.go`:
```go
type GitEnvironment int

const (
    Ready        GitEnvironment = iota
    NeedsSetup
    MissingGit
    MissingSSH
)

func (e GitEnvironment) String() string {
    switch e {
    case Ready:
        return "ready"
    case NeedsSetup:
        return "needs_setup"
    case MissingGit:
        return "missing_git"
    case MissingSSH:
        return "missing_ssh"
    default:
        return "unknown"
    }
}
```

2. Create `environment.go` with detection:
```go
func DetectGitEnvironment() GitEnvironment {
    // Check git
    if !commandExists("git") {
        return MissingGit
    }
    
    // Check ssh
    if !commandExists("ssh") {
        return MissingSSH
    }
    
    // Check SSH key exists
    home, _ := os.UserHomeDir()
    rsaKey := filepath.Join(home, ".ssh", "id_rsa")
    ed25519Key := filepath.Join(home, ".ssh", "id_ed25519")
    
    if fileExists(rsaKey) || fileExists(ed25519Key) {
        return Ready
    }
    
    return NeedsSetup
}

func commandExists(cmd string) bool {
    _, err := exec.LookPath(cmd)
    return err == nil
}

func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}
```

**Test:**
```bash
./build.sh
# Temporarily rename ~/.ssh/id_rsa
# Run tit, verify DetectGitEnvironment() returns NeedsSetup
# Restore key, verify returns Ready
```

---

## Phase 2: Wizard Step Enum + State Tracking

**Files:** `internal/app/modes.go`, `internal/app/app.go`

**Work:**

1. Add `ModeSetupWizard` to modes.go:
```go
const (
    // ... existing modes ...
    ModeSetupWizard  // Git environment setup wizard
)
```

2. Add `SetupWizardStep` enum:
```go
type SetupWizardStep int

const (
    SetupStepWelcome SetupWizardStep = iota
    SetupStepPrerequisites
    SetupStepEmail
    SetupStepGenerate
    SetupStepDisplayKey
    SetupStepComplete
)
```

3. Add wizard state to Application struct:
```go
type Application struct {
    // ... existing fields ...
    
    // GitEnvironment state
    gitEnvironment    git.GitEnvironment
    setupWizardStep   SetupWizardStep
    setupEmail        string  // Email for key generation
}
```

**Test:**
```bash
./build.sh
# Verify build succeeds
# Add debug print in app.go to show gitEnvironment value
```

---

## Phase 3: App Startup Integration

**Files:** `internal/app/app.go`

**Work:**

1. In `NewApplication()` or `Init()`, check GitEnvironment FIRST:
```go
func (a *Application) Init() tea.Cmd {
    // Check git environment before anything else
    a.gitEnvironment = git.DetectGitEnvironment()
    
    if a.gitEnvironment != git.Ready {
        a.mode = ModeSetupWizard
        a.setupWizardStep = SetupStepWelcome
        return nil
    }
    
    // Existing git state detection...
    a.gitState = git.DetectState()
    // ...
}
```

2. In `View()`, handle ModeSetupWizard:
```go
case ModeSetupWizard:
    return a.renderSetupWizard()
```

3. Create placeholder `renderSetupWizard()`:
```go
func (a *Application) renderSetupWizard() string {
    return "Setup wizard placeholder - Step: " + fmt.Sprint(a.setupWizardStep)
}
```

**Test:**
```bash
./build.sh
# Temporarily rename ~/.ssh/id_rsa
# Run tit, verify shows "Setup wizard placeholder"
# Restore key, verify shows normal menu
```

---

## Phase 4: Step 1 — Welcome Confirmation

**Files:** `internal/app/setup_wizard.go` (new file)

**Work:**

1. Create `setup_wizard.go` with step rendering:
```go
func (a *Application) renderSetupWizard() string {
    switch a.setupWizardStep {
    case SetupStepWelcome:
        return a.renderSetupWelcome()
    default:
        return "Unknown step"
    }
}

func (a *Application) renderSetupWelcome() string {
    // Use existing confirmation dialog component
    // Title: "Welcome to TIT"
    // Explanation: "We'll set up SSH authentication for git.\nThis only needs to be done once per machine."
    // Button: "Continue"
}
```

2. Handle ENTER key to advance to next step:
```go
// In keyboard.go or setup_wizard.go
func (a *Application) handleSetupWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if a.setupWizardStep == SetupStepWelcome && msg.String() == "enter" {
        a.setupWizardStep = SetupStepPrerequisites
        return a, nil
    }
    return a, nil
}
```

**Test:**
```bash
./build.sh
# Remove SSH key, run tit
# Verify welcome screen appears
# Press ENTER, verify advances to step 2
```

---

## Phase 5: Step 2 — Prerequisites Check

**Files:** `internal/app/setup_wizard.go`

**Work:**

1. Render prerequisites status:
```go
func (a *Application) renderSetupPrerequisites() string {
    gitOK := git.CommandExists("git")
    sshOK := git.CommandExists("ssh")
    
    // Build status lines
    // ✓ Git found (git version 2.x.x)
    // ✗ SSH not found
    //
    // Install SSH:
    //   macOS:   brew install openssh
    //   Linux:   sudo apt install openssh-client
    //   Windows: Included with Git for Windows
    //
    // [Press ENTER to re-check]
}
```

2. ENTER re-checks prerequisites:
```go
if a.setupWizardStep == SetupStepPrerequisites && msg.String() == "enter" {
    a.gitEnvironment = git.DetectGitEnvironment()
    
    if a.gitEnvironment == git.MissingGit || a.gitEnvironment == git.MissingSSH {
        // Stay on this step, re-render with updated status
        return a, nil
    }
    
    // Prerequisites OK, advance
    a.setupWizardStep = SetupStepEmail
    return a, nil
}
```

**Test:**
```bash
./build.sh
# Verify git ✓ and ssh ✓ shown
# Temporarily break PATH to hide git, verify ✗ shown with instructions
# Fix PATH, press ENTER, verify advances
```

---

## Phase 6: Step 3 — Email Input

**Files:** `internal/app/setup_wizard.go`

**Work:**

1. Reuse existing input field rendering (like ModeInput):
```go
func (a *Application) renderSetupEmail() string {
    // Prompt: "Enter your email (used as key identifier):"
    // Input field with cursor
    // Hint: "This email appears in your public key"
}
```

2. Handle text input + ENTER to submit:
```go
if a.setupWizardStep == SetupStepEmail {
    if msg.String() == "enter" && a.inputValue != "" {
        a.setupEmail = a.inputValue
        a.inputValue = ""
        a.setupWizardStep = SetupStepGenerate
        return a, a.cmdGenerateSSHKey()
    }
    // Handle character input (existing pattern)
}
```

**Test:**
```bash
./build.sh
# Navigate to email step
# Type email, press ENTER
# Verify advances to generate step
```

---

## Phase 7: Step 4 — Generate Key (Console)

**Files:** `internal/app/setup_wizard.go`, `internal/git/ssh.go` (new)

**Work:**

1. Create `ssh.go` with key generation:
```go
func GenerateSSHKey(email string) error {
    home, _ := os.UserHomeDir()
    keyPath := filepath.Join(home, ".ssh", "id_rsa")
    
    // ssh-keygen -t rsa -b 4096 -C "email" -f keyPath -N ""
    cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096",
        "-C", email, "-f", keyPath, "-N", "")
    return cmd.Run()
}

func AddKeyToAgent() error {
    home, _ := os.UserHomeDir()
    keyPath := filepath.Join(home, ".ssh", "id_rsa")
    
    // Start agent if needed
    // ssh-add keyPath
    cmd := exec.Command("ssh-add", keyPath)
    return cmd.Run()
}

func WriteSSHConfig() error {
    // Append to ~/.ssh/config if not already configured
    // Host *
    //   AddKeysToAgent yes
    //   IdentityFile ~/.ssh/id_rsa
}

func GetPublicKey() (string, error) {
    home, _ := os.UserHomeDir()
    pubKeyPath := filepath.Join(home, ".ssh", "id_rsa.pub")
    data, err := os.ReadFile(pubKeyPath)
    return string(data), err
}
```

2. Create async command for generation:
```go
func (a *Application) cmdGenerateSSHKey() tea.Cmd {
    email := a.setupEmail
    
    return func() tea.Msg {
        buffer := ui.GetBuffer()
        buffer.Clear()
        
        buffer.Append("Generating SSH key...", ui.TypeStatus)
        if err := git.GenerateSSHKey(email); err != nil {
            return SetupErrorMsg{Step: "keygen", Error: err.Error()}
        }
        buffer.Append("✓ Created ~/.ssh/id_rsa", ui.TypeSuccess)
        
        buffer.Append("Adding to SSH agent...", ui.TypeStatus)
        if err := git.AddKeyToAgent(); err != nil {
            buffer.Append("⚠ Could not add to agent (add manually)", ui.TypeWarning)
        } else {
            buffer.Append("✓ Added to SSH agent", ui.TypeSuccess)
        }
        
        buffer.Append("Configuring ~/.ssh/config...", ui.TypeStatus)
        if err := git.WriteSSHConfig(); err != nil {
            buffer.Append("⚠ Could not write config (configure manually)", ui.TypeWarning)
        } else {
            buffer.Append("✓ Configured ~/.ssh/config", ui.TypeSuccess)
        }
        
        return SetupCompleteMsg{Step: SetupStepGenerate}
    }
}
```

3. Show console output, ENTER to continue:
```go
func (a *Application) renderSetupGenerate() string {
    // Render console output buffer
    // Show "Press ENTER to continue" when complete
}
```

**Test:**
```bash
./build.sh
# Remove existing SSH key first (backup!)
# Run through wizard to generate step
# Verify key created at ~/.ssh/id_rsa
# Verify ssh-add succeeded
# Verify ~/.ssh/config updated
```

---

## Phase 8: Step 5 — Display Key + Clipboard

**Files:** `internal/app/setup_wizard.go`

**Work:**

1. Render public key display:
```go
func (a *Application) renderSetupDisplayKey() string {
    pubKey, _ := git.GetPublicKey()
    
    // Copy to clipboard on first render
    if !a.setupKeyCopied {
        clipboard.WriteAll(pubKey)
        a.setupKeyCopied = true
    }
    
    // Render:
    // "Your public key (copied to clipboard):"
    // ┌──────────────────────────────────────────────────┐
    // │ ssh-rsa AAAAB3Nza... user@example.com            │
    // └──────────────────────────────────────────────────┘
    //
    // Add this key to your git provider:
    //   • GitHub:    github.com/settings/ssh/new
    //   • GitLab:    gitlab.com/-/user_settings/ssh_keys
    //   • Bitbucket: bitbucket.org/account/settings/ssh-keys
    //
    // [Press ENTER when key is added]
}
```

**Test:**
```bash
./build.sh
# Navigate to display key step
# Verify key shown
# Verify clipboard contains key (paste somewhere)
# Press ENTER, verify advances
```

---

## Phase 9: Step 6 — Complete

**Files:** `internal/app/setup_wizard.go`

**Work:**

1. Render completion screen:
```go
func (a *Application) renderSetupComplete() string {
    // ✓ Setup complete!
    //
    // TIT is ready. You can now init or clone repositories.
    //
    // [Press ENTER to continue]
}
```

2. ENTER transitions to normal TIT operation:
```go
if a.setupWizardStep == SetupStepComplete && msg.String() == "enter" {
    a.gitEnvironment = git.Ready
    a.mode = ModeMenu
    a.gitState = git.DetectState()
    a.menuItems = a.GenerateMenu()
    return a, nil
}
```

**Test:**
```bash
./build.sh
# Complete full wizard
# Verify transitions to normal menu (Init/Clone options)
```

---

## Phase 10: End-to-End Testing + Docs

**Work:**

1. Full flow test (clean machine simulation):
   - Backup ~/.ssh
   - Remove SSH keys
   - Run TIT, complete wizard
   - Verify can clone repo via SSH
   - Restore ~/.ssh

2. Update documentation:
   - ARCHITECTURE.md: Add GitEnvironment section
   - SPEC.md: Add setup wizard specification
   - README.md: Mention first-time setup

3. Update AGENTS.md if needed

**Test Checklist:**
- [ ] Fresh machine (no SSH key) → wizard appears
- [ ] Machine with SSH key → wizard skipped
- [ ] Git missing → shows install instructions
- [ ] SSH missing → shows install instructions
- [ ] Email input validates
- [ ] Key generated successfully
- [ ] Key copied to clipboard
- [ ] Wizard completion → normal TIT menu
- [ ] ESC at any step → appropriate behavior

---

## Files Summary

**New files:**
- `internal/git/environment.go` — GitEnvironment detection
- `internal/git/ssh.go` — SSH key generation utilities
- `internal/app/setup_wizard.go` — Wizard rendering + handlers

**Modified files:**
- `internal/git/types.go` — Add GitEnvironment type
- `internal/app/modes.go` — Add ModeSetupWizard, SetupWizardStep
- `internal/app/app.go` — Add wizard state, startup integration
- `internal/app/keyboard.go` — Add wizard key handlers

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| ssh-keygen not in PATH | Detect and show instructions |
| ssh-agent not running | Start agent or show instructions |
| ~/.ssh doesn't exist | Create directory with correct permissions |
| Existing key overwrite | Check before generating, skip if exists |
| macOS Keychain differences | Detect OS, use --apple-use-keychain flag |
| Windows SSH differences | Detect OS, adjust commands |

---

**End of Plan**
