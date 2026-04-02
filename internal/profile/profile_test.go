package profile

import (
	"testing"
	"time"
)

func setupTestDir(t *testing.T) {
	t.Helper()
	profileDirOverride = t.TempDir()
	t.Cleanup(func() { profileDirOverride = "" })
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"work-2026", "work-2026"},
		{"my profile", "my-profile"},
		{"test.v2", "test-v2"},
		{"UPPER", "UPPER"},
		{"a/b\\c", "a-b-c"},
		{"", ""},
	}
	for _, tt := range tests {
		got := sanitizeName(tt.input)
		if got != tt.want {
			t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	setupTestDir(t)

	p := &Profile{
		Name:      "test",
		CreatedAt: time.Now().Truncate(time.Second),
		Machine:   "test-machine",
		Homebrew: HomebrewProfile{
			Formulas: []string{"git", "ripgrep"},
			Casks:    []string{"iterm2"},
		},
		Shell: ShellProfile{
			Shell:   "zsh",
			Aliases: []string{"alias ll='ls -la'"},
		},
		Git: GitProfile{
			UserName:  "Test User",
			UserEmail: "test@example.com",
		},
		System: SystemProfile{
			Hostname:     "test-host",
			MacOSVersion: "14.0",
			ChipArch:     "arm64",
		},
	}

	if _, err := Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load("test")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.Name != p.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, p.Name)
	}
	if loaded.Machine != p.Machine {
		t.Errorf("Machine = %q, want %q", loaded.Machine, p.Machine)
	}
	if len(loaded.Homebrew.Formulas) != 2 {
		t.Errorf("Formulas len = %d, want 2", len(loaded.Homebrew.Formulas))
	}
	if loaded.Git.UserName != "Test User" {
		t.Errorf("UserName = %q, want %q", loaded.Git.UserName, "Test User")
	}
}

func TestSaveLoadRoundTripWithNewFields(t *testing.T) {
	setupTestDir(t)

	p := &Profile{
		Name:      "full",
		CreatedAt: time.Now().Truncate(time.Second),
		Machine:   "test-machine",
		Homebrew: HomebrewProfile{
			Taps:     []string{"homebrew/cask-fonts"},
			Formulas: []string{"git"},
			Casks:    []string{"firefox"},
			MasApps:  []MasApp{{ID: "123", Name: "Xcode"}},
		},
		Shell: ShellProfile{
			Shell:              "zsh",
			FishConfig:         "set -x PATH /usr/local/bin $PATH",
			FishPlugins:        []string{"z", "fzf"},
			FishAbbreviations:  []string{"abbr g git"},
			BashrcContent:      "export PS1='$ '",
			BashProfileContent: "source ~/.bashrc",
			BashAliases:        []string{"alias ll='ls -la'"},
		},
		Editor: EditorProfile{
			VSCode: true,
			Neovim: true,
			NeovimPlugins: []NeovimPlugin{
				{Name: "nvim-treesitter", Source: "lazy"},
				{Name: "telescope.nvim", Source: "lazy"},
			},
			JetBrains: []JetBrainsIDE{
				{Name: "GoLand", Version: "2025.1", Plugins: []string{"go-template"}},
			},
		},
		SSH: SSHProfile{
			Keys: []SSHKey{
				{Filename: "id_ed25519", Type: "ED25519", Fingerprint: "SHA256:abc123", Comment: "test@host"},
			},
		},
		Languages: LanguageProfile{
			PHPVersion:      "8.3",
			RustVersion:     "1.75.0",
			JavaVersion:     "21",
			ComposerGlobals: []string{"laravel/installer"},
			PipGlobals:      []string{"requests"},
			GemGlobals:      []string{"rails"},
			CargoPackages:   []string{"ripgrep"},
			YarnGlobals:     []string{"typescript"},
			PnpmGlobals:     []string{"turbo"},
		},
		ConfigFiles: map[string]string{
			".config/kitty/kitty.conf": "font_size 14",
		},
	}

	if _, err := Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load("full")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(loaded.Homebrew.Taps) != 1 {
		t.Errorf("Taps len = %d, want 1", len(loaded.Homebrew.Taps))
	}
	if len(loaded.Shell.FishPlugins) != 2 {
		t.Errorf("FishPlugins len = %d, want 2", len(loaded.Shell.FishPlugins))
	}
	if loaded.Shell.BashrcContent != "export PS1='$ '" {
		t.Errorf("BashrcContent = %q", loaded.Shell.BashrcContent)
	}
	if len(loaded.Editor.JetBrains) != 1 {
		t.Errorf("JetBrains len = %d, want 1", len(loaded.Editor.JetBrains))
	}
	if loaded.Editor.JetBrains[0].Name != "GoLand" {
		t.Errorf("JetBrains[0].Name = %q, want GoLand", loaded.Editor.JetBrains[0].Name)
	}
	if len(loaded.Editor.NeovimPlugins) != 2 {
		t.Errorf("NeovimPlugins len = %d, want 2", len(loaded.Editor.NeovimPlugins))
	}
	if len(loaded.SSH.Keys) != 1 {
		t.Errorf("SSH keys len = %d, want 1", len(loaded.SSH.Keys))
	}
	if loaded.SSH.Keys[0].Fingerprint != "SHA256:abc123" {
		t.Errorf("SSH fingerprint = %q, want SHA256:abc123", loaded.SSH.Keys[0].Fingerprint)
	}
	if loaded.Languages.RustVersion != "1.75.0" {
		t.Errorf("RustVersion = %q, want 1.75.0", loaded.Languages.RustVersion)
	}
	if loaded.ConfigFiles[".config/kitty/kitty.conf"] != "font_size 14" {
		t.Error("ConfigFiles content mismatch")
	}
}

