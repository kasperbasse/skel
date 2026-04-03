package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

// ---------------------------------------------------------------------------
// Profile sections - individual list fields for diff / drift / list / manage
// ---------------------------------------------------------------------------

// ProfileSection defines a named, comparable slice of items from a profile.
// Adding a new field here automatically updates diff, drift, list, manage, and item counts.
type ProfileSection struct {
	Icon  string
	Label string
	Items func(p *profile.Profile) []string
}

var (
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))            // Dark Gray
	versionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))             // Cyan
	countStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true) // Pink/Hot
)

// profileSections is the ordered list of all comparable list fields.
var profileSections = []ProfileSection{
	// Homebrew
	{"🔌", "Homebrew Taps", func(p *profile.Profile) []string { return p.Homebrew.Taps }},
	{"🍺", "Homebrew Formulas", func(p *profile.Profile) []string { return p.Homebrew.Formulas }},
	{"📦", "Casks", func(p *profile.Profile) []string { return p.Homebrew.Casks }},
	// Mas
	{"🛍️ ", "App Store Apps", func(p *profile.Profile) []string {
		items := make([]string, len(p.Homebrew.MasApps))
		for i, a := range p.Homebrew.MasApps {
			items[i] = fmt.Sprintf("%s (%s)", a.Name, a.ID)
		}
		return items
	}},
	// Editor
	{"💻", "VS Code Extensions", func(p *profile.Profile) []string { return p.Editor.VSCodeExts }},
	{"🖱️ ", "Cursor Extensions", func(p *profile.Profile) []string { return p.Editor.CursorExts }},
	{"📝", "Neovim Plugins", func(p *profile.Profile) []string {
		items := make([]string, len(p.Editor.NeovimPlugins))
		for i, np := range p.Editor.NeovimPlugins {
			items[i] = np.Name
		}
		return items
	}},
	{"🧠", "JetBrains Plugins", func(p *profile.Profile) []string {
		var items []string
		for _, jb := range p.Editor.JetBrains {
			for _, plugin := range jb.Plugins {
				items = append(items, jb.Name+": "+plugin)
			}
		}
		return items
	}},
	// Language versions
	{"🟢", "Node", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.NodeVersion) }},
	{"🐹", "Go", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.GoVersion) }},
	{"🐍", "Python", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.PythonVersion) }},
	{"💎", "Ruby", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.RubyVersion) }},
	{"🐘", "PHP", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.PHPVersion) }},
	{"🦀", "Rust", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.RustVersion) }},
	{"☕", "Java", func(p *profile.Profile) []string { return wrapIfNotEmpty(p.Languages.JavaVersion) }},
	// Language packages/globals
	{"📦", "npm Globals", func(p *profile.Profile) []string { return p.Languages.NpmGlobals }},
	{"📦", "Yarn Globals", func(p *profile.Profile) []string { return p.Languages.YarnGlobals }},
	{"📦", "pnpm Globals", func(p *profile.Profile) []string { return p.Languages.PnpmGlobals }},
	{"📦", "Composer Globals", func(p *profile.Profile) []string { return p.Languages.ComposerGlobals }},
	{"📦", "pip Packages", func(p *profile.Profile) []string { return p.Languages.PipGlobals }},
	{"💎", "Ruby Gems", func(p *profile.Profile) []string { return p.Languages.GemGlobals }},
	{"📦", "Cargo Packages", func(p *profile.Profile) []string { return p.Languages.CargoPackages }},
	// Shell
	{"🐟", "Fish Plugins", func(p *profile.Profile) []string { return p.Shell.FishPlugins }},
	// Config files
	{"⚙️ ", "Config Files", func(p *profile.Profile) []string {
		items := make([]string, 0, len(p.ConfigFiles))
		for path := range p.ConfigFiles {
			items = append(items, "~/"+path)
		}
		return items
	}},
	// Mac defaults
	{"🖥️ ", "macOS Defaults", func(p *profile.Profile) []string {
		items := make([]string, len(p.Defaults.Settings))
		for i, s := range p.Defaults.Settings {
			items[i] = fmt.Sprintf("%s %s = %s", s.Domain, s.Key, s.Value)
		}
		return items
	}},
	// Security
	{"🔑", "SSH Keys", func(p *profile.Profile) []string {
		items := make([]string, len(p.SSH.Keys))
		for i, k := range p.SSH.Keys {
			items[i] = fmt.Sprintf("%s (%s)", k.Filename, k.Fingerprint)
		}
		return items
	}},
}

