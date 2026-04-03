package scanner

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/kasperbasse/skel/internal/profile"
)

// Run scans the current Mac and returns a populated Profile along with any
// non-fatal warnings (e.g. tools that were not found).
func Run(name string) (*profile.Profile, []string, error) {
	return RunWithProgress(name, nil)
}

// RunWithProgress is like Run but calls onProgress(label) before each section
// so callers can display live feedback. onProgress may be nil.
func RunWithProgress(name string, onProgress func(string)) (*profile.Profile, []string, error) {
	var warnings []string
	warn := func(msg string) { warnings = append(warnings, msg) }
	progress := func(label string) {
		if onProgress != nil {
			onProgress(label)
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		warn("could not read hostname: " + err.Error())
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, warnings, err
	}

	p := &profile.Profile{
		Name:      name,
		CreatedAt: time.Now(),
		Machine:   hostname,
	}

	progress("System")
	p.System = scanSystem()

	progress("Homebrew")
	p.Homebrew = scanHomebrew(warn)

	progress("Shell")
	p.Shell = scanShell(home, warn)

	progress("Editors")
	p.Editor = scanEditor()

	progress("Git")
	p.Git = scanGit(home, warn)

	progress("Languages")
	p.Languages = scanLanguages()

	progress("Configs")
	p.ConfigFiles = scanConfigFiles(home)

	progress("SSH")
	p.SSH = scanSSH(home, warn)

	progress("Defaults")
	p.Defaults = scanDefaults(warn)

	return p, warnings, nil
}

// --- System ---

func scanSystem() profile.SystemProfile {
	hostname, _ := os.Hostname()
	macOSVersion := runCommand("sw_vers", "-productVersion")

	return profile.SystemProfile{
		Hostname:     hostname,
		MacOSVersion: macOSVersion,
		ChipArch:     runtime.GOARCH,
	}
}

// --- Homebrew ---

func scanHomebrew(warn func(string)) profile.HomebrewProfile {
	h := profile.HomebrewProfile{}

	if !which("brew") {
		warn("Homebrew not found - skipping formula/cask scan. Install it from https://brew.sh")
		return h
	}

	// Taps
	h.Taps = splitLines(runCommand("brew", "tap"))

	// Formulas
	h.Formulas = splitLines(runCommand("brew", "list", "--formula", "-1"))

	// Casks
	h.Casks = splitLines(runCommand("brew", "list", "--cask", "-1"))

	// Mac App Store
	if which("mas") {
		masOutput := runCommand("mas", "list")
		for _, line := range splitLines(masOutput) {
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				h.MasApps = append(h.MasApps, profile.MasApp{
					ID:   strings.TrimSpace(parts[0]),
					Name: strings.TrimSpace(parts[1]),
				})
			}
		}
	} else {
		warn("mas not found - skipping App Store scan. Install it with: brew install mas")
	}

	return h
}

// --- Shell ---

func scanShell(home string, warn func(string)) profile.ShellProfile {
	s := profile.ShellProfile{}

	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		warn("$SHELL is not set - cannot detect shell")
	} else {
		parts := strings.Split(shellPath, "/")
		s.Shell = parts[len(parts)-1]
	}

	// Zsh
	s = scanZsh(home, s, warn)

	// Fish
	s = scanFish(home, s)

	// Bash
	s = scanBash(home, s)

	return s
}

func scanZsh(home string, s profile.ShellProfile, warn func(string)) profile.ShellProfile {
	zshrcPath := filepath.Join(home, ".zshrc")
	content, err := readFileBounded(zshrcPath)
	if err != nil {
		if !os.IsNotExist(err) {
			warn("could not read ~/.zshrc: " + err.Error())
		}
		return s
	}
	if content == nil {
		warn("~/.zshrc exceeds 1MB, skipping content capture")
		return s
	}

	zshrc := string(content)
	s.ZshrcContent = zshrc
	s.Aliases = extractAliases(zshrc)

	if strings.Contains(zshrc, "oh-my-zsh") {
		s.OhMyZsh = true
		s.OhMyZshTheme = extractZshValue(zshrc, "ZSH_THEME")
		s.OhMyZshPlugins = extractZshPlugins(zshrc)
	}

	if strings.Contains(zshrc, "starship init") {
		s.Starship = true
		starshipPath := filepath.Join(home, ".config", "starship.toml")
		starshipConfig, err := readFileBounded(starshipPath)
		if err != nil && !os.IsNotExist(err) {
			warn("could not read starship.toml: " + err.Error())
		} else if starshipConfig != nil {
			s.StarshipConfig = string(starshipConfig)
		}
	}

	return s
}

