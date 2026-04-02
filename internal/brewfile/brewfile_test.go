package brewfile

import (
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestGenerate(t *testing.T) {
	h := profile.HomebrewProfile{
		Taps:     []string{"homebrew/cask-fonts"},
		Formulas: []string{"git", "ripgrep"},
		Casks:    []string{"firefox", "iterm2"},
		MasApps:  []profile.MasApp{{Name: "Xcode", ID: "497799835"}},
	}

	got := Generate(h)

	if !strings.Contains(got, `tap "homebrew/cask-fonts"`) {
		t.Error("missing tap line")
	}
	if !strings.Contains(got, `brew "git"`) {
		t.Error("missing brew git line")
	}
	if !strings.Contains(got, `brew "ripgrep"`) {
		t.Error("missing brew ripgrep line")
	}
	if !strings.Contains(got, `cask "firefox"`) {
		t.Error("missing cask firefox line")
	}
	if !strings.Contains(got, `mas "Xcode", id: 497799835`) {
		t.Error("missing mas line")
	}
}

func TestGenerateEmpty(t *testing.T) {
	got := Generate(profile.HomebrewProfile{})
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestParse(t *testing.T) {
	input := `tap "homebrew/cask-fonts"

brew "git"
brew "ripgrep"

cask "firefox"

mas "Xcode", id: 497799835
`
	h, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(h.Taps) != 1 || h.Taps[0] != "homebrew/cask-fonts" {
		t.Errorf("Taps = %v", h.Taps)
	}
	if len(h.Formulas) != 2 {
		t.Errorf("Formulas len = %d, want 2", len(h.Formulas))
	}
	if len(h.Casks) != 1 || h.Casks[0] != "firefox" {
		t.Errorf("Casks = %v", h.Casks)
	}
	if len(h.MasApps) != 1 || h.MasApps[0].Name != "Xcode" || h.MasApps[0].ID != "497799835" {
		t.Errorf("MasApps = %v", h.MasApps)
	}
}

func TestParseWithComments(t *testing.T) {
	input := `# This is my Brewfile
# Managed by skel

tap "homebrew/core"

# Development tools
brew "git"
# brew "svn"  # disabled
`
	h, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(h.Taps) != 1 {
		t.Errorf("Taps len = %d, want 1", len(h.Taps))
	}
	if len(h.Formulas) != 1 {
		t.Errorf("Formulas len = %d, want 1", len(h.Formulas))
	}
}

func TestParseBareNames(t *testing.T) {
	input := `brew git
cask firefox
`
	h, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(h.Formulas) != 1 || h.Formulas[0] != "git" {
		t.Errorf("Formulas = %v", h.Formulas)
	}
	if len(h.Casks) != 1 || h.Casks[0] != "firefox" {
		t.Errorf("Casks = %v", h.Casks)
	}
}

func TestParseWithOptions(t *testing.T) {
	input := `brew "imagemagick", args: ["--with-webp"]
brew "mysql", restart_service: true
`
	h, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(h.Formulas) != 2 {
		t.Errorf("Formulas len = %d, want 2", len(h.Formulas))
	}
	if h.Formulas[0] != "imagemagick" {
		t.Errorf("Formulas[0] = %q, want imagemagick", h.Formulas[0])
	}
}

func TestParseMasLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantApp profile.MasApp
		wantErr bool
	}{
		{
			"standard format",
			`mas "Xcode", id: 497799835`,
			profile.MasApp{Name: "Xcode", ID: "497799835"},
			false,
		},
		{
			"app with spaces",
			`mas "Final Cut Pro", id: 424389933`,
			profile.MasApp{Name: "Final Cut Pro", ID: "424389933"},
			false,
		},
		{
			"missing id",
			`mas "SomeApp"`,
			profile.MasApp{},
			true,
		},
		{
			"non-numeric id",
			`mas "SomeApp", id: abc`,
			profile.MasApp{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := parseMasLine(tt.line)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if app.Name != tt.wantApp.Name || app.ID != tt.wantApp.ID {
				t.Errorf("got %+v, want %+v", app, tt.wantApp)
			}
		})
	}
}

func TestParseInvalidPackageName(t *testing.T) {
	tests := []struct {
		name, input string
	}{
		{"shell injection semicolon", `brew "git; rm -rf /"` + "\n"},
		{"shell injection backtick", "brew \"`whoami`\"\n"},
		{"shell injection pipe", `brew "git | cat"` + "\n"},
		{"shell injection dollar", `brew "$(evil)"` + "\n"},
		{"path traversal", `brew "../../etc/passwd"` + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Error("expected error for unsafe package name")
			}
		})
	}
}

func TestParseEmptyFile(t *testing.T) {
	h, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h.Formulas) != 0 && h.Formulas != nil {
		t.Errorf("expected empty formulas, got %v", h.Formulas)
	}
}

func TestParseMalformedDirective(t *testing.T) {
	_, err := Parse("install git\n")
	if err == nil {
		t.Error("expected error for unknown directive")
	}
	if !strings.Contains(err.Error(), "unknown directive") {
		t.Errorf("error = %q, should mention unknown directive", err.Error())
	}
}

func TestParseFileTooLarge(t *testing.T) {
	huge := strings.Repeat("brew \"x\"\n", MaxBrewfileSize/9+1)
	_, err := Parse(huge)
	if err == nil {
		t.Error("expected error for oversized file")
	}
}

func TestRoundTrip(t *testing.T) {
	original := profile.HomebrewProfile{
		Taps:     []string{"homebrew/cask-fonts", "homebrew/services"},
		Formulas: []string{"git", "ripgrep", "fd"},
		Casks:    []string{"firefox", "iterm2"},
		MasApps:  []profile.MasApp{{Name: "Xcode", ID: "497799835"}},
	}

	generated := Generate(original)
	parsed, err := Parse(generated)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(parsed.Taps) != len(original.Taps) {
		t.Errorf("Taps: got %d, want %d", len(parsed.Taps), len(original.Taps))
	}
	if len(parsed.Formulas) != len(original.Formulas) {
		t.Errorf("Formulas: got %d, want %d", len(parsed.Formulas), len(original.Formulas))
	}
	if len(parsed.Casks) != len(original.Casks) {
		t.Errorf("Casks: got %d, want %d", len(parsed.Casks), len(original.Casks))
	}
	if len(parsed.MasApps) != len(original.MasApps) {
		t.Errorf("MasApps: got %d, want %d", len(parsed.MasApps), len(original.MasApps))
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"simple", "git", false},
		{"with slash", "homebrew/core", false},
		{"with at", "php@8.2", false},
		{"with dots", "node@18.0.1", false},
		{"with plus", "c++", false},
		{"empty", "", true},
		{"with space", "my package", true},
		{"with semicolon", "git;evil", true},
		{"with backtick", "`whoami`", true},
		{"with dollar", "$(cmd)", true},
		{"path traversal", "../etc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
