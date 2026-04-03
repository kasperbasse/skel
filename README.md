# skel 💀

**The skeleton of your Mac environment. Captured, shared, and restored in minutes.**

<p>
    <img src="https://img.shields.io/github/go-mod/go-version/kasperbasse/skel?style=flat-square&logo=go&color=00ADD8&v=1" alt="Go Version">
    <img src="https://img.shields.io/github/license/kasperbasse/skel?style=flat-square&color=ff79c6&v=1" alt="License">
    <img src="https://img.shields.io/github/actions/workflow/status/kasperbasse/skel/ci.yml?style=flat-square&v=1" alt="Build Status">
</p>

---

### 🤔 Why "skel"?

In Unix-like systems, `/etc/skel` is the "skeleton" directory used to initialize new user environments. **skel** brings that philosophy to the modern Mac.

It eliminates the "New Mac Headache" by capturing the "bones" of your setup—configs, packages, and settings. It creates a portable profile of your environment and intelligently "re-fleshes" any Mac in minutes, installing only what is missing.

---

> [!IMPORTANT]
> **Early preview** - this is a v0.x release. Things may change. [Report issues](https://github.com/kasperbasse/skel/issues) if something breaks.

---

## 📦 Installation

### Homebrew

```bash
brew tap kasperbasse/tap
brew install skel
```

---

## 🚀 Quick Start

```bash
# 1. Capture your current setup
skel scan

# 2. View what was found
skel show default

# 3. Restore on a new machine
skel restore default
```

---

## 🛠 Commands

| Command           | Description                                                                 |
|:------------------|:----------------------------------------------------------------------------|
| list              | Lists all saved profiles. Use arrow keys to browse, enter to show details.  |
| show [profile]    | Shows the full contents of a profile.                                       |
| scan [profile]    | Scans your Mac and saves a profile (defaults to "default").                 |
| restore [profile] | Interactive section picker, then restores - skips what's already installed. |
| update [profile]  | Re-scans and updates an existing profile.                                   |
| drift [profile]   | Detects changes since the last scan.                                        |
| delete [profile]  | Deletes a saved profile.                                                    |
| diff [a] [b]      | Compares two profiles side-by-side.                                         |
| export [profile]  | Exports a profile to a shareable JSON file.                                 |
| import [file]     | Imports a profile from a JSON file.                                         |
| clone [source]    | Clone a profile from a GitHub Gist (URL or github:user/id).                 |
| publish [profile] | Publish a profile as a GitHub Gist (requires `GITHUB_TOKEN` or `gh` CLI).   |
| brewfile export   | Exports Homebrew packages as a standard Brewfile.                           |
| brewfile import   | Imports a Brewfile into a profile.                                          |

### Advanced Management

<details>
<summary><b>View detailed command usage</b></summary>

### skel scan [profile-name]
Scans your Mac and saves a profile. Defaults to "default" if no name is given.

```bash
skel scan work-2026
skel scan --force  # overwrite without confirmation
```

### skel restore [profile-name]
Restores a profile on the current Mac. Only installs what's missing.

```bash
skel restore work-2026
skel restore work-2026 --dry-run         # preview without changes
skel restore work-2026 --only homebrew   # restore only Homebrew packages
skel restore work-2026 --only shell,git  # restore shell + git config
```

**Available --only sections:** homebrew, mas, shell, editors, git, languages, configs, defaults

### skel drift [profile-name]
Detects what's changed on your Mac since the last scan.

```bash
skel drift            # compare against "default" profile
skel drift work-2026  # compare against a specific profile
```

### skel clone / publish
Share profiles via GitHub Gists.

```bash
skel publish my-setup                                  # publish to a gist
skel clone https://gist.github.com/user/abc123         # clone from URL
skel clone github:user/abc123                          # clone via shorthand
skel clone github:user/abc123 --force                  # skip safety prompt
```

`publish` requires a GitHub token (`GITHUB_TOKEN` env var or `gh auth login`).
`clone` works with public gists without authentication. If the profile contains shell or git configs, you'll be prompted to confirm before saving.

### skel brewfile export [profile-name]
Exports Homebrew packages as a standard Brewfile.

```bash
skel brewfile export work-2026              # creates Brewfile
skel brewfile export work-2026 -o dev.brew  # custom filename
```

</details>

---

## 🔍 What Gets Saved

| Category         | Details                                                     |
|:-----------------|:------------------------------------------------------------|
| 🍺 Homebrew      | Taps, formulas, casks, Mac App Store apps                   |
| 🐚 Shell         | Zsh, Fish, and Bash configs + plugins (Oh My Zsh, Starship) |
| 💻 Editors       | VS Code, Cursor, Neovim, JetBrains IDEs (configs + plugins) |
| 🔧 Git           | .gitconfig, global .gitignore, user identity                |
| 🌐 Languages     | Node, Python, Go, Ruby, PHP, Rust, Java versions            |
| 📦 Packages      | npm, yarn, pnpm, pip, Composer, Ruby gems, Cargo globals    |
| ⚙️ Config files  | Any app config in ~/.config/ (auto-discovered)              |
| 🖥️ Defaults     | Dock, keyboard, trackpad, Finder, screenshot preferences    |
| 🔑 SSH keys      | Public fingerprints only (private keys are never read)      |
| 🖥 System        | macOS version, hostname, architecture                       |

### JetBrains IDEs

`skel` detects and backs up configs for: IntelliJ IDEA, WebStorm, GoLand, PyCharm, PhpStorm, CLion, RubyMine, Rider, DataGrip, RustRover, and more.

---

## 🛡️ Security & Privacy

* **Zero-Knowledge:** Private SSH keys, .env files, passwords, and tokens are never read or stored.
* **Fingerprinting:** Only public key SHA256 fingerprints are stored to help you identify which keys to re-add manually.
* **Safe Restore:** Config files are restored with 0600 permissions (owner-only). Path traversal is blocked.
* **Validation:** Imported and cloned profiles are validated before saving.
* **Publish safety:** `skel publish` redacts your hostname before uploading.
* **Import warnings:** Profiles containing shell/git configs show a prominent warning - review before restoring.

---

## 🧠 Smart Restore

Unlike a simple script, `skel` checks your system state first:
* Interactive section picker lets you choose what to restore before it starts.
* Homebrew packages already present are skipped to avoid errors.
* VS Code extensions already installed are skipped.
* macOS preferences (Dock, keyboard, Finder, etc.) are captured and restored.
* If core tools like `brew` are missing, `skel` provides helpful instructions instead of failing silently.

---

## 🎨 Built With

* [Go](https://go.dev/)
* [Cobra](https://github.com/spf13/cobra) - CLI framework
* [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI engine
* [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling

---

## License

[MIT](LICENSE)