func scanFish(home string, s profile.ShellProfile) profile.ShellProfile {
	configPath := filepath.Join(home, ".config", "fish", "config.fish")
	content, err := readFileBounded(configPath)
	if err != nil || content == nil {
		return s
	}

	fishConfig := string(content)
	s.FishConfig = fishConfig

	// Extract fish abbreviations and aliases
	scanner := bufio.NewScanner(strings.NewReader(fishConfig))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "abbr ") || strings.HasPrefix(line, "alias ") {
			s.FishAbbreviations = append(s.FishAbbreviations, line)
		}
	}

	// Fisher plugins
	pluginsPath := filepath.Join(home, ".config", "fish", "fish_plugins")
	if pluginsContent, err := os.ReadFile(pluginsPath); err == nil {
		for _, line := range strings.Split(string(pluginsContent), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				s.FishPlugins = append(s.FishPlugins, line)
			}
		}
	}

	return s
}

func scanBash(home string, s profile.ShellProfile) profile.ShellProfile {
	bashrcPath := filepath.Join(home, ".bashrc")
	if content, err := readFileBounded(bashrcPath); err == nil && content != nil {
		s.BashrcContent = string(content)
		s.BashAliases = extractAliases(string(content))
	}

	bashProfilePath := filepath.Join(home, ".bash_profile")
	if content, err := readFileBounded(bashProfilePath); err == nil && content != nil {
		s.BashProfileContent = string(content)
		s.BashAliases = append(s.BashAliases, extractAliases(string(content))...)
	}

	return s
}

func extractAliases(content string) []string {
	var aliases []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "alias ") {
			aliases = append(aliases, line)
		}
	}
	return aliases
}

func extractZshValue(content, key string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, key+"=") {
			val := strings.TrimPrefix(line, key+"=")
			return strings.Trim(val, `"'`)
		}
	}
	return ""
}

func extractZshPlugins(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "plugins=(") {
			inner := strings.TrimPrefix(line, "plugins=(")
			inner = strings.TrimSuffix(inner, ")")
			return strings.Fields(inner)
		}
	}
	return nil
}

// --- Editor ---

func scanEditor() profile.EditorProfile {
	e := profile.EditorProfile{}

	if which("code") {
		e.VSCode = true
		e.VSCodeExts = splitLines(runCommand("code", "--list-extensions"))
	}

	if which("cursor") {
		e.Cursor = true
		e.CursorExts = splitLines(runCommand("cursor", "--list-extensions"))
	}

	e.Neovim = which("nvim")
	if e.Neovim {
		e.NeovimPlugins = scanNeovimPlugins()
	}

	// JetBrains IDEs
	e.JetBrains = scanJetBrains()

	return e
}

// scanNeovimPlugins detects plugins from lazy.nvim (lazy-lock.json) and packer (start/opt dirs).
func scanNeovimPlugins() []profile.NeovimPlugin {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var plugins []profile.NeovimPlugin
	seen := make(map[string]bool)

	// lazy.nvim: parse lazy-lock.json
	lazyPaths := []string{
		filepath.Join(home, ".config", "nvim", "lazy-lock.json"),
		filepath.Join(home, ".local", "share", "nvim", "lazy", "lazy-lock.json"),
	}
	for _, lockPath := range lazyPaths {
		parsed := parseLazyLock(lockPath)
		for _, p := range parsed {
			if !seen[p.Name] {
				seen[p.Name] = true
				plugins = append(plugins, p)
			}
		}
	}

	// packer.nvim: scan start/ and opt/ directories
	packerDirs := []string{
		filepath.Join(home, ".local", "share", "nvim", "site", "pack", "packer", "start"),
		filepath.Join(home, ".local", "share", "nvim", "site", "pack", "packer", "opt"),
	}
	for _, dir := range packerDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() && !seen[e.Name()] {
				seen[e.Name()] = true
				plugins = append(plugins, profile.NeovimPlugin{
					Name:   e.Name(),
					Source: "packer",
				})
			}
		}
	}

	return plugins
}

// parseLazyLock parses a lazy-lock.json file and returns plugin entries.
func parseLazyLock(path string) []profile.NeovimPlugin {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// lazy-lock.json is a JSON object: {"plugin-name": {"branch": "...", "commit": "..."}, ...}
	var lockfile map[string]json.RawMessage
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil
	}

	var plugins []profile.NeovimPlugin
	for name := range lockfile {
		plugins = append(plugins, profile.NeovimPlugin{
			Name:   name,
			Source: "lazy",
		})
	}
	return plugins
}

