---
applyTo: "internal/profile/**"
---

# internal/profile - Review Instructions

## Security: never store private keys

`SSHProfile` stores key metadata only - filename, type, fingerprint of the PUBLIC
key, and comment. Private key content must NEVER appear in any field.
`Validate()` checks for PEM markers (`PRIVATE KEY`, `BEGIN RSA`, etc.) as defense
in depth. Do not weaken or remove this check.

## Security: path traversal in ConfigFiles

`ConfigFiles` map keys are relative paths written under the user's home directory.
`Validate()` rejects any key that is absolute or escapes via `..`.
Any new code writing `ConfigFiles` to disk MUST call `Validate()` first.

## PII: Redact() before publishing

`Redact()` clears: git name/email, raw gitconfig content, hostname, machine name,
SSH key comments. It does NOT clear shell config contents or `.gitignore_global`
because those are the value of a shared profile.
Call `Redact()` before any publish/export operation that sends data externally.

## New profile fields

- Always add `omitempty` to new JSON tags for backward compatibility.
  Old profiles that don't have the field must still parse cleanly.
- If the field contains user-controlled strings that will be used as paths,
  add a validation rule in `Validate()`.

## File permissions

`Save()` writes with `0600`. Profile directory is created with `0700`.
Never change these. Config files restored to disk also use `0600`.

## Test isolation

`profileDirOverride` redirects disk operations to a temp dir.
Use `SetProfileDirOverride(t.TempDir())` in tests; reset with `defer SetProfileDirOverride("")`.
Do not access `profileDirOverride` directly from test code outside this package.