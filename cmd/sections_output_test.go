package cmd

import (
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

// ---------------------------------------------------------------------------
// wrapIfNotEmpty
// ---------------------------------------------------------------------------

func TestWrapIfNotEmpty(t *testing.T) {
	if got := wrapIfNotEmpty("v20"); len(got) != 1 || got[0] != "v20" {
		t.Errorf("wrapIfNotEmpty(%q) = %v, want [v20]", "v20", got)
	}
	if got := wrapIfNotEmpty(""); got != nil {
		t.Errorf("wrapIfNotEmpty(\"\") = %v, want nil", got)
	}
}

// ---------------------------------------------------------------------------
// summarizeVersions
// ---------------------------------------------------------------------------

func TestSummarizeVersionsWithVersions(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion:   "v20.0.0",
			GoVersion:     "go version go1.22 linux/amd64",
			PythonVersion: "Python 3.12.0",
			RubyVersion:   "ruby 3.2.0",
			PHPVersion:    "PHP 8.2.0",
			RustVersion:   "rustc 1.75.0",
			JavaVersion:   `openjdk version "17.0.8"`,
		},
	}
	s := summarizeVersions(p)
	for _, want := range []string{"Node", "Go", "Python", "Ruby", "PHP", "Rust", "Java"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in summarizeVersions output: %q", want, s)
		}
	}
}

func TestSummarizeVersionsWithPackages(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NpmGlobals:      []string{"typescript", "yarn"},
			YarnGlobals:     []string{"create-react-app"},
			PnpmGlobals:     []string{"turbo"},
			PipGlobals:      []string{"requests"},
			ComposerGlobals: []string{"laravel/installer"},
			GemGlobals:      []string{"rails", "bundler"},
			CargoPackages:   []string{"ripgrep", "bat", "fd"},
		},
	}
	s := summarizeVersions(p)
	for _, want := range []string{"NPM", "Yarn", "PNPM", "Pip", "Composer", "Ruby Gems", "Cargo"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in summarizeVersions output: %q", want, s)
		}
	}
}

func TestSummarizeVersionsEmpty(t *testing.T) {
	p := &profile.Profile{}
	if s := summarizeVersions(p); s != "" {
		t.Errorf("expected empty string for empty profile, got: %q", s)
	}
}

// ---------------------------------------------------------------------------
// printDiffSection
// ---------------------------------------------------------------------------

func TestPrintDiffSection(t *testing.T) {
	out := captureStdout(func() {
		printDiffSection("🍺", "Homebrew Formulas", []string{"ripgrep"}, []string{"wget"})
	})
	for _, want := range []string{"Homebrew Formulas", "ripgrep", "wget", "+", "-"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in printDiffSection output: %q", want, out)
		}
	}
}

func TestPrintDiffSectionNoChanges(t *testing.T) {
	out := captureStdout(func() {
		printDiffSection("🍺", "Homebrew", nil, nil)
	})
	if !strings.Contains(out, "Homebrew") {
		t.Errorf("expected title even with no changes: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showHomebrew
// ---------------------------------------------------------------------------

func TestShowHomebrew(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
			Casks:    []string{"iterm2"},
			MasApps:  []profile.MasApp{{ID: "497799835", Name: "Xcode"}},
			Taps:     []string{"homebrew/cask-fonts"},
		},
	}
	out := captureStdout(func() { showHomebrew(p) })
	for _, want := range []string{"Homebrew", "git", "ripgrep", "iterm2", "Xcode", "homebrew/cask-fonts"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showHomebrew output: %q", want, out)
		}
	}
}