func scanJetBrains() []profile.JetBrainsIDE {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	appSupport := filepath.Join(home, "Library", "Application Support", "JetBrains")
	entries, err := os.ReadDir(appSupport)
	if err != nil {
		return nil
	}

	// Map directory prefixes to IDE names
	ideNames := map[string]string{
		"IntelliJIdea":   "IntelliJ IDEA",
		"IntelliJIdeaCE": "IntelliJ IDEA CE",
		"WebStorm":       "WebStorm",
		"GoLand":         "GoLand",
		"PyCharm":        "PyCharm",
		"PyCharmCE":      "PyCharm CE",
		"PhpStorm":       "PhpStorm",
		"CLion":          "CLion",
		"RubyMine":       "RubyMine",
		"Rider":          "Rider",
		"DataGrip":       "DataGrip",
		"RustRover":      "RustRover",
		"DataSpell":      "DataSpell",
		"Aqua":           "Aqua",
	}

	type ideEntry struct {
		name    string
		version string
		dir     string
	}
	latest := make(map[string]ideEntry)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		for prefix, displayName := range ideNames {
			if strings.HasPrefix(dirName, prefix) {
				version := strings.TrimPrefix(dirName, prefix)
				if existing, ok := latest[prefix]; !ok || version > existing.version {
					latest[prefix] = ideEntry{
						name:    displayName,
						version: version,
						dir:     filepath.Join(appSupport, dirName),
					}
				}
			}
		}
	}

	var ides []profile.JetBrainsIDE
	for _, entry := range latest {
		ide := profile.JetBrainsIDE{
			Name:    entry.name,
			Version: entry.version,
			Configs: make(map[string]string),
		}

		// Read installed plugins
		pluginsDir := filepath.Join(entry.dir, "plugins")
		if pluginEntries, err := os.ReadDir(pluginsDir); err == nil {
			for _, pe := range pluginEntries {
				if pe.IsDir() {
					ide.Plugins = append(ide.Plugins, pe.Name())
				}
			}
		}

		// Discover config files from known config directories (skip files over 1MB)
		configDirs := []string{"options", "codestyles", "keymaps"}
		for _, dir := range configDirs {
			dirPath := filepath.Join(entry.dir, dir)
			files, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}
			for _, f := range files {
				if f.IsDir() || !strings.HasSuffix(f.Name(), ".xml") {
					continue
				}
				relPath := filepath.Join(dir, f.Name())
				data, err := os.ReadFile(filepath.Join(entry.dir, relPath))
				if err == nil && len(data) < 1024*1024 {
					ide.Configs[relPath] = string(data)
				}
			}
		}

		ides = append(ides, ide)
	}

	return ides
}

// --- Git ---

func scanGit(home string, warn func(string)) profile.GitProfile {
	g := profile.GitProfile{}

	if !which("git") {
		warn("git not found - skipping git config scan")
		return g
	}

	g.UserName = runCommand("git", "config", "--global", "user.name")
	g.UserEmail = runCommand("git", "config", "--global", "user.email")
	g.DefaultBranch = runCommand("git", "config", "--global", "init.defaultBranch")

	if content, err := readFileBounded(filepath.Join(home, ".gitconfig")); err == nil && content != nil {
		g.GitConfigContent = string(content)
	} else if err != nil && !os.IsNotExist(err) {
		warn("could not read ~/.gitconfig: " + err.Error())
	}

	if content, err := readFileBounded(filepath.Join(home, ".gitignore_global")); err == nil && content != nil {
		g.GlobalIgnore = string(content)
	} else if err != nil && !os.IsNotExist(err) {
		warn("could not read ~/.gitignore_global: " + err.Error())
	}

	return g
}

// --- Languages ---