// ---------------------------------------------------------------------------
// Version and content fields - for drift / language summaries / import warnings
// ---------------------------------------------------------------------------

// VersionField defines a named version string in a profile.
type VersionField struct {
	Label        string // drift label ("Node", "Go", ...)
	DisplayLabel string // prepended in summaries; empty = version string already includes it
	Value        func(p *profile.Profile) string
}

var versionFields = []VersionField{
	{"Node", "Node", func(p *profile.Profile) string { return p.Languages.NodeVersion }},
	{"Go", "Go", func(p *profile.Profile) string { return p.Languages.GoVersion }},
	{"Python", "Python", func(p *profile.Profile) string { return p.Languages.PythonVersion }},
	{"Ruby", "Ruby", func(p *profile.Profile) string { return p.Languages.RubyVersion }},
	{"PHP", "PHP", func(p *profile.Profile) string { return p.Languages.PHPVersion }},
	{"Rust", "Rust", func(p *profile.Profile) string { return p.Languages.RustVersion }},
	{"Java", "Java", func(p *profile.Profile) string { return p.Languages.JavaVersion }},
}

// ContentField defines a named content string that runs as the user on restore.
type ContentField struct {
	Label string
	Value func(p *profile.Profile) string
}

var shellContentFields = []ContentField{
	{".zshrc", func(p *profile.Profile) string { return p.Shell.ZshrcContent }},
	{".bashrc", func(p *profile.Profile) string { return p.Shell.BashrcContent }},
	{".bash_profile", func(p *profile.Profile) string { return p.Shell.BashProfileContent }},
	{"fish config", func(p *profile.Profile) string { return p.Shell.FishConfig }},
}

// ---------------------------------------------------------------------------
// Scan groups - high-level sections for scan / show / dry-run / import display
// ---------------------------------------------------------------------------

// ScanGroup defines a high-level section with ALL its display behavior.
// This is the single source of truth: scan output, show detail, dry-run, import warnings.
type ScanGroup struct {
	Icon           string
	Label          string
	RestoreKeys    []string                                        // --only flag values; nil = not restorable
	ScanSummary    func(p *profile.Profile) string                 // "" = skip row
	ShowDetail     func(p *profile.Profile)                        // nil = skip in show
	DryRun         func(p *profile.Profile, opts *restore.Options) // nil = no dry-run
	ImportWarnings func(p *profile.Profile) []string               // nil = no warnings
}

