package restore

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

const (
	// MsgAlreadyInstalled is the step message emitted when a tool is already present.
	MsgAlreadyInstalled     = "already installed"
	msgPathTraversalBlocked = "path traversal blocked"
	dirPermissions          = os.FileMode(0700)
)

// Result reports the outcome of a single restore step.
type Result struct {
	Step    string
	Success bool
	Message string
	Index   int
	Total   int
}

// InstalledState represents what's already on the machine.
type InstalledState struct {
	Formulas   map[string]bool
	Casks      map[string]bool
	VSCodeExts map[string]bool
	CursorExts map[string]bool
	MasApps    map[string]bool // keyed by app ID
}

// Options configures which sections to restore.
type Options struct {
	Sections map[string]bool // nil = all sections
}

func (o *Options) ShouldRestore(section string) bool {
	if o == nil || len(o.Sections) == 0 {
		return true
	}
	return o.Sections[section]
}

// detectInstalled scans the current machine for what's already present.
func detectInstalled() InstalledState {
	s := InstalledState{
		Formulas:   toSet(splitOutput(runOutput("brew", "list", "--formula", "-1"))),
		Casks:      toSet(splitOutput(runOutput("brew", "list", "--cask", "-1"))),
		VSCodeExts: toSet(splitOutput(runOutput("code", "--list-extensions"))),
		CursorExts: toSet(splitOutput(runOutput("cursor", "--list-extensions"))),
		MasApps:    make(map[string]bool),
	}

	for _, line := range splitOutput(runOutput("mas", "list")) {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) >= 1 {
			s.MasApps[strings.TrimSpace(parts[0])] = true
		}
	}

	return s
}

// Run applies a profile's configuration to the current machine, step by step.
// Calls onStep for each operation (for progress reporting).
func Run(p *profile.Profile, opts *Options, onStep func(Result)) {
	homeDir := home()
	if homeDir == "" {
		onStep(Result{Step: "Setup", Success: false, Message: "could not determine home directory", Index: 1, Total: 1})
		return
	}

	installed := detectInstalled()
	total := countSteps(p, opts)
	stepIdx := 0

	emitStep := func(name string, success bool, message string) {
		stepIdx++
		onStep(Result{Step: name, Success: success, Message: message, Index: stepIdx, Total: total})
	}

	restoreSection := func(name string, fn func()) {
		if opts.ShouldRestore(name) {
			fn()
		}
	}

	restoreSection("homebrew", func() { restoreHomebrew(installed, p, emitStep) })
	restoreSection("mas", func() { restoreMacAppStore(p, emitStep) })
	restoreSection("shell", func() { restoreShellConfigs(p, homeDir, emitStep) })
	restoreSection("git", func() { restoreGitConfigs(p, homeDir, emitStep) })
	restoreSection("editors", func() { restoreEditorConfigs(p, homeDir, installed, emitStep) })
	restoreSection("configs", func() { restoreConfigFiles(p, homeDir, emitStep) })
	restoreSection("defaults", func() { restoreDefaults(p, emitStep) })
	restoreSection("languages", func() { restoreLanguageTools(p, emitStep) })
}

func restoreHomebrew(installed InstalledState, p *profile.Profile, emit func(string, bool, string)) {
	if !commandExists("brew") {
		emit("Homebrew", false, "brew not found - install Homebrew first (https://brew.sh)")
		return
	}

	for _, tap := range p.Homebrew.Taps {
		emitResult(fmt.Sprintf("brew tap %s", tap), runSilent("brew", "tap", tap), emit)
	}

	for _, formula := range p.Homebrew.Formulas {
		if installed.Formulas[formula] {
			emit(fmt.Sprintf("brew install %s", formula), true, MsgAlreadyInstalled)
			continue
		}
		emitResult(fmt.Sprintf("brew install %s", formula), brewInstall(formula, false), emit)
	}

	for _, cask := range p.Homebrew.Casks {
		if installed.Casks[cask] {
			emit(fmt.Sprintf("brew install --cask %s", cask), true, MsgAlreadyInstalled)
			continue
		}
		emitResult(fmt.Sprintf("brew install --cask %s", cask), brewInstall(cask, true), emit)
	}
}

