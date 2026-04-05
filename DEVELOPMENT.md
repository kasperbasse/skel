# Development Guide

This document is for **developers and contributors** who want to understand or modify the Skel codebase.

> **If you just want to use Skel,** see [README.md](./README.md) instead.

## 📚 Documentation Roadmap

### **Required Reading** (30 minutes total)
Read these in order to understand the codebase:

1. **[docs/GLOSSARY.md](./docs/GLOSSARY.md)** ← **Start here** (~5 min)
   - Domain terminology: "profile", "section", "ScanGroup", etc.

2. **[docs/CODE_CLARITY.md](./docs/CODE_CLARITY.md)** (~10 min)
   - How this codebase stays understandable through clear structure
   - Self-documenting functions, single responsibility, strategic comments

3. **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** (~15 min)
   - High-level system design, package organization
   - Step-by-step: adding a new section

### **Reference Docs**
- **Go patterns & conventions:** [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md)
- **Full docs index:** [docs/README.md](./docs/README.md)
- **Contributor guidelines:** [CONTRIBUTING.md](./CONTRIBUTING.md)

## 🛠 Local Development

### Prerequisites
```bash
Go 1.25+        # Install from https://golang.org/dl/
Make            # Usually pre-installed on Mac/Linux
golangci-lint   # For linting: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

### Setup & Build
```bash
git clone https://github.com/kasperbasse/skel.git
cd skel
make build      # Compile ./skel binary
```

### Development Workflow
```bash
# Make a change
vim cmd/cmd_restore.go

# Test it
make test           # Run tests with race detector
make test-v         # Run tests verbosely

# Full quality checks (recommended before committing)
make check          # Runs: vet + lint + test

