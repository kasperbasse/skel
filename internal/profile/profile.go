package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxImportSize = 50 * 1024 * 1024 // 50 MB

// profileDirOverride allows tests to redirect profile storage to a temp dir.
var profileDirOverride string

// SetProfileDirOverride redirects profile storage to the given directory.
// Pass an empty string to reset. Intended for use in tests outside this package.
func SetProfileDirOverride(dir string) { profileDirOverride = dir }

// Profile represents a full snapshot of a developer's Mac setup.
type Profile struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Machine   string    `json:"machine"`

	Homebrew    HomebrewProfile   `json:"homebrew"`
	Shell       ShellProfile      `json:"shell"`
	Editor      EditorProfile     `json:"editor"`
	Git         GitProfile        `json:"git"`
	Languages   LanguageProfile   `json:"languages"`
	System      SystemProfile     `json:"system"`
	ConfigFiles map[string]string `json:"config_files,omitempty"`
	SSH         SSHProfile        `json:"ssh,omitempty"`
	Defaults    DefaultsProfile   `json:"defaults,omitempty"`
}

// SSHProfile stores an inventory of SSH keys (fingerprints only, NEVER private key contents).
type SSHProfile struct {
	Keys []SSHKey `json:"keys,omitempty"`
}

// SSHKey records metadata about an SSH key. Only public key fingerprints are stored.
type SSHKey struct {
	Filename    string `json:"filename"`              // e.g. "id_ed25519"
	Type        string `json:"type,omitempty"`        // e.g. "ED25519", "RSA"
	Fingerprint string `json:"fingerprint,omitempty"` // SHA256 fingerprint of the PUBLIC key
	Comment     string `json:"comment,omitempty"`     // key comment (usually email)
	PublicOnly  bool   `json:"public_only,omitempty"` // true if only .pub file exists
}

type HomebrewProfile struct {
	Taps     []string `json:"taps,omitempty"`
	Formulas []string `json:"formulas"`
	Casks    []string `json:"casks"`
	MasApps  []MasApp `json:"mas_apps"`
}

type MasApp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ShellProfile struct {
	Shell          string   `json:"shell"`
	ZshrcContent   string   `json:"zshrc_content,omitempty"`
	OhMyZsh        bool     `json:"oh_my_zsh"`
	OhMyZshTheme   string   `json:"oh_my_zsh_theme,omitempty"`
	OhMyZshPlugins []string `json:"oh_my_zsh_plugins,omitempty"`
	Starship       bool     `json:"starship"`
	StarshipConfig string   `json:"starship_config,omitempty"`
	Aliases        []string `json:"aliases,omitempty"`

	// Fish
	FishConfig        string   `json:"fish_config,omitempty"`
	FishPlugins       []string `json:"fish_plugins,omitempty"`
	FishAbbreviations []string `json:"fish_abbreviations,omitempty"`

	// Bash
	BashrcContent      string   `json:"bashrc_content,omitempty"`
	BashProfileContent string   `json:"bash_profile_content,omitempty"`
	BashAliases        []string `json:"bash_aliases,omitempty"`
}

type EditorProfile struct {
	VSCode     bool     `json:"vscode"`
	VSCodeExts []string `json:"vscode_extensions,omitempty"`
	Cursor     bool     `json:"cursor"`
	CursorExts []string `json:"cursor_extensions,omitempty"`
	Neovim     bool     `json:"neovim"`

	// Neovim plugins (lazy.nvim, packer)
	NeovimPlugins []NeovimPlugin `json:"neovim_plugins,omitempty"`

	// JetBrains IDEs
	JetBrains []JetBrainsIDE `json:"jetbrains,omitempty"`
}

type NeovimPlugin struct {
	Name   string `json:"name"`             // e.g. "nvim-treesitter/nvim-treesitter" or short name
	Source string `json:"source,omitempty"` // "lazy" or "packer"
}

type JetBrainsIDE struct {
	Name    string            `json:"name"`              // e.g. "IntelliJIdea", "GoLand"
	Version string            `json:"version"`           // e.g. "2025.1"
	Plugins []string          `json:"plugins,omitempty"` // installed plugin IDs
	Configs map[string]string `json:"configs,omitempty"` // key config files
}

type GitProfile struct {
	UserName         string `json:"user_name"`
	UserEmail        string `json:"user_email"`
	DefaultBranch    string `json:"default_branch,omitempty"`
	GitConfigContent string `json:"gitconfig_content,omitempty"`
	GlobalIgnore     string `json:"global_ignore,omitempty"`
}

type LanguageProfile struct {
	NodeVersion   string `json:"node_version,omitempty"`
	PythonVersion string `json:"python_version,omitempty"`
	GoVersion     string `json:"go_version,omitempty"`
	RubyVersion   string `json:"ruby_version,omitempty"`
	PHPVersion    string `json:"php_version,omitempty"`
	RustVersion   string `json:"rust_version,omitempty"`
	JavaVersion   string `json:"java_version,omitempty"`

	NpmGlobals      []string `json:"npm_globals,omitempty"`
	YarnGlobals     []string `json:"yarn_globals,omitempty"`
	PnpmGlobals     []string `json:"pnpm_globals,omitempty"`
	ComposerGlobals []string `json:"composer_globals,omitempty"`
	PipGlobals      []string `json:"pip_globals,omitempty"`
	GemGlobals      []string `json:"gem_globals,omitempty"`
	CargoPackages   []string `json:"cargo_packages,omitempty"`
}

