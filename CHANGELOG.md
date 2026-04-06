## [0.4.0] - 2026-04-06

### 🚀 Features

- Refactor code to make it easier to maintain
- Add auto release workflow
- Add release token for auto release
- Implement doctor checks in restore command
- Add docter checks to restore command
- Make sure it supports only flag restore
- Add extra line before verify
- Add random fun subtitles to command headers

### 🐛 Bug Fixes

- Prevent duplicate if a version is change for a language
- Doctor test after rebuilding functions
- Linting issues with doctor cmd
- Move import into the block for the cli
- Change comment to fit function and use RunChecks
- Remove unused metadata.json
- Lint issue with indent
- Wrap all command errors with enhanceError for fuzzy profile matching
- Singular/plural count labels, empty shell section fallback, and error indentation
- MAS app ID parsing, path traversal guards, and EOF handling in ConfirmOverwrite
- Reduce excess whitespace and clean up error handling
- HTTP timeout, raw URL validation, defaults type validation
- Make unnecessarily exported identifiers private
- Redact PII from exported profiles by default, add --no-redact flag
- Brewfile import metadata, runDelete error handling, and add pre-read size check
- Skip no-RestoreKeys groups in hasRestorableData, dynamic --only help, scoped tool warning
- Enhance URL validation for gist with case-insensitivity and port checks
- Improve error handling for FetchGist and enhance URL validation
- Correct commit message prefix for documentation and add continuous integration group
- Improve error handling for fetching raw gist content
- Standardize dash usage in docs and CLI output, clarify error and warning messages
- Improve error message formatting and standardize macOS capitalization in UI and output
- Add indentation to hint in enhanced error message and update related test

### 🚜 Refactor

- Rename cmd files for clearer naming conventions
- Replace repeated divider literals with dividerLine constant
- Export restore.MsgAlreadyInstalled and use it across cmd/tui
- Extract summarizeGit and split printVersionDetails
- Tighten Go API surface and centralize date formats
- Remove dead UI/test helpers and update command docs
- Rename cmd test files to match their source file names

### 📚 Documentation

- Remove completed refactor candidates from CODE_CLARITY.md
- Refresh cmd package guide after layout refactor
- Replace website plain text with shields.io badge

### 🧪 Testing

- Cover extracted cmd helpers and restore selection logic
- Add regression tests for list interactive error path and double-enhancement guard
- Add cmd_delete_test.go for delete command error and confirmation paths
- Update cmd_export_test.go

### ⚙️ Miscellaneous Tasks

- Run smoke after darwin arm64/amd64 build; docs: add website link