func scanLanguages() profile.LanguageProfile {
	l := profile.LanguageProfile{}

	l.NodeVersion = runCommand("node", "--version")
	l.PythonVersion = runCommand("python3", "--version")
	l.GoVersion = runCommand("go", "version")
	l.RubyVersion = runCommand("ruby", "--version")
	l.PHPVersion = firstLine(runCommand("php", "--version"))
	l.RustVersion = runCommand("rustc", "--version")
	l.JavaVersion = firstLine(runCommand("java", "--version"))

	// npm globals
	if which("npm") {
		out := runCommand("npm", "list", "-g", "--depth=0", "--parseable")
		for _, line := range splitLines(out) {
			parts := strings.Split(line, "/")
			pkg := parts[len(parts)-1]
			if pkg != "" && pkg != "lib" {
				l.NpmGlobals = append(l.NpmGlobals, pkg)
			}
		}
	}

	// yarn globals
	if which("yarn") {
		out := runCommand("yarn", "global", "list", "--depth=0")
		for _, line := range splitLines(out) {
			if strings.HasPrefix(line, "info \"") {
				// yarn outputs: info "pkg@version" has binaries:
				line = strings.TrimPrefix(line, "info \"")
				if idx := strings.Index(line, "@"); idx > 0 {
					l.YarnGlobals = append(l.YarnGlobals, line[:idx])
				}
			}
		}
	}

	// pnpm globals
	if which("pnpm") {
		out := runCommand("pnpm", "list", "-g", "--depth=0")
		for _, line := range splitLines(out) {
			// pnpm output format: "package-name version"
			if parts := strings.Fields(line); len(parts) >= 1 && !strings.HasPrefix(line, "/") {
				pkg := parts[0]
				if pkg != "" && !strings.Contains(pkg, "/") {
					l.PnpmGlobals = append(l.PnpmGlobals, pkg)
				}
			}
		}
	}

	// Composer globals (PHP)
	if which("composer") {
		out := runCommand("composer", "global", "show", "--format=json")
		l.ComposerGlobals = parseComposerGlobals(out)
	}

	// pip user packages
	if which("pip3") {
		out := runCommand("pip3", "list", "--user", "--format=json")
		l.PipGlobals = parsePipPackages(out)
	}

	// Ruby gems
	if which("gem") {
		l.GemGlobals = getOnlyUserGems()
	}

	// Cargo packages
	if which("cargo") {
		out := runCommand("cargo", "install", "--list")
		l.CargoPackages = parseCargoPackages(out)
	}

	return l
}

