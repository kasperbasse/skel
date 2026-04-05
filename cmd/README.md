# cmd package guide

Use this package as the CLI orchestration layer.

## File prefixes

- `cmd_*.go`: user-facing command entry files
- `cli_*.go`: command helpers (args, completions, help templates, shared profile-loading)
- `section_*.go`: section registry + section behavior used across commands
- `ui_*.go`: output/rendering helpers
- `error_*.go`: error enhancement helpers
- `test_*` + `*_test.go`: test support and tests

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
```

## Current command files

- `cmd_scan.go`
- `cmd_restore.go`
- `cmd_list.go`
- `cmd_show.go`
- `cmd_update.go`
- `cmd_delete.go`
- `cmd_status.go`
- `cmd_drift.go`
- `cmd_diff.go`
- `cmd_doctor.go`
- `cmd_export.go`
- `cmd_import.go`
- `cmd_clone.go`
- `cmd_publish.go`
- `cmd_brewfile.go` (contains `brewfile export` + `brewfile import` subcommands)

## Important entry points

- Root registration: `cmd/cmd_root.go`
- Shared profile helpers: `cmd/cli_profiles.go`
- Shared output helpers: `cmd/ui_output.go`
- Section registry: `cmd/section_registry.go`

## Add a new command

1. Create `cmd/cmd_<name>.go`
2. Define `var <name>Cmd = &cobra.Command{ RunE: run<Name> }`
3. Implement `run<Name>` and keep helpers in the same file
4. Register it in `cmd/cmd_root.go`

## Keep it simple

- Prefer one command per file
- Keep handlers small; extract helpers early
- Put display logic in `ui_*.go`
- Put section-related logic in `section_*.go`
