package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractAliases(t *testing.T) {
	content := `# some comment
export PATH="/usr/local/bin:$PATH"
alias ll='ls -la'
alias gs='git status'
source ~/.zsh_custom
alias dc='docker compose'
`
	got := extractAliases(content)
	want := []string{"alias ll='ls -la'", "alias gs='git status'", "alias dc='docker compose'"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExtractAliasesEmpty(t *testing.T) {
	got := extractAliases("")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestExtractZshValue(t *testing.T) {
	tests := []struct {
		name, content, key, want string
	}{
		{
			"double quotes",
			`ZSH_THEME="robbyrussell"`,
			"ZSH_THEME",
			"robbyrussell",
		},
		{
			"single quotes",
			`ZSH_THEME='agnoster'`,
			"ZSH_THEME",
			"agnoster",
		},
		{
			"no quotes",
			`ZSH_THEME=powerlevel10k`,
			"ZSH_THEME",
			"powerlevel10k",
		},
		{
			"key not found",
			`SOMETHING_ELSE="value"`,
			"ZSH_THEME",
			"",
		},
		{
			"empty content",
			"",
			"ZSH_THEME",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractZshValue(tt.content, tt.key)
			if got != tt.want {
				t.Errorf("extractZshValue(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestExtractZshPlugins(t *testing.T) {
	content := `
ZSH_THEME="robbyrussell"
plugins=(git zsh-autosuggestions zsh-syntax-highlighting)
source $ZSH/oh-my-zsh.sh
`
	got := extractZshPlugins(content)
	want := []string{"git", "zsh-autosuggestions", "zsh-syntax-highlighting"}

	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExtractZshPluginsNotFound(t *testing.T) {
	got := extractZshPlugins("no plugins line here")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single", "hello", 1},
		{"multiple", "a\nb\nc", 3},
		{"with blanks", "a\n\nb\n\nc\n", 3},
		{"whitespace", "  a  \n  b  ", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.input)
			if tt.want == 0 && got != nil {
				t.Errorf("expected nil, got %v", got)
			} else if len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"empty", "", ""},
		{"single line", "hello world", "hello world"},
		{"multi line", "first\nsecond\nthird", "first"},
		{"trailing newline", "only\n", "only"},
		{"whitespace", "  padded  \nother", "padded"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := firstLine(tt.input)
			if got != tt.want {
				t.Errorf("firstLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseComposerGlobals(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		input := `{"installed":[{"name":"laravel/installer"},{"name":"phpunit/phpunit"}]}`
		got := parseComposerGlobals(input)
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0] != "laravel/installer" {
			t.Errorf("[0] = %q, want %q", got[0], "laravel/installer")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if parseComposerGlobals("") != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if parseComposerGlobals("{bad json") != nil {
			t.Error("expected nil for invalid json")
		}
	})

	t.Run("empty installed", func(t *testing.T) {
		got := parseComposerGlobals(`{"installed":[]}`)
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})
}

func TestParsePipPackages(t *testing.T) {
	t.Run("valid json", func(t *testing.T) {
		input := `[{"name":"requests"},{"name":"flask"},{"name":"numpy"}]`
		got := parsePipPackages(input)
		if len(got) != 3 {
			t.Fatalf("len = %d, want 3", len(got))
		}
		if got[0] != "requests" {
			t.Errorf("[0] = %q, want %q", got[0], "requests")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if parsePipPackages("") != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		if parsePipPackages("not json") != nil {
			t.Error("expected nil for invalid json")
		}
	})

	t.Run("empty array", func(t *testing.T) {
		if parsePipPackages("[]") != nil {
			t.Error("expected nil for empty array")
		}
	})
}

func TestParseCargoPackages(t *testing.T) {
	t.Run("typical output", func(t *testing.T) {
		input := `cargo-edit v0.12.2:
    cargo-add
    cargo-rm
ripgrep v14.1.0:
    rg`
		got := parseCargoPackages(input)
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0] != "cargo-edit" {
			t.Errorf("[0] = %q, want %q", got[0], "cargo-edit")
		}
		if got[1] != "ripgrep" {
			t.Errorf("[1] = %q, want %q", got[1], "ripgrep")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		if parseCargoPackages("") != nil {
			t.Error("expected nil for empty input")
		}
	})

	t.Run("indented lines skipped", func(t *testing.T) {
		input := `mypackage v1.0.0:
    mybinary`
		got := parseCargoPackages(input)
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
	})
}

// --- Neovim plugin tests ---

func TestParseLazyLock(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "lazy-lock.json")

	lockContent := map[string]any{
		"nvim-treesitter": map[string]string{"branch": "master", "commit": "abc123"},
		"telescope.nvim":  map[string]string{"branch": "main", "commit": "def456"},
		"plenary.nvim":    map[string]string{"branch": "master", "commit": "ghi789"},
	}
	data, _ := json.Marshal(lockContent)
	if err := os.WriteFile(lockPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	plugins := parseLazyLock(lockPath)
	if len(plugins) != 3 {
		t.Fatalf("len = %d, want 3", len(plugins))
	}

	// All should have source = "lazy"
	for _, p := range plugins {
		if p.Source != "lazy" {
			t.Errorf("plugin %q source = %q, want lazy", p.Name, p.Source)
		}
	}
}

func TestParseLazyLockMissing(t *testing.T) {
	plugins := parseLazyLock("/nonexistent/lazy-lock.json")
	if plugins != nil {
		t.Errorf("expected nil, got %v", plugins)
	}
}

func TestParseLazyLockInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "lazy-lock.json")
	if err := os.WriteFile(lockPath, []byte("{invalid json"), 0600); err != nil {
		t.Fatal(err)
	}

	plugins := parseLazyLock(lockPath)
	if plugins != nil {
		t.Errorf("expected nil for invalid JSON, got %v", plugins)
	}
}

// --- SSH key tests ---

func TestParseSSHKeygenOutput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantType    string
		wantFP      string
		wantComment string
	}{
		{
			"ed25519",
			"256 SHA256:abcdef123456 user@host (ED25519)",
			"ED25519",
			"SHA256:abcdef123456",
			"user@host",
		},
		{
			"rsa with email comment",
			"3072 SHA256:xyzabc789012 alice@example.com (RSA)",
			"RSA",
			"SHA256:xyzabc789012",
			"alice@example.com",
		},
		{
			"ecdsa no comment",
			"256 SHA256:ecdsa123 (ECDSA)",
			"ECDSA",
			"SHA256:ecdsa123",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseSSHKeygenOutput(tt.input)
			if info == nil {
				t.Fatal("expected non-nil result")
			}
			if info.keyType != tt.wantType {
				t.Errorf("type = %q, want %q", info.keyType, tt.wantType)
			}
			if info.fingerprint != tt.wantFP {
				t.Errorf("fingerprint = %q, want %q", info.fingerprint, tt.wantFP)
			}
			if info.comment != tt.wantComment {
				t.Errorf("comment = %q, want %q", info.comment, tt.wantComment)
			}
		})
	}
}

func TestParseSSHKeygenOutputInvalid(t *testing.T) {
	tests := []struct {
		name, input string
	}{
		{"empty", ""},
		{"too few fields", "256 SHA256:abc"},
		{"garbage", "not ssh-keygen output"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if parseSSHKeygenOutput(tt.input) != nil {
				t.Error("expected nil for invalid input")
			}
		})
	}
}

func TestParseSSHPubKeyFingerprintRejectsNonPub(t *testing.T) {
	// Defense in depth: must refuse to process non-.pub files
	result := parseSSHPubKeyFingerprint("/home/user/.ssh/id_ed25519")
	if result != nil {
		t.Error("expected nil for non-.pub file")
	}
}

func TestScanSSHEmptyDir(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	result := scanSSH(dir, func(string) {})
	if len(result.Keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(result.Keys))
	}
}

func TestScanSSHNoDir(t *testing.T) {
	dir := t.TempDir()
	// No .ssh directory exists
	result := scanSSH(dir, func(string) {})
	if len(result.Keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(result.Keys))
	}
}

func TestScanSSHSkipsPrivateKeys(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Create a fake private key and public key pair
	privateKey := filepath.Join(sshDir, "id_test")
	pubKey := filepath.Join(sshDir, "id_test.pub")

	// Write private key - this should NEVER be read
	if err := os.WriteFile(privateKey, []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nfake\n-----END OPENSSH PRIVATE KEY-----"), 0600); err != nil {
		t.Fatal(err)
	}
	// Write public key
	if err := os.WriteFile(pubKey, []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAItest test@host"), 0600); err != nil {
		t.Fatal(err)
	}
	// Also create a standalone private key with no .pub - should not appear
	if err := os.WriteFile(filepath.Join(sshDir, "deploy_key"), []byte("-----BEGIN RSA PRIVATE KEY-----"), 0600); err != nil {
		t.Fatal(err)
	}

	var warnings []string
	result := scanSSH(dir, func(w string) { warnings = append(warnings, w) })

	// Should find exactly 1 key (id_test, from the .pub file)
	if len(result.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result.Keys))
	}
	if result.Keys[0].Filename != "id_test" {
		t.Errorf("filename = %q, want id_test", result.Keys[0].Filename)
	}
	// PublicOnly should be false since the private key exists
	if result.Keys[0].PublicOnly {
		t.Error("PublicOnly should be false when private key exists")
	}
}