# Try the binary
./skel --help
./skel scan
./skel restore --dry-run
```

## 🏗 Common Tasks

### I want to...

#### **Understand how restore works**
1. Open [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) → "High-Level Flow"
2. Start with `cmd/README.md` to see how the `cmd/` package is organized
3. Read the comments in `cmd/cmd_restore.go` (heavily commented with explanations)
3. Trace the code:
   - `cmd/cmd_restore.go` → `cmd/section_registry.go` / `cmd/section_*.go` → `internal/restore/restore.go`

#### **Add a new restore section**
See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) → "Adding a New Section" (step-by-step guide)

#### **Understand tool requirements**
1. Read `internal/app/doctor/service.go` (has detailed comments)
2. See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) → "Key Design Decisions"
3. Understand the rules-based pattern in `internal/app/doctor/service.go`

#### **Write a test**
1. See [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) → "Testing Go Code"
2. Use test helpers: `cmd/test_support.go` has `NewTestProfile()` and `TestOptions()`
3. Examples: Look at `cmd/*_test.go` files

#### **Fix a bug**
1. Write a test that reproduces the bug
2. Fix the code
3. Verify test passes: `make test`
4. Run full checks: `make check`

#### **Learn Go patterns**
See [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) for:
- Error handling (wrap with %w)
- Multiple return values
- Interfaces (implicit, structural)
- Callbacks & closures
- Nil checks & early returns
- Table-driven tests

## 📋 Before Submitting a PR

### Checklist
- [ ] Read [CONTRIBUTING.md](./CONTRIBUTING.md)
- [ ] Read [docs/GLOSSARY.md](./docs/GLOSSARY.md), [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md), and relevant parts of [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- [ ] Code follows Go style guide (`make fmt` applied)
- [ ] Tests added for new functionality
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Full checks pass (`make check`)
- [ ] Comments added for exported functions
- [ ] Code comments explain the "why", not just the "what"
- [ ] No hardcoded paths or credentials
- [ ] Errors wrapped with `%w`
- [ ] Type assertions checked (not just `result.(Type)`)

### Make Targets
```bash
make help                   # List all available targets
make build                  # Compile ./skel
make install                # Install to $GOPATH/bin
make test                   # Run tests with race detector
make test-v                 # Run tests verbosely
make lint                   # Run golangci-lint
make vet                    # Run go vet
make fmt                    # Format all .go files
make check                  # vet + lint + test (local CI)
make tidy-check             # Ensure go.mod/go.sum are tidy
make vulncheck              # Check for known CVEs
make build-darwin           # Build arm64 and amd64 binaries
make ci-local               # Full CI checks (like GitHub Actions)
```

## 🗂 Project Structure

```
skel/
├─ README.md                 # Product overview (for users)
├─ CONTRIBUTING.md           # Contributor guide
├─ DEVELOPMENT.md            # ← This file (for developers)
├─ Makefile                  # Build targets
├─ go.mod / go.sum           # Go dependencies
│
├─ docs/                     # Developer documentation
│  ├─ README.md              # Navigation hub
│  ├─ GLOSSARY.md            # Domain terms
│  ├─ CONTRIBUTING_GO.md     # Go patterns
│  └─ ARCHITECTURE.md        # System design
│
├─ main.go                   # Entry point
├─ cmd/                      # CLI package (flat, grouped by prefix)
│  ├─ README.md              # ← Start here for cmd/ navigation
│  ├─ cmd_*.go               # User-facing commands
│  ├─ cli_*.go               # Command helpers (args, completions, profile loading)
│  ├─ section_*.go           # Section registry and section behavior
│  ├─ ui_*.go                # Output and display helpers
│  ├─ error_*.go / test_*.go # Focused support code
│  └─ tui/                   # Terminal UI (Bubble Tea)
│
├─ internal/                 # Internal packages (not exported)
│  ├─ app/
│  │  ├─ doctor/             # Tool requirement validation
│  │  │  ├─ service.go       # ← Rules-based design (read comments)
│  │  │  ├─ checks.go        # Validation rows
│  │  │  └─ tools.go         # Tool metadata
│  │  └─ errorx/             # Error handling utilities
│  ├─ profile/               # Profile data model & persistence
│  ├─ restore/               # Restore execution
│  ├─ scanner/               # Profile capture/scanning
│  ├─ ui/                    # UI helpers (colors, icons, etc.)
│  └─ version/               # Version information
│
└─ tests/                    # Integration tests (if any)
```

## 🎓 Learning Path

### For New Contributors (30 minutes)

**Goal:** Understand enough to make your first contribution

1. **Read README.md** (3 min) - What Skel does
2. **Read docs/GLOSSARY.md** (5 min) - Domain terms
3. **Read docs/CODE_CLARITY.md** (10 min) - How code is organized
4. **Trace one command** (12 min)
   - Pick: `cmd/cmd_restore.go` or `cmd/cmd_scan.go`
   - The code reads like prose—follow the function names
   - Each function is small and does one thing

### For Deeper Understanding (1 hour additional)

1. **Read docs/ARCHITECTURE.md** (15 min)
   - Package structure, how sections work
   
2. **Study a pattern closely** (15 min)
   - Pick: rules pattern in `internal/app/doctor/service.go`
   - Or: section organization in `cmd/section_registry.go`
   
3. **Read test examples** (15 min)
   - `cmd/restore_test.go`, `cmd/scan_test.go`
   - `cmd/test_support.go` — reusable test builders
   
4. **Go patterns reference** (15 min)
   - [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) — callbacks, interfaces, error handling


## 🤔 FAQ

**Q: Why is code organized this way?**
A: See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) → "Package Organization"

**Q: Why use callbacks?**
A: See [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) → "Callbacks (Functional Programming)"

**Q: Why wrap errors with %w?**
A: See [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) → "Error Handling"

**Q: How do I add a new section?**
A: See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) → "Adding a New Section"

**Q: Why is this function called a "builder"?**
A: See code comments or [docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md) → "Common Patterns"

## 🔗 Useful Resources

- **Effective Go:** https://golang.org/doc/effective_go
- **Go Code Review Comments:** https://github.com/golang/go/wiki/CodeReviewComments
- **Bubble Tea (TUI framework):** https://github.com/charmbracelet/bubbletea
- **Go Testing:** https://golang.org/pkg/testing/

## 💬 Questions?

1. **Check the docs first:** See the "I want to..." section above
2. **Search code comments:** Start with `cmd/README.md`, then open focused files like `cmd/cmd_restore.go`
3. **Open an issue:** Describe what you're trying to understand
4. **Ask on discussions:** Use GitHub Discussions for design questions

---

**Happy coding! 🚀**

