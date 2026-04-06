## [0.4.1] - 2026-04-06

### 🚀 Features

- Add git-cliff changelog generation and integrate with release workflow

### 🚜 Refactor

- Extract drift and diff comparison helpers, unify output, and add tests for diff detection
- Unify drift and diff output by extracting printChangedItemsWithHeader helper
- Make helper functions unexported and update tests for full diff/drift coverage
## [0.4.0] - 2026-04-06

### 🚀 Features

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

## [0.3.0] - 2026-04-05

### 🚀 Features

- Default to 'default' profile for restore
- Default to 'default' profile for delete
- Default to 'default' profile for export
- Default to 'default' profile for publish
- Default to 'default' profile for brewfile export
- Add title to mention the profile which is being published
- Add profile name to title for exporting
- Add profile name to title for brewfile export
- Add line after title and align spacing
- Add needed permissions
- Add confirmation
- Quit by ctrl+c
- Align spacing/header, next steps, boxes for important messages
- Implement enhance error detection visual
- Refactor code to make it easier to maintain

## [0.2.0] - 2026-04-03

### 🚀 Features

- Adding ui improvements
- Add autocompletion feature
- Add progress to scan
- Add timeago feature
- Add doctor command
- Enable doctor and status command
- Add onboarding prompt
- Add status command
- Add option to show all for show command instead of being truncated
- Security - reset git and hostname/os specific from publish
- Ui improvements for spaceing under title for commands
- Add commands to autocompletions
- Sort profiles after CreatedAt
- Add version diffs + remove the "no changes" line
- Add languages to diff checker

### 🐛 Bug Fixes

- Remove dead code and fix lint issues

### 🚜 Refactor

- Remove dead code

### 📚 Documentation

- Update badges
- Update documentation
- Update readme

### 🧪 Testing

- Add tests to completions and sections
- Add function used to test with for override dir

### ⚙️ Miscellaneous Tasks

- Remove new line after no drift detected

## [0.1.0] - 2026-04-03

### 🚀 Features

- Initial release of skel
- Activate homebrew tap
