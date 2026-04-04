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

func runShow(_ *cobra.Command, args []string) error {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	p, err := profile.Load(name)
	if err != nil {
		return enhanceError(err)
	}

	printProfileHeader("Profile", p.Name)
	fmt.Println()

	var x = 1
	var scanGroupsLength = len(scanGroups)
	for _, g := range scanGroups {
		if g.ShowDetail == nil {
			continue
		}
		if summary := g.ScanSummary(p); summary == "" {
			continue
		}
		g.ShowDetail(p)
		x++
		if x < scanGroupsLength {
			fmt.Println()
		}
	}

	return nil
}

func init() {
	showCmd.Flags().BoolVar(&showAll, "all", false, "Show all items without truncation")
}
