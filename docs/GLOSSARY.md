# Skel Glossary

## Domain Terms

### Profile
A saved snapshot of your macOS configuration. Contains sections like Homebrew packages, shell config, git settings, etc.
- Stored as JSON in `~/.skel/profiles/`
- Loaded and modified during `scan` and `restore` commands

### Section
A logical group of related config (e.g., "homebrew", "shell", "git").
- Each section is restorable independently (with `--only`)
- Blocked if required tools are missing

### Restorable Sections
Sections that have actual data to restore. Not all sections may have data in a profile.
- "Homebrew" is restorable if profile.Homebrew.Formulas is not empty
- "Shell" is restorable if shell config was saved

### Blocked Section
A section that *can't* be restored because required tools are missing.
- Example: Can't restore "homebrew" section if `brew` command isn't installed
- UI shows blocked sections with ⚠ icon in interactive mode

### Section Keys / RestoreKeys
The keys that identify a section in the Sections map.
- Defined per `ScanGroup` in `cmd/section_registry.go`
- Example: ScanGroup for Homebrew has RestoreKeys = ["homebrew", "mas"]
- Both "homebrew" and "mas" are blocked if "brew" is missing

---

## Technical Terms

### ScanGroup
Metadata + functions for a restore section.
```go
type ScanGroup struct {
    Icon        string  // 🍺 for Homebrew
    Label       string  // "Homebrew"
    RestoreKeys []string  // ["homebrew", "mas"]
    ScanFunc    func(...)  // how to capture this section
    RestoreFunc func(...)  // how to restore this section
}
```
Defined in `cmd/section_registry.go` and used throughout the app.

### Tool Requirement Rule
A rule that says: "If profile has data X, tool Y is required to restore it".
```go
{
    Section: "homebrew",
    Tool:    "brew",
    Needed: func(p *Profile) bool { 
        return len(p.Homebrew.Formulas) > 0 
    },
}
```
See `internal/app/doctor/service.go`.

### Check (Doctor Check)
A validation row showing:
- Tool name (label)
- Whether it's installed (OK)
- How to install it if missing (Fix)

Example output:
```
✓ Homebrew
✗ Node.js
  → brew install node
```

### Command / Validator Command
A shell command that `exec.LookPath()` searches for in PATH.
- Usually the tool name (e.g., "brew", "git")
- Sometimes different (e.g., "pip3" validates Python)

---

## UI Terms

### Bubble Tea
A terminal UI framework (go-tea/bubbletea).
- Used for interactive selection in `cmd/tui/`
- Provides keyboard input, rendering, state management

### DRY-RUN Mode
Preview mode: show what *would* be restored without actually making changes.
- Each section prints what it would do (e.g., "Would install 3 packages")
- Use: `skel restore --dry-run`

### Interactive Mode
User is prompted to select sections to restore via keyboard.
- Requires TTY (terminal)
- Alternative: non-interactive mode skips UI and restores all sections

### --only Flag
Limit restore to specific sections.
- Example: `skel restore --only shell,git`
- Skips interactive UI if specified
- Used to restore without user input

---

## File/Directory Terms

### Profile Directory

`~/.skel/profiles/` - where profiles are stored as JSON files.

### Profile Name
The filename (without .json) of a profile.
- Example: `default.json` → profile name is "default"
- Passed to `skel restore default`

### Dry-Run Output
Preview text showing what would be installed/changed.
- Printed to stdout before actual restore begins

---

## Error Handling Terms

### Error Wrapping
Using `%w` format verb to preserve error chain.
```go
return fmt.Errorf("restore failed: %w", err)
```
Allows caller to unwrap errors and find root cause.

### Error Enhancement
Adding context to an error before returning (via `enhanceError()`).
- Makes error messages more user-friendly
- Adds suggestions for fixing the problem

---

## Common Patterns

### Early Return
Exiting a function as soon as you know it will fail or succeed.
```go
if someCondition {
    return defaultValue
}
// Continue with main logic
```
Reduces nesting and improves readability.

### Callback (Closure)
A function passed as a parameter that captures variables from outer scope.
```go
RequiredToolsForSections(p, func(section string) bool {
    return opts.Sections[section]  // captures opts from outer scope
})
```

### Nil Check
Verifying a pointer isn't nil before using it.
```go
if g.ScanSummary != nil {
    summary := g.ScanSummary(p)
}
```
Required in Go; no automatic null checking.

### Struct Literal
Creating a struct with named fields.
```go
SelectItem{
    Icon:     g.Icon,
    Label:    g.Label,
    Selected: true,
}
```
Self-documenting; easier to read than positional arguments.

---

## Command Glossary

### `skel scan`
Capture current Mac configuration → save to profile file.

### `skel restore [profile-name]`
Load a profile and restore it (interactive or non-interactive).

### `skel doctor [profile-name]`
Validate that all tools needed by profile are installed.

### `skel list`
Show all saved profiles.

### `skel show [profile-name]`
Display profile contents (what's in it).

### `skel restore --dry-run`
Preview what would be restored (no actual changes).

### `skel restore --only SECTIONS`
Restore only specific sections (comma-separated).

