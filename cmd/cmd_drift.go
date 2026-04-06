package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var driftCmd = &cobra.Command{
	Use:   "drift [profile-name]",
	Short: "Detect what's changed since last scan",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDrift,
}

// runDrift detects changes between saved profile and current machine state.
func runDrift(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	// Load saved profile
	saved, err := profile.Load(name)
	if err != nil {
		return enhanceError(err)
	}

	PrintCommandHeader("drift", fmt.Sprintf("Checking for drift against %s", bold("'"+name+"'")))
	fmt.Printf("  %s\n\n", dim(fmt.Sprintf("Saved %s from %s", saved.CreatedAt.Format(dateFormat), saved.Machine)))

	// Scan current state
	current, warnings, err := performScan(name)
	if err != nil {
		return enhanceError(err)
	}

	// Display warnings
	PrintWarnings(warnings)

	// Compute and display changes
	return displayDriftComparison(name, saved, current)
}

// displayDriftComparison shows all changes found.
func displayDriftComparison(name string, saved, current *profile.Profile) error {
	changes := computeDrift(saved, current)

	if len(changes) == 0 {
		fmt.Printf("  %s No drift detected. Your Mac matches the profile.\n\n", iconCheck())
		return nil
	}

	total := countDriftItems(changes)
	fmt.Printf("  %s Found %s changes since last scan:\n\n",
		iconWarn(), cyan(fmt.Sprintf("%d", total)))

	for _, change := range changes {
		if !hasChanges(change) {
			continue
		}

		count := len(change.changed) + len(change.added) + len(change.removed)
		fmt.Printf("  %s %s %s\n", change.icon, bold(change.title), dim(fmt.Sprintf("(%d)", count)))

		printChangedItems("~", cyan, change.changed)
		printChangedItems("+", green, change.added)
		printChangedItems("-", red, change.removed)

		fmt.Println()
	}

	fmt.Printf("  %s\n\n", dim("Run 'skel update "+name+"' to save the current state"))
	return nil
}

type driftSection struct {
	icon    string
	title   string
	changed []string
	added   []string
	removed []string
}

func isLanguageVersionSection(label string) bool {
	for _, v := range versionFields {
		if v.Label == label {
			return true
		}
	}
	return false
}

func computeDrift(saved, current *profile.Profile) []driftSection {
	sections := computeListDrift(saved, current)
	if versionChanges := computeVersionDrift(saved, current); hasChanges(versionChanges) {
		sections = append(sections, versionChanges)
	}
	if configChanges := computeConfigDrift(saved, current); hasChanges(configChanges) {
		sections = append(sections, configChanges)
	}
	if shellChanges := computeShellDrift(saved, current); hasChanges(shellChanges) {
		sections = append(sections, shellChanges)
	}
	if gitConfigChanges := computeGitConfigDrift(saved, current); hasChanges(gitConfigChanges) {
		sections = append(sections, gitConfigChanges)
	}
	if jetBrainsConfigChanges := computeJetBrainsConfigsDrift(saved, current); hasChanges(jetBrainsConfigChanges) {
		sections = append(sections, jetBrainsConfigChanges)
	}
	if systemChanges := computeSystemDrift(saved, current); hasChanges(systemChanges) {
		sections = append(sections, systemChanges)
	}
	return sections
}

// computeListDrift compares regular profile sections.
func computeListDrift(saved, current *profile.Profile) []driftSection {
	var sections []driftSection
	for _, s := range profileSections {
		if s.Label == "Config Files" || isLanguageVersionSection(s.Label) {
			continue
		}
		savedItems := s.Items(saved)
		currentItems := s.Items(current)
		if len(currentItems) == 0 && len(savedItems) > 0 {
			continue
		}
		added, removed := diffSlices(savedItems, currentItems)
		if len(added) > 0 || len(removed) > 0 {
			sections = append(sections, driftSection{icon: s.Icon, title: s.Label, added: added, removed: removed})
		}
	}
	return sections
}

// countDriftItems counts total number of added/removed/changed items in drift sections.
func countDriftItems(sections []driftSection) int {
	n := 0
	for _, s := range sections {
		n += len(s.changed) + len(s.added) + len(s.removed)
	}
	return n
}
