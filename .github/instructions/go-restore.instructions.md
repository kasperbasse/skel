---
applyTo: "internal/restore/**"
---

# internal/restore - Review Instructions

## Path traversal guard - safeSubPath

`safeSubPath(base, target)` must be called for EVERY file write that uses a path
derived from profile data (`ConfigFiles` keys, JetBrains `Configs` keys).
It uses `filepath.Rel` rather than `strings.HasPrefix` to avoid the
`/foo/bar` vs `/foo/bar2` false-positive. Do not replace it with a prefix check.

## File/directory permissions

- All files written with `0600` (`writeFile`)
- All directories created with `dirPermissions` = `0700`
  Never raise these permissions.

## Adding a new restore section

Two places must stay in sync:

1. `Run()` - add `restoreSection("name", func() { restoreXxx(...) })`
2. `countSteps()` - add the count for the new section

If `countSteps` is not updated, the progress indicator will be wrong (e.g., `[5/4]`).

## Section function pattern

Each section is a standalone function: `restoreXxx(profile, homeDir, emit)`.
Tool presence is checked inside the section function (not in `Run`), so the section
can emit a descriptive failure message rather than silently skipping.

## detectInstalled - best-effort

`detectInstalled()` calls brew/code/cursor/mas. If a tool is missing, the command
returns empty string silently. This is intentional: absence of a tool is handled
per-section, not globally.

## restoreGems - system Ruby warning

If `/usr/bin/ruby` is detected, gems are NOT installed and a warning is emitted.
This prevents native extension failures. Do not remove this check.

## emitResult

Standardized success/failure emit. Use it for all external command invocations.
Never call `emit(name, false, ...)` inline when `emitResult` would do.