package cmd

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestDisplayComparison_IdenticalProfiles(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
		},
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
		},
	}
	if got := DisplayComparison(p, p); got {
		t.Errorf("expected no differences, got true")
	}
}

func TestDisplayComparison_AddedFormula(t *testing.T) {
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
	if got := DisplayComparison(saved, current); !got {
		t.Errorf("expected differences, got false")
	}
}

func TestDisplayComparison_RemovedFormula(t *testing.T) {
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
	if got := DisplayComparison(saved, current); !got {
		t.Errorf("expected differences, got false")
	}
}

func TestDisplayComparison_VersionChange(t *testing.T) {
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
	if got := DisplayComparison(saved, current); !got {
		t.Errorf("expected differences, got false")
	}
}
