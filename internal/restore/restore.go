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

// Run applies a profile to the current machine, reporting results step by step.
func Run(p *profile.Profile, opts *Options, onStep func(Result)) {
	if home() == "" {
		onStep(Result{Step: "Setup", Success: false, Message: "could not determine home directory", Index: 1, Total: 1})
		return
	}

	installed := detectInstalled()

	// Count total steps for progress tracking.
	total := countSteps(p, opts)
	idx := 0

	emit := func(name string, success bool, msg string) {
		idx++
		onStep(Result{Step: name, Success: success, Message: msg, Index: idx, Total: total})
	}
	emitResult := func(name string, err error) {
		if err != nil {
			emit(name, false, err.Error())
		} else {
			emit(name, true, "done")
		}
	}

	if opts.ShouldRestore("homebrew") {
		if !commandExists("brew") {
			emit("Homebrew", false, "brew not found - install Homebrew first (https://brew.sh)")
		} else {
			// Taps first - formulas may depend on custom taps.
			for _, tap := range p.Homebrew.Taps {
				emitResult(fmt.Sprintf("brew tap %s", tap), runSilent("brew", "tap", tap))
			}

			// Formulas - skip already installed.
			for _, formula := range p.Homebrew.Formulas {
				if installed.Formulas[formula] {
					emit(fmt.Sprintf("brew install %s", formula), true, "already installed")
					continue
				}
				emitResult(fmt.Sprintf("brew install %s", formula), brewInstall(formula, false))
			}

			// Casks - skip already installed.
			for _, cask := range p.Homebrew.Casks {
				if installed.Casks[cask] {
					emit(fmt.Sprintf("brew install --cask %s", cask), true, "already installed")
					continue
				}
				emitResult(fmt.Sprintf("brew install --cask %s", cask), brewInstall(cask, true))
			}
		}
	}

	if opts.ShouldRestore("mas") {
		if !commandExists("mas") {
			if len(p.Homebrew.MasApps) > 0 {
				emit("Mac App Store", false, "mas not found - install with: brew install mas")
			}
		} else {
			for _, app := range p.Homebrew.MasApps {
				if installed.MasApps[app.ID] {
					emit(fmt.Sprintf("mas install %s (%s)", app.Name, app.ID), true, "already installed")
					continue
				}
				emitResult(fmt.Sprintf("mas install %s (%s)", app.Name, app.ID),
					runSilent("mas", "install", app.ID))
			}
		}
	}

	if opts.ShouldRestore("shell") {
		if p.Shell.ZshrcContent != "" {
			emitResult("Restore ~/.zshrc", writeFile(home()+"/.zshrc", p.Shell.ZshrcContent))
		}

		if p.Shell.Starship && p.Shell.StarshipConfig != "" {
			configDir := home() + "/.config"
			if err := os.MkdirAll(configDir, 0700); err != nil {
				emit("Create ~/.config", false, err.Error())
			} else {
				emitResult("Restore starship.toml", writeFile(configDir+"/starship.toml", p.Shell.StarshipConfig))
			}
		}

		if p.Shell.FishConfig != "" {
			fishDir := filepath.Join(home(), ".config", "fish")
			if err := os.MkdirAll(fishDir, 0700); err != nil {
				emit("Create ~/.config/fish", false, err.Error())
			} else {
				emitResult("Restore fish config", writeFile(fishDir+"/config.fish", p.Shell.FishConfig))
			}
		}

		if p.Shell.BashrcContent != "" {
			emitResult("Restore ~/.bashrc", writeFile(home()+"/.bashrc", p.Shell.BashrcContent))
		}
		if p.Shell.BashProfileContent != "" {
			emitResult("Restore ~/.bash_profile", writeFile(home()+"/.bash_profile", p.Shell.BashProfileContent))
		}
	}

	if opts.ShouldRestore("git") {
		if p.Git.GitConfigContent != "" {
			emitResult("Restore ~/.gitconfig", writeFile(home()+"/.gitconfig", p.Git.GitConfigContent))
		}
		if p.Git.GlobalIgnore != "" {
			emitResult("Restore ~/.gitignore_global", writeFile(home()+"/.gitignore_global", p.Git.GlobalIgnore))
		}
	}

	if opts.ShouldRestore("editors") {
		if p.Editor.VSCode && commandExists("code") {
			for _, ext := range p.Editor.VSCodeExts {
				if installed.VSCodeExts[ext] {
					emit(fmt.Sprintf("VS Code: %s", ext), true, "already installed")
					continue
				}
				emitResult(fmt.Sprintf("VS Code: %s", ext), runSilent("code", "--install-extension", ext))
			}
		}

		if p.Editor.Cursor && commandExists("cursor") {
			for _, ext := range p.Editor.CursorExts {
				if installed.CursorExts[ext] {
					emit(fmt.Sprintf("Cursor: %s", ext), true, "already installed")
					continue
				}
				emitResult(fmt.Sprintf("Cursor: %s", ext), runSilent("cursor", "--install-extension", ext))
			}
		}

		// JetBrains IDE configs
		for _, jb := range p.Editor.JetBrains {
			for cfgPath, content := range jb.Configs {
				appSupport := filepath.Join(home(), "Library", "Application Support", "JetBrains")
				// Find matching IDE directory
				entries, err := os.ReadDir(appSupport)
				if err != nil {
					emit(fmt.Sprintf("%s: %s", jb.Name, cfgPath), false, "JetBrains config dir not found")
					break
				}
				var targetDir string
				for _, e := range entries {
					if e.IsDir() && strings.Contains(e.Name(), strings.ReplaceAll(jb.Name, " ", "")) {
						targetDir = filepath.Join(appSupport, e.Name())
					}
				}
				if targetDir == "" {
					emit(fmt.Sprintf("%s: %s", jb.Name, cfgPath), false, jb.Name+" not installed")
					break
				}
				fullPath := filepath.Clean(filepath.Join(targetDir, cfgPath))
				if !strings.HasPrefix(fullPath, targetDir) {
					emit(fmt.Sprintf("%s: %s", jb.Name, cfgPath), false, "path traversal blocked")
					continue
				}
				if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
					emit(fmt.Sprintf("%s: %s", jb.Name, cfgPath), false, err.Error())
					continue
				}
				emitResult(fmt.Sprintf("%s: %s", jb.Name, cfgPath), writeFile(fullPath, content))
			}
		}
	}

	if opts.ShouldRestore("configs") {
		for relPath, content := range p.ConfigFiles {
			fullPath := filepath.Clean(filepath.Join(home(), relPath))
			if !strings.HasPrefix(fullPath, home()) {
				emit(fmt.Sprintf("Restore %s", relPath), false, "path traversal blocked")
				continue
			}
			if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
				emit(fmt.Sprintf("Create dir for %s", relPath), false, err.Error())
				continue
			}
			emitResult(fmt.Sprintf("Restore %s", relPath), writeFile(fullPath, content))
		}
	}

	if opts.ShouldRestore("defaults") {
		for _, d := range p.Defaults.Settings {
			label := fmt.Sprintf("defaults write %s %s", d.Domain, d.Key)
			if current := runOutput("defaults", "read", d.Domain, d.Key); current == d.Value {
				emit(label, true, "already installed")
				continue
			}
			typeFlag := "-" + d.Type
			emitResult(label, runSilent("defaults", "write", d.Domain, d.Key, typeFlag, d.Value))
		}
	}

	if opts.ShouldRestore("languages") {
		if commandExists("npm") {
			for _, pkg := range p.Languages.NpmGlobals {
				emitResult(fmt.Sprintf("npm install -g %s", pkg), runSilent("npm", "install", "-g", pkg))
			}
		}
		if commandExists("yarn") {
			for _, pkg := range p.Languages.YarnGlobals {
				emitResult(fmt.Sprintf("yarn global add %s", pkg), runSilent("yarn", "global", "add", pkg))
			}
		}
		if commandExists("pnpm") {
			for _, pkg := range p.Languages.PnpmGlobals {
				emitResult(fmt.Sprintf("pnpm add -g %s", pkg), runSilent("pnpm", "add", "-g", pkg))
			}
		}
		if commandExists("composer") {
			for _, pkg := range p.Languages.ComposerGlobals {
				emitResult(fmt.Sprintf("composer global require %s", pkg), runSilent("composer", "global", "require", pkg))
			}
		}
		if commandExists("gem") {
			// Check if we are using the restricted System Ruby
			rubyPath, err := exec.Command("which", "ruby").Output()
			isSystemRuby := err == nil && strings.HasPrefix(strings.TrimSpace(string(rubyPath)), "/usr/bin/ruby")

			if isSystemRuby {
				msg := "System ruby detected (/usr/bin/ruby). Native extensions will likely fail."
				if commandExists("brew") {
					msg += " Fix with: 'brew install ruby'"
				} else {
					msg += " Please install a non-system Ruby (e.g., via rbenv or brew)."
				}
				emitResult("Ruby Environment Check", fmt.Errorf("%s", msg))
			} else {
				for _, pkg := range p.Languages.GemGlobals {
					emitResult(fmt.Sprintf("gem install %s", pkg), runSilent("gem", "install", pkg))
				}
			}
		}
		if commandExists("cargo") {
			for _, pkg := range p.Languages.CargoPackages {
				emitResult(fmt.Sprintf("cargo install %s", pkg), runSilent("cargo", "install", pkg))
			}
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