func TestScanSSHPublicOnlyKey(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}

	// Only a .pub file, no corresponding private key
	pubKey := filepath.Join(sshDir, "old_key.pub")
	if err := os.WriteFile(pubKey, []byte("ssh-rsa AAAAB3NzaC1yc2EAAAAtest old@host"), 0600); err != nil {
		t.Fatal(err)
	}

	result := scanSSH(dir, func(string) {})
	if len(result.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result.Keys))
	}
	if !result.Keys[0].PublicOnly {
		t.Error("PublicOnly should be true when private key is missing")
	}
}

func TestRunWithProgressCallsCallback(t *testing.T) {
	var sections []string
	_, _, err := RunWithProgress("test", func(s string) {
		sections = append(sections, s)
	})
	if err != nil {
		t.Fatalf("RunWithProgress: %v", err)
	}

	// All nine scanner sections must fire in order.
	want := []string{"System", "Homebrew", "Shell", "Editors", "Git", "Languages", "Configs", "SSH", "Defaults"}
	if len(sections) != len(want) {
		t.Fatalf("got %d progress callbacks, want %d: %v", len(sections), len(want), sections)
	}
	for i, s := range want {
		if sections[i] != s {
			t.Errorf("sections[%d] = %q, want %q", i, sections[i], s)
		}
	}
}

func TestRunWithProgressNilCallback(t *testing.T) {
	// nil callback must not panic.
	_, _, err := Run("test-nil")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}
