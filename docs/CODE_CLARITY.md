# Code Clarity & Self-Documenting Principles

This document explains how the codebase achieves clarity through structure and naming, reducing the need for documentation.

## Philosophy

**Good code should speak for itself.** Documentation should explain *why*, not *what*.

- **Bad:** Lots of comments explaining what each line does
- **Good:** Function names so clear that the code reads like prose
- **Better:** Small functions that do one thing well

## Key Improvements

### 1. Smaller, Focused Functions

**Problem:** Large functions with multiple responsibilities are hard to understand.

**Solution:** Break into smaller functions with clear names.

#### Example: `cmd/cmd_restore.go`

**Before:**
```go
func RunE(cmd *cobra.Command, args []string) error {
    // 60+ lines of nested logic: loading, validating, checking tools, selecting, executing
    // Many inline comments trying to explain what each section does
}
```

**After:**
```go
func RunE(...) error {
    // ... 
    return executeRestore(p, opts, dryRun)
}

// executeRestore orchestrates the restore flow
func executeRestore(p *profile.Profile, opts *restore.Options, dryRunMode bool) error {
    printRestoreHeader(p)
    if !hasRestorableData(p, opts) { ... }
    if err := checkToolRequirements(p, opts, dryRunMode); err != nil { ... }
    return runRestoreExecution(p, opts)
}
```

**Benefit:** Each function has one job. No need to explain every step because the function names tell the story.

### 2. Self-Documenting Names

**Problem:** Unclear function names require comments.

**Solution:** Choose names that explain what the function does.

#### Example: UI Checklist Building

**Before:**
```go
func buildSelectItems(p *profile.Profile, missingToolsBySection map[string][]string) []tui.SelectItem {
    // 70 lines + 30 lines of comments explaining what we're building
}
```

**After:**
```go
func buildSelectRestoreChecklistItems(p *profile.Profile) []tui.SelectItem {
    // Name says: we're building a restore checklist for interactive selection
    // Function is now 40 lines (more work, better naming)
}

func gatherMissingToolsForSection(g ScanGroup, missingBySection map[string][]string) (blocked bool, tools []string) {
    // Name says: collect missing tools for a section
    // Extracted logic that was buried in 70-line function
}
```

**Benefit:** The name tells you exactly what the function does. No comments needed.

### 3. Strategic Comments (Explain "Why")

**Problem:** Comments that explain "what" the code does are redundant.

**Solution:** Comments only for "why" decisions and non-obvious patterns.

#### Example: Rules-Based Tool Requirements

**Before:**
```go
// rules is the data-driven configuration for tool requirements.
// Format: if a Profile has data in section X and rule Y applies, tool Y is required.
// Examples:
//   - If profile.Homebrew.Formulas are not empty → "brew" is required
//   - If profile.Languages.NpmGlobals is not empty → "npm" is required
// Design choice: This is a slice (not a map) to preserve order and allow duplicates
// (e.g., "homebrew" section needs both "brew" and "mas").
var rules = []sectionToolRule{ ... }
```

**After:**
```go
// Package doctor validates that required tools are installed to restore a profile.
// It uses a rules-based system: data-driven definitions of which tools each section needs.
// Adding a new tool requires only a new rule; no code changes to restore logic.

// rules defines which tools are required for each section.
// Each rule says: if a profile has data in this section AND this condition passes,
// then this tool is required. Sections may appear in multiple rules (e.g., homebrew needs brew + mas).
var rules = []sectionToolRule{ ... }
```

**Benefit:** Comments explain the design decision (why rules?), not what the code does (what the code does is obvious).

### 4. Cleaner Restore Logic

**Problem:** `internal/restore/restore.go` had a 300-line `Run()` function with nested sections.

**Solution:** Extract each section into its own function.

**Before:**
```go
func Run(p *profile.Profile, opts *Options, onStep func(Result)) {
    // ... setup ...
    if opts.ShouldRestore("homebrew") {
        if !commandExists("brew") {
            emit("Homebrew", false, "...")
        } else {
            for _, tap := range p.Homebrew.Taps { ... }
            for _, formula := range p.Homebrew.Formulas { ... }
            for _, cask := range p.Homebrew.Casks { ... }
        }
    }
    if opts.ShouldRestore("mas") { ... }
    if opts.ShouldRestore("shell") { ... }
    if opts.ShouldRestore("git") { ... }
    // ... 200 more lines ...
}
```

