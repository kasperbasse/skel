package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var showCmd = &cobra.Command{
	Use:   "show [profile-name]",
	Short: "Show the contents of a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runShow,
}

// runShow displays all sections of a profile in detail.
func runShow(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	printProfileHeader("Profile", p.Name)
	fmt.Println()

	displayAllSections(p)
	fmt.Println()

	return nil
}

// displayAllSections iterates through all sections and displays details.
func displayAllSections(p *profile.Profile) {
	sectionCount := 0
	totalSections := countActiveSections(p)

	for _, g := range scanGroups {
		if g.ShowDetail == nil {
			continue
		}
		if summary := g.ScanSummary(p); summary == "" {
			continue
		}

		g.ShowDetail(p)
		sectionCount++

		if sectionCount < totalSections {
			fmt.Println()
		}
	}
}

// countActiveSections returns the number of sections with data to display.
func countActiveSections(p *profile.Profile) int {
	count := 0
	for _, g := range scanGroups {
		if g.ShowDetail != nil && g.ScanSummary(p) != "" {
			count++
		}
	}
	return count
}

func init() {
	showCmd.Flags().BoolVar(&showAll, "all", false, "Show all items without truncation")
}
