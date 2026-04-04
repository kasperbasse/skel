package cmd

import (
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

// ---------------------------------------------------------------------------
// shortVer
// ---------------------------------------------------------------------------

func TestShortVer(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "none"},
		{"v20.0.0", "v20.0.0"},
		{"go version go1.22 linux/amd64", "1.22"},
		{"Python 3.12.0", "3.12.0"},
		{"ruby 3.2.0 (2023-03-30 revision e51014f9c0)", "3.2.0"},
		{"rustc 1.75.0 (82e1608df 2023-12-21)", "1.75.0"},
		{"20.0.0", "20.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shortVer(tt.input)
			if got != tt.want {
				t.Errorf("shortVer(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// printUpdateDiff
// ---------------------------------------------------------------------------

func TestPrintUpdateDiffChanges(t *testing.T) {
	old := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
		Languages: profile.LanguageProfile{NodeVersion: "v18.0.0"},
	}
	updated := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep", "fzf"},
		},
		Languages: profile.LanguageProfile{NodeVersion: "v20.0.0"},
	}
	out := captureStdout(func() { printUpdateDiff(old, updated) })
	if !strings.Contains(out, "Homebrew Formulas") {
		t.Errorf("expected 'Homebrew Formulas' in printUpdateDiff output: %q", out)
	}
}

func TestPrintUpdateDiffNoChanges(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
	}
	out := captureStdout(func() { printUpdateDiff(p, p) })
	// No change => no output
	if out != "" {
		t.Errorf("expected no output when profiles are identical, got: %q", out)
	}
}

func TestPrintUpdateDiffVersionAdded(t *testing.T) {
	old := &profile.Profile{}
	updated := &profile.Profile{
		Languages: profile.LanguageProfile{GoVersion: "go version go1.22 linux/amd64"},
	}
	out := captureStdout(func() { printUpdateDiff(old, updated) })
	if !strings.Contains(out, "Go") {
		t.Errorf("expected 'Go' in diff when version added: %q", out)
	}
}

func TestPrintUpdateDiffVersionRemoved(t *testing.T) {
	old := &profile.Profile{
		Languages: profile.LanguageProfile{GoVersion: "go version go1.22 linux/amd64"},
	}
	updated := &profile.Profile{}
	out := captureStdout(func() { printUpdateDiff(old, updated) })
	if !strings.Contains(out, "Go") {
		t.Errorf("expected 'Go' in diff when version removed: %q", out)
	}
}

// ---------------------------------------------------------------------------
// collectImportWarnings
// ---------------------------------------------------------------------------

func TestCollectImportWarnings(t *testing.T) {
	p := &profile.Profile{
		Shell: profile.ShellProfile{ZshrcContent: "# zsh"},
		Git:   profile.GitProfile{GitConfigContent: "[user]"},
	}
	warnings := collectImportWarnings(p)
	if len(warnings) == 0 {
		t.Error("expected import warnings for profile with shell and git content")
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, ".zshrc") || strings.Contains(w, "gitconfig") || strings.Contains(w, ".gitconfig") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected zshrc or gitconfig warning, got %v", warnings)
	}
}

func TestCollectImportWarningsEmpty(t *testing.T) {
	p := &profile.Profile{Name: "clean"}
	warnings := collectImportWarnings(p)
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings for clean profile, got %d: %v", len(warnings), warnings)
	}
}