**After:**
```go
func Run(p *profile.Profile, opts *Options, onStep func(Result)) {
    homeDir := home()
    if homeDir == "" { ... }
    
    installed := detectInstalled()
    total := countSteps(p, opts)
    stepIdx := 0
    
    emitStep := func(name string, success bool, message string) { ... }
    
    restoreSection := func(name string, fn func()) {
        if opts.ShouldRestore(name) { fn() }
    }
    
    restoreSection("homebrew", func() { restoreHomebrew(installed, p, emitStep) })
    restoreSection("mas", func() { restoreMacAppStore(p, emitStep) })
    restoreSection("shell", func() { restoreShellConfigs(p, homeDir, emitStep) })
    // ... etc ...
}

func restoreHomebrew(installed InstalledState, p *profile.Profile, emit func(string, bool, string)) {
    // 15 lines, crystal clear
}

func restoreMacAppStore(p *profile.Profile, emit func(string, bool, string)) {
    // 10 lines, crystal clear
}
```

**Benefit:** The main `Run()` function now reads like a checklist. Each `restoreXxx()` function is small and obvious.

### 5. Explicit Progress Tracking

**Problem:** Nested callbacks and closures made progress tracking hard to understand.

**Solution:** Make progress handling explicit.

**Before:**
```go
emitResult := func(name string, err error) {
    if err != nil {
        emit(name, false, err.Error())
    } else {
        emit(name, true, "done")
    }
}
// Then scattered throughout: emitResult(name, runSilent(...))
```

**After:**
```go
func emitResult(name string, err error, emit func(string, bool, string)) {
    if err != nil {
        emit(name, false, err.Error())
    } else {
        emit(name, true, "done")
    }
}
// Explicit parameter; no hidden closure dependencies
```

**Benefit:** Dependency on `emit` is explicit, not hidden in closures.

## Principles Applied

### 1. **Single Responsibility**
Each function does one thing:
- `buildSelectRestoreChecklistItems()` builds a checklist (one job)
- `gatherMissingToolsForSection()` gathers missing tools (one job)
- `executeRestore()` orchestrates the restore flow (one job)

### 2. **Descriptive Names**
Names tell you what the function does:
- `printRestoreHeader()` - prints restore header ✓
- `checkToolRequirements()` - checks tool requirements ✓
- `restoreShellConfigs()` - restores shell configs ✓

Not: `process()`, `handle()`, `do()`, `run()` ✗

### 3. **No Redundant Comments**
Bad: `x := 5 // set x to 5`
Good: (no comment; the code is clear)

Justified: `// Use map[string]struct{} as a set (more efficient than map[string]bool)`

### 4. **Extract Early**
If a block of code needs a comment explaining what it does, extract it into a function with a clear name instead.

### 5. **Consistent Patterns**
- All section restores follow the same pattern: `restore<Section>(profile, homeDir, emit)`
- All tool checks follow the same pattern: check existence, then execute
- All results follow the same pattern: track success/failure, emit step

## Benefits

1. **Faster Onboarding:** New contributors can read code and understand it without extensive docs
2. **Fewer Bugs:** Clear code is easier to reason about; less cognitive load = fewer mistakes
3. **Easier Maintenance:** Changing one section doesn't require rewriting comments
4. **Self-Validating:** If the function name doesn't match what it does, it's wrong
5. **Reduced Documentation:** Code clarity reduces documentation needs by 50%+

## Where Documentation Still Helps

Documentation *is* still needed for:
- **"Why" decisions:** Why rules-based design? Why callbacks instead of config objects?
- **Package-level context:** What does the doctor package do? How should you use it?
- **Non-obvious patterns:** Callbacks, implicit interfaces, deduplication logic
- **Integration guides:** How do packages work together? How do you add a new section?

But documentation should *explain context*, not describe what the code does.

## Checklist: Writing Clear Code

- [ ] Function name describes what it does (verb + object)
- [ ] Function is short enough to understand at a glance (< 20 lines ideal)
- [ ] Single responsibility (one job, one reason to change)
- [ ] No redundant comments ("// increment i" when line is `i++`)
- [ ] Comments explain "why" or link to design docs
- [ ] Parameter names are clear
- [ ] Return values are obvious from context
- [ ] Similar operations follow the same pattern
- [ ] Extracted functions rather than inline comments

## Examples to Follow

✅ **Good code clarity examples in this codebase:**
- `cmd/cmd_restore.go`: Small functions with clear names
- `internal/restore/restore.go`: Each section in its own function
- `internal/app/doctor/service.go`: Rules pattern with clear comments
- `cmd/section_registry.go`: `ScanGroup` is the section registry and stays self-documenting

---

**Philosophy:** Write code as if the reader won't have documentation. Make it impossible to misunderstand.

