package cmd

import (
	"fmt"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// hasChanges checks if a drift section has any changes.
func hasChanges(d driftSection) bool {
	return len(d.changed) > 0 || len(d.added) > 0 || len(d.removed) > 0
}

// printChangedItems prints a list of changes with a prefix.
func printChangedItems(prefix string, colorFunc func(string) string, items []string) {
	for _, item := range items {
		fmt.Printf("     %s %s\n", colorFunc(prefix), colorFunc(item))
	}
}

// computeVersionDrift compares language version fields.
func computeVersionDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: iconLanguages, title: "Language Versions"}
	for _, v := range versionFields {
		savedVer := v.Value(saved)
		currentVer := v.Value(current)
		if savedVer == currentVer {
			continue
		}
		if savedVer == "" && currentVer != "" {
			changes.added = append(changes.added, fmt.Sprintf("%s %s", v.Label, currentVer))
		} else if savedVer != "" && currentVer == "" {
			changes.removed = append(changes.removed, fmt.Sprintf("%s %s", v.Label, savedVer))
		} else {
			changes.changed = append(changes.changed, fmt.Sprintf("%s %s (was %s)", v.Label, currentVer, savedVer))
		}
	}
	return changes
}

// computeConfigDrift compares config file presence and content.
func computeConfigDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: iconConfigs, title: "Config Files"}
	for path := range current.ConfigFiles {
		if _, ok := saved.ConfigFiles[path]; !ok {
			changes.added = append(changes.added, "~/"+path)
		}
	}
	for path := range saved.ConfigFiles {
		if _, ok := current.ConfigFiles[path]; !ok {
			changes.removed = append(changes.removed, "~/"+path)
		}
	}
	for path, currentContent := range current.ConfigFiles {
		if savedContent, ok := saved.ConfigFiles[path]; ok && savedContent != currentContent {
			changes.changed = append(changes.changed, fmt.Sprintf("~/%s (modified)", path))
		}
	}
	return changes
}

// computeShellDrift compares shell configuration.
func computeShellDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: "🐚", title: "Shell Config"}
	for _, f := range shellContentFields {
		if f.Value(saved) != f.Value(current) && f.Value(current) != "" {
			changes.changed = append(changes.changed, f.Label+" (modified)")
		}
	}
	if strings.Join(saved.Shell.Aliases, "\n") != strings.Join(current.Shell.Aliases, "\n") {
		added, removed := diffSlices(saved.Shell.Aliases, current.Shell.Aliases)
		changes.added = append(changes.added, added...)
		changes.removed = append(changes.removed, removed...)
	}
	if strings.Join(saved.Shell.BashAliases, "\n") != strings.Join(current.Shell.BashAliases, "\n") {
		added, removed := diffSlices(saved.Shell.BashAliases, current.Shell.BashAliases)
		changes.added = append(changes.added, added...)
		changes.removed = append(changes.removed, removed...)
	}
	if saved.Shell.Shell != current.Shell.Shell {
		changes.changed = append(changes.changed, fmt.Sprintf("Shell %s (was %s)", current.Shell.Shell, saved.Shell.Shell))
	}
	if saved.Shell.OhMyZsh != current.Shell.OhMyZsh {
		changes.changed = append(changes.changed, fmt.Sprintf("Oh My Zsh %v (was %v)", current.Shell.OhMyZsh, saved.Shell.OhMyZsh))
	}
	if saved.Shell.OhMyZshTheme != current.Shell.OhMyZshTheme {
		changes.changed = append(changes.changed, fmt.Sprintf("Oh My Zsh theme %s (was %s)", current.Shell.OhMyZshTheme, saved.Shell.OhMyZshTheme))
	}
	if saved.Shell.Starship != current.Shell.Starship {
		changes.changed = append(changes.changed, fmt.Sprintf("Starship %v (was %v)", current.Shell.Starship, saved.Shell.Starship))
	}
	if strings.Join(saved.Shell.OhMyZshPlugins, "\n") != strings.Join(current.Shell.OhMyZshPlugins, "\n") {
		added, removed := diffSlices(saved.Shell.OhMyZshPlugins, current.Shell.OhMyZshPlugins)
		changes.added = append(changes.added, added...)
		changes.removed = append(changes.removed, removed...)
	}
	if strings.Join(saved.Shell.FishPlugins, "\n") != strings.Join(current.Shell.FishPlugins, "\n") {
		added, removed := diffSlices(saved.Shell.FishPlugins, current.Shell.FishPlugins)
		changes.added = append(changes.added, added...)
		changes.removed = append(changes.removed, removed...)
	}
	if strings.Join(saved.Shell.FishAbbreviations, "\n") != strings.Join(current.Shell.FishAbbreviations, "\n") {
		added, removed := diffSlices(saved.Shell.FishAbbreviations, current.Shell.FishAbbreviations)
		changes.added = append(changes.added, added...)
		changes.removed = append(changes.removed, removed...)
	}
	return changes
}