func restoreMacAppStore(p *profile.Profile, emit func(string, bool, string)) {
	if !commandExists("mas") {
		if len(p.Homebrew.MasApps) > 0 {
			emit("Mac App Store", false, "mas not found - install with: brew install mas")
		}
		return
	}

	installed := toSet(splitOutput(runOutput("mas", "list")))
	for _, app := range p.Homebrew.MasApps {
		if installed[app.ID] {
			emit(fmt.Sprintf("mas install %s (%s)", app.Name, app.ID), true, MsgAlreadyInstalled)
			continue
		}
		emitResult(fmt.Sprintf("mas install %s (%s)", app.Name, app.ID), runSilent("mas", "install", app.ID), emit)
	}
}

func restoreShellConfigs(p *profile.Profile, homeDir string, emit func(string, bool, string)) {
	if p.Shell.ZshrcContent != "" {
		emitResult("Restore ~/.zshrc", writeFile(homeDir+"/.zshrc", p.Shell.ZshrcContent), emit)
	}

	if p.Shell.Starship && p.Shell.StarshipConfig != "" {
		configDir := homeDir + "/.config"
		ensureDirThenWrite(configDir, "Create ~/.config", configDir+"/starship.toml", "Restore starship.toml", p.Shell.StarshipConfig, emit)
	}

	if p.Shell.FishConfig != "" {
		fishDir := filepath.Join(homeDir, ".config", "fish")
		ensureDirThenWrite(fishDir, "Create ~/.config/fish", fishDir+"/config.fish", "Restore fish config", p.Shell.FishConfig, emit)
	}

	if p.Shell.BashrcContent != "" {
		emitResult("Restore ~/.bashrc", writeFile(homeDir+"/.bashrc", p.Shell.BashrcContent), emit)
	}
	if p.Shell.BashProfileContent != "" {
		emitResult("Restore ~/.bash_profile", writeFile(homeDir+"/.bash_profile", p.Shell.BashProfileContent), emit)
	}
}

func restoreGitConfigs(p *profile.Profile, homeDir string, emit func(string, bool, string)) {
	if p.Git.GitConfigContent != "" {
		emitResult("Restore ~/.gitconfig", writeFile(homeDir+"/.gitconfig", p.Git.GitConfigContent), emit)
	}
	if p.Git.GlobalIgnore != "" {
		emitResult("Restore ~/.gitignore_global", writeFile(homeDir+"/.gitignore_global", p.Git.GlobalIgnore), emit)
	}
}

func restoreEditorConfigs(p *profile.Profile, homeDir string, installed InstalledState, emit func(string, bool, string)) {
	if p.Editor.VSCode && commandExists("code") {
		for _, ext := range p.Editor.VSCodeExts {
			if installed.VSCodeExts[ext] {
				emit(fmt.Sprintf("VS Code: %s", ext), true, MsgAlreadyInstalled)
				continue
			}
			emitResult(fmt.Sprintf("VS Code: %s", ext), runSilent("code", "--install-extension", ext), emit)
		}
	}

	if p.Editor.Cursor && commandExists("cursor") {
		for _, ext := range p.Editor.CursorExts {
			if installed.CursorExts[ext] {
				emit(fmt.Sprintf("Cursor: %s", ext), true, MsgAlreadyInstalled)
				continue
			}
			emitResult(fmt.Sprintf("Cursor: %s", ext), runSilent("cursor", "--install-extension", ext), emit)
		}
	}

	restoreJetBrainsConfigs(p.Editor.JetBrains, homeDir, emit)
}

// restoreJetBrainsConfigs restores config files for each JetBrains IDE.
func restoreJetBrainsConfigs(jetBrains []profile.JetBrainsIDE, homeDir string, emit func(string, bool, string)) {
	for _, jb := range jetBrains {
		restoreSingleJetBrainsApp(jb, homeDir, emit)
	}
}

