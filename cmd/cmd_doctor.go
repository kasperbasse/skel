package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [profile-name]",
	Short: "Check that a profile can be restored on this machine",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDoctor,
}

// runDoctor validates if a profile can be restored.
func runDoctor(_ *cobra.Command, args []string) error {
	name := SelectProfileName(args)

	p, err := LoadAnyProfile(name)
	if err != nil {
		return err
	}

	PrintCommandHeader("doctor", fmt.Sprintf("Checking %s", bold("'"+p.Name+"'")))

	issues, empty := appdoctor.RunChecks(appdoctor.RequiredTools(p))
	if empty {
		fmt.Printf("  %s Profile has no restorable sections.\n\n", iconDash())
		return nil
	}

	fmt.Println()
	return reportDoctorResults(name, issues)
}

// reportDoctorResults displays doctor check results and next steps.
func reportDoctorResults(name string, issues int) error {
	if issues == 0 {
		fmt.Printf("  %s All tools present. Ready to restore.\n\n", iconCheck())
		printNextSteps(
			nextStep("skel restore "+name, "to apply this profile"),
		)
		return nil
	}

	fmt.Printf("  %s %s - install missing tools then run %s\n",
		iconWarn(),
		bold(fmt.Sprintf("%d issue%s found", issues, pluralS(issues))),
		cyan("skel restore "+name),
	)
	printNextSteps(
		nextStep("Install the missing tools", "listed above"),
		nextStep("skel doctor "+name, "to verify"),
	)
	return nil
}

// pluralS returns "s" if n != 1, else empty string.
func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	doctorCmd.ValidArgsFunction = singleProfileCompletion
}