// computeGitConfigDrift compares Git config content, global ignore, and default branch.
func computeGitConfigDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: "🔧", title: "Git Config"}
	if saved.Git.GitConfigContent != current.Git.GitConfigContent {
		if saved.Git.GitConfigContent == "" && current.Git.GitConfigContent != "" {
			changes.added = append(changes.added, ".gitconfig (added)")
		} else if saved.Git.GitConfigContent != "" && current.Git.GitConfigContent == "" {
			changes.removed = append(changes.removed, ".gitconfig (removed)")
		} else {
			changes.changed = append(changes.changed, ".gitconfig (modified)")
		}
	}
	if saved.Git.GlobalIgnore != current.Git.GlobalIgnore {
		if saved.Git.GlobalIgnore == "" && current.Git.GlobalIgnore != "" {
			changes.added = append(changes.added, ".gitignore_global (added)")
		} else if saved.Git.GlobalIgnore != "" && current.Git.GlobalIgnore == "" {
			changes.removed = append(changes.removed, ".gitignore_global (removed)")
		} else {
			changes.changed = append(changes.changed, ".gitignore_global (modified)")
		}
	}
	if saved.Git.DefaultBranch != current.Git.DefaultBranch {
		if saved.Git.DefaultBranch == "" && current.Git.DefaultBranch != "" {
			changes.added = append(changes.added, "Default branch: "+current.Git.DefaultBranch)
		} else if saved.Git.DefaultBranch != "" && current.Git.DefaultBranch == "" {
			changes.removed = append(changes.removed, "Default branch: "+saved.Git.DefaultBranch)
		} else {
			changes.changed = append(changes.changed, fmt.Sprintf("Default branch: %s (was %s)", current.Git.DefaultBranch, saved.Git.DefaultBranch))
		}
	}
	return changes
}

// computeJetBrainsConfigsDrift compares JetBrains config files (per-IDE, per-file).
func computeJetBrainsConfigsDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: "🧠", title: "JetBrains Configs"}
	// For each JetBrains IDE in current, compare configs to saved
	savedIDEs := make(map[string]map[string]string)
	for _, jb := range saved.Editor.JetBrains {
		savedIDEs[jb.Name] = jb.Configs
	}
	for _, jb := range current.Editor.JetBrains {
		savedConfigs := savedIDEs[jb.Name]
		for path, content := range jb.Configs {
			if savedConfigs == nil {
				changes.added = append(changes.added, fmt.Sprintf("%s: %s (added)", jb.Name, path))
				continue
			}
			savedContent, ok := savedConfigs[path]
			if !ok {
				changes.added = append(changes.added, fmt.Sprintf("%s: %s (added)", jb.Name, path))
			} else if savedContent != content {
				changes.changed = append(changes.changed, fmt.Sprintf("%s: %s (modified)", jb.Name, path))
			}
		}
		for path := range savedConfigs {
			if _, ok := jb.Configs[path]; !ok {
				changes.removed = append(changes.removed, fmt.Sprintf("%s: %s (removed)", jb.Name, path))
			}
		}
	}
	return changes
}

// computeSystemDrift compares Hostname, MacOSVersion, and ChipArch.
func computeSystemDrift(saved, current *profile.Profile) driftSection {
	changes := driftSection{icon: "🖥", title: "System Info"}
	if saved.System.Hostname != current.System.Hostname {
		if saved.System.Hostname == "" && current.System.Hostname != "" {
			changes.added = append(changes.added, "Hostname: "+current.System.Hostname)
		} else if saved.System.Hostname != "" && current.System.Hostname == "" {
			changes.removed = append(changes.removed, "Hostname: "+saved.System.Hostname)
		} else {
			changes.changed = append(changes.changed, fmt.Sprintf("Hostname: %s (was %s)", current.System.Hostname, saved.System.Hostname))
		}
	}
	if saved.System.MacOSVersion != current.System.MacOSVersion {
		if saved.System.MacOSVersion == "" && current.System.MacOSVersion != "" {
			changes.added = append(changes.added, "macOS: "+current.System.MacOSVersion)
		} else if saved.System.MacOSVersion != "" && current.System.MacOSVersion == "" {
			changes.removed = append(changes.removed, "macOS: "+saved.System.MacOSVersion)
		} else {
			changes.changed = append(changes.changed, fmt.Sprintf("macOS: %s (was %s)", current.System.MacOSVersion, saved.System.MacOSVersion))
		}
	}
	if saved.System.ChipArch != current.System.ChipArch {
		if saved.System.ChipArch == "" && current.System.ChipArch != "" {
			changes.added = append(changes.added, "Chip: "+current.System.ChipArch)
		} else if saved.System.ChipArch != "" && current.System.ChipArch == "" {
			changes.removed = append(changes.removed, "Chip: "+saved.System.ChipArch)
		} else {
			changes.changed = append(changes.changed, fmt.Sprintf("Chip: %s (was %s)", current.System.ChipArch, saved.System.ChipArch))
		}
	}
	return changes
}

// printChangedItemsWithHeader prints a drift section with a header and lists of changes.
func printChangedItemsWithHeader(diff driftSection) {
	count := len(diff.changed) + len(diff.added) + len(diff.removed)
	fmt.Printf("  %s %s %s\n", diff.icon, bold(diff.title), dim(fmt.Sprintf("(%d)", count)))
	printChangedItems("~", cyan, diff.changed)
	printChangedItems("+", green, diff.added)
	printChangedItems("-", red, diff.removed)
}
