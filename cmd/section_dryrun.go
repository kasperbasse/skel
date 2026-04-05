package cmd

import (
	"fmt"

	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/restore"
)

func dryRunHomebrew(p *profile.Profile, opts *restore.Options) {
	if opts.ShouldRestore("homebrew") {
		if len(p.Homebrew.Taps) > 0 {
			printSection(iconTaps, fmt.Sprintf("Would add %s Homebrew taps", num(len(p.Homebrew.Taps))))
			for _, t := range p.Homebrew.Taps {
				dryRunBullet("brew tap " + t)
			}
			fmt.Println()
		}
		printSection(iconHomebrew, fmt.Sprintf("Would install %s Homebrew formulas", num(len(p.Homebrew.Formulas))))
		for _, f := range p.Homebrew.Formulas {
			dryRunBullet("brew install " + f)
		}
		if len(p.Homebrew.Casks) > 0 {
			fmt.Println()
			printSection(iconPackage, fmt.Sprintf("Would install %s casks", num(len(p.Homebrew.Casks))))
			for _, c := range p.Homebrew.Casks {
				dryRunBullet("brew install --cask " + c)
			}
		}
	}
	if opts.ShouldRestore("mas") && len(p.Homebrew.MasApps) > 0 {
		fmt.Println()
		printSection(iconMas, fmt.Sprintf("Would install %s App Store apps", num(len(p.Homebrew.MasApps))))
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
	printSection(iconDoc, "Would restore shell configs")
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
		printSection(iconEditors, fmt.Sprintf("Would install %s VS Code extensions", num(len(p.Editor.VSCodeExts))))
	}
	if p.Editor.Cursor {
		fmt.Println()
		printSection(iconCursor, fmt.Sprintf("Would install %s Cursor extensions", num(len(p.Editor.CursorExts))))
	}
}

func dryRunGit(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("git") {
		return
	}
	if p.Git.GitConfigContent != "" {
		fmt.Println()
		printSection(iconGit, "Would restore git config")
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
		printSection(iconPackage, fmt.Sprintf("Would install %s %s globals", num(len(g.items)), g.label))
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
	printSection(iconConfigs, fmt.Sprintf("Would restore %s config files", num(len(p.ConfigFiles))))
	for path := range p.ConfigFiles {
		dryRunBullet("~/" + path)
	}
}

func dryRunDefaults(p *profile.Profile, opts *restore.Options) {
	if !opts.ShouldRestore("defaults") || len(p.Defaults.Settings) == 0 {
		return
	}
	fmt.Println()
	printSection(iconDefaults, fmt.Sprintf("Would apply %s macOS preferences", num(len(p.Defaults.Settings))))
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
	printSection(iconSSH, fmt.Sprintf("SSH Keys (%s) - manual setup required", num(len(p.SSH.Keys))))
	for _, key := range p.SSH.Keys {
		printBullet(dim(formatSSHKey(key)))
	}
	fmt.Printf("     %s\n", dim("SSH keys are never restored automatically - regenerate or transfer manually"))
}

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