func TestShowHomebrewMinimal(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
	}
	out := captureStdout(func() { showHomebrew(p) })
	if !strings.Contains(out, "git") {
		t.Errorf("expected 'git' in showHomebrew output: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showShell
// ---------------------------------------------------------------------------

func TestShowShell(t *testing.T) {
	p := &profile.Profile{
		Shell: profile.ShellProfile{
			Shell:          "zsh",
			OhMyZsh:        true,
			OhMyZshPlugins: []string{"git", "z"},
			Aliases:        []string{"alias ll='ls -la'"},
			FishPlugins:    []string{"bass"},
		},
	}
	out := captureStdout(func() { showShell(p) })
	for _, want := range []string{"Shell", "zsh"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showShell output: %q", want, out)
		}
	}
}

func TestShowShellIncludesFallbackDetailWhenNoPluginsOrAliases(t *testing.T) {
	p := &profile.Profile{Shell: profile.ShellProfile{Shell: "zsh"}}
	out := captureStdout(func() { showShell(p) })
	if !strings.Contains(out, "No plugins or aliases captured") {
		t.Errorf("expected fallback shell detail in output: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showEditors
// ---------------------------------------------------------------------------

func TestShowEditors(t *testing.T) {
	p := &profile.Profile{
		Editor: profile.EditorProfile{
			VSCode:     true,
			VSCodeExts: []string{"golang.go"},
			Cursor:     true,
			CursorExts: []string{"ext1"},
			Neovim:     true,
			NeovimPlugins: []profile.NeovimPlugin{
				{Name: "telescope.nvim", Source: "lazy"},
			},
			JetBrains: []profile.JetBrainsIDE{
				{Name: "GoLand", Version: "2025.1", Plugins: []string{"IdeaVim"}},
			},
		},
	}
	out := captureStdout(func() { showEditors(p) })
	for _, want := range []string{"VS Code", "Cursor", "Neovim", "GoLand"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showEditors output: %q", want, out)
		}
	}
}

func TestCountLabelPluralization(t *testing.T) {
	if got := countLabel(1, "formula", "formulas"); !strings.Contains(got, "formula") || strings.Contains(got, "formulas") {
		t.Errorf("countLabel singular = %q, want singular noun", got)
	}
	if got := countLabel(2, "formula", "formulas"); !strings.Contains(got, "formulas") {
		t.Errorf("countLabel plural = %q, want plural noun", got)
	}
}

func TestShowEditorsNone(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showEditors(p) })
	if out != "" {
		t.Errorf("expected no output for profile with no editors: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showGit
// ---------------------------------------------------------------------------

func TestShowGit(t *testing.T) {
	p := &profile.Profile{
		Git: profile.GitProfile{
			UserName:      "Alice",
			UserEmail:     "alice@example.com",
			DefaultBranch: "main",
		},
	}
	out := captureStdout(func() { showGit(p) })
	for _, want := range []string{"Git", "Alice", "alice@example.com", "main"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showGit output: %q", want, out)
		}
	}
}

func TestShowGitEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showGit(p) })
	if out != "" {
		t.Errorf("expected no output for empty git profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showLanguages
// ---------------------------------------------------------------------------

func TestShowLanguages(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
			GoVersion:   "go version go1.22 linux/amd64",
		},
	}
	out := captureStdout(func() { showLanguages(p) })
	if !strings.Contains(out, "Languages") {
		t.Errorf("expected 'Languages' in showLanguages output: %q", out)
	}
}

func TestShowLanguagesEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showLanguages(p) })
	if out != "" {
		t.Errorf("expected no output for empty languages profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showConfigs
// ---------------------------------------------------------------------------

func TestShowConfigs(t *testing.T) {
	p := &profile.Profile{
		ConfigFiles: map[string]string{
			".config/kitty/kitty.conf": "font_size 14",
			".tmux.conf":               "set -g prefix C-a",
		},
	}
	out := captureStdout(func() { showConfigs(p) })
	if !strings.Contains(out, "Config") {
		t.Errorf("expected 'Config' in showConfigs output: %q", out)
	}
}

func TestShowConfigsEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showConfigs(p) })
	if out != "" {
		t.Errorf("expected no output for empty config files: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showDefaults
// ---------------------------------------------------------------------------

func TestShowDefaults(t *testing.T) {
	p := &profile.Profile{
		Defaults: profile.DefaultsProfile{
			Settings: []profile.DefaultsSetting{
				{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Value: "36"},
			},
		},
	}
	out := captureStdout(func() { showDefaults(p) })
	for _, want := range []string{"Defaults", "com.apple.dock", "tilesize"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showDefaults output: %q", want, out)
		}
	}
}

func TestShowDefaultsEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showDefaults(p) })
	if out != "" {
		t.Errorf("expected no output for empty defaults profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// showSSH
// ---------------------------------------------------------------------------

func TestShowSSH(t *testing.T) {
	p := &profile.Profile{
		SSH: profile.SSHProfile{
			Keys: []profile.SSHKey{
				{Filename: "id_ed25519", Type: "ED25519", Fingerprint: "SHA256:abc"},
			},
		},
	}
	out := captureStdout(func() { showSSH(p) })
	for _, want := range []string{"SSH", "id_ed25519"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in showSSH output: %q", want, out)
		}
	}
}

func TestShowSSHEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { showSSH(p) })
	if out != "" {
		t.Errorf("expected no output for empty SSH profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunHomebrew
// ---------------------------------------------------------------------------

func TestDryRunHomebrew(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
			Casks:    []string{"iterm2"},
			Taps:     []string{"homebrew/cask-fonts"},
			MasApps:  []profile.MasApp{{ID: "497799835", Name: "Xcode"}},
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunHomebrew(p, opts) })
	for _, want := range []string{"brew install git", "brew install --cask iterm2", "brew tap homebrew/cask-fonts", "mas install"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dryRunHomebrew output: %q", want, out)
		}
	}
}

