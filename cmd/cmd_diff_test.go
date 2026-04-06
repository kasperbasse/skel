package cmd

import (
	"bytes"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestDisplayComparisonIdenticalProfiles(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
		},
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
		},
	}
	if got := displayComparison(p, p); got {
		t.Errorf("expected no differences, got true")
	}
}

func TestDisplayComparisonAddedFormula(t *testing.T) {
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
	output := captureOutput(func() {
		displayComparison(saved, current)
	})
	if !bytes.Contains([]byte(output), []byte("+")) {
		t.Errorf("expected '+' marker in output, got: %q", output)
	}
}

func TestDisplayComparisonRemovedFormula(t *testing.T) {
	saved := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
		},
	}
	current := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
		},
	}
	output := captureOutput(func() {
		displayComparison(saved, current)
	})
	if !bytes.Contains([]byte(output), []byte("-")) {
		t.Errorf("expected '-' marker in output, got: %q", output)
	}
}

func TestDisplayComparisonVersionChange(t *testing.T) {
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
	output := captureOutput(func() {
		displayComparison(saved, current)
	})
	if !bytes.Contains([]byte(output), []byte("~")) {
		t.Errorf("expected '~' marker in output, got: %q", output)
	}
}
