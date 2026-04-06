package cmd

import (
	"fmt"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

func showHomebrew(p *profile.Profile) {
	printSection(iconHomebrew, fmt.Sprintf("Homebrew (%s, %s)",
		countLabel(len(p.Homebrew.Formulas), "formula", "formulas"),
		countLabel(len(p.Homebrew.Casks), "cask", "casks"),
	))
	printList(p.Homebrew.Formulas, 15)
	if len(p.Homebrew.Casks) > 0 {
		fmt.Println()
		printSection(iconPackage, fmt.Sprintf("Casks (%s)", num(len(p.Homebrew.Casks))))
		printList(p.Homebrew.Casks, 15)
	}
	if len(p.Homebrew.MasApps) > 0 {
		fmt.Println()
		printSection(iconMas, fmt.Sprintf("App Store (%s)", num(len(p.Homebrew.MasApps))))
		for _, app := range p.Homebrew.MasApps {
			printBullet(fmt.Sprintf("%s (%s)", app.Name, app.ID))
		}
	}
	if len(p.Homebrew.Taps) > 0 {
		fmt.Println()
		printSection(iconTaps, fmt.Sprintf("Homebrew Taps (%s)", num(len(p.Homebrew.Taps))))
		for _, t := range p.Homebrew.Taps {
			printBullet(t)
		}
	}
}

func showShell(p *profile.Profile) {
	printSection(iconShell, "Shell: "+summarizeShell(p.Shell))
	hasDetails := false
	if p.Shell.OhMyZsh && len(p.Shell.OhMyZshPlugins) > 0 {
		printBullet("Plugins: " + strings.Join(p.Shell.OhMyZshPlugins, ", "))
		hasDetails = true
	}
	if len(p.Shell.FishPlugins) > 0 {
		printBullet("Fish plugins: " + strings.Join(p.Shell.FishPlugins, ", "))
		hasDetails = true
	}
	if len(p.Shell.Aliases) > 0 {
		printBullet(fmt.Sprintf("%s aliases", num(len(p.Shell.Aliases))))
		hasDetails = true
	}
	if !hasDetails {
		printBullet(dim("No plugins or aliases captured"))
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
		printSection(iconEditors, fmt.Sprintf("VS Code (%s extensions)", num(len(p.Editor.VSCodeExts))))
		printList(p.Editor.VSCodeExts, 15)
	}
	if p.Editor.Cursor {
		printEditorGap()
		printSection(iconCursor, fmt.Sprintf("Cursor (%s extensions)", num(len(p.Editor.CursorExts))))
		printList(p.Editor.CursorExts, 15)
	}
	if p.Editor.Neovim && len(p.Editor.NeovimPlugins) > 0 {
		printEditorGap()
		printSection(iconNeovim, fmt.Sprintf("Neovim (%s plugins)", num(len(p.Editor.NeovimPlugins))))
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
		printSection(iconJetBrains, fmt.Sprintf("%s %s%s%s", jb.Name, cyan(jb.Version), pluginInfo, configInfo))
		printList(jb.Plugins, 15)
	}
}

func showGit(p *profile.Profile) {
	if p.Git.UserName == "" && p.Git.UserEmail == "" && p.Git.DefaultBranch == "" {
		return
	}

	printSection(iconGit, "Git")
	if p.Git.UserName != "" || p.Git.UserEmail != "" {
		printBullet(fmt.Sprintf("%s %s", p.Git.UserName, dim("<"+p.Git.UserEmail+">")))
	}
	if p.Git.DefaultBranch != "" {
		printBullet("Default branch: " + cyan(p.Git.DefaultBranch))
	}
}

func showLanguages(p *profile.Profile) {
	if summarizeVersions(p) == "" {
		return
	}
	printSection(iconLanguages, "Languages")
	printVersionDetails(p)
}

func showConfigs(p *profile.Profile) {
	if len(p.ConfigFiles) == 0 {
		return
	}
	printSection(iconConfigs, fmt.Sprintf("Config files (%s)", num(len(p.ConfigFiles))))
	for path := range p.ConfigFiles {
		printBullet("~/" + path)
	}
}

func showDefaults(p *profile.Profile) {
	if len(p.Defaults.Settings) == 0 {
		return
	}
	printSection(iconDefaults, fmt.Sprintf("macOS Defaults (%s)", num(len(p.Defaults.Settings))))
	for _, d := range p.Defaults.Settings {
		printBullet(fmt.Sprintf("%s %s = %s", dim(d.Domain), d.Key, cyan(d.Value)))
	}
}

func showSSH(p *profile.Profile) {
	if len(p.SSH.Keys) == 0 {
		return
	}
	printSection(iconSSH, fmt.Sprintf("SSH Keys (%s) - fingerprints only, no private keys stored", num(len(p.SSH.Keys))))
	for _, key := range p.SSH.Keys {
		printBullet(formatSSHKey(key))
	}
}
