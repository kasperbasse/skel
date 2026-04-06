package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

// profileSection defines a named, comparable slice of items from a profile.
// Adding a new field here automatically updates diff, drift, list, manage, and item counts.
type profileSection struct {
	Icon  string
	Label string
	Items func(p *profile.Profile) []string
}

var (
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	versionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	countStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
)

const (
	iconTaps      = "🔌"
	iconHomebrew  = "🍺"
	iconPackage   = "📦"
	iconMas       = "🛍"
	iconEditors   = "💻"
	iconCursor    = "🖱"
	iconNeovim    = "📝"
	iconJetBrains = "🧠"
	iconGit       = "🔧"
	iconLanguages = "🌐"
	iconConfigs   = "⚙️"
	iconDefaults  = "🖥"
	iconSSH       = "🔑"
	iconShell     = "🐚"
	iconFish      = "🐟"
	iconNode      = "🟢"
	iconGo        = "🐹"
	iconPython    = "🐍"
	iconRuby      = "💎"
	iconPHP       = "🐘"
	iconRust      = "🦀"
	iconJava      = "☕"
	iconDoc       = "📄"
)

// profileSections is the ordered list of all comparable list fields.
var profileSections = []profileSection{
	{iconTaps, "Homebrew Taps", func(p *profile.Profile) []string { return p.Homebrew.Taps }},
	{iconHomebrew, "Homebrew Formulas", func(p *profile.Profile) []string { return p.Homebrew.Formulas }},
	{iconPackage, "Casks", func(p *profile.Profile) []string { return p.Homebrew.Casks }},
	{iconMas, "App Store Apps", func(p *profile.Profile) []string {
		items := make([]string, len(p.Homebrew.MasApps))
		for i, a := range p.Homebrew.MasApps {
			items[i] = fmt.Sprintf("%s (%s)", a.Name, a.ID)
		}
		return items
	}},
	{iconEditors, "VS Code Extensions", func(p *profile.Profile) []string { return p.Editor.VSCodeExts }},
	{iconCursor, "Cursor Extensions", func(p *profile.Profile) []string { return p.Editor.CursorExts }},
	{iconNeovim, "Neovim Plugins", func(p *profile.Profile) []string {
		items := make([]string, len(p.Editor.NeovimPlugins))
		for i, np := range p.Editor.NeovimPlugins {
			items[i] = np.Name
		}
		return items
	}},
	{iconJetBrains, "JetBrains Plugins", func(p *profile.Profile) []string {
		var items []string
		for _, jb := range p.Editor.JetBrains {
			for _, plugin := range jb.Plugins {
				items = append(items, jb.Name+": "+plugin)
			}
		}
		return items
	}},
	{iconNode, "Node", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.NodeVersion) }},
	{iconGo, "Go", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.GoVersion) }},
	{iconPython, "Python", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.PythonVersion) }},
	{iconRuby, "Ruby", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.RubyVersion) }},
	{iconPHP, "PHP", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.PHPVersion) }},
	{iconRust, "Rust", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.RustVersion) }},
	{iconJava, "Java", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.JavaVersion) }},
	{iconPackage, "npm Globals", func(p *profile.Profile) []string { return p.Languages.NpmGlobals }},
	{iconPackage, "Yarn Globals", func(p *profile.Profile) []string { return p.Languages.YarnGlobals }},
	{iconPackage, "pnpm Globals", func(p *profile.Profile) []string { return p.Languages.PnpmGlobals }},
	{iconPackage, "Composer Globals", func(p *profile.Profile) []string { return p.Languages.ComposerGlobals }},
	{iconPackage, "pip Packages", func(p *profile.Profile) []string { return p.Languages.PipGlobals }},
	{iconRuby, "Ruby Gems", func(p *profile.Profile) []string { return p.Languages.GemGlobals }},
	{iconPackage, "Cargo Packages", func(p *profile.Profile) []string { return p.Languages.CargoPackages }},
	{iconFish, "Fish Plugins", func(p *profile.Profile) []string { return p.Shell.FishPlugins }},
	{iconConfigs, "Config Files", func(p *profile.Profile) []string {
		items := make([]string, 0, len(p.ConfigFiles))
		for path := range p.ConfigFiles {
			items = append(items, "~/"+path)
		}
		return items
	}},
	{iconDefaults, "macOS Defaults", func(p *profile.Profile) []string {
		items := make([]string, len(p.Defaults.Settings))
		for i, s := range p.Defaults.Settings {
			items[i] = fmt.Sprintf("%s %s = %s", s.Domain, s.Key, s.Value)
		}
		return items
	}},
	{iconSSH, "SSH Keys", func(p *profile.Profile) []string {
		items := make([]string, len(p.SSH.Keys))
		for i, k := range p.SSH.Keys {
			items[i] = fmt.Sprintf("%s (%s)", k.Filename, k.Fingerprint)
		}
		return items
	}},
}

// versionField defines a named version string in a profile.
type versionField struct {
	Label        string
	DisplayLabel string
	Value        func(p *profile.Profile) string
}

