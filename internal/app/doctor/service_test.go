package doctor

import (
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestRequiredToolsEmpty(t *testing.T) {
	p := &profile.Profile{Name: "empty"}
	if got := RequiredTools(p); len(got) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(got))
	}
}

func TestRequiredToolsOrderAndUniqueness(t *testing.T) {
	p := &profile.Profile{
		Name: "all",
		Homebrew: profile.HomebrewProfile{
			Formulas: []string{"git"},
			MasApps:  []profile.MasApp{{ID: "1", Name: "Xcode"}},
		},
		Editor: profile.EditorProfile{
			VSCode: true,
			Cursor: true,
			Neovim: true,
		},
		Git: profile.GitProfile{
			UserName: "Kasper",
		},
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

	want := []string{"brew", "mas", "code", "cursor", "nvim", "git", "node", "npm", "yarn", "pnpm", "pip3", "gem", "cargo", "composer"}
	got := RequiredTools(p)
	if len(got) != len(want) {
		t.Fatalf("unexpected length: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("index %d: got %q want %q", i, got[i], want[i])
		}
	}
}
