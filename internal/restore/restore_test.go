package restore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kasperbasse/skel/internal/profile"
)

func TestHome(t *testing.T) {
	h := home()
	if h == "" {
		t.Error("home() returned empty string")
	}
}

func TestShouldRestore(t *testing.T) {
	t.Run("nil options restores all", func(t *testing.T) {
		opts := &Options{}
		if !opts.ShouldRestore("homebrew") {
			t.Error("expected true for empty sections")
		}
	})

	t.Run("specific sections", func(t *testing.T) {
		opts := &Options{Sections: map[string]bool{"homebrew": true, "shell": true}}
		if !opts.ShouldRestore("homebrew") {
			t.Error("expected true for homebrew")
		}
		if !opts.ShouldRestore("shell") {
			t.Error("expected true for shell")
		}
		if opts.ShouldRestore("git") {
			t.Error("expected false for git")
		}
	})

	t.Run("nil receiver", func(t *testing.T) {
		var opts *Options
		if !opts.ShouldRestore("anything") {
			t.Error("expected true for nil receiver")
		}
	})
}

func TestToSet(t *testing.T) {
	s := toSet([]string{"a", "b", "c"})
	if len(s) != 3 {
		t.Errorf("len = %d, want 3", len(s))
	}
	if !s["a"] || !s["b"] || !s["c"] {
		t.Error("missing expected keys")
	}

	empty := toSet(nil)
	if len(empty) != 0 {
		t.Errorf("expected empty set, got %d", len(empty))
	}
}

func TestSplitOutput(t *testing.T) {
	got := splitOutput("a\n\nb\n  c  \n")
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
	if got[2] != "c" {
		t.Errorf("[2] = %q, want %q", got[2], "c")
	}

	if splitOutput("") != nil {
		t.Error("expected nil for empty input")
	}
}

func TestCommandExists(t *testing.T) {
	if !commandExists("ls") {
		t.Error("expected ls to exist")
	}

	if commandExists("definitely-not-a-real-command-xyz") {
		t.Error("expected nonexistent command to return false")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := writeFile(path, "hello world"); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("content = %q, want %q", string(data), "hello world")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permissions = %o, want 0600", perm)
	}
}

func TestWriteFileInvalidPath(t *testing.T) {
	err := writeFile("/nonexistent/dir/file.txt", "content")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestCountSteps(t *testing.T) {
	p := &profile.Profile{
		Homebrew: profile.HomebrewProfile{
			Taps:     []string{"tap1"},
			Formulas: []string{"git", "ripgrep"},
			Casks:    []string{"firefox"},
			MasApps:  []profile.MasApp{{ID: "1", Name: "App"}},
		},
		Shell: profile.ShellProfile{
			ZshrcContent:  "content",
			BashrcContent: "content",
		},
		Git: profile.GitProfile{
			GitConfigContent: "content",
		},
		Editor: profile.EditorProfile{
			VSCode:     true,
			VSCodeExts: []string{"ext1", "ext2"},
		},
		ConfigFiles: map[string]string{"a": "b"},
		Languages: profile.LanguageProfile{
			NpmGlobals: []string{"pkg1"},
		},
	}

	// All sections enabled
	opts := &Options{}
	n := countSteps(p, opts)
	// taps(1) + formulas(2) + casks(1) + mas(1) + zshrc(1) + bashrc(1) + gitconfig(1) + vscode(2) + configs(1) + npm(1) = 12
	if n != 12 {
		t.Errorf("countSteps = %d, want 12", n)
	}

	// Only homebrew
	opts = &Options{Sections: map[string]bool{"homebrew": true}}
	n = countSteps(p, opts)
	// taps(1) + formulas(2) + casks(1) = 4
	if n != 4 {
		t.Errorf("countSteps(homebrew only) = %d, want 4", n)
	}

	// Only shell
	opts = &Options{Sections: map[string]bool{"shell": true}}
	n = countSteps(p, opts)
	// zshrc(1) + bashrc(1) = 2
	if n != 2 {
		t.Errorf("countSteps(shell only) = %d, want 2", n)
	}
}
