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
	RunE: runStatus,
}

// runStatus displays a summary of a profile.
func runStatus(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	printProfileStatusSummary(p)
	return nil
}

// printProfileStatusSummary prints profile summary information.
func printProfileStatusSummary(p *profile.Profile) {
	printProfileHeader("Status", p.Name)

	ago := timeAgo(p.CreatedAt)
	counts := appmeta.CountsForProfile(p)
	readiness := string(appmeta.ReadinessForProfile(p))

	fmt.Printf("  %s\n", internalui.ReadinessBadge(readiness))
	fmt.Printf("  %s · %s items · %s · macOS %s\n\n",
		dim(ago),
		num(counts.Total),
		dim(p.Machine),
		p.System.MacOSVersion,
	)
}

func init() {
	statusCmd.ValidArgsFunction = singleProfileCompletion
}
