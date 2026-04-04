package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	appdoctor "github.com/kasperbasse/skel/internal/app/doctor"
	"github.com/kasperbasse/skel/internal/profile"
)

type check struct {
	label string
	ok    bool
	fix   string // install hint shown when not ok
}

var doctorCmd = &cobra.Command{
	Use:   "doctor [profile-name]",
	Short: "Check that a profile can be restored on this machine",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "default"
		if len(args) > 0 {
			name = args[0]
		}

		p, err := profile.Load(name)
		if err != nil {
			return enhanceError(err)
		}

		fmt.Printf("\n  %s Checking %s\n", cyan(headlineIcon("doctor")), bold("'"+p.Name+"'"))
		fmt.Printf("  %s\n\n", dividerStyle.Render("────────────────────────────────────────────"))

		checks := buildChecks(p)
		if len(checks) == 0 {
			fmt.Printf("  %s Profile has no restorable sections.\n\n", iconDash())
			return nil
		}

		issues := 0
		for _, c := range checks {
			printCheck(c)
			if !c.ok {
				issues++
			}
		}

		fmt.Println()
		if issues == 0 {
			fmt.Printf("  %s All tools present. Ready to restore.\n\n", iconCheck())
			printNextSteps(
				nextStep("skel restore "+name, "to apply this profile"),
			)
		} else {
			fmt.Printf("  %s %s - install missing tools then run %s\n",
				iconWarn(),
				bold(fmt.Sprintf("%d issue%s found", issues, pluralS(issues))),
				cyan("skel restore "+name),
			)
			printNextSteps(
				nextStep("Install the missing tools", "listed above"),
				nextStep("skel doctor "+name, "to verify"),
			)
		}
		return nil
	},
}

func buildChecks(p *profile.Profile) []check {
	rawChecks := appdoctor.BuildChecks(p, appdoctor.ToolDoctorInfo, appdoctor.CommandExists)
	checks := make([]check, 0, len(rawChecks))
	for _, c := range rawChecks {
		checks = append(checks, check{label: c.Label, ok: c.OK, fix: c.Fix})
	}

	return checks
}

func printCheck(c check) {
	if c.ok {
		fmt.Printf("  %s  %s\n", iconCheck(), c.label)
	} else {
		fmt.Printf("  %s  %s\n", iconCross(), bold(c.label))
		fmt.Printf("       %s  %s\n", dim("→"), dim(c.fix))
	}
}

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	doctorCmd.ValidArgsFunction = singleProfileCompletion
}
