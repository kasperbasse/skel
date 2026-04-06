# 💀 skel 

**The skeleton of your Mac environment. Captured, shared, and restored in minutes.**

<p>
    <img src="https://img.shields.io/github/go-mod/go-version/kasperbasse/skel?style=flat-square&logo=go&color=00ADD8&v=1" alt="Go Version">
    <img src="https://img.shields.io/github/license/kasperbasse/skel?style=flat-square&color=ff79c6&v=1" alt="License">
    <img src="https://img.shields.io/github/actions/workflow/status/kasperbasse/skel/ci.yml?style=flat-square&v=1" alt="Build Status">
</p>

Website: https://basse.dev/skel/

---

### 🤔 Why "skel"?

In Unix-like systems, `/etc/skel` is the "skeleton" directory used to initialize new user environments. **skel** brings that philosophy to the modern Mac.

It eliminates the "New Mac Headache" by capturing the "bones" of your setup - configs, packages, and settings - into a portable profile, then intelligently "re-fleshes" any Mac in minutes, installing only what's missing.

---

> [!IMPORTANT]
> **Early preview** - v0.x release. Things may change. [Report issues](https://github.com/kasperbasse/skel/issues) if something breaks.

---

## 📦 Installation

```bash
brew tap kasperbasse/tap
brew install skel
```

---

## 🚀 Quick Start

```bash
# On your current Mac - capture your setup
skel scan

# See what was saved
skel show default

# On a new Mac - check what's missing first
skel doctor default

# Then restore
skel restore default
```

---

## 🛠 Commands

### Profile

| Command             | Description                                                               |
|:--------------------|:--------------------------------------------------------------------------|
| `scan [profile]`    | Scan your Mac and save a profile (defaults to `default`)                  |
| `restore [profile]` | Interactive section picker, then restore - skips what's already installed |
| `list`              | Browse saved profiles (interactive) or print them (piped)                 |
| `show [profile]`    | Show full profile contents                                                |
| `update [profile]`  | Re-scan and update an existing profile                                    |
| `delete [profile]`  | Delete a saved profile                                                    |

### Inspect

| Command            | Description                                                                |
|:-------------------|:---------------------------------------------------------------------------|
| `status [profile]` | One-line summary: name, last scanned, item count - great for shell prompts |
| `drift [profile]`  | Show what's changed on this machine since the last scan                    |
| `diff [a] [b]`     | Compare two profiles side-by-side                                          |
| `doctor [profile]` | Check that all required tools are present before restoring                 |

### Share

| Command              | Description                                                              |
|:---------------------|:-------------------------------------------------------------------------|
| `export [profile]`   | Export a profile to a shareable JSON file                                |
| `import [file]`      | Import a profile from a JSON file                                        |
| `clone [source]`     | Clone a profile from a GitHub Gist (URL or `github:user/id`)             |
| `publish [profile]`  | Publish a profile as a GitHub Gist                                       |
| `brewfile export`    | Export Homebrew packages as a standard Brewfile                          |
| `brewfile import`    | Import a Brewfile into a profile                                         |

---

### Detailed Usage

<details>
<summary><b>doctor</b> ⚡ run this before restore on a new machine</summary>

`skel doctor` checks your machine against the profile - every tool that restore would need (Homebrew, mas, editors, language runtimes, package managers) - and tells you exactly what's missing and how to fix it. Nothing gets installed. It just checks.

**Run it first. Every time. On any machine you haven't restored to before.**

```bash
skel doctor             # checks default profile
skel doctor work-2026
```

```
  ✓  Homebrew
  ✓  Git
  ✗  mas (App Store)
       →  brew install mas
  ✗  Yarn
       →  npm install -g yarn

  ⚠ 2 issues found - install missing tools then run skel restore work-2026
```

Once all checks are green, `skel restore` will run cleanly with no mid-flight failures.
</details>

<details>
<summary><b>scan</b></summary>

Scans your Mac and saves a profile. Defaults to `default` if no name is given. Shows live progress as each section is scanned.

```bash
skel scan
skel scan work-2026
skel scan --force        # overwrite without confirmation
```
</details>

<details>
<summary><b>restore</b></summary>

Restores a profile on the current Mac. Opens an interactive section picker so you choose what to restore. Only installs what's missing.

```bash
skel restore work-2026
skel restore work-2026 --dry-run          # preview without making changes
skel restore work-2026 --only homebrew    # restore only Homebrew packages
skel restore work-2026 --only shell,git   # restore shell + git config
```

**Available `--only` sections:** `homebrew` `mas` `shell` `editors` `git` `languages` `configs` `defaults`
</details>

<details>
<summary><b>status</b></summary>

Prints a one-line summary of a profile. Fast - reads the saved file, no rescan. Useful in shell prompts or scripts.

```bash
skel status             # uses default profile
skel status work-2026
```

Output: `work-2026  3 days ago  247 items`
</details>

<details>
<summary><b>drift</b></summary>

Shows what has changed on your Mac since the last scan - new packages, removed tools, version bumps, config changes.

```bash
skel drift              # compare against default profile
skel drift work-2026
```

Run `skel update` to save the current state after reviewing drift.
</details>

<details>
<summary><b>clone / publish</b></summary>

Share profiles via GitHub Gists.

```bash
skel publish my-setup                                   # publish to a gist (PII redacted)
skel publish my-setup --no-redact                       # publish without redaction (not recommended)
skel clone https://gist.github.com/user/abc123          # clone from URL
skel clone github:user/abc123                           # clone via shorthand
skel clone github:user/abc123 --force                   # skip safety prompt
```

`publish` requires a GitHub token (`GITHUB_TOKEN` env var or `gh auth login`).
Before uploading, `skel` automatically redacts: **git name & email**, **raw gitconfig**, **hostname**, and **SSH key comments**. Shell config contents (`.zshrc`, aliases, etc.) are kept as-is since they are the primary value of a shared profile — review with `skel show` before publishing if yours contains tokens or personal paths.

`clone` works with public gists without authentication. Profiles containing shell or git configs show a warning — review with `skel show` before restoring.
</details>

---

## ⌨️ Shell Completions

Tab-complete profile names in `scan`, `show`, `restore`, `drift`, `update`, `delete`, `publish`, `diff`, `status`, and `doctor`.

```bash
# Zsh
echo 'source <(skel completion zsh)' >> ~/.zshrc

# Bash
echo 'source <(skel completion bash)' >> ~/.bashrc

# Fish
skel completion fish | source
```

---

## 📦 What Gets Saved

| Category         | Details                                                     |
|:-----------------|:------------------------------------------------------------|
| 🍺 Homebrew      | Taps, formulas, casks, Mac App Store apps                   |
| 🐚 Shell         | Zsh, Fish, Bash configs + plugins (Oh My Zsh, Starship)     |
| 💻 Editors       | VS Code, Cursor, Neovim, JetBrains IDEs (configs + plugins) |
| 🔧 Git           | `.gitconfig`, global `.gitignore`, user identity            |
| 🌐 Languages     | Node, Python, Go, Ruby, PHP, Rust, Java versions            |
| 📦 Packages      | npm, yarn, pnpm, pip, Composer, Ruby gems, Cargo globals    |
| ⚙️ Config files  | Any app config in `~/.config/` (auto-discovered)            |
| 🖥️ Defaults     | Dock, keyboard, trackpad, Finder, screenshot preferences    |
| 🔑 SSH keys      | Public fingerprints only - private keys are never read      |
| 🖥 System        | macOS version, hostname, architecture                       |

**JetBrains IDEs detected:** IntelliJ IDEA, WebStorm, GoLand, PyCharm, PhpStorm, CLion, RubyMine, Rider, DataGrip, RustRover, and more.

---

## 🛡️ Security & Privacy

- **Private keys never touched.** SSH private keys, `.env` files, passwords, and tokens are never read or stored.
- **Fingerprints only.** SSH key SHA256 fingerprints help you identify which keys to add manually on a new machine.
- **Safe restore.** Config files are written with `0600` permissions. Path traversal is blocked at validation time.
- **Import warnings.** Profiles with shell or git configs show a prominent warning before saving - always review with `skel show` first.
- **Publish safety.** `skel publish` automatically redacts git name/email, raw gitconfig, hostname, and SSH key comments before uploading. Shell config contents are kept (they are the point of sharing) — use `skel show` to review before publishing. Pass `--no-redact` to opt out.

---

## 🎨 Built With

- [Go](https://go.dev/)
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI engine
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

---

## 🤝 Contributing

Want to understand or modify the codebase quickly?

- **[DEVELOPMENT.md](./DEVELOPMENT.md)** — fast contributor path and daily workflow
- **[cmd/README.md](./cmd/README.md)** — how the flat `cmd/` package is grouped (`cmd_*`, `cli_*`, `section_*`, `ui_*`)
- **[docs/GLOSSARY.md](./docs/GLOSSARY.md)** — domain terms (10 min)
- **[docs/CONTRIBUTING_GO.md](./docs/CONTRIBUTING_GO.md)** — Go patterns and conventions (15 min)
- **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** — system design and component interactions (20 min)
- **[docs/README.md](./docs/README.md)** — full docs index

For a guided walkthrough, see the "Documentation" section in [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE)