// DefaultsProfile stores macOS user preferences captured via `defaults read`.
type DefaultsProfile struct {
	Settings []DefaultsSetting `json:"settings,omitempty"`
}

// DefaultsSetting is a single macOS preference key-value pair.
type DefaultsSetting struct {
	Domain string `json:"domain"`
	Key    string `json:"key"`
	Type   string `json:"type"`  // "string", "int", "float", "bool"
	Value  string `json:"value"` // string representation of the value
}

type SystemProfile struct {
	Hostname     string `json:"hostname"`
	MacOSVersion string `json:"macos_version"`
	ChipArch     string `json:"chip_arch"`
}

// Redact returns a deep copy of the profile with personally identifiable fields cleared:
// git identity (name, email, raw gitconfig), hostname, machine name, and SSH key comments.
// Fields like git.default_branch, git.global_ignore, and shell/config file contents
// are kept because they are the primary value of a shared profile.
func (p *Profile) Redact() *Profile {
	out := *p
	out.Machine = "shared"
	out.System.Hostname = ""
	out.Git.UserName = ""
	out.Git.UserEmail = ""
	out.Git.GitConfigContent = ""

	// Deep-copy ConfigFiles so the redacted copy is fully independent.
	if len(p.ConfigFiles) > 0 {
		out.ConfigFiles = make(map[string]string, len(p.ConfigFiles))
		for k, v := range p.ConfigFiles {
			out.ConfigFiles[k] = v
		}
	}

	// Deep-copy SSH keys and clear comments.
	if len(p.SSH.Keys) > 0 {
		keys := make([]SSHKey, len(p.SSH.Keys))
		for i, k := range p.SSH.Keys {
			k.Comment = ""
			keys[i] = k
		}
		out.SSH = SSHProfile{Keys: keys}
	}

	return &out
}

// Validate checks that a profile does not contain dangerous paths or oversized data.
func (p *Profile) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("profile is missing a name field")
	}
	for relPath := range p.ConfigFiles {
		cleaned := filepath.Clean(relPath)
		if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
			return fmt.Errorf("config file path %q escapes home directory", relPath)
		}
	}

	// SSH key validation - defense in depth against private key leakage.
	for _, key := range p.SSH.Keys {
		if strings.Contains(key.Filename, "..") || filepath.IsAbs(key.Filename) {
			return fmt.Errorf("SSH key filename %q contains path traversal", key.Filename)
		}
		if containsPrivateKeyMarker(key.Fingerprint) || containsPrivateKeyMarker(key.Comment) {
			return fmt.Errorf("SSH key entry for %q appears to contain private key material", key.Filename)
		}
	}

	return nil
}

// containsPrivateKeyMarker checks if a string contains PEM private key markers.
func containsPrivateKeyMarker(s string) bool {
	upper := strings.ToUpper(s)
	return strings.Contains(upper, "PRIVATE KEY") || strings.Contains(upper, "BEGIN RSA") || strings.Contains(upper, "BEGIN EC") || strings.Contains(upper, "BEGIN OPENSSH")
}

// --- Storage ---

func GetProfileDir() (string, error) {
	if profileDirOverride != "" {
		return profileDirOverride, os.MkdirAll(profileDirOverride, 0700)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	dir := filepath.Join(home, ".skel", "profiles")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("creating profiles directory: %w", err)
	}
	return dir, nil
}

// Save writes the profile to disk and returns the size in bytes.
func Save(p *Profile) (int, error) {
	dir, err := GetProfileDir()
	if err != nil {
		return 0, fmt.Errorf("saving profile: %w", err)
	}

	filename := fmt.Sprintf("%s.json", sanitizeName(p.Name))
	path := filepath.Join(dir, filename)

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("encoding profile: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return 0, fmt.Errorf("writing profile to %s: %w", path, err)
	}
	return len(data), nil
}

func Load(name string) (*Profile, error) {
	dir, err := GetProfileDir()
	if err != nil {
		return nil, fmt.Errorf("loading profile: %w", err)
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.json", sanitizeName(name)))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile '%s' not found", name)
		}
		return nil, fmt.Errorf("reading profile '%s': %w", name, err)
	}

	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parsing profile '%s': %w", name, err)
	}
	return &p, nil
}

func Exists(name string) bool {
	dir, err := GetProfileDir()
	if err != nil {
		return false
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.json", sanitizeName(name)))
	_, err = os.Stat(path)
	return err == nil
}

func ListAll() ([]*Profile, error) {
	dir, err := GetProfileDir()
	if err != nil {
		return nil, fmt.Errorf("listing profiles: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading profiles directory: %w", err)
	}

	var profiles []*Profile
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5]
		p, err := Load(name)
		if err == nil {
			profiles = append(profiles, p)
		}
	}
	return profiles, nil
}

func Delete(name string) error {
	dir, err := GetProfileDir()
	if err != nil {
		return fmt.Errorf("deleting profile: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s.json", sanitizeName(name)))
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("profile '%s' not found", name)
		}
		return fmt.Errorf("removing profile '%s': %w", name, err)
	}
	return nil
}

func sanitizeName(name string) string {
	result := make([]byte, len(name))
	for i, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			result[i] = byte(c & 0x7f) //nolint:gosec // c is always ASCII here (matched above)
		} else {
			result[i] = '-'
		}
	}
	return string(result)
}
