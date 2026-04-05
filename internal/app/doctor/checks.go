package doctor

import (
	"fmt"

	internalui "github.com/kasperbasse/skel/internal/ui"
)

// Check is a rendered doctor check row.
type Check struct {
	Label string
	OK    bool
	Fix   string
}

// ToolResolver maps a tool command to display metadata.
type ToolResolver func(command string) (label, validatorCmd, fix string, ok bool)

// ToolExists checks whether a tool command is available.
type ToolExists func(command string) bool

// buildChecksWith builds doctor checks for a list of required tool commands.
func buildChecksWith(requiredTools []string, resolve ToolResolver, exists ToolExists) []Check {
	if resolve == nil || exists == nil {
		return nil
	}

	checks := make([]Check, 0, len(requiredTools))
	for _, cmd := range requiredTools {
		label, validatorCmd, fix, ok := resolve(cmd)
		if !ok {
			label = cmd
			validatorCmd = cmd
			fix = "Install and ensure it's in your PATH"
		}
		checks = append(checks, Check{Label: label, OK: exists(validatorCmd), Fix: fix})
	}
	return checks
}

func BuildChecks(requiredTools []string) []Check {
	return buildChecksWith(requiredTools, ToolDoctorInfo, CommandExists)
}

func PrintCheck(c Check) {
	if c.OK {
		fmt.Printf("  %s  %s\n", internalui.IconCheck(), c.Label)
	} else {
		fmt.Printf("  %s  %s %s %s\n",
			internalui.IconCross(),
			internalui.Bold(c.Label),
			internalui.Dim("·"),
			internalui.Dim(c.Fix),
		)
	}
}

// RunChecks prints all checks for a list of required tool commands and returns
// the number of issues found and whether any checks were produced.
func RunChecks(requiredTools []string) (issues int, empty bool) {
	checks := BuildChecks(requiredTools)
	if len(checks) == 0 {
		return 0, true
	}
	for _, c := range checks {
		PrintCheck(c)
		if !c.OK {
			issues++
		}
	}
	return issues, false
}
