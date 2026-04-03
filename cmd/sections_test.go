package cmd

import (
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

func TestProfileItemCount(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep", "fzf"},
			Casks:    []string{"iterm2"},
		},
		Editor: profile.EditorProfile{
			VSCodeExts: []string{"golang.go", "esbenp.prettier-vscode"},
		},
	}
	n := profileItemCount(p)
	// 3 formulas + 1 cask + 2 VS Code extensions = 6
	if n != 6 {
		t.Errorf("profileItemCount = %d, want 6", n)
	}
}

func TestProfileItemCountEmpty(t *testing.T) {
	p := &profile.Profile{}
	if n := profileItemCount(p); n != 0 {
		t.Errorf("expected 0 for empty profile, got %d", n)
	}
}

func TestProfileSummaryParts(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
			Casks:    []string{"iterm2", "firefox"},
		},
	}
	parts := profileSummaryParts(p)
	if len(parts) == 0 {
		t.Fatal("expected non-empty summary parts")
	}
	// At least one part should mention Homebrew Formulas and one Casks.
	joined := strings.Join(parts, " ")
	if !strings.Contains(joined, "Formulas") {
		t.Errorf("expected Formulas in summary parts, got: %s", joined)
	}
	if !strings.Contains(joined, "Casks") {
		t.Errorf("expected Casks in summary parts, got: %s", joined)
	}
}

func TestProfileSummaryPartsEmpty(t *testing.T) {
	p := &profile.Profile{}
	if parts := profileSummaryParts(p); len(parts) != 0 {
		t.Errorf("expected empty parts for empty profile, got %v", parts)
	}
}

func TestAllRestoreKeys(t *testing.T) {
	keys := allRestoreKeys()
	if len(keys) == 0 {
		t.Fatal("expected non-empty restore keys")
	}
	// Each key must be non-empty and match what ScanGroups define.
	seen := make(map[string]bool)
	for _, k := range keys {
		if k == "" {
			t.Error("empty restore key")
		}
		if seen[k] {
			t.Errorf("duplicate restore key: %q", k)
		}
		seen[k] = true
	}
	// Required keys that the rest of the app depends on.
	for _, required := range []string{"homebrew", "shell", "git", "editors", "configs", "languages"} {
		if !seen[required] {
			t.Errorf("missing required restore key %q", required)
		}
	}
}

func TestParseOnlyFlag(t *testing.T) {
	t.Run("empty returns all", func(t *testing.T) {
		opts, err := parseOnlyFlag("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.ShouldRestore("homebrew") {
			t.Error("empty --only should restore all sections")
		}
	})

	t.Run("single valid section", func(t *testing.T) {
		opts, err := parseOnlyFlag("homebrew")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.ShouldRestore("homebrew") {
			t.Error("expected homebrew to be enabled")
		}
		if opts.ShouldRestore("shell") {
			t.Error("expected shell to be disabled")
		}
	})

	t.Run("multiple sections", func(t *testing.T) {
		opts, err := parseOnlyFlag("shell,git")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.ShouldRestore("shell") || !opts.ShouldRestore("git") {
			t.Error("expected shell and git to be enabled")
		}
		if opts.ShouldRestore("homebrew") {
			t.Error("expected homebrew to be disabled")
		}
	})

	t.Run("invalid section", func(t *testing.T) {
		_, err := parseOnlyFlag("notasection")
		if err == nil {
			t.Fatal("expected error for invalid section")
		}
	})

	t.Run("mixed valid and invalid", func(t *testing.T) {
		_, err := parseOnlyFlag("shell,notreal")
		if err == nil {
			t.Fatal("expected error when any section is invalid")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		opts, err := parseOnlyFlag("Homebrew")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.ShouldRestore("homebrew") {
			t.Error("expected case-insensitive match")
		}
	})

	t.Run("whitespace trimmed", func(t *testing.T) {
		opts, err := parseOnlyFlag(" shell , git ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !opts.ShouldRestore("shell") {
			t.Error("expected shell after whitespace trim")
		}
	})
}

func TestSummarizeBrew(t *testing.T) {
	h := profile.HomebrewProfile{
		Formulas: []string{"git", "ripgrep"},
		Casks:    []string{"iterm2"},
		MasApps:  []profile.MasApp{{ID: "1", Name: "Xcode"}},
		Taps:     []string{"homebrew/cask-fonts"},
	}
	s := summarizeBrew(h)
	if !strings.Contains(s, "2") {
		t.Errorf("expected formula count in summary: %q", s)
	}
	if !strings.Contains(s, "1") {
		t.Errorf("expected cask count in summary: %q", s)
	}
}

func TestSummarizeBrewNoTaps(t *testing.T) {
	h := profile.HomebrewProfile{
		Formulas: []string{"git"},
	}
	s := summarizeBrew(h)
	if strings.Contains(s, "taps") {
		t.Errorf("should not mention taps when none present: %q", s)
	}
}

func TestSummarizeShell(t *testing.T) {
	s := profile.ShellProfile{
		Shell:          "zsh",
		OhMyZsh:        true,
		OhMyZshTheme:   "robbyrussell",
		OhMyZshPlugins: []string{"git", "z", "fzf"},
		Aliases:        []string{"alias ll='ls -la'", "alias gs='git status'"},
	}
	result := summarizeShell(s)
	if !strings.Contains(result, "zsh") {
		t.Errorf("expected shell name in summary: %q", result)
	}
	if !strings.Contains(result, "3") {
		t.Errorf("expected plugin count in summary: %q", result)
	}
	if !strings.Contains(result, "2") {
		t.Errorf("expected alias count in summary: %q", result)
	}
}

var _ = restore.Options{} // keep import used
