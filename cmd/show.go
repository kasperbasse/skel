package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var showCmd = &cobra.Command{
	Use:   "show [profile-name]",
	Short: "Show the contents of a profile",
	Args:  requireArgs("show <profile-name>"),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	p, err := profile.Load(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("\n  %s  %s\n\n", cyan("📦"), bold(p.Name))
	fmt.Printf("      %s · %s · macOS %s\n\n",
		dim(timeAgo(p.CreatedAt)),
		p.System.ChipArch,
		p.System.MacOSVersion,
	)

	for _, g := range scanGroups {
		if g.ShowDetail == nil {
			continue
		}
		if summary := g.ScanSummary(p); summary == "" {
			continue
		}
		g.ShowDetail(p)
		fmt.Println()
	}

	fmt.Println()
	return nil
}

func init() {
	showCmd.Flags().BoolVar(&showAll, "all", false, "Show all items without truncation")
}