// restoreSingleJetBrainsApp locates the IDE's config directory and writes each config file.
func restoreSingleJetBrainsApp(jb profile.JetBrainsIDE, homeDir string, emit func(string, bool, string)) {
	appSupport := filepath.Join(homeDir, "Library", "Application Support", "JetBrains")
	targetDir, err := findJetBrainsAppDir(appSupport, jb.Name)
	if err != nil {
		emit(fmt.Sprintf("%s: setup", jb.Name), false, "JetBrains config dir not found")
		return
	}
	if targetDir == "" {
		emit(fmt.Sprintf("%s: setup", jb.Name), false, jb.Name+" not installed")
		return
	}

	for cfgPath, content := range jb.Configs {
		label := fmt.Sprintf("%s: %s", jb.Name, cfgPath)
		fullPath := filepath.Clean(filepath.Join(targetDir, cfgPath))
		if !strings.HasPrefix(fullPath, targetDir) {
			emit(label, false, msgPathTraversalBlocked)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fullPath), dirPermissions); err != nil {
			emit(label, false, err.Error())
			continue
		}
		emitResult(label, writeFile(fullPath, content), emit)
	}
}

// findJetBrainsAppDir finds the versioned config directory for a given JetBrains IDE name.
func findJetBrainsAppDir(appSupport, ideName string) (string, error) {
	entries, err := os.ReadDir(appSupport)
	if err != nil {
		return "", fmt.Errorf("JetBrains config dir not found")
	}
	nameKey := strings.ReplaceAll(ideName, " ", "")
	for _, e := range entries {
		if e.IsDir() && strings.Contains(e.Name(), nameKey) {
			return filepath.Join(appSupport, e.Name()), nil
		}
	}
	return "", nil
}

func restoreConfigFiles(p *profile.Profile, homeDir string, emit func(string, bool, string)) {
	for relPath, content := range p.ConfigFiles {
		fullPath := filepath.Clean(filepath.Join(homeDir, relPath))
		if !strings.HasPrefix(fullPath, homeDir) {
			emit(fmt.Sprintf("Restore %s", relPath), false, msgPathTraversalBlocked)
			continue
		}
		ensureDirThenWrite(
			filepath.Dir(fullPath), fmt.Sprintf("Create dir for %s", relPath),
			fullPath, fmt.Sprintf("Restore %s", relPath),
			content, emit,
		)
	}
}

func restoreDefaults(p *profile.Profile, emit func(string, bool, string)) {
	for _, d := range p.Defaults.Settings {
		label := fmt.Sprintf("defaults write %s %s", d.Domain, d.Key)
		if current := runOutput("defaults", "read", d.Domain, d.Key); current == d.Value {
			emit(label, true, MsgAlreadyInstalled)
			continue
		}
		typeFlag := "-" + d.Type
		emitResult(label, runSilent("defaults", "write", d.Domain, d.Key, typeFlag, d.Value), emit)
	}
}

func restoreLanguageTools(p *profile.Profile, emit func(string, bool, string)) {
	if commandExists("npm") {
		for _, pkg := range p.Languages.NpmGlobals {
			emitResult(fmt.Sprintf("npm install -g %s", pkg), runSilent("npm", "install", "-g", pkg), emit)
		}
	}

	if commandExists("yarn") {
		for _, pkg := range p.Languages.YarnGlobals {
			emitResult(fmt.Sprintf("yarn global add %s", pkg), runSilent("yarn", "global", "add", pkg), emit)
		}
	}

	if commandExists("pnpm") {
		for _, pkg := range p.Languages.PnpmGlobals {
			emitResult(fmt.Sprintf("pnpm add -g %s", pkg), runSilent("pnpm", "add", "-g", pkg), emit)
		}
	}

	if commandExists("composer") {
		for _, pkg := range p.Languages.ComposerGlobals {
			emitResult(fmt.Sprintf("composer global require %s", pkg), runSilent("composer", "global", "require", pkg), emit)
		}
	}

	if commandExists("gem") {
		restoreGems(p, emit)
	}

	if commandExists("cargo") {
		for _, pkg := range p.Languages.CargoPackages {
			emitResult(fmt.Sprintf("cargo install %s", pkg), runSilent("cargo", "install", pkg), emit)
		}
	}
}