func parseComposerGlobals(jsonStr string) []string {
	if jsonStr == "" {
		return nil
	}
	var data struct {
		Installed []struct {
			Name string `json:"name"`
		} `json:"installed"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil
	}
	var pkgs []string
	for _, p := range data.Installed {
		pkgs = append(pkgs, p.Name)
	}
	return pkgs
}

func parsePipPackages(jsonStr string) []string {
	if jsonStr == "" {
		return nil
	}
	var packages []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &packages); err != nil {
		return nil
	}
	var pkgs []string
	for _, p := range packages {
		pkgs = append(pkgs, p.Name)
	}
	return pkgs
}

func parseCargoPackages(output string) []string {
	if output == "" {
		return nil
	}
	var pkgs []string
	for _, line := range strings.Split(output, "\n") {
		// Cargo output: "package-name v1.2.3:" for installed packages (no leading whitespace)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.Contains(line, " v") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				pkgs = append(pkgs, parts[0])
			}
		}
	}
	return pkgs
}

// --- Config files ---

// scanConfigFiles discovers developer config files from ~/.config/ dynamically.
// It scans each app directory for files with config-like extensions, skipping
// directories handled by dedicated scanners and known large app data stores.
func scanConfigFiles(home string) map[string]string {
	configDir := filepath.Join(home, ".config")
	appDirs, err := os.ReadDir(configDir)
	if err != nil {
		return nil
	}

	// Directories handled by dedicated scanners or known to be large data stores.
	skipDirs := map[string]bool{
		"fish": true, "nvim": true, // handled by dedicated scanners
		"google-chrome": true, "chromium": true, "BraveSoftware": true, "vivaldi": true, // browsers
		"Code": true, "Cursor": true, // IDE data
		"discord": true, "Slack": true, // chat apps
		"configstore": true, "gcloud": true, // cloud tooling
	}

	configExts := map[string]bool{
		".conf": true, ".toml": true, ".yml": true, ".yaml": true,
		".lua": true, ".ini": true,
	}

	const maxFileSize = 1024 * 1024 // 1MB

	configs := make(map[string]string)

	for _, app := range appDirs {
		if !app.IsDir() || skipDirs[app.Name()] {
			continue
		}

		appPath := filepath.Join(configDir, app.Name())
		files, err := os.ReadDir(appPath)
		if err != nil || len(files) > 50 {
			continue // skip directories with too many entries (likely not simple config)
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			ext := filepath.Ext(name)
			// Accept known config extensions or files named "config" (no extension)
			if !configExts[ext] && name != "config" {
				continue
			}

			info, err := f.Info()
			if err != nil || info.Size() > maxFileSize {
				continue
			}

			fullPath := filepath.Join(appPath, name)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			relPath, _ := filepath.Rel(home, fullPath)
			configs[relPath] = string(content)
		}
	}

	if len(configs) == 0 {
		return nil
	}
	return configs
}

// --- SSH ---

// scanSSH inventories SSH keys by reading ONLY public key files (.pub).
// Private key files are NEVER read - only their existence is checked to set PublicOnly.
func scanSSH(home string, warn func(string)) profile.SSHProfile {
	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return profile.SSHProfile{}
	}

	var keys []profile.SSHKey
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pub") {
			continue
		}

		pubFile := e.Name()
		baseName := strings.TrimSuffix(pubFile, ".pub")

		// Check if corresponding private key exists (without reading it).
		_, privateExists := os.Stat(filepath.Join(sshDir, baseName))

		key := profile.SSHKey{
			Filename:   baseName,
			PublicOnly: privateExists != nil, // true if private key file does NOT exist
		}

		// Use ssh-keygen to extract fingerprint from the PUBLIC key only.
		pubPath := filepath.Join(sshDir, pubFile)
		if parsed := parseSSHPubKeyFingerprint(pubPath); parsed != nil {
			key.Type = parsed.keyType
			key.Fingerprint = parsed.fingerprint
			key.Comment = parsed.comment
		} else {
			warn("could not read fingerprint for " + pubFile)
		}

		keys = append(keys, key)
	}

	return profile.SSHProfile{Keys: keys}
}

type sshKeyInfo struct {
	fingerprint string
	keyType     string
	comment     string
}

// parseSSHPubKeyFingerprint runs ssh-keygen -lf on a .pub file to extract metadata.
// It ONLY accepts .pub files as a safety check.
func parseSSHPubKeyFingerprint(pubPath string) *sshKeyInfo {
	// Defense in depth: refuse to run on anything that isn't a .pub file.
	if !strings.HasSuffix(pubPath, ".pub") {
		return nil
	}

	output := runCommand("ssh-keygen", "-lf", pubPath)
	if output == "" {
		return nil
	}

	return parseSSHKeygenOutput(output)
}

// parseSSHKeygenOutput parses the output of `ssh-keygen -lf`.
// Format: "256 SHA256:abcdef... [comment] (TYPE)"
func parseSSHKeygenOutput(output string) *sshKeyInfo {
	// Example: "256 SHA256:abc123def456 user@host (ED25519)"
	// Minimum: "256 SHA256:abc (TYPE)" = 3 fields
	parts := strings.Fields(output)
	if len(parts) < 3 {
		return nil
	}

	// Fingerprint must start with a hash algorithm prefix
	if !strings.Contains(parts[1], ":") {
		return nil
	}

	// Last field must be (TYPE)
	last := parts[len(parts)-1]
	if !strings.HasPrefix(last, "(") || !strings.HasSuffix(last, ")") {
		return nil
	}

	info := &sshKeyInfo{
		fingerprint: parts[1],
		keyType:     last[1 : len(last)-1],
	}

	// Middle fields are the comment (everything between fingerprint and type)
	if len(parts) > 3 {
		info.comment = strings.Join(parts[2:len(parts)-1], " ")
	}

	return info
}

// --- Gems ---

func getUserGemPath() string {
	// Run the full environment command
	lines := strings.Split(runCommand("gem", "env"), "\n")
	for _, line := range lines {
		if strings.Contains(line, "USER INSTALLATION DIRECTORY") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func getOnlyUserGems() []string {
	userPath := getUserGemPath()
	if userPath == "" {
		return nil
	}

	// Look in the /gems subfolder
	actualGemsDir := filepath.Join(userPath, "gems")

	files, err := os.ReadDir(actualGemsDir)
	if err != nil {
		return nil
	}

	var myGems []string
	for _, f := range files {
		if f.IsDir() {
			name := strings.Split(f.Name(), "-")[0]
			myGems = append(myGems, name)
		}
	}
	return myGems
}

// --- Helpers ---

const maxFileRead = 1024 * 1024           // 1 MB per config file
const maxCommandOutput = 10 * 1024 * 1024 // 10 MB

// readFileBounded reads a file up to maxFileRead bytes. Returns nil, nil if the file
// is too large (not an error, just skipped).
func readFileBounded(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > maxFileRead {
		return nil, nil
	}
	return os.ReadFile(path)
}

func runCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ""
	}
	if err := cmd.Start(); err != nil {
		return ""
	}
	limited := io.LimitReader(stdout, maxCommandOutput)
	out, err := io.ReadAll(limited)
	if err != nil {
		_ = cmd.Wait()
		return ""
	}
	_ = cmd.Wait()
	return strings.TrimSpace(string(out))
}

func which(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func firstLine(s string) string {
	if s == "" {
		return ""
	}
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}