// scanGroups defines every high-level section, in display order.
// Adding a new section here makes it appear in scan, show, dry-run, and import automatically.
var scanGroups = []ScanGroup{
	{
		Icon: "🍺", Label: "Homebrew", RestoreKeys: []string{"homebrew", "mas"},
		ScanSummary: func(p *profile.Profile) string { return summarizeBrew(p.Homebrew) },
		ShowDetail:  showHomebrew,
		DryRun:      dryRunHomebrew,
	},
	{
		Icon: "🐚", Label: "Shell", RestoreKeys: []string{"shell"},
		ScanSummary:    func(p *profile.Profile) string { return summarizeShell(p.Shell) },
		ShowDetail:     showShell,
		DryRun:         dryRunShell,
		ImportWarnings: importWarningsShell,
	},
	{
		Icon: "💻", Label: "Editors", RestoreKeys: []string{"editors"},
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
		Icon: "🔧", Label: "Git", RestoreKeys: []string{"git"},
		ScanSummary: func(p *profile.Profile) string {
			return fmt.Sprintf("%s %s", p.Git.UserName, dim("<"+p.Git.UserEmail+">"))
		},
		ShowDetail:     showGit,
		DryRun:         dryRunGit,
		ImportWarnings: importWarningsGit,
	},
	{
		Icon: "🌐", Label: "Languages", RestoreKeys: []string{"languages"},
		ScanSummary: func(p *profile.Profile) string { return summarizeVersions(p) },
		ShowDetail:  showLanguages,
		DryRun:      dryRunLanguages,
	},
	{
		Icon: "⚙️ ", Label: "Configs", RestoreKeys: []string{"configs"},
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
		Icon: "🔑", Label: "SSH",
		ScanSummary: func(p *profile.Profile) string {
			if len(p.SSH.Keys) == 0 {
				return ""
			}
			return fmt.Sprintf("%s keys (fingerprints only)", num(len(p.SSH.Keys)))
		},
		ShowDetail: showSSH,
		DryRun:     dryRunSSH, // informational only
	},
	{
		Icon: "🖥️ ", Label: "Defaults", RestoreKeys: []string{"defaults"},
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
		Icon: "🖥️ ", Label: "System",
		ScanSummary: func(p *profile.Profile) string {
			return fmt.Sprintf("%s · macOS %s (%s)", p.System.Hostname, cyan(p.System.MacOSVersion), p.System.ChipArch)
		},
		// System is scan-only - no show/dry-run/import
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

// ---------------------------------------------------------------------------
// Helpers used by multiple registries
// ---------------------------------------------------------------------------

func wrapIfNotEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

func profileItemCount(p *profile.Profile) int {
	n := 0
	for _, s := range profileSections {
		n += len(s.Items(p))
	}
	return n
}

func profileSummaryParts(p *profile.Profile) []string {
	var parts []string
	for _, s := range profileSections {
		items := s.Items(p)
		if len(items) > 0 {
			parts = append(parts, num(len(items))+" "+s.Label)
		}
	}
	return parts
}

func printDiffSection(icon, title string, added, removed []string) {
	count := len(added) + len(removed)
	fmt.Printf("  %s %s %s\n", icon, bold(title), dim(fmt.Sprintf("(%d)", count)))
	for _, f := range added {
		fmt.Printf("     %s %s\n", green("+"), green(f))
	}
	for _, f := range removed {
		fmt.Printf("     %s %s\n", red("-"), red(f))
	}
	fmt.Println()
}

func diffSlices(a, b []string) (added, removed []string) {
	setA := toSet(a)
	setB := toSet(b)
	for item := range setB {
		if !setA[item] {
			added = append(added, item)
		}
	}
	for item := range setA {
		if !setB[item] {
			removed = append(removed, item)
		}
	}
	return
}

func toSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}

func summarizeVersions(p *profile.Profile) string {
	var parts []string
	dot := subtleStyle.Render(" · ")

	for _, v := range versionFields {
		ver := v.Value(p)
		if ver == "" {
			continue
		}

		// Strip noise like "go version" or "(cli) (built...)"
		cleanVer := ver
		if strings.Contains(ver, "go version") {
			cleanVer = strings.Split(strings.Replace(ver, "go version go", "", 1), " ")[0]
		} else if strings.Contains(ver, "PHP") ||
			strings.Contains(ver, "ruby") ||
			strings.Contains(ver, "Python") ||
			strings.Contains(ver, "Rust") {
			// Just get the version number
			fields := strings.Fields(ver)
			if len(fields) > 1 {
				cleanVer = fields[1]
			}
		} else if v.Label == "Java" {
			fields := strings.Fields(ver)
			for _, f := range fields {
				if strings.Contains(f, ".") {
					cleanVer = strings.Trim(f, "\"")
					break
				}
			}
		}

		parts = append(parts, fmt.Sprintf("%s %s", subtleStyle.Render(v.DisplayLabel), versionStyle.Render(cleanVer)))
	}

	pkgs := []struct {
		label string
		count int
	}{
		{"NPM", len(p.Languages.NpmGlobals)},
		{"Yarn", len(p.Languages.YarnGlobals)},
		{"PNPM", len(p.Languages.PnpmGlobals)},
		{"Pip", len(p.Languages.PipGlobals)},
		{"Composer", len(p.Languages.ComposerGlobals)},
		{"Ruby Gems", len(p.Languages.GemGlobals)},
		{"Cargo", len(p.Languages.CargoPackages)},
	}

	for _, pkg := range pkgs {
		if pkg.count > 0 {
			// High contrast: Bold Pink Number + Dark Gray Label
			item := fmt.Sprintf("%s %s", countStyle.Render(strconv.Itoa(pkg.count)), subtleStyle.Render(pkg.label))
			parts = append(parts, item)
		}
	}

	return strings.Join(parts, dot)
}

// ===========================================================================
// Show detail functions - one per scan group
// ===========================================================================

func showHomebrew(p *profile.Profile) {
	printSection("🍺", fmt.Sprintf("Homebrew (%s formulas, %s casks)", num(len(p.Homebrew.Formulas)), num(len(p.Homebrew.Casks))))
	printList(p.Homebrew.Formulas, 15)
	if len(p.Homebrew.Casks) > 0 {
		fmt.Println()
		printSection("📦", fmt.Sprintf("Casks (%s)", num(len(p.Homebrew.Casks))))
		printList(p.Homebrew.Casks, 15)
	}
	if len(p.Homebrew.MasApps) > 0 {
		fmt.Println()
		printSection("🛍️ ", fmt.Sprintf("App Store (%s)", num(len(p.Homebrew.MasApps))))
		for _, app := range p.Homebrew.MasApps {
			printBullet(fmt.Sprintf("%s (%s)", app.Name, app.ID))
		}
	}
	if len(p.Homebrew.Taps) > 0 {
		fmt.Println()
		printSection("🔌", fmt.Sprintf("Homebrew Taps (%s)", num(len(p.Homebrew.Taps))))
		for _, t := range p.Homebrew.Taps {
			printBullet(t)
		}
	}
}

func showShell(p *profile.Profile) {
	printSection("🐚", "Shell: "+summarizeShell(p.Shell))
	if p.Shell.OhMyZsh && len(p.Shell.OhMyZshPlugins) > 0 {
		printBullet("Plugins: " + strings.Join(p.Shell.OhMyZshPlugins, ", "))
	}
	if len(p.Shell.FishPlugins) > 0 {
		printBullet("Fish plugins: " + strings.Join(p.Shell.FishPlugins, ", "))
	}
	if len(p.Shell.Aliases) > 0 {
		printBullet(fmt.Sprintf("%s aliases", num(len(p.Shell.Aliases))))
	}
}

func showEditors(p *profile.Profile) {
	if !hasEditors(p.Editor) {
		return
	}
	first := true
	printEditorGap := func() {
		if !first {
			fmt.Println()
		}
		first = false
	}
	if p.Editor.VSCode {
		printEditorGap()
		printSection("💻", fmt.Sprintf("VS Code (%s extensions)", num(len(p.Editor.VSCodeExts))))
		printList(p.Editor.VSCodeExts, 15)
	}
	if p.Editor.Cursor {
		printEditorGap()
		printSection("🖱️ ", fmt.Sprintf("Cursor (%s extensions)", num(len(p.Editor.CursorExts))))
		printList(p.Editor.CursorExts, 15)
	}
	if p.Editor.Neovim && len(p.Editor.NeovimPlugins) > 0 {
		printEditorGap()
		printSection("📝", fmt.Sprintf("Neovim (%s plugins)", num(len(p.Editor.NeovimPlugins))))
		pluginNames := make([]string, len(p.Editor.NeovimPlugins))
		for i, np := range p.Editor.NeovimPlugins {
			pluginNames[i] = np.Name
			if np.Source != "" {
				pluginNames[i] += dim(" (" + np.Source + ")")
			}
		}
		printList(pluginNames, 15)
	}
	for _, jb := range p.Editor.JetBrains {
		printEditorGap()
		pluginInfo := ""
		if len(jb.Plugins) > 0 {
			pluginInfo = fmt.Sprintf(", %s plugins", num(len(jb.Plugins)))
		}
		configInfo := ""
		if len(jb.Configs) > 0 {
			configInfo = fmt.Sprintf(", %s config files", num(len(jb.Configs)))
		}
		printSection("🧠", fmt.Sprintf("%s %s%s%s", jb.Name, cyan(jb.Version), pluginInfo, configInfo))
		printList(jb.Plugins, 15)
	}
}

func showGit(p *profile.Profile) {
	printSection("🔧", "Git")
	printBullet(fmt.Sprintf("%s %s", p.Git.UserName, dim("<"+p.Git.UserEmail+">")))
	if p.Git.DefaultBranch != "" {
		printBullet("Default branch: " + cyan(p.Git.DefaultBranch))
	}
}

func showLanguages(p *profile.Profile) {
	if summarizeVersions(p) == "" {
		return
	}
	printSection("🌐", "Languages")
	printVersionDetails(p)
}

func showConfigs(p *profile.Profile) {
	if len(p.ConfigFiles) == 0 {
		return
	}
	printSection("⚙️ ", fmt.Sprintf("Config files (%s)", num(len(p.ConfigFiles))))
	for path := range p.ConfigFiles {
		printBullet("~/" + path)
	}
}

func showDefaults(p *profile.Profile) {
	if len(p.Defaults.Settings) == 0 {
		return
	}
	printSection("🖥️ ", fmt.Sprintf("macOS Defaults (%s)", num(len(p.Defaults.Settings))))
	for _, d := range p.Defaults.Settings {
		printBullet(fmt.Sprintf("%s %s = %s", dim(d.Domain), d.Key, cyan(d.Value)))
	}
}

func showSSH(p *profile.Profile) {
	if len(p.SSH.Keys) == 0 {
		return
	}
	printSection("🔑", fmt.Sprintf("SSH Keys (%s) - fingerprints only, no private keys stored", num(len(p.SSH.Keys))))
	for _, key := range p.SSH.Keys {
		printBullet(formatSSHKey(key))
	}
}

// ===========================================================================
// Dry-run display functions - one per restorable group
// ===========================================================================

func dryRunHomebrew(p *profile.Profile, opts *restore.Options) {
	if opts.ShouldRestore("homebrew") {
		if len(p.Homebrew.Taps) > 0 {
			printSection("🔌", fmt.Sprintf("Would add %s Homebrew taps", num(len(p.Homebrew.Taps))))
			for _, t := range p.Homebrew.Taps {
				dryRunBullet("brew tap " + t)
			}
			fmt.Println()
		}
		printSection("🍺", fmt.Sprintf("Would install %s Homebrew formulas", num(len(p.Homebrew.Formulas))))
		for _, f := range p.Homebrew.Formulas {
			dryRunBullet("brew install " + f)
		}
		if len(p.Homebrew.Casks) > 0 {
			fmt.Println()
			printSection("📦", fmt.Sprintf("Would install %s casks", num(len(p.Homebrew.Casks))))
			for _, c := range p.Homebrew.Casks {
				dryRunBullet("brew install --cask " + c)
			}
		}
	}
	if opts.ShouldRestore("mas") && len(p.Homebrew.MasApps) > 0 {
		fmt.Println()
		printSection("🛍️ ", fmt.Sprintf("Would install %s App Store apps", num(len(p.Homebrew.MasApps))))
		for _, app := range p.Homebrew.MasApps {
			dryRunBullet(fmt.Sprintf("mas install %s (%s)", app.Name, app.ID))
		}
	}
}

func dryRunShell(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("shell") {
		return
	}
	fmt.Println()
	printSection("📄", "Would restore shell configs")
	if p.Shell.ZshrcContent != "" {
		dryRunBullet("~/.zshrc")
	}
	if p.Shell.Starship {
		dryRunBullet("~/.config/starship.toml")
	}
	if p.Shell.FishConfig != "" {
		dryRunBullet("~/.config/fish/config.fish")
	}
	if p.Shell.BashrcContent != "" {
		dryRunBullet("~/.bashrc")
	}
	if p.Shell.BashProfileContent != "" {
		dryRunBullet("~/.bash_profile")
	}
}

func dryRunEditors(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("editors") {
		return
	}
	if p.Editor.VSCode {
		fmt.Println()
		printSection("💻", fmt.Sprintf("Would install %s VS Code extensions", num(len(p.Editor.VSCodeExts))))
	}
	if p.Editor.Cursor {
		fmt.Println()
		printSection("🖱️ ", fmt.Sprintf("Would install %s Cursor extensions", num(len(p.Editor.CursorExts))))
	}
}

func dryRunGit(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("git") {
		return
	}
	if p.Git.GitConfigContent != "" {
		fmt.Println()
		printSection("🔧", "Would restore git config")
		dryRunBullet("~/.gitconfig")
	}
}

func dryRunLanguages(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("languages") {
		return
	}
	type langGlobal struct {
		label string
		cmd   string
		items []string
	}
	globals := []langGlobal{
		{"npm", "npm install -g", p.Languages.NpmGlobals},
		{"yarn", "yarn global add", p.Languages.YarnGlobals},
		{"pnpm", "pnpm add -g", p.Languages.PnpmGlobals},
		{"composer", "composer global require", p.Languages.ComposerGlobals},
		{"gem", "gem install", p.Languages.GemGlobals},
		{"cargo", "cargo install", p.Languages.CargoPackages},
	}
	for _, g := range globals {
		if len(g.items) == 0 {
			continue
		}
		fmt.Println()
		printSection("📦", fmt.Sprintf("Would install %s %s globals", num(len(g.items)), g.label))
		for _, pkg := range g.items {
			dryRunBullet(g.cmd + " " + pkg)
		}
	}
}

func dryRunConfigs(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("configs") || len(p.ConfigFiles) == 0 {
		return
	}
	fmt.Println()
	printSection("⚙️ ", fmt.Sprintf("Would restore %s config files", num(len(p.ConfigFiles))))
	for path := range p.ConfigFiles {
		dryRunBullet("~/" + path)
	}
}

func dryRunDefaults(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("defaults") || len(p.Defaults.Settings) == 0 {
		return
	}
	fmt.Println()
	printSection("🖥️ ", fmt.Sprintf("Would apply %s macOS preferences", num(len(p.Defaults.Settings))))
	for _, d := range p.Defaults.Settings {
		dryRunBullet(fmt.Sprintf("defaults write %s %s -%s %s", d.Domain, d.Key, d.Type, d.Value))
	}
	fmt.Printf("     %s\n", dim("You may need to restart Dock/Finder for changes to take effect"))
}

func dryRunSSH(p *profile.Profile, _ *restore.Options) {
	if len(p.SSH.Keys) == 0 {
		return
	}
	fmt.Println()
	printSection("🔑", fmt.Sprintf("SSH Keys (%s) - manual setup required", num(len(p.SSH.Keys))))
	for _, key := range p.SSH.Keys {
		printBullet(dim(formatSSHKey(key)))
	}
	fmt.Printf("     %s\n", dim("SSH keys are never restored automatically - regenerate or transfer manually"))
}

// ===========================================================================
// Import warning functions
// ===========================================================================

func importWarningsShell(p *profile.Profile) []string {
	var warnings []string
	for _, f := range shellContentFields {
		if f.Value(p) != "" {
			warnings = append(warnings, f.Label)
		}
	}
	return warnings
}

func importWarningsGit(p *profile.Profile) []string {
	if p.Git.GitConfigContent != "" {
		return []string{".gitconfig"}
	}
	return nil
}
