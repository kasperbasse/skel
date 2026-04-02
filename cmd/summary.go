package cmd

import (
	"fmt"
	"strings"

	"github.com/kasperbasse/skel/internal/profile"
)

// num formats an integer with cyan color for use in summaries.
func num(n int) string { return cyan(fmt.Sprintf("%d", n)) }

// summarizeBrew returns a one-line summary of the Homebrew section.
func summarizeBrew(h profile.HomebrewProfile) string {
	detail := fmt.Sprintf("%s formulas · %s casks · %s App Store apps",
		num(len(h.Formulas)), num(len(h.Casks)), num(len(h.MasApps)))
	if len(h.Taps) > 0 {
		detail += fmt.Sprintf(" · %s taps", num(len(h.Taps)))
	}
	return detail
}

// summarizeShell returns a one-line summary of the Shell section.
func summarizeShell(s profile.ShellProfile) string {
	detail := cyan(s.Shell)
	if s.OhMyZsh {
		detail += fmt.Sprintf(" · Oh My Zsh (%s, %s plugins)",
			dim(s.OhMyZshTheme), num(len(s.OhMyZshPlugins)))
	}
	if s.Starship {
		detail += " · Starship"
	}
	totalAliases := len(s.Aliases) + len(s.BashAliases) + len(s.FishAbbreviations)
	if totalAliases > 0 {
		detail += fmt.Sprintf(" · %s aliases", num(totalAliases))
	}
	if s.FishConfig != "" {
		detail += " · Fish"
	}
	if s.BashrcContent != "" {
		detail += " · Bash"
	}
	return detail
}

// summarizeEditors returns a one-line summary of all detected editors.
func summarizeEditors(e profile.EditorProfile) string {
	var editors []string
	if e.VSCode {
		editors = append(editors, fmt.Sprintf("VS Code (%s exts)", num(len(e.VSCodeExts))))
	}
	if e.Cursor {
		editors = append(editors, fmt.Sprintf("Cursor (%s exts)", num(len(e.CursorExts))))
	}
	if e.Neovim {
		nvimDetail := "Neovim"
		if len(e.NeovimPlugins) > 0 {
			nvimDetail += fmt.Sprintf(" (%s plugins)", num(len(e.NeovimPlugins)))
		}
		editors = append(editors, nvimDetail)
	}
	for _, jb := range e.JetBrains {
		detail := jb.Name
		if len(jb.Plugins) > 0 {
			detail += fmt.Sprintf(" (%s plugins)", num(len(jb.Plugins)))
		}
		editors = append(editors, detail)
	}
	return strings.Join(editors, " · ")
}

// hasEditors returns true if any editor was detected.
func hasEditors(e profile.EditorProfile) bool {
	return e.VSCode || e.Cursor || e.Neovim || len(e.JetBrains) > 0
}

// printSection prints a section header with icon and bold title.
func printSection(icon, title string) {
	fmt.Printf("  %s %s\n\n", icon, bold(title))
}

// printBullet prints an indented bullet point.
func printBullet(s string) {
	fmt.Printf("     %s %s\n", dim("·"), s)
}

// printList prints a list of items, truncating after max with a "and N more..." line.
func printList(items []string, max int) {
	for i, item := range items {
		if i >= max {
			fmt.Printf("     %s %s\n", dim("·"), dim(fmt.Sprintf("and %d more...", len(items)-max)))
			break
		}
		printBullet(item)
	}
}

// dryRunBullet prints a dry-run command line with $ prefix.
func dryRunBullet(s string) { fmt.Printf("     %s %s\n", dim("$"), dim(s)) }

// printRow prints a compact row for scan output: icon + label + detail.
func printRow(icon, label, detail string) {
	fmt.Printf("  %s  %-10s %s\n", icon, bold(label), detail)
}

// printVersionDetails prints the detailed language version + package breakdown.
func printVersionDetails(p *profile.Profile) {
	for _, v := range versionFields {
		ver := v.Value(p)
		if ver == "" {
			continue
		}
		if v.DisplayLabel != "" {
			printBullet(v.DisplayLabel + " " + cyan(ver))
		} else {
			printBullet(cyan(ver))
		}
	}
	// Package manager globals
	type pkgGroup struct {
		label string
		items func(p *profile.Profile) []string
	}
	groups := []pkgGroup{
		{"npm globals", func(p *profile.Profile) []string { return p.Languages.NpmGlobals }},
		{"yarn globals", func(p *profile.Profile) []string { return p.Languages.YarnGlobals }},
		{"pnpm globals", func(p *profile.Profile) []string { return p.Languages.PnpmGlobals }},
		{"pip packages", func(p *profile.Profile) []string { return p.Languages.PipGlobals }},
		{"composer globals", func(p *profile.Profile) []string { return p.Languages.ComposerGlobals }},
		{"gems", func(p *profile.Profile) []string { return p.Languages.GemGlobals }},
		{"cargo packages", func(p *profile.Profile) []string { return p.Languages.CargoPackages }},
	}
	for _, g := range groups {
		items := g.items(p)
		if len(items) > 0 {
			printBullet(fmt.Sprintf("%s: %s packages", g.label, num(len(items))))
		}
	}
}

// formatSSHKey formats an SSH key for display with type, fingerprint, and comment.
func formatSSHKey(key profile.SSHKey) string {
	detail := key.Filename
	if key.Type != "" {
		detail += " (" + cyan(key.Type) + ")"
	}
	if key.Fingerprint != "" {
		detail += " " + dim(key.Fingerprint)
	}
	if key.Comment != "" {
		detail += " " + dim(key.Comment)
	}
	if key.PublicOnly {
		detail += " " + yellow("[pub only]")
	}
	return detail
}