func TestDryRunHomebrewSkipped(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{Formulas: []string{"git"}},
	}
	opts, _ := parseOnlyFlag("shell") // only=shell => homebrew skipped
	out := captureStdout(func() { dryRunHomebrew(p, opts) })
	if strings.Contains(out, "brew install") {
		t.Errorf("expected no homebrew output when section skipped: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunShell
// ---------------------------------------------------------------------------

func TestDryRunShell(t *testing.T) {
	p := &profile.Profile{
		Shell: profile.ShellProfile{
			ZshrcContent:       "# zsh",
			Starship:           true,
			FishConfig:         "# fish",
			BashrcContent:      "# bash",
			BashProfileContent: "# bash_profile",
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunShell(p, opts) })
	for _, want := range []string{".zshrc", "starship.toml", "config.fish", ".bashrc", ".bash_profile"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dryRunShell output: %q", want, out)
		}
	}
}

func TestDryRunShellSkipped(t *testing.T) {
	p := &profile.Profile{Shell: profile.ShellProfile{ZshrcContent: "# zsh"}}
	opts, _ := parseOnlyFlag("homebrew")
	out := captureStdout(func() { dryRunShell(p, opts) })
	if strings.Contains(out, ".zshrc") {
		t.Errorf("expected no shell output when section skipped: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunEditors
// ---------------------------------------------------------------------------

func TestDryRunEditors(t *testing.T) {
	p := &profile.Profile{
		Editor: profile.EditorProfile{
			VSCode:     true,
			VSCodeExts: []string{"golang.go"},
			Cursor:     true,
			CursorExts: []string{"ext1"},
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunEditors(p, opts) })
	for _, want := range []string{"VS Code", "Cursor"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dryRunEditors output: %q", want, out)
		}
	}
}

func TestDryRunEditorsSkipped(t *testing.T) {
	p := &profile.Profile{Editor: profile.EditorProfile{VSCode: true}}
	opts, _ := parseOnlyFlag("homebrew")
	out := captureStdout(func() { dryRunEditors(p, opts) })
	if strings.Contains(out, "VS Code") {
		t.Errorf("expected no editor output when section skipped: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunGit
// ---------------------------------------------------------------------------

func TestDryRunGit(t *testing.T) {
	p := &profile.Profile{
		Git: profile.GitProfile{GitConfigContent: "[user]\n\tname = Alice"},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunGit(p, opts) })
	if !strings.Contains(out, ".gitconfig") {
		t.Errorf("expected .gitconfig in dryRunGit output: %q", out)
	}
}

func TestDryRunGitNoContent(t *testing.T) {
	p := &profile.Profile{Git: profile.GitProfile{UserName: "Alice"}}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunGit(p, opts) })
	if strings.Contains(out, ".gitconfig") {
		t.Errorf("expected no .gitconfig when GitConfigContent is empty: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunLanguages
// ---------------------------------------------------------------------------

func TestDryRunLanguages(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NpmGlobals:      []string{"typescript"},
			YarnGlobals:     []string{"create-react-app"},
			PnpmGlobals:     []string{"turbo"},
			ComposerGlobals: []string{"laravel/installer"},
			GemGlobals:      []string{"rails"},
			CargoPackages:   []string{"ripgrep"},
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunLanguages(p, opts) })
	for _, want := range []string{"npm install -g", "yarn global add", "pnpm add -g", "composer global require", "gem install", "cargo install"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dryRunLanguages output: %q", want, out)
		}
	}
}

func TestDryRunLanguagesEmpty(t *testing.T) {
	p := &profile.Profile{}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunLanguages(p, opts) })
	if out != "" {
		t.Errorf("expected no output for empty languages: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunConfigs
// ---------------------------------------------------------------------------

func TestDryRunConfigs(t *testing.T) {
	p := &profile.Profile{
		ConfigFiles: map[string]string{
			".config/kitty/kitty.conf": "font_size 14",
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunConfigs(p, opts) })
	if !strings.Contains(out, ".config/kitty/kitty.conf") {
		t.Errorf("expected config path in dryRunConfigs output: %q", out)
	}
}

func TestDryRunConfigsEmpty(t *testing.T) {
	p := &profile.Profile{}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunConfigs(p, opts) })
	if out != "" {
		t.Errorf("expected no output for empty config files: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunDefaults
// ---------------------------------------------------------------------------

func TestDryRunDefaults(t *testing.T) {
	p := &profile.Profile{
		Defaults: profile.DefaultsProfile{
			Settings: []profile.DefaultsSetting{
				{Domain: "com.apple.dock", Key: "tilesize", Type: "int", Value: "36"},
			},
		},
	}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunDefaults(p, opts) })
	if !strings.Contains(out, "defaults write com.apple.dock tilesize -int 36") {
		t.Errorf("expected defaults write command in output: %q", out)
	}
}

func TestDryRunDefaultsEmpty(t *testing.T) {
	p := &profile.Profile{}
	opts, _ := parseOnlyFlag("")
	out := captureStdout(func() { dryRunDefaults(p, opts) })
	if out != "" {
		t.Errorf("expected no output for empty defaults profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// dryRunSSH
// ---------------------------------------------------------------------------

func TestDryRunSSH(t *testing.T) {
	p := &profile.Profile{
		SSH: profile.SSHProfile{
			Keys: []profile.SSHKey{
				{Filename: "id_ed25519", Type: "ED25519"},
			},
		},
	}
	opts := &restore.Options{}
	out := captureStdout(func() { dryRunSSH(p, opts) })
	for _, want := range []string{"SSH", "id_ed25519", "manual"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in dryRunSSH output: %q", want, out)
		}
	}
}

func TestDryRunSSHEmpty(t *testing.T) {
	p := &profile.Profile{}
	opts := &restore.Options{}
	out := captureStdout(func() { dryRunSSH(p, opts) })
	if out != "" {
		t.Errorf("expected no output for empty SSH profile: %q", out)
	}
}

// ---------------------------------------------------------------------------
// importWarningsShell / importWarningsGit
// ---------------------------------------------------------------------------

func TestImportWarningsShell(t *testing.T) {
	p := &profile.Profile{
		Shell: profile.ShellProfile{
			ZshrcContent:       "# zsh",
			BashrcContent:      "# bash",
			BashProfileContent: "# bash_profile",
			FishConfig:         "# fish",
		},
	}
	warnings := importWarningsShell(p)
	if len(warnings) != 4 {
		t.Errorf("expected 4 shell warnings, got %d: %v", len(warnings), warnings)
	}
}

func TestImportWarningsShellEmpty(t *testing.T) {
	p := &profile.Profile{}
	warnings := importWarningsShell(p)
	if len(warnings) != 0 {
		t.Errorf("expected 0 shell warnings for empty profile, got %d", len(warnings))
	}
}

func TestImportWarningsGit(t *testing.T) {
	p := &profile.Profile{
		Git: profile.GitProfile{GitConfigContent: "[user]\n\tname = Alice"},
	}
	warnings := importWarningsGit(p)
	if len(warnings) != 1 || warnings[0] != ".gitconfig" {
		t.Errorf("expected [.gitconfig] git warning, got %v", warnings)
	}
}

func TestImportWarningsGitEmpty(t *testing.T) {
	p := &profile.Profile{}
	warnings := importWarningsGit(p)
	if len(warnings) != 0 {
		t.Errorf("expected 0 git warnings for empty profile, got %d", len(warnings))
	}
}
