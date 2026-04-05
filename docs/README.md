# Skel Documentation

Welcome to Skel! This folder contains comprehensive documentation to help you understand and contribute to the codebase.

## 🚀 Quick Navigation

### **First Time Here?**
1. Start with **[GLOSSARY.md](./GLOSSARY.md)** (~5 min)
   - Learn what terms like "profile", "section", "restorable data" mean
   - Understand concepts like "ScanGroup", "blocked section", "rule"

2. Then read **[CODE_CLARITY.md](./CODE_CLARITY.md)** (~10 min)
   - Learn how this codebase stays understandable
   - Self-documenting functions, single responsibility, strategic comments
   - Why small focused functions beat heavy documentation

3. Then read **[CONTRIBUTING_GO.md](./CONTRIBUTING_GO.md)** (~10 min)
   - Go patterns used throughout the codebase
   - Error handling, callbacks, interfaces, testing

4. Finally read **[ARCHITECTURE.md](./ARCHITECTURE.md)** (~15 min)
   - See how packages connect
   - Understand data flow and design decisions
   - Follow the step-by-step guide: "Adding a New Section"

**Total time: ~40 minutes. Totally worth it.**

### **I Want To...**

| Goal | Start Here |
|------|-----------|
| Understand what "section" means | GLOSSARY.md |
| Write a test | CONTRIBUTING_GO.md → Testing, then ARCHITECTURE.md → "Adding a New Section" |
| Add a new restore section | ARCHITECTURE.md → "Adding a New Section" |
| Understand error handling | CONTRIBUTING_GO.md → Error Handling |
| Learn Go callbacks/closures | CONTRIBUTING_GO.md → Callbacks |
| See how restore works | ARCHITECTURE.md → High-Level Flow, then trace cmd/cmd_restore.go |
| Understand tool requirements | ARCHITECTURE.md → Key Design Decisions, then internal/app/doctor/service.go |

## 📋 Document Overview

### **[GLOSSARY.md](./GLOSSARY.md)** — Domain Vocabulary
Explains every term used in the codebase. If you see a word you don't understand, start here.

**Contains:**
- Domain terms (Profile, Section, Restorable Sections, Blocked Section)
- Technical terms (ScanGroup, Tool Requirement Rule, Check)
- UI terms (Bubble Tea, Dry-Run, Interactive Mode, --only flag)
- File/directory terms (Profile Directory, Profile Name)
- Error handling terms (Error Wrapping, Error Enhancement)
- Common patterns (Early Return, Callback, Nil Check, Struct Literal)
- Command reference (scan, restore, doctor, etc.)

### **[CODE_CLARITY.md](./CODE_CLARITY.md)** — Self-Documenting Code
Learn how this codebase achieves clarity through structure and naming.

**Contains:**
- Philosophy: good code speaks for itself
- Small focused functions with clear names
- Strategic comments (explain "why", not "what")
- Examples of before/after refactoring
- Principles: single responsibility, descriptive names, consistent patterns
- Benefits: faster onboarding, fewer bugs, easier maintenance
- Checklist for writing clear code

### **[CONTRIBUTING_GO.md](./CONTRIBUTING_GO.md)** — Go Patterns & Idioms
Deep dive into Go conventions and patterns used in this codebase.

**Contains:**
- Error handling (check errors immediately, wrap with %w)
- Multiple return values (why Go uses them, how to interpret them)
- Interfaces (implicit, structural typing, when to use)
- Callbacks (functional programming, closures, dependency injection)
- Nil checks & early returns (idiomatic flow)
- Table-driven tests (Go testing standard)
- Common gotchas (nil slices, uninitialized maps, goroutine timing)
- External resources (Effective Go, Code Review Comments, Bubble Tea)

### **[ARCHITECTURE.md](./ARCHITECTURE.md)** — System Design
High-level overview of how the system works, how to add features.

**Contains:**
- High-level flow diagrams (user runs command → what happens)
- Package organization (cmd/, internal/, responsibilities)
- Dependency flow (which package calls which)
- Key design decisions (why rules are data-driven, why sections can be blocked)
- How sections work (ScanGroup definition, how tools are required)
- Error handling pattern (errors bubble up cleanly)
- Testing strategy (unit, integration, what to test)
- Step-by-step: Adding a new section
- Common workflows (interactive mode, dry-run, --only flag)
- Performance notes
- Future improvements

## 💡 Pro Tips

1. **Understand "rules"** — Most tool requirement logic is data-driven (see service.go). This makes the codebase very flexible.

2. **Test before refactoring** — Run `make test` before AND after changes. Catch regressions immediately.

3. **Use test helpers** — Don't manually build Profiles. Use `NewTestProfile()` from `cmd/test_support.go`.

4. **Follow the flow** — All commands route through `cmd/` orchestrators, which call `internal/` packages.

5. **Errors should wrap** — Always use `%w` in `fmt.Errorf()`. This preserves the error chain for debugging.

6. **Read the code comments** — Recent changes added detailed comments explaining the "why", not just the "what".

7. **Link to docs from code** — When something is tricky, code comments link to relevant doc sections.

## 🤝 Contributing

See [../CONTRIBUTING.md](../CONTRIBUTING.md) for setup, development workflow, code style, and PR guidelines.

**Quick checklist before submitting:**
- [ ] Read GLOSSARY.md, CONTRIBUTING_GO.md, ARCHITECTURE.md
- [ ] Tests added/updated
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Comments added for exported functions
- [ ] Errors wrapped with `%w`

## 📚 External Resources

- **Effective Go:** https://golang.org/doc/effective_go
- **Go Code Review Comments:** https://github.com/golang/go/wiki/CodeReviewComments
- **Bubble Tea:** https://github.com/charmbracelet/bubbletea

## ❓ Questions?

- **Word not clear?** → GLOSSARY.md
- **How to do X in Go?** → CONTRIBUTING_GO.md
- **How does Y work?** → ARCHITECTURE.md
- **How do I set up dev?** → ../CONTRIBUTING.md
- **Found a bug?** → Open an issue on GitHub

---

**Happy learning! 🎉**

