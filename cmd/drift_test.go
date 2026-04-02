package cmd

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestComputeDriftNoChanges(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
			Casks:    []string{"firefox"},
		},
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
		},
	}

	changes := computeDrift(p, p)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes, got %d", len(changes))
	}
}

func TestComputeDriftAddedFormula(t *testing.T) {
	saved := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
	}
	current := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
		},
	}

	changes := computeDrift(saved, current)
	if len(changes) != 1 {
		t.Fatalf("expected 1 section, got %d", len(changes))
	}
	if len(changes[0].added) != 1 || changes[0].added[0] != "ripgrep" {
		t.Errorf("expected ripgrep added, got %v", changes[0].added)
	}
}

func TestComputeDriftRemovedCask(t *testing.T) {
	saved := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Casks: []string{"firefox", "iterm2"},
		},
	}
	current := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Casks: []string{"firefox"},
		},
	}

	changes := computeDrift(saved, current)
	if len(changes) != 1 {
		t.Fatalf("expected 1 section, got %d", len(changes))
	}
	if len(changes[0].removed) != 1 || changes[0].removed[0] != "iterm2" {
		t.Errorf("expected iterm2 removed, got %v", changes[0].removed)
	}
}

func TestComputeDriftSkipsEmptyCurrent(t *testing.T) {
	// When current scan returns empty (tool not in PATH), don't report as mass removal.
	saved := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep", "fd"},
		},
	}
	current := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: nil,
		},
	}

	changes := computeDrift(saved, current)
	for _, c := range changes {
		if c.title == "Homebrew Formulas" {
			t.Error("should not report formulas drift when current is empty")
		}
	}
}

func TestComputeDriftVersionChange(t *testing.T) {
	saved := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion: "v18.0.0",
		},
	}
	current := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
		},
	}

	changes := computeDrift(saved, current)
	found := false
	for _, c := range changes {
		if c.title == "Language Versions" {
			found = true
			if len(c.added) != 1 {
				t.Errorf("expected 1 version change, got %d", len(c.added))
			}
		}
	}
	if !found {
		t.Error("expected Language Versions section")
	}
}

func TestComputeDriftVersionEmptyCurrentIgnored(t *testing.T) {
	saved := &profile.Profile{
		Languages: profile.LanguageProfile{
			GoVersion: "go1.22",
		},
	}
	current := &profile.Profile{
		Languages: profile.LanguageProfile{
			GoVersion: "", // tool not in PATH
		},
	}

	changes := computeDrift(saved, current)
	for _, c := range changes {
		if c.title == "Language Versions" {
			t.Error("should not report version drift when current is empty")
		}
	}
}

func TestComputeDriftShellModified(t *testing.T) {
	saved := &profile.Profile{
		Shell: profile.ShellProfile{
			ZshrcContent: "old content",
		},
	}
	current := &profile.Profile{
		Shell: profile.ShellProfile{
			ZshrcContent: "new content",
		},
	}

	changes := computeDrift(saved, current)
	found := false
	for _, c := range changes {
		if c.title == "Shell Config" {
			found = true
		}
	}
	if !found {
		t.Error("expected Shell Config section for modified .zshrc")
	}
}

func TestComputeDriftConfigFileAdded(t *testing.T) {
	saved := &profile.Profile{
		ConfigFiles: map[string]string{},
	}
	current := &profile.Profile{
		ConfigFiles: map[string]string{
			".config/kitty/kitty.conf": "font_size 14",
		},
	}

	changes := computeDrift(saved, current)
	found := false
	for _, c := range changes {
		if c.title == "Config Files" {
			found = true
			if len(c.added) != 1 {
				t.Errorf("expected 1 added config, got %d", len(c.added))
			}
		}
	}
	if !found {
		t.Error("expected Config Files section")
	}
}

func TestCountDriftItems(t *testing.T) {
	sections := []driftSection{
		{added: []string{"a", "b"}, removed: []string{"c"}},
		{added: []string{"d"}},
	}
	if n := countDriftItems(sections); n != 4 {
		t.Errorf("countDriftItems = %d, want 4", n)
	}

	if n := countDriftItems(nil); n != 0 {
		t.Errorf("countDriftItems(nil) = %d, want 0", n)
	}
}

func TestDiffSlices(t *testing.T) {
	added, removed := diffSlices(
		[]string{"a", "b", "c"},
		[]string{"b", "c", "d"},
	)
	if len(added) != 1 {
		t.Errorf("added len = %d, want 1", len(added))
	}
	if len(removed) != 1 {
		t.Errorf("removed len = %d, want 1", len(removed))
	}
}
