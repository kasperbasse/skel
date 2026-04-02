# Contributing to skel

Thanks for your interest in contributing! Here's how to get started.

## Prerequisites

| Tool          | Version                         | Install                                                                 |
|---------------|---------------------------------|-------------------------------------------------------------------------|
| Go            | 1.25+                           | [go.dev/dl](https://go.dev/dl/)                                         |
| golangci-lint | latest                          | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| GoReleaser    | latest (optional, for releases) | `brew install goreleaser`                                               |

## Getting Started

```bash
git clone https://github.com/kasperbasse/skel
cd skel
go build -o skel .
go test ./...
```

## Making Changes

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Add tests for new functionality
4. Run `go test ./...`, `go vet ./...`, `go mod tidy`, and `go build ./...`
5. Open a pull request

## Code Style

- Keep it simple. No premature abstractions.
- Use `fmt.Errorf("context: %w", err)` for error wrapping.
- New fields on profile structs should use `omitempty` for backward compatibility.
- Config files are written with `0600` permissions, directories with `0700`.
- User-facing strings should use regular dashes (`-`), not em dashes.
- Run `golangci-lint run ./...` before submitting. The project uses a `.golangci.yml` config that enforces formatting (`gofmt`, `goimports`), static analysis, and security checks.
- Import groups: stdlib, then external packages, then local (`github.com/kasperbasse/skel/...`), separated by blank lines.

## Testing

Every new pure function should have a test. Tests live next to the code they test (`*_test.go`).

```bash
go test ./...         # run all tests
go test -v ./...      # verbose
go test -race ./...   # race detector
```

Scanner/restore functions that call external tools (`brew`, `code`, etc.) are tested manually. Pure functions (parsers, validators, helpers) should have unit tests.

## Security

If you find a security issue, please email kasperbasse@gmail.com instead of opening a public issue.

Key things to watch for:
- Path traversal in `ConfigFiles` map keys
- File permissions on restored configs (must be `0600`)
- Validate imported profiles before saving
- Never store SSH private keys, `.env` files, or tokens

## Releases

Releases are automated via GitHub Actions and GoReleaser. Tag a version to trigger:

```bash
git tag v0.1.0
git push --tags
```

This builds macOS binaries (arm64 + amd64) and updates the Homebrew tap.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
