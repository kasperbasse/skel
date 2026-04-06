# Contributing to Skel: Go Guide

This guide explains Go-specific patterns and idioms used in this codebase. If you know other languages, this bridges the gap.

## Error Handling

Go uses explicit error returns instead of exceptions:

```go
// Go pattern: always check errors
result, err := someFunction()
if err != nil {
    return fmt.Errorf("context: %w", err)  // %w wraps the error chain
}
// Use result...
```

**Key points:**
- `err != nil` checks are common (not verbose, it's the idiomatic way)
- Always wrap errors with `%w` for error chain preservation
- Don't ignore errors even if you think they won't happen
- Early returns keep code readable (exit on error, continue on success)

## Multiple Return Values

Go functions can return multiple values without tuples or special syntax:

```go
// Go: multiple returns are normal and idiomatic
func LoadProfile(name string) (*Profile, error) {
    // ...
    return &profile, nil  // success case
    // or
    return nil, fmt.Errorf("not found: %s", name)  // error case
}

// Caller pattern
p, err := LoadProfile("default")
if err != nil {
    return err  // handle error
}
// p is now guaranteed non-nil
```

**Why this matters:**
- Forces caller to handle errors explicitly
- No hidden state or exceptions
- Very clear what function can return

## Interfaces (Implicit, Structural)

Go interfaces are **implicit** - you don't declare "I implement this":

```go
// Define interface (what the function needs)
type ToolResolver func(command string) (label, validatorCmd, fix string, ok bool)
type ToolExists func(command string) bool

// In buildChecksWith, we accept ANY function matching these signatures:
func buildChecksWith(resolve ToolResolver, exists ToolExists) []Check {
    // We can call resolve() and exists() without caring who implements them
    label, _, _, ok := resolve(cmd)
    if exists(validatorCmd) { ... }
}

// Caller provides concrete implementations:
checksWith(ToolDoctorInfo, CommandExists)  // both functions match the interface
```

**Why this is powerful:**
- No inheritance hierarchy needed
- Functions can be passed like strategy pattern
- Easy to test (provide mock functions)
- Decoupled from implementations

## Callbacks (Functional Programming)

Notice this pattern in `cmd/cmd_restore.go`:

```go
// Function that takes a callback
func RequiredToolsForSections(p *Profile, shouldInclude func(string) bool) []string {
    for _, r := range rules {
        if !shouldInclude(r.Section) {  // call the callback
            continue
        }
        // ...
    }
}

// Caller provides the logic inline
requiredTools := RequiredToolsForSections(p, func(section string) bool {
    if len(opts.Sections) > 0 {
        return opts.Sections[section]
    }
    return true
})
```

**This is idiomatic Go:**
- Pass behavior (functions) as parameters
- Closures can capture local variables (`opts`, `section` map)
- Very clean way to customize behavior without subclassing

## Pointer vs Value Receivers

```go
// Method on pointer receiver - can modify the struct
func (m *SelectRestoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.cursor++  // modifies the model
    return m, nil
}

// Read-only? Use value receiver OR pointer if the struct is large
func (c Check) String() string {
    return c.Label
}
```

**When to use each:**
- **Pointer receiver**: when method modifies struct, or struct is large (>128 bytes)
- **Value receiver**: when you don't modify, and struct is small

## Nil Checks and Early Returns

```go
// Go pattern: check preconditions early, return early on error
func hasRestorableData(p *profile.Profile, opts *restore.Options) bool {
    for _, g := range scanGroups {
        if g.ScanSummary == nil {  // nil check
            continue
        }
        // ... rest of logic ...
    }
    return false
}
```

**Philosophy:**
- Check preconditions first
- Fail fast (return early)
- Main logic is at the "happy path" level of indentation
- Easier to read than deeply nested if-else

## Deferred Cleanup

```go
// Go pattern: guarantee cleanup even if error occurs
func processFile(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer f.Close()  // guarantees Close() happens, even if panic
    
    // process file...
    return nil
}
```

**Key:**
- `defer` statements run in reverse order when function exits
- Cleanup guaranteed even if function panics
- Great for file handles, locks, etc.

## Structs with Embedded Fields

```go
// If a struct has only one field, it's often the primary "subject"
type Check struct {
    Label string
    OK    bool
    Fix   string
}

// When building, use struct literals (common in Go)
items = append(items, tui.SelectItem{
    Icon:         g.Icon,
    Label:        g.Label,
    Keys:         g.RestoreKeys,
    Summary:      summary,
    Selected:     !blocked,
    Blocked:      blocked,
    MissingTools: missingTools,
})
```

**Why named fields?**
- Self-documenting (no positional args confusion)
- Adding fields later doesn't break callers
- Easier to read than `SelectItem{"icon", []string{}, ...}`

## Range Loop Patterns

```go
// Discard index if you don't need it
for _, g := range scanGroups {
    // use g, ignore loop index
}

// Keep both if needed
for i, g := range scanGroups {
    if i == selectedIndex {
        // ...
    }
}

// Range over map (order is random, so don't rely on order)
for section, missingTools := range missingToolsBySection {
    // ...
}
```

## Type Aliases for Clarity

```go
// Define semantic types for callbacks
type ToolResolver func(command string) (label, validatorCmd, fix string, ok bool)
type ToolExists func(command string) bool
type SectionFilter func(section string) bool

// Now function signatures are more readable:
func buildChecksWith(resolve ToolResolver, exists ToolExists) []Check { ... }
// vs:
func buildChecksWith(resolve func(...) (...), exists func(...) bool) []Check { ... }
```

## Common Patterns in This Codebase

### 1. "Split Logic into Helper Functions"
```go
// Main function orchestrates high-level flow
RunE: func(...) error {
    if !hasRestorableData(p, opts) { return ... }
    checkToolRequirements(p, opts, dryRun)
    // ...
}

// Helpers do the actual work
func hasRestorableData(...) bool { ... }
func checkToolRequirements(...) error { ... }
```
**Benefit:** Main logic is a readable story; details in helpers.

### 2. "Validate Early, Return Early"
```go
// Check all preconditions upfront
if p == nil { return nil }
if resolve == nil || exists == nil { return nil }

// If we reach here, all invariants are satisfied
checks := make([]Check, 0, len(requiredTools))
```
**Benefit:** Rest of code doesn't need defensive checks.

### 3. "Use Closures for Dependency Injection"
```go
RequiredToolsForSections(p, func(section string) bool {
    // Closure captures 'opts' from outer scope
    return opts.Sections[section]
})
```
**Benefit:** No need to pass opts through function params.

## Testing Go Code

```go
func TestCommandExists(t *testing.T) {
    tests := []struct {
        name     string
        command  string
        expected bool
    }{
        {"nonexistent", "xyz_invalid_cmd", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CommandExists(tt.command)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

**Table-driven tests** are standard in Go:
- Define test cases as slice of structs
- Loop and use `t.Run()` for subtest names
- Easy to add cases without repeating code

## Go Gotchas to Avoid

1. **Forgetting error checks** - linter will catch but don't ignore
2. **Shadowing variables** - `err` in nested scope hides outer `err`
   ```go
   if err != nil { ... }  // outer err
   if x, err := foo(); err != nil { ... }  // SHADOWS outer err!
   ```
3. **Nil pointer dereference** - Go won't catch until runtime
   ```go
   var p *Profile  // nil
   p.Name  // PANIC at runtime
   ```
4. **Modifying map while iterating** - undefined behavior
5. **Goroutines without wait** - program exits before they complete

## Resources

- [Effective Go](https://golang.org/doc/effective_go) - official Go best practices
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - idioms & conventions
- [Errors are values](https://go.dev/blog/errors-are-values) - how to think about error handling

## Next Steps

Start with these patterns in the codebase:
1. Find a function with clear error handling → trace how errors flow
2. Find a callback (`func(...)`) → understand why it's used vs. alternative
3. Write a small test using table-driven style
4. Add a helper function to split some complex logic

Ask questions! Open PRs with comments if patterns confuse you.

