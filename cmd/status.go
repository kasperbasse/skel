package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	appmeta "github.com/kasperbasse/skel/internal/app/profilemeta"
	"github.com/kasperbasse/skel/internal/profile"
	internalui "github.com/kasperbasse/skel/internal/ui"
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
			return enhanceError(err)
		}

		ago := timeAgo(p.CreatedAt)
		counts := appmeta.CountsForProfile(p)

		printProfileHeader("Status", p.Name)
		fmt.Printf("  %s\n", internalui.ReadinessBadge(string(appmeta.ReadinessForProfile(p))))
		fmt.Printf("  %s · %s items · %s · macOS %s\n\n", dim(ago), num(counts.Total), dim(p.Machine), p.System.MacOSVersion)
		return nil
	},
}

func init() {
	statusCmd.ValidArgsFunction = singleProfileCompletion
}