func restoreGems(p *profile.Profile, emit func(string, bool, string)) {
	rubyPath, err := exec.Command("which", "ruby").Output()
	isSystemRuby := err == nil && strings.HasPrefix(strings.TrimSpace(string(rubyPath)), "/usr/bin/ruby")

	if isSystemRuby {
		msg := "System ruby detected (/usr/bin/ruby). Native extensions will likely fail."
		if commandExists("brew") {
			msg += " Fix with: 'brew install ruby'"
		} else {
			msg += " Please install a non-system Ruby (e.g., via rbenv or brew)."
		}
		emit("Ruby Environment Check", false, msg)
	} else {
		for _, pkg := range p.Languages.GemGlobals {
			emitResult(fmt.Sprintf("gem install %s", pkg), runSilent("gem", "install", pkg), emit)
		}
	}
}

func countSteps(p *profile.Profile, opts *Options) int {
	n := 0
	if opts.ShouldRestore("homebrew") {
		n += len(p.Homebrew.Taps) + len(p.Homebrew.Formulas) + len(p.Homebrew.Casks)
	}
	if opts.ShouldRestore("mas") {
		n += len(p.Homebrew.MasApps)
	}
	if opts.ShouldRestore("shell") {
		if p.Shell.ZshrcContent != "" {
			n++
		}
		if p.Shell.Starship && p.Shell.StarshipConfig != "" {
			n++
		}
		if p.Shell.FishConfig != "" {
			n++
		}
		if p.Shell.BashrcContent != "" {
			n++
		}
		if p.Shell.BashProfileContent != "" {
			n++
		}
	}
	if opts.ShouldRestore("git") {
		if p.Git.GitConfigContent != "" {
			n++
		}
		if p.Git.GlobalIgnore != "" {
			n++
		}
	}
	if opts.ShouldRestore("editors") {
		if p.Editor.VSCode {
			n += len(p.Editor.VSCodeExts)
		}
		if p.Editor.Cursor {
			n += len(p.Editor.CursorExts)
		}
		for _, jb := range p.Editor.JetBrains {
			n += len(jb.Configs)
		}
	}
	if opts.ShouldRestore("configs") {
		n += len(p.ConfigFiles)
	}
	if opts.ShouldRestore("defaults") {
		n += len(p.Defaults.Settings)
	}
	if opts.ShouldRestore("languages") {
		n += len(p.Languages.NpmGlobals) + len(p.Languages.YarnGlobals) +
			len(p.Languages.PnpmGlobals) + len(p.Languages.ComposerGlobals) +
			len(p.Languages.GemGlobals) + len(p.Languages.CargoPackages)
	}
	return n
}

// --- Helpers ---

// emitResult emits a step result based on error (success or failure).
func emitResult(name string, err error, emit func(string, bool, string)) {
	if err != nil {
		emit(name, false, err.Error())
	} else {
		emit(name, true, "done")
	}
}

// ensureDirThenWrite creates dirPath if needed, then writes content to filePath.
// Emits dirLabel on MkdirAll failure, fileLabel on write success/failure.
func ensureDirThenWrite(dirPath, dirLabel, filePath, fileLabel, content string, emit func(string, bool, string)) {
	if err := os.MkdirAll(dirPath, dirPermissions); err != nil {
		emit(dirLabel, false, err.Error())
		return
	}
	emitResult(fileLabel, writeFile(filePath, content), emit)
}

func brewInstall(pkg string, cask bool) error {
	args := []string{"install"}
	if cask {
		args = append(args, "--cask")
	}
	args = append(args, pkg)
	return runSilent("brew", args...)
}

func runSilent(name string, args ...string) error {
	var stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w\n    %s", err, msg)
		}
		return err
	}
	return nil
}

func runOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

func home() string {
	h, err := os.UserHomeDir()
	if err == nil {
		return h
	}
	if fallback := os.Getenv("HOME"); fallback != "" {
		return fallback
	}
	return ""
}

func splitOutput(s string) []string {
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

func toSet(s []string) map[string]bool {
	m := make(map[string]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}