var versionFields = []versionField{
	{"Node", "Node", func(p *profile.Profile) string { return p.Languages.NodeVersion }},
	{"Go", "Go", func(p *profile.Profile) string { return p.Languages.GoVersion }},
	{"Python", "Python", func(p *profile.Profile) string { return p.Languages.PythonVersion }},
	{"Ruby", "Ruby", func(p *profile.Profile) string { return p.Languages.RubyVersion }},
	{"PHP", "PHP", func(p *profile.Profile) string { return p.Languages.PHPVersion }},
	{"Rust", "Rust", func(p *profile.Profile) string { return p.Languages.RustVersion }},
	{"Java", "Java", func(p *profile.Profile) string { return p.Languages.JavaVersion }},
}

// contentField defines a named content string that runs as the user on restore.
type contentField struct {
	Label string
	Value func(p *profile.Profile) string
}

var shellContentFields = []contentField{
	{".zshrc", func(p *profile.Profile) string { return p.Shell.ZshrcContent }},
	{".bashrc", func(p *profile.Profile) string { return p.Shell.BashrcContent }},
	{".bash_profile", func(p *profile.Profile) string { return p.Shell.BashProfileContent }},
	{"fish config", func(p *profile.Profile) string { return p.Shell.FishConfig }},
}

// scanGroup defines a high-level section with all its display behavior.
type scanGroup struct {
	Icon           string
	Label          string
	RestoreKeys    []string
	ScanSummary    func(p *profile.Profile) string
	ShowDetail     func(p *profile.Profile)
	DryRun         func(p *profile.Profile, opts *restore.Options)
	ImportWarnings func(p *profile.Profile) []string
}

// scanGroups defines every high-level section, in display order.
var scanGroups = []scanGroup{
	{
		Icon: iconHomebrew, Label: "Homebrew", RestoreKeys: []string{"homebrew", "mas"},
		ScanSummary: func(p *profile.Profile) string { return summarizeBrew(p.Homebrew) },
		ShowDetail:  showHomebrew,
		DryRun:      dryRunHomebrew,
	},
	{
		Icon: iconShell, Label: "Shell", RestoreKeys: []string{"shell"},
		ScanSummary: func(p *profile.Profile) string {
			return summarizeShell(p.Shell)
		},
		ShowDetail:     showShell,
		DryRun:         dryRunShell,
		ImportWarnings: importWarningsShell,
	},
	{
		Icon: iconEditors, Label: "Editors", RestoreKeys: []string{"editors"},
		ScanSummary: func(p *profile.Profile) string {
			if !hasEditors(p.Editor) {
				return ""
			}
			return summarizeEditors(p.Editor)
		},
		ShowDetail: showEditors,
		DryRun:     dryRunEditors,
	},
	{
		Icon: iconGit, Label: "Git", RestoreKeys: []string{"git"},
		ScanSummary:    func(p *profile.Profile) string { return summarizeGit(p.Git) },
		ShowDetail:     showGit,
		DryRun:         dryRunGit,
		ImportWarnings: importWarningsGit,
	},
	{
		Icon: iconLanguages, Label: "Languages", RestoreKeys: []string{"languages"},
		ScanSummary: func(p *profile.Profile) string { return summarizeVersions(p) },
		ShowDetail:  showLanguages,
		DryRun:      dryRunLanguages,
	},
	{
		Icon: iconConfigs, Label: "Configs", RestoreKeys: []string{"configs"},
		ScanSummary: func(p *profile.Profile) string {
			if len(p.ConfigFiles) == 0 {
				return ""
			}
			return fmt.Sprintf("%s config files", num(len(p.ConfigFiles)))
		},
		ShowDetail: showConfigs,
		DryRun:     dryRunConfigs,
	},
	{
		Icon: iconSSH, Label: "SSH",
		ScanSummary: func(p *profile.Profile) string {
			if len(p.SSH.Keys) == 0 {
				return ""
			}
			return fmt.Sprintf("%s keys (fingerprints only)", num(len(p.SSH.Keys)))
		},
		ShowDetail: showSSH,
		DryRun:     dryRunSSH,
	},
	{
		Icon: iconDefaults, Label: "Defaults", RestoreKeys: []string{"defaults"},
		ScanSummary: func(p *profile.Profile) string {
			if len(p.Defaults.Settings) == 0 {
				return ""
			}
			return fmt.Sprintf("%s macOS preferences", num(len(p.Defaults.Settings)))
		},
		ShowDetail: showDefaults,
		DryRun:     dryRunDefaults,
	},
	{
		Icon: iconDefaults, Label: "System",
		ScanSummary: func(p *profile.Profile) string {
			if p.System.Hostname == "" && p.System.MacOSVersion == "" && p.System.ChipArch == "" {
				return ""
			}
			return fmt.Sprintf("%s · macOS %s (%s)", p.System.Hostname, cyan(p.System.MacOSVersion), p.System.ChipArch)
		},
	},
}

// allRestoreKeys derives the valid --only flag values from scanGroups.
func allRestoreKeys() []string {
	seen := make(map[string]bool)
	var keys []string
	for _, g := range scanGroups {
		for _, k := range g.RestoreKeys {
			if !seen[k] {
				seen[k] = true
				keys = append(keys, k)
			}
		}
	}
	return keys
}
