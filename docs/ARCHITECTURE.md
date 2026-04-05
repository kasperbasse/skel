# Skel Architecture Guide

## High-Level Flow

```
User runs: skel restore [profile-name]
    ↓
cmd/cmd_restore.go (restore command handler)
    ├─ Load profile from disk
    ├─ Validate restorable data exists
    ├─ Check tool requirements via appdoctor
    ├─ If interactive: show TUI selector (cmd/tui/selectrestore.go)
    ├─ Run restore (internal/restore/)
    └─ Display results
```

## Package Organization

### `cmd/` — CLI Command Handlers
- **cmd_restore.go** — orchestrates the restore workflow
- **cmd_doctor.go** — validates tool availability (uses internal/app/doctor)
- **cmd_scan.go** — captures current system state
- **tui/** — Terminal UI components (Bubble Tea)

**Key pattern:** Commands are orchestrators. They call internal packages and print results.

### `internal/app/doctor/` — Tool Validation
- **checks.go** — builds check rows (tool label, status, fix message)
- **service.go** — maps profile sections to required tools (rules-based)
- **tools.go** — tool metadata (display name, install instructions)
- **runtime.go** — `exec.LookPath()` wrapper

**Dependency flow:**
```
cmd/cmd_restore.go 
    → internal/app/doctor/service.go (RequiredToolsForSections)
        → internal/app/doctor/checks.go (BuildChecks)
            → internal/app/doctor/tools.go (ToolDoctorInfo)
            → internal/app/doctor/runtime.go (CommandExists)
```

### `internal/profile/` — Profile Data Model
- **profile.go** — struct definitions (Homebrew, Editor, Git, Languages, etc.)
- Loaded from JSON files in `~/.skel/profiles/`

### `internal/restore/` — Restore Execution
- **restore.go** — iterates over sections and executes restore logic
- Calls section-specific restore functions (brewfile, git config, etc.)

### `internal/scanner/` — Profile Capture
- **scanner.go** — scans current system state
- Builds Profile struct from system data

## Key Design Decisions

### 1. **Section-Based Grouping** (`scanGroups` in `cmd/section_registry.go`)
Each section (homebrew, shell, git, editors, etc.) has metadata:
```go
type ScanGroup struct {
    Icon        string                              // 🍺
    Label       string                              // "Homebrew"
    RestoreKeys []string                            // ["homebrew", "mas"]
    ScanFunc    func(*Profile) error                // how to capture
    RestoreFunc func(*Profile, *Options) error      // how to restore
    ScanSummary func(*Profile) string               // "3 formulas, 1 cask"
    DryRun      func(*Profile, *Options)            // preview
}
```
**Why:** Decouples section UI metadata from restore logic.

### 2. **Tool Requirement Rules** (`internal/app/doctor/service.go`)
Instead of hardcoding which sections need which tools:
```go
var rules = []sectionToolRule{
    {
        Section: "homebrew",
        Tool:    "brew",
        Needed:  func(p *Profile) bool { 
            return len(p.Homebrew.Formulas) > 0 || ... 
        },
    },
    // ...
}
```
**Why:** Rules are data-driven. Easy to add new tools or sections without code changes.

### 3. **Blocked Sections** (`appdoctor.BlockedSectionTools`)
Separates "what's restorable" from "what's executable":
- `hasRestorableData()` — does profile have data for selected sections?
- `BlockedSectionTools()` — which sections need missing tools?
- UI can then deselect blocked sections by default

**Why:** User can still restore sections that don't have tool dependencies.

## Error Handling Pattern

```
User action
    ↓
cmd/ function (returns error)
    ├─ If critical error → enhanceError() adds context
    └─ Return error up to cobra
    
cobra catches error and prints to stderr
```

**Why this matters:** Errors bubble up cleanly. No panic(), no silent failures.

## Testing Strategy

### Unit Tests (no side effects)
- `internal/app/doctor/checks_test.go` — test check building
- `internal/app/doctor/runtime_test.go` — test tool detection
- `internal/profile/profile_test.go` — test profile parsing

### Integration Tests (invoke real functions, check results)
- `cmd/root_test.go` — test CLI argument parsing
- May use temp directories for file I/O

### What NOT to Test
- TUI interactions (too complex, use manual testing)
- OS-specific behavior like `exec.LookPath` (environment-dependent)

## Adding a New Section

To add a new restore section (e.g., "databases"):

1. **Define ScanGroup in `cmd/section_registry.go`:**
   ```go
   {
       Icon:     "🗄️",
       Label:    "Databases",
       RestoreKeys: []string{"databases"},
       ScanFunc:   scanDatabases,
       RestoreFunc: restoreDatabases,
       ScanSummary: func(p *Profile) string { 
           if p.Databases == nil { return "" }
           return fmt.Sprintf("%d configs", len(p.Databases))
       },
   }
   ```

2. **Add Profile field in internal/profile/profile.go:**
   ```go
   type Profile struct {
       Databases map[string]string `json:"databases"`
   }
   ```

3. **Implement scanners/restores (internal/scanner/, internal/restore/):**
   ```go
   func scanDatabases(p *Profile) error { ... }
   func restoreDatabases(p *Profile, opts *Options) error { ... }
   ```

4. **Add tool rule in internal/app/doctor/service.go (if needed):**
   ```go
   {
       Section: "databases",
       Tool:    "psql",  // if postgres config restoration needs psql
       Needed: func(p *Profile) bool { return len(p.Databases) > 0 },
   }
   ```

5. **Add tests** for the scanner and restorer.

## Common Workflows

### Running in Interactive Mode
```
skel restore
    ├─ Load profile
    ├─ Show TUI checklist
    │  └─ User toggles sections (X key)
    │  └─ User confirms (enter)
    └─ Execute selected sections
```

### Running with --only flag
```
skel restore --only shell,git
    ├─ Parse flag → opts.Sections = {shell: true, git: true}
    ├─ Skip interactive (onlyStr != "")
    ├─ Restore only shell + git sections
```

### Dry-Run Mode
```
skel restore --dry-run
    ├─ Load profile
    ├─ Call printDryRun() for each section
    │  └─ Sections print what WOULD happen (e.g., "Would install 3 packages")
    └─ Exit (don't actually restore)
```

## Performance Considerations

- **Profile load:** JSON parse, minimal cost
- **Tool checks:** `exec.LookPath()` calls (fixed cost, ~7 tools)
- **TUI rendering:** Bubble Tea, fast (terminal-based, no browser)
- **Restore execution:** Depends on package manager (could be slow)

No major bottlenecks identified. Lazy loading not needed.

## Future Improvements

1. **Parallel tool checks** — check all tools concurrently
2. **Profile versioning** — handle schema migrations
3. **Rollback capability** — save state before restore
4. **Selective dry-run per section** — not just global `--dry-run`
5. **Config file** — `~/.skelrc` for default options

