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
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := profile.Load(args[0])
		if err != nil {
			return enhanceError(err)
		}

		p, err := profile.Load(args[1])
		if err != nil {
			return enhanceError(err)
		}

		fmt.Printf("\n  %s Comparing %s → %s\n", cyan(headlineIcon("diff")), bold(args[0]), bold(args[1]))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		hasDiff := false
		for _, s := range profileSections {
			added, removed := diffSlices(s.Items(a), s.Items(p))
			if len(added) == 0 && len(removed) == 0 {
				continue
			}
			hasDiff = true
			printDiffSection(s.Icon, s.Label, added, removed)
		}

		if !hasDiff {
			fmt.Printf("  %s These profiles are identical. No differences found.\n\n", iconCheck())
		}

		return nil
	},
}
