package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/cmd/tui"
	"github.com/kasperbasse/skel/internal/profile"
	"github.com/kasperbasse/skel/internal/scanner"
)

var driftCmd = &cobra.Command{
	Use:   "drift [profile-name]",
	Short: "Detect what's changed since last scan",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		saved, err := profile.Load(name)
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s Checking for drift against %s\n", cyan(headlineIcon("drift")), bold("'"+name+"'"))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s\n\n", dim(fmt.Sprintf("Saved %s from %s", saved.CreatedAt.Format("Jan 02 2006"), saved.Machine)))

		var current *profile.Profile
		var warnings []string

		if IsInteractive() {
			m := tui.NewScanModel(name, "Scanning current state...")
			prog := tea.NewProgram(m)
			finalModel, err := prog.Run()
			if err != nil {
				return fmt.Errorf("scan failed: %w", err)
			}
			scanModel, ok := finalModel.(tui.ScanModel)
			if !ok {
				return fmt.Errorf("unexpected model type from scan")
			}
			result := scanModel.Result()
			if result == nil {
				return fmt.Errorf("scan was interrupted")
			}
			if result.Err != nil {
				return result.Err
			}
			current = result.Profile
			warnings = result.Warnings
		} else {
			spin := NewSpinner("Scanning current state...")
			spin.Start()
			current, warnings, err = scanner.Run(name)
			spin.Stop()
			if err != nil {
				return err
			}
		}

		fmt.Println()

		if len(warnings) > 0 {
			for _, w := range warnings {
				fmt.Printf("  %s %s\n", iconWarn(), dim(w))
			}
			fmt.Println()
		}

		changes := computeDrift(saved, current)

		if len(changes) == 0 {
			fmt.Printf("  %s No drift detected. Your Mac matches the profile.\n", iconCheck())
			return nil
		}

		total := countDriftItems(changes)
		fmt.Printf("  %s Found %s changes since last scan:\n\n",
			iconWarn(), cyan(fmt.Sprintf("%d", total)))

		for _, c := range changes {
			if len(c.added) == 0 && len(c.removed) == 0 {
				continue
			}
			count := len(c.added) + len(c.removed)
			fmt.Printf("  %s %s %s\n", c.icon, bold(c.title), dim(fmt.Sprintf("(%d)", count)))
			for _, item := range c.added {
				fmt.Printf("     %s %s\n", green("+"), green(item))
			}
			for _, item := range c.removed {
				fmt.Printf("     %s %s\n", red("-"), red(item))
			}
			fmt.Println()
		}

		fmt.Printf("  %s\n\n", dim("Run 'skel update "+name+"' to save the current state"))
		return nil
	},
}

type driftSection struct {
	icon    string
	title   string
	added   []string
	removed []string
}

func computeDrift(saved, current *profile.Profile) []driftSection {
	var sections []driftSection

	// List-based diffs from section registry.
	// Config Files is excluded because drift also checks content modifications.
	for _, s := range profileSections {
		if s.Label == "Config Files" {
			continue
		}
		savedItems := s.Items(saved)
		currentItems := s.Items(current)
		// If the current scan returned nothing but saved had items,
		// the tool likely isn't in PATH - don't report as mass removal.
		if len(currentItems) == 0 && len(savedItems) > 0 {
			continue
		}
		added, removed := diffSlices(savedItems, currentItems)
		if len(added) > 0 || len(removed) > 0 {
			sections = append(sections, driftSection{s.Icon, s.Label, added, removed})
		}
	}

	// Version changes
	var versionChanges driftSection
	versionChanges.icon = iconLanguages
	versionChanges.title = "Language Versions"
	for _, v := range versionFields {
		savedVer := v.Value(saved)
		currentVer := v.Value(current)
		if savedVer == currentVer || currentVer == "" {
			continue
		}
		if savedVer == "" {
			versionChanges.added = append(versionChanges.added, fmt.Sprintf("%s %s", v.Label, currentVer))
		} else {
			versionChanges.added = append(versionChanges.added, fmt.Sprintf("%s %s (was %s)", v.Label, currentVer, savedVer))
		}
	}
	if len(versionChanges.added) > 0 {
		sections = append(sections, versionChanges)
	}

	// Config files (also checks content modifications, not just added/removed)
	var configAdded, configRemoved []string
	for path := range current.ConfigFiles {
		if _, ok := saved.ConfigFiles[path]; !ok {
			configAdded = append(configAdded, "~/"+path)
		}
	}
	for path := range saved.ConfigFiles {
		if _, ok := current.ConfigFiles[path]; !ok {
			configRemoved = append(configRemoved, "~/"+path)
		}
	}
	for path, currentContent := range current.ConfigFiles {
		if savedContent, ok := saved.ConfigFiles[path]; ok && savedContent != currentContent {
			configAdded = append(configAdded, fmt.Sprintf("~/%s (modified)", path))
		}
	}
	if len(configAdded) > 0 || len(configRemoved) > 0 {
		sections = append(sections, driftSection{iconConfigs, "Config Files", configAdded, configRemoved})
	}

	// Shell config changes
	var shellChanges driftSection
	shellChanges.icon = "🐚"
	shellChanges.title = "Shell Config"
	for _, f := range shellContentFields {
		if f.Value(saved) != f.Value(current) && f.Value(current) != "" {
			shellChanges.added = append(shellChanges.added, f.Label+" (modified)")
		}
	}
	savedAliases := strings.Join(saved.Shell.Aliases, "\n")
	currentAliases := strings.Join(current.Shell.Aliases, "\n")
	if savedAliases != currentAliases {
		added, removed := diffSlices(saved.Shell.Aliases, current.Shell.Aliases)
		shellChanges.added = append(shellChanges.added, added...)
		shellChanges.removed = append(shellChanges.removed, removed...)
	}
	if len(shellChanges.added) > 0 || len(shellChanges.removed) > 0 {
		sections = append(sections, shellChanges)
	}

	return sections
}

func countDriftItems(sections []driftSection) int {
	n := 0
	for _, s := range sections {
		n += len(s.added) + len(s.removed)
	}
	return n
}
