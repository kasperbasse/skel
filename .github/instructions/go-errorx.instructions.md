---
applyTo: "internal/app/errorx/**"
---

# internal/app/errorx - Review Instructions

## Purpose

Adds actionable hints to CLI errors. Rules match error message substrings and
append context-specific guidance. Keeps error enhancement logic separate from
Cobra command code so rules are independently testable.

## Rule ordering matters

`EnhanceMessage` returns the first matching rule's result. More specific rules
must appear before broader ones. For example, the `"profile not found"` rule
must precede the generic `"not found"` tool-hint rule.

## Rules must be additive

An `Enhance` function appends to the existing message (`errMsg + "\n\n..."`)
rather than replacing it. Never discard the original error text - it may contain
file paths, IDs, or other diagnostic context.

## EnhanceOptions - dependency injection

`ToolHint`, `SuggestProfile`, and `ValidSections` are injected by `cmd/error_enhancement.go`.
This keeps the `errorx` package free of `cmd/` imports.
Tests use stub implementations of these fields directly.

## Adding a new rule

1. Add to `buildDefaultRules()` - ordered correctly relative to existing rules.
2. Write a test in `enhancer_test.go` that covers both the match and the enhanced message.
3. If the rule needs runtime context, add a field to `EnhanceOptions` and wire it in `cmd/error_enhancement.go`.