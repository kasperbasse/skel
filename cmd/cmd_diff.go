package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var diffCmd = &cobra.Command{
	Use:   "diff [profile-a] [profile-b]",
	Short: "Compare two profiles",
	Args:  requireExactArgs(2, "diff <profile-a> <profile-b>"),
	RunE:  runDiff,
}

// runDiff compares two profiles and displays differences.
func runDiff(_ *cobra.Command, args []string) error {
	profileA, err := loadProfileOrFail(args[0])
	if err != nil {
		return err
	}

	profileB, err := loadProfileOrFail(args[1])
	if err != nil {
		return err
	}

	printCommandHeader("diff", fmt.Sprintf("Comparing %s → %s", bold(args[0]), bold(args[1])))
	fmt.Println()

	hasDifferences := displayComparison(profileA, profileB)

	if !hasDifferences {
		fmt.Printf("  %s These profiles are identical. No differences found.\n\n", iconCheck())
	}

	return nil
}

// displayComparison prints differences between two profiles.
// Returns true if any differences were found.
func displayComparison(profileA, profileB *profile.Profile) bool {
	hasDiff := false

	// Compare regular profile sections (lists)
	for _, section := range profileSections {
		added, removed := diffSlices(section.Items(profileA), section.Items(profileB))
		if len(added) == 0 && len(removed) == 0 {
			continue
		}
		hasDiff = true
		printDiffSection(section.Icon, section.Label, added, removed)
	}

	// Define a slice of comparison functions
	type diffFunc struct {
		fn func(*profile.Profile, *profile.Profile) driftSection
	}

	diffs := []diffFunc{
		{computeVersionDrift},
		{computeConfigDrift},
		{computeShellDrift},
		{computeGitConfigDrift},
		{computeJetBrainsConfigsDrift},
		{computeSystemDrift},
	}

	for _, d := range diffs {
		diff := d.fn(profileA, profileB)
		if hasChanges(diff) {
			hasDiff = true
			printChangedItemsWithHeader(diff)
			fmt.Println()
		}
	}

	return hasDiff
}
