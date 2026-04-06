---
applyTo: "internal/scanner/**"
---

# internal/scanner - Review Instructions

## Purpose

Captures the current machine state into a `profile.Profile`. Read-only: never
modifies system state. Non-fatal errors become warnings, not hard failures.

## Warning pattern

`warn(msg)` appends to the warnings slice; it never stops the scan.
Missing tools or unreadable files should warn, not error.
Only truly unrecoverable errors (e.g. cannot determine home dir) return an error.

## Progress callback

`RunWithProgress` accepts an `onProgress(label)` callback for live TUI feedback.
`Run` is just a convenience wrapper with `nil` progress. Both must stay in sync.

## New section checklist

1. Add a `scanXxx(...)` function (read-only, warn on missing tools)
2. Call `progress("Section Label")` before it in `RunWithProgress`
3. Assign result to `p.XxxField`
4. Add tests for the parser/extractor helpers (not the exec calls)
5. Add a `scanGroup` entry in `cmd/section_registry.go`
6. Add a `profileSection` entry in `cmd/section_registry.go`

## Defaults scanning

`defaults read` output is parsed into `DefaultsSetting` entries. The `Type` field
must be one of the `validDefaultsTypes` allowlist in `profile.go` - this is
validated on `Validate()` before any write. Scanner sets the type based on the
detected value format.