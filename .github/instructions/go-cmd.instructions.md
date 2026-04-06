---
applyTo: "cmd/**"
---

# cmd/ - Review Instructions

## Layer contract

`cmd/` is the orchestration layer only: parse flags/args → call internal packages →
render output. Business logic does NOT belong here. If a PR adds decision-making
logic to `cmd/`, it should move to `internal/app/`.

## File naming conventions

- `cmd_*` - one Cobra command per file, with `RunE: runXxx` and extracted handler
- `cli_*` - shared CLI helpers (arg selection, completions, help templates)
- `section_*` - section registry and per-section behavior
- `ui_*` - terminal output helpers (icons, tables, summaries)
- `error_*` - error enhancement wiring

## Section registries - the core extensibility point

There are two registries in `section_registry.go`:

**`scanGroups`** - high-level sections driving: restore TUI checklist, dry-run
preview, show detail, import warnings, and `--only` flag values.
Each entry has `RestoreKeys []string` which maps to the keys in `restore.Options`.

**`profileSections`** - flat list of comparable fields driving: diff, drift, list,
manage item counts.

Adding a new profile section means entries in BOTH registries. Forgetting one will
cause the section to silently disappear from one of the views.

## `--only` flag validation

`validSections` is derived from `allRestoreKeys()` at startup. The valid values
always reflect the current registry. Never hardcode the list elsewhere.

## Interactive vs non-interactive restore

`IsInteractive()` gates TUI usage. Non-interactive path (`runNonInteractiveRestore`)
is used in CI and pipe contexts. Both paths must reach the same restore logic
(`restore.Run`); they differ only in output rendering.

## TUI components (`cmd/tui/`)

Not covered by unit tests. Test manually. Do not add business logic inside TUI models.