func TestLoadNotFound(t *testing.T) {
	setupTestDir(t)

	_, err := Load("does-not-exist")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "profile 'does-not-exist' not found"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestExists(t *testing.T) {
	setupTestDir(t)

	if Exists("nope") {
		t.Error("Exists returned true for non-existent profile")
	}

	p := &Profile{Name: "exists-test", CreatedAt: time.Now()}
	if _, err := Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if !Exists("exists-test") {
		t.Error("Exists returned false for saved profile")
	}
}

func TestDeleteExisting(t *testing.T) {
	setupTestDir(t)

	p := &Profile{Name: "to-delete", CreatedAt: time.Now()}
	if _, err := Save(p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := Delete("to-delete"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := Load("to-delete")
	if err == nil {
		t.Fatal("expected error loading deleted profile")
	}
}

func TestDeleteNotFound(t *testing.T) {
	setupTestDir(t)

	err := Delete("ghost")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListAll(t *testing.T) {
	setupTestDir(t)

	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		if _, err := Save(&Profile{Name: n, CreatedAt: time.Now()}); err != nil {
			t.Fatalf("Save(%s): %v", n, err)
		}
	}

	profiles, err := ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}

	if len(profiles) != 3 {
		t.Errorf("ListAll returned %d profiles, want 3", len(profiles))
	}
}

func TestListAllEmpty(t *testing.T) {
	setupTestDir(t)

	profiles, err := ListAll()
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if profiles != nil {
		t.Errorf("expected nil, got %d profiles", len(profiles))
	}
}

func TestValidate(t *testing.T) {
	t.Run("valid profile", func(t *testing.T) {
		p := &Profile{
			Name:        "test",
			ConfigFiles: map[string]string{".config/kitty/kitty.conf": "content"},
		}
		if err := p.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		p := &Profile{}
		if err := p.Validate(); err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("path traversal with dotdot", func(t *testing.T) {
		p := &Profile{
			Name:        "evil",
			ConfigFiles: map[string]string{"../../etc/passwd": "root:x:0:0"},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for path traversal")
		}
	})

	t.Run("absolute path", func(t *testing.T) {
		p := &Profile{
			Name:        "evil",
			ConfigFiles: map[string]string{"/etc/passwd": "root:x:0:0"},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for absolute path")
		}
	})

	t.Run("valid nested path", func(t *testing.T) {
		p := &Profile{
			Name:        "ok",
			ConfigFiles: map[string]string{".config/alacritty/alacritty.toml": "content"},
		}
		if err := p.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("nil config files", func(t *testing.T) {
		p := &Profile{Name: "minimal"}
		if err := p.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("valid SSH keys", func(t *testing.T) {
		p := &Profile{
			Name: "ssh-test",
			SSH: SSHProfile{
				Keys: []SSHKey{
					{Filename: "id_ed25519", Type: "ED25519", Fingerprint: "SHA256:abc123"},
				},
			},
		}
		if err := p.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("SSH key path traversal", func(t *testing.T) {
		p := &Profile{
			Name: "evil",
			SSH: SSHProfile{
				Keys: []SSHKey{
					{Filename: "../../etc/passwd"},
				},
			},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for SSH path traversal")
		}
	})

	t.Run("SSH key absolute path", func(t *testing.T) {
		p := &Profile{
			Name: "evil",
			SSH: SSHProfile{
				Keys: []SSHKey{
					{Filename: "/etc/shadow"},
				},
			},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for SSH absolute path")
		}
	})

	t.Run("SSH private key material in fingerprint", func(t *testing.T) {
		p := &Profile{
			Name: "evil",
			SSH: SSHProfile{
				Keys: []SSHKey{
					{Filename: "id_rsa", Fingerprint: "-----BEGIN RSA PRIVATE KEY-----"},
				},
			},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for private key material")
		}
	})

	t.Run("SSH private key material in comment", func(t *testing.T) {
		p := &Profile{
			Name: "evil",
			SSH: SSHProfile{
				Keys: []SSHKey{
					{Filename: "id_rsa", Comment: "BEGIN OPENSSH PRIVATE KEY"},
				},
			},
		}
		if err := p.Validate(); err == nil {
			t.Error("expected error for private key material in comment")
		}
	})
}
