package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

// captureStdout redirects os.Stdout to a pipe, runs fn, and returns the captured output.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// ---------------------------------------------------------------------------
// hasEditors
// ---------------------------------------------------------------------------

func TestHasEditors(t *testing.T) {
	tests := []struct {
		name string
		e    profile.EditorProfile
		want bool
	}{
		{"none", profile.EditorProfile{}, false},
		{"vscode", profile.EditorProfile{VSCode: true}, true},
		{"cursor", profile.EditorProfile{Cursor: true}, true},
		{"neovim", profile.EditorProfile{Neovim: true}, true},
		{"jetbrains", profile.EditorProfile{JetBrains: []profile.JetBrainsIDE{{Name: "GoLand"}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasEditors(tt.e); got != tt.want {
				t.Errorf("hasEditors() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// summarizeEditors
// ---------------------------------------------------------------------------

func TestSummarizeEditors(t *testing.T) {
	e := profile.EditorProfile{
		VSCode:     true,
		VSCodeExts: []string{"golang.go", "esbenp.prettier-vscode"},
		Cursor:     true,
		CursorExts: []string{"ext1"},
		Neovim:     true,
		NeovimPlugins: []profile.NeovimPlugin{
			{Name: "nvim-treesitter"},
			{Name: "telescope.nvim"},
		},
		JetBrains: []profile.JetBrainsIDE{
			{Name: "GoLand", Version: "2025.1", Plugins: []string{"IdeaVim", "Go"}},
		},
	}
	s := summarizeEditors(e)
	for _, want := range []string{"VS Code", "Cursor", "Neovim", "GoLand"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in summarizeEditors output: %q", want, s)
		}
	}
}

func TestSummarizeEditorsNoEditors(t *testing.T) {
	s := summarizeEditors(profile.EditorProfile{})
	if s != "" {
		t.Errorf("expected empty string for no editors, got: %q", s)
	}
}

func TestSummarizeEditorsNeovimNoPlugins(t *testing.T) {
	e := profile.EditorProfile{Neovim: true}
	s := summarizeEditors(e)
	if !strings.Contains(s, "Neovim") {
		t.Errorf("expected Neovim in output: %q", s)
	}
}

// ---------------------------------------------------------------------------
// summarizeGit
// ---------------------------------------------------------------------------

func TestSummarizeGit(t *testing.T) {
	tests := []struct {
		name string
		git  profile.GitProfile
		want string
	}{
		{
			name: "empty",
			git:  profile.GitProfile{},
			want: "",
		},
		{
			name: "name and email",
			git: profile.GitProfile{
				UserName:  "Kasper",
				UserEmail: "kasper@example.com",
			},
			want: "Kasper",
		},
		{
			name: "default branch only",
			git: profile.GitProfile{
				DefaultBranch: "main",
			},
			want: "Default branch",
		},
		{
			name: "config only",
			git: profile.GitProfile{
				GitConfigContent: "[user]",
			},
			want: "Git configuration present",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := summarizeGit(tt.git)
			if !strings.Contains(got, tt.want) {
				t.Errorf("summarizeGit() = %q, want substring %q", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatSSHKey
// ---------------------------------------------------------------------------

func TestFormatSSHKey(t *testing.T) {
	key := profile.SSHKey{
		Filename:    "id_ed25519",
		Type:        "ED25519",
		Fingerprint: "SHA256:abc123",
		Comment:     "user@example.com",
	}
	s := formatSSHKey(key)
	for _, want := range []string{"id_ed25519", "ED25519", "SHA256:abc123", "user@example.com"} {
		if !strings.Contains(s, want) {
			t.Errorf("expected %q in formatSSHKey output: %q", want, s)
		}
	}
}

func TestFormatSSHKeyPublicOnly(t *testing.T) {
	key := profile.SSHKey{Filename: "id_rsa", PublicOnly: true}
	s := formatSSHKey(key)
	if !strings.Contains(s, "pub only") {
		t.Errorf("expected 'pub only' flag in output: %q", s)
	}
}

func TestFormatSSHKeyMinimal(t *testing.T) {
	key := profile.SSHKey{Filename: "id_ed25519"}
	s := formatSSHKey(key)
	if !strings.Contains(s, "id_ed25519") {
		t.Errorf("expected filename in output: %q", s)
	}
}

// ---------------------------------------------------------------------------
// normalizeIcon
// ---------------------------------------------------------------------------

func TestNormalizeIcon(t *testing.T) {
	if got := normalizeIcon("  🐹  "); got != "🐹" {
		t.Errorf("normalizeIcon() = %q, want %q", got, "🐹")
	}
	if got := normalizeIcon("🐹"); got != "🐹" {
		t.Errorf("normalizeIcon() = %q, want %q", got, "🐹")
	}
}

// ---------------------------------------------------------------------------
// dryRunBullet / printSection / printBullet / printRow
// ---------------------------------------------------------------------------

func TestDryRunBullet(t *testing.T) {
	out := captureStdout(func() { dryRunBullet("brew install git") })
	if !strings.Contains(out, "brew install git") {
		t.Errorf("expected command in dryRunBullet output: %q", out)
	}
	if !strings.Contains(out, "$") {
		t.Errorf("expected '$' prefix in dryRunBullet output: %q", out)
	}
}

func TestPrintSection(t *testing.T) {
	out := captureStdout(func() { printSection("🍺", "Homebrew") })
	if !strings.Contains(out, "Homebrew") {
		t.Errorf("expected title in printSection output: %q", out)
	}
}

func TestPrintBullet(t *testing.T) {
	out := captureStdout(func() { printBullet("some item") })
	if !strings.Contains(out, "some item") {
		t.Errorf("expected item in printBullet output: %q", out)
	}
}

func TestPrintRow(t *testing.T) {
	out := captureStdout(func() { printRow("Shell", "zsh · 5 aliases") })
	if !strings.Contains(out, "Shell") || !strings.Contains(out, "zsh") {
		t.Errorf("expected label and detail in printRow output: %q", out)
	}
}

// ---------------------------------------------------------------------------
// printVersionDetails
// ---------------------------------------------------------------------------

func TestPrintVersionDetails(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
			GoVersion:   "go version go1.22 linux/amd64",
			NpmGlobals:  []string{"typescript", "yarn"},
			PipGlobals:  []string{"requests"},
		},
	}
	out := captureStdout(func() { printVersionDetails(p) })
	for _, want := range []string{"Node", "Go", "npm", "pip"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in printVersionDetails output: %q", want, out)
		}
	}
}

func TestPrintLanguageVersions(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NodeVersion: "v20.0.0",
			GoVersion:   "go1.22.0",
		},
	}
	out := captureStdout(func() { printLanguageVersions(p) })
	for _, want := range []string{"Node", "v20.0.0", "Go", "go1.22.0"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in printLanguageVersions output: %q", want, out)
		}
	}
}

func TestPrintPackageManagerGlobals(t *testing.T) {
	p := &profile.Profile{
		Languages: profile.LanguageProfile{
			NpmGlobals: []string{"typescript", "tsx"},
			PipGlobals: []string{"requests"},
		},
	}
	out := captureStdout(func() { printPackageManagerGlobals(p) })
	for _, want := range []string{"npm globals", "2", "pip packages", "1"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in printPackageManagerGlobals output: %q", want, out)
		}
	}
}

func TestPrintVersionDetailsEmpty(t *testing.T) {
	p := &profile.Profile{}
	out := captureStdout(func() { printVersionDetails(p) })
	if out != "" {
		t.Errorf("expected no output for empty profile, got: %q", out)
	}
}
