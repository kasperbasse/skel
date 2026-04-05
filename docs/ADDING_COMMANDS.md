# Adding New Commands to Skel

This guide explains how to add new commands using the established patterns in the Skel codebase. The goal is to make adding new functionality as simple and predictable as possible.

## Command Structure Overview

Every command follows these patterns:

1. **Command Definition** - Define the cobra.Command with Use, Short, Args, RunE
2. **Extract Run Function** - Pull business logic into a `runXxx` function
3. **UI Calls** - Use `PrintCommandHeader`, `PrintWarnings`, etc. from `cmd/ui_output.go`
4. **Error Handling** - Use helper functions for loading profiles and handling errors
5. **Profile Loading** - Use `SelectProfileName()` and `LoadAnyProfile()` helpers

## Template: Simple Profile Command

For commands that just load a profile and display information:

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var myCmd = &cobra.Command{
	Use:   "mycommand [profile-name]",
	Short: "Short description",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runMyCommand,
}

// runMyCommand implements the command logic.
func runMyCommand(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	PrintCommandHeader("mycommand", "Processing " + bold(name))
	
	// Do your work here
	doSomethingWith(p)

	fmt.Printf("  %s Command completed\n\n", iconCheck())
	return nil
}

func init() {
	// Add command-specific flags here if needed.
}
```

Then register the command in `cmd/cmd_root.go` and wire completions in `cmd/cli_completions.go`.

## Template: Command with Confirmation

For commands that need user confirmation before proceeding:

```go
var updateCmd = &cobra.Command{
	Use:   "update [profile-name]",
	Short: "Update an existing profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

func runUpdate(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	// Get confirmation
	ok, err := ConfirmOverwrite(name)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("  %s Canceled.\n\n", iconDash())
		return nil
	}

	// Proceed with update
	return performUpdate(name)
}
```

## Template: Command that Compares Profiles

For commands that compare two profiles:

```go
var compareCmd = &cobra.Command{
	Use:   "compare [profile-a] [profile-b]",
	Short: "Compare two profiles",
	Args:  requireExactArgs(2, "compare <profile-a> <profile-b>"),
	RunE:  runCompare,
}

func runCompare(_ *cobra.Command, args []string) error {
	profileA, err := loadProfileOrFail(args[0])
	if err != nil {
		return err
	}

	profileB, err := loadProfileOrFail(args[1])
	if err != nil {
		return err
	}

	PrintCommandHeader("compare", fmt.Sprintf("Comparing %s → %s", 
		bold(args[0]), bold(args[1])))

	// Do comparison...
	displayComparison(profileA, profileB)

	return nil
}
```

## Key Helper Functions

### Profile Loading

```go
// Get profile name from args, default to "default"
name := SelectProfileName(args)

// Load profile with first-run fallback
p, err := LoadAnyProfile(name)

// Load profile and return error if fails
p, err := loadProfileOrFail(name)
```

### UI Output

```go
// Print command header
PrintCommandHeader("cmdname", "Your message")

// Print warnings
PrintWarnings(warnings)

// Print a formatted error
PrintError(err)

// For command-specific success/info output, print directly using the shared
// styling helpers (`iconCheck`, `iconDot`, `bold`, `dim`, etc.).
```

### Confirmation Dialogs

```go
// Confirm overwriting a profile
ok, err := ConfirmOverwrite(name)
if !ok {
    fmt.Printf("  %s Canceled.\n\n", iconDash())
    return nil
}
```

## Adding a Section-Based Command

For commands that work with multiple sections (like scan, restore, drift):

1. The sections are defined in `section_registry.go` in the `scanGroups` slice
2. Each section has: Icon, Label, RestoreKeys, ScanSummary, ShowDetail, DryRun

Example - iterating sections:

```go
for _, g := range scanGroups {
	if g.ScanSummary == nil {
		continue
	}
	if summary := g.ScanSummary(p); summary != "" {
		// This section has data
		fmt.Printf("  %s %s: %s\n", g.Icon, g.Label, summary)
	}
}
```

## Error Handling Pattern

Always wrap errors with context:

```go
// Instead of:
return err

// Use:
return fmt.Errorf("operation name: %w", err)
```

Always check errors immediately after operations:

```go
p, err := profile.Load(name)
if err != nil {
	return fmt.Errorf("loading profile: %w", err)
}
```

## Command Registration

All commands must be registered in `cmd_root.go` during init:

```go
func init() {
	rootCmd.AddCommand(myCmd)
	myCmd.ValidArgsFunction = singleProfileCompletion
}
```

## Testing

Test your command with:

```bash
# Build and test
go build -o skel .
./skel mycommand --help
./skel mycommand

# Run tests
go test ./cmd -v
```

## Checklist for New Commands

- [ ] Command uses `SelectProfileName()` for profile name handling
- [ ] Command uses `LoadAnyProfile()` or `loadProfileOrFail()` for loading
- [ ] Command uses `PrintCommandHeader()` for output
- [ ] Error messages include context (using `%w`)
- [ ] Errors are checked immediately after operations
- [ ] Command function name follows `runXxx` pattern
- [ ] Command is extracted from inline closure
- [ ] Command added to `cmd_root.go` init()
- [ ] Completion set up (if applicable)
- [ ] Tests updated/added
- [ ] `make check` passes (vet + lint + test)

## Common Patterns

### Loading Profiles with Fallback
```go
p, err := LoadAnyProfile(name)
if err != nil {
	return err  // Already enhanced with context
}
```

### Displaying Sections
```go
for _, g := range scanGroups {
	if summary := g.ScanSummary(p); summary != "" {
		printRow(g.Label, summary)
	}
}
```

### Handling Interactive vs Non-Interactive
```go
if IsInteractive() {
	return runInteractiveMode(p)
}
return runNonInteractiveMode(p)
```

### Reporting Results
```go
printNextSteps(
	nextStep("skel show " + name, "to see details"),
	nextStep("skel restore " + name, "to apply"),
)
```

## Design Principles

1. **Consistency** - All commands follow the same pattern
2. **Clarity** - Function names describe what they do
3. **Reusability** - Common patterns extracted to helpers
4. **Testability** - Logic separated from UI rendering
5. **Flexibility** - Easy to add new sections or options

## Examples in Codebase

- **Simple command:** `cmd/cmd_show.go` - Just load and display
- **With confirmation:** `cmd/cmd_scan.go` - Confirm before overwriting
- **Comparison:** `cmd/cmd_diff.go` - Compare two profiles
- **Complex:** `cmd/cmd_restore.go` - Multiple phases and user interaction
- **Analysis:** `cmd/cmd_drift.go` - Scan and compare

Study these examples to understand each pattern better.

