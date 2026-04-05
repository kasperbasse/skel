# Contributing to skel

Thanks for your interest in contributing! Here's how to get started.

## Prerequisites

| Tool          | Version                         | Install                                                                  |
|---------------|---------------------------------|--------------------------------------------------------------------------|
| Go            | 1.25+                           | [go.dev/dl](https://go.dev/dl/)                                          |
| golangci-lint | v2.11.4 (matches CI)            | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4` |
| GoReleaser    | latest (optional, for releases) | `brew install goreleaser`                                                |

## Getting Started

```bash
git clone https://github.com/kasperbasse/skel
cd skel
make build   # compile ./skel
make test    # run all tests with race detector
```

Run `make help` to list all available targets.

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Run `make check` (vet + lint + test — fast local gate)
5. Open a pull request

## Code Style

- Keep it simple. No premature abstractions.
- Use `fmt.Errorf("context: %w", err)` for error wrapping.
- New fields on profile structs should use `omitempty` for backward compatibility.
- Config files are written with `0600` permissions, directories with `0700`.
- User-facing strings should use regular dashes (`-`), not em dashes.
- Run `golangci-lint run ./...` before submitting. The project uses a `.golangci.yml` config that enforces formatting (`gofmt`, `goimports`), static analysis, and security checks.
- Import groups: stdlib, then external packages, then local (`github.com/kasperbasse/skel/...`), separated by blank lines.

## Architecture Guidelines

- Keep `cmd/*` thin: parse args/flags, call app logic, render output.
- Put business decisions in `internal/app/*` packages.
- Put terminal styling/printing helpers in `internal/ui`.
- Prefer registry-driven additions (sections, tools, error rules) over repeated `if/switch` blocks.
- Add tests for each new pure helper and each new registry rule.

## Testing

Every new pure function should have a test. Tests live next to the code they test (`*_test.go`).

```bash
make test    # run all tests with race detector (shortcut)
make test-v  # verbose output

# or directly:
go test ./...
go test -v ./...
go test -race ./...
```

Scanner/restore functions that call external tools (`brew`, `code`, etc.) are tested manually. Pure functions (parsers, validators, helpers) should have unit tests.

### Memory Profiling

Use the bundled target to spot allocation hotspots and potential retention issues while tests run:

```bash
make memcheck
make memcheck MEM_PKG=./internal/profile
make memcheck-loop MEM_PKG=./internal/scanner MEM_RUNS=50
make memcheck-report MEM_RUNS=25
make memcheck-baseline MEM_RUNS=25
make memcheck-delta
```

This writes `mem.out` in repo root and prints a quick `pprof` top table.
`MEM_PKG` defaults to `./cmd` because `-memprofile` can only be used with one package per run.
`memcheck-loop` repeats tests (`MEM_RUNS`, default `25`) to make slow retention issues easier to spot.
`memcheck-report` profiles multiple packages (`MEM_PKGS`) and writes one `.out` + one `.txt` report per package into `memreports/`.
`memcheck-baseline` snapshots `memreports/` into `memreports-baseline/`, and `memcheck-delta` prints per-package in-use/alloc changes.

Useful follow-ups:

```bash
go tool pprof -top mem.out
go tool pprof -sample_index=alloc_space -top mem.out
go tool pprof -list Run mem.out
```

## CI Required Checks

For branch protection on `main`, require these GitHub Actions job names from `.github/workflows/ci.yml`:

- `Quality (vet + lint + tidy)`
- `Test (race)`
- `Security (govulncheck)`
- `Build (darwin binaries)`

This keeps protection rules stable even if workflow internals evolve.

## CI Troubleshooting

When a CI job fails, run `make ci-local` first — it reproduces the full CI gate in one command:

```bash
make ci-local
```

For faster iteration, run `make check` (core vet/lint/test only):

```bash
make check
```

Or run individual checks to match each CI job exactly:

```bash
go mod verify
go mod tidy && git diff --exit-code -- go.mod go.sum
go vet ./...
golangci-lint run ./...
go test -v -race ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./...
GOOS=darwin GOARCH=arm64 go build -o skel-arm64 .
GOOS=darwin GOARCH=amd64 go build -o skel-amd64 .
```

Tips:
- If `tidy` fails in CI, commit updated `go.mod`/`go.sum`.
- If lint differs locally vs CI, ensure you are on the pinned linter version (`v2.11.4`) used in the workflow.
- If `govulncheck` finds an issue, update the affected dependency to a non-vulnerable version and re-run tests.

## Maintainer Checklist

Before merge:
- Ensure CI is green for `Quality (vet + lint + tidy)`, `Test (race)`, `Security (govulncheck)`, and `Build (darwin binaries)`.
- Confirm user-facing output changes are intentional and consistent with existing command style.
- Check that architecture boundaries still hold (`cmd` thin, `internal/app` logic, `internal/ui` rendering).
- Verify new registries/helpers include tests and docs comments where exported.

Before release:
- Re-run full checks locally (`go test ./...`, `go vet ./...`, `golangci-lint run ./...`).
- Confirm `go mod tidy` produces no changes.
- Verify darwin binaries build for both `arm64` and `amd64`.
- Review release notes/changelog entries for user-visible changes.
- Ensure CI is green on `main`, then run the **Auto Release** workflow manually with the version you want to cut.

## Security

If you find a security issue, please email kasperbasse@gmail.com instead of opening a public issue.

Key things to watch for:
- Path traversal in `ConfigFiles` map keys
- File permissions on restored configs (must be `0600`)
- Validate imported profiles before saving
- Never store SSH private keys, `.env` files, or tokens

## Releases

Releases are created manually from GitHub Actions on `main`. This keeps the flow simple and trusted while still avoiding manual local tag pushes.

When you're ready to release:

1. Make sure CI is green on `main`
2. Go to GitHub → Actions → **Auto Release**
3. Click **Run workflow**
4. Enter the exact version to cut (for example `v0.2.0`)
5. Leave the target branch as `main`

When the workflow runs it:

1. Validates the requested version (`vMAJOR.MINOR.PATCH`)
2. Verifies the tag does not already exist
3. Creates and pushes a real Git tag from `main`
4. Creates a GitHub release with generated release notes
5. Marks `0.x` releases as **pre-release** automatically

The separate **Release** workflow then:
- Builds macOS binaries (arm64 + amd64) using GoReleaser
- Updates the Homebrew tap
- Publishes artifacts to the GitHub release

### Repository Rules / Tokens

If your repository rules restrict tag creation, configure a `RELEASE_TOKEN` secret (PAT or GitHub App token) with permission to create tags and, if needed, bypass the tag/ref ruleset. For protected repositories, the release workflow expects this token explicitly instead of falling back to `GITHUB_TOKEN`.

Recommended setup:
- create a fine-grained PAT or GitHub App token with repository `Contents: Read and write`
- save it as a repository secret named `RELEASE_TOKEN`
- allow that token/app to create `v*` tags in your repository ruleset (or grant bypass where required)

Also make sure branch protection / rulesets require the current CI job names from `.github/workflows/ci.yml`, especially `Test (race)` rather than an old `test` check name.

**Note:** All 0.x releases are marked as pre-releases on GitHub. Once v1.0.0 is released, releases will be marked as stable.

The GitHub release uses generated release notes based on merged work since the previous release.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
