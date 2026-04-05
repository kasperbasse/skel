# cmd package guide

Use this package as the CLI orchestration layer.

## File prefixes

- `cmd_*.go`: user-facing command entry files
- `cli_*.go`: command helpers (args, completions, help templates, shared profile-loading)
- `section_*.go`: section registry + section behavior used across commands
- `ui_*.go`: output/rendering helpers
- `error_*.go`: error enhancement helpers
- `test_support.go`: shared test builders/helpers
- `*_test.go`: command-specific and domain-level tests
- `tui/`: interactive Bubble Tea models only; keep CLI orchestration in `cmd`

## Command structure

Each command file should follow this shape:

```go
var xxxCmd = &cobra.Command{
    Use:  "xxx",
    RunE: runXxx,
}

func runXxx(_ *cobra.Command, args []string) error {
    // parse args
    // load profile if needed
    // execute
    // print results
    return nil
}

func init() {
    // command-specific flags only
}
```

Notes:

- Register top-level commands in `cmd/cmd_root.go`
- Put shared completions in `cmd/cli_completions.go`
- Keep `runXxx` small and extract named helpers when branching grows
- Use `RunE` unless the command truly cannot fail

## Important entry points

- Root command and `Execute()`: `cmd/cmd_root.go`
- Shared profile helpers: `cmd/cli_profiles.go`
- Shared completions: `cmd/cli_completions.go`
- Shared output helpers: `cmd/ui_output.go`
- Section registry (`scanGroups`): `cmd/section_registry.go`
- Interactive models: `cmd/tui/`

## Add a new command

1. Create `cmd/cmd_<name>.go`
2. Define `var <name>Cmd = &cobra.Command{ RunE: run<Name> }`
3. Implement `run<Name>` and keep helpers in the same file
4. Add command-specific flags in the file's `init()` if needed
5. Register it in `cmd/cmd_root.go`
6. Hook up tab completion in `cmd/cli_completions.go` if it takes profile args

## Keep it simple

- Prefer one command per file
- Keep handlers small; extract helpers early
- Put display logic in `ui_*.go`
- Put section-related logic in `section_*.go`
- Put interactive state machines in `cmd/tui/`, not in command files
- Add focused tests for extracted helpers in `*_test.go`
