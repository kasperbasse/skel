package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kasperbasse/skel/internal/profile"
)

var statusCmd = &cobra.Command{
	Use:   "status [profile-name]",
	Short: "Show a one-line summary of a profile",
	Long: `Show a one-line summary: profile name, last scan time, and item count.
Fast by default (reads the saved profile, no rescan). Use 'skel drift' to
check what has changed since the last scan.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		p, err := profile.Load(name)
		if err != nil {
			all, listErr := profile.ListAll()
			if listErr == nil && len(all) == 0 {
				printFirstRun()
				return nil
			}
			return err
		}

		ago := timeAgo(p.CreatedAt)
		items := profileItemCount(p)

		fmt.Printf("\n  %s %s\n", cyan("📦"), bold(p.Name))
		fmt.Printf("  %s\n", dividerStyle.Render("────────────────────────────────────────────"))
		fmt.Printf("  %s · %d items\n\n", dim(ago), items)
		return nil
	},
}

func init() {
	statusCmd.ValidArgsFunction = singleProfileCompletion
}
