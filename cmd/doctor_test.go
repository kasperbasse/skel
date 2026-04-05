package cmd

import (
	"testing"

	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
	"github.com/kasperbasse/skel/internal/profile"
)

func TestBuildChecksEmpty(t *testing.T) {
	p := &profile.Profile{Name: "empty"}
	checks := appdoctor.BuildChecks(p)
	if len(checks) != 0 {
		t.Errorf("expected 0 checks for empty profile, got %d", len(checks))
	}
}

func TestBuildChecksHomebrew(t *testing.T) {
	p := &profile.Profile{
		Name: "brew",
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
		},
	}
	checks := appdoctor.BuildChecks(p)
	if len(checks) == 0 {
		t.Fatal("expected at least one check for profile with formulas")
	}
	found := false
	for _, c := range checks {
		if c.Label == "Homebrew" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a Homebrew check")
	}
}

func TestBuildChecksMas(t *testing.T) {
	p := &profile.Profile{
		Name: "mas",
		Homebrew: profile.HomebrewProfile{
			MasApps: []profile.MasApp{{ID: "497799835", Name: "Xcode"}},
		},
	}
	checks := appdoctor.BuildChecks(p)
	masFound := false
	for _, c := range checks {
		if c.Label == "mas (App Store)" {
			masFound = true
		}
	}
	if !masFound {
		t.Error("expected mas check when profile has App Store apps")
	}
}

func TestBuildChecksEditors(t *testing.T) {
	p := &profile.Profile{
		Name: "editors",
		Editor: profile.EditorProfile{
			VSCode: true,
			Cursor: true,
			Neovim: true,
		},
	}
	checks := appdoctor.BuildChecks(p)
	labels := make(map[string]bool)
	for _, c := range checks {
		labels[c.Label] = true
	}
	for _, want := range []string{"VS Code", "Cursor", "Neovim"} {
		if !labels[want] {
			t.Errorf("expected check for %q", want)
		}
	}
}

func TestBuildChecksLanguages(t *testing.T) {
	p := &profile.Profile{
		Name: "langs",
		Languages: profile.LanguageProfile{
			NodeVersion:     "20.0.0",
			NpmGlobals:      []string{"typescript"},
			YarnGlobals:     []string{"create-react-app"},
			PnpmGlobals:     []string{"turbo"},
			PipGlobals:      []string{"requests"},
			GemGlobals:      []string{"rails"},
			CargoPackages:   []string{"ripgrep"},
			ComposerGlobals: []string{"laravel/installer"},
		},
	}
	checks := appdoctor.BuildChecks(p)
	labels := make(map[string]bool)
	for _, c := range checks {
		labels[c.Label] = true
	}
	for _, want := range []string{"Node.js", "npm", "Yarn", "pnpm", "pip3", "gem (Ruby)", "cargo (Rust)", "Composer"} {
		if !labels[want] {
			t.Errorf("expected check for %q", want)
		}
	}
}

func TestBuildChecksHasFix(t *testing.T) {
	p := &profile.Profile{
		Name: "fix-test",
		Homebrew: profile.HomebrewProfile{
			MasApps: []profile.MasApp{{ID: "1", Name: "App"}},
		},
	}
	for _, c := range appdoctor.BuildChecks(p) {
		if c.Label == "mas (App Store)" && c.Fix == "" {
			t.Error("expected a non-empty fix hint for mas check")
		}
	}
}

func TestPluralS(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}
	for _, tt := range tests {
		if got := pluralS(tt.n); got != tt.want {
			t.Errorf("pluralS(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
